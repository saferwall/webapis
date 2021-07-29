// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package local

import (
	"io"
	"os"
	"path/filepath"
)

// Service provides abstraction to cloud object storage.
type Service struct {
	// Root directory in the local file system.
	root string
}

// New generates new object storage service.
func New(root string) (Service, error) {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		if err := os.MkdirAll(root, os.ModePerm); err != nil {
			return Service{}, err
		}
	}
	return Service{root}, nil
}

// Upload upload an object to s3.
func (s Service) Upload(bucket, key string, file io.Reader, timeout int) error {

	// Create new file.
	dest := filepath.Join(s.root, bucket, key)
	new, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer new.Close()

	// Perform the copy.
	if _, err := io.Copy(new, file); err != nil {
		return err
	}

	return nil
}
