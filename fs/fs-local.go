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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"e2e/crypto"
	"e2e/utils"

	homedir "github.com/mitchellh/go-homedir"
)

// Local is the local file system
// This implementation does not rely on tags, as it's assumed that concurrency isn't an issue on a single machine
type Local struct {
	basePath  string
	masterKey []byte
}

func (f *Local) Init(connection string) error {
	// Ensure that connection starts with "local:" or "file:"
	if !strings.HasPrefix(connection, "local:") && !strings.HasPrefix(connection, "file:") {
		return fmt.Errorf("invalid scheme")
	}

	// Get the path
	path := connection[strings.Index(connection, ":")+1:]

	// Expand the tilde if needed
	path, err := homedir.Expand(path)
	if err != nil {
		return err
	}

	// Get the absolute path
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	// Ensure the path ends with a /
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Lastly, ensure the path exists
	exists, err := utils.PathExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("path does not exist: %s", path)
	}

	f.basePath = path

	return nil
}

func (f *Local) SetMasterKey(key []byte) {
	f.masterKey = key
}

func (f *Local) Get(name string, out io.Writer, headerCb func(*crypto.Header)) (found bool, tag *interface{}, err error) {
	found = true

	// Open the file
	file, err := os.Open(f.basePath + name)
	if err != nil {
		if os.IsNotExist(err) {
			found = false
		}
		return
	}

	// Check if the file has any content
	stat, err := file.Stat()
	if err != nil {
		return
	}
	if stat.Size() == 0 {
		found = false
		return
	}

	// Decrypt the data
	err = crypto.DecryptFile(out, file, f.masterKey, headerCb)
	if err != nil {
		return
	}

	return
}

func (f *Local) Set(name string, in io.Reader, tag *interface{}, fileName string, mimeType string, size int64) (tagOut *interface{}, err error) {
	// Create the file
	file, err := os.Create(f.basePath + name)
	if err != nil {
		return nil, err
	}

	// Encrypt the data and write it to file
	err = crypto.EncryptFile(file, in, f.masterKey, fileName, mimeType, size)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
