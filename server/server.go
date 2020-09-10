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
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/repository"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
)

type Server struct {
	Store     fs.Fs
	Verbose   bool
	Repo      *repository.Repository
	Infofile  *infofile.InfoFile
	LogWriter io.Writer
	ReadOnly  bool
}

func (s *Server) Start(ctx context.Context, address, port string) error {
	// Log writer
	if s.LogWriter == nil {
		s.LogWriter = os.Stdout
	}

	// Set gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Start gin server
	router := gin.New()

	// Add middlewares: logger (if desired) and recovery
	if s.Verbose {
		router.Use(gin.LoggerWithWriter(s.LogWriter))
	}
	router.Use(gin.Recovery())

	// Add routes
	router.GET("/api/info", s.GetInfoHandler)
	router.POST("/api/repo/select", s.PostRepoSelectHandler)

	// Routes that require an open repository
	{
		requireRepo := router.Group("", s.MiddlewareRequireRepo)

		// Routes that require the repository to be unlocked
		{
			requireUnlock := requireRepo.Group("", s.MiddlewareRequireUnlock)

			// Request a file
			requireUnlock.GET("/file/:fileId", s.FileHandler)
			requireUnlock.HEAD("/file/:fileId", s.FileHandler)

			// APIs
			apis := requireUnlock.Group("/api")
			apis.GET("/tree", s.GetTreeHandler)
			apis.GET("/tree/*path", s.GetTreeHandler)
			apis.POST("/tree/*path",
				s.MiddlewareRequireReadWrite,
				s.MiddlewareRequireInfoFileVersion(3),
				s.PostTreeHandler,
			)
			apis.DELETE("/tree/*path",
				s.MiddlewareRequireReadWrite,
				s.MiddlewareRequireInfoFileVersion(3),
				s.DeleteTreeHandler,
			)
			apis.GET("/metadata/*file", s.GetMetadataHandler)
			apis.POST("/repo/key",
				s.MiddlewareRequireReadWrite,
				s.MiddlewareRequireInfoFileVersion(2),
				s.PostRepoKeyHandler,
			)
			apis.DELETE("/repo/key/:keyId",
				s.MiddlewareRequireReadWrite,
				s.MiddlewareRequireInfoFileVersion(2),
				s.DeleteRepoKeyHandler,
			)
		}

		// Other APIs that don't require the repository to be unlocked
		{
			apis := requireRepo.Group("/api")
			apis.GET("/repo/key", s.MiddlewareRequireInfoFileVersion(2), s.GetRepoKeyHandler)

			// These APIs accept requests to unlock the repo
			apis.POST("/repo/unlock", s.PostRepoUnlockHandler(false))
			apis.POST("/repo/keytest", s.PostRepoUnlockHandler(true))
		}
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
		Addr:    address + ":" + port,
		Handler: router,
		//ReadTimeout:    10 * time.Second,
		//WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Handle graceful shutdown on SIGINT
	idleConnsClosed := make(chan struct{})
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		select {
		case <-ctx.Done():
		case <-ch:
		}

		// We received an interrupt signal, shut down
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := server.Shutdown(shutdownCtx); err != nil {
			// Error from closing listeners, or context timeout:
			fmt.Fprintf(s.LogWriter, "HTTP server shutdown error: %v\n", err)
		}
		shutdownCancel()
		close(idleConnsClosed)
	}()

	// Listen to connections
	fmt.Fprintf(s.LogWriter, "Listening on http://%s:%s\n", address, port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	<-idleConnsClosed

	return nil
}
