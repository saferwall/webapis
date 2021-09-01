// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package secure

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"hash"

	"golang.org/x/crypto/bcrypt"
)

// New initializes security service.
func New(h hash.Hash) *Service {
	return &Service{h: h}
}

// Service holds security related methods.
type Service struct {
	h hash.Hash
}

// HashPW hashes the password using bcrypt.
func (Service) HashPW(password string) string {
	hashedPW, _ := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	return string(hashedPW)
}

// HashMatchesPassword matches hash with password. Returns true if hash and
// password match.
func (Service) HashMatchesPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Token generates new unique token.
func (Service) Token(str string) string {
	token, _ := GenerateRandomStringURLSafe(32)
	return token
}

// Hash hashes a stream of bytes using sha2 algorihtm.
func (s Service) Hash(b []byte) string {
	s.h.Reset()
	s.h.Write(b)
	return hex.EncodeToString(s.h.Sum(nil))
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomStringURLSafe returns a URL-safe, base64 encoded securely
// generated random string. It will return an error if the system's secure
// random number generator fails to function correctly, in which case the
// caller should not continue.
func GenerateRandomStringURLSafe(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}
