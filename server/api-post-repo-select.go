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
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// PostRepoSelectHandler is the handler for POST /api/repo/select, which selects a repository
func (s *Server) PostRepoSelectHandler(c *gin.Context) {
	// Get a set of key-values from the body
	args := make(map[string]string)
	if err := c.Bind(&args); err != nil || len(args) == 0 {
		if err != nil {
			c.Error(err)
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Could not parse request body"})
		return
	}

	// Check if we have a name key, which would be the name of a saved connection
	name, ok := args["name"]
	if ok {
		// Sanitize the name
		name = utils.SanitizeConnectionName(name)
		if name == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Value of 'name' is invalid"})
			return
		}

		// Load the connection and check if it exists
		args = viper.GetStringMapString("connections." + name)
		if args == nil || len(args) == 0 || args["type"] == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{fmt.Sprintf("Connection not found: %s", name)})
			return
		}
	} else {
		// Get the storage type
		if args["type"] == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Key 'type' is required"})
			return
		}
	}

	// Create the store object
	store, err := fs.GetWithOptionsMap(args)
	if err != nil || store == nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Could not initialize the store"})
		return
	}

	// If there's an existing store object, release locks (if any)
	err = s.releaseRepoLock()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Acquire a lock
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	err = store.AcquireLock(ctx)
	cancel()
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusConflict, "Could not acquire a lock. Please make sure that no other instance of prvt is running with the same repo.\nIf you believe this is a mistake, you can forcefully break all locks with the \"prvt repo lock-break\" command.")
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

	// Set the store and info file objects
	// Note that the repo is still locked at this stage
	s.Store = store
	s.Infofile = info
	s.Repo = nil

	// Check if we the repo can be unlocked with a GPG key
	gpgUnlock := false
	for _, k := range s.Infofile.Keys {
		if k.GPGKey != "" {
			gpgUnlock = true
			break
		}
	}

	// Response
	repoId := s.Infofile.RepoId
	if repoId == "" {
		repoId = "(Repository ID missing)"
	}
	fmt.Fprintln(s.LogWriter, "Selected repository:", repoId)
	c.JSON(http.StatusOK, RepoInfoResponse{
		StoreType:    s.Store.FSName(),
		StoreAccount: s.Store.AccountName(),
		RepoID:       repoId,
		RepoVersion:  s.Infofile.Version,
		GPGUnlock:    gpgUnlock,
	})
}
