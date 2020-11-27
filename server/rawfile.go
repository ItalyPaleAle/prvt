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
	"time"

	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-gonic/gin"
)

// RawFileGetHandler is the handler for GET /rawfile/:path, which returns a raw file from the fs
func (s *Server) RawFileGetHandler(c *gin.Context) {
	// Get the path and ensure the path does not start with /
	path := strings.TrimPrefix(c.Param("path"), "/")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("empty path"))
		return
	}

	// Check if we have a range
	rngHeader, err := utils.ParseRange(c.GetHeader("Range"))
	if err != nil {
		c.AbortWithError(http.StatusRequestedRangeNotSatisfiable, err)
		return
	}
	var start, count int64
	if rngHeader != nil {
		start = rngHeader.Start
		count = rngHeader.Length
	}

	// Context with a cancel function
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Detect closed connections (this should be redundant, but still)
	notifier := c.Writer.CloseNotify()

	// Add a timeout to detect idle connections (which are not closed but are hanging)
	t := time.NewTimer(IdleTimeout * time.Second)
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
				t.Reset(IdleTimeout * time.Second)
			}
		}
	}()
	out := utils.CtxWriter(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			read <- 1
			return c.Writer.Write(p)
		}
	})

	// Stream the file to the response writer
	found, _, err := s.Store.RawGet(ctx, path, out, start, count)
	if err != nil {
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
