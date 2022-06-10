// Copyright 2022 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/db"
	e "github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

var (
	userReg = regexp.MustCompile(`^[a-zA-Z0-9]{1,20}$`)
)

type middleware struct {
	service Service
	logger  log.Logger
}

// NewMiddleware creates a new user Middleware.
func NewMiddleware(service Service, logger log.Logger) middleware {
	return middleware{service, logger}
}

// VerifyUser is the middleware function.
func (m middleware) VerifyUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		username := strings.ToLower(c.Param("username"))
		if !userReg.MatchString(username) {
			m.logger.Error("failed to match regex for username %v", username)
			return e.BadRequest("invalid username string")
		}
		docExists, err := m.service.Exists(c.Request().Context(), username)
		if err != nil {
			return err
		}
		if !docExists {
			return db.ErrDocumentNotFound
		}

		return next(c)
	}
}
