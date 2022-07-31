// Copyright 2022 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

import "github.com/google/uuid"

// ID returns a unique ID to identify a document.
func ID() string {
	id := uuid.New()
	return id.String()
}

// IsValidID returns true when ID is valid.
func IsValidID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
