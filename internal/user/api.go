// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
	"github.com/saferwall/saferwall-api/pkg/pagination"
)

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin echo.MiddlewareFunc, optionalLogin echo.MiddlewareFunc,
	logger log.Logger) {

	res := resource{service, logger}

	g.POST("/users/", res.create)
	//g.GET("/users/:username/", res.get)
	//g.PATCH("/users/:username/", res.update, requireLogin)
	//g.DELETE("/users/:username/", res.delete, requireLogin)

	g.GET("/users/activities/", res.activities, optionalLogin)
	// g.GET("/users/:username/likes", res.likes, requireLogin)
	// g.GET("/users/:username/following", res.following, requireLogin)
	//g.GET("/users/:username/followers", res.followers, requireLogin)
	// g.GET("/users/:username/submissions", res.submissions)
	// g.GET("/users/:username/comments", res.comments)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c echo.Context) error {
	user, err := r.service.Get(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusOK, user)
}

func (r resource) create(c echo.Context) error {
	var input CreateUserRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return err
	}

	user, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusCreated, user)
}

func (r resource) update(c echo.Context) error {
	var input UpdateUserRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return err
	}

	user, err := r.service.Update(c.Request().Context(),
		c.Param("username"), input)
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusCreated, user)
}

func (r resource) delete(c echo.Context) error {
	user, err := r.service.Delete(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusOK, user)
}

func (r resource) activities(c echo.Context) error {
	ctx := c.Request().Context()
	var id string
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		id = user.ID()
	}
	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	activities, err := r.service.Activities(ctx, id, pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = activities
	return c.JSON(http.StatusOK, pages)
}
