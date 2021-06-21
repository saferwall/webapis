// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(g *echo.Group, service Service, logger log.Logger) {
	g.POST("/auth/login/", login(service, logger))
}

// login returns a handler that handles user login request.
func login(service Service, logger log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Username string `json:"username" validate:"required,alphanum,min=1,max=20"`
			Password string `json:"password" validate:"required,alphanum,min=8,max=30"`
		}

		if err := c.Bind(&req); err != nil {
			logger.With(c.Request().Context()).Errorf("invalid request: %v", err)
			return errors.BadRequest("")
		}

		token, err := service.Login(c.Request().Context(), req.Username, req.Password)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, struct {
			Token string `json:"token"`
		}{token})
	}
}
