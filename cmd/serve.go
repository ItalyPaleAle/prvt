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
	"bytes"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/server"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/spf13/cobra"
)

func init() {
	var (
		flagBindPort    string
		flagBindAddress string
	)

	c := &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Long: `Starts a web server on the local machine, so you can access your encrypted files using a web browser.

Usage: "prvt serve --store <string>"

You can use the optional "--address" and "--port" flags to control what address and port the server listens on. To enable connections from remote clients (not running on the local machine), set the address to "0.0.0.0".
`,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Create the store object
			store, err := fs.Get(storeConnectionString)
			if err != nil || store == nil {
				utils.ExitWithError(utils.ErrorUser, "Could not initialize store", err)
				return
			}

			// Request the info file
			info, err := store.GetInfoFile()
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Error requesting the info file", err)
				return
			}
			if info == nil {
				utils.ExitWithError(utils.ErrorUser, "Store is not initialized", err)
				return
			}

			// Get the passphrase and derive the master key
			passphrase, err := utils.PromptPassphrase()
			if err != nil {
				utils.ExitWithError(utils.ErrorUser, "Error getting passphrase", err)
				return
			}
			masterKey, confirmationHash, err := crypto.KeyFromPassphrase(passphrase, info.Salt)
			if bytes.Compare(info.ConfirmationHash, confirmationHash) != 0 {
				utils.ExitWithError(utils.ErrorUser, "Invalid passphrase", err)
				return
			}
			store.SetMasterKey(masterKey)

			// Set up the index
			index.Instance.SetStore(store)

			// Start the server
			srv := server.Server{
				Store: store,
			}
			err = srv.Start(flagBindAddress, flagBindPort)
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, "Could not start server", err)
				return
			}
		},
	}
	rootCmd.AddCommand(c)

	// Flags
	c.Flags().StringVarP(&flagBindAddress, "address", "a", "127.0.0.1", "address to bind to")
	c.Flags().StringVarP(&flagBindPort, "port", "p", "3129", "port to bind to")
}
