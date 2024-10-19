// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"context"
	"regexp"
)

// contextKey defines a custom time to get/set values from a context.
type contextKey int

const (
	// filtersKey identifies the current filters during the request life.
	FiltersKey contextKey = iota
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

// WithFilters returns a context that contains the API filters.
func WithFilters(ctx context.Context, value map[string][]string) context.Context {
	return context.WithValue(ctx, FiltersKey, value)
}
