// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package behavior

import (
	"context"

	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Service encapsulates usecase logic for behaviors.
type Service interface {
	Get(ctx context.Context, id string, fields []string) (Behavior, error)
	Exists(ctx context.Context, id string) (bool, error)
	Count(ctx context.Context) (int, error)
	CountAPIs(ctx context.Context, id string) (int, error)
	CountEvents(ctx context.Context, id string) (int, error)
	APIs(ctx context.Context, id string, offset, limit int) (interface{}, error)
	Events(ctx context.Context, id string, offset, limit int) (interface{}, error)
}

// Behavior represents the data about a behavior scan.
type Behavior struct {
	entity.Behavior
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the file behavior scan given its ID.
func (s service) Get(ctx context.Context, id string, fields []string) (Behavior, error) {
	behavior, err := s.repo.Get(ctx, id, fields)
	if err != nil {
		return Behavior{}, err
	}
	return Behavior{behavior}, nil
}

// Exists checks if a document exists for the given id.
func (s service) Exists(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
}

// Count returns the number of behavior scans.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s service) CountAPIs(ctx context.Context, id string) (int, error) {
	count, err := s.repo.CountAPIs(ctx, id)
	if err != nil {
		return 0, err
	}
	return count, err
}

func (s service) CountEvents(ctx context.Context, id string) (int, error) {
	count, err := s.repo.CountEvents(ctx, id)
	if err != nil {
		return 0, err
	}
	return count, err
}

func (s service) APIs(ctx context.Context, id string, offset, limit int) (
	interface{}, error) {

	result, err := s.repo.APIs(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s service) Events(ctx context.Context, id string, offset, limit int) (
	interface{}, error) {

	result, err := s.repo.Events(ctx, id, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}
