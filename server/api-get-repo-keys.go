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
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRepoKeysHandler is the handler for GET /api/repo/key, which returns the list of keys allowed to unlock this repository
func (s *Server) GetRepoKeysHandler(c *gin.Context) {
	result := repoKeyListResponse{}

	if s.Infofile != nil && s.Infofile.Keys != nil && len(s.Infofile.Keys) > 0 {
		result.Keys = make([]repoKeyListItem, len(s.Infofile.Keys))

		// Iterate through the keys
		for i, k := range s.Infofile.Keys {
			// Get the key id and type
			item := repoKeyListItem{}
			if k.GPGKey != "" {
				item.KeyID = k.GPGKey
				item.Type = "gpg"
			} else {
				hash := sha256.Sum256(k.MasterKey)
				item.KeyID = fmt.Sprintf("p:%X", hash[0:8])
				item.Type = "passphrase"
			}

			// Add the key
			result.Keys[i] = item
		}
	}
	c.JSON(http.StatusOK, result)
}
