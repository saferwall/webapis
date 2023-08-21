// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package behavior

import (
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
