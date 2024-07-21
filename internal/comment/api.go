// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
)

type resource struct {
	service Service
	logger  log.Logger
}

func RegisterHandlers(g *echo.Group, service Service,
	logger log.Logger, requireLogin echo.MiddlewareFunc,
	verifyID echo.MiddlewareFunc) {

	res := resource{service, logger}

	g.GET("/comments/:id/", res.get, verifyID)
	g.POST("/comments/", res.create, requireLogin)
	g.PATCH("/comments/:id/", res.update, verifyID, requireLogin)
	g.DELETE("/comments/:id/", res.delete, verifyID, requireLogin)
}

// @Summary Create a new comment
// @Description Create a new comment.
// @Tags Comment
// @Accept json
// @Produce json
// @Param data body CreateCommentRequest true "Comment body"
// @Success 201 {object} entity.Comment
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 413 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /comments/ [post]
// @Security Bearer
func (r resource) create(c echo.Context) error {

	var input CreateCommentRequest
	ctx := c.Request().Context()
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Info(err)
		return err
	}
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		input.Username = user.ID()
	}
	com, err := r.service.Create(ctx, input)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, com)
}

// @Summary Get comment by ID
// @Description Retrieves information about a comment.
// @Tags Comment
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Success 200 {object} entity.Comment
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /comments/{id} [get]
func (r resource) get(c echo.Context) error {
	ctx := c.Request().Context()
	comment, err := r.service.Get(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, comment)
}

// @Summary Update a comment object (full update)
// @Description Replace a cocument with a new comment's document.
// @Tags Comment
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param data body UpdateCommentRequest true "New comment data"
// @Success 200 {object} entity.Comment
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /comments/{id}/ [patch]
// @Security Bearer
func (r resource) update(c echo.Context) error {
	var input UpdateCommentRequest

	ctx := c.Request().Context()

	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Info(err)
		return err
	}

	comment, err := r.service.Update(ctx, c.Param("id"), input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, comment)
}

// @Summary Deletes a comment
// @Description Deletes a comment by ID.
// @Tags Comment
// @Produce json
// @Param id path string true "Comment ID"
// @Success 204 {object} entity.Comment
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /comments/{id}/ [delete]
// @Security Bearer
func (r resource) delete(c echo.Context) error {

	var curUsername string
	ctx := c.Request().Context()

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		curUsername = user.ID()
	}

	comment, err := r.service.Get(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	if comment.Username != curUsername {
		return errors.Forbidden("")
	}

	comment, err = r.service.Delete(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, comment)
}
