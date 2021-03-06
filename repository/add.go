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
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/utils"

	mime "github.com/cubewise-code/go-mime"
	"github.com/gofrs/uuid"
)

// Flag that controls whether SHA-256 digests are calculated for files
var CalculateDigest = true

// AddStream adds a document to the repository by reading it from a stream
func (repo *Repository) AddStream(ctx context.Context, in io.ReadCloser, filename, destinationFolder, mimeType string, size int64, force bool) (fileIdStr string, status int, err error) {
	// Generate a file id
	fileId, err := uuid.NewV4()
	if err != nil {
		return "", RepositoryStatusInternalError, err
	}

	// Sanitize the file name added to the index
	sanitizedFilename := utils.SanitizePath(filename)
	sanitizedPath := utils.SanitizePath(destinationFolder + filename)

	// Sanitize the mime type
	mimeType = utils.SanitizeMimeType(mimeType)

	// Check if the file exists in the index already
	existing, err := repo.Index.GetFileByPath(repo.tx, sanitizedPath)
	if err != nil {
		return "", RepositoryStatusInternalError, err
	}
	// Unless we're force-adding the file, return an error if the file already exists
	// Path "/" always exists and is always an error
	if (!force && existing != nil) || sanitizedPath == "/" {
		return "", RepositoryStatusExisting, nil
	}

	// Create the hash
	hash := sha256.New()

	// Tee the in stream into the hash
	tee := io.TeeReader(in, hash)

	// Write the data to an encrypted file
	metadata := &crypto.Metadata{
		Name:        sanitizedFilename,
		ContentType: mimeType,
		Size:        size,
	}
	_, err = repo.Store.Set(ctx, fileId.String(), tee, nil, metadata)
	if err != nil {
		return "", RepositoryStatusInternalError, err
	}

	// Complete the file's digest
	digest := hash.Sum(nil)

	// If there was an existing file, now it's time to remove it
	if existing != nil {
		// Remove it from the index
		var objs []string
		objs, _, err = repo.Index.DeleteFile(repo.tx, sanitizedPath)
		if err != nil {
			return "", RepositoryStatusInternalError, err
		}
		if len(objs) != 1 {
			return "", RepositoryStatusInternalError, errors.New("invalid number of files removed from index")
		}
		// Delete the file data
		err = repo.Store.Delete(ctx, objs[0], nil)
		if err != nil {
			return "", RepositoryStatusInternalError, err
		}
	}

	// Add to the index
	err = repo.Index.AddFile(repo.tx, sanitizedPath, fileId.Bytes(), mimeType, size, digest, force)
	if err != nil {
		return "", RepositoryStatusInternalError, err
	}

	return fileId.String(), RepositoryStatusOK, nil
}

// AddFile adds a file to the repository
// This accepts any regular file, and it does not ignore any file
func (repo *Repository) AddFile(ctx context.Context, folder, target, destinationFolder string, force bool) (fileIdStr string, status int, err error) {
	path := filepath.Join(folder, target)

	// Check if target exists and it's a regular file
	exists, err := utils.IsRegularFile(path)
	if err != nil {
		return "", RepositoryStatusInternalError, err
	}
	if !exists {
		return "", RepositoryStatusUserError, errors.New("target does not exist: " + target)
	}

	// Get a stream to the input file
	in, err := os.Open(path)
	if err != nil {
		return "", RepositoryStatusInternalError, err
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
		return "", RepositoryStatusInternalError, err
	}
	size := stat.Size()

	// Add the file's stream
	return repo.AddStream(ctx, in, target, destinationFolder, mimeType, size, force)
}

// AddPath adds a path (a file or a folder, recursively) and reports each element added in the res channel
func (repo *Repository) AddPath(ctx context.Context, folder, target, destinationFolder string, force bool, res chan<- PathResultMessage) {
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
		fileId, status, err := repo.AddFile(ctx, folder, target, destinationFolder, force)
		res <- PathResultMessage{
			Path:   destinationFolder + target,
			Status: status,
			FileId: fileId,
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
			repo.AddPath(ctx, path, el.Name(), destinationFolder+target+"/", force, res)
		}
	}

	return
}
