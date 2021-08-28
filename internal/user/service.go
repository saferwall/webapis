// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"context"
	"encoding/json"
	"errors"
	"time"

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
}

var (
	errEmailAlreadyExists = errors.New("email already exists")
	errUserAlreadyExists  = errors.New("username already exists")
)

// User represents the data about a user.
type User struct {
	entity.User
}

// Securer represents security interface.
type Securer interface {
	HashPW(string) string
}

type service struct {
	sec    Securer
	repo   Repository
	logger log.Logger
	actSvc activity.Service
}

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,alphanum,min=1,max=20"`
	Password string `json:"password" validate:"required,alphanum,min=8,max=30"`
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	Name     string `json:"name" validate:"omitempty,min=1,max=32"`
	Location string `json:"location" validate:"omitempty,alphanumunicode,min=2,max=16"`
	URL      string `json:"url" validate:"omitempty,url,max=64"`
	Bio      string `json:"bio" validate:"omitempty,printascii,min=1,max=64"`
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger,
	sec Securer, actSvc activity.Service) Service {
	return service{sec, repo, logger, actSvc}
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
	currentUser, err := s.Get(ctx, loggedInUser.ID())
	if err != nil {
		return err
	}

	currentUsername := currentUser.ID()
	targetUsername := targetUser.ID()

	if !isStringInSlice(targetUsername, currentUser.Following) {
		currentUser.Following = append(currentUser.Following, targetUsername)
		currentUser.FollowingCount += 1

		// add new activity
		if _, err = s.actSvc.Create(ctx, activity.CreateActivityRequest{
			Kind:     "follow",
			Username: currentUsername,
			Target:   targetUsername,
		}); err != nil {
			return err
		}
		if err = s.repo.Update(ctx, currentUser.User); err != nil {
			return err
		}

	}
	if !isStringInSlice(currentUsername, targetUser.Followers) {
		targetUser.Followers = append(targetUser.Followers, currentUsername)
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
	currentUser, err := s.Get(ctx, loggedInUser.ID())
	if err != nil {
		return err
	}

	targetUsername := targetUser.ID()
	currentUsername := currentUser.ID()

	if isStringInSlice(targetUsername, currentUser.Following) {
		currentUser.Following = removeStringFromSlice(
			currentUser.Following, targetUsername)
		currentUser.FollowingCount -= 1

		// delete corresponsing activity.
		if s.repo.DeleteActivity(ctx, "follow", currentUsername,
		 targetUsername) ; err != nil {
			return err
		}
	}
	if isStringInSlice(currentUsername, targetUser.Followers) {
		targetUser.Followers = removeStringFromSlice(
			targetUser.Followers, currentUsername)
		targetUser.FollowersCount -= 1
	}

	if s.repo.Update(ctx, currentUser.User); err != nil {
		return err
	}
	if s.repo.Update(ctx, targetUser.User); err != nil {
		return err
	}

	return nil
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
