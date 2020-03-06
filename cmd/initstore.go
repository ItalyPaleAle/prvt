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

	"e2e/fs"
	"e2e/index"
	"e2e/utils"

	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:               "initstore",
		Short:             "Initialize a new store",
		Long:              ``,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Get the master key and create the filesystem object
			store, err := fs.Get(storeConnectionString)
			if err != nil || store == nil {
				utils.ExitWithError(utils.ErrorUser, "Could not initialize store", err)
				return
			}
			masterKey, err := utils.PromptMasterKey()
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, "Error getting master key", err)
				return
			}
			store.SetMasterKey([]byte(masterKey))
			index.Instance.SetStore(store)

			// Create the info file, which is encrypted also to verify the passphrase
			infoFile, err := utils.InfoCreate()
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Error creating info file", err)
				return
			}

			// Store the info file
			buf := bytes.NewReader(infoFile)
			_, err = store.Set("info", buf, nil, "info.json", "application/json", int64(len(infoFile)))
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Cannot store the info file", err)
				return
			}

			fmt.Println("Initialized new store")
		},
	}
	rootCmd.AddCommand(c)
}
