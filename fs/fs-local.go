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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	err = utils.EnsureFolder(path)
	if err != nil {
		return err
	}

	f.basePath = path

	return nil
}

func (f *Local) SetMasterKey(key []byte) {
	f.masterKey = key
}

func (f *Local) GetInfoFile() (info *InfoFile, err error) {
	// Read the file
	data, err := ioutil.ReadFile(f.basePath + "_info.json")
	if err != nil {
		return
	}

	// Check if the file has any content
	if len(data) == 0 {
		return
	}

	// Parse the JSON data
	info = &InfoFile{}
	if err = json.Unmarshal(data, info); err != nil {
		info = nil
		return
	}

	// Validate the content
	if err = InfoValidate(info); err != nil {
		info = nil
		return
	}

	return
}

func (f *Local) SetInfoFile(info *InfoFile) (err error) {
	// Encode the content as JSON
	data, err := json.Marshal(info)
	if err != nil {
		return
	}

	// Write to file
	err = ioutil.WriteFile(f.basePath+"_info.json", data, 0644)
	if err != nil {
		return
	}

	return
}

func (f *Local) Get(name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	found = true

	// If the file doesn't start with _, it lives in a sub-folder
	folder := ""
	if len(name) > 4 && name[0] != '_' {
		folder = name[0:2] + "/" + name[2:4] + "/"
	}

	// Open the file
	file, err := os.Open(f.basePath + folder + name)
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
	err = crypto.DecryptFile(out, file, f.masterKey, metadataCb)
	if err != nil {
		return
	}

	return
}

func (f *Local) Set(name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, place it in a sub-folder
	folder := ""
	if len(name) > 4 && name[0] != '_' {
		folder = name[0:2] + "/" + name[2:4] + "/"

		// Ensure the folder exists
		err = utils.EnsureFolder(f.basePath + folder)
		if err != nil {
			return
		}
	}

	// Create the file
	file, err := os.Create(f.basePath + folder + name)
	if err != nil {
		return nil, err
	}

	// Encrypt the data and write it to file
	err = crypto.EncryptFile(file, in, f.masterKey, metadata)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (f *Local) Delete(name string, tag interface{}) (err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder
	folder := ""
	if len(name) > 4 && name[0] != '_' {
		folder = name[0:2] + "/" + name[2:4] + "/"
	}

	// Delete the file
	err = os.Remove(f.basePath + folder + name)
	return
}
