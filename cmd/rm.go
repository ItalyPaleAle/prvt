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
	"bytes"
	"fmt"

	"e2e/crypto"
	"e2e/fs"
	"e2e/index"
	"e2e/utils"

	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:               "rm",
		Short:             "remove a file or folder",
		Long:              ``,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Create the store object
			store, err := fs.Get(storeConnectionString)
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

			// Get the passphrase and derive the master key
			passphrase, err := utils.PromptPassphrase()
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, "Error getting passphrase", err)
				return
			}
			masterKey, confirmationHash, err := crypto.KeyFromPassphrase(passphrase, info.Salt)
			if bytes.Compare(info.ConfirmationHash, confirmationHash) != 0 {
				utils.ExitWithError(utils.ErrorUser, "Invalid passphrase", err)
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
	rootCmd.AddCommand(c)
}
