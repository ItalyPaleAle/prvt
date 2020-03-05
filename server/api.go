/*
Copyright © 2019 Alessandro Segala (@ItalyPaleAle)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, version 3 of the License.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package server

import (
	"e2e/index"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TreeHandler is the handler for GET /api/tree/:path, which returns the contents of a path
func TreeHandler(c *gin.Context) {
	// Get the path (can be empty if requesting the root)
	path := c.Param("path")
	// Ensure that the path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Get the list of files in the folder
	list, err := index.Instance.ListFolder(path)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, list)
}
