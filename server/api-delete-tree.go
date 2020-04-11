/*
Copyright © 2019 Alessandro Segala (@ItalyPaleAle)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, version 3 of the License.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ItalyPaleAle/prvt/repository"

	"github.com/gin-gonic/gin"
)

type itemErrorResponse struct {
	Path  string `json:"path,omitempty"`
	Error string `json:"error,omitempty"`
}

type deleteTreeResponse struct {
	Removed  []string            `json:"removed,omitempty"`
	NotFound []string            `json:"notFound,omitempty"`
	Error    []itemErrorResponse `json:"error,omitempty"`
}

// DeleteTreeHandler is the handler for DELETE /api/tree/:path, which removes an object
// The :path value must be an exact object, or must end with "/*" to remove a folder
func (s *Server) DeleteTreeHandler(c *gin.Context) {
	// Get the path (can be empty if requesting the root)
	path := c.Param("path")
	// Ensure that the path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Remove the path
	res := make(chan repository.PathResultMessage)
	go func() {
		s.Repo.RemovePath(path, res)
		close(res)
	}()

	// Response
	response := deleteTreeResponse{}
	for el := range res {
		switch el.Status {
		case repository.RepositoryStatusOK:
			if response.Removed == nil {
				response.Removed = []string{el.Path}
			} else {
				response.Removed = append(response.Removed, el.Path)
			}
		case repository.RepositoryStatusNotFound:
			if response.NotFound == nil {
				response.NotFound = []string{el.Path}
			} else {
				response.NotFound = append(response.NotFound, el.Path)
			}
		case repository.RepositoryStatusInternalError:
			errEl := itemErrorResponse{
				Path:  el.Path,
				Error: fmt.Sprintf("Internal error: %s", el.Err),
			}
			if response.Error == nil {
				response.Error = []itemErrorResponse{errEl}
			} else {
				response.Error = append(response.Error, errEl)
			}
		case repository.RepositoryStatusUserError:
			errEl := itemErrorResponse{
				Path:  el.Path,
				Error: fmt.Sprintf("Error: %s", el.Err),
			}
			if response.Error == nil {
				response.Error = []itemErrorResponse{errEl}
			} else {
				response.Error = append(response.Error, errEl)
			}
		}
	}

	c.JSON(http.StatusOK, response)
}
