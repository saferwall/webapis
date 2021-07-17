// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

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
	Username  string `json:"username,omitempty"`
	Timestamp int64  `json:"ts,omitempty"`
}

// Submissions represents a User subsmission: either a file or an URL.
type UserSubmission struct {
	Type      string `json:"type,omitempty"`
	Hash      string `json:"hash,omitempty"`
	Username  string `json:"username,omitempty"`
	Timestamp int64  `json:"ts,omitempty"`
}