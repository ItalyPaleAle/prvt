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
	"github.com/ItalyPaleAle/prvt/index"

	"github.com/spf13/cobra"
)

// NewRepoInfoCmd is for "prvt repo info"
func NewRepoInfoCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "info",
		Short: "Show information about the repository",
		Long: `This command returns information about a repository, such as its version (based on the version of the info file) and number of files. Certain information is only available if the repository is unlocked.
`,
		DisableAutoGenTag: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'store'", err)
			}
			flagNoUnlock, err := cmd.Flags().GetBool("no-unlock")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'no-unlock'", err)
			}

			// Create the store object
			store, err := fs.GetWithConnectionString(flagStoreConnectionString)
			if err != nil || store == nil {
				return NewExecError(ErrorUser, "Could not initialize store", err)
			}

			// Request the info file
			info, err := store.GetInfoFile()
			if err != nil {
				return NewExecError(ErrorApp, "Error requesting the info file", err)
			}
			if info == nil {
				return NewExecError(ErrorUser, "Repository is not initialized", err)
			}

			var stat *index.IndexStats

			if !flagNoUnlock {
				// Unlock the repository
				masterKey, keyId, errMessage, err := GetMasterKey(info)
				if err != nil {
					return NewExecError(ErrorUser, errMessage, err)
				}
				store.SetMasterKey(keyId, masterKey)

				// Set up the index
				idx := &index.Index{}
				idx.SetStore(store)

				// Get stats
				stat, err = idx.Stat()
				if err != nil {
					return NewExecError(ErrorApp, "Could not get the stats from the repository", err)
				}
			}

			// Show the version
			fmt.Fprintf(cmd.OutOrStdout(), "Repository version:  %d\n", info.Version)

			// Show the stats, if any
			if stat != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Total files stored:  %d\n", stat.FileCount)
			}

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)
	c.Flags().Bool("no-unlock", false, "do not unlock the repo")

	return c
}
