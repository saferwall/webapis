// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/h2non/filetype"

	"github.com/saferwall/saferwall-api/internal/activity"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Service encapsulates usecase logic for users.
type Service interface {
	Get(ctx context.Context, id string) (User, error)
	Query(ctx context.Context, offset, limit int) ([]User, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateUserRequest) (User, error)
	Update(ctx context.Context, id string, input interface{}) (User, error)
	Patch(ctx context.Context, id, path string, input interface{}) error
	Delete(ctx context.Context, id string) (User, error)
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
	Unfollow(ctx context.Context, id string) error
	UpdateAvatar(ctx context.Context, id string, src io.Reader) error
	UpdatePassword(ctx context.Context, input UpdatePasswordRequest) error
}

var (
	// avatar upload timeout in seconds.
	avatarUploadTimeout   = time.Duration(time.Second * 10)
	errEmailAlreadyExists = errors.New("email already exists")
	errUserAlreadyExists  = errors.New("username already exists")
	errWrongPassword      = errors.New("wrong password")
	errUserSelfFollow     = errors.New(
		"source and target user in follow request is the same")
	errImageFormatNotSupported = errors.New("unsupported file type")
)

// User represents the data about a user.
type User struct {
	entity.User
}

// Securer represents security interface.
type Securer interface {
	HashPW(string) string
	HashMatchesPassword(string, string) bool
}

// Uploader represents the file upload interface.
type Uploader interface {
	Upload(ctx context.Context, bucket, key string, file io.Reader) error
}

type service struct {
	sec    Securer
	repo   Repository
	logger log.Logger
	actSvc activity.Service
	bucket string
	objsto Uploader
}

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,alphanum,min=1,max=20"`
	Password string `json:"password" validate:"required,min=8,max=30"`
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	Name     string `json:"name" validate:"omitempty,min=1,max=32"`
	Location string `json:"location" validate:"omitempty,min=1,max=16"`
	URL      string `json:"url" validate:"omitempty,url,max=64"`
	Bio      string `json:"bio" validate:"omitempty,min=1,max=64"`
}

// UpdatePasswordRequest represents a password update request.
type UpdatePasswordRequest struct {
	OldPassword string `json:"old" validate:"required,min=8,max=30"`
	NewPassword string `json:"new" validate:"required,necsfield=OldPassword,min=8,max=30"`
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger, sec Securer,
	bucket string, upl Uploader, actSvc activity.Service) Service {
	return service{sec, repo, logger, actSvc, bucket, upl}
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
		Password:    s.sec.HashPW(req.Password),
		Email:       req.Email,
		MemberSince: now.Unix(),
		LastSeen:    now.Unix(),
	})
	if err != nil {
		return User{}, err
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
	return user.LikesCount, err
}

func (s service) CountFollowing(ctx context.Context, id string) (int, error) {
	user, err := s.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return user.FollowingCount, err
}

func (s service) CountFollowers(ctx context.Context, id string) (int, error) {
	user, err := s.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return user.FollowersCount, err
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
	return user.SubmissionsCount, err
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

	curUsername := curUser.ID()
	targetUsername := targetUser.ID()

	if curUsername == targetUsername {
		return errUserSelfFollow
	}

	if !isStringInSlice(targetUsername, curUser.Following) {
		curUser.Following = append(curUser.Following, targetUsername)
		curUser.FollowingCount += 1

		// add new activity
		if _, err = s.actSvc.Create(ctx, activity.CreateActivityRequest{
			Kind:     "follow",
			Username: curUser.Username,
			Target:   targetUser.Username,
		}); err != nil {
			return err
		}
		if err = s.repo.Update(ctx, curUser.User); err != nil {
			return err
		}

	}
	if !isStringInSlice(curUsername, targetUser.Followers) {
		targetUser.Followers = append(targetUser.Followers, curUsername)
		targetUser.FollowersCount += 1
		return s.repo.Update(ctx, targetUser.User)
	}

	return nil
}

func (s service) Unfollow(ctx context.Context, id string) error {
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

	if isStringInSlice(targetUsername, curUser.Following) {
		curUser.Following = removeStringFromSlice(
			curUser.Following, targetUsername)
		curUser.FollowingCount -= 1

		// delete corresponsing activity.
		if s.repo.DeleteActivity(ctx, "follow", curUser.Username,
			targetUsername); err != nil {
			return err
		}
	}
	if isStringInSlice(curUsername, targetUser.Followers) {
		targetUser.Followers = removeStringFromSlice(
			targetUser.Followers, curUsername)
		targetUser.FollowersCount -= 1
	}

	if s.repo.Update(ctx, curUser.User); err != nil {
		return err
	}
	if s.repo.Update(ctx, targetUser.User); err != nil {
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

	user.Password = s.sec.HashPW(input.NewPassword)
	return s.repo.Update(ctx, user)
}

func (s service) UpdateAvatar(ctx context.Context, id string, src io.Reader) error {

	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	fileContent, err := ioutil.ReadAll(src)
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

	err = s.objsto.Upload(uploadCtx, s.bucket, id, bytes.NewReader(fileContent))
	if err != nil {
		return err
	}

	user.HasAvatar = true

	return s.repo.Update(ctx, user)
}

// isStringInSlice check if a string exist in a list of strings
func isStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

// removeStringFromSlice removes a string item from a list of strings.
func removeStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
