// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/sha1"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/saferwall/saferwall-api/internal/config"
	"github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/event"
	"github.com/saferwall/saferwall-api/internal/secure"
	"github.com/saferwall/saferwall-api/internal/server"
	"github.com/saferwall/saferwall-api/internal/storage"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

func main() {

	// Create root logger tagged with server version.
	logger := log.New().With(nil, "version", Version)

	if err := run(logger); err != nil {
		logger.Errorf("failed to run the server: %s", err)
		os.Exit(-1)
	}
}

// run was explicitely created to allow main() to receive an error when server
// creation fails.
func run(logger log.Logger) error {

	// Load application configuration.
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	cfg, err := config.Load(dir + "./../config")
	if err != nil {
		return err
	}

	// Connect to the database.
	dbx, err := db.Open(cfg.DB.Server, cfg.DB.Username,
		cfg.DB.Password, cfg.DB.BucketName)
	if err != nil {
		return err
	}

	// Create a translator for validation error messages.
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)

	// Create a securer for auth.
	sec := secure.New(sha1.New())

	// Create an uploader to uplaod file to object storage.
	uploader, err := storage.New(cfg.DeploymentKind, cfg.ObjStorage)
	if err != nil {
		return err
	}

	// Create a producer to write messages to stream processing framework.
	producer, err := event.New(cfg.Broker.Network,
		 cfg.Broker.Topic, cfg.Broker.Address)
	if err != nil {
		return err
	}

	// Build HTTP server.
	hs := &http.Server{
		Addr:    cfg.Address,
		Handler: server.BuildHandler(logger, dbx, sec, cfg, Version, trans,
			 uploader, producer),
	}

	// Start server.
	go func() {
		logger.Infof("server is running at %s", cfg.Address)
		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err)
			os.Exit(-1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a
	// timeout of 10 seconds. Use a buffered channel to avoid missing
	// signals as recommended for signal.Notify.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := hs.Shutdown(ctx); err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	return nil
}
