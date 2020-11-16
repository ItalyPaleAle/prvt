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
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ItalyPaleAle/prvt/infofile"

	"github.com/stretchr/testify/assert"
)

// RunCLI runs the sequence of tests for CLI commands; must be run before the server tests
func (s *funcTestSuite) RunCLI(t *testing.T) {
	t.Run("init repo", s.cmdRepoInit)
	t.Run("key management", s.cmdRepoKey)
	t.Run("add files", s.cmdAdd)
	t.Run("list and remove files", s.cmdLsAndRm)
	t.Run("repo info", s.cmdRepoInfo)
}

func (s *funcTestSuite) cmdRepoInit(t *testing.T) {
	// Init the first repo with a passphrase
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"repo", "init", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			if !strings.HasPrefix(stdout, "Initialized new repository in the store local:") {
				t.Fatal("output does not match prefix", stdout)
			}
		},
		nil,
	)

	// Error: empty passphrase
	s.promptPwd.SetPasswords()
	runCmd(t,
		[]string{"repo", "init", "--store", "local:" + s.dirs[1]},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] Error adding the key\n^D\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// Init the second repo with a GPG key
	runCmd(t,
		[]string{"repo", "init", "--gpg", s.gpgKeyId, "--store", "local:" + s.dirs[1]},
		nil,
		func(stdout string) {
			if !strings.HasPrefix(stdout, "Initialized new repository in the store local:") {
				t.Fatal("output does not match prefix", stdout)
			}
		},
		nil,
	)

	// Error: repository already initialized
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"repo", "init", "--store", "local:" + s.dirs[1]},
		func(err error) {
			if !assert.EqualError(t, err, "[Fatal error] Error initializing repository\nA repository is already initialized in this store\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// Check the info file for the first repo
	{
		read, err := ioutil.ReadFile(s.dirs[0] + "/_info.json")
		if !assert.NoError(t, err) {
			t.Fatal(err)
		}
		if !assert.NotEmpty(t, read) {
			t.Fatal("info file is empty")
		}
		info := &infofile.InfoFile{}
		err = json.Unmarshal(read, info)
		assert.NoError(t, err)
		assert.Equal(t, "prvt", info.App)
		assert.Equal(t, uint16(5), info.Version)
		assert.NotEmpty(t, info.RepoId)
		assert.Len(t, info.Keys, 1)
		assert.Len(t, info.Keys[0].ConfirmationHash, 32)
		assert.Len(t, info.Keys[0].MasterKey, 40)
		assert.Len(t, info.Keys[0].Salt, 16)
		assert.Len(t, info.Keys[0].GPGKey, 0)
		s.repoIds[0] = info.RepoId
	}

	// Check the info file for the second repo
	{
		read, err := ioutil.ReadFile(s.dirs[1] + "/_info.json")
		if !assert.NoError(t, err) {
			t.Fatal(err)
		}
		if !assert.NotEmpty(t, read) {
			t.Fatal("info file is empty")
		}
		info := &infofile.InfoFile{}
		err = json.Unmarshal(read, info)
		assert.NoError(t, err)
		assert.Equal(t, "prvt", info.App)
		assert.Equal(t, uint16(5), info.Version)
		assert.NotEmpty(t, info.RepoId)
		assert.Len(t, info.Keys, 1)
		assert.Len(t, info.Keys[0].ConfirmationHash, 0)
		assert.True(t, len(info.Keys[0].MasterKey) > 100)
		assert.Len(t, info.Keys[0].Salt, 0)
		assert.True(t, len(info.Keys[0].GPGKey) > 0)
		s.repoIds[1] = info.RepoId
	}
}

func (s *funcTestSuite) cmdRepoKey(t *testing.T) {
	// Skip when running a short test
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Test the passphrase on the first repo
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"repo", "key", "test", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			if !strings.HasPrefix(stdout, "Repository unlocked using key with ID: p:") {
				t.Fatal("output does not match prefix", stdout)
			}
		},
		nil,
	)

	// Test the GPG key with the second repo
	// Should pick up the GPG key automatically
	runCmd(t,
		[]string{"repo", "key", "test", "--store", "local:" + s.dirs[1]},
		nil,
		func(stdout string) {
			if stdout != "Repository unlocked using key with ID: "+s.gpgKeyId+"\n" {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)

	// Error: invalid passphrase
	s.promptPwd.SetPasswords("no match")
	runCmd(t,
		[]string{"repo", "key", "test", "--store", "local:" + s.dirs[0]},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] Cannot unlock the repository\nInvalid passphrase\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// List keys for the first repo
	runCmd(t,
		[]string{"repo", "key", "ls", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			match := "KEY TYPE    | KEY ID\n------------|--------------------\nPassphrase  | p:"
			if !strings.HasPrefix(stdout, match) {
				t.Fatal("output does not match prefix", stdout)
			}
		},
		nil,
	)

	// List keys for the second repo
	runCmd(t,
		[]string{"repo", "key", "ls", "--store", "local:" + s.dirs[1]},
		nil,
		func(stdout string) {
			match := "KEY TYPE    | KEY ID\n------------|--------------------\nGPG Key     | " + s.gpgKeyId
			if s.gpgKeyUser != "" {
				match += "  (" + s.gpgKeyUser + ")"
			}
			if !strings.HasPrefix(stdout, match) {
				t.Fatal("output does not match prefix", stdout)
			}
		},
		nil,
	)

	// Add a passphrase to the second repo
	s.promptPwd.SetPasswords("hi")
	var passphraseId string
	runCmd(t,
		[]string{"repo", "key", "add", "--store", "local:" + s.dirs[1]},
		nil,
		func(stdout string) {
			i := strings.Index(stdout, "Key added with id: p:")
			if i < 0 {
				t.Fatal("output does not contain required string\n", stdout)
			}
			start := strings.Index(stdout[i:], "p:")
			// Exclude last character which is a newline
			passphraseId = stdout[(i + start):(len(stdout) - 1)]
		},
		nil,
	)

	// Check the info file to ensure it's updated
	{
		read, err := ioutil.ReadFile(s.dirs[1] + "/_info.json")
		if !assert.NoError(t, err) {
			t.Fatal(err)
		}
		if !assert.NotEmpty(t, read) {
			t.Fatal("info file is empty")
		}
		info := &infofile.InfoFile{}
		err = json.Unmarshal(read, info)
		assert.NoError(t, err)
		assert.Equal(t, "prvt", info.App)
		assert.Equal(t, uint16(5), info.Version)
		assert.NotEmpty(t, info.RepoId)
		assert.Len(t, info.Keys, 2)
		assert.Len(t, info.Keys[0].ConfirmationHash, 0)
		assert.True(t, len(info.Keys[0].MasterKey) > 100)
		assert.Len(t, info.Keys[0].Salt, 0)
		assert.True(t, len(info.Keys[0].GPGKey) > 0)
		assert.Len(t, info.Keys[1].ConfirmationHash, 32)
		assert.Len(t, info.Keys[1].MasterKey, 40)
		assert.Len(t, info.Keys[1].Salt, 16)
		assert.Len(t, info.Keys[1].GPGKey, 0)
	}

	// Key should be added
	runCmd(t,
		[]string{"repo", "key", "ls", "--store", "local:" + s.dirs[1]},
		nil,
		func(stdout string) {
			if !strings.Contains(stdout, "GPG Key     | "+s.gpgKeyId) {
				t.Fatal("output does not contain required string\n", stdout)
			}
			if !strings.Contains(stdout, "Passphrase  | p:") {
				t.Fatal("output does not contain required string\n", stdout)
			}
		},
		nil,
	)

	// Error: cannot remove the same key we're using to unlock the repo
	runCmd(t,
		[]string{"repo", "key", "rm", "--key", s.gpgKeyId, "--store", "local:" + s.dirs[1]},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] Invalid key ID\nYou cannot remove the same key you're using to unlock the repository\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// Error: cannot add the same passphrase twice
	s.promptPwd.SetPasswords("hi")
	runCmd(t,
		[]string{"repo", "key", "add", "--store", "local:" + s.dirs[1]},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] Key already added\nThis passphrase has already been added to the repository\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// Error: the same GPG key has already been added
	runCmd(t,
		[]string{"repo", "key", "add", "--store", "local:" + s.dirs[1], "--gpg", s.gpgKeyId},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] A GPG key with the same ID is already authorized to unlock this repository\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// Remove the passphrase that was added
	runCmd(t,
		[]string{"repo", "key", "rm", "--key", passphraseId, "--store", "local:" + s.dirs[1]},
		nil,
		func(stdout string) {
			if !strings.Contains(stdout, "Key removed") {
				t.Fatal("output does not contain required string\n", stdout)
			}
		},
		nil,
	)

	// Error: cannot remove the last key
	runCmd(t,
		[]string{"repo", "key", "rm", "--key", s.gpgKeyId, "--store", "local:" + s.dirs[1]},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] Cannot remove the only key\nThis repository has only one key, which cannot be removed\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)
}

func (s *funcTestSuite) cmdAdd(t *testing.T) {
	var addPath string

	// Add a file to the first repo
	s.promptPwd.SetPasswords("hello world")
	addPath = filepath.Join(s.fixtures, "divinacommedia.txt")
	runCmd(t,
		[]string{"add", addPath, "--destination", "/text", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			if stdout != "Added: /text/divinacommedia.txt\n" {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)
	checkRepoDirectory(t, s.dirs[0], 1)

	// Add multiple files to the repo
	s.promptPwd.SetPasswords("hello world")
	addPath = filepath.Join(s.fixtures, "photos")
	runCmd(t,
		[]string{"add", addPath, "--destination", "/", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := []string{
				"",
				"Added: /photos/elton-sa-_3g60mG4N80-unsplash.jpg",
				"Added: /photos/joshua-woroniecki-dyEaBD5uiio-unsplash.jpg",
				"Added: /photos/leigh-williams-CCABYukxjHs-unsplash.jpg",
				"Added: /photos/nathan-thomassin-E6xV-UxrKSg-unsplash.jpg",
				"Added: /photos/partha-narasimhan-kT5Syi2Ll3w-unsplash.jpg",
			}
			actual := strings.Split(stdout, "\n")
			sort.Strings(actual)
			if !reflect.DeepEqual(expected, actual) {
				t.Error("output does not match", stdout)
			}
		},
		nil,
	)
	checkRepoDirectory(t, s.dirs[0], 6)

	// Add multiple files, including existing ones
	s.promptPwd.SetPasswords("hello world")
	addPath1 := filepath.Join(s.fixtures, "photos")
	addPath2 := filepath.Join(s.fixtures, "short.txt")
	runCmd(t,
		[]string{"add", addPath1, addPath2, "--destination", "/", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := []string{
				"",
				"Added: /short.txt",
				"Skipping existing file: /photos/elton-sa-_3g60mG4N80-unsplash.jpg",
				"Skipping existing file: /photos/joshua-woroniecki-dyEaBD5uiio-unsplash.jpg",
				"Skipping existing file: /photos/leigh-williams-CCABYukxjHs-unsplash.jpg",
				"Skipping existing file: /photos/nathan-thomassin-E6xV-UxrKSg-unsplash.jpg",
				"Skipping existing file: /photos/partha-narasimhan-kT5Syi2Ll3w-unsplash.jpg",
			}
			actual := strings.Split(stdout, "\n")
			sort.Strings(actual)
			if !reflect.DeepEqual(expected, actual) {
				t.Error("output does not match", stdout)
			}
		},
		nil,
	)
	checkRepoDirectory(t, s.dirs[0], 7)

	// File does not exist
	s.promptPwd.SetPasswords("hello world")
	addPath = filepath.Join(s.fixtures, "notfound.txt")
	runCmd(t,
		[]string{"add", addPath, "--destination", "/", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			if stdout != "Error adding file '/notfound.txt': target does not exist\n" {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)
	checkRepoDirectory(t, s.dirs[0], 7)

	// Error: repository is not initialized
	s.promptPwd.SetPasswords("hello world")
	addPath = filepath.Join(s.fixtures, "divinacommedia.txt")
	runCmd(t,
		[]string{"add", addPath, "--destination", "/text", "--store", "local:foo"},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] Repository is not initialized\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)
}

func (s *funcTestSuite) cmdLsAndRm(t *testing.T) {
	// List files and folders in the root folder
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"ls", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := strings.Join([]string{
				"photos/",
				"text/",
				"short.txt",
				"",
			}, "\n")
			if stdout != expected {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)

	// Same but with explicit root folder
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"ls", "/", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := strings.Join([]string{
				"photos/",
				"text/",
				"short.txt",
				"",
			}, "\n")
			if stdout != expected {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)

	// List a sub-directory
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"ls", "/photos", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := strings.Join([]string{
				"elton-sa-_3g60mG4N80-unsplash.jpg",
				"joshua-woroniecki-dyEaBD5uiio-unsplash.jpg",
				"leigh-williams-CCABYukxjHs-unsplash.jpg",
				"nathan-thomassin-E6xV-UxrKSg-unsplash.jpg",
				"partha-narasimhan-kT5Syi2Ll3w-unsplash.jpg",
				"",
			}, "\n")
			if stdout != expected {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)

	// Run on empty repository
	runCmd(t,
		[]string{"ls", "--store", "local:" + s.dirs[1]},
		nil,
		func(stdout string) {
			if stdout != "" {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)

	// Remove a single file
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"rm", "/photos/elton-sa-_3g60mG4N80-unsplash.jpg", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := strings.Join([]string{
				"Removed: /photos/elton-sa-_3g60mG4N80-unsplash.jpg",
				"",
			}, "\n")
			if stdout != expected {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)
	checkRepoDirectory(t, s.dirs[0], 6)

	// List the directory again
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"ls", "/photos", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := strings.Join([]string{
				"joshua-woroniecki-dyEaBD5uiio-unsplash.jpg",
				"leigh-williams-CCABYukxjHs-unsplash.jpg",
				"nathan-thomassin-E6xV-UxrKSg-unsplash.jpg",
				"partha-narasimhan-kT5Syi2Ll3w-unsplash.jpg",
				"",
			}, "\n")
			if stdout != expected {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)

	// Remove various files, including a not found one
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"rm", "/photos/*", "/short.txt", "/notfound.txt", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := []string{
				"",
				"Not found: /notfound.txt",
				"Removed: /photos/joshua-woroniecki-dyEaBD5uiio-unsplash.jpg",
				"Removed: /photos/leigh-williams-CCABYukxjHs-unsplash.jpg",
				"Removed: /photos/nathan-thomassin-E6xV-UxrKSg-unsplash.jpg",
				"Removed: /photos/partha-narasimhan-kT5Syi2Ll3w-unsplash.jpg",
				"Removed: /short.txt",
			}
			actual := strings.Split(stdout, "\n")
			sort.Strings(actual)
			if !reflect.DeepEqual(expected, actual) {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)
	checkRepoDirectory(t, s.dirs[0], 1)

	// File must start with /
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"rm", "divinacommedia.txt", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			if stdout != "Internal error removing path 'divinacommedia.txt': Error while removing path from index: path must start with /\n" {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)

	// Error: repository is not initialized (rm)
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"rm", "/divinacommedia.txt", "--store", "local:foo"},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] Repository is not initialized\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)

	// Error: repository is not initialized (ls)
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"ls", "--store", "local:foo"},
		func(err error) {
			if !assert.EqualError(t, err, "[Error] Repository is not initialized\n") {
				t.Fatal("error does not match", err)
			}
		},
		nil,
		nil,
	)
}

func (s *funcTestSuite) cmdRepoInfo(t *testing.T) {
	storePath, err := filepath.Abs(s.dirs[0])
	if err != nil {
		t.Fatal(err)
		return
	}

	// Test repo info on a locked repo
	runCmd(t,
		[]string{"repo", "info", "--store", "local:" + s.dirs[0], "--no-unlock"},
		nil,
		func(stdout string) {
			expected := `Repository ID:       ` + s.repoIds[0] + `
Repository version:  5
Store type:          local
Store account:       ` + storePath + `/
`
			if stdout != expected {
				t.Fatal("output does not match", stdout, expected)
			}
		},
		nil,
	)

	// Test repo info on an unlocked repo
	s.promptPwd.SetPasswords("hello world")
	runCmd(t,
		[]string{"repo", "info", "--store", "local:" + s.dirs[0]},
		nil,
		func(stdout string) {
			expected := `Repository ID:       ` + s.repoIds[0] + `
Repository version:  5
Store type:          local
Store account:       ` + storePath + `/
Total files stored:  1
`
			if stdout != expected {
				t.Fatal("output does not match", stdout)
			}
		},
		nil,
	)
}
