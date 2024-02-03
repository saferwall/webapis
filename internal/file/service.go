// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"time"

	"github.com/saferwall/saferwall-api/internal/activity"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/user"
	"github.com/saferwall/saferwall-api/pkg/log"
)

var (
	// ErrDocumentNotFound is returned when the doc does not exist in the DB.
	ErrDocumentNotFound = "document not found"
	// ErrObjectNotFound is returned when an object does not exist in Obj storage.
	ErrObjectNotFound = errors.New("object not found")
	// file upload timeout in seconds.
	fileUploadTimeout = time.Duration(time.Second * 30)
)

// Progress of a file scan.
const (
	queued = iota + 1
	processing
	finished
)

// Service encapsulates use case logic for files.
type Service interface {
	Get(ctx context.Context, id string, fields []string) (File, error)
	Count(ctx context.Context) (int, error)
	Exists(ctx context.Context, id string) (bool, error)
	Create(ctx context.Context, input CreateFileRequest) (File, error)
	Update(ctx context.Context, id string, input UpdateFileRequest) (File, error)
	Delete(ctx context.Context, id string) (File, error)
	Query(ctx context.Context, offset, limit int, fields []string) ([]File, error)
	Patch(ctx context.Context, key, path string, val interface{}) error
	Summary(ctx context.Context, id string) (interface{}, error)
	Like(ctx context.Context, id string) error
	Unlike(ctx context.Context, id string) error
	Rescan(ctx context.Context, id string) error
	Comments(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	CountStrings(ctx context.Context, id string) (int, error)
	Strings(ctx context.Context, id string, offset, limit int) (interface{}, error)
	Download(ctx context.Context, id string, zipFile *string) error
	GeneratePresignedURL(ctx context.Context, id string) (string, error)
}

// File represents the data about a File.
type File struct {
	entity.File
}

// FileScanCfg represents a file scanning config. This map to a 1:1 mapping between
// the config stored in the main saferwall repo.
type FileScanCfg struct {
	// SHA256 hash of the file.
	SHA256 string `json:"sha256,omitempty"`
	// Config used during dynamic file scan.
	DynFileScanCfg `json:"dynamic,omitempty"`
}

// DynFileScanCfg represents the config used to detonate a file.
type DynFileScanCfg struct {
	// Destination path where the sample will be located in the VM.
	DestPath string `json:"dest_path,omitempty"`
	// Arguments used to run the sample.
	Arguments string `json:"args,omitempty"`
	// Timeout in seconds for how long to keep the VM running.
	Timeout int `json:"timeout,omitempty"`
	// Country to route traffic through.
	Country string `json:"country,omitempty"`
	// Operating System used to run the sample.
	OS string `json:"os,omitempty"`
}

type UploadDownloader interface {
	Upload(ctx context.Context, bucket, key string, file io.Reader) error
	Download(ctx context.Context, bucket, key string, file io.Writer) error
	Exists(ctx context.Context, bucket, key string) (bool, error)
	GeneratePresignedURL(ctx context.Context, bucket, key string) (string, error)
}

// Producer represents event stream message producer interface.
type Producer interface {
	Produce(string, []byte) error
}

// Archiver represents the archiving interface for files.
type Archiver interface {
	Archive(string, string, io.Reader) error
}

// CreateFileRequest represents a file creation request.
type CreateFileRequest struct {
	src      io.Reader
	filename string
	geoip    string
}

type service struct {
	repo          Repository
	logger        log.Logger
	objSto        UploadDownloader
	producer      Producer
	topic         string
	bucket        string
	samplesZipPwd string
	userSvc       user.Service
	actSvc        activity.Service
	archiver      Archiver
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
func NewService(repo Repository, logger log.Logger,
	updown UploadDownloader, producer Producer, topic, bucket, samplesZipPwd string,
	userSvc user.Service, actSvc activity.Service, arch Archiver) Service {
	return service{repo, logger, updown, producer, topic, bucket, samplesZipPwd,
		userSvc, actSvc, arch}
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

	fileContent, err := io.ReadAll(req.src)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return File{}, err
	}

	sha256 := hash(fileContent)
	file, err := s.Get(ctx, sha256, nil)
	if err != nil && err.Error() != ErrDocumentNotFound {
		return File{}, err
	}

	now := time.Now().Unix()

	// When a new file has been uploaded, we create a new doc in the db.
	if err != nil && err.Error() == ErrDocumentNotFound {

		// Create a context with a timeout that will abort the upload if it takes
		// more than the passed in timeout.
		uploadCtx, cancelFn := context.WithTimeout(context.Background(), fileUploadTimeout)

		// Ensure the context is canceled to prevent leaking.
		defer cancelFn()

		// TODO: check if the file is already in Obj.
		err = s.objSto.Upload(uploadCtx, s.bucket, sha256,
			bytes.NewReader(fileContent))
		if err != nil {
			return File{}, err
		}

		// Get the source of the HTTP request from the ctx.
		source, _ := ctx.Value(entity.SourceKey).(string)

		// Create a new submission.
		submission := entity.Submission{
			Timestamp: now,
			Filename:  req.filename,
			Source:    source,
			Country:   req.geoip,
		}

		// Create a new file.
		err = s.repo.Create(ctx, sha256, entity.File{
			SHA256:      sha256,
			Type:        "file",
			FirstSeen:   now,
			Submissions: append(file.Submissions, submission),
			Status:      queued,
		})
		if err != nil {
			s.logger.With(ctx).Error(err)
			return File{}, err
		}

		loggedInUser, _ := ctx.Value(entity.UserKey).(entity.User)
		user, err := s.userSvc.Get(ctx, loggedInUser.ID())
		if err != nil {
			return File{}, err
		}

		// Create a new `submit` activity.
		if _, err = s.actSvc.Create(ctx, activity.CreateActivityRequest{
			Kind:     "submit",
			Username: user.Username,
			Target:   sha256,
			Source:   source,
		}); err != nil {
			return File{}, err
		}

		// Update submissions count on user object.
		err = s.Patch(ctx, user.ID(), "submissions_count", user.SubmissionsCount+1)
		if err != nil {
			return File{}, err
		}

		// Serialize the msg to send to the orchestrator.
		msg, err := json.Marshal(FileScanCfg{SHA256: sha256})
		if err != nil {
			s.logger.With(ctx).Error(err)
			return File{}, err
		}

		// Push a message to the queue to scan this file.
		err = s.producer.Produce(s.topic, msg)
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

// Count returns the number of files.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Exists checks if a document exists for the given id.
func (s service) Exists(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
}

// Query returns the files with the specified offset and limit.
func (s service) Query(ctx context.Context, offset, limit int, fields []string) (
	[]File, error) {

	items, err := s.repo.Query(ctx, offset, limit, fields)
	if err != nil {
		return nil, err
	}
	result := []File{}
	for _, item := range items {
		result = append(result, File{item})
	}
	return result, nil
}

// Patch performs an atomic file sub document update.
func (s service) Patch(ctx context.Context, id, path string,
	input interface{}) error {
	return s.repo.Patch(ctx, id, path, input)
}

// Summary returns a summary of a file scan.
func (s service) Summary(ctx context.Context, id string) (interface{}, error) {
	res, err := s.repo.Summary(ctx, id)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s service) Like(ctx context.Context, sha256 string) error {

	loggedInUser, _ := ctx.Value(entity.UserKey).(entity.User)
	user, err := s.userSvc.Get(ctx, loggedInUser.ID())
	if err != nil {
		return err
	}

	// Get the source of the HTTP request from the ctx.
	source, _ := ctx.Value(entity.SourceKey).(string)

	if !isStringInSlice(sha256, user.Likes) {
		user.Likes = append(user.Likes, sha256)
		user.LikesCount += 1
		user, err = s.userSvc.Update(ctx, user.ID(), user)
		if err != nil {
			return err
		}

		// add new `like` activity.
		if _, err = s.actSvc.Create(ctx, activity.CreateActivityRequest{
			Kind:     "like",
			Username: user.Username,
			Target:   sha256,
			Source:   source,
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

	if isStringInSlice(sha256, user.Likes) {
		user.Likes = removeStringFromSlice(user.Likes, sha256)
		user.LikesCount -= 1
		user, err = s.userSvc.Update(ctx, user.ID(), user)
		if err != nil {
			return err
		}

		// delete corresponding activity.
		if s.actSvc.DeleteWith(ctx, "like", user.ID(),
			sha256); err != nil {
			return err
		}
	}

	return nil
}

func (s service) Rescan(ctx context.Context, sha256 string) error {

	// Serialize the msg to send to the orchestrator.
	msg, err := json.Marshal(FileScanCfg{SHA256: sha256})
	if err != nil {
		s.logger.With(ctx).Error(err)
		return err
	}

	// Push a message to the queue to scan this file.
	err = s.producer.Produce(s.topic, msg)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return err
	}

	return nil
}

func (s service) Download(ctx context.Context, sha256 string, zipFile *string) error {

	// Create a context with a timeout that will abort the download if it takes
	// more than the passed in timeout.
	downloadCtx, cancelFn := context.WithTimeout(
		context.Background(), time.Duration(time.Second*30))
	defer cancelFn()

	found, err := s.objSto.Exists(ctx, s.bucket, sha256)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return err
	}

	if !found {
		return ErrObjectNotFound
	}

	buf := new(bytes.Buffer)
	err = s.objSto.Download(downloadCtx, s.bucket, sha256, buf)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return err
	}

	*zipFile = filepath.Join("/tmp", sha256+".zip")
	err = s.archiver.Archive(*zipFile, s.samplesZipPwd, buf)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return err
	}
	return nil
}

func (s service) Comments(ctx context.Context, id string, offset, limit int) (
	[]interface{}, error) {

	result, err := s.repo.Comments(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) CountStrings(ctx context.Context, id string) (int, error) {
	count, err := s.repo.CountStrings(ctx, id)
	if err != nil {
		return 0, err
	}
	return count, err
}

func (s service) Strings(ctx context.Context, id string, offset, limit int) (
	interface{}, error) {

	result, err := s.repo.Strings(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) GeneratePresignedURL(ctx context.Context, id string) (string, error) {

	found, err := s.objSto.Exists(ctx, s.bucket, id)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return "", err
	}

	if !found {
		return "", ErrObjectNotFound
	}

	return s.objSto.GeneratePresignedURL(ctx, s.bucket, id)
}
