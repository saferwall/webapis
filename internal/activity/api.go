// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package activity

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/pkg/log"
	"github.com/saferwall/saferwall-api/pkg/pagination"
)

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin echo.MiddlewareFunc, logger log.Logger) {

	res := resource{service, logger}

	g.GET("/activities/:id", res.get)
	g.GET("/activities/", res.query)

}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c echo.Context) error {

	// the `fields` query parameter is used to limit the fields
	// to include in the response.
	var fields []string
	if fieldsParam := c.QueryParam("fields"); fieldsParam != "" {
		fields = strings.Split(fieldsParam, ",")
	}

	activity, err := r.service.Get(c.Request().Context(), c.Param("id"), fields)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, activity)
}

func (r resource) query(c echo.Context) error {

	ctx := c.Request().Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	activities, err := r.service.Query(ctx, pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = activities
	return c.JSON(http.StatusOK, pages)
}
