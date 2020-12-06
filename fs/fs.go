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
	"github.com/ItalyPaleAle/prvt/fs/fsutils"
	"github.com/ItalyPaleAle/prvt/infofile"
)

var fsTypes = map[string]reflect.Type{}

// GetFsOptions returns the list of options for a specific fs
func GetFsOptions(name string) *FsOptionsList {
	// Get the store object using some reflection magic
	fsTyp, ok := fsTypes[name]
	if !ok || fsTyp == nil {
		return nil
	}
	store := reflect.New(fsTyp).Interface().(Fs)

	// Return options
	return store.OptionsList()
}

// GetAllFsOptions returns the list of options for all fs
func GetAllFsOptions() map[string]*FsOptionsList {
	res := make(map[string]*FsOptionsList)
	for _, fsTyp := range fsTypes {
		// Get the store object using some reflection magic
		store := reflect.New(fsTyp).Interface().(Fs)
		// Add this only for the canonical name
		k := store.FSName()
		if res[k] == nil {
			res[k] = store.OptionsList()
		}
	}

	// Return options
	return res
}

// GetWithOptionsMap returns a store given the options map
// The dictionary must have a key "type" for the type of fs to use
func GetWithOptionsMap(opts map[string]string) (store Fs, err error) {
	// Init the cache
	cache := &fsutils.MetadataCache{}
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
	err = store.InitWithOptionsMap(opts, cache)

	return
}

// GetWithConnectionString returns a store for the given connection string
func GetWithConnectionString(connection string) (store Fs, err error) {
	// Init the cache
	cache := &fsutils.MetadataCache{}
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
	// OptionsList returns the list of options (dictionary keys)
	OptionsList() *FsOptionsList

	// InitWithOptionsMap inits the object by passing an options map
	InitWithOptionsMap(opts map[string]string, cache *fsutils.MetadataCache) error

	// InitWithConnectionString inits the object by passing a connection string and the cache object
	InitWithConnectionString(connection string, cache *fsutils.MetadataCache) error

	// FSName returns the identifier of this fs
	FSName() string

	// AccountName returns a string that can be used to identify this account
	AccountName() string

	// SetMasterKey sets the master key (used to encrypt/decrypt files) in the object
	SetMasterKey(keyId string, key []byte)

	// GetMasterKey returns the master key
	GetMasterKey() []byte

	// GetKeyId returns the ID of the key used
	GetKeyId() string

	// RawGet gets a file from the store as-is, without processing or decrypting its content
	RawGet(ctx context.Context, name string, out io.Writer, start int64, count int64) (found bool, tag interface{}, err error)

	// RawSet sets a file in the store as-is, without encrypting its content
	RawSet(ctx context.Context, name string, in io.Reader, tag interface{}) (tagOut interface{}, err error)

	// GetInfoFile returns the contents of the info file
	GetInfoFile() (info *infofile.InfoFile, err error)

	// SetInfoFile stores the info file
	SetInfoFile(info *infofile.InfoFile) (err error)

	// Get returns a stream to a file in the filesystem
	// It also returns a tag (which might be empty) that should be passed to the Set method if you want to subsequentially update the contents of the file
	Get(ctx context.Context, name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error)

	// GetWithRange is like Get, but accepts a custom range
	GetWithRange(ctx context.Context, name string, out io.Writer, rng *fsutils.RequestRange, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error)

	// Set writes a stream to the file in the filesystem
	// If you pass a tag, the implementation might use that to ensure that the file on the filesystem hasn't been changed since it was read (optional)
	Set(ctx context.Context, name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error)

	// Delete a file from the filesystem
	// If you pass a tag, the implementation might use that to ensure that the file on the filesystem hasn't been changed since it was read (optional)
	Delete(ctx context.Context, name string, tag interface{}) (err error)

	// AcquireLock acquires an exclusive lock, to help making sure that no other process is using the same repository
	AcquireLock(ctx context.Context) (err error)

	// ReleaseLock releases the lock
	ReleaseLock(ctx context.Context) (err error)

	// BreakLock breaks the lock that another process might be holding
	BreakLock(ctx context.Context) (err error)
}

// Base class for filesystems, which contains the key and data path
type fsBase struct {
	keyId     string
	masterKey []byte
	dataPath  string
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

// Individual for the filesystem
type FsOption struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
	Private     bool   `json:"private,omitempty"`
}

// List of options for each filesystem
type FsOptionsList struct {
	Label    string     `json:"label"`
	Required []FsOption `json:"required"`
	Optional []FsOption `json:"optional"`
}
