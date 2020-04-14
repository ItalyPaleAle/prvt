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

// Requires a minimum version of the info file to continue
func (s *Server) MiddlewareRequireInfoFileVersion(version uint16) func(c *gin.Context) {
	return func(c *gin.Context) {
		if s.Infofile.Version < version {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, map[string]string{"error": `This repository needs to be upgraded. Please run "prvt repo upgrade --store <string>" to upgrade this repository to the latest format`})
		}
	}
}
