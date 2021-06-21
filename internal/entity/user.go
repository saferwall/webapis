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
	Name             string `json:"name,omitempty"`
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

// Follows keeps track of User followers and following.
type Follows struct {
	Type      string `json:"type,omitempty"`
	Source    string `json:"src,omitempty"`
	Target    string `json:"target,omitempty"`
	Timestamp int64  `json:"ts,omitempty"`
}

// Like represents User likes: either a file or an URL.
type Like struct {
	Type      string `json:"type,omitempty"`
	Hash      string `json:"hash,omitempty"`
	Timestamp int64  `json:"ts,omitempty"`
	Username  string `json:"username,omitempty"`
}

// Submissions represents a User subsmission: either a file or an URL.
type Submission struct {
	Type      string `json:"type,omitempty"`
	Hash      string `json:"hash,omitempty"`
	Username  string `json:"username,omitempty"`
	Timestamp int64  `json:"ts,omitempty"`
}

// GetID returns the user ID.
func (u User) GetID() string {
	return strings.ToLower(u.Username)
}

// GetName returns the user name.
func (u User) GetName() string {
	return u.Name
}
