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

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/spf13/cobra"
)

func init() {
	var (
		flagStoreConnectionString string
	)

	c := &cobra.Command{
		Use:   "test",
		Short: "Test a key for unlocking the repo",
		Long: `Tests a key and returns the ID of the key used to unlock the repo.

Usage: "prvt repo key test --store <string>"

This command is particularly useful to determine the ID of a key that you want to remove.
`,
		DisableAutoGenTag: true,

		Run: func(cmd *cobra.Command, args []string) {
			// Create the store object
			store, err := fs.Get(flagStoreConnectionString)
			if err != nil || store == nil {
				utils.ExitWithError(utils.ErrorUser, "Could not initialize store", err)
				return
			}

			// Request the info file
			info, err := store.GetInfoFile()
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Error requesting the info file", err)
				return
			}
			if info == nil {
				utils.ExitWithError(utils.ErrorUser, "Repository is not initialized", err)
				return
			}

			// Require info files version 2 or higher
			if !requireInfoFileVersion(info, 2, flagStoreConnectionString) {
				return
			}

			// Unlock the repository
			_, keyId, errMessage, err := GetMasterKey(info)
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, errMessage, err)
				return
			}

			// Show the key ID
			fmt.Println("Repository unlocked using key with ID:", keyId)
		},
	}

	// Flags
	addStoreFlag(c, &flagStoreConnectionString, true)

	// Add the command
	repoKeyCmd.AddCommand(c)
}
