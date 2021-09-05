// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package password

import (
	"hash"

	"golang.org/x/crypto/bcrypt"
)

// Service holds security related methods.
type Service struct {
	h hash.Hash
}

// New initializes security service.
func New(h hash.Hash) Service {
	return Service{h: h}
}

// HashPassword hashes the password using bcrypt.
func (Service) HashPassword(password string) string {
	hashedPW, _ := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	return string(hashedPW)
}

// HashMatchesPassword matches hash with password. Returns true if hash and
// password match.
func (Service) HashMatchesPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
