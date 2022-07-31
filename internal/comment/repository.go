// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"context"
	"strings"

	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Repository encapsulates the logic to access comments from the data source.
type Repository interface {
	// Get returns the comment with the specified comment ID.
	Get(ctx context.Context, id string) (entity.Comment, error)
	// Create saves a new comment in the storage.
	Create(ctx context.Context, Comment entity.Comment) error
	// Update updates the comment with given ID in the storage.
	Update(ctx context.Context, User entity.Comment) error
	// Delete removes the comment with given ID from the storage.
	Delete(ctx context.Context, id string) error
	// Exists checks if a comment exists with a given ID.
	Exists(ctx context.Context, id string) (bool, error)
}

// repository persists comments in database.
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new comment repository.
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the comment with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.Comment, error) {
	var com entity.Comment
	key := strings.ToLower(id)
	err := r.db.Get(ctx, key, &com)
	return com, err
}

// Create saves a new comment record in the database.
// It returns the ID of the newly inserted comment record.
func (r repository) Create(ctx context.Context, comment entity.Comment) error {
	return r.db.Create(ctx, comment.ID, &comment)
}

// Update updates the changes to a comment in the database.
func (r repository) Update(ctx context.Context, comment entity.Comment) error {
	return r.db.Update(ctx, comment.ID, &comment)
}

// Delete deletes a comment with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	return r.db.Delete(ctx, id)
}

// Exists checks if a comment exists for the given id.
func (r repository) Exists(ctx context.Context, id string) (bool, error) {
	docExists := false
	key := strings.ToLower(id)
	err := r.db.Exists(ctx, key, &docExists)
	return docExists, err
}
