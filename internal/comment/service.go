// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/saferwall/saferwall-api/internal/activity"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/internal/file"
	"github.com/saferwall/saferwall-api/internal/user"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Comment represents a comment made by a user for a file.
type Comment struct {
	entity.Comment
}

// Service encapsulates usecase logic for files.
type Service interface {
	Exists(ctx context.Context, id string) (bool, error)
	Get(ctx context.Context, id string) (Comment, error)
	Create(ctx context.Context, input CreateCommentRequest) (Comment, error)
	Update(ctx context.Context, id string, input UpdateCommentRequest) (Comment, error)
	Delete(ctx context.Context, id string) (Comment, error)
}

type service struct {
	repo    Repository
	logger  log.Logger
	actSvc  activity.Service
	userSvc user.Service
	fileSvc file.Service
}

// CreateCommentRequest represents a comment creation request.
type CreateCommentRequest struct {
	Body     string `json:"body" validate:"required"`
	SHA256   string `json:"sha256" validate:"required,alphanum,len=64"`
	Username string
}

// UpdateCommentRequest represents a comment update request.
type UpdateCommentRequest struct {
	Body string `json:"body" validate:"required"`
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger, actSvc activity.Service,
	userSvc user.Service, fileSvc file.Service) Service {
	return service{repo, logger, actSvc, userSvc, fileSvc}
}

// Exists checks if a comment exists for the given id.
func (s service) Exists(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
}

// Create creates a new comment.
func (s service) Create(ctx context.Context, req CreateCommentRequest) (
	Comment, error) {

	now := time.Now()
	id := entity.ID()
	err := s.repo.Create(ctx, entity.Comment{
		Type:      "comment",
		ID:        id,
		Body:      req.Body,
		SHA256:    req.SHA256,
		Username:  req.Username,
		Timestamp: now.Unix(),
	})
	if err != nil {
		return Comment{}, err
	}

	user, err := s.userSvc.Get(ctx, req.Username)
	if err != nil {
		return Comment{}, err
	}

	file, err := s.fileSvc.Get(ctx, req.SHA256, nil)
	if err != nil {
		return Comment{}, err
	}

	// Update comments count on user object.
	err = s.userSvc.Patch(ctx, req.Username, "comments_count", user.CommentsCount+1)
	if err != nil {
		return Comment{}, err
	}

	// Update comments count on file object.
	err = s.fileSvc.Patch(ctx, req.SHA256, "comments_count", *file.CommentsCount+1)
	if err != nil {
		return Comment{}, err
	}

	// Get the source of the HTTP request from the ctx.
	source, _ := ctx.Value(entity.SourceKey).(string)

	// Create a new `comment` activity.
	if _, err = s.actSvc.Create(ctx, activity.CreateActivityRequest{
		Kind:     "comment",
		Username: user.Username,
		Target:   id,
		Source:   source,
	}); err != nil {
		return Comment{}, err
	}
	return s.Get(ctx, id)
}

// Get returns the comment with the specified comment ID.
func (s service) Get(ctx context.Context, id string) (Comment, error) {
	com, err := s.repo.Get(ctx, id)
	if err != nil {
		return Comment{}, err
	}
	return Comment{com}, nil
}

// Update updates the comment with the specified ID.
func (s service) Update(ctx context.Context, id string, input UpdateCommentRequest) (
	Comment, error) {

	var curUsername string

	comment, err := s.Get(ctx, id)
	if err != nil {
		return comment, err
	}

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		curUsername = user.ID()
	}

	if comment.Username != curUsername {
		return comment, errors.Forbidden("")
	}

	data, err := json.Marshal(input)
	if err != nil {
		return comment, err
	}

	err = json.Unmarshal(data, &comment)
	if err != nil {
		return comment, err
	}

	if err := s.repo.Update(ctx, comment.Comment); err != nil {
		return comment, err
	}

	return comment, nil
}

// Delete deletes the comment with the specified ID.
func (s service) Delete(ctx context.Context, id string) (Comment, error) {
	com, err := s.Get(ctx, id)
	if err != nil {
		return Comment{}, err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		return Comment{}, err
	}

	// delete corresponsing activity.
	if s.actSvc.DeleteWith(ctx, "comment", com.Username, id); err != nil {
		return Comment{}, err
	}

	return com, nil
}
