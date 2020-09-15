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

	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// DeleteConnectionHandler is the handler for DELETE /api/connection/:name, which removes a connection
func (s *Server) DeleteConnectionHandler(c *gin.Context) {
	// Get the connection name
	name := c.Param("name")
	name = utils.SanitizeConnectionName(name)
	if name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Missing or invalid connection name from URL"})
		return
	}

	// Check if the connection exists
	val := viper.GetStringMapString("connections." + name)
	if val == nil || len(val) == 0 || val["type"] == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{fmt.Sprintf("Connection not found: %s", name)})
		return
	}

	// Delete the connection
	viper.Set("connections."+name, map[string]string{})
	err := viper.WriteConfig()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Respond with the key ID
	c.JSON(http.StatusOK, struct {
		Removed string `json:"removed"`
	}{
		Removed: name,
	})
}
