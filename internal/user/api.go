// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

func RegisterHandlers(g *echo.Group, service Service,
	authHandler echo.MiddlewareFunc, logger log.Logger) {

	res := resource{service, logger}

	g.GET("/users/:username/", res.get)
	g.POST("/users/:username/", res.create)
	// g.DELETE("/users/:username/", user.DeleteUser, m.RequireLogin)

	// r.Get("/albums/<id>", res.get)
	// r.Get("/albums", res.query)

	// r.Use(authHandler)

	// // the following endpoints require a valid JWT
	// r.Post("/albums", res.create)
	// r.Put("/albums/<id>", res.update)
	// r.Delete("/albums/<id>", res.delete)
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

	return c.JSON(http.StatusOK, user)
}

func (r resource) create(c echo.Context) error {
	var input CreateUserRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return errors.BadRequest("")
	}

	user, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, user)
}
