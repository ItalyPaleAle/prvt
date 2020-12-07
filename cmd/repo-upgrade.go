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

// NewRepoUpgradeCmd is for "prvt repo upgrade"
func NewRepoUpgradeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade a repository",
		Long: `Upgrades a repository to the latest info file format.

Usage: "prvt repo upgrade --store <string>"
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
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err = store.AcquireLock(ctx)
			cancel()
			if err != nil {
				return NewExecError(ErrorApp, "Could not acquire a lock. Please make sure that no other instance of prvt is running with the same repo.\nIf you believe this is a mistake, you can forcefully break all locks with the \"prvt repo lock-break\" command.", err)
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

			// Upgrade the info file
			errMessage, err := UpgradeInfoFile(info)
			if err != nil {
				return NewExecError(ErrorUser, errMessage, err)
			}

			// Store the info file
			err = store.SetInfoFile(info)
			if err != nil {
				return NewExecError(ErrorApp, "Cannot store the info file", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Repository upgraded")

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)

	return c
}
