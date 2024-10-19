// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

type Comment struct {
	// Meta represents document metadata.
	Meta *DocMetadata `json:"doc,omitempty"`
	// Type represents the document type.
	Type string `json:"type,omitempty"`
	// ID represents the activity identifier.
	ID string `json:"id,omitempty"`
	// Body represents the content of the comment.
	Body string `json:"body,omitempty"`
	// SHA256 references the hash of the file
	// where the comment has been made.
	SHA256 string `json:"sha256,omitempty"`
	// Timestamp when this activity happened.
	Timestamp int64 `json:"timestamp,omitempty"`
	// Username represents the author of the comment.
	Username string `json:"username,omitempty"`
}
