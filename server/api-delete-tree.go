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
	"net/http"
	"strings"

	"github.com/ItalyPaleAle/prvt/repository"

	"github.com/gin-gonic/gin"
)

// DeleteTreeHandler is the handler for DELETE /api/tree/:path, which removes an object
// The :path value must be an exact object, or must end with "/*" to remove a folder
func (s *Server) DeleteTreeHandler(c *gin.Context) {
	// Get the path
	path := c.Param("path")
	// Ensure that the path starts with /
	// The path will be validated by the Repo.RemovePath command (in the index module)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Start a transaction
	err := s.Repo.BeginTransaction()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Remove the path
	res := make(chan repository.PathResultMessage)
	go func() {
		s.Repo.RemovePath(c.Request.Context(), path, res)
		close(res)
	}()

	// Response
	response := make([]TreeOperationResponse, 0)
	for el := range res {
		r := TreeOperationResponse{
			Path:   el.Path,
			FileId: el.FileId,
		}
		switch el.Status {
		case repository.RepositoryStatusOK:
			r.Status = "removed"
		case repository.RepositoryStatusNotFound:
			r.Status = "not-found"
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
