// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package errors

import "net/http"

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
