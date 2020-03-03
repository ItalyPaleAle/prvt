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
	"golang.org/x/crypto/argon2"
)

func DecryptFile(out io.Writer, in io.Reader, masterKey []byte) error {
	// Peek the first 2kb at most
	peek := make([]byte, 2048)
	n, err := io.ReadFull(in, peek)
	// Ignore the ErrUnexpectedEOF, which means that we read less than the requested size
	if err != nil && err != io.ErrUnexpectedEOF {
		return err
	}

	// Ensure we have at least 3 bytes
	if n < 3 {
		return errors.New("Input stream ended too quickly")
	}

	// Get the length of the header
	headerLen := binary.LittleEndian.Uint16(peek[0:2])
	header := Header{}
	err = json.Unmarshal(peek[2:headerLen+2], &header)
	if err != nil {
		return err
	}

	// Ensure the header is valid
	if header.Version != 0x01 {
		return fmt.Errorf("File header uses version %d which is not supported", header.Version)
	}
	if len(header.Salt) != 16 {
		return errors.New("Invalid salt found in file header")
	}

	// Put the first bytes after the header back into the stream
	in = io.MultiReader(bytes.NewReader(peek[headerLen+2:]), in)

	// Derive the encryption key using Argon2id
	// From the docs: "The draft RFC recommends[2] time=1, and memory=64*1024 is a sensible number. If using that amount of memory (64 MB) is not possible in some contexts then the time parameter can be increased to compensate.""
	key := argon2.IDKey(masterKey, header.Salt, 1, 64*1024, 4, 32)

	// Decrypt the data using minio/sio
	dec, err := sio.DecryptWriter(out, sio.Config{
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

	return nil
}
