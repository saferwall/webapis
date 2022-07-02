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
	"github.com/saferwall/saferwall-api/internal/secure"
	"github.com/saferwall/saferwall-api/internal/user"
	"github.com/saferwall/saferwall-api/pkg/log"
)

var (
	errUserNotFound     = e.New("user not found")
	errUserNotConfirmed = e.New("account non confirmed")
	errWrongPassword    = e.New("wrong password")
	errExpiredToken     = e.New("token expired")
	errMalformedToken   = e.New("malformed token")
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
	// VerifyAccount confirms the user account by verifying the token.
	VerifyAccount(ctx context.Context, id, token string) error
	// create a new password if the user has already a reset password token.
	CreateNewPassword(ctx context.Context, id, password, token string) error
	// resend a new confirmation email for the user's account.
	ResendConfirmation(ctx context.Context, email string) (ResendConfirmationResponse, error)
}

type ResetPasswordResponse struct {
	token    string
	guid     string
	username string
}

type ResendConfirmationResponse struct {
	token    string
	guid     string
	username string
}

// Identity represents an authenticated user identity.
type Identity interface {
	// ID returns the user ID.
	ID() string
	// IsAdmin return true if the user have admin priviliges.
	IsAdmin() bool
}

type service struct {
	signingKey      string
	tokenExpiration int
	logger          log.Logger
	sec             secure.Password
	tokenGen        secure.TokenGenerator
	userSvc         user.Service
}

// NewService creates a new authentication service.
func NewService(signingKey string, tokenExpiration int,
	logger log.Logger, sec secure.Password, userSvc user.Service,
	tokenGen secure.TokenGenerator) Service {
	return service{signingKey, tokenExpiration, logger, sec, tokenGen, userSvc}
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

func (s service) VerifyAccount(ctx context.Context, id,
	token string) error {

	confirmAccTok, err := s.tokenGen.GetByID(ctx, id)
	if err != nil {
		return err
	}

	exp := time.Unix(confirmAccTok.Expiration, 0)
	if exp.Before(time.Now()) {
		return errExpiredToken
	}

	if !s.tokenGen.HashMatchesToken(ctx, confirmAccTok.Secret, token) {
		return errMalformedToken
	}

	userID := strings.ToLower(confirmAccTok.OwnerID)
	err = s.userSvc.Patch(ctx, userID, "confirmed", true)
	if err != nil {
		return err
	}
	return s.tokenGen.Delete(ctx, id)
}

func (s service) ResetPassword(ctx context.Context, email string) (
	ResetPasswordResponse, error) {

	user, err := s.userSvc.GetByEmail(ctx, email)
	if err != nil {
		if err.Error() == "user not found" {
			return ResetPasswordResponse{}, errUserNotFound
		}
		return ResetPasswordResponse{}, err
	}

	rpt, err := s.tokenGen.Create(ctx, user.Username)
	if err != nil {
		return ResetPasswordResponse{}, err
	}

	resp := ResetPasswordResponse{
		token:    rpt.Token,
		guid:     rpt.ID,
		username: user.Username,
	}

	return resp, nil
}

func (s service) CreateNewPassword(ctx context.Context, id,
	token, password string) error {

	confirmAccTok, err := s.tokenGen.GetByID(ctx, id)
	if err != nil {
		return err
	}

	exp := time.Unix(confirmAccTok.Expiration, 0)
	if exp.Before(time.Now()) {
		return errExpiredToken
	}

	if !s.tokenGen.HashMatchesToken(ctx, confirmAccTok.Secret, token) {
		return errMalformedToken
	}

	passwordHash := s.sec.HashPassword(password)
	userID := strings.ToLower(confirmAccTok.OwnerID)
	err = s.userSvc.Patch(ctx, userID, "password", passwordHash)
	if err != nil {
		return err
	}
	return s.tokenGen.Delete(ctx, id)
}

func (s service) ResendConfirmation(ctx context.Context, email string) (ResendConfirmationResponse, error) {

	user, err := s.userSvc.GetByEmail(ctx, email)
	if err != nil {
		s.logger.With(ctx).Errorf("get by email failed: %v", err)
		return ResendConfirmationResponse{}, err
	}

	resp, err := s.userSvc.GenerateConfirmationEmail(ctx, user)
	if err != nil {
		s.logger.With(ctx).Errorf("generate confirmation email failed: %v", err)
		return ResendConfirmationResponse{}, err
	}

	r := ResendConfirmationResponse{
		token:    resp.Token,
		guid:     resp.Guid,
		username: user.Username,
	}

	return r, nil
}
