// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/pkg/log"
)

type resource struct {
	service Service
	logger  log.Logger
}

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin echo.MiddlewareFunc, optionalLogin echo.MiddlewareFunc,
	logger log.Logger) {

	res := resource{service, logger}

	g.POST("/comments/", res.create, requireLogin)
	g.GET("/comments/:id/", res.get)

}

func (r resource) create(c echo.Context) error {

	var input CreateCommentRequest
	ctx := c.Request().Context()
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Info(err)
		return err
	}

	comment, err := r.service.Create(ctx, input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, comment)
}

func (r resource) get(c echo.Context) error {
	ctx := c.Request().Context()
	comment, err := r.service.Get(ctx, c.Param("if"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, comment)
}
