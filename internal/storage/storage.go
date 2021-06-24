// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package storage

import (
	"github.com/graymeta/stow"
	// support local storage
	_ "github.com/graymeta/stow/local"
	// support s3 storage
	_ "github.com/graymeta/stow/s3"
)

// Dial dials stow storage.
// See stow.Dial for more information.
func Dial(kind string, config stow.Config) (stow.Location, error) {
	return stow.Dial(kind, config)
}
