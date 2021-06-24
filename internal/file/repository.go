// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"context"
	"strings"

	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Repository encapsulates the logic to access users from the data source.
type Repository interface {
	// Get returns the file with the specified file ID.
	Get(ctx context.Context, id string) (entity.File, error)
	// Count returns the number of users.
	Count(ctx context.Context) (int, error)
	// Query returns the list of users with the given offset and limit.
	Query(ctx context.Context, offset, limit int) ([]entity.File, error)
	// Create saves a new file in the storage.
	Create(ctx context.Context, file entity.File) error
	// Update updates the file with given ID in the storage.
	Update(ctx context.Context, file entity.File) error
	// Delete removes the file with given ID from the storage.
	Delete(ctx context.Context, id string) error
}

// repository persists users in database.
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new file repository.
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the file with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.File, error) {
	var file entity.File
	key := strings.ToLower("files::" + id)
	err := r.db.Get(ctx, key, &file)
	return file, err
}

// Create saves a new file record in the database.
// It returns the ID of the newly inserted file record.
func (r repository) Create(ctx context.Context, file entity.File) error {
	key := "files::" + file.SHA256
	return r.db.Create(ctx, key, &file)
}

// Update saves the changes to a file in the database.
func (r repository) Update(ctx context.Context, file entity.File) error {
	key := "files::" + file.SHA256
	return r.db.Update(ctx, key, &file)
}

// Delete deletes a file with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	key := "files::" + id
	return r.db.Delete(ctx, key)
}

// Count returns the number of the file records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int

	err := r.db.Count(ctx, "files", &count)
	return count, err
}

// Query retrieves the file records with the specified offset and limit
// from the database.
func (r repository) Query(ctx context.Context, offset, limit int) (
	[]entity.File, error) {
	var users []entity.File
	// err := r.db.With(ctx).
	// 	Select().
	// 	OrderBy("id").
	// 	Offset(int64(offset)).
	// 	Limit(int64(limit)).
	// 	All(&users)
	return users, nil
}
