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

// PostRepoUnlockHandler is the handler for POST /api/repo/unlock, which unlocks a repository
// The same middleware is also used by POST /api/repo/keytest, which tests a key and returns its ID (essentially performing a "dry run" for the unlock operation)
func (s *Server) PostRepoUnlockHandler(c *gin.Context) {
	// If we get here, the repository was unlocked successfully (see MiddlewareUnlockRepo)
	c.JSON(http.StatusOK, RepoKeyListItem{
		KeyId: c.GetString("KeyId"),
		Type:  c.GetString("KeyType"),
	})
}
