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

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// GetConnectionInfoHandler is the handler for GET /api/connection/:name, which returns information about the repo for a specific connection
func (s *Server) GetConnectionInfoHandler(c *gin.Context) {
	// Get the connection name
	name := c.Param("name")
	name = utils.SanitizeConnectionName(name)
	if name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Missing or invalid connection name from URL"})
		return
	}

	// Get the details of the connection
	val := viper.GetStringMapString("connections." + name)
	if val == nil || len(val) == 0 || val["type"] == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{fmt.Sprintf("Connection not found: %s", name)})
		return
	}

	// Create the store object
	store, err := fs.GetWithOptionsMap(val)
	if err != nil || store == nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Could not initialize the store"})
		return
	}

	// Request the info file
	info, err := store.GetInfoFile()
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{"Could not initialize the store"})
		return
	}
	if info == nil {
		c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{"Repository is not initialized"})
		return
	}

	// Check if we the repo can be unlocked with a GPG key
	gpgUnlock := false
	for _, k := range info.Keys {
		if k.GPGKey != "" {
			gpgUnlock = true
			break
		}
	}

	// Response is RepoInfoResponse (a subset of InfoResponse)
	repoId := info.RepoId
	if repoId == "" {
		repoId = "(Repository ID missing)"
	}
	res := RepoInfoResponse{
		StoreType:    store.FSName(),
		StoreAccount: store.AccountName(),
		RepoID:       repoId,
		RepoVersion:  info.Version,
		GPGUnlock:    gpgUnlock,
	}

	c.JSON(http.StatusOK, res)
}
