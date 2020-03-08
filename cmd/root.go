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
	"os"

	"github.com/spf13/cobra"
)

// Store connection string
var storeConnectionString string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "prvt",
	Short:             "",
	Long:              ``,
	DisableAutoGenTag: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&storeConnectionString, "store", "s", "", "connection string for the store")
	rootCmd.MarkFlagFilename("store")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(10)
	}
}
