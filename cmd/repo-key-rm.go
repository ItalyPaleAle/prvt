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
	"errors"
	"fmt"
	"time"

	"github.com/ItalyPaleAle/prvt/fs"

	"github.com/spf13/cobra"
)

// NewRepoKeyRmCmd is for "prvt repo key rm"
func NewRepoKeyRmCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "rm",
		Short: "Remove a passphrase or GPG key",
		Long: `Removes a passphrase or GPG key from those allowed to unlock the repository.

Usage: "prvt repo key rm --store <string> --key <string>"

You can find the list of passphrases and GPG keys authorized to unlock the repository using "prvt repo key ls --store <string>".

To identify a passphrase or a GPG key among those authorized, you can use the "prvt repo key test --store <string>" command.
`,
		DisableAutoGenTag: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'store'", err)
			}
			flagKeyId, err := cmd.Flags().GetString("key")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'key'", err)
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

			// Require info files version 2 or higher
			err = requireInfoFileVersion(info, 2, flagStoreConnectionString)
			if err != nil {
				return err
			}

			// Require at least 2 keys in the repository
			if len(info.Keys) < 2 {
				return NewExecError(ErrorUser, "Cannot remove the only key", errors.New("This repository has only one key, which cannot be removed"))
			}

			// First, unlock the repository
			fmt.Fprintln(cmd.OutOrStdout(), "Unlocking the repository")
			_, keyId, errMessage, err := GetMasterKey(info)
			if err != nil {
				return NewExecError(ErrorUser, errMessage, err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Repository unlocked")

			// The key we're removing must not be the same as the key used to unlock the repository
			if flagKeyId == keyId {
				return NewExecError(ErrorUser, "Invalid key ID", errors.New("You cannot remove the same key you're using to unlock the repository"))
			}

			// Remove the key
			err = info.RemoveKey(flagKeyId)
			if err != nil {
				return NewExecError(ErrorUser, "Cannot remove the key", err)
			}

			// Store the info file
			err = store.SetInfoFile(info)
			if err != nil {
				return NewExecError(ErrorApp, "Cannot store the info file", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Key removed")

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)
	c.Flags().StringP("key", "k", "", "ID of the key to remove")
	c.MarkFlagRequired("key")

	return c
}
