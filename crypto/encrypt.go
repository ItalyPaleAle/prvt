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
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"

	"github.com/minio/sio"
	"golang.org/x/crypto/argon2"
)

func EncryptFile(out io.WriteCloser, in io.Reader, masterKey []byte, fileName string, fileContentType string, fileSize uint32) error {
	defer out.Close()

	// Get the salt that will be used to generate the file's key
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}

	// First, build the header
	head := Header{
		Version:     0x01,
		Salt:        salt,
		Name:        fileName,
		ContentType: fileContentType,
		Size:        fileSize,
	}
	headJSON, err := json.Marshal(head)
	if err != nil {
		return err
	}

	// Write the header to the stream
	// Start with the length
	headLen := make([]byte, 2)
	// Header must be at most 2kb - 2 bytes (length)
	if len(headJSON) > 2046 {
		return errors.New("Header too big")
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

	// Derive the encryption key using Argon2id
	// From the docs: "The draft RFC recommends[2] time=1, and memory=64*1024 is a sensible number. If using that amount of memory (64 MB) is not possible in some contexts then the time parameter can be increased to compensate.""
	key := argon2.IDKey(masterKey, salt, 1, 64*1024, 4, 32)

	// Encrypt the data using minio/sio
	enc, err := sio.EncryptWriter(out, sio.Config{
		Key: key,
	})
	if err != nil {
		return err
	}

	// Copy the buffer
	if _, err := io.Copy(enc, in); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}

	return nil
}
