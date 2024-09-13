// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/db"
	e "github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

var (
	sha256reg = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

type middleware struct {
	service Service
	logger  log.Logger
}

// NewMiddleware creates a new file Middleware.
func NewMiddleware(service Service, logger log.Logger) middleware {
	return middleware{service, logger}
}

// VerifyHash is the middleware function.
func (m middleware) VerifyHash(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		sha256 := strings.ToLower(c.Param("sha256"))
		if !sha256reg.MatchString(sha256) {
			m.logger.Error("failed to match sha256 regex for doc %v", sha256)
			return e.BadRequest("invalid sha256 hash")
		}

		// Change the <sha256> path parameter to lower case. This will reflect on
		// any handler that uses `VerifyHash` middleware.
		c.SetParamValues(sha256)

		docExists, err := m.service.Exists(c.Request().Context(), sha256)
		if err != nil {
			return err
		}
		if !docExists {
			return db.ErrDocumentNotFound
		}

		return next(c)
	}
}
