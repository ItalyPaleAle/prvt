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

package repository

import (
	"errors"
	"mime"
	"os"
	"path/filepath"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/utils"
)

// AddFile adds a file to the repository
// This accepts any regular file, and it does not ignore any file
func (repo *Repository) AddFile(folder, target, destinationFolder string) (int, error) {
	path := filepath.Join(folder, target)

	// Check if target exists and it's a regular file
	exists, err := utils.IsRegularFile(path)
	if err != nil {
		return RepositoryStatusInternalError, err
	}
	if !exists {
		return RepositoryStatusUserError, errors.New("target does not exist: " + target)
	}

	// Generate a file id
	fileId, err := index.GenerateFileId()
	if err != nil {
		return RepositoryStatusInternalError, err
	}

	// Sanitize the file name added to the index
	sanitizedTarget := utils.SanitizePath(target)
	sanitizedPath := utils.SanitizePath(destinationFolder + target)

	// Check if the file exists in the index already
	exists, err = index.Instance.FileExists(sanitizedPath)
	if err != nil {
		return RepositoryStatusInternalError, err
	}
	if exists {
		return RepositoryStatusExisting, nil
	}

	// Get a stream to the input file
	in, err := os.Open(path)
	if err != nil {
		return RepositoryStatusInternalError, err
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
		return RepositoryStatusInternalError, err
	}

	// Write the data to an encrypted file
	metadata := &crypto.Metadata{
		Name:        sanitizedTarget,
		ContentType: mimeType,
		Size:        stat.Size(),
	}
	_, err = repo.Store.Set(fileId, in, nil, metadata)
	if err != nil {
		return RepositoryStatusInternalError, err
	}

	// Add to the index
	err = index.Instance.AddFile(sanitizedPath, fileId)
	if err != nil {
		return RepositoryStatusInternalError, err
	}

	return RepositoryStatusOK, nil
}

// AddPath adds a path (a file or a folder, recursively) and reports each element added in the res channel
func (repo *Repository) AddPath(folder, target, destinationFolder string, res chan<- PathResultMessage) {
	path := filepath.Join(folder, target)

	// Check if target exists
	exists, err := utils.PathExists(path)
	if err != nil {
		res <- PathResultMessage{
			Path:   path,
			Status: RepositoryStatusInternalError,
			Err:    err,
		}
		return
	}
	if !exists {
		res <- PathResultMessage{
			Path:   path,
			Status: RepositoryStatusUserError,
			Err:    errors.New("target does not exist"),
		}
		return
	}

	// Check if we should ignore this path
	if utils.IsIgnoredFile(path) {
		res <- PathResultMessage{
			Path:   path,
			Status: RepositoryStatusIgnored,
		}
		return
	}

	// Check if it's a directory
	isFile, err := utils.IsRegularFile(path)
	if err != nil {
		res <- PathResultMessage{
			Path:   path,
			Status: RepositoryStatusInternalError,
			Err:    err,
		}
		return
	}

	// For files, add that
	if isFile {
		status, err := repo.AddFile(folder, target, destinationFolder)
		res <- PathResultMessage{
			Path:   destinationFolder + target,
			Status: status,
			Err:    err,
		}
	} else {
		// Recursively read all the elements in the directory
		f, err := os.Open(path)
		if err != nil {
			res <- PathResultMessage{
				Path:   path,
				Status: RepositoryStatusInternalError,
				Err:    err,
			}
			return
		}
		list, err := f.Readdir(-1)
		f.Close()
		for _, el := range list {
			// Recursion
			repo.AddPath(path, el.Name(), destinationFolder+target+"/", res)
		}
	}

	return
}
