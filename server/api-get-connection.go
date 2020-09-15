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
	"regexp"

	"github.com/ItalyPaleAle/prvt/fs"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var connectionKeyRegex = regexp.MustCompile("^connections\\.([a-z0-9-_]+)(|\\.type)$")

// GetConnectionHandler is the handler for GET /api/connection, which returns the list of connections saved
func (s *Server) GetConnectionHandler(c *gin.Context) {
	// Get the list of connections
	connections := ConnectionList{}
	for _, k := range viper.AllKeys() {
		// Get names of connections
		var name string
		if match := connectionKeyRegex.FindStringSubmatch(k); match != nil && len(match) > 1 {
			name = match[1]
		} else {
			continue
		}

		// Get the details and account name
		v := viper.GetStringMapString("connections." + name)
		store, err := fs.GetWithOptionsMap(v)
		if err != nil {
			continue
		}

		// Add the element
		connections[name] = ConnectionListItem{
			Type:    v["type"],
			Account: store.AccountName(),
		}
	}

	c.JSON(http.StatusOK, connections)
}
