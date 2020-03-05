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

package fs

import (
	"io"
	"os"

	"e2e/crypto"
)

// Local is the local file system
// This implementation does not rely on tags, as it's assumed that concurrency isn't an issue on a single machine
type Local struct {
	masterKey []byte
}

func (f *Local) SetMasterKey(key []byte) {
	f.masterKey = key
}

func (f *Local) Get(name string, out io.Writer, headerCb func(*crypto.Header)) (found bool, tag *interface{}, err error) {
	found = true

	// Open the file
	file, err := os.Open("test/" + name)
	if err != nil {
		if os.IsNotExist(err) {
			found = false
		}
		return
	}

	// Decrypt the data
	err = crypto.DecryptFile(out, file, f.masterKey, headerCb)
	if err != nil {
		return
	}

	return
}

func (f *Local) Set(name string, in io.Reader, tag *interface{}, fileName string, mimeType string, size int64) (err error) {
	// Create the file
	file, err := os.Create("test/" + name)
	if err != nil {
		return err
	}

	// Encrypt the data and write it to file
	err = crypto.EncryptFile(file, in, f.masterKey, fileName, mimeType, size)
	if err != nil {
		return err
	}

	return nil
}
