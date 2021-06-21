// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

// File represent a sample
type File struct {
	Type           string                 `json:"type,omitempty"`
	MD5            string                 `json:"md5,omitempty"`
	SHA1           string                 `json:"sha1,omitempty"`
	SHA256         string                 `json:"sha256,omitempty"`
	SHA512         string                 `json:"sha512,omitempty"`
	Ssdeep         string                 `json:"ssdeep,omitempty"`
	CRC32          string                 `json:"crc32,omitempty"`
	Magic          string                 `json:"magic,omitempty"`
	Size           uint64                 `json:"size,omitempty"`
	Exif           map[string]string      `json:"exif,omitempty"`
	Tags           map[string]interface{} `json:"tags,omitempty"`
	TriD           []string               `json:"trid,omitempty"`
	Packer         []string               `json:"packer,omitempty"`
	FirstSeen      int64                  `json:"first_seen,omitempty"`
	LastSubmission int64                  `json:"last_submission,omitempty"`
	LastScanned    int64                  `json:"last_scanned,omitempty"`
	Submissions    []submission           `json:"submissions,omitempty"`
	Strings        []interface{}          `json:"strings,omitempty"`
	MultiAV        map[string]interface{} `json:"multiav,omitempty"`
	PE             interface{}            `json:"pe,omitempty"`
	Histogram      []int                  `json:"histogram,omitempty"`
	ByteEntropy    []int                  `json:"byte_entropy,omitempty"`
	Ml             map[string]interface{} `json:"ml,omitempty"`
	CommentsCount  int                    `json:"comments_count"`
	FileType       string                 `json:"filetype,omitempty"`
	Status         int                    `json:"status,omitempty"`
}

type submission struct {
	Date     int64  `json:"ts,omitempty"`
	Filename string `json:"filename,omitempty"`
	Source   string `json:"source,omitempty"`
	Country  string `json:"country,omitempty"`
}
