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
	"io"
	"os"
	"path/filepath"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/utils"

	mime "github.com/cubewise-code/go-mime"
	"github.com/gofrs/uuid"
)

// AddStream adds a document to the repository by reading it from a stream
func (repo *Repository) AddStream(in io.ReadCloser, filename, destinationFolder, mimeType string, size int64) (int, error) {
	// Generate a file id
	fileId, err := uuid.NewV4()
	if err != nil {
		return RepositoryStatusInternalError, err
	}

	// Sanitize the file name added to the index
	sanitizedFilename := utils.SanitizePath(filename)
	sanitizedPath := utils.SanitizePath(destinationFolder + filename)

	// Sanitize the mime type
	mimeType = utils.SanitizeMimeType(mimeType)

	// Check if the file exists in the index already
	exists, err := index.Instance.FileExists(sanitizedPath)
	if err != nil {
		return RepositoryStatusInternalError, err
	}
	if exists {
		return RepositoryStatusExisting, nil
	}

	// Write the data to an encrypted file
	metadata := &crypto.Metadata{
		Name:        sanitizedFilename,
		ContentType: mimeType,
		Size:        size,
	}
	_, err = repo.Store.Set(fileId.String(), in, nil, metadata)
	if err != nil {
		return RepositoryStatusInternalError, err
	}

	// Add to the index
	err = index.Instance.AddFile(sanitizedPath, fileId.Bytes(), mimeType)
	if err != nil {
		return RepositoryStatusInternalError, err
	}

	return RepositoryStatusOK, nil
}

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

	// Get a stream to the input file
	in, err := os.Open(path)
	if err != nil {
		return RepositoryStatusInternalError, err
	}
	defer in.Close()

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
	size := stat.Size()

	// Add the file's stream
	return repo.AddStream(in, target, destinationFolder, mimeType, size)
}

// AddPath adds a path (a file or a folder, recursively) and reports each element added in the res channel
func (repo *Repository) AddPath(folder, target, destinationFolder string, res chan<- PathResultMessage) {
	path := filepath.Join(folder, target)

	// Check if target exists
	exists, err := utils.PathExists(path)
	if err != nil {
		res <- PathResultMessage{
			Path:   destinationFolder + target,
			Status: RepositoryStatusInternalError,
			Err:    err,
		}
		return
	}
	if !exists {
		res <- PathResultMessage{
			Path:   destinationFolder + target,
			Status: RepositoryStatusUserError,
			Err:    errors.New("target does not exist"),
		}
		return
	}

	// Check if we should ignore this path
	if utils.IsIgnoredFile(path) {
		res <- PathResultMessage{
			Path:   destinationFolder + target,
			Status: RepositoryStatusIgnored,
		}
		return
	}

	// Check if it's a directory
	isFile, err := utils.IsRegularFile(path)
	if err != nil {
		res <- PathResultMessage{
			Path:   destinationFolder + target,
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
		// Do not defer the call to Close, or it will be closed at the end of the function, after the recursion
		f, err := os.Open(path)
		if err != nil {
			res <- PathResultMessage{
				Path:   destinationFolder + target,
				Status: RepositoryStatusInternalError,
				Err:    err,
			}
			return
		}

		list, err := f.Readdir(-1)
		f.Close() // Close here
		if err != nil {
			res <- PathResultMessage{
				Path:   destinationFolder + target,
				Status: RepositoryStatusInternalError,
				Err:    err,
			}
			return
		}
		for _, el := range list {
			// Recursion
			repo.AddPath(path, el.Name(), destinationFolder+target+"/", res)
		}
	}

	return
}
