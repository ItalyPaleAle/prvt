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

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/spf13/cobra"
)

func init() {
	var (
		flagStoreConnectionString string
		flagKeyId                 string
	)

	c := &cobra.Command{
		Use:   "rm",
		Short: "Remove a passphrase or GPG key",
		Long: `Removes a passphrase or GPG key from those allowed to unlock the repository.

Usage: "prvt repo key rm --store <string> --key <string>"

You can find the list of passphrases and GPG keys authorized to unlock the repository using "prvt repo key ls --store <string>".

To identify a passphrase or a GPG key among those authorized, you can use the "prvt repo key test --store <string>" command.
`,
		DisableAutoGenTag: true,

		Run: func(cmd *cobra.Command, args []string) {
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
				utils.ExitWithError(utils.ErrorUser, "Repository is not initialized", err)
				return
			}

			// Require info files version 2 or higher
			if info.Version < 2 {
				utils.ExitWithError(utils.ErrorUser, "Repository needs to be upgraded", errors.New(`Please run "prvt repo upgrade --store <string>" to upgrade this repository to the latest format`))
				return
			}

			// Require at least 2 keys in the repository
			if len(info.Keys) < 2 {
				utils.ExitWithError(utils.ErrorUser, "Cannot remove the only key", errors.New("This repository has only one key, which cannot be removed"))
				return
			}

			// First, unlock the repository
			fmt.Println("Unlocking the repository")
			_, keyId, errMessage, err := GetMasterKey(info)
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, errMessage, err)
				return
			}
			fmt.Println("Repository unlocked")

			// The key we're removing must not be the same as the key used to unlock the repository
			if flagKeyId == keyId {
				utils.ExitWithError(utils.ErrorUser, "Invalid key ID", errors.New("You cannot remove the same key you're using to unlock the repository"))
				return
			}

			// Remove the key
			errMessage, err = RemoveKey(info, flagKeyId)
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, errMessage, err)
				return
			}

			// Store the info file
			err = store.SetInfoFile(info)
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Cannot store the info file", err)
				return
			}

			fmt.Println("Key removed")
		},
	}

	// Flags
	c.Flags().StringVarP(&flagStoreConnectionString, "store", "s", "", "connection string for the store")
	c.MarkFlagRequired("store")
	c.Flags().StringVarP(&flagKeyId, "key", "k", "", "ID of the key to remove")
	c.MarkFlagRequired("key")

	// Add the command
	repoKeyCmd.AddCommand(c)
}
