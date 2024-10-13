// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package comment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Repository encapsulates the logic to access comments from the data source.
type Repository interface {
	// Get returns the comment with the specified comment ID.
	Get(ctx context.Context, id string, fields []string) (entity.Comment, error)
	// Create saves a new comment in the storage.
	Create(ctx context.Context, Comment entity.Comment) error
	// Update updates the comment with given ID in the storage.
	Update(ctx context.Context, User entity.Comment) error
	// Delete removes the comment with given ID from the storage.
	Delete(ctx context.Context, id string) error
	// Exists checks if a comment exists with a given ID.
	Exists(ctx context.Context, id string) (bool, error)
	// Count returns the number of comments.
	Count(ctx context.Context, fields []string) (int, error)
	// Query returns the list of comments with the given offset and limit.
	Query(ctx context.Context, offset, limit int, fields []string) ([]entity.Comment, error)
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
func (r repository) Get(ctx context.Context, id string, fields []string) (entity.Comment, error) {
	var com entity.Comment
	var err error

	key := strings.ToLower(id)

	// if only some fields are wanted from the whole document.
	if len(fields) > 0 {
		err = r.db.Lookup(ctx, key, fields, &com)
	} else {
		err = r.db.Get(ctx, key, &com)
	}

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

// Count returns the number of the comment records in the database.
func (r repository) Count(ctx context.Context, fields []string) (int, error) {
	var count int

	params := make(map[string]interface{}, 1)
	params["docType"] = "comment"

	statement :=
		"SELECT RAW COUNT(*) AS count FROM `" + r.db.Bucket.Name() + "` " +
			"WHERE `type`=$docType"

	if len(fields) > 0 {
		for _, field := range fields {
			statement += fmt.Sprintf(" AND %s=$%s", field, field)
			params[field] = field

		}
	}
	err := r.db.Count(ctx, statement, params, &count)
	return count, err
}

// Query retrieves the comment records with the specified offset and limit
// from the database.
func (r repository) Query(ctx context.Context, offset, limit int, fields []string) (
	[]entity.Comment, error) {
	var res interface{}

	params := make(map[string]interface{}, 1)
	params["docType"] = "comment"
	params["offset"] = offset
	params["limit"] = limit
	statement := r.db.N1QLQuery[dbcontext.GetAllDocType]

	if len(fields) > 0 {
		statement = "SELECT "
		for _, field := range fields {
			statement += fmt.Sprintf("%s,", field)
		}
		statement = strings.TrimSuffix(statement, ",")
		statement += fmt.Sprintf(" FROM `%s` WHERE type = $docType OFFSET $offset LIMIT $limit",
			r.db.Bucket.Name())

	}

	err := r.db.Query(ctx, statement, params, &res)
	if err != nil {
		return []entity.Comment{}, err
	}
	comments := []entity.Comment{}
	for _, u := range res.([]interface{}) {
		comment := entity.Comment{}
		b, _ := json.Marshal(u)
		_ = json.Unmarshal(b, &comment)
		comments = append(comments, comment)
	}
	return comments, nil
}
