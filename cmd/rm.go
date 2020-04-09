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
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/spf13/cobra"
)

func init() {
	var (
		flagStoreConnectionString string
	)

	c := &cobra.Command{
		Use:   "rm",
		Short: "Remove a file or folder",
		Long: `Removes a file (or folder) from the repository.

Usage: "prvt rm <path> [<path> ...] --store <string>"

Removes a file or folder (recursively) from the repository. Once removed, files cannot be recovered.

To remove a file, specify its exact path. To remove a folder recursively, specify the name of the folder, ending with /*
`,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				utils.ExitWithError(utils.ErrorUser, "No file to remove", nil)
				return
			}

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
				utils.ExitWithError(utils.ErrorUser, "Store is not initialized", err)
				return
			}

			// Derive the master key
			masterKey, errMessage, err := GetMasterKey(info)
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, errMessage, err)
				return
			}
			store.SetMasterKey(masterKey)

			// Set up the index
			index.Instance.SetStore(store)

			// Iterate through the args and remove all files
			for _, e := range args {
				// Remove from the index
				objects, err := index.Instance.DeleteFile(e)
				if err != nil {
					utils.ExitWithError(utils.ErrorApp, "Failed to remove path from index: "+e, err)
					return
				}
				if objects == nil || len(objects) < 1 {
					fmt.Println("Nothing removed:", e)
					continue
				}

				// Delete the files
				for _, o := range objects {
					err = store.Delete(o, nil)
					if err != nil {
						utils.ExitWithError(utils.ErrorApp, "Failed to remove object from store: "+o+" (for path "+e+")", err)
						return
					}
				}
				var removed string
				if len(objects) == 1 {
					removed = "(1 file)"
				} else {
					removed = fmt.Sprintf("(%d files)", len(objects))
				}
				fmt.Println("Removed path:", e, removed)
			}
		},
	}

	// Flags
	c.Flags().StringVarP(&flagStoreConnectionString, "store", "s", "", "connection string for the store")
	c.MarkFlagRequired("store")

	// Add the command
	rootCmd.AddCommand(c)
}
