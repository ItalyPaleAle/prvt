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
	"github.com/ItalyPaleAle/prvt/repository"

	"github.com/spf13/cobra"
)

// NewRmCmd is for "prvt rm"
func NewRmCmd() *cobra.Command {
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
				ExitWithError(cmd.ErrOrStderr(), ErrorUser, "No file to remove", nil)
				return
			}

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

			// Require info files version 3 or higher before any operation that changes the store (which would update the index to the protobuf-based format)
			if !requireInfoFileVersion(cmd.ErrOrStderr(), info, 3, flagStoreConnectionString) {
				return
			}

			// Derive the master key
			masterKey, keyId, errMessage, err := GetMasterKey(info)
			if err != nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorUser, errMessage, err)
				return
			}
			store.SetMasterKey(keyId, masterKey)

			// Set up the index
			index.Instance.SetStore(store)

			// Set up the repository
			repo := repository.Repository{
				Store: store,
			}

			// Iterate through the args and remove all files
			res := make(chan repository.PathResultMessage)
			go func() {
				for _, e := range args {
					repo.RemovePath(e, res)
				}

				close(res)
			}()

			// Print each message
			for el := range res {
				switch el.Status {
				case repository.RepositoryStatusOK:
					fmt.Fprintln(cmd.OutOrStdout(), "Removed:", el.Path)
				case repository.RepositoryStatusNotFound:
					fmt.Fprintln(cmd.OutOrStdout(), "Not found:", el.Path)
				case repository.RepositoryStatusInternalError:
					fmt.Fprintf(cmd.OutOrStdout(), "Internal error removing path '%s': %s\n", el.Path, el.Err)
				case repository.RepositoryStatusUserError:
					fmt.Fprintf(cmd.OutOrStdout(), "Error removing path '%s': %s\n", el.Path, el.Err)
				}
			}
		},
	}

	// Flags
	addStoreFlag(c, true)

	return c
}
