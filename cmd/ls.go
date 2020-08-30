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
	"sort"
	"strings"

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "ls",
		Short: "List files and folders",
		Long: `List files and folders contained in the repository at a specific path.

Usage: "prvt ls [<path>] --store <string>"

Shows the list of all files and folders contained in the repository at a given path. If the path is omitted, it's assumed to be "/", which is the root of the repository.
`,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				utils.ExitWithError(utils.ErrorUser, "Can only pass one path", nil)
				return
			}

			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Cannot get flag 'store'", err)
				return
			}

			// Create the store object
			store, err := fs.GetWithConnectionString(flagStoreConnectionString)
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

			// Derive the master key
			masterKey, keyId, errMessage, err := GetMasterKey(info)
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, errMessage, err)
				return
			}
			store.SetMasterKey(keyId, masterKey)

			// Set up the index
			index.Instance.SetStore(store)

			// Get the path and ensure it starts with /
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			// Get the list of files in the folder
			list, err := index.Instance.ListFolder(path)
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Error listing the contents of the folder", err)
				return
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
					fmt.Println(el.Path + "/")
				} else {
					fmt.Println(el.Path)
				}
			}
		},
	}

	// Flags
	addStoreFlag(c, true)

	// Add the command
	rootCmd.AddCommand(c)
}
