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
	"github.com/saferwall/saferwall-api/internal/comment"
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
	ReScan(ctx context.Context, id string, input FileScanRequest) error
	Comments(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	CountStrings(ctx context.Context, id string, queryString string) (int, error)
	CountComments(ctx context.Context, id string) (int, error)
	Strings(ctx context.Context, id string, queryString string, offset, limit int) (interface{}, error)
	Download(ctx context.Context, id string, zipFile *string) error
	DownloadRaw(ctx context.Context, id string, file io.Writer) (int64, error)
	GeneratePresignedURL(ctx context.Context, id string) (string, error)
	MetaUI(ctx context.Context, id string) (interface{}, error)
	Search(ctx context.Context, input FileSearchRequest) (FileSearchResponse, error)
}

type UploadDownloader interface {
	Upload(ctx context.Context, bucket, key string, file io.Reader) error
	Download(ctx context.Context, bucket, key string, file io.Writer) error
	DownloadWithSize(ctx context.Context, bucket, key string, file io.Writer) (int64, error)
	Exists(ctx context.Context, bucket, key string) (bool, error)
	GeneratePresignedURL(ctx context.Context, bucket, key string) (string, error)
}

// File represents the data about a File.
type File struct {
	entity.File
}

// Producer represents event stream message producer interface.
type Producer interface {
	Produce(string, []byte) error
}

// Archiver represents the archiving interface for files.
type Archiver interface {
	Archive(string, string, io.Reader) error
}

// DynFileScanCfg represents the config used to detonate a file.
type DynFileScanCfg struct {
	// Destination path where the sample will be located in the VM.
	DestPath string `json:"dest_path,omitempty" form:"dest_path"`
	// Arguments used to run the sample.
	Arguments string `json:"args,omitempty" form:"args"`
	// Timeout in seconds for how long to keep the VM running.
	Timeout int `json:"timeout,omitempty" form:"timeout"`
	// Country to route traffic through.
	Country string `json:"country,omitempty" form:"country"`
	// Operating System used to run the sample.
	OS string `json:"os,omitempty" form:"os"`
}

// FileScanRequest represents a File scan request.
type FileScanRequest struct {
	// Disable Sandbox
	SkipDetonation bool `json:"skip_detonation,omitempty" form:"skip_detonation"`
	// Dynamic scan config
	DynFileScanCfg `json:"scan_cfg,omitempty"`
}

// FileScanCfg represents a file scanning config. This map to a 1:1 mapping between
// the config stored in the main saferwall repo.
type FileScanCfg struct {
	// SHA256 hash of the file.
	SHA256 string `json:"sha256,omitempty"`
	// Represents the dynamic scan configuration plus some option fields.
	FileScanRequest
}

// CreateFileRequest represents a file creation request.
type CreateFileRequest struct {
	src      io.Reader
	filename string
	geoip    string
	scanCfg  FileScanRequest
}

// UpdateUserRequest represents a File update request.
type UpdateFileRequest struct {
	MD5         string                 `json:"md5,omitempty"`
	SHA1        string                 `json:"sha1,omitempty"`
	SHA256      string                 `json:"sha256,omitempty"`
	SHA512      string                 `json:"sha512,omitempty"`
	Ssdeep      string                 `json:"ssdeep,omitempty"`
	TLSH        string                 `json:"tlsh,omitempty"`
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

// FileSearchRequest represents a file search request.
type FileSearchRequest struct {
	Query   string `json:"query" validate:"required,min=3" example:"type=pe and tag=upx"`
	Page    int    `json:"page" validate:"omitempty,gte=0,lte=10000" example:"1"`
	PerPage int    `json:"per_page" validate:"omitempty,gte=0,lte=1000" example:"100"`
	SortBy  string `json:"sort_by" validate:"omitempty,printascii,min=1,max=20,lowercase" example:"first_seen"`
	Order   string `json:"order" validate:"omitempty,oneof=asc desc" example:"asc"`
}

// FileSearchResponse represents file search response results.
type FileSearchResponse struct {
	Results   interface{}
	TotalHits uint64
}

// AutoCompleteEntry represents a file search autocomplete entry.
type AutoCompleteEntry struct {
	Query   string `json:"query"`
	Comment string `json:"comment"`
}

// FileSearchAutocomplete represents the autocomplete example when using file search.
type FileSearchAutocomplete struct {
	Examples        []AutoCompleteEntry `json:"examples"`
	SearchModifiers []AutoCompleteEntry `json:"search_modifiers"`
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
	comSvc        comment.Service
	archiver      Archiver
}

// NewService creates a new File service.
func NewService(repo Repository, logger log.Logger,
	updown UploadDownloader, producer Producer, topic, bucket, samplesZipPwd string,
	userSvc user.Service, actSvc activity.Service, commentSvc comment.Service, arch Archiver) Service {
	return service{repo, logger, updown, producer, topic, bucket, samplesZipPwd,
		userSvc, actSvc, commentSvc, arch}
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

		go func() {
			existsCtx, cancelExistsFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelExistsFn()

			if ok, _ := s.objSto.Exists(existsCtx, s.bucket, sha256); ok {
				return
			}

			var err error

			for attempt := 0; attempt < 3; attempt++ {
				// Create a context with a timeout that will abort the upload if it takes
				// more than the passed in timeout.
				uploadCtx, cancelUploadFn := context.WithTimeout(context.Background(), fileUploadTimeout)

				// Ensure the context is canceled to prevent leaking.
				defer cancelUploadFn()

				err = s.objSto.Upload(uploadCtx, s.bucket, sha256, bytes.NewReader(fileContent))
				if err == nil {
					break
				}
				s.logger.With(uploadCtx).Error(err)

				// Give time to the system to recover
				time.Sleep(10 * time.Second)
			}

			// Check if the upload failed.
			if err != nil {
				s.logger.Error(err)
				return
			}

			// Serialize the msg to send to the orchestrator.
			msg, err := json.Marshal(FileScanCfg{SHA256: sha256, FileScanRequest: req.scanCfg})
			if err != nil {
				s.logger.Error(err)
				return
			}

			// Push a message to the queue to scan this file.
			err = s.producer.Produce(s.topic, msg)
			if err != nil {
				s.logger.Error(err)
				return
			}

		}()

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
			Meta:        &entity.DocMetadata{CreatedAt: now, LastUpdated: now, Version: 1},
			SHA256:      sha256,
			Type:        "file",
			FirstSeen:   now,
			Submissions: append(file.Submissions, submission),
			Status:      entity.FileScanProgressQueued,
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

		// Update user submissions.
		newSubmission := entity.UserSubmission{
			SHA256:    sha256,
			Timestamp: now,
		}
		if err = s.userSvc.Submit(ctx, user.ID(), newSubmission); err != nil {
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

	// update the last modified time
	file.Meta.LastUpdated = time.Now().Unix()

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

	// Add new `like` activity even if the file is already liked.
	if _, err = s.actSvc.Create(ctx, activity.CreateActivityRequest{
		Kind:     "like",
		Username: user.Username,
		Target:   sha256,
		Source:   source,
	}); err != nil {
		return err
	}

	newLike := entity.UserLike{
		SHA256:    sha256,
		Timestamp: time.Now().Unix(),
	}
	return s.userSvc.Like(ctx, user.ID(), newLike)
}

func (s service) Unlike(ctx context.Context, sha256 string) error {

	loggedInUser, _ := ctx.Value(entity.UserKey).(entity.User)
	user, err := s.userSvc.Get(ctx, loggedInUser.ID())
	if err != nil {
		return err
	}

	// delete corresponding activity.
	if err = s.actSvc.DeleteWith(ctx, "like", user.ID(),
		sha256); err != nil {
		return err
	}

	return s.userSvc.Unlike(ctx, user.ID(), sha256)
}

func (s service) ReScan(ctx context.Context, sha256 string, input FileScanRequest) error {

	// Serialize the msg to send to the orchestrator.
	msg, err := json.Marshal(FileScanCfg{SHA256: sha256, FileScanRequest: input})
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

func (s service) DownloadRaw(ctx context.Context, sha256 string, file io.Writer) (size int64, err error) {

	// Create a context with a timeout that will abort the download if it takes
	// more than the passed in timeout.
	downloadCtx, cancelFn := context.WithTimeout(
		context.Background(), time.Duration(time.Second*30))
	defer cancelFn()

	found, err := s.objSto.Exists(ctx, s.bucket, sha256)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return
	}

	if !found {
		return size, ErrObjectNotFound
	}

	size, err = s.objSto.DownloadWithSize(downloadCtx, s.bucket, sha256, file)
	if err != nil {
		s.logger.With(ctx).Error(err)
		return size, err
	}
	return size, nil
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

func (s service) CountStrings(ctx context.Context, id string, queryString string) (int, error) {
	count, err := s.repo.CountStrings(ctx, id, queryString)
	if err != nil {
		return 0, err
	}
	return count, err
}

func (s service) CountComments(ctx context.Context, id string) (int, error) {
	filterVal := map[string][]string{
		"sha256": {id},
	}
	ctx = comment.WithFilters(ctx, filterVal)
	count, err := s.comSvc.Count(ctx)
	if err != nil {
		return 0, err
	}
	return count, err
}

func (s service) Strings(ctx context.Context, id string, queryString string, offset, limit int) (
	interface{}, error) {

	result, err := s.repo.Strings(ctx, id, queryString, offset, limit)
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

func (s service) MetaUI(ctx context.Context, id string) (interface{}, error) {
	res, err := s.repo.MetaUI(ctx, id)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s service) Search(ctx context.Context, input FileSearchRequest) (
	FileSearchResponse, error) {

	result, err := s.repo.Search(ctx, input)
	if err != nil {
		return FileSearchResponse{}, err
	}
	return result, nil
}
