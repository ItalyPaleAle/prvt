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
	"runtime"

	"github.com/ItalyPaleAle/prvt/buildinfo"

	"github.com/gin-gonic/gin"
)

// GetInfoHandler is the handler for GET /api/info, which returns the info for the app
func (s *Server) GetInfoHandler(c *gin.Context) {
	if buildinfo.BuildID == "" || buildinfo.CommitHash == "" {
		c.JSON(http.StatusOK, InfoResponse{
			Name:       "prvt",
			AppVersion: "canary",
			Info:       "This prvt build does not contain a build identifier, and it was probably fetched from the repository as source",
			Runtime:    runtime.Version(),
			ReadOnly:   s.ReadOnly,
		})
	} else {
		c.JSON(http.StatusOK, InfoResponse{
			Name:       "prvt",
			AppVersion: buildinfo.AppVersion,
			BuildID:    buildinfo.BuildID,
			BuildTime:  buildinfo.BuildTime,
			CommitHash: buildinfo.CommitHash,
			Runtime:    runtime.Version(),
			ReadOnly:   s.ReadOnly,
		})
	}
}
