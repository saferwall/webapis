// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.


package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saferwall/saferwall-api/internal/auth"
	"github.com/saferwall/saferwall-api/internal/config"
	"github.com/saferwall/saferwall-api/internal/db"
	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/healthcheck"
	"github.com/saferwall/saferwall-api/internal/user"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

func main() {

	// Create root logger tagged with server version
	logger := log.New().With(nil, "version", Version)

	// Load application configurations
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	cfg, err := config.Load(dir + "./../config")
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	// Connect to the database.
	dbx, err := db.Open(cfg.DB.Server, cfg.DB.Username,
		cfg.DB.Password, cfg.DB.BucketName)
	if err != nil {
		logger.Errorf("failed to connect to database: %s", err)
		os.Exit(-1)
	}

	// Build HTTP server.
	hs := &http.Server{
		Addr:    cfg.Address,
		Handler: buildHandler(logger, dbx, cfg),
	}

	// Start server.
	go func() {
		logger.Infof("server is running at %s", cfg.Address)
		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err)
			os.Exit(-1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := hs.Shutdown(ctx); err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

}

// buildHandler sets up the HTTP routing and builds an HTTP handler.
func buildHandler(logger log.Logger, db *dbcontext.DB,
	cfg *config.Config) http.Handler {

	// Create `echo` instance
	e := echo.New()

	// Setup middlware.
	e.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Format: `{"remote_ip":"${remote_ip}","host":"${host}",` +
				`"method":"${method}","uri":"${uri}","status":${status},` +
				`"latency":${latency},"latency_human":"${latency_human}",` +
				`"bytes_in":${bytes_in},bytes_out":${bytes_out}}` + "\n",
		}))

	// CORS middlware.
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{cfg.UI.Address},
		AllowMethods: []string{
			echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowCredentials: true,
	}))

	// Add Trailing slash for consistent URIs.
	e.Pre(middleware.AddTrailingSlash())

	// Healthcheck endpoint.
	healthcheck.RegisterHandlers(e, Version)

	// Creates a new group for v1.
	g := e.Group("/v1")

	// Setup JWT Auth handler.
	authHandler := auth.Handler(cfg.JWTSigningKey)

	user.RegisterHandlers(g, user.NewService(
		user.NewRepository(db, logger), logger),
		authHandler, logger,
	)

	return e
}
