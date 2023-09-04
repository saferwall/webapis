// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package behavior

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
	"github.com/saferwall/saferwall-api/pkg/pagination"
)

// contextKey defines a custom time to get/set values from a context.
type contextKey int

const (
	// filtersKey identifies the current filters during the request life.
	filtersKey contextKey = iota
)

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin, verifyID echo.MiddlewareFunc, logger log.Logger) {

	res := resource{service, logger}

	g.GET("/behaviors/:id/", res.get, verifyID)
	g.GET("/behaviors/:id/api-trace/", res.apis, verifyID)
	g.GET("/behaviors/:id/sys-events/", res.events, verifyID)

}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c echo.Context) error {

	var fields []string
	if fieldsParam := c.QueryParam("fields"); fieldsParam != "" {
		fields = strings.Split(fieldsParam, ",")
	}

	if len(fields) > 0 {
		allowed := areFieldsAllowed(fields)
		if !allowed {
			return errors.BadRequest("field not allowed")
		}
	}

	behavior, err := r.service.Get(c.Request().Context(), c.Param("id"), fields)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, behavior)
}

func (r resource) apis(c echo.Context) error {
	ctx := c.Request().Context()

	if len(c.QueryParams()) > 0 {
		ctx = WithFilters(ctx, c.QueryParams())
	}

	count, err := r.service.CountAPIs(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	apis, err := r.service.APIs(
		ctx, c.Param("id"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = apis
	return c.JSON(http.StatusOK, pages)
}

func (r resource) events(c echo.Context) error {
	ctx := c.Request().Context()

	if len(c.QueryParams()) > 0 {
		ctx = WithFilters(ctx, c.QueryParams())
	}

	count, err := r.service.CountEvents(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	events, err := r.service.Events(
		ctx, c.Param("id"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = events
	return c.JSON(http.StatusOK, pages)
}

// WithFilters returns a context that contains the API filters.
func WithFilters(ctx context.Context, value map[string][]string) context.Context {
	return context.WithValue(ctx, filtersKey, value)
}
