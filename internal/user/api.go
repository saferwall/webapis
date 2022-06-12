// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	tpl "github.com/saferwall/saferwall-api/internal/template"
	"github.com/saferwall/saferwall-api/pkg/log"
	"github.com/saferwall/saferwall-api/pkg/pagination"
)

func RegisterHandlers(g *echo.Group, service Service,
	requireLogin, optionalLogin, verifyUser echo.MiddlewareFunc,
	logger log.Logger, mailer Mailer, templater tpl.Service) {

	res := resource{service, logger, mailer, templater}

	g.POST("/users/", res.create)
	g.GET("/users/:username/", res.get, optionalLogin, verifyUser)
	g.PATCH("/users/:username/", res.update, requireLogin)
	g.PATCH("/users/:username/password/", res.password, requireLogin)
	g.PATCH("/users/:username/email/", res.email, requireLogin)
	g.DELETE("/users/:username/", res.delete, requireLogin)
	g.GET("/users/", res.getUsers, requireLogin)
	g.POST("/users/resend-confirmation/", res.resendConfirmation)
	g.GET("/users/activities/", res.activities, optionalLogin)
	g.GET("/users/:username/likes/", res.likes, optionalLogin)
	g.GET("/users/:username/following/", res.following, optionalLogin)
	g.GET("/users/:username/followers/", res.followers, optionalLogin)
	g.GET("/users/:username/submissions/", res.submissions, optionalLogin)
	g.GET("/users/:username/comments/", res.comments, optionalLogin)
	g.POST("/users/:username/follow/", res.follow, requireLogin)
	g.POST("/users/:username/unfollow/", res.unfollow, requireLogin)
	g.POST("/users/:username/avatar/", res.avatar, requireLogin)
}

// Mailer represents the mailer interface.
type Mailer interface {
	Send(body, subject, from, to string) error
}

type resource struct {
	service   Service
	logger    log.Logger
	mailer    Mailer
	templater tpl.Service
}

// @Summary Get user information by user ID
// @Description Retrieves information about a user
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "User ID"
// @Success 200 {object} entity.User
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username} [get]
func (r resource) get(c echo.Context) error {
	ctx := c.Request().Context()
	user, err := r.service.Get(ctx, c.Param("username"))
	if err != nil {
		return err
	}

	// Hide the email unless the logged-in user is asking its own
	// information.
	curUser, ok := ctx.Value(entity.UserKey).(entity.User)
	if !ok || curUser.ID() != strings.ToLower(c.Param("username")) {
		user.Email = ""
	}

	// Always hide the apssword.
	user.Password = ""
	return c.JSON(http.StatusOK, user)
}

// @Summary Create a new user
// @Description Create a new user.
// @Tags user
// @Accept json
// @Produce json
// @Param data body CreateUserRequest true  "User data"
// @Success 201 {object} entity.User
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 413 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/ [post]
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

	// Hide sensible data,
	user.Password = ""

	// No need to generate a confirmation email when smtp
	// is not configured.
	if len(r.templater.EmailRequestTemplate) == 0 {
		user.Confirmed = true
		err = r.service.Patch(ctx, user.ID(), "confirmed", true)
		if err != nil {
			r.logger.With(ctx).Errorf("failed to confirm user: %v", err)
			return err
		}
		return c.JSON(http.StatusCreated, user)
	}

	resp, err := r.service.GenerateConfirmationEmail(ctx, user)
	if err != nil {
		r.logger.With(ctx).Errorf("generate confirmation email failed: %v", err)
		return err
	}

	body := new(bytes.Buffer)
	actionURL := c.Request().Host + "/v1/auth/verify-account/?token=" +
		resp.token + "&guid=" + resp.guid
	templateData := struct {
		Username     string
		ActionURL    string
		LoginURL     string
		LiveChatURL  string
		HelpURL      string
		SupportEmail string
	}{
		Username:     user.Username,
		ActionURL:    actionURL,
		LoginURL:     "https://saferwall.com/auth/login",
		LiveChatURL:  "https://discord.gg/an37PYHeZP",
		HelpURL:      "https://about.saferwall.com/",
		SupportEmail: "contact@saferwall.com",
	}

	confirmAccountTpl := r.templater.EmailRequestTemplate[tpl.ConfirmAccount]
	if err = confirmAccountTpl.Execute(templateData, body); err != nil {
		return err
	}

	go r.mailer.Send(body.String(),
		confirmAccountTpl.Subject, confirmAccountTpl.From, user.Email)

	return c.JSON(http.StatusCreated, user)
}

// @Summary Update a user object (full update)
// @Description Replace a user document with a new user's document
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} entity.User
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username} [put]
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

	if curUser != strings.ToLower(c.Param("username")) {
		return errors.Forbidden("")
	}

	user, err := r.service.Update(ctx, c.Param("username"), input)
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusOK, user)
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
	c.Response().Header().Set("Cache-Control", "public, max-age=120")
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
		switch err {
		case errUserSelfFollow:
			return errors.Forbidden("You can't follow yourself.")
		}
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
		switch err {
		case errUserSelfFollow:
			return errors.Forbidden("You can't unfollow yourself.")
		}
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

func (r resource) avatar(c echo.Context) error {
	var curUsername string
	ctx := c.Request().Context()

	targetUsername := strings.ToLower(c.Param("username"))

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		curUsername = user.ID()
	}

	if curUsername != targetUsername {
		return errors.Forbidden("")
	}

	f, err := c.FormFile("file")
	if err != nil {
		r.logger.With(ctx).Info(err)
		return errors.BadRequest("missing file in form request")
	}

	if f.Size > 1000000 {
		r.logger.With(ctx).Infof("image size too large: %v", f.Size)
		return errors.TooLargeEntity("The file size is too large, maximm allowed: 1MB")
	}

	src, err := f.Open()
	if err != nil {
		r.logger.With(ctx).Error(err)
		return errors.InternalServerError("")
	}
	defer src.Close()

	err = r.service.UpdateAvatar(ctx, curUsername, src)
	if err != nil {
		switch err {
		case errImageFormatNotSupported:
			return errors.UnsupportedMediaType("The image format is not supported.")
		default:
			r.logger.With(ctx).Error(err)
			return err
		}

	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

// password handle update password request for authenticated users.
func (r resource) password(c echo.Context) error {
	var req UpdatePasswordRequest
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Errorf("invalid request: %v", err)
		return err
	}

	var curUsername string
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		curUsername = user.ID()
	}

	if curUsername != strings.ToLower(c.Param("username")) {
		return errors.Forbidden("")
	}

	err := r.service.UpdatePassword(ctx, req)
	if err != nil {
		switch err {
		case errWrongPassword:
			return errors.Forbidden("")
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

// email handle update email request for authenticated users.
func (r resource) email(c echo.Context) error {
	var req UpdateEmailRequest
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Errorf("invalid request: %v", err)
		return err
	}

	var curUsername string
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		curUsername = user.ID()
	}

	if curUsername != strings.ToLower(c.Param("username")) {
		return errors.Forbidden("")
	}

	err := r.service.UpdateEmail(ctx, req)
	if err != nil {
		if err != nil {
			switch err {
			case errWrongPassword:
				return errors.Forbidden("")
			default:
				return err
			}
		}
	}

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

func (r resource) resendConfirmation(c echo.Context) error {

	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		r.logger.With(ctx).Infof("invalid request: %v", err)
		return errors.BadRequest(err.Error())
	}

	user, err := r.service.GetByEmail(ctx, req.Email)
	if err != nil {
		r.logger.With(ctx).Errorf("get by email failed: %v", err)
		return c.JSON(http.StatusOK, struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
		}{"ok", http.StatusOK})
	}

	resp, err := r.service.GenerateConfirmationEmail(ctx, user)
	if err != nil {
		r.logger.With(ctx).Errorf("generate confirmation email failed: %v", err)
		return c.JSON(http.StatusOK, struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
		}{"ok", http.StatusOK})
	}

	body := new(bytes.Buffer)
	link := c.Request().Host + "/v1/auth/verify-account/?token=" +
		resp.token + "&guid=" + resp.guid
	templateData := struct {
		Username     string
		ActionURL    string
		LoginURL     string
		LiveChatURL  string
		HelpURL      string
		SupportEmail string
	}{
		Username:     resp.username,
		ActionURL:    link,
		LoginURL:     "https://saferwall.com/auth/login",
		LiveChatURL:  "https://discord.gg/an37PYHeZP",
		HelpURL:      "https://about.saferwall.com/",
		SupportEmail: "contact@saferwall.com",
	}

	confirmAccountTpl := r.templater.EmailRequestTemplate[tpl.ConfirmAccount]
	if err = confirmAccountTpl.Execute(templateData, body); err != nil {
		return err
	}

	if r.mailer != nil {
		go r.mailer.Send(body.String(),
			confirmAccountTpl.Subject, confirmAccountTpl.From, req.Email)
	}

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})

}
