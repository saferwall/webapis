// Package resetpwd implements `ResetPasswordTokens` for couchbase driver.
package resetpwd

import (
	"context"
	"encoding/hex"
	"hash"
	"time"

	store "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	pr "github.com/saferwall/saferwall-api/pkg/passwordreset"
)

// Service represents the password reset token management service.
type Service struct {
	db              *store.DB
	h               hash.Hash
	tokenExpiration int
}

// New initializes the token generation service.
func New(db *store.DB, h hash.Hash, exp int) Service {
	return Service{db, h, exp}
}

// Hash hashes a stream of bytes using sha2 algorihtm.
func (s Service) Hash(b []byte) string {
	s.h.Reset()
	s.h.Write(b)
	return hex.EncodeToString(s.h.Sum(nil))
}

// Create creates new reset password token.
func (s Service) Create(ctx context.Context, ownerID string) (
	pr.ResetPasswordToken, error) {
	secret, err := pr.NewResetPasswordSecret()
	if err != nil {
		return pr.ResetPasswordToken{}, err
	}

	ID := entity.ID()
	rpt := pr.ResetPasswordToken{
		Token:      secret.String(),
		Secret:     s.Hash([]byte(secret.String())),
		ID:         ID,
		OwnerID:    ownerID,
		Expiration: time.Now().Add(time.Duration(s.tokenExpiration) * time.Minute).Unix(),
	}

	err = s.db.Create(ctx, ID, rpt)
	if err != nil {
		return pr.ResetPasswordToken{}, err
	}

	return rpt, nil
}

// GetByOwnerID retrieves ResetPasswordToken by ownerID
func (s Service) GetByOwnerID(ctx context.Context, ownerID string) (
	pr.ResetPasswordToken, error) {
	rpt := pr.ResetPasswordToken{}
	err := s.db.Get(ctx, ownerID, &rpt)
	if err != nil {
		return pr.ResetPasswordToken{}, err
	}
	return rpt, nil
}

// Delete deletes ResetPasswordToken by ResetPasswordSecret
func (s Service) Delete(ctx context.Context, ownerID string) error {
	return s.db.Delete(ctx, ownerID)
}
