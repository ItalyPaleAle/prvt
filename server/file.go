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
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

// FileHandler is the handler for GET /file/:fileId, which returns a (decrypted) file
func (s *Server) FileHandler(c *gin.Context) {
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

	// Check if we have the dl=1 option, which forces a download
	forceDownload := false
	dlQs := c.Query("dl")
	if dlQs == "1" || dlQs == "true" || dlQs == "t" || dlQs == "y" || dlQs == "yes" {
		forceDownload = true
	}

	// Check if we have a range
	rngHeader, err := utils.ParseRange(c.GetHeader("Range"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	var rng *fs.RequestRange
	if rngHeader != nil {
		rng = fs.NewRequestRange(rngHeader, 0, 0)
	}

	// Load and decrypt the file, then pipe it to the response writer (but not for HEAD requests)
	var out io.Writer
	if c.Request.Method != "HEAD" {
		out = c.Writer
	}
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	metadataCb := func(metadata *crypto.Metadata, metadataSize int32) {
		// Send headers before the data is sent
		// Start with Content-Type and Content-Disposition
		if metadata.ContentType != "" {
			c.Header("Content-Type", metadata.ContentType)
		} else {
			c.Header("Content-Type", "application/octet-stream")
		}
		contentDisposition := "inline"
		if forceDownload {
			contentDisposition = "attachment"
		}
		if metadata.Name != "" {
			fileName := strings.ReplaceAll(metadata.Name, "\"", "")
			contentDisposition += "; filename=\"" + fileName + "\""
		}
		c.Header("Content-Disposition", contentDisposition)

		// Handle range requests
		if rng != nil {
			// Spec: https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests
			// Content-Length is the length of the range itself
			c.Header("Content-Length", strconv.FormatInt(rng.Length, 10))
			c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rng.Start, rng.Start+rng.Length-1, metadata.Size))
			c.Status(http.StatusPartialContent)
		} else {
			// Content-Length and Accept-Ranges
			if metadata.Size > 0 {
				c.Header("Content-Length", strconv.FormatInt(metadata.Size, 10))
				c.Header("Accept-Ranges", "bytes")
			}
		}

		// If this is a HEAD request, stop requesting the body
		if c.Request.Method == "HEAD" {
			cancel()
		}
	}
	var found bool
	if rng != nil {
		found, _, err = s.Store.GetWithRange(ctx, fileId, out, rng, metadataCb)
	} else {
		found, _, err = s.Store.GetWithContext(ctx, fileId, out, metadataCb)
	}
	if err != nil {
		if c.Request.Method == "HEAD" && err == crypto.ErrMetadataOnly {
			c.AbortWithStatus(200)
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
