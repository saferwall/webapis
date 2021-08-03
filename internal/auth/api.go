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

type resource struct {
	service Service
	logger  log.Logger
}

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(g *echo.Group, service Service, logger log.Logger) {

	res := resource{service, logger}

	g.POST("/auth/login/", res.login)
	//g.POST("/auth/resend-confirmation/", res.resendConfirmation)
}

// login returns a handler that handles user login request.
func (r resource) login(c echo.Context) error {
	var req struct {
		Username string `json:"username" validate:"required,alphanum,min=1,max=20"`
		Password string `json:"password" validate:"required,min=8,max=30"`
	}

	r.logger.Info(c.Request().Host)
	if err := c.Bind(&req); err != nil {
		r.logger.With(c.Request().Context()).Errorf("invalid request: %v", err)
		return errors.BadRequest("")
	}

	token, err := r.service.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return err
	}

	cookie := JWTCookie(token, "localhost", 72)
	c.SetCookie(&cookie)

	return c.JSON(http.StatusOK, struct {
		Token string `json:"token"`
	}{token})
}


func (r resource) resendConfirmation(c echo.Context) error {
	return nil
}