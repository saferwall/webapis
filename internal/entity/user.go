// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

import "strings"

// User represent a user.
type User struct {
	Email            string `json:"email,omitempty"`
	Username         string `json:"username,omitempty"`
	Password         string `json:"password,omitempty"`
	FullName         string `json:"name,omitempty"`
	Location         string `json:"location,omitempty"`
	URL              string `json:"url,omitempty"`
	Bio              string `json:"bio,omitempty"`
	Confirmed        bool   `json:"confirmed,omitempty"`
	MemberSince      int64  `json:"member_since,omitempty"`
	LastSeen         int64  `json:"last_seen,omitempty"`
	Admin            bool   `json:"admin,omitempty"`
	HasAvatar        bool   `json:"has_avatar,omitempty"`
	FollowingCount   int    `json:"following_count"`
	FollowersCount   int    `json:"followers_count"`
	LikesCount       int    `json:"likes_count"`
	SubmissionsCount int    `json:"submissions_count"`
	CommentsCount    int    `json:"comments_count"`
}

// ID returns a unique ID to identify a User object.
func (f User) ID() string {
	return "users::" + strings.ToLower(f.Username)
}

// Name returns the user name.
func (u User) Name() string {
	return u.Username
}
