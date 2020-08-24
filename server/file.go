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
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		c.AbortWithError(http.StatusRequestedRangeNotSatisfiable, err)
		return
	}
	var rng *fs.RequestRange
	if rngHeader != nil {
		rng = fs.NewRequestRange(rngHeader)
	}

	// Context
	ctx := c.Request.Context()

	// Load and decrypt the file, then pipe it to the response writer (but not for HEAD requests)
	var out io.Writer
	if c.Request.Method != "HEAD" {
		// Context with a cancel function
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()

		// Detect closed connections (this should be redundant, but still)
		notifier := c.Writer.CloseNotify()

		// Add a timeout to detect idle connections (which are not closed but are hanging)
		timeout := 20 * time.Second
		t := time.NewTimer(timeout)
		read := make(chan int)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-notifier:
					cancel()
				case <-t.C:
					cancel()
				case <-read:
					if !t.Stop() {
						<-t.C
					}
					t.Reset(timeout)
				}
			}
		}()

		out = utils.CtxWriter(func(p []byte) (int, error) {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
				read <- 1
				return c.Writer.Write(p)
			}
		})
	}
	metadataCb := func(metadata *crypto.Metadata, metadataSize int32) bool {
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
			if rng.Start >= rng.FileSize {
				c.Status(http.StatusRequestedRangeNotSatisfiable)
			} else {
				// Spec: https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests
				// Content-Length is the length of the range itself
				c.Header("Content-Length", strconv.FormatInt(rng.Length, 10))
				c.Header("Content-Range", rng.ResponseHeaderValue())
				c.Status(http.StatusPartialContent)
			}
		} else {
			// Content-Length and Accept-Ranges
			if metadata.Size > 0 {
				c.Header("Content-Length", strconv.FormatInt(metadata.Size, 10))
				c.Header("Accept-Ranges", "bytes")
			}
		}
		return true
	}
	var found bool
	if rng != nil {
		found, _, err = s.Store.GetWithRange(ctx, fileId, out, rng, metadataCb)
	} else {
		found, _, err = s.Store.GetWithContext(ctx, fileId, out, metadataCb)
	}
	if err != nil {
		// Ignore error ErrMetadataOnly if we're making a head request
		if c.Request.Method == "HEAD" && err == crypto.ErrMetadataOnly {
			c.AbortWithStatus(200)
			return
		}
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
