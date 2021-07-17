package server

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/labstack/echo/v4"
)

// customContext is the new context in the request / response cycle
// We can use it to pass some dependencies to the handlers.
type customContext struct {
	echo.Context
	trans  ut.Translator
}

func (c *customContext) Error(err error) {
	c.Echo().HTTPErrorHandler(err, c)
}

func newCustomContextMiddleware(trans ut.Translator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &customContext{c, trans}
			if err := next(cc); err != nil {
				cc.Error(err)
			}
			return nil
		}
	}
}
