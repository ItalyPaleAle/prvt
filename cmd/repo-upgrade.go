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

// NewRepoUpgradeCmd is for "prvt repo upgrade"
func NewRepoUpgradeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade a repository",
		Long: `Upgrades a repository to the latest info file format.

Usage: "prvt repo upgrade --store <string>"
`,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorApp, "Cannot get flag 'store'", err)
				return
			}

			// Create the store object
			store, err := fs.GetWithConnectionString(flagStoreConnectionString)
			if err != nil || store == nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorUser, "Could not initialize store", err)
				return
			}

			// Request the info file
			info, err := store.GetInfoFile()
			if err != nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorApp, "Error requesting the info file", err)
				return
			}
			if info == nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorUser, "Repository is not initialized", err)
				return
			}

			// Upgrade the info file
			errMessage, err := UpgradeInfoFile(info)
			if err != nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorUser, errMessage, err)
				return
			}

			// Store the info file
			err = store.SetInfoFile(info)
			if err != nil {
				ExitWithError(cmd.ErrOrStderr(), ErrorApp, "Cannot store the info file", err)
				return
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Repository upgraded")
		},
	}

	// Flags
	addStoreFlag(c, true)

	return c
}
