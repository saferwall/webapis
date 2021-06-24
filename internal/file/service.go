// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"time"

	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

const (
	// ErrDocumentNotFound is returned when the doc does not exist in the DB.
	ErrDocumentNotFound = "document not found"
)

// Service encapsulates usecase logic for files.
type Service interface {
	Get(ctx context.Context, id string) (File, error)
	//Query(ctx context.Context, offset, limit int) ([]File, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateFileRequest) (File, error)
	Update(ctx context.Context, id string, input UpdateFileRequest) (File, error)
	Delete(ctx context.Context, id string) (File, error)
}

// File represents the data about a File.
type File struct {
	entity.File
}

// Securer represents security interface.
type Securer interface {
	HashFile([]byte) string
}

type Uploader interface {
	Upload(bucket, key string, file io.Reader, timeout int) error
}

// CreateFileRequest represents a file creation request.
type CreateFileRequest struct {
	src io.Reader
}

type service struct {
	sec    Securer
	repo   Repository
	logger log.Logger
	upl    Uploader
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
	upl Uploader) Service {
	return service{sec, repo, logger, upl}
}

// Get returns the File with the specified File ID.
func (s service) Get(ctx context.Context, id string) (File, error) {
	file, err := s.repo.Get(ctx, id)
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
	file, err := s.Get(ctx, sha256)
	if err != nil && err.Error() != ErrDocumentNotFound {
		return File{}, err
	}

	// Is this a new file ? if yes, we create a new doc in the db.
	if err.Error() == ErrDocumentNotFound {
		err = s.upl.Upload("files", sha256, req.src, 10)
		if err != nil {
			return File{}, err
		}
		err = s.repo.Create(ctx, entity.File{
			SHA256: sha256,
			Type:   "file",
		})
		if err != nil {
			return File{}, err
		}
	}

	// If not, we append this new submission to the file doc.
	file.LastScanned = time.Now().Unix()
	return s.Get(ctx, sha256)
}

// Update updates the File with the specified ID.
func (s service) Update(ctx context.Context, id string, req UpdateFileRequest) (
	File, error) {

	File, err := s.Get(ctx, id)
	if err != nil {
		return File, err
	}

	// merge the structures.
	data, err := json.Marshal(req)
	if err != nil {
		return File, err
	}
	err = json.Unmarshal(data, &File)
	if err != nil {
		return File, err
	}

	// check if File.Username == id
	if err := s.repo.Update(ctx, File.File); err != nil {
		return File, err
	}

	return File, nil
}

// Delete deletes the File with the specified ID.
func (s service) Delete(ctx context.Context, id string) (File, error) {
	file, err := s.Get(ctx, id)
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
