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
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PostRepoCloseHandler is the handler for POST /api/repo/close, which closes any open repository
func (s *Server) PostRepoCloseHandler(c *gin.Context) {
	// If there's an existing store object, release locks (if any)
	err := s.releaseRepoLock()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Reset the store, infofile, and repo variables
	s.Store = nil
	s.Infofile = nil
	s.Repo = nil

	fmt.Fprintln(s.LogWriter, "Repository closed")

	// Respond with a 200 OK
	c.Status(http.StatusOK)
}
