// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"context"
	"encoding/json"
	"time"

	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Service encapsulates usecase logic for users.
type Service interface {
	Get(ctx context.Context, id string) (User, error)
	//Query(ctx context.Context, offset, limit int) ([]User, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateUserRequest) (User, error)
	Update(ctx context.Context, id string, input UpdateUserRequest) (User, error)
	Delete(ctx context.Context, id string) (User, error)
}

// User represents the data about a user.
type User struct {
	entity.User
}

// Securer represents security interface.
type Securer interface {
	Hash(string) string
}

type service struct {
	sec    Securer
	repo   Repository
	logger log.Logger
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
	Bio      string `json:"bio" validate:"omitempty,alphanumunicode,min=1,max=256"`
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger, sec Securer) Service {
	return service{sec, repo, logger}
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

	now := time.Now()
	err := s.repo.Create(ctx, entity.User{
		Username:    req.Username,
		Password:    s.sec.Hash(req.Password),
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
func (s service) Update(ctx context.Context, id string, req UpdateUserRequest) (
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

	// check if user.Username == id
	if err := s.repo.Update(ctx, user.User); err != nil {
		return user, err
	}

	return user, nil
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
