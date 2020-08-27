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

// GetRepoKeysHandler is the handler for GET /api/repo/key, which returns the list of keys allowed to unlock this repository
func (s *Server) GetRepoKeysHandler(c *gin.Context) {
	result := repoKeyListResponse{
		PassphrasesCount: 0,
		GPGKeys:          make([]string, 0),
	}

	// Iterate through the keys
	for _, k := range s.Infofile.Keys {
		if k.GPGKey == "" {
			result.PassphrasesCount++
		} else {
			result.GPGKeys = append(result.GPGKeys, k.GPGKey)
		}
	}
	c.JSON(http.StatusOK, result)
}
