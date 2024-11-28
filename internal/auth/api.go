// Copyright 2018 Saferwall. All rights reserved.
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
	"github.com/saferwall/saferwall-api/internal/mailer"
	"github.com/saferwall/saferwall-api/pkg/log"
)

const (
	msgEmailSent  = "A request to reset your password has been sent to your email address"
	msgPwdChanged = "You password has been successfully changed."
)

type resource struct {
	service   Service
	logger    log.Logger
	mailer    mailer.Mailer
	templater tpl.Service
	UIAddress string
}

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(g *echo.Group, service Service, logger log.Logger,
	mailer mailer.Mailer, templater tpl.Service, UIAddress string) {

	res := resource{service, logger, mailer, templater, UIAddress}

	g.POST("/auth/login/", res.login)
	g.DELETE("/auth/logout/", res.logout)
	g.POST("/auth/reset-password/", res.resetPassword)
	g.POST("/auth/password/", res.createNewPassword)
	g.GET("/auth/verify-account/", res.verifyAccount)
	g.POST("/auth/resend-confirmation/", res.resendConfirmation)

}

// loginRequest describes a login authentication request.
type loginRequest struct {
	Username string `json:"username" validate:"required,username_or_email" example:"mrrobot or mr-robot@protonmail.com"`
	Password string `json:"password" validate:"required,min=8,max=30" example:"control123"`
}

// resetPasswordRequest describes a password reset request for anonymous users.
type resetPwdRequest struct {
	Email string `json:"email" validate:"required,email" example:"mike@protonmail.com"`
}

// createNewPwdRequest describes a request to create a new password via a JWT token
// received in email.
type createNewPwdRequest struct {
	Token    string `json:"token" validate:"required" example:"eyJhbGciOiJIUzI1Ni"`
	GUID     string `json:"guid" validate:"required" example:"f47ac10b-58cc-8372-8567-0e02b2c3d479"`
	Password string `json:"password" validate:"required,min=8,max=30" example:"secretControl"`
}

// confirmAccountRequest describes an account confirmation email request.
type confirmAccountRequest struct {
	Email string `json:"email" validate:"required,email" example:"mike@protonmail.com"`
}

// @Summary Log in
// @Description Users logins by username and password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param auth-request body loginRequest true "Username and password"
// @Success 200 {string} json "{"token": "value"}"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/login/ [post]
func (r resource) login(c echo.Context) error {
	ctx := c.Request().Context()
	req := loginRequest{}
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Errorf("invalid request: %v", err)
		return errors.BadRequest("Invalid username or password")
	}

	loginResponse, err := r.service.Login(ctx, req.Username, req.Password)
	if err != nil {
		return errors.Unauthorized("Invalid username or password")
	}

	cookie := &http.Cookie{
		Value:    loginResponse.token,
		HttpOnly: true,
		Path:     "/",
		Name:     jwtCookieName,
		Domain:   c.Request().Host,
		Expires:  time.Now().Add(time.Duration(72 * time.Hour)),
		SameSite: http.SameSiteLaxMode,
	}

	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, struct {
		Token    string `json:"token"`
		Username string `json:"username"`
	}{loginResponse.token, loginResponse.username})
}

// @Summary Log out from current session
// @Description Delete the cookie used for authentication.
// @Tags Authentication
// @Success 204 "logout success"
// @Router /auth/logout/ [delete]
func (r resource) logout(c echo.Context) error {

	// Delete the cookie by setting a cookie with
	// the same name and an expired date.
	cookie := &http.Cookie{
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Name:     jwtCookieName,
		Domain:   c.Request().Host,
		Expires:  time.Unix(0, 0),
	}

	c.SetCookie(cookie)
	return c.NoContent(204)
}

// @Summary Confirm a new account creation
// @Description Verify the JWT token received during account creation.
// @Tags Authentication
// @Param guid query string true "GUID to identify the token"
// @Param token query string true "JWT token generated for account creation"
// @Success 200 {string} json "{"token": "value"}"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/verify-account/ [get]
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

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})

}

// @Summary Reset password for non-logged users by email
// @Description Request a reset password for anonymous users.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param reset-pwd body resetPwdRequest true "Email used during account sign-up"
// @Success 200 {string} json "{"token": "value"}"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/reset-password/ [post]
func (r resource) resetPassword(c echo.Context) error {
	ctx := c.Request().Context()
	req := resetPwdRequest{}
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

	templateData := struct {
		Username     string
		Token        string
		Guid         string
		HelpURL      string
		SupportEmail string
	}{
		Username:     resp.username,
		Token:        resp.token,
		Guid:         resp.guid,
		HelpURL:      "https://about.saferwall.com/",
		SupportEmail: "contact@saferwall.com",
	}

	resetPasswordTpl := r.templater.EmailRequestTemplate[tpl.ResetPassword]
	if err = resetPasswordTpl.Execute(templateData, body); err != nil {
		return err
	}

	go func() {
		var attachments []mailer.Attachment
		for _, attachment := range resetPasswordTpl.InlineImgs {
			attachments = append(attachments, attachment)
		}
		_ = r.mailer.Send(body.String(), resetPasswordTpl.Subject,
			resetPasswordTpl.From, req.Email, attachments)
	}()

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{msgEmailSent, http.StatusOK})

}

// @Summary Create a new password from a token received in email
// @Description Update the password from the auth token received in email.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param reset-pwd body createNewPwdRequest true "New password request"
// @Success 200 {string} json "{"token": "value"}"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/password/ [post]
func (r resource) createNewPassword(c echo.Context) error {
	req := createNewPwdRequest{}
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
	}{msgPwdChanged, http.StatusOK})

}

// @Summary Resend a confirmation email
// @Description Send a new confirmation email link to confirm user's account.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param reset-pwd body confirmAccountRequest true "Account confirmation request"
// @Success 200 {string} json "{"token": "value"}"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/resend-confirmation/ [post]
func (r resource) resendConfirmation(c echo.Context) error {
	var req confirmAccountRequest
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Infof("invalid request: %v", err)
		return errors.BadRequest(err.Error())
	}

	resp, err := r.service.ResendConfirmation(ctx, req.Email)
	if err != nil {
		return c.JSON(http.StatusOK, struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
		}{msgEmailSent, http.StatusOK})
	}

	body := new(bytes.Buffer)

	templateData := struct {
		Username     string
		Token        string
		Guid         string
		LoginURL     string
		LiveChatURL  string
		HelpURL      string
		SupportEmail string
	}{
		Username:     resp.username,
		Token:        resp.token,
		Guid:         resp.guid,
		LoginURL:     "https://saferwall.com/auth/login",
		LiveChatURL:  "https://discord.gg/an37PYHeZP",
		HelpURL:      "https://about.saferwall.com/",
		SupportEmail: "contact@saferwall.com",
	}

	confirmAccountTpl := r.templater.EmailRequestTemplate[tpl.ConfirmAccount]
	if err = confirmAccountTpl.Execute(templateData, body); err != nil {
		return err
	}

	go func() {
		var attachments []mailer.Attachment
		for _, attachment := range confirmAccountTpl.InlineImgs {
			attachments = append(attachments, attachment)
		}
		_ = r.mailer.Send(body.String(), confirmAccountTpl.Subject,
			confirmAccountTpl.From, req.Email, attachments)
	}()

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})

}
