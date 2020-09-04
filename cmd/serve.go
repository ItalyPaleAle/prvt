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
	"errors"

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/repository"
	"github.com/ItalyPaleAle/prvt/server"

	"github.com/spf13/cobra"
)

// NewServeCmd is for "prvt serve"
func NewServeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Long: `Starts a web server on the local machine, so you can access your encrypted files using a web browser.

Usage: "prvt serve --store <string>"

You can use the optional "--address" and "--port" flags to control what address and port the server listens on. To enable connections from remote clients (not running on the local machine), set the address to "0.0.0.0".
`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				repo  *repository.Repository
				store fs.Fs
				info  *infofile.InfoFile
				err   error
			)

			// Flags
			flagStoreConnectionString, err := cmd.Flags().GetString("store")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'store'", err)
			}
			flagBindPort, err := cmd.Flags().GetString("port")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'port'", err)
			}
			flagBindAddress, err := cmd.Flags().GetString("address")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'address'", err)
			}
			flagVerbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'verbose'", err)
			}
			flagNoUnlock, err := cmd.Flags().GetBool("no-unlock")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'no-unlock'", err)
			}
			flagNoRepo, err := cmd.Flags().GetBool("no-repo")
			if err != nil {
				return NewExecError(ErrorApp, "Cannot get flag 'no-repo'", err)
			}

			// Check if we have a store flag
			if !flagNoRepo {
				// Ensure the connection string is set
				if flagStoreConnectionString == "" {
					return NewExecError(ErrorUser, "Missing store connection string", errors.New("Use the '--store' flag to pass a store when '--no-repo' is not set."))
				}

				// Create the store object
				store, err = fs.GetWithConnectionString(flagStoreConnectionString)
				if err != nil || store == nil {
					return NewExecError(ErrorUser, "Could not initialize store", err)
				}

				// Request the info file
				info, err = store.GetInfoFile()
				if err != nil {
					return NewExecError(ErrorApp, "Error requesting the info file", err)
				}
				if info == nil {
					return NewExecError(ErrorUser, "Repository is not initialized", err)
				}

				// Unlock the repo if needed
				if !flagNoUnlock {
					// Derive the master key
					masterKey, keyId, errMessage, err := GetMasterKey(info)
					if err != nil {
						return NewExecError(ErrorUser, errMessage, err)
					}
					store.SetMasterKey(keyId, masterKey)

					// Set up the index
					index.Instance.SetStore(store)

					// Set up the repository
					repo = &repository.Repository{
						Store: store,
					}
				}
			}

			// Start the server
			srv := server.Server{
				Store:     store,
				Verbose:   flagVerbose,
				Repo:      repo,
				Infofile:  info,
				LogWriter: cmd.OutOrStdout(),
			}
			err = srv.Start(cmd.Context(), flagBindAddress, flagBindPort)
			if err != nil {
				return NewExecError(ErrorApp, "Could not start server", err)
			}

			return nil
		},
	}

	// Flags
	addStoreFlag(c, false)
	c.Flags().StringP("address", "a", "127.0.0.1", "address to bind to")
	c.Flags().StringP("port", "p", "3129", "port to bind to")
	c.Flags().BoolP("verbose", "v", false, "show request log")
	c.Flags().Bool("no-unlock", false, "do not unlock the repo")
	c.Flags().Bool("no-repo", false, "do not connect to a repository")

	return c
}
