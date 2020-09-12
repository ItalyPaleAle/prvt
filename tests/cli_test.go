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
	"strings"
	"testing"

	"github.com/ItalyPaleAle/prvt/cmd"

	"github.com/spf13/cobra"
)

// TestCLICore tests base CLI functionality, ensuring that all required commands are defined and loaded
func TestCLICore(t *testing.T) {
	// Root command
	t.Run("root command", func(t *testing.T) {
		runCmd(t,
			nil,
			nil,
			func(stdout string) {
				// Ensure this contains the description
				if !strings.HasPrefix(stdout, "prvt") {
					t.Error("Output does not have the desired prefix")
				}
			},
			nil)
	})

	// Test commands from a table
	table := []struct {
		Name    string
		Args    []string
		Command *cobra.Command
	}{
		{"prvt add", []string{"add", "--help"}, cmd.NewAddCmd()},
		{"prvt ls", []string{"ls", "--help"}, cmd.NewLsCmd()},
		{"prvt repo", []string{"repo", "--help"}, cmd.NewRepoCmd()},
		{"prvt repo info", []string{"repo", "info", "--help"}, cmd.NewRepoInfoCmd()},
		{"prvt repo init", []string{"repo", "init", "--help"}, cmd.NewRepoInitCmd()},
		{"prvt repo key", []string{"repo", "key", "--help"}, cmd.NewRepoKeyCmd()},
		{"prvt repo key add", []string{"repo", "key", "add", "--help"}, cmd.NewRepoKeyAddCmd()},
		{"prvt repo key ls", []string{"repo", "key", "ls", "--help"}, cmd.NewRepoKeyLsCmd()},
		{"prvt repo key rm", []string{"repo", "key", "rm", "--help"}, cmd.NewRepoKeyRmCmd()},
		{"prvt repo key test", []string{"repo", "key", "test", "--help"}, cmd.NewRepoKeyTestCmd()},
		{"prvt repo upgrade", []string{"repo", "upgrade", "--help"}, cmd.NewRepoUpgradeCmd()},
		{"prvt rm", []string{"rm", "--help"}, cmd.NewRmCmd()},
		{"prvt serve", []string{"serve", "--help"}, cmd.NewServeCmd()},
		{"prvt version", []string{"version", "--help"}, cmd.NewVersionCmd()},
	}

	for _, el := range table {
		t.Run(el.Name, func(t *testing.T) {
			runCmd(t,
				el.Args,
				nil,
				func(stdout string) {
					// Ensure this contains the description
					if !strings.HasPrefix(stdout, el.Command.Long) {
						t.Error("Output does not have the desired prefix")
					}
				},
				nil)
		})
	}
}

// TestCLIVersionCommand tests the "prvt version" command and ensures it's returning the right version
func TestCLIVersionCommand(t *testing.T) {
	// prvt version
	t.Run("no version defined", func(t *testing.T) {
		runCmd(t,
			[]string{"version"},
			nil,
			func(stdout string) {
				// Ensure this starts with "This prvt build does not contain a build identifier,"
				if !strings.HasPrefix(string(stdout), "This prvt build does not contain a build identifier,") {
					t.Error("Output does not have the desired prefix")
				}
			},
			nil)
	})

	t.Run("with version defined", func(t *testing.T) {
		reset := setBuildInfo()
		runCmd(t,
			[]string{"version"},
			nil,
			func(stdout string) {
				expect := "prvt ci\nBuild ID: 1 (2020)\nGit commit: a1b2c3d4e5f6\nRuntime: go1."
				if !strings.HasPrefix(string(stdout), expect) {
					t.Error("Output does not have the desired prefix")
				}
			},
			nil)
		reset()
	})
}
