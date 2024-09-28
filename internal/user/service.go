// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/h2non/filetype"

	"github.com/saferwall/saferwall-api/internal/activity"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/secure"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Service encapsulates use case logic for users.
type Service interface {
	Get(ctx context.Context, id string) (User, error)
	Query(ctx context.Context, offset, limit int) ([]User, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateUserRequest) (User, error)
	Update(ctx context.Context, id string, input interface{}) (User, error)
	Patch(ctx context.Context, id, path string, input interface{}) error
	Delete(ctx context.Context, id string) (User, error)
	Exists(ctx context.Context, id string) (bool, error)
	Activities(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Likes(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Followers(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Following(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Submissions(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Comments(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	CountActivities(ctx context.Context) (int, error)
	CountLikes(ctx context.Context, id string) (int, error)
	CountFollowing(ctx context.Context, id string) (int, error)
	CountFollowers(ctx context.Context, id string) (int, error)
	CountComments(ctx context.Context, id string) (int, error)
	CountSubmissions(ctx context.Context, id string) (int, error)
	Follow(ctx context.Context, id string) error
	UnFollow(ctx context.Context, id string) error
	GetByEmail(ctx context.Context, id string) (User, error)
	UpdateAvatar(ctx context.Context, id string, src io.Reader) error
	UpdatePassword(ctx context.Context, input UpdatePasswordRequest) error
	UpdateEmail(ctx context.Context, input UpdateEmailRequest) error
	GenerateConfirmationEmail(ctx context.Context, user User) (
		ConfirmAccountResponse, error)
	Like(ctx context.Context, id string, userLike entity.UserLike) error
	Unlike(ctx context.Context, id, sha256 string) error
}

var (
	// avatar upload timeout in seconds.
	avatarUploadTimeout        = time.Duration(time.Second * 10)
	errEmailAlreadyExists      = errors.New("email already exists")
	errUserAlreadyExists       = errors.New("username already exists")
	errUserNotFound            = errors.New("user not found")
	errUserAlreadyConfirmed    = errors.New("email already confirmed")
	errWrongPassword           = errors.New("wrong password")
	errUserSelfFollow          = errors.New("user can't self follow")
	errImageFormatNotSupported = errors.New("unsupported file type")
)

// User represents the data about a user.
type User struct {
	entity.User
}

// Uploader represents the file upload interface.
type Uploader interface {
	Upload(ctx context.Context, bucket, key string, file io.Reader) error
}

type service struct {
	repo     Repository
	logger   log.Logger
	tokenGen secure.TokenGenerator
	sec      secure.Password
	actSvc   activity.Service
	bucket   string
	objSto   Uploader
}

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email" example:"mike@protonmail.com"`
	Username string `json:"username" validate:"required,alphanum,min=1,max=20" example:"mike"`
	Password string `json:"password" validate:"required,min=8,max=30" example:"control123"`
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	Name     string `json:"name" validate:"omitempty,min=1,max=32" example:"Ibn Taymiyyah"`
	Location string `json:"location" validate:"omitempty,min=1,max=16" example:"Damascus"`
	URL      string `json:"url" validate:"omitempty,url,max=64" example:"https://en.wikipedia.org/wiki/Ibn_Taymiyyah"`
	Bio      string `json:"bio" validate:"omitempty,min=1,max=64" example:"What really counts are good endings, not flawed beginnings."`
}

// UpdatePasswordRequest represents a password update request.
type UpdatePasswordRequest struct {
	OldPassword string `json:"old" validate:"required,min=8,max=30" example:"control123"`
	NewPassword string `json:"new_password" validate:"required,necsfield=OldPassword,min=8,max=30" example:"secretControl"`
}

// UpdateEmailRequest represents an email update request.
type UpdateEmailRequest struct {
	Password string `json:"password" validate:"required,min=8,max=30" example:"control123"`
	NewEmail string `json:"email" validate:"required,email" example:"mike@proton.me"`
}

// ConfirmAccountResponse holds data coming from the token generator.
type ConfirmAccountResponse struct {
	Token    string
	Guid     string
	Username string
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger, tokenGen secure.TokenGenerator,
	sec secure.Password, bucket string, upl Uploader, actSvc activity.Service) Service {
	return service{repo, logger, tokenGen, sec, actSvc, bucket, upl}
}

// Get returns the user with the specified user ID.
func (s service) Get(ctx context.Context, id string) (User, error) {
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return User{}, err
	}
	return User{user}, nil
}

// Create creates a new user.
func (s service) Create(ctx context.Context, req CreateUserRequest) (
	User, error) {

	// check if the username is already taken.
	user, err := s.Get(ctx, req.Username)
	if err != nil && err.Error() != "document not found" {
		return User{}, err
	}
	if user.Username != "" {
		return User{}, errUserAlreadyExists
	}

	// check if the email already exists.
	present, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return User{}, err
	}
	if present {
		return User{}, errEmailAlreadyExists
	}

	now := time.Now()
	err = s.repo.Create(ctx, entity.User{
		Type:        "user",
		Username:    req.Username,
		Password:    s.sec.HashPassword(req.Password),
		Email:       strings.ToLower(req.Email),
		MemberSince: now.Unix(),
		LastSeen:    now.Unix(),
	})
	if err != nil {
		return User{}, err
	}

	// Set a default avatar for the user with the help of `Robohash`. A web
	// service used to generate avatars. Do not fail if these operations returns
	// an error.
	id := strings.ToLower(req.Username)
	url := fmt.Sprintf("https://robohash.org/%s?set=set1&bgset=bg1&size=200x200", id)
	buffer, err := downloadURLContent(url)
	if err != nil {
		s.logger.Errorf("failed to download user's avatar: %v", err)
	} else {
		err = s.objSto.Upload(ctx, s.bucket, id, buffer)
		if err != nil {
			s.logger.Errorf("failed to upload user's avatar: %v", err)
		}
	}

	return s.Get(ctx, req.Username)
}

// Update updates the user with the specified ID.
func (s service) Update(ctx context.Context, id string, req interface{}) (
	User, error) {

	user, err := s.Get(ctx, id)
	if err != nil {
		return user, err
	}
	data, err := json.Marshal(req)
	if err != nil {
		return user, err
	}
	err = json.Unmarshal(data, &user)
	if err != nil {
		return user, err
	}

	if err := s.repo.Update(ctx, user.User); err != nil {
		return user, err
	}

	return user, nil
}

// Patch performs an atomic user sub document update.
func (s service) Patch(ctx context.Context, id, path string,
	input interface{}) error {
	return s.repo.Patch(ctx, id, path, input)
}

// Delete deletes the user with the specified ID.
func (s service) Delete(ctx context.Context, id string) (User, error) {
	user, err := s.Get(ctx, id)
	if err != nil {
		return User{}, err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		return User{}, err
	}
	return user, nil
}

// Exists checks if a document exists for the given id.
func (s service) Exists(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
}

// Count returns the number of users.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Query returns the users with the specified offset and limit.
func (s service) Query(ctx context.Context, offset, limit int) (
	[]User, error) {

	items, err := s.repo.Query(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []User{}
	for _, item := range items {
		result = append(result, User{item})
	}
	return result, nil
}

func (s service) Activities(ctx context.Context, id string, offset, limit int) (
	[]interface{}, error) {

	result, err := s.repo.Activities(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) Following(ctx context.Context, id string, offset, limit int) (
	[]interface{}, error) {

	result, err := s.repo.Following(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) Followers(ctx context.Context, id string, offset, limit int) (
	[]interface{}, error) {

	result, err := s.repo.Followers(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) Likes(ctx context.Context, id string, offset, limit int) (
	[]interface{}, error) {

	result, err := s.repo.Likes(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) Submissions(ctx context.Context, id string, offset, limit int) (
	[]interface{}, error) {

	result, err := s.repo.Submissions(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) Comments(ctx context.Context, id string, offset, limit int) (
	[]interface{}, error) {

	result, err := s.repo.Comments(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) CountActivities(ctx context.Context) (int, error) {
	count, err := s.repo.CountActivities(ctx)
	if err != nil {
		return 0, err
	}
	return count, err
}

func (s service) CountLikes(ctx context.Context, id string) (int, error) {
	user, err := s.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return len(user.Likes), err
}

func (s service) CountFollowing(ctx context.Context, id string) (int, error) {
	user, err := s.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return len(user.Following), err
}

func (s service) CountFollowers(ctx context.Context, id string) (int, error) {
	user, err := s.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return len(user.Followers), err
}

func (s service) CountComments(ctx context.Context, id string) (int, error) {
	user, err := s.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return user.CommentsCount, err
}

func (s service) CountSubmissions(ctx context.Context, id string) (int, error) {
	user, err := s.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return len(user.Submissions), err
}

func (s service) Follow(ctx context.Context, id string) error {
	var err error
	targetUser, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	loggedInUser, _ := ctx.Value(entity.UserKey).(entity.User)
	curUser, err := s.Get(ctx, loggedInUser.ID())
	if err != nil {
		return err
	}

	// Get the source of the HTTP request from the ctx.
	source, _ := ctx.Value(entity.SourceKey).(string)

	curUsername := curUser.ID()
	targetUsername := targetUser.ID()

	if curUsername == targetUsername {
		return errUserSelfFollow
	}

	// Add new activity even if the user is already followed.
	if _, err = s.actSvc.Create(ctx, activity.CreateActivityRequest{
		Kind:     "follow",
		Username: curUser.Username,
		Target:   targetUser.Username,
		Source:   source,
	}); err != nil {
		return err
	}

	newFollow := entity.UserFollows{
		Username:  targetUser.Username,
		Timestamp: time.Now().Unix(),
	}
	if err = s.repo.Follow(ctx, curUser.Username, newFollow); err != nil {
		return err
	}
	return nil
}

func (s service) UnFollow(ctx context.Context, id string) error {
	var err error
	targetUser, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	loggedInUser, _ := ctx.Value(entity.UserKey).(entity.User)
	curUser, err := s.Get(ctx, loggedInUser.ID())
	if err != nil {
		return err
	}

	targetUsername := targetUser.ID()
	curUsername := curUser.ID()
	if curUsername == targetUsername {
		return errUserSelfFollow
	}

	if err = s.repo.Unfollow(ctx, curUser.Username, targetUser.Username); err != nil {
		return err
	}

	if err = s.repo.Update(ctx, curUser.User); err != nil {
		return err
	}
	if err = s.repo.Update(ctx, targetUser.User); err != nil {
		return err
	}

	return nil
}

func (s service) UpdatePassword(ctx context.Context, input UpdatePasswordRequest) error {

	var id string
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		id = user.ID()
	}

	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if !s.sec.HashMatchesPassword(user.Password, input.OldPassword) {
		return errWrongPassword
	}

	user.Password = s.sec.HashPassword(input.NewPassword)
	return s.repo.Update(ctx, user)
}

func (s service) UpdateEmail(ctx context.Context, input UpdateEmailRequest) error {

	var id string
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		id = user.ID()
	}

	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if !s.sec.HashMatchesPassword(user.Password, input.Password) {
		return errWrongPassword
	}

	return s.repo.Patch(ctx, id, "email", input.NewEmail)
}

func (s service) UpdateAvatar(ctx context.Context, id string, src io.Reader) error {

	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	fileContent, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	if !filetype.IsImage(fileContent) {
		return errImageFormatNotSupported
	}

	// Create a context with a timeout that will abort the upload if it takes
	// more than the passed in timeout.
	uploadCtx, cancelFn := context.WithTimeout(context.Background(),
		avatarUploadTimeout)

	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	defer cancelFn()

	err = s.objSto.Upload(uploadCtx, s.bucket, id, bytes.NewReader(fileContent))
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, user)
}

// GetByEmail returns the user given its email address.
func (s service) GetByEmail(ctx context.Context, email string) (User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}
	return User{user}, nil
}

func (s service) GenerateConfirmationEmail(ctx context.Context, user User) (
	ConfirmAccountResponse, error) {

	if user.Confirmed {
		return ConfirmAccountResponse{}, errUserAlreadyConfirmed
	}

	rpt, err := s.tokenGen.Create(ctx, user.ID())
	if err != nil {
		return ConfirmAccountResponse{}, err
	}

	resp := ConfirmAccountResponse{
		Username: user.Username,
		Token:    rpt.Token,
		Guid:     rpt.ID,
	}

	return resp, nil
}

func (s service) Like(ctx context.Context, id string, userLike entity.UserLike) error {
	return s.repo.Like(ctx, id, userLike)
}

func (s service) Unlike(ctx context.Context, id, sha256 string) error {
	return s.repo.Unlike(ctx, id, sha256)
}
