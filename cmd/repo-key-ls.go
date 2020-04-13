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
	"fmt"
	"strconv"

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/spf13/cobra"
)

func init() {
	var (
		flagStoreConnectionString string
	)

	c := &cobra.Command{
		Use:   "ls",
		Short: "List all keys for the repo",
		Long: `Prints the list of keys (passphrases and GPG keys) that can unlock the repo.

Usage: "prvt repo key ls --store <string>"
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
			if info.Version < 2 {
				utils.ExitWithError(utils.ErrorUser, "Repository needs to be upgraded", errors.New(`Please run "prvt repo upgrade --store <string>" to upgrade this repository to the latest format`))
				return
			}

			// Table headers
			fmt.Println("KEY TYPE    | KEY ID")
			fmt.Println("------------|------------------------")

			// Show all keys in a table
			// First, show all passphrases
			i := 0
			for _, k := range info.Keys {
				if k.GPGKey == "" {
					fmt.Println("Passphrase  | p:" + strconv.Itoa(i))
					i++
				}
			}
			// Now, show all GPG keys
			for _, k := range info.Keys {
				if k.GPGKey != "" {
					fmt.Println("GPG Key     | " + k.GPGKey)
				}
			}
		},
	}

	// Flags
	c.Flags().StringVarP(&flagStoreConnectionString, "store", "s", "", "connection string for the store")
	c.MarkFlagRequired("store")

	// Add the command
	repoKeyCmd.AddCommand(c)
}
