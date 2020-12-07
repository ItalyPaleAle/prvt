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

	"github.com/ItalyPaleAle/prvt/fs"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// NewRepoLockBreakCmd is for "prvt repo lock-break"
func NewRepoLockBreakCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "lock-break",
		Short: "Forcefully remove all locks on a repository",
		Long: `DATA LOSS WARNING: USE THIS COMMAND WITH CAUTION

This command forcefully removes all locks placed on a repository, allowing the usage of a repository that was otherwise locked. 

Most prvt commands acquire an exclusive lock on the repository before performing operations that would change the index, to forbid other instances of prvt from accessing the same repository and so to preserve the integrity of the data. Locks are generally removed automatically once the comamnd ends or the app is closed; however, in situations such as when the app suddenly crashes, repositories may remain in a locked state.

When that happens, using the "prvt repo break-locks --store <string>" command can help by removing all locks placed on a repository.

This command should ONLY be invoked if you're sure that no other instance of prvt is accessing the data in the repository. Forcefully unlocking a repository that is in use by another instance of prvt could cause the index to be corrupted and data loss.
`,
		DisableAutoGenTag: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'store'", err)
			}

			// Ask for confirmation
			prompt := promptui.Prompt{
				Label:     "Force break locks for repository " + flagStoreConnectionString,
				IsConfirm: true,
				Stdin:     PromptuiStdin,
				Stdout:    PromptuiStdout,
			}
			result, err := prompt.Run()
			if err != nil || (result != "y" && result != "Y") {
				// The prompt returns an error even when the user selects "n", so just abort
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted")
				return nil
			}

			// Create the store object
			// No need for a lock for this command
			store, err := fs.GetWithConnectionString(flagStoreConnectionString)
			if err != nil || store == nil {
				return NewExecError(ErrorUser, "Could not initialize store", err)
			}

			// Remove all locks
			err = store.BreakLock(context.Background())
			if err != nil {
				return NewExecError(ErrorApp, "Could not break locks on the repository", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "All locks on the repository have been removed")

			return nil
		},
	}

	// Flags
	addStoreFlag(c, true)

	return c
}
