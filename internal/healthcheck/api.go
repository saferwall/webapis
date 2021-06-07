// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package healthcheck

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RegisterHandlers registers the handlers that perform healthchecks.
func RegisterHandlers(e *echo.Echo, version string) {
	e.Match([]string{"GET", "HEAD"}, "/healthcheck/", healthcheck(version))
}

// healthcheck responds to a healthcheck request.
func healthcheck(version string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "OK "+version)
	}
}
