// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"context"
	"strings"

	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Repository encapsulates the logic to access users from the data source.
type Repository interface {
	// Get returns the user with the specified user ID.
	Get(ctx context.Context, id string) (entity.User, error)
	// Count returns the number of users.
	Count(ctx context.Context) (int, error)
	// Query returns the list of users with the given offset and limit.
	Query(ctx context.Context, offset, limit int) ([]entity.User, error)
	// Create saves a new user in the storage.
	Create(ctx context.Context, User entity.User) error
	// Update updates the user with given ID in the storage.
	Update(ctx context.Context, User entity.User) error
	// Delete removes the user with given ID from the storage.
	Delete(ctx context.Context, id string) error
}

// repository persists users in database.
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new user repository.
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the user with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.User, error) {
	var user entity.User
	key := strings.ToLower("users::" + id)
	err := r.db.Get(ctx, key, &user)
	return user, err
}

// Create saves a new user record in the database.
// It returns the ID of the newly inserted user record.
func (r repository) Create(ctx context.Context, user entity.User) error {
	key := strings.ToLower("users::" + user.Username)
	return r.db.Create(ctx, key, &user)
}

// Update saves the changes to a user in the database.
func (r repository) Update(ctx context.Context, user entity.User) error {
	key := "users::" + user.Username
	return r.db.Update(ctx, key, &user)
}

// Delete deletes a user with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	key := strings.ToLower("users::" + id)
	return r.db.Delete(ctx, key)
}

// Count returns the number of the user records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int

	err := r.db.Count(ctx, "users", &count)
	return count, err
}

// Query retrieves the user records with the specified offset and limit
// from the database.
func (r repository) Query(ctx context.Context, offset, limit int) (
	[]entity.User, error) {
	var users []entity.User
	// err := r.db.With(ctx).
	// 	Select().
	// 	OrderBy("id").
	// 	Offset(int64(offset)).
	// 	Limit(int64(limit)).
	// 	All(&users)
	return users, nil
}
