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
	"time"

	"github.com/ItalyPaleAle/prvt/fs"

	"github.com/spf13/cobra"
)

// NewRepoKeyTestCmd is for "prvt repo key test"
func NewRepoKeyTestCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "test",
		Short: "Test a key for unlocking the repo",
		Long: `Tests a key and returns the ID of the key used to unlock the repo.

Usage: "prvt repo key test --store <string>"

This command is particularly useful to determine the ID of a key that you want to remove.
`,
		DisableAutoGenTag: true,

		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Unlock the repository
			_, keyId, errMessage, err := GetMasterKey(info)
			if err != nil {
				return NewExecError(ErrorUser, errMessage, err)
			}

			// Show the key ID
			fmt.Fprintln(cmd.OutOrStdout(), "Repository unlocked using key with ID:", keyId)

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)

	return c
}
