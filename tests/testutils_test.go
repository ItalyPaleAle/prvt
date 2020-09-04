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
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ItalyPaleAle/prvt/cmd"
	"github.com/ItalyPaleAle/prvt/utils"
)

// Run a CLI command
func runCmd(t *testing.T, args []string, errCb func(error), stdoutValidate func(string), stderrValidate func(string)) {
	if args == nil {
		args = []string{}
	}
	bStdout := &bytes.Buffer{}
	bStderr := &bytes.Buffer{}

	// Invoke the command
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetOut(bStdout)
	rootCmd.SetErr(bStderr)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()

	// If there's an error, pass it to the callback if present; otherwise, fail with it
	if err != nil {
		if errCb != nil {
			errCb(err)
		} else {
			t.Fatal(err)
		}
	}

	// Read the output
	stdout, err := ioutil.ReadAll(bStdout)
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := ioutil.ReadAll(bStderr)
	if err != nil {
		t.Fatal(err)
	}

	// Validate stdout if requested, otherwise accept everything
	if stdoutValidate != nil {
		stdoutValidate(string(stdout))
	}

	// Validate stderr
	if stderrValidate != nil {
		stderrValidate(string(stderr))
	} else {
		// Ensure it's empty
		if len(stderr) != 0 {
			t.Errorf("stderr is not empty:\n%s\n", stderr)
		}
	}
}

// Checks if the directory containing the repository has the correct number of files
func checkRepoDirectory(t *testing.T, path string, expectFiles int) {
	// Check if the info and index files exists
	if exists, _ := utils.IsRegularFile(filepath.Join(path, "_info.json")); !exists {
		t.Error("file does not exist: _info.json")
		return
	}
	if exists, _ := utils.IsRegularFile(filepath.Join(path, "_index")); !exists {
		t.Error("file does not exist: _index")
		return
	}

	// Count the number of files in the data folder, recursively
	found := 0
	err := filepath.Walk(filepath.Join(path, "data"),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() {
				found++
			}
			return nil
		})
	if err != nil {
		t.Error(err)
		return
	}
	if found != expectFiles {
		t.Errorf("expected to find %d files, found %d", expectFiles, found)
	}
}

// Used to pass a password to promptui
type passwordPrompter struct {
	passwords []string
}

func (o *passwordPrompter) SetPasswords(passwords ...string) {
	o.passwords = passwords
}

func (o *passwordPrompter) Read(p []byte) (n int, err error) {
	if len(o.passwords) == 0 {
		n = 0
		err = io.EOF
	} else {
		n = copy(p, o.passwords[0]+"\n")
		if len(o.passwords) > 1 {
			// pop from the queue
			o.passwords = o.passwords[1:]
		} else {
			o.passwords = []string{}
		}
	}
	return
}

func (passwordPrompter) Write(p []byte) (n int, err error) {
	// Do nothing with what we read from p, but respond as if we read it all
	return len(p), nil
}

func (passwordPrompter) Close() error {
	return nil
}
