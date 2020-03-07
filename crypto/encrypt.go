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
	"io"

	"github.com/minio/sio"
)

// EncryptFile encrypts a stream (in), streaming the result to out
// The function requires a masterKey, a 32-byte key for AES-256, which is used to wrap a key unique for this file
// The function optionally accepts a metadata argument that will be encrypted at the beginning of the file
func EncryptFile(out io.WriteCloser, in io.Reader, masterKey []byte, metadata *Metadata) error {
	defer out.Close()

	// Generate a new key for this file, wrapped with the master key
	key, err := NewKey()
	if err != nil {
		return err
	}
	wrappedKey, err := WrapKey(masterKey, key)
	if err != nil {
		return err
	}

	// First, build the header
	// This contains the wrapped key too
	head := Header{
		Version: 0x01,
		Key:     wrappedKey,
	}
	headJSON, err := json.Marshal(head)
	if err != nil {
		return err
	}

	// Write the header to the stream
	// Start with the length
	headLen := make([]byte, 2)
	// Header must be at most (256-2) bytes (first 2 bytes are the length)
	if len(headJSON) > 254 {
		return errors.New("header object is too big")
	}
	binary.LittleEndian.PutUint16(headLen, uint16(len(headJSON)))
	_, err = out.Write(headLen)
	if err != nil {
		return err
	}
	_, err = out.Write(headJSON)
	if err != nil {
		return err
	}

	// Metadata, which is encrypted
	if metadata == nil {
		metadata = &Metadata{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// Metadata must be at most (1024-2) bytes (first 2 bytes are the length)
	if len(metadataJSON) > 1022 {
		return errors.New("metadata object is too big")
	}
	metadataLen := make([]byte, 2)
	binary.LittleEndian.PutUint16(metadataLen, uint16(len(metadataJSON)))

	// Write the metadata to a buffer
	metadataBuf := &bytes.Buffer{}
	_, err = metadataBuf.Write(metadataLen)
	if err != nil {
		return err
	}
	_, err = metadataBuf.Write(metadataJSON)
	if err != nil {
		return err
	}

	// Prepend the metadata to the data to encrypt
	reader := io.MultiReader(metadataBuf, in)

	// Encrypt the data using minio/sio and the file-specific key (un-wrapped)
	enc, err := sio.EncryptWriter(out, sio.Config{
		MinVersion: sio.Version20,
		Key:        key,
	})
	if err != nil {
		return err
	}

	// Copy the buffer
	if _, err := io.Copy(enc, reader); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}

	return nil
}
