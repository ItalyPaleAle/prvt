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
	"errors"
	"os"

	"github.com/ItalyPaleAle/prvt/infofile"

	"github.com/spf13/cobra"
)

// Requires a minimum version of the info file to continue
func requireInfoFileVersion(info *infofile.InfoFile, version uint16, connectionString string) error {
	if connectionString == "" {
		connectionString = "<string>"
	}

	if info.Version < version {
		return NewExecError(ErrorUser, "Repository needs to be upgraded", errors.New(`Please run "prvt repo upgrade --store `+connectionString+`" to upgrade this repository to the latest format`))
	}

	return nil
}

// Adds the --store flag, with a default value read from the environment
func addStoreFlag(c *cobra.Command, required bool) {
	// Check if we have a value in the PRVT_STORE env var
	env := os.Getenv("PRVT_STORE")
	c.Flags().StringP("store", "s", env, "connection string for the store")
	if env == "" && required {
		c.MarkFlagRequired("store")
	}
}
