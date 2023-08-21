// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package behavior

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin, verifyID echo.MiddlewareFunc, logger log.Logger) {

	res := resource{service, logger}

	g.GET("/behaviors/:id/", res.get, verifyID)

}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c echo.Context) error {

	var fields []string
	if fieldsParam := c.QueryParam("fields"); fieldsParam != "" {
		fields = strings.Split(fieldsParam, ",")
	}

	if len(fields) > 0 {
		allowed := areFieldsAllowed(fields)
		if !allowed {
			return errors.BadRequest("field not allowed")
		}
	}
	behavior, err := r.service.Get(c.Request().Context(), c.Param("id"), fields)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, behavior)
}