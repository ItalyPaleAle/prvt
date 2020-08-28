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

	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/keys"
	"github.com/ItalyPaleAle/prvt/repository"

	"github.com/gin-gonic/gin"
)

// MiddlewareUnlockRepo tries to unlock the repository
func (s *Server) MiddlewareUnlockRepo(dryRun bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		// If this repository is already unlocked and we're not doing a dry-run, abort
		if !dryRun && s.Repo != nil {
			c.AbortWithStatusJSON(http.StatusConflict, errorResponse{"This repository has already been unlocked"})
		}

		// Get the information to unlock the repository from the body
		args := &unlockKeyRequest{}
		if ok := args.FromBody(c); !ok {
			return
		}

		// Try unlocking the repository
		var (
			masterKey  []byte
			keyId      string
			errMessage string
			err        error
		)
		if args.Type == "gpg" {
			masterKey, keyId, errMessage, err = keys.GetMasterKeyWithGPG(s.Infofile)
		} else {
			masterKey, keyId, errMessage, err = keys.GetMasterKeyWithPassphrase(s.Infofile, args.Passphrase)
		}
		if err != nil {
			msg := fmt.Sprintf("%s: %s", err, errMessage)
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{msg})
			return
		}

		// Skip if doing a dry-run (e.g. while testing a key)
		if !dryRun {
			// Set the master key
			s.Store.SetMasterKey(keyId, masterKey)

			// Set up the index
			index.Instance.SetStore(s.Store)

			// Set up the repository
			s.Repo = &repository.Repository{
				Store: s.Store,
			}

			fmt.Println("Repository unlocked with key:", keyId)
		}

		// Store the key ID in the context
		c.Set("KeyId", keyId)
		c.Set("KeyType", args.Type)
	}
}

// MiddlewareRequireRepo requires a repository to be selected (even if not unlocked)
func (s *Server) MiddlewareRequireRepo(c *gin.Context) {
	if s.Store == nil || s.Infofile == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{"No repository has been selected yet"})
	}
}

// MiddlewareRequireUnlock requires the repository to be unlocked before processing
func (s *Server) MiddlewareRequireUnlock(c *gin.Context) {
	if s.Repo == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{"The repository has not been unlocked"})
	}
}

// MiddlewareRequireInfoFileVersion requires a minimum version of the info file to continue
func (s *Server) MiddlewareRequireInfoFileVersion(version uint16) func(c *gin.Context) {
	return func(c *gin.Context) {
		if s.Infofile.Version < version {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, errorResponse{`This repository needs to be upgraded. Please run "prvt repo upgrade --store <string>" to upgrade this repository to the latest format`})
		}
	}
}
