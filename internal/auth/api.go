// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

const (
	msgEmailSent = "A password reset message was sent to your email address"
)

type resource struct {
	service Service
	logger  log.Logger
	mailer  Mailer
}

// Mailer represents the mailer interface/
type Mailer interface {
	Send(body, subject, from, to string) error
}

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(g *echo.Group, service Service, logger log.Logger,
	mailer Mailer) {

	res := resource{service, logger, mailer}

	g.POST("/auth/login/", res.login)
	g.POST("/auth/resend-confirmation/", res.resendConfirmation)
	g.POST("/auth/reset-password/", res.resetPassword)
}

// login handles authentication request.
func (r resource) login(c echo.Context) error {
	var req struct {
		Username string `json:"username" validate:"required,alphanum,min=1,max=20"`
		Password string `json:"password" validate:"required,min=8,max=30"`
	}
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Errorf("invalid request: %v", err)
		return errors.BadRequest("Invalid username or password")
	}

	token, err := r.service.Login(ctx, req.Username, req.Password)
	if err != nil {
		return errors.Unauthorized("Invalid username or password")
	}

	cookie := &http.Cookie{
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		Name:     jwtCookieName,
		Domain:   c.Request().Host,
		Expires:  time.Now().Add(time.Duration(72 * time.Hour)),
		SameSite: http.SameSiteLaxMode,
	}

	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, struct {
		Token string `json:"token"`
	}{token})
}

func (r resource) resendConfirmation(c echo.Context) error {
	return nil
}

func (r resource) resetPassword(c echo.Context) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Infof("invalid request: %v", err)
		return err
	}

	resp, err := r.service.ResetPassword(ctx, req.Email)
	if err != nil {
		r.logger.With(ctx).Infof("invalid request: %v", err)
		return c.JSON(http.StatusOK, struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
		}{msgEmailSent, http.StatusOK})
	}

	link := "https://saferwall.com/auth/reset-password?token=" +
		resp.token + "&token=" + "&guid=" + resp.guid

	go r.mailer.Send(link, "saferwall - reset password",
		"noreply@saferwall.com", req.Email)

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{msgEmailSent, http.StatusOK})

}
