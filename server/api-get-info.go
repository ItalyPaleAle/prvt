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
	// Response object
	res := InfoResponse{
		Name:         "prvt",
		Runtime:      runtime.Version(),
		ReadOnly:     s.ReadOnly,
		RepoSelected: c.GetBool("RepoSelected"),
		RepoUnlocked: c.GetBool("RepoUnlocked"),
	}

	// If repo is selected, add repository ID and version
	if res.RepoSelected {
		// RepoId might be empty, so the value will be ignored in the resulting JSON
		res.RepoID = s.Infofile.RepoId
		res.RepoVersion = s.Infofile.Version

		// Store type and account
		res.StoreType = s.Store.FSName()
		res.StoreAccount = s.Store.AccountName()

		// If the repo is unlocked, add stats too
		if res.RepoUnlocked {
			stat, err := s.Repo.Index.Stat(0)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			res.FileCount = stat.FileCount
		}
	}

	// Add version and build info if present
	if buildinfo.BuildID == "" || buildinfo.CommitHash == "" {
		res.AppVersion = "dev"
		res.Info = "This prvt build does not contain a build identifier, and it was probably fetched from the repository as source"
	} else {
		res.AppVersion = buildinfo.AppVersion
		res.BuildID = buildinfo.BuildID
		res.BuildTime = buildinfo.BuildTime
		res.CommitHash = buildinfo.CommitHash
	}

	c.JSON(http.StatusOK, res)
}
