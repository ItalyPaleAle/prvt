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
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/ItalyPaleAle/prvt/buildinfo"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-gonic/gin"
	"github.com/markbates/pkger"
)

// Handles requests for the web ui
func (s *Server) handleWebUI(router *gin.Engine) error {
	// First, set up the handler for the assets
	assetsBox, err := pkger.Open("/ui/assets")
	if err != nil {
		return err
	}

	wasmMatch := regexp.MustCompile("^app(-([a-z0-9\\.-_]+))?.wasm$")
	acceptEncMatch := regexp.MustCompile("\\bbr\\b")

	// Serve the files from the assets directory
	router.GET("/assets/*path", func(c *gin.Context) {
		// Get the path and remove the optional / prefix
		path := strings.TrimPrefix(c.Param("path"), "/")

		// This method (for now at least) can only support the app.wasm files, with an optional build ID
		if !wasmMatch.MatchString(path) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// Try to return the file (uncompressed) if it exists in the box
		f, err := assetsBox.Open("/app.wasm")
		if err != nil && !os.IsNotExist(err) {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if err == nil {
			c.Header("Content-Type", "application/wasm")
			_, err = io.Copy(c.Writer, f)
			f.Close()
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			return
		}

		// If we're here, the .wasm file doesn't exist, so try the compressed one
		// First, ensure that the client accepts it
		acceptEnc := c.Request.Header.Get("Accept-Encoding")
		if acceptEnc == "" || !acceptEncMatch.MatchString(acceptEnc) {
			c.String(http.StatusBadRequest, "Client does not accept responses compressed with brotli\n")
			c.Abort()
			return
		}

		// Return the compressed file
		f, err = assetsBox.Open("/app.wasm.br")
		if err != nil && !os.IsNotExist(err) {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if err == nil {
			c.Header("Content-Type", "application/wasm")
			c.Header("Content-Encoding", "br")
			_, err = io.Copy(c.Writer, f)
			f.Close()
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			return
		}

		// We're still here: file doesn't exist
		c.AbortWithStatus(http.StatusNotFound)
		return
	})

	// Serve the webapp
	// This is the fallback if no other route has been specified
	if utils.IsTruthy(buildinfo.Production) {
		uiBox, err := pkger.Open("/ui/dist")
		if err != nil {
			return err
		}
		router.NoRoute(gin.WrapH(http.FileServer(uiBox)))
	} else {
		// In development, proxy to the webpack dev server
		target, _ := url.Parse("http://localhost:3000")
		router.NoRoute(gin.WrapH(httputil.NewSingleHostReverseProxy(target)))
	}

	return nil
}
