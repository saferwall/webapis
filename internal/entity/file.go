// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

import "strings"

// File represent a sample
type File struct {
	Type          string                 `json:"type,omitempty"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
	Tags          map[string]interface{} `json:"tags,omitempty"`
	FirstSeen     int64                  `json:"first_seen,omitempty"`
	LastScanned   int64                  `json:"last_scanned,omitempty"`
	Submissions   []Submission           `json:"submissions,omitempty"`
	Strings       []interface{}          `json:"strings,omitempty"`
	MultiAV       map[string]interface{} `json:"multiav,omitempty"`
	PE            interface{}            `json:"pe,omitempty"`
	Ml            map[string]interface{} `json:"ml,omitempty"`
	CommentsCount int                    `json:"comments_count"`
	Format        string                 `json:"format,omitempty"`
	Status        int                    `json:"status,omitempty"`
}

// Submission represents a file submission.
type Submission struct {
	Date     int64  `json:"ts,omitempty"`
	Filename string `json:"filename,omitempty"`
	Source   string `json:"src,omitempty"`
	Country  string `json:"country,omitempty"`
}

// ID returns a unique ID to identify a File object.
func (f File) ID(key string) string {
	return "files::" + strings.ToLower(key)
}
