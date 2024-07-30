// Copyright 2022 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
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

// isStringInSlice check if a string exist in a list of strings
func isStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

// removeStringFromSlice removes a string item from a list of strings.
func removeStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// hash calculates the sha256 hash over a stream of bytes.
func hash(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
