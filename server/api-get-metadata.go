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
	"errors"
	"net/http"
	"strings"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/index"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

// GetMetadataHandler is the handler for GET /api/metadata/:file, which returns the metadata for a file
func (s *Server) GetMetadataHandler(c *gin.Context) {
	// Get the file parameter and remove the leading /
	file := c.Param("file")
	if file == "" || file == "/" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"empty file name"})
		return
	}
	if strings.HasPrefix(file, "/") {
		file = file[1:]
	}

	// Get the element from the index
	var el *index.FolderList
	fileIdUUID, err := uuid.FromString(file)
	// Check if we have a file ID
	if err == nil && fileIdUUID.Version() == 4 {
		el, err = s.Repo.Index.GetFileById(0, file)
	} else {
		// Re-add the leading /
		el, err = s.Repo.Index.GetFileByPath(0, "/"+file)
	}
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if el == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{"file not found in index"})
		return
	}

	// Request the metadata
	found, _, err := s.Store.Get(c.Request.Context(), el.FileId, nil, func(metadata *crypto.Metadata, metadataSize int32) {
		pos := strings.LastIndex(el.Path, "/") + 1
		response := MetadataResponse{
			FileId:   el.FileId,
			Folder:   el.Path[0:pos],
			Name:     metadata.Name,
			Date:     el.Date,
			MimeType: metadata.ContentType,
			Size:     metadata.Size,
			Digest:   el.Digest,
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
