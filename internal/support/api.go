// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package support

import (
	"net/http"

	"github.com/MicahParks/recaptcha"
	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/internal/mailer"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// SupportEmailRequest represents a new email support request.
type SupportEmailRequest struct {
	Email             string `form:"email" validate:"required,email"`
	Subject           string `form:"subject" validate:"required,min=5,max=70"`
	Message           string `form:"message" validate:"required,min=10,max=1000"`
	RecaptchaResponse string `form:"g-recaptcha-response" validate:"required,min=1"`
}

func RegisterHandlers(e *echo.Echo, logger log.Logger, mailer mailer.Mailer, verifier recaptcha.VerifierV3) {
	res := resource{logger, mailer, verifier}
	e.POST("/contact/", res.contact)

}

type resource struct {
	logger   log.Logger
	mailer   mailer.Mailer
	verifier recaptcha.VerifierV3
}


// @Summary Contact Us
// @Description Handles contact us form-data sent via landing page.
// @Tags Support
// @Produce json
// @Param email body string true "The user's email"
// @Param subject body string true "The subject of the email"
// @Param message body string true "The content of the email"
// @Param g-recaptcha-response body string true "Google Recaptcha v3 response"
// @Success 200 {object} object{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /contact/ [post]
func (r resource) contact(c echo.Context) error {
	var input SupportEmailRequest
	ctx := c.Request().Context()
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Info(err)
		return err
	}

	remoteAddr := c.Request().RemoteAddr
	response, err := r.verifier.Verify(ctx, input.RecaptchaResponse, remoteAddr)
	if err != nil {
		r.logger.Errorf("Failed to verify reCAPTCHA: %v", err)
		return err
	}

	r.logger.Infof("reCAPTCHA V3 response: %#v", response)
	err = response.Check(recaptcha.V3ResponseCheckOptions{
		Action:   []string{"submit"},
		Hostname: []string{"saferwall.com"},
		Score:    0.5,
	})
	if err != nil {
		r.logger.Errorf("Failed check for reCAPTCHA response: %v", err)
		return errors.BadRequest("captcha failed failed")
	}

	go func() {
		err := r.mailer.Send(input.Message,
			input.Subject, input.Email, "support@saferwall.com", nil)
		if err != nil {
			r.logger.Errorf("failed to send email: %v", err)
		}
	}()

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}
