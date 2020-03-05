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
	"e2e/fs"
	"e2e/server"
	"e2e/utils"

	"github.com/spf13/cobra"
)

// serveCmd represents the auth command
var serveCmd = &cobra.Command{
	Use:               "serve",
	Short:             "Start the server",
	Long:              ``,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the master key and create the filesystem object
		store := &fs.Local{}
		store.SetMasterKey([]byte("hello world"))

		// Start the server
		srv := server.Server{
			Store: store,
		}
		err := srv.Start()
		if err != nil {
			utils.ExitWithError(utils.ErrorApp, "Could not start server", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
