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

	"github.com/gin-gonic/gin"
)

// GetInfoHandler is the handler for GET /api/repo/info, which returns info about the repository
func (s *Server) GetRepoInfoHandler(c *gin.Context) {
	response := RepoInfoResponse{
		Version: s.Infofile.Version,
	}

	// Count files if we have an index
	if s.Repo != nil && s.Repo.Index != nil {
		stat, err := s.Repo.Index.Stat()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		response.FileCount = stat.FileCount
	}

	c.JSON(http.StatusOK, response)
}
