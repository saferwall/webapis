package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saferwall/saferwall-api/internal/config"
	"github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/healthcheck"
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

}

// buildHandler sets up the HTTP routing and builds an HTTP handler.
func buildHandler(logger log.Logger, db *dbcontext.DB,
	cfg *config.Config) http.Handler {

	// Create `echo` instance
	e := echo.New()

	// Setup middlware.
	e.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Format: `{"remote_ip":"${remote_ip}","host":"${host}",
			method":"${method}","uri":"${uri}","status":${status},
			"latency":${latency},"latency_human":"${latency_human}",
			"bytes_in":${bytes_in},bytes_out":${bytes_out}}` + "\n",
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

	healthcheck.RegisterHandlers(e, Version)

	// Creates a new group for v1.
	g := e.Group("/v1")


	authHandler := auth.Handler(cfg.JWTSigningKey)

	album.RegisterHandlers(rg.Group(""),
		album.NewService(album.NewRepository(db, logger), logger),
		authHandler, logger,
	)

	auth.RegisterHandlers(rg.Group(""),
		auth.NewService(cfg.JWTSigningKey, cfg.JWTExpiration, logger),
		logger,
	)

	return router
}
