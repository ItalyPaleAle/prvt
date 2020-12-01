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
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/fs/fsindex"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/repository"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// NewAddCmd is for "prvt add"
func NewAddCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "add",
		Short: "Add a file or folder",
		Long: `Adds a file or folder to a repository.

Usage: "prvt add <file> [<file> ...] --store <string> --destination <string>"

You can add multiple files or folders from the local file system; folders will be added recursively.

You must specify a destination, which is a folder inside the repository where your files will be added. The value for the --destination flag must begin with a slash.
`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'store'", err)
			}
			flagDestination, err := cmd.Flags().GetString("destination")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'destination'", err)
			}
			flagForce, err := cmd.Flags().GetBool("force")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'force'", err)
			}

			// Get the file/folder name from the args
			if len(args) < 1 {
				return NewExecError(ErrorUser, "no file or folder specified", nil)
			}

			// Destination must start with "/"
			if !strings.HasPrefix(flagDestination, "/") {
				return NewExecError(ErrorUser, "destination must start with /", nil)
			}

			// Ensure destination ends with a "/"
			if !strings.HasSuffix(flagDestination, "/") {
				flagDestination += "/"
			}

			// Create the store object
			store, err := fs.GetWithConnectionString(flagStoreConnectionString)
			if err != nil || store == nil {
				return NewExecError(ErrorUser, "Could not initialize store", err)
			}

			// Acquire a lock
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err = store.AcquireLock(ctx)
			cancel()
			if err != nil {
				return NewExecError(ErrorApp, "Could not acquire a lock. Please make sure that no other instance of prvt is running with the same repo.", err)
			}
			defer store.ReleaseLock(context.Background())

			// Request the info file
			info, err := store.GetInfoFile()
			if err != nil {
				return NewExecError(ErrorApp, "Error requesting the info file", err)
			}
			if info == nil {
				return NewExecError(ErrorUser, "Repository is not initialized", err)
			}

			// Require info files version 3 or higher before any operation that changes the store (which would update the index to the protobuf-based format)
			err = requireInfoFileVersion(info, 3, flagStoreConnectionString)
			if err != nil {
				return err
			}

			// Derive the master key
			masterKey, keyId, errMessage, err := GetMasterKey(info)
			if err != nil {
				return NewExecError(ErrorUser, errMessage, err)
			}
			store.SetMasterKey(keyId, masterKey)

			// Set up the repository and index
			repo := repository.Repository{
				Store: store,
				Index: &index.Index{},
			}
			indexProvider := &fsindex.IndexProviderFs{
				Store: store,
			}
			repo.Index.SetProvider(indexProvider)

			// Start a transaction with the index to add all files
			err = repo.BeginTransaction()
			if err != nil {
				return NewExecError(ErrorApp, "Error starting a transaction", err)
			}

			// Iterate through the args and add them all
			res := make(chan repository.PathResultMessage)
			go func() {
				var err error
				var expanded string
				for _, e := range args {
					// Get the target and folder
					expanded, err = homedir.Expand(e)
					if err != nil {
						res <- repository.PathResultMessage{
							Path:   e,
							Status: repository.RepositoryStatusInternalError,
							Err:    err,
						}
						break
					}
					folder := filepath.Dir(expanded)
					target := filepath.Base(expanded)

					repo.AddPath(context.Background(), folder, target, flagDestination, flagForce, res)
				}

				close(res)
			}()

			// Print each message
			for el := range res {
				switch el.Status {
				case repository.RepositoryStatusOK:
					fmt.Fprintln(cmd.OutOrStdout(), "Added:", el.Path)
				case repository.RepositoryStatusIgnored:
					fmt.Fprintln(cmd.OutOrStdout(), "Ignoring:", el.Path)
				case repository.RepositoryStatusExisting:
					fmt.Fprintln(cmd.OutOrStdout(), "Skipping existing file:", el.Path)
				case repository.RepositoryStatusInternalError:
					fmt.Fprintf(cmd.OutOrStdout(), "Internal error adding file '%s': %s\n", el.Path, el.Err)
				case repository.RepositoryStatusUserError:
					fmt.Fprintf(cmd.OutOrStdout(), "Error adding file '%s': %s\n", el.Path, el.Err)
				}
			}

			// End the transaction
			err = repo.CommitTransaction()
			if err != nil {
				return NewExecError(ErrorApp, "Error committing a transaction", err)
			}

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)
	c.Flags().StringP("destination", "d", "", "destination folder")
	c.MarkFlagRequired("destination")
	c.Flags().BoolP("force", "f", false, "overwrite a file if existing")

	return c
}
