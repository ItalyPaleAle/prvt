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

	"github.com/ItalyPaleAle/prvt/keys"

	"github.com/gin-gonic/gin"
)

// PostRepoKeyHandler is the handler for POST /api/repo/key, which adds a new key
func (s *Server) PostRepoKeyHandler(c *gin.Context) {
	// Get the new key
	args := &AddKeyRequest{}
	if ok := args.FromBody(c); !ok {
		return
	}

	// Add the key
	var (
		keyId      string
		errMessage string
		err        error
	)
	if args.GPGKeyId != "" {
		keyId, errMessage, err = keys.AddKeyGPG(s.Infofile, s.Store.GetMasterKey(), args.GPGKeyId)
	} else {
		keyId, errMessage, err = keys.AddKeyPassphrase(s.Infofile, s.Store.GetMasterKey(), args.Passphrase)
	}
	if err != nil {
		msg := fmt.Sprintf("%s: %s", err, errMessage)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{msg})
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
		KeyId string `json:"keyId"`
	}{
		KeyId: keyId,
	})
}
