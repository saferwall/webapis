// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package secure

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

// Secret stores secret of registration token.
type Secret [32]byte

// TokenGenerator is interface for working with tokens.
type TokenGenerator interface {
	// Create creates a new token.
	Create(ctx context.Context, ownerID string) (Token, error)
	// GetByOwnerID retrieves the Token object by id.
	GetByID(ctx context.Context, id string) (Token, error)
	// Delete deletes the Token object by id.
	Delete(ctx context.Context, id string) error
	// Hash hashes the clear text token and returns a string.
	Hash(ctx context.Context, b []byte) string
	// HashMatchesToken matches hash with token.
	HashMatchesToken(ctx context.Context, hash, token string) bool
}

// Password is abstract interface for dealing with password security.
type Password interface {
	// HashPassword hashes the password.
	HashPassword(password string) string
	// HashMatchesPassword matches hash with password.
	HashMatchesPassword(hash, password string) bool
}

// Token represents a token model in the database.
type Token struct {
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

// NewSecret generates a random stream of bytes.
func NewSecret() (Secret, error) {
	var b [32]byte

	_, err := rand.Read(b[:])
	if err != nil {
		return b, err
	}

	return b, nil
}

// SecretFromBase64 creates new reset password secret from base64 string
func SecretFromBase64(s string) (Secret, error) {
	var secret Secret

	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return secret, err
	}

	copy(secret[:], b)

	return secret, nil
}

// String implements Stringer, it returns a URL-safe, base64 encoded string
// of the secret.
func (secret Secret) String() string {
	return base64.URLEncoding.EncodeToString(secret[:])
}
