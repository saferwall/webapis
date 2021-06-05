package main

import (
	"os"
	"path/filepath"

	"github.com/saferwall/saferwall-api/internal/config"
	"github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

func main() {

	// Create root logger tagged with server version
	logger := log.New().With(nil, "version", Version)

	// Load application configurations
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	cfg, err := config.Load(dir+"./../config")
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	// Connect to the database.
	db.Open(cfg.DB.Server, cfg.DB.Username, cfg.DB.Password, cfg.DB.BucketName)

}
