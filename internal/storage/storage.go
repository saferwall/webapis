// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/saferwall/saferwall-api/internal/config"
	"github.com/saferwall/saferwall-api/internal/storage/local"
	"github.com/saferwall/saferwall-api/internal/storage/minio"
	"github.com/saferwall/saferwall-api/internal/storage/s3"
)

var (
	errDeploymentNotFound = errors.New("deployment not found")
	timeout               = time.Duration(time.Second * 5)
)

// UploadDownloader abstract uploading and download files from different
// object storage solutions.
type UploadDownloader interface {
	// Upload uploads a file to an object storage.
	Upload(ctx context.Context, bucket, key string, file io.Reader) error
	// Download downloads a file from a remote object storage location.
	Download(ctx context.Context, bucket, key string, file io.Writer) error
	// DownloadWithSize downloads a file from a remote object storage location and returns it's size.
	DownloadWithSize(ctx context.Context, bucket, key string, file io.Writer) (int64, error)
	// MakeBucket creates a new bucket.
	MakeBucket(ctx context.Context, bucket, location string) error
	// Exists checks whether an object exists.
	Exists(ctx context.Context, bucket, key string) (bool, error)
	// GeneratePresignedURL generates a pre-signed URL for downloading samples.
	GeneratePresignedURL(ctx context.Context, bucket, key string)(string, error)
}

func New(cfg config.StorageCfg) (UploadDownloader, error) {

	// Create a context with a timeout that will abort the upload if it takes
	// more than the passed in timeout.
	ctx := context.Background()
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}

	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	if cancelFn != nil {
		defer cancelFn()
	}

	switch cfg.DeploymentKind {
	case "aws":
		svc, err := s3.New(cfg.S3.Region, cfg.S3.AccessKey, cfg.S3.SecretKey)
		if err != nil {
			return nil, err
		}
		err = svc.MakeBucket(ctx, cfg.FileContainerName, cfg.S3.Region)
		if err != nil {
			return nil, err
		}
		err = svc.MakeBucket(ctx, cfg.AvatarsContainerName, cfg.S3.Region)
		if err != nil {
			return nil, err
		}
		return svc, nil

	case "minio":
		svc, err := minio.New(cfg.Minio.Endpoint, cfg.Minio.AccessKey,
			cfg.Minio.SecretKey)
		if err != nil {
			return nil, err
		}
		err = svc.MakeBucket(ctx, cfg.FileContainerName, cfg.Minio.Region)
		if err != nil {
			return nil, err
		}
		err = svc.MakeBucket(ctx, cfg.AvatarsContainerName, cfg.Minio.Region)
		if err != nil {
			return nil, err
		}
		return svc, nil
	case "local":
		svc, err := local.New(cfg.Local.RootDir)
		if err != nil {
			return nil, err
		}
		err = svc.MakeBucket(ctx, cfg.FileContainerName, "")
		if err != nil {
			return nil, err
		}
		err = svc.MakeBucket(ctx, cfg.AvatarsContainerName, "")
		if err != nil {
			return nil, err
		}
		return svc, nil
	}

	return nil, errDeploymentNotFound
}
