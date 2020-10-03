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

// DeleteRepoKeyHandler is the handler for DELETE /api/repo/key/:keyId, which removes a key
func (s *Server) DeleteRepoKeyHandler(c *gin.Context) {
	// Get the key ID
	keyId := c.Param("keyId")
	if keyId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Missing key ID from URL"})
		return
	}

	// Require at least 2 keys in the repository
	if len(s.Infofile.Keys) < 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"This repository has only one key, which cannot be removed"})
		return
	}

	// The key we're removing must not be the same as the key used to unlock the repository
	if keyId == s.Store.GetKeyId() {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"You cannot remove the same key you're using to unlock the repository"})
		return
	}

	// Remove the key
	err := s.Infofile.RemoveKey(keyId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Store the info file
	err = s.Store.SetInfoFile(s.Infofile)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Respond with the key ID
	c.JSON(http.StatusOK, struct {
		Removed string `json:"removed"`
	}{
		Removed: keyId,
	})
}
