// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
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

type toolBodyWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (r toolBodyWriter) Write(b []byte) (int, error) {
	return r.body.Write(b)
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
			m.logger.Errorf("failed to match sha256 regex for doc %v", sha256)
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

// CacheResponse is the middleware function for handing HTTP caching.
func (m middleware) CacheResponse(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Retrieve an ETag for the resource.
		file, err := m.service.Get(c.Request().Context(), c.Param("sha256"),
			[]string{"doc.last_updated"})
		if err != nil {
			m.logger.Errorf("failed to get file object %v", err)
			return next(c)
		}
		etag := strconv.FormatInt(file.Meta.LastUpdated, 10)
		if etag == "" {
			m.logger.Errorf("file.Meta.LastUpdated is not set: %v", err)
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

// ModifyResponse modifies the JSON response to include some metadata for the UI.
func (m middleware) ModifyResponse(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		var err error
		var isUI int

		writer := &toolBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Response().Writer}
		c.Response().Writer = writer

		if err = next(c); err != nil {
			return err
		}

		isUIStr := c.Request().Header.Get("X-Get-Ui")
		if len(isUIStr) > 0 {
			isUI, err = strconv.Atoi(isUIStr)
			if err != nil {
				return err
			}
		}

		// Determines the source of the API request, if it originates from the UI,
		// we want to attach some more UI metadata.
		if isUI != 1 {
			_, err := writer.ResponseWriter.Write(writer.body.Bytes())
			return err
		}

		metaUI, err := m.service.MetaUI(c.Request().Context(), c.Param("sha256"))
		if err != nil {
			return err
		}

		oldResponseBody := make(map[string]any)

		err = json.Unmarshal(writer.body.Bytes(), &oldResponseBody)
		if err != nil {
			return err
		}

		oldResponseBody["ui"] = metaUI

		newResponseBody, err := json.Marshal(oldResponseBody)
		if err != nil {
			return err
		}

		n, err := writer.ResponseWriter.Write(newResponseBody)
		if err != nil {
			return err
		}

		c.Response().Header().Set("Content-Length", strconv.Itoa(n))

		return nil
	}
}
