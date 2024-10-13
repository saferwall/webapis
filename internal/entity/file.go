// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

import (
	"strings"
)

// File represent a binary file.
type File struct {
	Type             string                 `json:"type,omitempty"`
	MD5              string                 `json:"md5,omitempty"`
	SHA1             string                 `json:"sha1,omitempty"`
	SHA256           string                 `json:"sha256,omitempty"`
	SHA512           string                 `json:"sha512,omitempty"`
	SSDeep           string                 `json:"ssdeep,omitempty"`
	TLSH             string                 `json:"tlsh,omitempty"`
	Crc32            string                 `json:"crc32,omitempty"`
	Size             int64                  `json:"size,omitempty"`
	Tags             map[string]interface{} `json:"tags,omitempty"`
	Magic            string                 `json:"magic,omitempty"`
	Exif             map[string]string      `json:"exif,omitempty"`
	TriD             []string               `json:"trid,omitempty"`
	Packer           []string               `json:"packer,omitempty"`
	FirstSeen        int64                  `json:"first_seen,omitempty"`
	LastScanned      int64                  `json:"last_scanned,omitempty"`
	Submissions      []Submission           `json:"submissions,omitempty"`
	Strings          interface{}            `json:"strings,omitempty"`
	MultiAV          map[string]interface{} `json:"multiav,omitempty"`
	PE               interface{}            `json:"pe,omitempty"`
	Histogram        []int                  `json:"histogram,omitempty"`
	ByteEntropy      []int                  `json:"byte_entropy,omitempty"`
	Ml               map[string]interface{} `json:"ml,omitempty"`
	Format           string                 `json:"file_format,omitempty"`
	Extension        string                 `json:"file_extension,omitempty"`
	DefaultBhvReport interface{}            `json:"default_behavior_report,omitempty"`
	BhvScans         interface{}            `json:"behavior_scans,omitempty"`
	Status           FileScanProgressType   `json:"status,omitempty"`
}

// Submission represents a file submission.
type Submission struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Filename  string `json:"filename,omitempty"`
	Source    string `json:"src,omitempty"`
	Country   string `json:"country,omitempty"`
}

// FileScanProgressType represents the file scan progress type.
type FileScanProgressType uint8

// Progress of a file scan.
const (
	FileScanProgressQueued     FileScanProgressType = 1
	FileScanProgressProcessing FileScanProgressType = 2
	FileScanProgressFinished   FileScanProgressType = 3
)

// ID returns a unique ID to identify a File object.
func (f File) ID(key string) string {
	return strings.ToLower(key)
}
