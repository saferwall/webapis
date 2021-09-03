// Package passwordreset provides password reset capabilities.
package passwordreset

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

// ResetPasswordTokens is interface for working with reset password tokens
type ResetPasswordTokens interface {
	// Create creates new reset password token
	Create(ctx context.Context, ownerID string) (ResetPasswordToken, error)
	// GetByOwnerID retrieves ResetPasswordToken by ownerID
	GetByOwnerID(ctx context.Context, ownerID string) (ResetPasswordToken, error)
	// Delete deletes ResetPasswordToken by ResetPasswordSecret
	Delete(ctx context.Context, secret ResetPasswordSecret) error
}

// ResetPasswordSecret stores secret of registration token.
type ResetPasswordSecret [32]byte

// ResetPasswordToken describing reset password model in the database.
type ResetPasswordToken struct {
	// Token in non-hashed format.
	// Useful to return back to the user, but not stored in the DB.
	Token string `json:"-"`
	// Secret stores the hash of the token.
	Secret string `json:"secret"`
	// OwnerID stores current token owner ID.
	OwnerID string `json:"ownerID"`
	// ID to uniquely identify a doc in the DB. This field is also not stored
	// in the DB, but only used to quickly find the doc when the user send back
	// the user send back the clear text hash.
	ID string `json:"-"`
	// Expiration represents the time the token will be  expired.
	Expiration int64 `json:"exp"`
}

// String implements Stringer, it returns a URL-safe, base64 encoded string
// of the secret.
func (secret ResetPasswordSecret) String() string {
	return base64.URLEncoding.EncodeToString(secret[:])
}

// NewResetPasswordSecret creates new reset password secret
func NewResetPasswordSecret() (ResetPasswordSecret, error) {
	var b [32]byte

	_, err := rand.Read(b[:])
	if err != nil {
		return b, err
	}

	return b, nil
}

// ResetPasswordSecretFromBase64 creates new reset password secret from base64 string
func ResetPasswordSecretFromBase64(s string) (ResetPasswordSecret, error) {
	var secret ResetPasswordSecret

	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return secret, err
	}

	copy(secret[:], b)

	return secret, nil
}
