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

	"e2e/crypto"
)

// Fs is the interface for the filesystem
type Fs interface {
	// Init the object, by passing a connection string
	Init(connection string) error

	// SetMasterKey sets the master passphrase (used to encrypt/decrypt files) in the object
	SetMasterKey(key []byte)

	// Get returns a stream to a file in the filesystem
	// It also returns a tag (which might be empty) that should be passed to the Set method if you want to subsequentially update the contents of the file
	Get(name string, out io.Writer, headerCb func(*crypto.Header)) (found bool, tag interface{}, err error)

	// Set writes a stream to the file in the filesystem
	// If you pass a tag, the implementation might use that to ensure that the file on the filesystem hasn't been changed since it was read (optional)
	Set(name string, in io.Reader, tag interface{}, fileName string, mimeType string, size int64) (tagOut interface{}, err error)
}
