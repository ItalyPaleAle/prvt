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
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ItalyPaleAle/prvt/cmd"
)

// TestFunctional performs a series of functional tests on the CLI and on the server
func TestFunctional(t *testing.T) {
	// Run the test suite
	suite := funcTestSuite{
		t: t,
	}
	suite.Run()
}

// Test suite
type funcTestSuite struct {
	t *testing.T

	promptPwd *passwordPrompter
	dirs      []string
}

// Run the test suite
func (s *funcTestSuite) Run() {
	// Setup and teardown
	teardown := s.setup()
	defer teardown()

	s.t.Run("init repo", func(t *testing.T) {
		s.promptPwd.SetPasswords("hello world")
		runCmd(t,
			[]string{"repo", "init", "--store", "local:" + s.dirs[0]},
			nil,
			func(stdout string) {
				if !strings.HasPrefix(stdout, "Initialized new repository in the store local:") {
					t.FailNow()
				}
			},
			nil,
		)
	})
}

// Set up testing environment; returns a callback whose execution should be deferred till all tests are run
func (s *funcTestSuite) setup() func() {
	// Set streams for promptui
	s.promptPwd = &passwordPrompter{}
	cmd.PromptuiStdin = s.promptPwd
	cmd.PromptuiStdout = s.promptPwd

	// Create 2 temporary folders for the tests
	s.dirs = []string{
		s.t.TempDir(),
		s.t.TempDir(),
	}

	return func() {
		// Restore globals
		cmd.PromptuiStdin = os.Stdin
		cmd.PromptuiStdout = os.Stdout
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
