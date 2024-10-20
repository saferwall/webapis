// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package behavior

import (
	"net/http"
	"strconv"
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

// CacheResponse is the middleware function for handing HTTP caching.
func (m middleware) CacheResponse(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Retrieve an ETag for the resource.
		bhvReport, err := m.service.Get(c.Request().Context(), c.Param("id"),
			[]string{"doc.last_updated", "status"})
		if err != nil {
			m.logger.Errorf("failed to get behavior report: %v", err)
			return next(c)
		}

		// Skip caching when dynamic scan status != finished.
		if bhvReport.Status != entity.FileScanProgressFinished {
			return next(c)
		}

		etag := strconv.FormatInt(bhvReport.Meta.LastUpdated, 10)
		if etag == "" {
			m.logger.Errorf("bhvReport.Meta.LastUpdated is not set: %v", err)
			return next(c)
		}

		// Check If-None-Match header from request.
		ifNoneMatch := c.Request().Header.Get("If-None-Match")
		if ifNoneMatch == etag {
			// Cache hit !
			return c.NoContent(http.StatusNotModified)
		}

		// Cache miss, set headers
		c.Response().Header().Set("ETag", etag)
		c.Response().Header().Set("Cache-Control", "max-age=3600, must-revalidate")
		return next(c)
	}
}
