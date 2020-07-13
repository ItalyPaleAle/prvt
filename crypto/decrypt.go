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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/minio/sio"
)

var ErrMetadataOnly = errors.New("output stream is nil, only metadata was returned")

// DecryptPackage decrypts a single package of encrypted data (64kb + 32 bytes), streaming the result to out
// It also requires a sequence number, that is the number of the chunk we expect to decrypt
// The function optionally accepts a metadata callback. When the metadata is extracted from the file (only from package #0), the callback is invoked with the metadata. The callback is invoked before the function starts streaming data to the out stream
func DecryptPackage(out io.Writer, in *bytes.Buffer, key []byte, seqNum uint32, metadataCb MetadataCb) error {
	if in.Len() > (64*1024 + 32) {
		return errors.New("input buffer is too long")
	}

	// If this is package #0, we have metadata to read
	dst := out
	if seqNum == 0 {
		dst = &bytes.Buffer{}
	}

	// Read the input stream, decrypt the data using minio/sio, and stream to the output stream
	_, err := sio.Decrypt(dst, in, sio.Config{
		Key:            key,
		SequenceNumber: seqNum,
	})
	if err != nil {
		return err
	}

	// Get the metadata if this is the first package
	if seqNum == 0 {
		buf := dst.(*bytes.Buffer)
		// Ensure we have at least 3 bytes
		if buf.Len() < 3 {
			return errors.New("decrypted stream ended too quickly")
		}

		// Get the length of the metadata
		metadataLen := int(binary.LittleEndian.Uint16(buf.Next(2)))
		if metadataLen > 1022 {
			return errors.New("invalid metadata length")
		}
		metadata := Metadata{}
		err = json.Unmarshal(buf.Next(metadataLen), &metadata)
		if err != nil {
			return err
		}

		// Metadata is ready, so invoke the callback
		if metadataCb != nil {
			metadataCb(&metadata)
		}

		// Write the rest to out if it's not nil
		// If it's nil, it meant we wanted the metadata only
		if out != nil {
			_, err = buf.WriteTo(out)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetFileHeader returns the wrapped key from the file header read from the stream "in"
// It returns the length of the header, the wrapped key as well as a new stream that should be used as input stream
func GetFileHeader(in io.Reader) (uint16, []byte, io.Reader, error) {
	// Peek the first 256 bytes at most
	peek := make([]byte, 256)
	n, err := io.ReadFull(in, peek)
	// Ignore the ErrUnexpectedEOF, which means that we read less than the requested size
	if err != nil && err != io.ErrUnexpectedEOF {
		return 0, nil, nil, err
	}

	// Ensure we have at least 3 bytes
	if n < 3 {
		return 0, nil, nil, errors.New("input stream ended too quickly")
	}

	// Get the length of the header then parse the header
	headerLen := binary.LittleEndian.Uint16(peek[0:2])
	header := Header{}
	err = json.Unmarshal(peek[2:headerLen+2], &header)
	if err != nil {
		return 0, nil, nil, err
	}

	// Ensure the header is valid
	if header.Version != 0x01 {
		return 0, nil, nil, fmt.Errorf("file header uses version %d which is not supported", header.Version)
	}
	if len(header.Key) == 0 {
		return 0, nil, nil, errors.New("invalid key found in file header")
	}

	// Put the first bytes after the header back into the stream
	in = io.MultiReader(bytes.NewReader(peek[headerLen+2:n]), in)

	return (headerLen + 2), header.Key, in, nil
}
