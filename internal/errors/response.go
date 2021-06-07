// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package errors

import (
	"errors"
	"net/http"

	gocb "github.com/couchbase/gocb/v2"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

// ErrorResponse is the response that represents an error.
type ErrorResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error is required by the error interface.
func (e ErrorResponse) Error() string {
	return e.Message
}

// StatusCode is required by routing.HTTPError interface.
func (e ErrorResponse) StatusCode() int {
	return e.Status
}

// InternalServerError creates a new error response representing an internal
// server error (HTTP 500)
func InternalServerError(msg string) ErrorResponse {
	if msg == "" {
		msg = "We encountered an error while processing your request."
	}
	return ErrorResponse{
		Status:  http.StatusInternalServerError,
		Message: msg,
	}
}

// NotFound creates a new error response representing a resource-not-found
// error (HTTP 404)
func NotFound(msg string) ErrorResponse {
	if msg == "" {
		msg = "The requested resource was not found."
	}
	return ErrorResponse{
		Status:  http.StatusNotFound,
		Message: msg,
	}
}

// Unauthorized creates a new error response representing an
// authentication/authorization failure (HTTP 401)
func Unauthorized(msg string) ErrorResponse {
	if msg == "" {
		msg = "You are not authenticated to perform the requested action."
	}
	return ErrorResponse{
		Status:  http.StatusUnauthorized,
		Message: msg,
	}
}

// Forbidden creates a new error response representing an authorization.
// failure (HTTP 403)
func Forbidden(msg string) ErrorResponse {
	if msg == "" {
		msg = "You are not authorized to perform the requested action."
	}
	return ErrorResponse{
		Status:  http.StatusForbidden,
		Message: msg,
	}
}

// BadRequest creates a new error response representing a bad request
// (HTTP 400).
func BadRequest(msg string) ErrorResponse {
	if msg == "" {
		msg = "Your request is in a bad format."
	}
	return ErrorResponse{
		Status:  http.StatusBadRequest,
		Message: msg,
	}
}

// BuildErrorResponse builds an error response from an error.
func BuildErrorResponse(err error, trans ut.Translator) ErrorResponse {
	switch err.(type) {
	case ErrorResponse:
		return err.(ErrorResponse)
	case validator.ValidationErrors:
		return invalidInput(err, trans)
	case *echo.HTTPError:
		switch err.(*echo.HTTPError).Code {
		case http.StatusNotFound:
			return NotFound("")
		default:
			return ErrorResponse{
				Status:  err.(*echo.HTTPError).Code,
				Message: err.Error(),
			}
		}
	}
	if errors.Is(err, gocb.ErrDocumentNotFound) {
		return NotFound("")
	}
	return InternalServerError("")
}

// invalidInput creates a new error response representing a data validation
// error (HTTP 400).
func invalidInput(err error, trans ut.Translator) ErrorResponse {

	var details []string

	// Translate all error at once.
	validatorErrs := err.(validator.ValidationErrors)

	for _, e := range validatorErrs {
		details = append(details, e.Translate(trans))
	}

	return ErrorResponse{
		Status:  http.StatusBadRequest,
		Message: "There is some problem with the data you submitted.",
		Details: details,
	}
}
