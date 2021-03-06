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

package server

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ItalyPaleAle/prvt/repository"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/go-homedir"
)

// PostTreeHandler is the handler for POST /api/tree/:path, which adds documents to the repository
// The path argument is the destination folder (just like the "-d" argument in the "prvt add" command)
// The post body can contain either one of:
// - One or more files transmitted in the request body, in the "file" field(s)
// - The path to a file or folder in the local filesystem, in the "localpath" field(s)
func (s *Server) PostTreeHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Get the path (can be empty if targeting the root)
	path := c.Param("path")
	// Ensure that the path starts with / and ends with "/"
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Get the data from the request body, which must be a multipart/form-data
	mpf, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Check if we're forcing the request
	force := false
	forceVal, ok := mpf.Value["force"]
	if ok && len(forceVal) > 0 {
		force = utils.IsTruthy(forceVal[0])
	}

	// Create a channel to listen to the responses
	res := make(chan repository.PathResultMessage)

	// Check if we have a path from the local filesystem or a file uploaded
	uploadFiles := mpf.File["file"]
	localPaths := mpf.Value["localpath"]
	var method func()
	if localPaths != nil && len(localPaths) > 0 && (uploadFiles == nil || len(uploadFiles) == 0) {
		method = func() {
			s.addLocalPath(ctx, localPaths, path, force, res)
		}
	} else if uploadFiles != nil && len(uploadFiles) > 0 && (localPaths == nil || len(localPaths) == 0) {
		method = func() {
			s.addUploadedFiles(ctx, uploadFiles, path, force, res)
		}
	} else {
		c.AbortWithError(http.StatusBadRequest, errors.New("need to specify one and only one of 'file' or 'localpath' form fields"))
		return
	}

	// Start a transaction
	err = s.Repo.BeginTransaction()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Add the file(s) in a goroutine
	go func() {
		method()
		close(res)
	}()

	// Read each response from the channel
	response := make([]TreeOperationResponse, 0)
	for el := range res {
		r := TreeOperationResponse{
			Path:   el.Path,
			FileId: el.FileId,
		}
		switch el.Status {
		case repository.RepositoryStatusOK:
			r.Status = "added"
		case repository.RepositoryStatusExisting:
			r.Status = "existing"
		case repository.RepositoryStatusIgnored:
			r.Status = "ignored"
		case repository.RepositoryStatusInternalError:
			r.Status = "internal-error"
			r.Error = el.Err.Error()
		case repository.RepositoryStatusUserError:
			r.Status = "error"
			r.Error = el.Err.Error()
		}
		response = append(response, r)
	}

	// Commit the transaction
	err = s.Repo.CommitTransaction()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Adds files from the local filesystem, passing the path
func (s *Server) addLocalPath(ctx context.Context, paths []string, destination string, force bool, res chan<- repository.PathResultMessage) {
	// Iterate through the paths and add them all
	var err error
	var expanded string
	for _, e := range paths {
		// Get the target and folder
		expanded, err = homedir.Expand(e)
		if err != nil {
			res <- repository.PathResultMessage{
				Path:   e,
				Status: repository.RepositoryStatusInternalError,
				Err:    err,
			}
			break
		}
		folder := filepath.Dir(expanded)
		target := filepath.Base(expanded)

		s.Repo.AddPath(ctx, folder, target, destination, force, res)
	}
}

// Add multiple files by streams
func (s *Server) addUploadedFiles(ctx context.Context, uploadFiles []*multipart.FileHeader, destination string, force bool, res chan<- repository.PathResultMessage) {
	for _, f := range uploadFiles {
		s.addUploadedFile(ctx, f, destination, force, res)
	}
}

// Add a file by a stream
func (s *Server) addUploadedFile(ctx context.Context, uploadFile *multipart.FileHeader, destination string, force bool, res chan<- repository.PathResultMessage) {
	// Filename
	filename := filepath.Base(uploadFile.Filename)
	if filename == "" || filename == ".." || filename == "." || filename == "/" {
		res <- repository.PathResultMessage{
			Status: repository.RepositoryStatusUserError,
			Err:    errors.New("invalid filename"),
		}
		return
	}

	// Mime type
	// Might be left empty
	mime := uploadFile.Header.Get("Content-Type")

	// Size
	size := uploadFile.Size
	if size < 1 {
		res <- repository.PathResultMessage{
			Status: repository.RepositoryStatusUserError,
			Err:    errors.New("invalid file size"),
		}
		return
	}

	// Stream to the file
	in, err := uploadFile.Open()
	if err != nil {
		res <- repository.PathResultMessage{
			Path:   destination + filename,
			Status: repository.RepositoryStatusInternalError,
			Err:    err,
		}
		return
	}

	// Add the file
	fileId, result, err := s.Repo.AddStream(ctx, in, filename, destination, mime, size, force)
	res <- repository.PathResultMessage{
		Path:   destination + filename,
		Status: result,
		FileId: fileId,
		Err:    err,
	}
}
