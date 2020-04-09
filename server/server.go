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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ItalyPaleAle/prvt/fs"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
)

type Server struct {
	Store   fs.Fs
	Verbose bool
}

func (s *Server) Start(address, port string) error {
	// Set gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Start gin server
	router := gin.New()

	// Add middlewares: logger (if desired) and recovery
	if s.Verbose {
		router.Use(gin.Logger())
	}
	router.Use(gin.Recovery())

	// Add routes
	router.GET("/file/:fileId", s.FileHandler)
	{
		// APIs
		apis := router.Group("/api")
		apis.GET("/tree/*path", s.TreeHandler)
	}

	// UI
	uiBox := packr.New("ui", "../ui/dist")
	router.GET("/ui/*page", gin.WrapH(http.StripPrefix("/ui/", http.FileServer(uiBox))))

	// Redirect from / to the UI
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui")
	})

	// HTTP Server
	server := &http.Server{
		Addr:           address + ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Handle graceful shutdown on SIGINT
	idleConnsClosed := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, os.Interrupt, syscall.SIGTERM)
		<-s

		// We received an interrupt signal, shut down.
		if err := server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			fmt.Printf("HTTP server shutdown error: %v\n", err)
		}
		close(idleConnsClosed)
	}()

	// Listen to connections
	fmt.Printf("Listening on http://%s:%s\n", address, port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	<-idleConnsClosed

	return nil
}
