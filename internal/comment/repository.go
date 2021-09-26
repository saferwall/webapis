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
