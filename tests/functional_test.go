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
	"os"
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
			func(stdout string) {
				// No-op
			},
			nil,
		)
	})
}

// Set up testing environment; returns a callback whose execution should be deferred till all tests are run
func (s *funcTestSuite) setup() func() {
	// Set streams for promptui and disable exiting on error
	s.promptPwd = &passwordPrompter{}
	cmd.PromptuiStdin = s.promptPwd
	cmd.PromptuiStdout = s.promptPwd
	cmd.NoExitOnError = true

	// Create 2 temporary folders for the tests
	s.dirs = []string{
		s.t.TempDir(),
		s.t.TempDir(),
	}

	return func() {
		// Restore globals
		cmd.PromptuiStdin = os.Stdin
		cmd.PromptuiStdout = os.Stdout
		cmd.NoExitOnError = false
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
	pass := ""
	if len(o.passwords) == 1 {
		pass = o.passwords[0]
		o.passwords = []string{}
	} else if len(o.passwords) > 1 {
		pass, o.passwords = o.passwords[0], o.passwords[1:]
	}
	n = copy(p, pass+"\n")
	return
}

func (passwordPrompter) Write(p []byte) (n int, err error) {
	// Do nothing with what we read from p, but respond as if we read it all
	return len(p), nil
}

func (passwordPrompter) Close() error {
	return nil
}
