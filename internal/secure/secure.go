// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package secure

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"strconv"
	"time"

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

// Hash hashes the password using bcrypt.
func (*Service) Hash(password string) string {
	hashedPW, _ := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	return string(hashedPW)
}

// HashMatchesPassword matches hash with password. Returns true if hash and
// password match.
func (*Service) HashMatchesPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Token generates new unique token.
func (s *Service) Token(str string) string {
	s.h.Reset()
	fmt.Fprintf(s.h, "%s%s", str, strconv.Itoa(time.Now().Nanosecond()))
	return fmt.Sprintf("%x", s.h.Sum(nil))
}

// HashFile hashes the password using bcrypt.
func (*Service) HashFile(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
