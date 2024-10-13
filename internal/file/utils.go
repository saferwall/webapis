// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	regPathNotation = regexp.MustCompile(`^[\w.]+$`)
)

// areFieldsAllowed check if we are allowed to filter GET with fields
func areFieldsAllowed(fields []string) bool {
	for _, field := range fields {
		if !regPathNotation.MatchString(field) {
			return false
		}
	}
	return true
}

// hash calculates the sha256 hash over a stream of bytes.
func hash(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

// isBrowser returns true when the HTTP request is coming from a known user agent.
func isBrowser(userAgent string) bool {
	browserList := []string{
		"Chrome", "Chromium", "Mozilla", "Opera", "Safari", "Edge", "MSIE",
	}

	for _, browserName := range browserList {
		if strings.Contains(userAgent, browserName) {
			return true
		}
	}
	return false
}
