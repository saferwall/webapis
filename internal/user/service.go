// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"context"
	"time"

	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Service encapsulates usecase logic for users.
type Service interface {
	Get(ctx context.Context, id string) (User, error)
	// Query(ctx context.Context, offset, limit int) ([]User, error)
	// Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateUserRequest) (User, error)
	// Update(ctx context.Context, id string, input UpdateUserRequest) (User, error)
	// Delete(ctx context.Context, id string) (User, error)
}

// User represents the data about a user.
type User struct {
	entity.User
}

type service struct {
	repo   Repository
	logger log.Logger
}

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,alphanum,min=1,max=20"`
	Password string `json:"password" validate:"required,alphanum,min=8,max=30"`
}

// Validate validates the CreateUserRequest fields.
// func (m CreateUserRequest) Validate() error {
// 	err := validate.Struct(m)
// 	validationErrors := err.(validator.ValidationErrors)
// 	return validationErrors
// }

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	Name string `json:"name"`
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
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
		Username: req.Username,
		Email: req.Email,
		MemberSince: now.Unix(),
		LastSeen: now.Unix(),
	})
	if err != nil {
		return User{}, err
	}
	return s.Get(ctx, req.Username)
}
