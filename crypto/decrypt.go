/*
Copyright © 2020 Alessandro Segala (@ItalyPaleAle)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package crypto

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/minio/sio"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// This error is returned if we're just returning the metadata from the file
var ErrMetadataOnly = errors.New("output stream is nil, only metadata was returned")

// DecryptFile decrypts a stream (in), streaming the result to out
// If the result stream is nil, it only returns the metadata and stops reading
// The function requires a masterKey, a 32-byte key for AES-256, which is used to un-wrap the unique key for the file
// The function optionally accepts a metadata callback. When the metadata is extracted from the file, the callback is invoked with the metadata. The callback is invoked before the function starts streaming data to the out stream
// The function returns the version and length of the header, the wrapped key, and an error if any
func DecryptFile(out io.Writer, in io.Reader, masterKey []byte, metadataCb MetadataCb) (uint16, int32, []byte, error) {
	// Get the file header which contains the wrapped key
	headerVersion, headerLen, wrappedKey, in, err := GetFileHeader(in)
	if err != nil {
		return headerVersion, headerLen, wrappedKey, err
	}

	// Decrypt the file starting from package #0
	err = DecryptPackages(out, in, headerVersion, wrappedKey, masterKey, 0, 0, -1, metadataCb)

	return headerVersion, headerLen, wrappedKey, err
}

// GetFileHeader returns the wrapped key from the file header read from the stream "in"
// It returns the version and length of the header, the wrapped key as well as a new stream that should be used as input stream
func GetFileHeader(in io.Reader) (uint16, int32, []byte, io.Reader, error) {
	// Peek the first 256 bytes at most
	peek := make([]byte, 256)
	n, err := io.ReadFull(in, peek)
	// Ignore the ErrUnexpectedEOF, which means that we read less than the requested size
	if err != nil && err != io.ErrUnexpectedEOF {
		return 0, 0, nil, nil, err
	}

	// Ensure we have at least 3 bytes
	if n < 3 {
		return 0, 0, nil, nil, errors.New("input stream ended too quickly")
	}

	// Get the length of the header then parse the header
	headerLen := binary.LittleEndian.Uint16(peek[0:2])
	header := Header{}
	err = json.Unmarshal(peek[2:headerLen+2], &header)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	// Ensure the header is valid
	if header.Version > 0x02 {
		return 0, 0, nil, nil, fmt.Errorf("file header uses version %d which is not supported", header.Version)
	}
	if len(header.Key) == 0 {
		return 0, 0, nil, nil, errors.New("invalid key found in file header")
	}

	// Put the first bytes after the header back into the stream
	in = io.MultiReader(bytes.NewReader(peek[headerLen+2:n]), in)

	return header.Version, int32(headerLen + 2), header.Key, in, nil
}

// DecryptPackages decrypts one or more packages/chunks of encrypted data (64kb + 32 bytes), streaming the result to out
// The function requires a wrapped key and the master key
// It also requires a sequence number, that is the number of the first package/chunk we expect to decrypt
// The function optionally accepts a metadata callback. When the metadata is extracted from the file (only from package #0), the callback is invoked with the metadata. The callback is invoked before the function starts streaming data to the out stream
func DecryptPackages(out io.Writer, in io.Reader, headerVersion uint16, wrappedKey []byte, masterKey []byte, seqNum, offset uint32, length int64, metadataCb MetadataCb) error {
	// Unwrap the key for the file, using the master key
	key, err := UnwrapKey(masterKey, wrappedKey)
	if err != nil {
		return err
	}

	// If we're reading from the first package, we need to extract metadata
	readMetadata := (seqNum == 0 && metadataCb != nil)
	// Create a writer that has a buffer of MaxMetadataLength+2 bytes, the maximum size of the metadata object (encoded as protobuf or JSON)
	w := &decryptWriter{
		OutStream:     out,
		Cb:            metadataCb,
		HeaderVersion: headerVersion,
		ReadMetadata:  readMetadata,
		Offset:        offset,
		Length:        length,
	}
	bw := bufio.NewWriterSize(w, MaxMetadataLength+2)

	// Decrypt the data using minio/sio
	dec, err := sio.DecryptWriter(bw, sio.Config{
		Key:            key,
		SequenceNumber: seqNum,
	})
	if err != nil {
		return err
	}

	// Copy the buffer
	if _, err := io.Copy(dec, in); err != nil {
		return err
	}
	if err := dec.Close(); err != nil {
		return err
	}

	// Flush whatever data is left in the buffer
	if err := bw.Flush(); err != nil {
		return err
	}

	return nil
}

// decryptWriter manages the data decrypted by sio, optionally returning the metadata
// If there's a length greater than -1, it only returns as many bytes from the decrypted streams
// Likewise, an offset greater than 0 makes it skip the first N bytes from the beginning of the stream (if there's an offset, there's no metadata parsing happening)
type decryptWriter struct {
	OutStream     io.Writer
	Cb            MetadataCb
	HeaderVersion uint16
	ReadMetadata  bool
	Offset        uint32
	Length        int64
}

func (w *decryptWriter) Write(p []byte) (n int, err error) {
	var start uint32 = 0
	// If we have an offset, we don't read metadata
	if w.Offset > 0 {
		w.ReadMetadata = false
	} else if w.ReadMetadata {
		// If the app wants us to start by reading metadata (from package #0)
		// This is ignored if we have an offset

		// Ensure we have at least 3 bytes
		if len(p) < 3 {
			return 0, errors.New("decrypted stream ended too quickly")
		}

		// Get the length of the metadata
		metadataLen := binary.LittleEndian.Uint16(p[0:2])
		if metadataLen > MaxMetadataLength {
			return 0, errors.New("invalid metadata length")
		}
		start = uint32(metadataLen) + 2
		metadata := Metadata{}
		// Version 0x01 uses JSON; newer versions use protobuf
		if w.HeaderVersion == 0x01 {
			err = protojson.Unmarshal(p[2:start], &metadata)
		} else {
			err = proto.Unmarshal(p[2:start], &metadata)
		}
		if err != nil {
			return 0, err
		}

		// Metadata is ready, so invoke the callback
		if w.Cb != nil {
			w.Cb(&metadata, int32(metadataLen+2))
		}
		w.ReadMetadata = false
	}

	// If the output stream is nil, we only wanted the headers, so return
	if w.OutStream == nil {
		return 0, ErrMetadataOnly
	}

	// Skip the bytes if we need to
	// Note that if we're here, we haven't read metadata so start is 0
	if w.Offset > 0 {
		l := uint32(len(p))
		if l <= w.Offset {
			// Do not copy anything
			w.Offset -= l
			return len(p), nil
		} else {
			// Skip the first bytes
			start += w.Offset
			w.Offset = 0
		}
	}

	// Pipe the (rest of the) data to the out stream
	if start < uint32(len(p)) {
		// Check if we need to copy certain bytes only (if Length >= 0)
		if w.Length == 0 {
			return len(p), nil
		} else if w.Length > 0 {
			l := int64(len(p)) - int64(start)
			if w.Length >= l {
				w.Length -= l
				_, err = w.OutStream.Write(p[start:])
			} else {
				end := w.Length + int64(start)
				_, err = w.OutStream.Write(p[start:end])
				w.Length = 0
			}
		} else {
			_, err = w.OutStream.Write(p[start:])
		}
		if err != nil {
			return 0, err
		}
	}

	return len(p), nil
}
