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

// MiddlewareRepoStatus adds the status of the repository to the context
// This is required by MiddlewareRequireRepo and MiddlewareRequireUnlock
func (s *Server) MiddlewareRepoStatus(c *gin.Context) {
	repoSelected := s.Store != nil && s.Infofile != nil
	repoUnlocked := repoSelected && s.Repo != nil && s.Repo.Index != nil

	// Set in the context
	c.Set("RepoSelected", repoSelected)
	c.Set("RepoUnlocked", repoUnlocked)
}

// MiddlewareRequireRepo requires a repository to be selected (even if not unlocked)
func (s *Server) MiddlewareRequireRepo(c *gin.Context) {
	if !c.GetBool("RepoSelected") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{"No repository has been selected yet."})
	}
}

// MiddlewareRequireUnlock requires the repository to be unlocked before processing
func (s *Server) MiddlewareRequireUnlock(c *gin.Context) {
	if !c.GetBool("RepoUnlocked") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{"The repository has not been unlocked."})
	}
}

// MiddlewareRequireInfoFileVersion requires a minimum version of the info file to continue
func (s *Server) MiddlewareRequireInfoFileVersion(version uint16) func(c *gin.Context) {
	return func(c *gin.Context) {
		if s.Infofile.Version < version {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, ErrorResponse{`This repository needs to be upgraded. Please run "prvt repo upgrade --store <string>" to upgrade this repository to the latest format.`})
		}
	}
}

// MiddlewareRequireReadWrite disables a route that would edit the repository if we're operating in read-only mode
func (s *Server) MiddlewareRequireReadWrite(c *gin.Context) {
	if s.ReadOnly {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, ErrorResponse{`This action is not allowed when the repository is opened in read-only mode.`})
	}
}
