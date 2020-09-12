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

package tests

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/stretchr/testify/assert"
)

// RunPreviousVersions runs the sequence of tests that ensure that prvt can work with repositories created with previous versions
func (s *funcTestSuite) RunPreviousVersions(t *testing.T) {
	// Skip when running a short test
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// v0.2 has the same data structures as v0.1
	t.Run("prvt 0.2", s.previousVersion_0_2)
	t.Run("prvt 0.3", s.previousVersion_0_3)
	t.Run("prvt 0.4", s.previousVersion_0_4)
	t.Run("prvt 0.5", s.previousVersion_0_5)
}

// Tests for working with repositories created by version 0.2
func (s *funcTestSuite) previousVersion_0_2(t *testing.T) {
	// Start the server
	s.promptPwd.SetPasswords("hello world")
	path := filepath.Join(s.fixtures, "previous-versions", "v0.2")
	close := s.startServer(t, "--store", "local:"+path)

	// Should be able to list files
	list, err := s.listRequest("/")
	if err != nil {
		close()
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 1)
	assert.Equal(t, "pg1000.txt", list[0].Path)
	assert.Empty(t, list[0].MimeType)
	assert.Empty(t, list[0].Date)

	// Should be able to request the file
	s.previousVersionRequestFile(t, list[0].FileId)

	// Stop the server
	close()

	// Commands such as "prvt repo key ls" requires a newer info file
	runCmd(t,
		[]string{"repo", "key", "ls", "--store", "local:" + path},
		func(err error) {
			if !strings.HasPrefix(err.Error(), "[Error] Repository needs to be upgraded") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// Upgrade the repo
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"repo", "upgrade", "--store", "local:" + path},
		nil,
		func(stdout string) {
			assert.Equal(t, "Repository upgraded\n", stdout)
		},
		nil,
	)

	// Check that the repo has been upgraded
	s.previousVersionCheckInfoFile(t, path, 1)

	// Try unlocking the repo again
	s.promptPwd.SetPasswords("hello world")
	close = s.startServer(t, "--store", "local:"+path)

	// Should be able to list files
	newList, err := s.listRequest("/")
	assert.NoError(t, err)
	assert.Len(t, newList, 1)
	assert.True(t, reflect.DeepEqual(list, newList))

	// Stop the server
	close()
}

// Tests for working with repositories created by version 0.3
func (s *funcTestSuite) previousVersion_0_3(t *testing.T) {
	// Start the server
	s.promptPwd.SetPasswords("hello world")
	path := filepath.Join(s.fixtures, "previous-versions", "v0.3")
	close := s.startServer(t, "--store", "local:"+path)

	// Should be able to list files
	list, err := s.listRequest("/")
	if err != nil {
		close()
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 1)
	assert.Equal(t, "pg1000.txt", list[0].Path)
	assert.Empty(t, list[0].MimeType)
	assert.Empty(t, list[0].Date)

	// Should be able to request the file
	s.previousVersionRequestFile(t, list[0].FileId)

	// Stop the server
	close()

	// Commands such as "prvt add" requires a newer info file (3+)
	s.promptPwd.SetPasswords("hello world")
	addPath := filepath.Join(s.fixtures, "photos")
	runCmd(t,
		[]string{"add", addPath, "--destination", "/text", "--store", "local:" + path},
		func(err error) {
			if !strings.HasPrefix(err.Error(), "[Error] Repository needs to be upgraded") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// Upgrade the repo
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"repo", "upgrade", "--store", "local:" + path},
		nil,
		func(stdout string) {
			assert.Equal(t, "Repository upgraded\n", stdout)
		},
		nil,
	)

	// Check that the repo has been upgraded
	s.previousVersionCheckInfoFile(t, path, 2)

	// Try unlocking the repo again
	s.promptPwd.SetPasswords("hello world")
	close = s.startServer(t, "--store", "local:"+path)

	// Should be able to list files
	newList, err := s.listRequest("/")
	assert.NoError(t, err)
	assert.Len(t, newList, 1)
	assert.True(t, reflect.DeepEqual(list, newList))

	// Stop the server
	close()
}

// Tests for working with repositories created by version 0.4
func (s *funcTestSuite) previousVersion_0_4(t *testing.T) {
	// Start the server
	s.promptPwd.SetPasswords("hello world")
	path := filepath.Join(s.fixtures, "previous-versions", "v0.4")
	close := s.startServer(t, "--store", "local:"+path)

	// Should be able to list files
	list, err := s.listRequest("/")
	if err != nil {
		close()
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 1)
	assert.Equal(t, "pg1000.txt", list[0].Path)
	assert.NotEmpty(t, list[0].MimeType)
	assert.NotEmpty(t, list[0].Date)

	// Should be able to request the file
	s.previousVersionRequestFile(t, list[0].FileId)

	// Stop the server
	close()

	// Upgrade the repo
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"repo", "upgrade", "--store", "local:" + path},
		nil,
		func(stdout string) {
			assert.Equal(t, "Repository upgraded\n", stdout)
		},
		nil,
	)

	// Check that the repo has been upgraded
	s.previousVersionCheckInfoFile(t, path, 3)

	// Try unlocking the repo again
	s.promptPwd.SetPasswords("hello world")
	close = s.startServer(t, "--store", "local:"+path)

	// Should be able to list files
	newList, err := s.listRequest("/")
	assert.NoError(t, err)
	assert.Len(t, newList, 1)
	assert.True(t, reflect.DeepEqual(list, newList))

	// Stop the server
	close()
}

// Tests for working with repositories created by version 0.5
func (s *funcTestSuite) previousVersion_0_5(t *testing.T) {
	// Start the server
	s.promptPwd.SetPasswords("hello world")
	path := filepath.Join(s.fixtures, "previous-versions", "v0.5")
	close := s.startServer(t, "--store", "local:"+path)

	// Should be able to list files
	list, err := s.listRequest("/")
	if err != nil {
		close()
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 1)
	assert.Equal(t, "pg1000.txt", list[0].Path)
	assert.NotEmpty(t, list[0].MimeType)
	assert.NotEmpty(t, list[0].Date)

	// Should be able to request the file
	s.previousVersionRequestFile(t, list[0].FileId)

	// Stop the server
	close()

	// Upgrade the repo
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"repo", "upgrade", "--store", "local:" + path},
		nil,
		func(stdout string) {
			assert.Equal(t, "Repository upgraded\n", stdout)
		},
		nil,
	)

	// Check that the repo has been upgraded
	s.previousVersionCheckInfoFile(t, path, 4)

	// Try unlocking the repo again
	s.promptPwd.SetPasswords("hello world")
	close = s.startServer(t, "--store", "local:"+path)

	// Should be able to list files
	newList, err := s.listRequest("/")
	assert.NoError(t, err)
	assert.Len(t, newList, 1)
	assert.True(t, reflect.DeepEqual(list, newList))

	// Stop the server
	close()
}

func (s *funcTestSuite) previousVersionRequestFile(t *testing.T, fileId string) {
	t.Helper()

	// Load the file from disk
	content, err := ioutil.ReadFile(filepath.Join(s.fixtures, "divinacommedia.txt"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Retrieve the file in full and compare the data
	res, err := s.client.Get(s.serverAddr + "/file/" + fileId)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Equal(t, content, read)

	// Request part of the file only
	req, err := http.NewRequest("GET", s.serverAddr+"/file/"+fileId, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	req.Header.Add("Range", "bytes=65409-65485")
	res, err = s.client.Do(req)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Equal(t, "77", res.Header.Get("Content-Length"))
	assert.Equal(t, content[65409:65486], read)
}

func (s *funcTestSuite) previousVersionCheckInfoFile(t *testing.T, path string, startingVer int) {
	t.Helper()

	// Read the info file
	data, err := ioutil.ReadFile(filepath.Join(path, "_info.json"))
	if err != nil {
		t.Error(err)
		return
	}
	info := &infofile.InfoFile{}
	err = json.Unmarshal(data, info)
	if err != nil {
		t.Error(err)
		return
	}

	// Validate the content
	err = info.Validate()
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, uint16(5), info.Version)
	assert.Len(t, info.Keys, 1)
	assert.NotEmpty(t, info.RepoId)

	// Key DataPath was added in info file v2, so the value remains empty if upgrading from v1
	if startingVer == 1 {
		assert.Equal(t, "", info.DataPath)
	} else {
		assert.Equal(t, "data", info.DataPath)
	}
}
