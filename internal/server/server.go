package server

import (
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saferwall/saferwall-api/internal/auth"
	"github.com/saferwall/saferwall-api/internal/config"
	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/internal/healthcheck"
	"github.com/saferwall/saferwall-api/internal/user"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Server represents our server, it include all dependencies and make it easy
// to understand what the server needs.
type Server struct {
	Echo   *echo.Echo     // HTTP middleware
	config *config.Config // Configuration
	db     *dbcontext.DB  // Database connection
}

// BuildHandler sets up the HTTP routing and builds an HTTP handler.
func BuildHandler(logger log.Logger, db *dbcontext.DB,
	cfg *config.Config, version string, trans ut.Translator) http.Handler {

	// Create `echo` instance.
	e := echo.New()

	// Logging middlware.
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

	// Recover from panic middleware.
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisablePrintStack: true,
	}))

	// Add Trailing slash for consistent URIs.
	e.Pre(middleware.AddTrailingSlash())

	// Pass-the-context middleware.
	e.Use(newCustomContextMiddleware(trans))

	// Register a custom fields validator.
	e.Validator = &CustomValidator{validator: validator.New()}

	// Register a custom binder.
	e.Binder = &CustomBinder{b: &echo.DefaultBinder{}}

	// Setup a custom HTTP error handler.
	e.HTTPErrorHandler = CustomHTTPErrorHandler

	// Healthcheck endpoint.
	healthcheck.RegisterHandlers(e, version)

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

// CustomValidator holds custom validator.
type CustomValidator struct {
	validator *validator.Validate
}

// Validate performs field validation.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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
	if err := cb.b.Bind(i, c); err != nil && err != echo.ErrUnsupportedMediaType {
		return err
	}
	return c.Validate(i)
}

// CustomHTTPErrorHandler handles errors encountered during HTTP request
// processing.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	l := c.Logger()
	cc, ok := c.(*customContext)
	if !ok {
		l.Errorf("failed to cast echo context to custom context")
	}
	res := errors.BuildErrorResponse(err, cc.trans)
	if res.StatusCode() == http.StatusInternalServerError {
		l.Errorf("encountered internal server error: %v", err)
	}
	if err = c.JSON(res.StatusCode(), res); err != nil {
		l.Errorf("failed writing error response: %v", err)
	}
}
