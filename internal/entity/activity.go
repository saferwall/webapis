// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

// Activity keeps track of activities made by users.
type Activity struct {
	// Type represents the document type.
	Type string `json:"type,omitempty"`
	// ID represents the activity identifier.
	ID string `json:"id,omitempty"`
	// Kind represents the type of the activity,
	// possible values: "follow", "comment", "like", "submit".
	Kind string `json:"kind,omitempty"`
	// Timestamp when this activity happened.
	Timestamp int64 `json:"timestamp,omitempty"`
	// Username represents the user who made this activity.
	Username string `json:"username,omitempty"`
	// Target could be a sha256, username or a comment id.
	Target string `json:"target,omitempty"`
	// Source describes weather the activity was generated from
	// a real web browser or a script.
	Source string `json:"src,omitempty"`
}
