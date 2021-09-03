// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	e "errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/internal/resetpwd"
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
	// reset password generates a password reset token. The hash of the token
	// is stored in the database, a GUID is also generated to retrieve the
	// document when the user send the new password from the html form.
	ResetPassword(ctx context.Context, email string) (ResetPasswordResponse, error)
}

type ResetPasswordResponse struct {
	token string
	guid  string
	user  entity.User
}

// Identity represents an authenticated user identity.
type Identity interface {
	// ID returns the user ID.
	ID() string
	// IsAdmin return true if the user have admin priviliges.
	IsAdmin() bool
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
	userSvc         user.Service
	resetpwd        resetpwd.Service
}

// NewService creates a new authentication service.
func NewService(signingKey string, tokenExpiration int,
	logger log.Logger, sec Securer, userSvc user.Service,
	resetPassword resetpwd.Service) Service {
	return service{signingKey, tokenExpiration, logger, sec,
		userSvc, resetPassword}
}

// Login authenticates a user and generates a JWT token if authentication
// succeeds. Otherwise, an error is returned.
func (s service) Login(ctx context.Context, username, password string) (
	string, error) {
	logger := s.logger.With(ctx, "user", username)
	username = strings.ToLower(username)
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

	user, err := s.userSvc.Get(ctx, username)
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
		"id":      identity.ID(),
		"isAdmin": identity.IsAdmin(),
		"exp":     time.Now().Add(time.Duration(s.tokenExpiration) * time.Hour).Unix(),
	}).SignedString([]byte(s.signingKey))
}

func (s service) ResetPassword(ctx context.Context, email string) (
	ResetPasswordResponse, error) {

	user, err := s.userSvc.GetByEmail(ctx, email)
	if err != nil {
		return ResetPasswordResponse{}, err
	}

	rpt, err := s.resetpwd.Create(ctx, user.Username)
	if err != nil {
		return ResetPasswordResponse{}, err
	}

	resp := ResetPasswordResponse{
		token: rpt.Token,
		guid:  rpt.ID,
		user:  user.User,
	}

	return resp, nil
}
