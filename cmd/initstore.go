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

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:               "initstore",
		Short:             "initialize a new store",
		Long:              ``,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Create the store object
			store, err := fs.Get(storeConnectionString)
			if err != nil || store == nil {
				utils.ExitWithError(utils.ErrorUser, "Could not initialize store", err)
				return
			}

			// Get the passphrase and derive the master key, after generating a new salt
			passphrase, err := utils.PromptPassphrase()
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, "Error getting passphrase", err)
				return
			}
			salt, err := crypto.NewSalt()
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Error generating a new salt", err)
				return
			}
			_, confirmationHash, err := crypto.KeyFromPassphrase(passphrase, salt)
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Error deriving the master key", err)
				return
			}

			// Set up the index
			index.Instance.SetStore(store)

			// Check if the file exists already
			// We are expecting this to be empty
			info, err := store.GetInfoFile()
			if err == nil {
				utils.ExitWithError(utils.ErrorApp, "Error initializing store", errors.New("store is already initialized"))
				return
			}
			if info != nil {
				utils.ExitWithError(utils.ErrorUser, "Error initializing store", errors.New("store is already initialized"))
				return
			}

			// Create the info file
			info, err = fs.InfoCreate(salt, confirmationHash)
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Error creating info file", err)
				return
			}

			// Store the info file
			err = store.SetInfoFile(info)
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Cannot store the info file", err)
				return
			}

			fmt.Println("Initialized new store")
		},
	}
	rootCmd.AddCommand(c)
}
