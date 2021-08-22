// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package secure_test

import (
	"crypto/sha1"
	"testing"

	"github.com/saferwall/saferwall-api/internal/secure"
	"github.com/stretchr/testify/assert"
)


func TestHashAndMatch(t *testing.T) {
	cases := []struct {
		name string
		pass string
		want bool
	}{
		{
			name: "Success",
			pass: "gamepad",
			want: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := secure.New(nil)
			hash := s.HashPW(tt.pass)
			assert.Equal(t, tt.want, s.HashMatchesPassword(hash, tt.pass))
		})
	}
}

func TestToken(t *testing.T) {
	s := secure.New(sha1.New())
	token := "token"
	tokenized := s.Token(token)
	assert.NotEqual(t, tokenized, token)
}
