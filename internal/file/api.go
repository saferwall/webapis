// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin echo.MiddlewareFunc, logger log.Logger) {

	res := resource{service, logger}

	g.POST("/files/", res.create)
	g.GET("/files/:sha256/", res.get)
	g.PUT("/files/:sha256/", res.update, requireLogin)
	g.DELETE("/files/:sha256/", res.delete, requireLogin)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c echo.Context) error {
	file, err := r.service.Get(c.Request().Context(), c.Param("sha256"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, file)
}

func (r resource) create(c echo.Context) error {

	f, err := c.FormFile("file")
	if err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return errors.BadRequest("missing file in form request")
	}

	if f.Size > 60000000 {
		r.logger.With(c.Request().Context()).Info("payload too large")
		return errors.TooLargeEntity("")
	}

	src, err := f.Open()
	if err != nil {
		r.logger.With(c.Request().Context()).Error(err)
		return errors.InternalServerError("")
	}
	defer src.Close()

	input := CreateFileRequest{src, f.Filename,
		c.Request().Header.Get("X-Geoip-Country")}
	file, err := r.service.Create(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, file)
}

func (r resource) update(c echo.Context) error {
	var input UpdateFileRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return errors.BadRequest("")
	}

	file, err := r.service.Update(c.Request().Context(),
		c.Param("sha256"), input)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, file)
}

func (r resource) delete(c echo.Context) error {
	file, err := r.service.Delete(c.Request().Context(), c.Param("sha256"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, file)
}
