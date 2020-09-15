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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/utils"

	homedir "github.com/mitchellh/go-homedir"
)

// Register the fs
func init() {
	t := reflect.TypeOf((*Local)(nil)).Elem()
	fsTypes["file"] = t
	fsTypes["local"] = t
}

// Local is the local file system
// This implementation does not rely on tags, as it's assumed that concurrency isn't an issue on a single machine
type Local struct {
	fsBase

	basePath string
	cache    *MetadataCache
	mux      sync.Mutex
}

func (f *Local) InitWithOptionsMap(opts map[string]string, cache *MetadataCache) error {
	// Required keys: "path"
	path := opts["path"]
	if path == "" {
		return errors.New("option 'path' is not defined")
	}

	return f.init(path, cache)
}

func (f *Local) InitWithConnectionString(connection string, cache *MetadataCache) error {
	// Connection string format: "local:<path>" or "file:<path>"
	// Get the path
	path := connection[strings.Index(connection, ":")+1:]

	return f.init(path, cache)
}

func (f *Local) init(path string, cache *MetadataCache) error {
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
	err = utils.EnsureFolder(path)
	if err != nil {
		return err
	}

	f.cache = cache
	f.basePath = path

	return nil
}

func (f *Local) AccountName() string {
	return f.basePath
}

func (f *Local) GetInfoFile() (info *infofile.InfoFile, err error) {
	// Read the file
	data, err := ioutil.ReadFile(f.basePath + "_info.json")
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}

	// Check if the file has any content
	if len(data) == 0 {
		return
	}

	// Parse the JSON data
	info = &infofile.InfoFile{}
	if err = json.Unmarshal(data, info); err != nil {
		info = nil
		return
	}

	// Validate the content
	if err = info.Validate(); err != nil {
		info = nil
		return
	}

	// Set the data path
	f.dataPath = info.DataPath

	return
}

func (f *Local) SetInfoFile(info *infofile.InfoFile) (err error) {
	// Encode the content as JSON
	data, err := json.Marshal(info)
	if err != nil {
		return
	}

	// Write to file
	err = ioutil.WriteFile(f.basePath+"_info.json.tmp", data, 0644)
	if err != nil {
		return
	}

	// Rename to make the write atomic
	err = os.Rename(f.basePath+"_info.json.tmp", f.basePath+"_info.json")
	if err != nil {
		return
	}

	return
}

func (f *Local) Get(ctx context.Context, name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	found = true

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if len(name) > 4 && name[0] != '_' {
		folder = f.dataPath + "/" + name[0:2] + "/" + name[2:4] + "/"
	}

	// Open the file
	file, err := os.Open(f.basePath + folder + name)
	if err != nil {
		if os.IsNotExist(err) {
			found = false
			err = nil
		}
		return
	}
	defer file.Close()

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
	var metadataLength int32
	var metadata *crypto.Metadata
	headerVersion, headerLength, wrappedKey, err := crypto.DecryptFile(ctx, out, file, f.masterKey, func(md *crypto.Metadata, sz int32) bool {
		metadata = md
		metadataLength = sz
		if metadataCb != nil {
			metadataCb(md, sz)
		}
		return true
	})
	// Ignore ErrMetadataOnly so the metadata is still added to cache
	if err != nil && err != crypto.ErrMetadataOnly {
		return
	}

	// Store the metadata in cache
	// Adding a lock here to prevent the case when adding this key causes the eviction of another one that's in use
	f.mux.Lock()
	f.cache.Add(name, headerVersion, headerLength, wrappedKey, metadataLength, metadata)
	f.mux.Unlock()

	return
}

func (f *Local) GetWithRange(ctx context.Context, name string, out io.Writer, rng *RequestRange, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if len(name) > 4 && name[0] != '_' {
		folder = f.dataPath + "/" + name[0:2] + "/" + name[2:4] + "/"
	}

	found = true

	// Open the file
	file, err := os.Open(f.basePath + folder + name)
	if err != nil {
		if os.IsNotExist(err) {
			found = false
			err = nil
		}
		return
	}
	defer file.Close()

	// Check if the file has any content
	stat, err := file.Stat()
	if err != nil {
		return
	}
	if stat.Size() == 0 {
		found = false
		return
	}

	// Look up the file's metadata in the cache
	f.mux.Lock()
	headerVersion, headerLength, wrappedKey, metadataLength, metadata := f.cache.Get(name)
	if headerVersion == 0 || headerLength < 1 || wrappedKey == nil || len(wrappedKey) < 1 {
		// Need to read the metadata and cache it
		// For that, we need to read the header and the first package, which are at most 64kb + (32+256) bytes
		read := make([]byte, 64*1024+32+256)
		var n int
		n, err = io.ReadFull(file, read)
		// Ignore ErrUnexpectedEOF which means that the file is shorter than what we were looking for
		if err != nil && err != io.ErrUnexpectedEOF {
			f.mux.Unlock()
			return
		}

		// Get a buffer and discard the bytes that were not read (and filled with 0's)
		buf := bytes.NewBuffer(read)
		buf.Truncate(n)

		// Decrypt the data
		headerVersion, headerLength, wrappedKey, err = crypto.DecryptFile(ctx, nil, buf, f.masterKey, func(md *crypto.Metadata, sz int32) bool {
			metadata = md
			metadataLength = sz
			return false
		})
		if err != nil && err != crypto.ErrMetadataOnly {
			f.mux.Unlock()
			return
		}

		// Store the metadata in cache
		f.cache.Add(name, headerVersion, headerLength, wrappedKey, metadataLength, metadata)
	}
	f.mux.Unlock()

	// Add the offsets to the range object and set the file size (it's guaranteed it's set, or we wouldn't have a range request)
	rng.HeaderOffset = int64(headerLength)
	rng.MetadataOffset = int64(metadataLength)
	rng.SetFileSize(metadata.Size)

	// Send the metadata to the callback
	if metadataCb != nil {
		metadataCb(metadata, metadataLength)
	}

	// Move the file pointer to the beginning of the range
	_, err = file.Seek(rng.StartBytes(), 0)
	if err != nil {
		return
	}

	// Create a pipe so we can stop reading after we read a certain amount of data
	pr, pw := io.Pipe()
	go func() {
		// Read only the required packages
		_, err := io.CopyN(pw, file, rng.LengthBytes())
		if err != nil && err != io.EOF {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()

	// Close the pipe if the context is canceled
	go func() {
		<-ctx.Done()
		pw.Close()
	}()

	// Decrypt the data
	err = crypto.DecryptPackages(ctx, out, pr, headerVersion, wrappedKey, f.masterKey, rng.StartPackage(), uint32(rng.SkipBeginning()), rng.Length, nil)
	if err != nil {
		return
	}

	return
}

func (f *Local) Set(ctx context.Context, name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if len(name) > 4 && name[0] != '_' {
		folder = f.dataPath + "/" + name[0:2] + "/" + name[2:4] + "/"

		// Ensure the folder exists
		err = utils.EnsureFolder(f.basePath + folder)
		if err != nil {
			return
		}
	}

	// Create a temporary file; we'll rename it later
	file, err := os.Create(f.basePath + folder + name + ".tmp")
	if err != nil {
		return nil, err
	}

	// Encrypt the data and write it to file
	err = crypto.EncryptFile(file, utils.ReaderFuncWithContext(ctx, in), f.masterKey, metadata)
	if err != nil {
		file.Close()
		return nil, err
	}
	file.Close()

	// Rename the file
	err = os.Rename(f.basePath+folder+name+".tmp", f.basePath+folder+name)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (f *Local) Delete(ctx context.Context, name string, tag interface{}) (err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if len(name) > 4 && name[0] != '_' {
		folder = f.dataPath + "/" + name[0:2] + "/" + name[2:4] + "/"
	}

	// Note that we're not removing the data from the cache, as it's identified by the UUID which will not be used by other files

	// Delete the file
	err = os.Remove(f.basePath + folder + name)
	return
}
