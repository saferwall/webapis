// Package token implements `TokenGenerator` for couchbase driver.
package token

import (
	"context"
	"encoding/hex"
	"hash"
	"time"

	store "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/secure"
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

// Create creates new reset password token.
func (s Service) Create(ctx context.Context, ownerID string) (
	secure.Token, error) {
	secret, err := secure.NewSecret()
	if err != nil {
		return secure.Token{}, err
	}

	ID := entity.ID()
	token := secure.Token{
		Token:      secret.String(),
		Secret:     s.Hash(ctx, []byte(secret.String())),
		ID:         ID,
		OwnerID:    ownerID,
		Expiration: time.Now().Add(time.Duration(s.tokenExpiration) * time.Minute).Unix(),
	}

	err = s.db.Create(ctx, ID, token)
	if err != nil {
		return secure.Token{}, err
	}

	return token, nil
}

// GetByID retrieves ResetPasswordToken by ownerID
func (s Service) GetByID(ctx context.Context, id string) (
	secure.Token, error) {
	token := secure.Token{}
	err := s.db.Get(ctx, id, &token)
	if err != nil {
		return secure.Token{}, err
	}
	return token, nil
}

// Delete deletes ResetPasswordToken by ResetPasswordSecret
func (s Service) Delete(ctx context.Context, id string) error {
	return s.db.Delete(ctx, id)
}

// Hash hashes a stream of bytes using sha2 algorihtm.
func (s Service) Hash(ctx context.Context, b []byte) string {
	s.h.Reset()
	s.h.Write(b)
	return hex.EncodeToString(s.h.Sum(nil))
}

// HashMatchesToken matches hash with token. Returns true if hash and
// token match.
func (s Service) HashMatchesToken(ctx context.Context, hash, token string) bool {
	return hash == s.Hash(ctx, []byte(token))
}
