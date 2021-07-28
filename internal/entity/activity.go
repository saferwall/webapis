// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

// Activity keeps track of activities made by users.
type Activity struct {
	// Type represents the type of the activity,
	// possible values: "follow", "comment", "like", "Submit".
	Activity string `json:"activity,omitempty"`
	// Timestamp when this activity happened.
	Timestamp int64 `json:"timestamp,omitempty"`
	// Username represents the user who made this activity.
	Username string `json:"username,omitempty"`
	// Target could be a sha256 or a username.
	Target string `json:"target,omitempty"`
	// Type represents the document type.
	Type string `json:"type,omitempty"`
}
