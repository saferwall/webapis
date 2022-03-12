// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

type resource struct {
	service Service
	logger  log.Logger
}

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin echo.MiddlewareFunc, logger log.Logger) {

	res := resource{service, logger}

	g.GET("/comments/:id/", res.get)
	g.POST("/comments/", res.create, requireLogin)
	g.PATCH("/comments/:id/", res.update, requireLogin)
	g.DELETE("/comments/:id/", res.delete, requireLogin)
}

func (r resource) create(c echo.Context) error {

	var input CreateCommentRequest
	ctx := c.Request().Context()
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Info(err)
		return err
	}
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		input.Username = user.ID()
	}
	com, err := r.service.Create(ctx, input)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, com)
}

func (r resource) get(c echo.Context) error {
	ctx := c.Request().Context()
	comment, err := r.service.Get(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, comment)
}

func (r resource) update(c echo.Context) error {
	var input UpdateCommentRequest

	ctx := c.Request().Context()

	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Info(err)
		return err
	}

	comment, err := r.service.Update(ctx, c.Param("id"), input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, comment)
}

func (r resource) delete(c echo.Context) error {

	var curUsername string
	ctx := c.Request().Context()

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		curUsername = user.ID()
	}

	comment, err := r.service.Get(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	if comment.Username != curUsername {
		return errors.Forbidden("")
	}

	comment, err = r.service.Delete(ctx, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, comment)
}
