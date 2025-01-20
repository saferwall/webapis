// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/saferwall/saferwall-api/internal/archive"
	"github.com/saferwall/saferwall-api/internal/config"
	"github.com/saferwall/saferwall-api/internal/db"
	smtpmailer "github.com/saferwall/saferwall-api/internal/mailer/smtp"
	"github.com/saferwall/saferwall-api/internal/queue"
	"github.com/saferwall/saferwall-api/internal/secure/password"
	"github.com/saferwall/saferwall-api/internal/secure/token"
	"github.com/saferwall/saferwall-api/internal/server"
	"github.com/saferwall/saferwall-api/internal/storage"
	tpl "github.com/saferwall/saferwall-api/internal/template"
	"github.com/saferwall/saferwall-api/pkg/log"
	"github.com/yeka/zip"
	"github.com/MicahParks/recaptcha"
)

// Version indicates the current version of the application.
var Version = "0.8.0"

var flagConfig = flag.String("config", "./../configs/", "path to the config file")
var flagN1QLFiles = flag.String("db", "./../db/", "path to the n1ql files")
var flagTplFiles = flag.String("tpl", "./../templates/", "path to html templates")

// @title Saferwall Web API
// @version 1.0
// @description Interact with Saferwall Malware Analysis Platform
// @termsOfService https://about.saferwall.com/tos

// @contact.name API Support
// @contact.url https://about.saferwall.com/contact.html
// @contact.email support@saferwall.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host api.saferwall.com
// @BasePath /v1

// @securityDefinitions.oauth2.password Bearer
// @tokenurl auth/login
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345".

// @schemes https
func main() {

	flag.Parse()

	// Create root logger tagged with server version.
	logger := log.New().With(context.TODO(), "version", Version)

	if err := run(logger); err != nil {
		logger.Errorf("failed to run the server: %s", err)
		os.Exit(-1)
	}
}

// run was explicitly created to allow main() to receive an error when server
// creation fails.
func run(logger log.Logger) error {

	// Load application configuration.
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		return err
	}
	logger.Info("successfully loaded config")

	// Connect to the database.
	dbx, err := db.Open(cfg.DB.Server, cfg.DB.Username,
		cfg.DB.Password, cfg.DB.BucketName)
	if err != nil {
		return err
	}
	logger.Info("connection to database has been established")

	// N1QL queries are stored separately from go code as the statement are
	// relatively complex and large.
	err = dbx.PrepareQueries(*flagN1QLFiles, cfg.DB.BucketName)
	if err != nil {
		return err
	}

	// Create a translator for validation error messages.
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()
	err = en_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return err
	}

	// Create a password securer for auth.
	sec := password.New(sha256.New())

	// Create a token generator service.
	tokenGen := token.New(dbx, sha256.New(), cfg.ResetPasswordTokenExp)

	// Create an uploader to upload file to object storage.
	updown, err := storage.New(cfg.ObjStorage)
	if err != nil {
		return err
	}

	// Create a producer to write messages to stream processing framework.
	producer, err := queue.New(cfg.Broker.Address, cfg.Broker.Topic)
	if err != nil {
		return err
	}

	// Create an archiver to zip files with a password in file download.
	archiver := archive.New(zip.AES256Encryption)

	// Create email client.
	var smtpMailer smtpmailer.SMTPMailer
	var emailTemplates tpl.Service
	if cfg.SMTP.Server != "" {
		smtpMailer = smtpmailer.New(cfg.SMTP.Server, cfg.SMTP.Port, cfg.SMTP.User,
			cfg.SMTP.Password)
		emailTemplates, err = tpl.New(*flagTplFiles)
		if err != nil {
			return err
		}
	}

	recaptchaVerifier := recaptcha.NewVerifierV3(cfg.RecaptchaKey, recaptcha.VerifierV3Options{})

	hs := &http.Server{
		Addr: cfg.Address,
		Handler: server.BuildHandler(logger, dbx, sec, cfg, Version, trans,
			updown, producer, smtpMailer, archiver, tokenGen, emailTemplates, recaptchaVerifier),
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
