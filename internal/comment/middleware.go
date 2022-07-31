// Copyright 2022 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

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

// NewMiddleware creates a new user Middleware.
func NewMiddleware(service Service, logger log.Logger) middleware {
	return middleware{service, logger}
}

// VerifyID validates the comment ID and check if the comment exists.
func (m middleware) VerifyID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		commentID := strings.ToLower(c.Param("id"))
		if !entity.IsValidID(commentID) {
			m.logger.Error("failed to match regex for comment ID %v", commentID)
			return e.BadRequest("invalid comment ID string")
		}

		docExists, err := m.service.Exists(c.Request().Context(), commentID)
		if err != nil {
			return err
		}

		if !docExists {
			return db.ErrDocumentNotFound
		}

		return next(c)
	}
}
