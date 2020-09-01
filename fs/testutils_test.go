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
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/keys"

	"github.com/stretchr/testify/assert"
)

// Performs tests for a store object, already initialized
type testFs struct {
	t     *testing.T
	store Fs

	info      *infofile.InfoFile
	files     map[string][]byte
	checksums map[string][]byte
}

// Starts the test
func (s *testFs) Run() {
	// Load fixtures
	s.loadFixtures()

	// Initialize repo
	s.testGetInfoFileNotInitialized()
	s.testSetInfoFile()
	s.testGetInfoFile()

	// Derive and set master key
	masterKey, keyId, _, err := keys.GetMasterKeyWithPassphrase(s.info, "hello world")
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	s.store.SetMasterKey(keyId, masterKey)

	// Set and retrieve encrypted files
	s.testSet()
}

// Loads fixtures
func (s *testFs) loadFixtures() {
	// Info file
	s.info = staticInfoFile()

	// Sample files
	s.files = map[string][]byte{}
	read, err := ioutil.ReadFile(path.Join("..", "tests", "fixtures", "divinacommedia.txt"))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	s.files["divinacommedia.txt"] = read
	read, err = ioutil.ReadFile(path.Join("..", "tests", "fixtures", "kitera-dent-BIj4LObC6es-unsplash.jpg"))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	s.files["kitera-dent-BIj4LObC6es-unsplash.jpg"] = read

	// Checksums
	s.checksums = map[string][]byte{}
	checksumsFile, err := ioutil.ReadFile(path.Join("..", "tests", "fixtures", "checksums.json"))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	err = json.Unmarshal(checksumsFile, &s.checksums)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
}

// Get info file, but repo is not initialized
func (s *testFs) testGetInfoFileNotInitialized() {
	// Must have no error but nil info, meaning the file was not found
	info, err := s.store.GetInfoFile()
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.Nil(s.t, info)
}

// Set info file
func (s *testFs) testSetInfoFile() {
	err := s.store.SetInfoFile(s.info)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
}

// Get info file
func (s *testFs) testGetInfoFile() {
	info, err := s.store.GetInfoFile()
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.NotNil(s.t, info)
	assert.True(s.t, reflect.DeepEqual(info, s.info))
}

// Store an encrypted file
func (s *testFs) testSet() {
	// Store text file
	in := bytes.NewReader(s.files["divinacommedia.txt"])
	metadata := &crypto.Metadata{
		Name:        "divinacommedia.txt",
		ContentType: "text/plain",
		Size:        int64(len(s.files["divinacommedia.txt"])),
	}
	_, err := s.store.Set(context.Background(), "divinacommedia.txt", in, nil, metadata)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}

	// Error: empty name
	_, err = s.store.Set(context.Background(), "", in, nil, metadata)
	if !assert.Error(s.t, err) {
		s.t.FailNow()
	}
}

// Return the info file object
func staticInfoFile() *infofile.InfoFile {
	// Create an info file with fixed data
	// Passphrase is "hello world"
	masterKey, _ := base64.StdEncoding.DecodeString("QGRFye4ebTr6U85Ja8V5d0ciZfDLXFz8gTjpqj+b6l1/N8q6oYC2hA==")
	salt, _ := base64.StdEncoding.DecodeString("Id5gT91MIeqMG7Pc1UFc8Q==")
	confirmationHash, _ := base64.StdEncoding.DecodeString("WL539+dtEvM5VDQ9LtCepF7nguCZMEzISvnFMK4UIeE=")
	return &infofile.InfoFile{
		App:      "prvt",
		Version:  4,
		DataPath: "data",
		RepoId:   "26346eac-6526-4093-a7b8-4640d4fa2f32",
		Keys: []infofile.InfoFileKey{
			{
				MasterKey:        masterKey,
				Salt:             salt,
				ConfirmationHash: confirmationHash,
			},
		},
	}
}
