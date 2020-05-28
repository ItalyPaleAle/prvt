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
)

var ErrMetadataOnly = errors.New("output stream is nil, only metadata was returned")

// EncryptFile decrypts a stream (in), streaming the result to out
// The function requires a masterKey, a 32-byte key for AES-256, which is used to un-wrap the unique key for the file
// The function optionally accepts a metadata callback. When the metadata is extracted from the file, the callback is invoked with the metadata. The callback is invoked before the function starts streaming data to the out stream
func DecryptFile(out io.Writer, in io.Reader, masterKey []byte, metadataCb MetadataCb) error {
	// Peek the first 256 bytes at most
	peek := make([]byte, 256)
	n, err := io.ReadFull(in, peek)
	// Ignore the ErrUnexpectedEOF, which means that we read less than the requested size
	if err != nil && err != io.ErrUnexpectedEOF {
		return err
	}

	// Ensure we have at least 3 bytes
	if n < 3 {
		return errors.New("input stream ended too quickly")
	}

	// Get the length of the header then parse the header
	headerLen := binary.LittleEndian.Uint16(peek[0:2])
	header := Header{}
	err = json.Unmarshal(peek[2:headerLen+2], &header)
	if err != nil {
		return err
	}

	// Ensure the header is valid
	if header.Version != 0x01 {
		return fmt.Errorf("file header uses version %d which is not supported", header.Version)
	}
	if len(header.Key) == 0 {
		return errors.New("invalid key found in file header")
	}

	// Put the first bytes after the header back into the stream
	in = io.MultiReader(bytes.NewReader(peek[headerLen+2:n]), in)

	// Unwrap the key for the file, using the master key
	key, err := UnwrapKey(masterKey, header.Key)
	if err != nil {
		return err
	}

	// Create a writer that has a buffer of 1024 bytes, the maximum size of the metadata object (JSON-encoded)
	w := &decryptWriter{
		OutStream: out,
		Cb:        metadataCb,
	}
	bw := bufio.NewWriterSize(w, 1024)

	// Decrypt the data using minio/sio
	dec, err := sio.DecryptWriter(bw, sio.Config{
		Key: key,
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

// decryptWriter manages the data decrypted by sio, to get the metadata first
type decryptWriter struct {
	OutStream    io.Writer
	Cb           MetadataCb
	metadataRead bool
}

func (w *decryptWriter) Write(p []byte) (n int, err error) {
	// If we haven't read metadata yet, this is the first chunk
	start := 0
	if !w.metadataRead {
		// Ensure we have at least 3 bytes
		if len(p) < 3 {
			return 0, errors.New("decrypted stream ended too quickly")
		}

		// Get the length of the metadata
		metadataLen := binary.LittleEndian.Uint16(p[0:2])
		if metadataLen > 1022 {
			return 0, errors.New("invalid metadata length")
		}
		start = int(metadataLen) + 2
		metadata := Metadata{}
		err = json.Unmarshal(p[2:start], &metadata)
		if err != nil {
			return 0, err
		}

		// Metadata is ready, so invoke the callback
		if w.Cb != nil {
			w.Cb(&metadata)
		}
		w.metadataRead = true
	}

	// If the output stream is nil, we only wanted the headers, so return
	if w.OutStream == nil {
		return 0, ErrMetadataOnly
	}

	// Pipe the (rest of the) data to the out stream
	if start < len(p) {
		_, err = w.OutStream.Write(p[start:])
		if err != nil {
			return 0, err
		}
	}

	return len(p), nil
}
