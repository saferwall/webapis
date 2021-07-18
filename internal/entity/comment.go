// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

type Comment struct {
	Date     int64  `json:"ts,omitempty"`
	Body     string `json:"body,omitempty"`
	Username string `json:"username,omitempty"`
}
