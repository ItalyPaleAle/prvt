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
	"io/ioutil"
	"testing"

	"github.com/ItalyPaleAle/prvt/cmd"
)

type errCbFunc func(error)
type validateFunc func(string)

func runCmd(t *testing.T, args []string, errCb errCbFunc, stdoutValidate validateFunc, stderrValidate validateFunc) {
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
