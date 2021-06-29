// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

import "strings"

// File represent a sample
type File struct {
	Type            string                 `json:"type,omitempty"`
	MD5             string                 `json:"md5,omitempty"`
	SHA1            string                 `json:"sha1,omitempty"`
	SHA256          string                 `json:"sha256,omitempty"`
	SHA512          string                 `json:"sha512,omitempty"`
	Ssdeep          string                 `json:"ssdeep,omitempty"`
	CRC32           string                 `json:"crc32,omitempty"`
	Magic           string                 `json:"magic,omitempty"`
	Size            uint64                 `json:"size,omitempty"`
	Exif            map[string]string      `json:"exif,omitempty"`
	Tags            map[string]interface{} `json:"tags,omitempty"`
	TriD            []string               `json:"trid,omitempty"`
	Packer          []string               `json:"packer,omitempty"`
	FirstSeen       int64                  `json:"first_seen,omitempty"`
	LastSubmission  int64                  `json:"last_submission,omitempty"`
	LastScanned     int64                  `json:"last_scanned,omitempty"`
	FileSubmissions []FileSubmission       `json:"submissions,omitempty"`
	Strings         []interface{}          `json:"strings,omitempty"`
	MultiAV         map[string]interface{} `json:"multiav,omitempty"`
	PE              interface{}            `json:"pe,omitempty"`
	Histogram       []int                  `json:"histogram,omitempty"`
	ByteEntropy     []int                  `json:"byte_entropy,omitempty"`
	Ml              map[string]interface{} `json:"ml,omitempty"`
	CommentsCount   int                    `json:"comments_count"`
	FileType        string                 `json:"filetype,omitempty"`
	Status          int                    `json:"status,omitempty"`
}

// FileSubmission represents a file submission.
type FileSubmission struct {
	Date     int64  `json:"ts,omitempty"`
	Filename string `json:"filename,omitempty"`
	Source   string `json:"src,omitempty"`
	Country  string `json:"country,omitempty"`
}

// ID returns a unique ID to identify a File object.
func (f File) ID() string {
	return "files::" + strings.ToLower(f.SHA256)
}

// Fields returns list of allowerd fields to be retrieved.
func (f File) Fields() string {
	return "files::" + strings.ToLower(f.SHA256)
}
