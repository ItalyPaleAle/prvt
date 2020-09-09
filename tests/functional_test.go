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
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ItalyPaleAle/prvt/cmd"
)

// TestFunctional performs a series of functional tests on the CLI and on the server
func TestFunctional(t *testing.T) {
	// Run the test suite
	suite := funcTestSuite{}
	suite.Run(t)
}

// Test suite
type funcTestSuite struct {
	gpgKeyId   string
	gpgKeyUser string
	promptPwd  *passwordPrompter
	dirs       []string
	fixtures   string
	client     *http.Client
	serverAddr string
	fileIds    map[string]string
}

// Run the test suite
func (s *funcTestSuite) Run(t *testing.T) {
	// Setup and teardown
	teardown := s.Setup(t)
	defer teardown()

	// CLI tests
	t.Run("CLI", s.RunCLI)

	// Server tests
	t.Run("server", s.RunServer)

	// Previous versions test
	t.Run("previous versions", s.RunPreviousVersions)
}

// Set up testing environment; returns a callback whose execution should be deferred till all tests are run
func (s *funcTestSuite) Setup(t *testing.T) func() {
	// GPG key ID
	s.gpgKeyId = os.Getenv("GPGKEY_ID")
	if s.gpgKeyId == "" {
		t.Fatal("empty GPG key ID: make sure you set the GPGKEY_ID environmental variable")
	}
	s.gpgKeyUser = os.Getenv("GPGKEY_USER")

	// Set streams for promptui
	s.promptPwd = &passwordPrompter{}
	cmd.PromptuiStdin = s.promptPwd
	cmd.PromptuiStdout = s.promptPwd

	// Path to the fixtures folder
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not get current path")
	}
	s.fixtures = filepath.Join(filepath.Dir(filename), "fixtures")

	// Create 2 temporary folders for the tests
	s.dirs = []string{
		t.TempDir(),
		t.TempDir(),
	}

	// HTTP client
	s.client = &http.Client{
		Timeout: 20 * time.Second,
	}

	// Other variables
	s.fileIds = make(map[string]string)

	return func() {
		// Restore globals
		cmd.PromptuiStdin = os.Stdin
		cmd.PromptuiStdout = os.Stdout
	}
}
