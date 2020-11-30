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

	index "github.com/ItalyPaleAle/prvt/index"

	"github.com/gin-gonic/gin"
)

// GetTreeHandler is the handler for GET /api/tree/:path, which returns the contents of a path
func (s *Server) GetTreeHandler(c *gin.Context) {
	// Get the path (can be empty if requesting the root)
	path := c.Param("path")
	// Ensure that the path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Get the list of files in the folder
	list, err := s.Repo.Index.ListFolder(0, path)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if list == nil {
		list = []index.FolderList{}
	}

	c.JSON(http.StatusOK, list)
}
