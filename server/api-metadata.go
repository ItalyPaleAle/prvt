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
	"context"
	"errors"
	"net/http"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/index"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

// GetMetadataHandler is the handler for GET /api/metadata/:fileId, which returns the metadata for a file
func (s *Server) GetMetadataHandler(c *gin.Context) {
	// Get the fileId
	fileId := c.Param("fileId")
	if fileId == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("empty fileId"))
		return
	}

	// Ensure fileId is a UUID
	fileIdUUID, err := uuid.FromString(fileId)
	if err != nil || fileIdUUID.Version() != 4 {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Get the element from the index
	el, err := index.Instance.GetFileById(fileId)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if el == nil {
		c.AbortWithError(http.StatusNotFound, errors.New("file not found in index"))
		return
	}

	// Request the metadata
	found, _, err := s.Store.GetWithContext(c.Request.Context(), fileId, nil, func(metadata *crypto.Metadata, metadataSize int32) {
		response := metadataResponse{
			FileId:   fileId,
			Path:     el.Path,
			Name:     metadata.Name,
			Date:     el.Date,
			MimeType: metadata.ContentType,
			Size:     metadata.Size,
		}
		c.JSON(http.StatusOK, response)
	})
	if err != nil && err != crypto.ErrMetadataOnly {
		// Ignore canceled contexts, e.g. if the browser canceled the request
		if err == context.Canceled {
			c.Abort()
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !found {
		c.AbortWithError(http.StatusNotFound, errors.New("file does not exist"))
		return
	}
}
