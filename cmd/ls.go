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
	"sort"
	"strings"
	"time"

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/fs/fsindex"
	"github.com/ItalyPaleAle/prvt/index"

	"github.com/spf13/cobra"
)

// NewLsCmd is for "prvt ls"
func NewLsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "ls",
		Short: "List files and folders",
		Long: `List files and folders contained in the repository at a specific path.

Usage: "prvt ls [<path>] --store <string>"

Shows the list of all files and folders contained in the repository at a given path. If the path is omitted, it's assumed to be "/", which is the root of the repository.
`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return NewExecError(ErrorUser, "Can only pass one path", nil)
			}

			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'store'", err)
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
			defer store.ReleaseLock()

			// Request the info file
			info, err := store.GetInfoFile()
			if err != nil {
				return NewExecError(ErrorApp, "Error requesting the info file", err)
			}
			if info == nil {
				return NewExecError(ErrorUser, "Repository is not initialized", err)
			}

			// Derive the master key
			masterKey, keyId, errMessage, err := GetMasterKey(info)
			if err != nil {
				return NewExecError(ErrorUser, errMessage, err)
			}
			store.SetMasterKey(keyId, masterKey)

			// Set up the index
			indexProvider := &fsindex.IndexProviderFs{
				Store: store,
			}
			idx := &index.Index{}
			idx.SetProvider(indexProvider)

			// Get the path and ensure it starts with /
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			// Get the list of files in the folder
			list, err := idx.ListFolder(0, path)
			if err != nil {
				return NewExecError(ErrorApp, "Error listing the contents of the folder", err)
			}

			// Sort the result
			sort.Slice(list, func(i, j int) bool {
				// Directories go always first
				if list[i].Directory != list[j].Directory {
					return list[i].Directory
				}
				return list[i].Path < list[j].Path
			})

			// Print the result
			for _, el := range list {
				if el.Directory {
					fmt.Fprintln(cmd.OutOrStdout(), el.Path+"/")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), el.Path)
				}
			}

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)

	return c
}
