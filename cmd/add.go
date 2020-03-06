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
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"e2e/fs"
	"e2e/index"
	"e2e/utils"

	"github.com/spf13/cobra"
)

func addFile(folder, target, destinationFolder string) (error, string) {
	// Get the master key and create the filesystem object
	store, err := fs.Get(storeConnectionString)
	if err != nil {
		return err, utils.ErrorUser
	}
	masterKey, err := utils.PromptMasterKey()
	if err != nil {
		return err, utils.ErrorUser
	}
	store.SetMasterKey([]byte(masterKey))
	index.Instance.SetStore(store)

	// Check if target exists
	path := filepath.Join(folder, target)
	exists, err := utils.PathExists(path)
	if err != nil {
		return err, utils.ErrorUser
	}
	if !exists {
		return errors.New("target does not exist"), utils.ErrorUser
	}

	// Check if it's a directory
	isFile, err := utils.IsRegularFile(path)
	if err != nil {
		return err, utils.ErrorUser
	}
	if !isFile {
		f, err := os.Open(path)
		if err != nil {
			return err, utils.ErrorApp
		}
		list, err := f.Readdir(-1)
		f.Close()
		for _, el := range list {
			err, errTyp := addFile(path, el.Name(), destinationFolder+target+"/")
			if err != nil {
				return err, errTyp
			}
		}
		return nil, ""
	}

	// Get a stream to the input file
	in, err := os.Open(path)
	if err != nil {
		return err, utils.ErrorApp
	}

	// Get the mime type
	extension := filepath.Ext(target)
	var mimeType string
	if extension != "" {
		mimeType = mime.TypeByExtension(extension)
	}

	// Get the size of the file
	stat, err := in.Stat()
	if err != nil {
		return err, utils.ErrorApp
	}

	// Generate a file id
	fileId, err := index.GenerateFileId()
	if err != nil {
		return err, utils.ErrorApp
	}

	// Write the data to an encrypted file
	_, err = store.Set(fileId, in, nil, target, mimeType, stat.Size())
	if err != nil {
		return err, utils.ErrorApp
	}

	// Add to the index
	err = index.Instance.AddFile(destinationFolder+target, fileId)
	if err != nil {
		return err, utils.ErrorApp
	}

	fmt.Println("Added:", destinationFolder+target)

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
				utils.ExitWithError(utils.ErrorUser, "no file or folder specified", nil)
				return
			}

			// Destination must start with "/"
			if !strings.HasPrefix(flagDestination, "/") {
				utils.ExitWithError(utils.ErrorUser, "destination must start with /", nil)
				return
			}

			// Ensure destination ends with a "/"
			if !strings.HasSuffix(flagDestination, "/") {
				flagDestination += "/"
			}

			// Iterate through the args and add them all
			for _, e := range args {
				// Get the target and folder
				folder := filepath.Dir(e)
				target := filepath.Base(e)
				err, errType := addFile(folder, target, flagDestination)
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
