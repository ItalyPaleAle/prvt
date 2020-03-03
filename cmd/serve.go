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
	"log"
	"net/http"

	"e2e/fs"

	"github.com/spf13/cobra"
)

// serveCmd represents the auth command
var serveCmd = &cobra.Command{
	Use:               "serve",
	Short:             "Start the server",
	Long:              ``,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Create the file server
		baseFs := fs.OsFs{}
		httpFs := fs.NewHttpFs(baseFs)
		fileserver := http.FileServer(httpFs.Dir("test/"))
		http.Handle("/", fileserver)

		// Create and listen to the web server
		err := http.ListenAndServe("127.0.0.1:3000", nil)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
