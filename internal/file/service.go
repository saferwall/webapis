// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"time"

	"github.com/saferwall/saferwall-api/internal/activity"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/user"
	"github.com/saferwall/saferwall-api/pkg/log"
)

var (
	// ErrDocumentNotFound is returned when the doc does not exist in the DB.
	ErrDocumentNotFound = "document not found"
	// file upload timeout in seconds.
	fileUploadTimeout = time.Duration(time.Second * 30)
)

// Service encapsulates usecase logic for files.
type Service interface {
	Get(ctx context.Context, id string, fields []string) (File, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateFileRequest) (File, error)
	Update(ctx context.Context, id string, input UpdateFileRequest) (File, error)
	Delete(ctx context.Context, id string) (File, error)
	Query(ctx context.Context, offset, limit int) ([]File, error)
	Like(ctx context.Context, id string) error
	Unlike(ctx context.Context, id string) error
}

// File represents the data about a File.
type File struct {
	entity.File
}

// Securer represents security interface.
type Securer interface {
	HashFile([]byte) string
}

type UploadDownloader interface {
	Upload(ctx context.Context, bucket, key string, file io.Reader) error
	Download(ctx context.Context, bucket, key string, file io.Writer) error
}

// Producer represents event stream message producer interface.
type Producer interface {
	Produce(string, []byte) error
}

// CreateFileRequest represents a file creation request.
type CreateFileRequest struct {
	src      io.Reader
	filename string
	geoip    string
}

type service struct {
	sec      Securer
	repo     Repository
	logger   log.Logger
	objsto   UploadDownloader
	producer Producer
	topic    string
	bucket   string
	userSvc  user.Service
	actSvc   activity.Service
}

// UpdateUserRequest represents a File update request.
type UpdateFileRequest struct {
	MD5         string                 `json:"md5,omitempty"`
	SHA1        string                 `json:"sha1,omitempty"`
	SHA256      string                 `json:"sha256,omitempty"`
	SHA512      string                 `json:"sha512,omitempty"`
	Ssdeep      string                 `json:"ssdeep,omitempty"`
	CRC32       string                 `json:"crc32,omitempty"`
	Magic       string                 `json:"magic,omitempty"`
	Size        uint64                 `json:"size,omitempty"`
	Exif        map[string]string      `json:"exif,omitempty"`
	Tags        map[string]interface{} `json:"tags,omitempty"`
	TriD        []string               `json:"trid,omitempty"`
	Packer      []string               `json:"packer,omitempty"`
	Strings     []interface{}          `json:"strings,omitempty"`
	MultiAV     map[string]interface{} `json:"multiav,omitempty"`
	PE          interface{}            `json:"pe,omitempty"`
	Histogram   []int                  `json:"histogram,omitempty"`
	ByteEntropy []int                  `json:"byte_entropy,omitempty"`
	Ml          map[string]interface{} `json:"ml,omitempty"`
	FileType    string                 `json:"filetype,omitempty"`
}

// NewService creates a new File service.
func NewService(repo Repository, logger log.Logger, sec Securer,
	updown UploadDownloader, producer Producer, topic, bucket string,
	userSvc user.Service, actSvc activity.Service) Service {
	return service{sec, repo, logger, updown,
		producer, topic, bucket, userSvc, actSvc}
}

// Get returns the File with the specified File ID.
func (s service) Get(ctx context.Context, id string, fields []string) (File, error) {
	file, err := s.repo.Get(ctx, id, fields)
	if err != nil {
		return File{}, err
	}
	return File{file}, nil
}

// Create creates a new File.
func (s service) Create(ctx context.Context, req CreateFileRequest) (
	File, error) {

	fileContent, err := ioutil.ReadAll(req.src)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return File{}, err
	}

	sha256 := s.sec.HashFile(fileContent)
	file, err := s.Get(ctx, sha256, nil)
	if err != nil && err.Error() != ErrDocumentNotFound {
		return File{}, err
	}

	now := time.Now().Unix()

	// When a new file has been uploader, we create a new doc in the db.
	if err != nil && err.Error() == ErrDocumentNotFound {

		// Create a context with a timeout that will abort the upload if it takes
		// more than the passed in timeout.
		uploadCtx, cancelFn := context.WithTimeout(context.Background(), fileUploadTimeout)

		// Ensure the context is canceled to prevent leaking.
		// See context package for more information, https://golang.org/pkg/context/
		defer cancelFn()

		err = s.objsto.Upload(uploadCtx, s.bucket, sha256,
			bytes.NewReader(fileContent))
		if err != nil {
			return File{}, err
		}

		// Create a new submission.
		submission := entity.Submission{
			Timestamp: now,
			Filename:  req.filename,
			Source:    "web",
			Country:   req.geoip,
		}

		err = s.repo.Create(ctx, sha256, entity.File{
			Type:        "file",
			FirstSeen:   now,
			LastScanned: now,
			Submissions: append(file.Submissions, submission),
		})
		if err != nil {
			s.logger.With(ctx).Error(err)
			return File{}, err
		}

		// Push a message to the queue to scan this file.
		err = s.producer.Produce(s.topic, []byte(sha256))
		if err != nil {
			s.logger.With(ctx).Error(err)
			return File{}, err
		}

		return s.Get(ctx, sha256, nil)

	} else {
		// If not, we append this new submission to the file doc.
		file.LastScanned = now
		return file, nil
	}
}

// Update updates the File with the specified ID.
func (s service) Update(ctx context.Context, id string, req UpdateFileRequest) (
	File, error) {

	file, err := s.Get(ctx, id, nil)
	if err != nil {
		return file, err
	}

	// merge the structures.
	data, err := json.Marshal(req)
	if err != nil {
		return file, err
	}
	err = json.Unmarshal(data, &file)
	if err != nil {
		return file, err
	}

	// check if File.Username == id
	if err := s.repo.Update(ctx, id, file.File); err != nil {
		return file, err
	}

	return file, nil
}

// Delete deletes the File with the specified ID.
func (s service) Delete(ctx context.Context, id string) (File, error) {
	file, err := s.Get(ctx, id, nil)
	if err != nil {
		return File{}, err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		return File{}, err
	}
	return file, nil
}

// Count returns the number of users.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Query returns the files with the specified offset and limit.
func (s service) Query(ctx context.Context, offset, limit int) (
	[]File, error) {

	items, err := s.repo.Query(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []File{}
	for _, item := range items {
		result = append(result, File{item})
	}
	return result, nil
}

func (s service) Like(ctx context.Context, sha256 string) error {

	loggedInUser, _ := ctx.Value(entity.UserKey).(entity.User)
	user, err := s.userSvc.Get(ctx, loggedInUser.ID())
	if err != nil {
		return err
	}
	_, err = s.Get(ctx, sha256, nil)
	if err != nil {
		return err
	}

	if !isStringInSlice(sha256, user.Likes) {
		user.Likes = append(user.Likes, sha256)
		user.LikesCount += 1
		user, err = s.userSvc.Update(ctx, user.ID(), user)
		if err != nil {
			return err
		}

		// add new activity
		if _, err = s.actSvc.Create(ctx, activity.CreateActivityRequest{
			Kind:     "like",
			Username: user.ID(),
			Target:   sha256,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s service) Unlike(ctx context.Context, sha256 string) error {

	loggedInUser, _ := ctx.Value(entity.UserKey).(entity.User)
	user, err := s.userSvc.Get(ctx, loggedInUser.ID())
	if err != nil {
		return err
	}
	_, err = s.Get(ctx, sha256, nil)
	if err != nil {
		return err
	}

	if isStringInSlice(sha256, user.Likes) {
		user.Likes = removeStringFromSlice(user.Likes, sha256)
		user.LikesCount -= 1
		user, err = s.userSvc.Update(ctx, user.ID(), user)
		if err != nil {
			return err
		}
	}

	return nil
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
