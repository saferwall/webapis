// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package s3

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Service provides abstraction to cloud object storage.
type Service struct {
	// s3 service.
	s3svc *awss3.S3
	// s3 uploader.
	uploader *s3manager.Uploader
	// s3 downloader.
	downloader *s3manager.Downloader
}

// FakeWriterAt represents a struct that provides the method WriteAt so it will
// satisfy interface io.WriterAt. It will ignore offset and therefore works
// like just io.Writer. AWS SDK is Using io.WriterAt because of concurrent
// download, so it can write at offset position (e.g. in middle of file).
// By disabling concurrent download we can safely ignore the offset argument
// because it will be downloaded sequentially.
type FakeWriterAt struct {
	w io.Writer
}

func (fw FakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we forced sequential downloads
	return fw.w.Write(p)
}

// New generates new s3 object storage service.
func New(region, accessKey, secretKey string) (Service, error) {

	// The session the S3 Uploader will use.
	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region),
		Credentials: creds})
	if err != nil {
		return Service{}, nil
	}

	// S3 service client the Upload manager will use.
	s3Svc := awss3.New(sess)

	// Create an uploader with S3 client and custom options
	uploader := s3manager.NewUploaderWithClient(s3Svc, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // 5MB per part
		u.LeavePartsOnError = true   // Don't delete the parts if the upload fails.
	})

	// Create a downloader with S3 client and custom options
	downloader := s3manager.NewDownloaderWithClient(s3Svc,
		func(u *s3manager.Downloader) {
			u.Concurrency = 1            // Guarantee sequential writes
			u.PartSize = 5 * 1024 * 1024 // 5MB per part
		})

	return Service{s3Svc, uploader, downloader}, nil
}

// Upload upload an object to s3.
func (s Service) Upload(ctx context.Context, bucket, key string,
	file io.Reader) error {

	// Upload input parameters
	upParams := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
	}

	// Perform an upload.
	_, err := s.uploader.UploadWithContext(ctx, upParams)

	return err
}

// Download downloads an object from s3.
func (s Service) Download(ctx context.Context, bucket, key string,
	file io.Writer) error {

	// Download input parameters.
	input := &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// Perform the download.
	_, err := s.downloader.DownloadWithContext(ctx, FakeWriterAt{file}, input)

	return err
}

// Download downloads an object from s3.
func (s Service) DownloadWithSize(ctx context.Context, bucket, key string,
	file io.Writer, done func()) (int64, error) {


	// Download input parameters.
	input := &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// Perform the download.

	obj, err := s.downloader.S3.GetObject(input)
	if err != nil {
		return 0, err
	}
	size := *obj.ContentLength
	go func () {
		defer obj.Body.Close()
		defer done()
		io.CopyN(file, obj.Body, size)
	}()
	return size, err
}

// GetFileSize gets an object's size from s3.
func (s Service) GetFileSize(ctx context.Context, bucket, key string, done func()) (int64, error) {

	// GetAttributes input parameters.
	input := &awss3.GetObjectAttributesInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ObjectAttributes: aws.StringSlice([]string{awss3.ObjectAttributesObjectSize}),
	}

	// Perform the GetAttributes request.
	obj, err := s.downloader.S3.GetObjectAttributes(input)
	if err != nil {
		return 0, err
	}
	defer done()
	return *obj.ObjectSize, err
}

// MakeBucket creates a new bucket in s2.
func (s Service) MakeBucket(ctx context.Context, bucketName, location string) error {

	input := &awss3.CreateBucketInput{Bucket: aws.String(bucketName)}

	// ignore location constraint for `us-east-1`, otherwise bucket
	// creation request fails.
	if location != "us-east-1" {
		input.CreateBucketConfiguration = &awss3.CreateBucketConfiguration{
			LocationConstraint: aws.String(location)}
	}

	_, err := s.s3svc.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case awss3.ErrCodeBucketAlreadyExists:
				return aerr
			case awss3.ErrCodeBucketAlreadyOwnedByYou:
				return nil
			default:
				return aerr
			}
		} else {
			return err
		}
	}

	return nil
}

func (s Service) Exists(ctx context.Context, bucketName,
	key string) (bool, error) {

	_, err := s.s3svc.HeadObjectWithContext(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			// s3.ErrCodeNoSuchKey does not work, aws is missing this error code
			// so we hardwire a string
			case "NotFound":
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

func (s Service) GeneratePresignedURL(ctx context.Context, bucketName, key string) (string, error) {
	req, _ := s.s3svc.GetObjectRequest(&awss3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(5 * time.Minute)
	if err != nil {
		return "", err
	}

	return urlStr, nil
}

// Delete removes an object from the store.
func (s Service) Delete(ctx context.Context, bucket, key string) (bool, error) {

	// Prepare the delete object input.
	input := &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.s3svc.DeleteObjectWithContext(ctx, input)
	if err != nil {
		return false, err
	}

	return true, nil
}
