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

// NewRepoKeyCmd is for "prvt repo key"
func NewRepoKeyCmd() *cobra.Command {
	c := &cobra.Command{
		Use:               "key",
		Short:             "Manage keys that can unlock the repository",
		Long:              `Commands to add and remove keys that can unlock the repository.`,
		DisableAutoGenTag: true,
	}

	// Sub-commands
	c.AddCommand(NewRepoKeyAddCmd())
	c.AddCommand(NewRepoKeyLsCmd())
	c.AddCommand(NewRepoKeyRmCmd())
	c.AddCommand(NewRepoKeyTestCmd())

	return c
}
