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

	"github.com/ItalyPaleAle/prvt/fs/fsindex"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/keys"
	"github.com/ItalyPaleAle/prvt/repository"

	"github.com/gin-gonic/gin"
)

// PostRepoUnlockHandler is the handler for POST /api/repo/unlock, which unlocks a repository
// The same handler is also used by POST /api/repo/keytest, which tests a key and returns its ID (essentially performing a "dry run" for the unlock operation)
func (s *Server) PostRepoUnlockHandler(dryRun bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Get the information to unlock the repository from the body
		args := &UnlockKeyRequest{}
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{msg})
			return
		}

		// Skip if doing a dry-run (e.g. while testing a key)
		if !dryRun {
			// Set the master key
			s.Store.SetMasterKey(keyId, masterKey)

			// Set up the repository and index
			s.Repo = &repository.Repository{
				Store: s.Store,
				Index: &index.Index{},
			}
			indexProvider := &fsindex.IndexProviderFs{
				Store: s.Store,
			}
			s.Repo.Index.SetProvider(indexProvider)

			fmt.Fprintln(s.LogWriter, "Repository unlocked with key:", keyId)
		}

		c.JSON(http.StatusOK, RepoKeyListItem{
			KeyId: keyId,
			Type:  args.Type,
		})
	}
}
