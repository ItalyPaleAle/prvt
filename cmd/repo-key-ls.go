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
	"github.com/ItalyPaleAle/prvt/keys"

	"github.com/spf13/cobra"
)

// NewRepoKeyLsCmd is for "prvt repo key ls"
func NewRepoKeyLsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "ls",
		Short: "List all keys for the repo",
		Long: `Prints the list of keys (passphrases and GPG keys) that can unlock the repo.

Usage: "prvt repo key ls --store <string>"
`,
		DisableAutoGenTag: true,

		Run: func(cmd *cobra.Command, args []string) {
			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorApp, "Cannot get flag 'store'", err)
				return
			}

			// Create the store object
			store, err := fs.GetWithConnectionString(flagStoreConnectionString)
			if err != nil || store == nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorUser, "Could not initialize store", err)
				return
			}

			// Request the info file
			info, err := store.GetInfoFile()
			if err != nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorApp, "Error requesting the info file", err)
				return
			}
			if info == nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorUser, "Repository is not initialized", err)
				return
			}

			// Require info files version 2 or higher
			if !requireInfoFileVersion(cmd.ErrOrStderr(), info, 2, flagStoreConnectionString) {
				return
			}

			// Table headers
			fmt.Fprintln(cmd.OutOrStdout(), "KEY TYPE    | KEY ID")
			fmt.Fprintln(cmd.OutOrStdout(), "------------|--------------------")

			// Show all keys in a table
			// First, show all passphrases
			for _, k := range info.Keys {
				if k.GPGKey == "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Passphrase  | p:%X\n", k.MasterKey[0:8])
				}
			}
			// Now, show all GPG keys
			for _, k := range info.Keys {
				if k.GPGKey != "" {
					// Get the owner of the GPG key
					uid := keys.GPGUID(k.GPGKey)
					if uid != "" {
						uid = "  (" + uid + ")"
					}

					fmt.Fprintln(cmd.OutOrStdout(), "GPG Key     | "+k.GPGKey+uid)
				}
			}
		},
	}

	// Flags
	addStoreFlag(c, true)

	return c
}
