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
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"
)

var fsTypes = map[string]reflect.Type{}

// GetWithDictionary returns a store given the options dictionary
// The dictionary must have a key "type" for the type of fs to use
func GetWithDictionary(opts map[string]string) (store Fs, err error) {
	// Init the cache
	cache := &MetadataCache{}
	err = cache.Init()
	if err != nil {
		return
	}

	// Get the store object using some reflection magic
	if opts["type"] == "" {
		err = errors.New("missing key 'type' in opts dictionary")
		return
	}
	fsTyp, ok := fsTypes[opts["type"]]
	if !ok || fsTyp == nil {
		err = fmt.Errorf("invalid fs type")
		return
	}
	store = reflect.New(fsTyp).Interface().(Fs)
	err = store.InitWithDictionary(opts, cache)

	return
}

// GetWithConnectionString returns a store for the given connection string
func GetWithConnectionString(connection string) (store Fs, err error) {
	// Init the cache
	cache := &MetadataCache{}
	err = cache.Init()
	if err != nil {
		return
	}

	// Get the name of the store
	pos := strings.Index(connection, ":")
	if pos < 1 {
		err = fmt.Errorf("invalid connection string")
		return
	}

	// Get the store object using some reflection magic
	fsTyp, ok := fsTypes[connection[0:pos]]
	if !ok || fsTyp == nil {
		err = fmt.Errorf("invalid connection string")
		return
	}
	store = reflect.New(fsTyp).Interface().(Fs)
	err = store.InitWithConnectionString(connection, cache)

	return
}

// Fs is the interface for the filesystem
type Fs interface {
	// InitWithDictionary inits the object by passing an options dictionary
	InitWithDictionary(opts map[string]string, cache *MetadataCache) error

	// InitWithConnectionString inits the object by passing a connection string and the cache object
	InitWithConnectionString(connection string, cache *MetadataCache) error

	// SetDataPath sets the path where the data is stored (read from the info file)
	SetDataPath(path string)

	// GetDataPath returns the path where the data is stored
	GetDataPath() string

	// SetMasterKey sets the master key (used to encrypt/decrypt files) in the object
	SetMasterKey(keyId string, key []byte)

	// GetMasterKey returns the master key
	GetMasterKey() []byte

	// GetKeyId returns the ID of the key used
	GetKeyId() string

	// GetInfoFile returns the contents of the info file
	GetInfoFile() (info *infofile.InfoFile, err error)

	// SetInfoFile stores the info file
	SetInfoFile(info *infofile.InfoFile) (err error)

	// Get returns a stream to a file in the filesystem
	// It also returns a tag (which might be empty) that should be passed to the Set method if you want to subsequentially update the contents of the file
	Get(ctx context.Context, name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error)

	// GetWithRange is like Get, but accepts a custom range
	GetWithRange(ctx context.Context, name string, out io.Writer, rng *RequestRange, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error)

	// Set writes a stream to the file in the filesystem
	// If you pass a tag, the implementation might use that to ensure that the file on the filesystem hasn't been changed since it was read (optional)
	Set(ctx context.Context, name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error)

	// Delete a file from the filesystem
	// If you pass a tag, the implementation might use that to ensure that the file on the filesystem hasn't been changed since it was read (optional)
	Delete(ctx context.Context, name string, tag interface{}) (err error)
}

// FsOptions is the interface for the options for the filesystem
type FsOptions interface {
	// SetOptions the options and validate them
	SetOptions(options map[string]string) (err error)
	// Get returns the value for an option, or the empty string if not set
	Get(key string) (value string)
}

// Base class for filesystems, which contains the key and data path
type fsBase struct {
	keyId     string
	masterKey []byte
	dataPath  string
}

// SetDataPath sets the path where the data is stored (read from the info file)
func (f *fsBase) SetDataPath(path string) {
	f.dataPath = path
}

// GetDataPath returns the path where the data is stored
func (f *fsBase) GetDataPath() string {
	return f.dataPath
}

// SetMasterKey sets the master key (used to encrypt/decrypt files) in the object
func (f *fsBase) SetMasterKey(keyId string, key []byte) {
	f.keyId = keyId
	f.masterKey = key
}

// GetMasterKey returns the master key
func (f *fsBase) GetMasterKey() []byte {
	return f.masterKey
}

// GetKeyId returns the ID of the key used
func (f *fsBase) GetKeyId() string {
	return f.keyId
}
