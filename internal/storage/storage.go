// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package storage

import (
	"errors"
	"io"

	"github.com/saferwall/saferwall-api/internal/config"
	"github.com/saferwall/saferwall-api/internal/storage/local"
	"github.com/saferwall/saferwall-api/internal/storage/s3"
)

var (
	errDeploymentNotFound = errors.New("deployment not found")
)

// Uploader abstract uploading files to different cloud locations.
type Uploader interface {
	// Upload uploads a file to an object storage.
	Upload(bucket, key string, file io.Reader, timeout int) error
}

func New(cfg config.StorageCfg) (Uploader, error) {

	switch cfg.DeploymentKind {
	case "aws":
		return s3.New(cfg.S3.Region, cfg.S3.AccessKey, cfg.S3.SecretKey)
	case "local":
		return local.New(cfg.Local.RootDir)
	}

	return nil, errDeploymentNotFound
}
