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
	"e2e/index"
	"errors"
	"fmt"
	"os"
	"strings"

	"e2e/crypto"
	"e2e/utils"

	"github.com/spf13/cobra"
)

func addFile(path string, destination string) (error, string) {
	// Check if file exists
	exists, err := utils.PathExists(path)
	if err != nil {
		return err, utils.ErrorUser
	}
	if !exists {
		return errors.New("File does not exist"), utils.ErrorUser
	}

	// Check if it's a directory
	isFile, err := utils.IsRegularFile(path)
	if err != nil {
		return err, utils.ErrorUser
	}
	if !isFile {
		// TODO: SCAN DIRECTORY AND RECURSIVELY DO THIS
		fmt.Println("TODO: SCAN DIRECTORY AND RECURSIVELY DO THIS")
		return nil, ""
	}

	// Get a stream to the file
	in, err := os.Open(path)
	if err != nil {
		return err, utils.ErrorApp
	}

	// Add to the index
	fileId, err := index.Instance.AddFile(destination + "fileout")
	if err != nil {
		return err, utils.ErrorApp
	}

	// Get a stream to the output
	out, err := os.Create("test/data/" + fileId)
	if err != nil {
		return err, utils.ErrorApp
	}

	// Encrypt the data
	err = crypto.EncryptFile(out, in, []byte("hello world"), "name", "image/jpeg", 0)
	if err != nil {
		return err, utils.ErrorApp
	}

	return nil, ""
}

func init() {
	var flagDestination string

	c := &cobra.Command{
		Use:               "add",
		Short:             "Add a file or folder",
		Long:              ``,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Get the file/folder name from the args
			if len(args) < 1 {
				utils.ExitWithError(utils.ErrorUser, "No file or folder specified", nil)
				return
			}

			// Destination must start with "/"
			if !strings.HasPrefix(flagDestination, "/") {
				utils.ExitWithError(utils.ErrorUser, "Destination must start with /", nil)
				return
			}

			// Ensure destination ends with a "/"
			if !strings.HasSuffix(flagDestination, "/") {
				flagDestination += "/"
			}

			// Create the destination folder
			_, err := index.Instance.AddFolder(flagDestination)
			if err != nil {
				utils.ExitWithError(utils.ErrorApp, err.Error(), err)
				return
			}

			// Iterate through the args and add them all
			for _, e := range args {
				err, errType := addFile(e, flagDestination)
				if err != nil {
					if errType == "" {
						errType = utils.ErrorApp
					}
					utils.ExitWithError(errType, err.Error(), err)
					return
				}
			}
		},
	}

	// Flags
	c.Flags().StringVarP(&flagDestination, "destination", "d", "", "destination folder")
	c.MarkFlagRequired("destination")

	// Add the command
	rootCmd.AddCommand(c)
}
