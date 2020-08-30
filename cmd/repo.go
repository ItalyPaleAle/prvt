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
	"github.com/spf13/cobra"
)

// NewRepoCmd is for "prvt repo"
func NewRepoCmd() *cobra.Command {
	c := &cobra.Command{
		Use:               "repo",
		Short:             "Create and manage a repository",
		Long:              `Commands to create and manage a repository.`,
		DisableAutoGenTag: true,
	}

	// Sub-commands
	c.AddCommand(NewRepoInitCmd())
	c.AddCommand(NewRepoKeyCmd())
	c.AddCommand(NewRepoUpgradeCmd())

	return c
}
