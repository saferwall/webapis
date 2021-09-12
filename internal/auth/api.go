// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

import (
	"bytes"
	"net/http"
	"time"

	tpl "github.com/saferwall/saferwall-api/internal/template"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

const (
	msgEmailSent = "A request to reset your password has been sent to your email address"
)

type resource struct {
	service   Service
	logger    log.Logger
	mailer    Mailer
	templater tpl.Service
	UIAddress string
}

// Mailer represents the mailer interface/
type Mailer interface {
	Send(body, subject, from, to string) error
}

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(g *echo.Group, service Service, logger log.Logger,
	mailer Mailer, templater tpl.Service, UIAddress string) {

	res := resource{service, logger, mailer, templater, UIAddress}

	g.POST("/auth/login/", res.login)
	g.POST("/auth/reset-password/", res.resetPassword)
	g.POST("/auth/password/", res.createNewPassword)
	g.GET("/auth/verify-account/", res.verifyAccount)
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

func (r resource) verifyAccount(c echo.Context) error {
	ctx := c.Request().Context()
	err := r.service.VerifyAccount(ctx, c.QueryParam("guid"), c.QueryParam("token"))
	if err != nil {
		r.logger.With(ctx).Errorf("verify account failed: %v", err)
		switch err {
		case errExpiredToken:
			return errors.Unauthorized(err.Error())
		case errMalformedToken:
			return errors.BadRequest(err.Error())
		}
		return err
	}

	return c.Redirect(http.StatusPermanentRedirect, r.UIAddress)

}

func (r resource) resetPassword(c echo.Context) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Infof("invalid request: %v", err)
		return errors.BadRequest(err.Error())
	}

	resp, err := r.service.ResetPassword(ctx, req.Email)
	if err != nil {
		switch err {
		case errUserNotFound:
			return c.JSON(http.StatusOK, struct {
				Message string `json:"message"`
				Status  int    `json:"status"`
			}{msgEmailSent, http.StatusOK})
		}
		r.logger.With(ctx).Error("reset password failed: %v", err)
		return err
	}

	body := new(bytes.Buffer)
	link := c.Request().Host + "/v1/auth/password/?token=" +
		resp.token + "&guid=" + resp.guid
	templateData := struct {
		Username     string
		ActionURL    string
		HelpURL      string
		SupportEmail string
	}{
		Username:     resp.user.Username,
		ActionURL:    link,
		HelpURL:      "https://about.saferwall.com/",
		SupportEmail: "contact@saferwall.com",
	}

	resetPasswordTpl := r.templater.EmailRequestTemplate[tpl.ResetPassword]
	if err = resetPasswordTpl.Execute(templateData, body); err != nil {
		return err
	}

	go r.mailer.Send(body.String(),
		resetPasswordTpl.Subject, resetPasswordTpl.From, resp.user.Email)

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{msgEmailSent, http.StatusOK})

}

func (r resource) createNewPassword(c echo.Context) error {
	var req struct {
		Token    string `json:"token" validate:"required"`
		GUID     string `json:"guid" validate:"required"`
		Password string `json:"password" validate:"required,min=8,max=30"`
	}
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Errorf("invalid request: %v", err)
		return err
	}

	err := r.service.CreateNewPassword(ctx, req.GUID, req.Token, req.Password)
	if err != nil {
		r.logger.With(ctx).Errorf("create new password failed: %v", err)
		switch err {
		case errExpiredToken:
			return errors.Unauthorized(err.Error())
		case errMalformedToken:
			return errors.BadRequest(err.Error())
		}
		return err
	}

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{msgEmailSent, http.StatusOK})

}
