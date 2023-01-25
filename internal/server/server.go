// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"runtime/debug"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saferwall/saferwall-api/internal/activity"
	"github.com/saferwall/saferwall-api/internal/archive"
	"github.com/saferwall/saferwall-api/internal/auth"
	"github.com/saferwall/saferwall-api/internal/comment"
	"github.com/saferwall/saferwall-api/internal/config"
	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/internal/file"
	"github.com/saferwall/saferwall-api/internal/healthcheck"
	"github.com/saferwall/saferwall-api/internal/mailer"
	"github.com/saferwall/saferwall-api/internal/queue"
	"github.com/saferwall/saferwall-api/internal/secure/password"
	"github.com/saferwall/saferwall-api/internal/secure/token"
	"github.com/saferwall/saferwall-api/internal/storage"
	tpl "github.com/saferwall/saferwall-api/internal/template"
	"github.com/saferwall/saferwall-api/internal/user"
	"github.com/saferwall/saferwall-api/pkg/log"
)

const (
	// Returned when request body length is null.
	errEmptyBody = "You have sent an empty body."
)

// BuildHandler sets up the HTTP routing and builds an HTTP handler.
func BuildHandler(logger log.Logger, db *dbcontext.DB, sec password.Service,
	cfg *config.Config, version string, trans ut.Translator,
	updown storage.UploadDownloader, p queue.Producer,
	smtpMailer mailer.SMTPMailer, arch archive.Archiver,
	tokenGen token.Service,
	emailTpl tpl.Service) http.Handler {

	// Create `echo` instance.
	e := echo.New()

	// Logging middleware.
	e.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Format: `{"remote_ip":"${remote_ip}","host":"${host}",` +
				`"method":"${method}","uri":"${uri}","status":${status},` +
				`"latency":${latency},"latency_human":"${latency_human}",` +
				`"bytes_in":${bytes_in},bytes_out":${bytes_out}}` + "\n",
		}))

	// CORS middleware.
	CORSAllowOrigins := cfg.CORSOrigins
	if cfg.DisableCORS {
		CORSAllowOrigins = []string{"*"}
	} else {
		CORSAllowOrigins = append(CORSAllowOrigins, cfg.UI.Address)
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: CORSAllowOrigins,
		AllowMethods: []string{
			echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowCredentials: true,
	}))

	// Recover from panic middleware.
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisablePrintStack: true,
	}))

	// Rate limiter middleware.
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// Add trailing slash for consistent URIs.
	e.Pre(middleware.AddTrailingSlash())

	// Setup JWT Auth handler.
	authHandler := auth.Handler(cfg.JWTSigningKey)
	optAuthHandler := auth.IsAuthenticated(authHandler)

	// Register a custom fields validator.
	e.Validator = &CustomValidator{validator: validator.New()}

	// Register a custom binder.
	e.Binder = &CustomBinder{b: &echo.DefaultBinder{}}

	// Setup a custom HTTP error handler.
	e.HTTPErrorHandler = CustomHTTPErrorHandler(trans)

	// Creates a new group for v1.
	g := e.Group("/v1")

	// Create the services and register the handlers.
	actSvc := activity.NewService(activity.NewRepository(db, logger), logger)
	userSvc := user.NewService(user.NewRepository(db, logger), logger, tokenGen,
		sec, cfg.ObjStorage.AvatarsContainerName, updown, actSvc)
	authSvc := auth.NewService(cfg.JWTSigningKey, cfg.JWTExpiration, logger,
		sec, userSvc, tokenGen)
	fileSvc := file.NewService(file.NewRepository(db, logger), logger, updown,
		p, cfg.Broker.Topic, cfg.ObjStorage.FileContainerName, userSvc, actSvc,
		arch)
	commentSvc := comment.NewService(comment.NewRepository(db, logger), logger,
		actSvc, userSvc, fileSvc)

	// Create the file, user and comment middleware.
	fileMiddleware := file.NewMiddleware(fileSvc, logger)
	userMiddleware := user.NewMiddleware(userSvc, logger)
	commentMiddleware := comment.NewMiddleware(commentSvc, logger)

	// Register the handlers.
	healthcheck.RegisterHandlers(e, version)
	user.RegisterHandlers(g, userSvc, authHandler, optAuthHandler, userMiddleware.VerifyUser,
		logger, smtpMailer, emailTpl)
	auth.RegisterHandlers(g, authSvc, logger, smtpMailer, emailTpl, cfg.UI.Address)
	file.RegisterHandlers(g, fileSvc, logger, authHandler, optAuthHandler, fileMiddleware.VerifyHash)
	activity.RegisterHandlers(g, actSvc, authHandler, logger)
	comment.RegisterHandlers(g, commentSvc, logger, authHandler, commentMiddleware.VerifyID)

	return e
}

// CustomValidator holds custom validator.
type CustomValidator struct {
	validator *validator.Validate
}

// Validate performs field validation.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

// NewBinder initializes custom server binder.
func NewBinder() *CustomBinder {
	return &CustomBinder{b: &echo.DefaultBinder{}}
}

// CustomBinder struct.
type CustomBinder struct {
	b echo.Binder
}

// Bind tries to bind request into interface, and if it does then validate it.
func (cb *CustomBinder) Bind(i interface{}, c echo.Context) error {
	if c.Request().ContentLength == 0 {
		return errors.BadRequest(errEmptyBody)
	}
	if err := cb.b.Bind(i, c); err != nil && err != echo.ErrUnsupportedMediaType {
		return err
	}
	return c.Validate(i)
}

// CustomHTTPErrorHandler handles errors encountered during HTTP request
// processing.
func CustomHTTPErrorHandler(trans ut.Translator) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		l := c.Logger()
		res := errors.BuildErrorResponse(err, trans)
		if res.StatusCode() == http.StatusInternalServerError {
			debug.PrintStack()
			l.Errorf("encountered internal server error: %v", err)
		}
		if err = c.JSON(res.StatusCode(), res); err != nil {
			l.Errorf("failed writing error response: %v", err)
		}
	}
}
