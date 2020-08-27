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

package keys

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"
)

// Cached path for GPG
var gpgPath string

// GPGEncrypt encrypts data using the GPG binary
func GPGEncrypt(in []byte, key string) (out []byte, err error) {
	return runGPG(in, "--output", "-", "--encrypt", "--recipient", key, "-")
}

// GPGDecrypt decrypts data using the GPG binary
func GPGDecrypt(in []byte) (out []byte, err error) {
	return runGPG(in, "--output", "-", "--decrypt", "-")
}

// runGPG runs the GPG command with the given flags
func runGPG(in []byte, flags ...string) (out []byte, err error) {
	// Get the GPG command
	path, err := lookupGPG()
	if err != nil {
		return
	}

	// Run GPG
	cmd := exec.Command(path, flags...)
	cmd.Stdin = bytes.NewReader(in)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	err = cmd.Run()
	if err != nil {
		return
	}
	return ioutil.ReadAll(&outBuf)
}

// lookupGPG returns the path of the GPG binary
func lookupGPG() (string, error) {
	// Cached value
	if gpgPath != "" {
		return gpgPath, nil
	}

	// First, look for gpg2
	path, err := exec.LookPath("gpg2")
	if err == nil && len(path) > 0 {
		gpgPath = path
		return path, nil
	}

	// Try gpg
	path, err = exec.LookPath("gpg")
	if err == nil && len(path) > 0 {
		gpgPath = path
		return path, nil
	}

	// Couldn't find the binary
	return "", errors.New("could not find GPG binary in PATH")
}
