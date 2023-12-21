// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Repository encapsulates the logic to access files from the data source.
type Repository interface {
	// Get returns the file with the specified file ID.
	Get(ctx context.Context, id string, fields []string) (entity.File, error)
	// Count returns the number of files.
	Count(ctx context.Context) (int, error)
	// Exists return true when the doc exists in the DB.
	Exists(ctx context.Context, id string) (bool, error)
	// Query returns the list of files with the given offset and limit.
	Query(ctx context.Context, offset, limit int, fields []string) ([]entity.File, error)
	// Create saves a new file in the storage.
	Create(ctx context.Context, id string, file entity.File) error
	// Update updates the whole file with given ID in the storage.
	Update(ctx context.Context, key string, file entity.File) error
	// Patch patches a sub entry in the file with given ID in the storage.
	Patch(ctx context.Context, key, path string, val interface{}) error
	// Delete removes the file with given ID from the storage.
	Delete(ctx context.Context, id string) error
	// Summary returns a summary of a file scan.
	Summary(ctx context.Context, id string) (interface{}, error)
	// Comments returns the list of comments over a file.
	Comments(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	CountStrings(ctx context.Context, id string) (int, error)
	Strings(ctx context.Context, id string, offset, limit int) (
		interface{}, error)
}

// repository persists files in database.
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new file repository.
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the file with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string, fields []string) (
	entity.File, error) {

	var err error
	var file entity.File

	key := file.ID(id)

	// if only some fields are wanted from the whole document.
	if len(fields) > 0 {
		err = r.db.Lookup(ctx, key, fields, &file)
	} else {
		err = r.db.Get(ctx, key, &file)
	}

	return file, err
}

// Create saves a new file record in the database.
// It returns the ID of the newly inserted file record.
func (r repository) Create(ctx context.Context, key string,
	file entity.File) error {
	return r.db.Create(ctx, file.ID(key), &file)
}

// Exists checks if a document exists for the given id.
func (r repository) Exists(ctx context.Context, key string) (bool, error) {
	docExists := false
	err := r.db.Exists(ctx, key, &docExists)
	return docExists, err
}

// Update saves the changes to a file in the database.
func (r repository) Update(ctx context.Context, key string,
	file entity.File) error {
	return r.db.Update(ctx, file.ID(key), &file)
}

// Patch performs a sub doc update to a file in the database.
func (r repository) Patch(ctx context.Context, key, path string,
	val interface{}) error {
	return r.db.Patch(ctx, key, path, val)
}

// Delete deletes a file with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	key := strings.ToLower(id)
	return r.db.Delete(ctx, key)
}

// Count returns the number of the file records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int

	params := make(map[string]interface{}, 1)
	params["docType"] = "file"

	statement :=
		"SELECT RAW COUNT(*) AS count FROM `" + r.db.Bucket.Name() + "` " +
			"WHERE `type`=$docType"

	err := r.db.Count(ctx, statement, params, &count)
	return count, err
}

// Query retrieves the file records with the specified offset and limit
// from the database.
func (r repository) Query(ctx context.Context, offset, limit int, fields []string) (
	[]entity.File, error) {
	var res interface{}

	params := make(map[string]interface{}, 1)
	params["docType"] = "file"
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
		return []entity.File{}, err
	}
	files := []entity.File{}
	for _, u := range res.([]interface{}) {
		file := entity.File{}
		b, _ := json.Marshal(u)
		json.Unmarshal(b, &file)
		files = append(files, file)
	}
	return files, nil
}

func (r repository) Summary(ctx context.Context, id string) (
	interface{}, error) {

	var results interface{}
	var query string
	params := make(map[string]interface{}, 1)
	params["sha256"] = strings.ToLower(id)

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		params["loggedInUser"] = user.ID()
	} else {
		params["loggedInUser"] = "_none_"
	}

	query = r.db.N1QLQuery[dbcontext.FileSummary]
	err := r.db.Query(ctx, query, params, &results)
	if err != nil {
		return nil, err
	}

	if len(results.([]interface{})) == 0 {
		return results, nil
	}
	return results.([]interface{})[0], nil
}

func (r repository) Comments(ctx context.Context, id string, offset,
	limit int) ([]interface{}, error) {

	var results interface{}

	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["sha256"] = strings.ToLower(id)

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		params["loggedInUser"] = user.ID()
	} else {
		params["loggedInUser"] = "_none_"

	}

	query := r.db.N1QLQuery[dbcontext.FileComments]
	err := r.db.Query(ctx, query, params, &results)
	if err != nil {
		return nil, err
	}
	return results.([]interface{}), nil
}

// CountStrings returns the number of strings in a file doc in the database.
func (r repository) CountStrings(ctx context.Context, id string) (int, error) {
	var count int

	params := make(map[string]interface{}, 1)
	params["sha256"] = strings.ToLower(id)

	query := r.db.N1QLQuery[dbcontext.CountStrings]
	err := r.db.Count(ctx, query, params, &count)
	return count, err
}

func (r repository) Strings(ctx context.Context, id string, offset,
	limit int) (interface{}, error) {

	var results interface{}

	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["sha256"] = strings.ToLower(id)

	query := r.db.N1QLQuery[dbcontext.FileStrings]
	err := r.db.Query(ctx, query, params, &results)
	if err != nil {
		return nil, err
	}
	if len(results.([]interface{})) == 0 {
		return results, nil
	}
	return results.([]interface{})[0], nil
}
