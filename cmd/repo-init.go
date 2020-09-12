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

	"github.com/spf13/cobra"
)

// NewRepoInitCmd is for "prvt repo init"
func NewRepoInitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new repository",
		Long: `Initializes a new, empty repository, and sets the encryption key to use.

Usage: "prvt repo init --store <string>"

See the help page for prvt ("prvt --help") for details on stores and how to configure them.

If you want to use a GPG key to protect this repository (including GPG keys stored in security tokens or smart cards), use the "--gpg" flag with the address or ID of a public GPG key. For example: "prvt repo init --store <string> --gpg mykey@example.com" 
In order to use GPG keys, you need to have GPG version 2 installed separately. You also need a GPG keypair (public and private) in your keyring.
`,
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
				return NewExecError(ErrorUser, "Could not initialize repository", err)
			}

			// Create the info file after generating a new master key
			info, errMessage, err := NewInfoFile(flagGPGKey)
			if err != nil {
				return NewExecError(ErrorUser, errMessage, err)
			}

			// Check if the file exists already
			// We are expecting this to be empty
			infoExisting, err := store.GetInfoFile()
			if err == nil && infoExisting != nil {
				return NewExecError(ErrorApp, "Error initializing repository", errors.New("A repository is already initialized in this store"))
			}

			// Store the info file
			err = store.SetInfoFile(info)
			if err != nil {
				return NewExecError(ErrorApp, "Cannot store the info file", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized new repository in the store %s\n", flagStoreConnectionString)

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)
	c.Flags().StringP("gpg", "g", "", "protect the master key with the gpg key with this address (optional)")

	return c
}
