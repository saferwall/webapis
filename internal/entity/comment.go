// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

type Comment struct {
	Body      string `json:"body"`
	SHA256    string `json:"sha256"`
	Timestamp int64  `json:"timestamp"`
	Username  string `json:"username"`
}
