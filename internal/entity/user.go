// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

import "strings"

// User represent a user.
type User struct {
	Type             string   `json:"type"`
	Email            string   `json:"email,omitempty"`
	Username         string   `json:"username"`
	Password         string   `json:"password,omitempty"`
	FullName         string   `json:"name"`
	Location         string   `json:"location"`
	URL              string   `json:"url"`
	Bio              string   `json:"bio"`
	Confirmed        bool     `json:"confirmed"`
	MemberSince      int64    `json:"member_since"`
	LastSeen         int64    `json:"last_seen"`
	Admin            bool     `json:"admin"`
	HasAvatar        bool     `json:"has_avatar"`
	Following        []string `json:"following"`
	FollowingCount   int      `json:"following_count"`
	Followers        []string `json:"followers"`
	FollowersCount   int      `json:"followers_count"`
	Likes            []string `json:"likes"`
	LikesCount       int      `json:"likes_count"`
	SubmissionsCount int      `json:"submissions_count"`
	CommentsCount    int      `json:"comments_count"`
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
)
