// Copyright 2018 Saferwall. All rights reserved.
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

const (
	KB = 1000
)

func RegisterHandlers(g *echo.Group, service Service, maxAvatarSize int,
	requireLogin, optionalLogin, verifyUser echo.MiddlewareFunc,
	logger log.Logger, mailer Mailer, templater tpl.Service) {

	res := resource{service, logger, mailer, templater, int64(maxAvatarSize*KB)}

	g.POST("/users/", res.create)
	g.GET("/users/", res.list, requireLogin)

	g.GET("/users/:username/", res.get, verifyUser, optionalLogin)
	g.PATCH("/users/:username/", res.update, verifyUser, requireLogin)
	g.PATCH("/users/:username/password/", res.password, verifyUser, requireLogin)
	g.PATCH("/users/:username/email/", res.email, verifyUser, requireLogin)
	g.DELETE("/users/:username/", res.delete, verifyUser, requireLogin)
	g.GET("/users/activities/", res.activities, optionalLogin)
	g.GET("/users/:username/likes/", res.likes, verifyUser, optionalLogin)
	g.GET("/users/:username/following/", res.following, verifyUser, optionalLogin)
	g.GET("/users/:username/followers/", res.followers, verifyUser, optionalLogin)
	g.GET("/users/:username/submissions/", res.submissions, verifyUser, optionalLogin)
	g.GET("/users/:username/comments/", res.comments, verifyUser, optionalLogin)
	g.POST("/users/:username/follow/", res.follow, verifyUser, requireLogin)
	g.POST("/users/:username/unfollow/", res.unFollow, verifyUser, requireLogin)
	g.POST("/users/:username/avatar/", res.avatar, verifyUser, requireLogin)
}

// Mailer represents the mailer interface.
type Mailer interface {
	Send(body, subject, from, to string) error
}

type resource struct {
	service       Service
	logger        log.Logger
	mailer        Mailer
	templater     tpl.Service
	maxAvatarSize int64 // expressed in bytes.
}

// @Summary Get user information by user ID
// @Description Retrieves information about a user.
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

	// Always hide the password.
	user.Password = ""
	return c.JSON(http.StatusOK, user)
}

// @Summary Create a new user
// @Description Create a new user.
// @Tags user
// @Accept json
// @Produce json
// @Param data body CreateUserRequest true "User data"
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
		resp.Token + "&guid=" + resp.Guid
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
// @Description Replace a user document with a new user's document.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param data body UpdateUserRequest true "New user data"
// @Success 200 {object} entity.User
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/ [patch]
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

// @Summary Deletes a user
// @Description Deletes a user by ID.
// @Tags user
// @Produce json
// @Param username path string true "Username"
// @Success 204 {object} entity.User
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/ [delete]
func (r resource) delete(c echo.Context) error {

	var isAdmin bool
	ctx := c.Request().Context()
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		isAdmin = user.IsAdmin()
	}
	if !isAdmin {
		return errors.Forbidden("")
	}

	user, err := r.service.Delete(ctx, c.Param("username"))
	if err != nil {
		return err
	}
	user.Email = ""
	user.Password = ""
	return c.JSON(http.StatusOK, user)
}

// @Summary Retrieves a paginated list of users
// @Description List users.
// @Tags user
// @Accept json
// @Produce json
// @Param per_page query uint false "Number of items per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages{items=[]entity.User}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/ [get]
func (r resource) list(c echo.Context) error {
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

// @Summary Returns a paginated list of a user's activities
// @Description List of activities of a user.
// @Tags activity
// @Accept json
// @Produce json
// @Param per_page query uint false "Number of items per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/activities/ [get]
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

// @Summary Returns a paginated list of a user's likes
// @Description List of likes of a user.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param per_page query uint false "Number of items per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/likes/ [get]
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

// @Summary Returns a paginated list of a user's following
// @Description List of users a user follows.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param per_page query uint false "Number of items per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/following/ [get]
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

// @Summary Returns a paginated list of a user's followers
// @Description List of users who follow a user.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param per_page query uint false "Number of items per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/followers/ [get]
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

// @Summary Returns a paginated list of a user's submissions
// @Description List of submissions by a user.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param per_page query uint false "Number of items per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/submissions/ [get]
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

// @Summary Returns a paginated list of a user's comments
// @Description List of comments by a user.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param per_page query uint false "Number of items per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/comments/ [get]
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

// @Summary Follow a user
// @Description Start following a user.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Target user to follow"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/follow/ [post]
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

// @Summary Unfollow a user
// @Description Stop following a user.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Target user to unfollow"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/unfollow/ [post]
func (r resource) unFollow(c echo.Context) error {
	ctx := c.Request().Context()
	err := r.service.UnFollow(ctx, c.Param("username"))
	if err != nil {
		switch err {
		case errUserSelfFollow:
			return errors.Forbidden("You can't un-follow yourself.")
		}
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

// @Summary Update user avatar
// @Description Change user avatar
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param data body UpdateEmailRequest true "User data"
// @Success 200 {object} object{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 413 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/avatar/ [post]
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

	if f.Size > r.maxAvatarSize {
		r.logger.With(ctx).Infof("image size too large: %v", f.Size)
		return errors.TooLargeEntity("The file size is too large, maximum allowed: 1MB")
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

// @Summary Update password for authenticated users
// @Description Change password for logged-in users.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param data body UpdateUserRequest true "User data"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/password/ [post]
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

// @Summary Update email for authenticated users
// @Description Change email for logged-in users.
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param data body UpdateEmailRequest true "User data"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /users/{username}/email/ [post]
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
