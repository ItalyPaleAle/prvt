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

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// PostConnectionHandler is the handler for POST /api/connection, which stores a new connection
func (s *Server) PostConnectionHandler(c *gin.Context) {
	// Get a set of key-values from the body
	args := make(map[string]string)
	if err := c.Bind(&args); err != nil || len(args) == 0 {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Could not parse response body"})
		return
	}

	// Ensure keys name and type are present and valid
	if args["name"] == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Key 'name' is required"})
		return
	}
	name := utils.SanitizeConnectionName(args["name"])
	if name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Value of 'name' is invalid"})
		return
	}
	delete(args, "name")
	if args["type"] == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Key 'type' is required"})
		return
	}

	// Create the store object to test if the args are correct
	store, err := fs.GetWithOptionsMap(args)
	if err != nil || store == nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Could not initialize the store"})
		return
	}

	// Save the connection
	viper.Set("connections."+name, args)
	err = viper.WriteConfig()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ConnectionListItem{
		Type:    args["type"],
		Account: store.AccountName(),
	})
}
