// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

import "strings"

// UserLike represents likes files by a user.
type UserLike struct {
	SHA256    string `json:"sha256"`
	Timestamp int64  `json:"ts"`
}

// User represents a user.
type User struct {
	Type             string     `json:"type"`
	Email            string     `json:"email,omitempty"`
	Username         string     `json:"username"`
	Password         string     `json:"password,omitempty"`
	FullName         string     `json:"name"`
	Location         string     `json:"location"`
	URL              string     `json:"url"`
	Bio              string     `json:"bio"`
	Confirmed        bool       `json:"confirmed"`
	MemberSince      int64      `json:"member_since"`
	LastSeen         int64      `json:"last_seen"`
	Admin            bool       `json:"admin"`
	Following        []string   `json:"following"`
	FollowingCount   int        `json:"following_count"`
	Followers        []string   `json:"followers"`
	FollowersCount   int        `json:"followers_count"`
	Likes            []UserLike `json:"likes"`
	SubmissionsCount int        `json:"submissions_count"`
	CommentsCount    int        `json:"comments_count"`
}

// UserPrivate represent a user with sensitive fields included.
type UserPrivate struct {
	User
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// ID returns a unique ID to identify a User object.
func (f User) ID() string {
	return strings.ToLower(f.Username)
}

// Name returns the user name.
func (u User) IsAdmin() bool {
	return u.Admin
}

// contextKey defines a custom time to get/set values from a context.
type contextKey int

const (
	// UserKey identifies the current user during the request life.
	UserKey contextKey = iota

	// SourceKey identifies the source of the HTTP request (web or api).
	SourceKey
)
