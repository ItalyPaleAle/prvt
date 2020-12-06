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
	"path/filepath"
	"syscall"
	"time"

	"github.com/ItalyPaleAle/prvt/buildinfo"
	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/repository"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Server is the main class for the prvt server
type Server struct {
	Store     fs.Fs
	Verbose   bool
	Repo      *repository.Repository
	Infofile  *infofile.InfoFile
	LogWriter io.Writer
	ReadOnly  bool
}

// Start the server
func (s *Server) Start(ctx context.Context, address, port string) error {
	// Log writer
	if s.LogWriter == nil {
		s.LogWriter = os.Stdout
	}

	// Load config
	err := s.loadConfig()
	if err != nil {
		return err
	}

	// Set gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Create a router
	router := gin.New()

	// Enable CORS when in development
	if !utils.IsTruthy(buildinfo.Production) {
		corsConfig := cors.DefaultConfig()
		corsConfig.AddAllowHeaders("Range")
		corsConfig.AddExposeHeaders("Date")
		corsConfig.AllowAllOrigins = true
		router.Use(cors.New(corsConfig))
	}

	// Add middlewares: logger (if desired) and recovery
	if s.Verbose {
		router.Use(gin.LoggerWithWriter(s.LogWriter))
	}
	router.Use(gin.Recovery())

	// Register all API routes
	s.registerAPIRoutes(router)

	// Web UI
	err = s.handleWebUI(router)
	if err != nil {
		return err
	}

	// Start the server
	err = s.launchServer(ctx, address, port, router)
	if err != nil {
		return err
	}

	return nil
}

// loadConfig loads the config file
func (s *Server) loadConfig() error {
	// Look for config in both $HOME/.config/.prvt (and equivalent on other OS's) and in the cwd
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("cannot determine config folder location: %s", err.Error())
	}
	configDir = filepath.Join(configDir, "prvt")

	// Set the config file name for viper
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(configDir)
	if !utils.IsTruthy(buildinfo.Production) {
		viper.AddConfigPath(".")
	}

	// Create the directory if it doesn't exist
	err = utils.EnsureFolder(configDir)
	if err != nil {
		return err
	}

	// Read config
	err = viper.ReadInConfig()
	if err != nil {
		// If the file doesn't exist, create it
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = utils.TouchFile(filepath.Join(configDir, "config.yaml"))
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// registerAPIRoutes registers the routes for the APIs
func (s *Server) registerAPIRoutes(router *gin.Engine) {
	// Middlewares for all APIs
	apis := router.Group("", s.MiddlewareRepoStatus)

	// Add routes
	apis.GET("/api/info", s.GetInfoHandler)
	apis.POST("/api/repo/select", s.PostRepoSelectHandler)
	apis.POST("/api/repo/close", s.PostRepoCloseHandler)
	apis.GET("/api/connection", s.GetConnectionHandler)
	apis.POST("/api/connection", s.PostConnectionHandler)
	apis.GET("/api/connection/:name", s.GetConnectionInfoHandler)
	apis.DELETE("/api/connection/:name", s.DeleteConnectionHandler)
	apis.GET("/api/fsoptions", s.GetFsOptionsHandler)
	apis.GET("/api/fsoptions/:fs", s.GetFsOptionsHandler)

	// Routes that require an open repository
	requireRepo := apis.Group("", s.MiddlewareRequireRepo)

	// Routes that require the repository to be unlocked
	{
		requireUnlock := requireRepo.Group("", s.MiddlewareRequireUnlock)

		// Request a file
		requireUnlock.GET("/file/:fileId", s.FileHandler)
		requireUnlock.HEAD("/file/:fileId", s.FileHandler)

		// APIs
		group := requireUnlock.Group("/api")
		group.GET("/tree", s.GetTreeHandler)
		group.GET("/tree/*path", s.GetTreeHandler)
		group.POST("/tree/*path",
			s.MiddlewareRequireReadWrite,
			s.MiddlewareRequireInfoFileVersion(3),
			s.PostTreeHandler,
		)
		group.DELETE("/tree/*path",
			s.MiddlewareRequireReadWrite,
			s.MiddlewareRequireInfoFileVersion(3),
			s.DeleteTreeHandler,
		)
		group.GET("/metadata/*file", s.GetMetadataHandler)
		group.POST("/repo/key",
			s.MiddlewareRequireReadWrite,
			s.MiddlewareRequireInfoFileVersion(5),
			s.PostRepoKeyHandler,
		)
		group.DELETE("/repo/key/:keyId",
			s.MiddlewareRequireReadWrite,
			s.MiddlewareRequireInfoFileVersion(2),
			s.DeleteRepoKeyHandler,
		)
	}

	// Other routes that don't require the repository to be unlocked
	{
		// Request a raw file
		requireRepo.GET("/rawfile/:path", s.RawFileGetHandler)

		// APIs
		group := requireRepo.Group("/api")
		group.GET("/repo/infofile", s.GetRepoInfofileHandler)
		group.GET("/repo/key", s.MiddlewareRequireInfoFileVersion(2), s.GetRepoKeyHandler)

		// These APIs accept requests to unlock the repo
		group.POST("/repo/unlock", s.PostRepoUnlockHandler(false))
		group.POST("/repo/keytest", s.PostRepoUnlockHandler(true))
	}
}

// launchServer launches the HTTP server
func (s *Server) launchServer(ctx context.Context, address, port string, router *gin.Engine) error {
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
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	// Release all locks, if any
	_ = s.releaseRepoLock()

	return nil
}

func (s *Server) releaseRepoLock() (err error) {
	// If there's an existing store object, release locks (if any)
	if s.Store != nil {
		err = s.Store.ReleaseLock(context.Background())
	}
	return
}
