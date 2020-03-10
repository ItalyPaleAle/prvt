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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prvt",
	Short: "Store encrypted end-to-end files and view them in your browser ",
	Long: `prvt lets you store files on local folders or on the cloud, encrypted with strong end-to-end encryption.

All commands require the address of a repository, which begins with the name of the store.

- For Azure Blob Storage: use "azure:" followed by the name of the container, for example "azure:myfiles". The container must already exist. Additionally, set the following environmental variables to authenticate with Azure Storage: "AZURE_STORAGE_ACCOUNT" with the storage account name, and "AZURE_STORAGE_ACCESS_KEY" with the storage account key.

- For AWS S3 and other S3-compatible services: use "s3:" followed by the name of the bucket, for example "s3:myfiles". The bucket must already exist. Additionally, set the following environmental variables to authenticate with S3: "AWS_ACCESS_KEY_ID" with the access key id, and "AWS_SECRET_ACCESS_KEY" with the secret access key. If you're connecting to a S3-compatible service, specify the endpoint with the "S3_ENDPOINT" environmental variable (defaults to "s3.amazonaws.com"), and optionally disable TLS with the environmental variable "S3_TLS=false".

- For storing on a local folder: use "local:" and the path to the folder (absolute or relative to the current working directory). For example: "local:/path/to/folder" or "local:subfolder-in-cwd".

Start by initializing the repository with the "prvt initrepo" command.

You can add files and folders to a repository with the "prvt add" command.

Use the "prvt serve" command to launch a local server so you can view the files with a web browser (decrypted on-the-fly).

Lastly, the "prvt rm" command lets you remove files from the repository.

prvt is open source, licensed under GNU General Public License version 3.0.
Project: https://github.com/ItalyPaleAle/prvt
`,
	DisableAutoGenTag: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(10)
	}
}
