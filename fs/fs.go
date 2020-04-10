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
	"strings"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"
)

// Get returns a store for the given connection string
func Get(connection string) (store Fs, err error) {
	store = nil

	// Get the name of the store
	pos := strings.Index(connection, ":")
	if pos < 1 {
		err = fmt.Errorf("invalid connection string")
		return
	}

	switch connection[0:pos] {
	case "file", "local":
		store = &Local{}
		err = store.Init(connection)
	case "azure", "azureblob":
		store = &AzureStorage{}
		err = store.Init(connection)
	case "s3", "minio":
		store = &S3{}
		err = store.Init(connection)
	default:
		err = fmt.Errorf("invalid connection string")
	}

	return
}

// Fs is the interface for the filesystem
type Fs interface {
	// Init the object, by passing a connection string
	Init(connection string) error

	// SetDataPath sets the path where the data is stored (read from the info file)
	SetDataPath(path string)

	// SetMasterKey sets the master passphrase (used to encrypt/decrypt files) in the object
	SetMasterKey(key []byte)

	// GetInfoFile returns the contents of the info file
	GetInfoFile() (info *infofile.InfoFile, err error)

	// SetInfoFile stores the info file
	SetInfoFile(info *infofile.InfoFile) (err error)

	// Get returns a stream to a file in the filesystem
	// It also returns a tag (which might be empty) that should be passed to the Set method if you want to subsequentially update the contents of the file
	Get(name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error)

	// Set writes a stream to the file in the filesystem
	// If you pass a tag, the implementation might use that to ensure that the file on the filesystem hasn't been changed since it was read (optional)
	Set(name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error)

	// Delete a file from the filesystem
	// If you pass a tag, the implementation might use that to ensure that the file on the filesystem hasn't been changed since it was read (optional)
	Delete(name string, tag interface{}) (err error)
}
