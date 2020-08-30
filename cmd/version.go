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

package cmd

import (
	"fmt"
	"runtime"

	"github.com/ItalyPaleAle/prvt/buildinfo"

	"github.com/spf13/cobra"
)

// NewVersionCmd is for "prvt version"
func NewVersionCmd() *cobra.Command {
	c := &cobra.Command{
		Use:               "version",
		Short:             "Show prvt version",
		Long:              `Prints the version of this prvt build, and other information on the binary.`,
		DisableAutoGenTag: true,

		Run: func(cmd *cobra.Command, args []string) {
			if buildinfo.BuildID == "" || buildinfo.CommitHash == "" {
				fmt.Println("This prvt build does not contain a build identifier, and it was probably fetched from the repository as source")
			} else {
				fmt.Printf("prvt %s\nBuild ID: %s (%s)\nGit commit: %s\nRuntime: %s\n", buildinfo.AppVersion, buildinfo.BuildID, buildinfo.BuildTime, buildinfo.CommitHash, runtime.Version())
			}
		},
	}

	return c
}
