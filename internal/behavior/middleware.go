// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package behavior

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	e "github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

type middleware struct {
	service Service
	logger  log.Logger
}

// NewMiddleware creates a new file Middleware.
func NewMiddleware(service Service, logger log.Logger) middleware {
	return middleware{service, logger}
}

// VerifyID is the middleware function.
func (m middleware) VerifyID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		id := strings.ToLower(c.Param("id"))
		if !entity.IsValidID(id) {
			m.logger.Error("failed to match behavior scan ID %s", id)
			return e.BadRequest("invalid behavior scan id")
		}
		docExists, err := m.service.Exists(c.Request().Context(), id)
		if err != nil {
			return err
		}
		if !docExists {
			return db.ErrDocumentNotFound
		}

		return next(c)
	}
}
