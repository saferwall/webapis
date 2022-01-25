// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package activity

import (
	"context"
	"encoding/json"
	"time"

	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Service encapsulates usecase logic for users.
type Service interface {
	Get(ctx context.Context, id string, fields []string) (Activity, error)
	Query(ctx context.Context, offset, limit int) ([]interface{}, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateActivityRequest) (Activity, error)
	Update(ctx context.Context, id string, input UpdateActivityRequest) (Activity, error)
	Delete(ctx context.Context, id string) (Activity, error)
}

// Activity represents the data about an activity.
type Activity struct {
	entity.Activity
}

type service struct {
	repo   Repository
	logger log.Logger
}

// CreateActivityRequest represents a user creation request.
type CreateActivityRequest struct {
	Kind     string `json:"kind" validate:"required,alpha"`
	Username string `json:"username" validate:"required,alpha"`
	Target   string `json:"target" validate:"required,alpha"`
	Source   string `json:"src" validate:"required,alpha"`
}

// UpdateActivityRequest represents a user update request.
type UpdateActivityRequest struct {
	Target string `json:"target" validate:"required,alpha"`
}

// NewService creates a new user service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the user with the specified user ID.
func (s service) Get(ctx context.Context, id string, fields []string) (Activity, error) {
	activity, err := s.repo.Get(ctx, id, fields)
	if err != nil {
		return Activity{}, err
	}
	return Activity{activity}, nil
}

// Create creates a new activity.
func (s service) Create(ctx context.Context, req CreateActivityRequest) (
	Activity, error) {

	id := entity.ID()
	now := time.Now()
	err := s.repo.Create(ctx, entity.Activity{
		ID:        id,
		Kind:      req.Kind,
		Type:      "activity",
		Username:  req.Username,
		Target:    req.Target,
		Timestamp: now.Unix(),
		Source:    req.Source,
	})
	if err != nil {
		return Activity{}, err
	}
	return s.Get(ctx, id, nil)
}

// Update updates the activity with the specified ID.
func (s service) Update(ctx context.Context, id string, req UpdateActivityRequest) (
	Activity, error) {

	activity, err := s.Get(ctx, id, nil)
	if err != nil {
		return activity, err
	}

	data, err := json.Marshal(req)
	if err != nil {
		return activity, err
	}
	err = json.Unmarshal(data, &activity)
	if err != nil {
		return activity, err
	}

	// check if activity.Username == id
	if err := s.repo.Update(ctx, id, activity.Activity); err != nil {
		return activity, err
	}

	return activity, nil
}

// Delete deletes the activity with the specified ID.
func (s service) Delete(ctx context.Context, id string) (Activity, error) {
	activity, err := s.Get(ctx, id, nil)
	if err != nil {
		return Activity{}, err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		return Activity{}, err
	}
	return activity, nil
}

// Count returns the number of activities.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Query returns the activities with the specified offset and limit.
func (s service) Query(ctx context.Context, offset, limit int) (
	[]interface{}, error) {

	result, err := s.repo.Query(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}
