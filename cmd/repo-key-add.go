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

	"github.com/spf13/cobra"
)

// NewRepoKeyAddCmd is for "prvt repo key add"
func NewRepoKeyAddCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "add",
		Short: "Add a passphrase or GPG key to the repo",
		Long: `Adds a passphrase or GPG key that can unlock a repo

Usage: "prvt repo key add --store <string>"

If you want to add a GPG key (including GPG keys stored in security tokens or smart cards), use the "--gpg" flag with the address or ID of a public GPG key. For example: "prvt repo key add --store <string> --gpg mykey@example.com" 
In order to use GPG keys, you need to have GPG version 2 installed separately. You also need a GPG keypair (public and private) in your keyring.`,
		DisableAutoGenTag: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'store'", err)
			}
			flagGPGKey, err := cmd.Flags().GetString("gpg")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'gpg'", err)
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

			// Require info files version 2 or higher
			err = requireInfoFileVersion(info, 2, flagStoreConnectionString)
			if err != nil {
				return err
			}

			// If we have a GPG key, ensure it's not already added
			if flagGPGKey != "" {
				for _, k := range info.Keys {
					if k.GPGKey == flagGPGKey {
						return NewExecError(ErrorUser, "A GPG key with the same ID is already authorized to unlock this repository", err)
					}
				}
			}

			// First, unlock the repository
			fmt.Fprintln(cmd.OutOrStdout(), "Unlocking the repository: if prompted for a passphrase, please type an existing one")
			masterKey, _, errMessage, err := GetMasterKey(info)
			if err != nil {
				return NewExecError(ErrorUser, errMessage, err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Repository unlocked")

			// Add the new key
			if flagGPGKey == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Type the new passphrase")
			}
			keyId, errMessage, err := AddKey(info, masterKey, flagGPGKey)
			if err != nil {
				return NewExecError(ErrorUser, errMessage, err)
			}

			// Store the info file
			err = store.SetInfoFile(info)
			if err != nil {
				return NewExecError(ErrorApp, "Cannot store the info file", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Key added with id:", keyId)

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)
	c.Flags().StringP("gpg", "g", "", "protect the master key with the gpg key with this address (optional)")

	return c
}
