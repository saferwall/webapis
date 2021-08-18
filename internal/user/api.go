// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
	"github.com/saferwall/saferwall-api/pkg/pagination"
)

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin echo.MiddlewareFunc, optionalLogin echo.MiddlewareFunc,
	logger log.Logger, mailer Mailer) {

	res := resource{service, logger, mailer}

	g.POST("/users/", res.create)
	g.GET("/users/:username/", res.get)
	g.PATCH("/users/:username/", res.update, requireLogin)
	g.DELETE("/users/:username/", res.delete, requireLogin)
	g.GET("/users/", res.getUsers)
	g.GET("/users/activities/", res.activities, optionalLogin)
	g.GET("/users/:username/likes/", res.likes, optionalLogin)
	g.GET("/users/:username/following/", res.following, optionalLogin)
	g.GET("/users/:username/followers/", res.followers, optionalLogin)
	g.GET("/users/:username/submissions/", res.submissions, optionalLogin)
	g.GET("/users/:username/comments/", res.comments, optionalLogin)
	g.POST("/users/:username/follow/", res.follow, requireLogin)
	g.POST("/users/:username/unfollow/", res.unfollow, requireLogin)
}

// Mailer represents the mailer interface/
type Mailer interface {
	Send(body, subject, from, to string) error
}

type resource struct {
	service Service
	logger  log.Logger
	mailer  Mailer
}

func (r resource) get(c echo.Context) error {
	user, err := r.service.Get(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusOK, user)
}

func (r resource) create(c echo.Context) error {
	var input CreateUserRequest
	ctx := c.Request().Context()
	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Info(err)
		return err
	}
	user, err := r.service.Create(ctx, input)
	if err != nil {
		switch err {
		case errEmailAlreadyExists:
			return errors.BadRequest(err.Error())
		case errUserAlreadyExists:
			return errors.BadRequest(err.Error())
		default:
			return err
		}
	}

	go r.mailer.Send("hello", "saferwall - confirm account", "noreply@saferwall.com",
					user.Email)
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusCreated, user)
}

func (r resource) update(c echo.Context) error {
	var input UpdateUserRequest
	var curUser string

	ctx := c.Request().Context()

	if err := c.Bind(&input); err != nil {
		r.logger.With(ctx).Info(err)
		return err
	}

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		curUser = user.ID()
	}

	if curUser != c.Param("username") {
		return errors.BadRequest("")
	}

	user, err := r.service.Update(ctx, c.Param("username"), input)
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusCreated, user)
}

func (r resource) delete(c echo.Context) error {

	var isAdmin bool
	ctx := c.Request().Context()
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		isAdmin = user.IsAdmin()
	}
	if !isAdmin {
		return errors.Forbidden("")
	}

	user, err := r.service.Delete(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusOK, user)
}

func (r resource) getUsers(c echo.Context) error {
	var isAdmin bool
	ctx := c.Request().Context()
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		isAdmin = user.IsAdmin()
	}
	if !isAdmin {
		return errors.Forbidden("")
	}

	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	users, err := r.service.Query(ctx, pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = users
	return c.JSON(http.StatusOK, pages)
}

func (r resource) activities(c echo.Context) error {
	ctx := c.Request().Context()
	var id string
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		id = user.ID()
	}
	count, err := r.service.CountActivities(ctx)
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	activities, err := r.service.Activities(ctx, id, pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = activities
	return c.JSON(http.StatusOK, pages)
}

func (r resource) likes(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.CountLikes(ctx, c.Param("username"))
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	activities, err := r.service.Likes(
		ctx, c.Param("username"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = activities
	return c.JSON(http.StatusOK, pages)
}

func (r resource) following(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.CountFollowing(ctx, c.Param("username"))
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	following, err := r.service.Following(
		ctx, c.Param("username"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = following
	return c.JSON(http.StatusOK, pages)
}

func (r resource) followers(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.CountFollowers(ctx, c.Param("username"))
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	followers, err := r.service.Followers(
		ctx, c.Param("username"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = followers
	return c.JSON(http.StatusOK, pages)
}

func (r resource) submissions(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.CountSubmissions(ctx, c.Param("username"))
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	submissions, err := r.service.Submissions(
		ctx, c.Param("username"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = submissions
	return c.JSON(http.StatusOK, pages)
}

func (r resource) comments(c echo.Context) error {
	ctx := c.Request().Context()
	count, err := r.service.CountComments(ctx, c.Param("username"))
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	comments, err := r.service.Comments(
		ctx, c.Param("username"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = comments
	return c.JSON(http.StatusOK, pages)
}

func (r resource) follow(c echo.Context) error {
	ctx := c.Request().Context()
	err := r.service.Follow(ctx, c.Param("username"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

func (r resource) unfollow(c echo.Context) error {
	ctx := c.Request().Context()
	err := r.service.Unfollow(ctx, c.Param("username"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}
