// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	e "errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/internal/user"
	"github.com/saferwall/saferwall-api/pkg/log"
)

var (
	errUserNotFound     = e.New("user not found")
	errWrongPassword    = e.New("wrong password")
	errUserNotConfirmed = e.New("account non confirmed")
)

// Service encapsulates the authentication logic.
type Service interface {
	// authenticate authenticates a user using username and password.
	// It returns a JWT token if authentication succeeds. Otherwise, an error is returned.
	Login(ctx context.Context, username, password string) (string, error)
}

// Identity represents an authenticated user identity.
type Identity interface {
	// ID returns the user ID.
	ID() string
	// Name returns the user name.
	Name() string
}

// Securer represents security interface.
type Securer interface {
	HashMatchesPassword(string, string) bool
}

type service struct {
	signingKey      string
	tokenExpiration int
	logger          log.Logger
	sec             Securer
	userService     user.Service
}

// NewService creates a new authentication service.
func NewService(signingKey string, tokenExpiration int,
	logger log.Logger, sec Securer, userService user.Service) Service {
	return service{signingKey, tokenExpiration, logger, sec, userService}
}

// Login authenticates a user and generates a JWT token if authentication
// succeeds. Otherwise, an error is returned.
func (s service) Login(ctx context.Context, username, password string) (
	string, error) {
	logger := s.logger.With(ctx, "user", username)
	identity, err := s.authenticate(ctx, username, password)
	if err != nil {
		logger.Debugf(err.Error())
		return "", errors.Unauthorized(err.Error())
	}

	logger.Debug("authentication successful")
	return s.generateJWT(identity)
}

// authenticate authenticates a user using username and password. If username
// and password are correct, an identity is returned. Otherwise, nil is returned.
func (s service) authenticate(ctx context.Context, username, password string) (
	Identity, error) {

	user, err := s.userService.Get(ctx, username)
	if err != nil {
		return nil, errUserNotFound
	}
	if !s.sec.HashMatchesPassword(user.Password, password) {
		return nil, errWrongPassword
	}
	if !user.Confirmed {
		return nil, errUserNotConfirmed
	}
	return entity.User{Username: username}, nil
}

// generateJWT generates a JWT that encodes an identity.
func (s service) generateJWT(identity Identity) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   identity.ID(),
		"name": identity.Name(),
		"exp":  time.Now().Add(time.Duration(s.tokenExpiration) * time.Hour).Unix(),
	}).SignedString([]byte(s.signingKey))
}
