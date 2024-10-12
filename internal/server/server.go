// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"regexp"
	"runtime/debug"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saferwall/saferwall-api/internal/activity"
	"github.com/saferwall/saferwall-api/internal/archive"
	"github.com/saferwall/saferwall-api/internal/auth"
	"github.com/saferwall/saferwall-api/internal/behavior"
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

	// Email regex string.
	emailRegexString = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"

	// Alphanum regex string.
	usernameRegexString = "^[a-zA-Z0-9]{1,20}$"
)

var (
	emailRegex    = regexp.MustCompile(emailRegexString)
	usernameRegex = regexp.MustCompile(usernameRegexString)
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
	validate := validator.New()
	_ = validate.RegisterValidation("username_or_email", validateUsernameOrEmail)
	e.Validator = &CustomValidator{validator: validate}

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
		p, cfg.Broker.Topic, cfg.ObjStorage.FileContainerName, cfg.SamplesZipPwd,
		userSvc, actSvc, arch)
	commentSvc := comment.NewService(comment.NewRepository(db, logger), logger,
		actSvc, userSvc, fileSvc)
	behaviorSvc := behavior.NewService(behavior.NewRepository(db, logger), logger)

	// Create the middlewares.
	fileMiddleware := file.NewMiddleware(fileSvc, logger)
	userMiddleware := user.NewMiddleware(userSvc, logger)
	commentMiddleware := comment.NewMiddleware(commentSvc, logger)
	behaviorMiddleware := behavior.NewMiddleware(behaviorSvc, logger)

	// Register the handlers.
	healthcheck.RegisterHandlers(e, version)
	user.RegisterHandlers(g, userSvc, cfg.MaxAvatarSize, authHandler, optAuthHandler, userMiddleware.VerifyUser,
		logger, smtpMailer, emailTpl)
	auth.RegisterHandlers(g, authSvc, logger, smtpMailer, emailTpl, cfg.UI.Address)
	file.RegisterHandlers(g, fileSvc, logger, cfg.MaxFileSize, authHandler, optAuthHandler, fileMiddleware.VerifyHash, fileMiddleware.ModifyResponse)
	activity.RegisterHandlers(g, actSvc, authHandler, logger)
	comment.RegisterHandlers(g, commentSvc, logger, authHandler, commentMiddleware.VerifyID)
	behavior.RegisterHandlers(g, behaviorSvc, authHandler, behaviorMiddleware.VerifyID, logger)

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

// validateUserOrEmail implements validator.Func.
func validateUsernameOrEmail(fl validator.FieldLevel) bool {
	usernameOrEmail := fl.Field().String()
	if !emailRegex.MatchString(usernameOrEmail) && !usernameRegex.MatchString(usernameOrEmail) {
		return false
	}
	return true
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
