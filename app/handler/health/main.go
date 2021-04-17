// Copyright 2020 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
}