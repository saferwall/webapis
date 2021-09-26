// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"context"
	"time"

	"github.com/saferwall/saferwall-api/internal/activity"
	"github.com/saferwall/saferwall-api/internal/entity"
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
	Get(ctx context.Context, id string) (Comment, error)
	Create(ctx context.Context, input CreateCommentRequest) (Comment, error)
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
	Username string `json:"username" validate:"required,alphanum,min=1,max=20"`
	SHA256   string `json:"sha256" validate:"required,alphanum,len=64"`
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger, actSvc activity.Service,
	userSvc user.Service, fileSvc file.Service) Service {
	return service{repo, logger, actSvc, userSvc, fileSvc}
}

// Create creates a new user.
func (s service) Create(ctx context.Context, req CreateCommentRequest) (
	Comment, error) {

	now := time.Now()
	err := s.repo.Create(ctx, entity.Comment{
		Type:      "comment",
		ID:        entity.ID(),
		Body:      req.Body,
		SHA256:    req.SHA256,
		Username:  req.Username,
		Timestamp: now.Unix(),
	})
	if err != nil {
		return Comment{}, err
	}
	return s.Get(ctx, req.Username)
}

// Get returns the comment with the specified comment ID.
func (s service) Get(ctx context.Context, id string) (Comment, error) {
	com, err := s.repo.Get(ctx, id)
	if err != nil {
		return Comment{}, err
	}
	return Comment{com}, nil
}
