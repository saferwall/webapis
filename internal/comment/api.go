// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
	"github.com/saferwall/saferwall-api/pkg/pagination"
)

// contextKey defines a custom time to get/set values from a context.
type contextKey int

const (
	// filtersKey identifies the current filters during the request life.
	filtersKey contextKey = iota
)

type resource struct {
	service Service
	logger  log.Logger
}

func RegisterHandlers(g *echo.Group, service Service,
	logger log.Logger, requireLogin echo.MiddlewareFunc,
	verifyID echo.MiddlewareFunc) {

	res := resource{service, logger}

	g.GET("/comments/", res.list)
	g.GET("/comments/:id/", res.get, verifyID)
	g.POST("/comments/", res.create, requireLogin)
	g.PATCH("/comments/:id/", res.update, verifyID, requireLogin)
	g.DELETE("/comments/:id/", res.delete, verifyID, requireLogin)
}

// @Summary Retrieves a paginated list of comments
// @Description List comments
// @Tags Comment
// @Accept json
// @Produce json
// @Param per_page query uint false "Number of comments  per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages{items=[]entity.Comment}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /comments/ [get]
func (r resource) list(c echo.Context) error {
	ctx := c.Request().Context()

	// the `fields` query parameter is used to limit the fields
	// to include in the response.
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

	if len(c.QueryParams()) > 0 || (len(c.QueryParams()) == 1 && c.QueryParam("fields") != "") {
		queryParams := c.QueryParams()
		delete(queryParams, pagination.PageSizeVar)
		delete(queryParams, pagination.PageVar)
		if len(queryParams) > 0 {
			ctx = WithFilters(ctx, queryParams)
		}
	}

	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	files, err := r.service.Query(ctx, pages.Offset(), pages.Limit(), fields)
	if err != nil {
		return err
	}
	pages.Items = files
	return c.JSON(http.StatusOK, pages)
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

	// the `fields` query parameter is used to limit the fields
	// to include in the response.
	var fields []string
	if fieldsParam := c.QueryParam("fields"); fieldsParam != "" {
		fields = strings.Split(fieldsParam, ",")
	}

	comment, err := r.service.Get(ctx, c.Param("id"), fields)
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

	comment, err := r.service.Get(ctx, c.Param("id"), nil)
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

// WithFilters returns a context that contains the API filters.
func WithFilters(ctx context.Context, value map[string][]string) context.Context {
	return context.WithValue(ctx, filtersKey, value)
}
