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

	"github.com/gin-gonic/gin"
)

// GetFsOptionsHandler is the handler for GET /api/fsoptions and /api/fsoptions/:fs, which returns the list of options used by the filesystems (or a specific one)
func (s *Server) GetFsOptionsHandler(c *gin.Context) {
	var res map[string]*fs.FsOptionsList

	// Check if we have a specific fs we want
	req := c.Param("fs")
	if req != "" {
		opts := fs.GetFsOptions(req)
		if opts != nil {
			res = map[string]*fs.FsOptionsList{
				req: opts,
			}
		}
	} else {
		res = fs.GetAllFsOptions()
	}

	c.JSON(http.StatusOK, res)
}
