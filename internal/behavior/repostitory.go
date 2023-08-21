// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package behavior

import (
	"context"
	"encoding/json"

	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Repository encapsulates the logic to access behavior scans from the data source.
type Repository interface {
	// Get returns the file behavior with the specified behavior scan ID.
	Get(ctx context.Context, id string, fields []string) (entity.Behavior, error)
	// Count returns the number of behavior scans.
	Count(ctx context.Context) (int, error)
	// Exists return true when the behavior scan ID exists in the DB.
	Exists(ctx context.Context, id string) (bool, error)
	// Query returns the list of file behavior scans with the given offset and limit.
	Query(ctx context.Context, offset, limit int) ([]entity.Behavior, error)
}

// repository persists file scan behaviors in database.
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new behavior repository.
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the behavior scan with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string, fields []string) (
	entity.Behavior, error) {

	var err error
	var behavior entity.Behavior

	// if only some fields are wanted from the whole document.
	if len(fields) > 0 {
		err = r.db.Lookup(ctx, id, fields, &behavior)
	} else {
		err = r.db.Get(ctx, id, &behavior)
	}

	return behavior, err
}

// Exists checks if a document exists for the given id.
func (r repository) Exists(ctx context.Context, key string) (bool, error) {
	docExists := false
	err := r.db.Exists(ctx, key, &docExists)
	return docExists, err
}

// Count returns the number of the behavior scan records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int

	params := make(map[string]interface{}, 1)
	params["docType"] = "behavior"

	statement :=
		"SELECT RAW COUNT(*) AS count FROM `" + r.db.Bucket.Name() + "` " +
			"WHERE `type`=$docType"

	err := r.db.Count(ctx, statement, params, &count)
	return count, err
}

// Query retrieves the file records with the specified offset and limit
// from the database.
func (r repository) Query(ctx context.Context, offset, limit int) (
	[]entity.Behavior, error) {
	var res interface{}

	params := make(map[string]interface{}, 1)
	params["docType"] = "behavior"
	params["offset"] = offset
	params["limit"] = limit

	statement := r.db.N1QLQuery[dbcontext.GetAllDocType]
	err := r.db.Query(ctx, statement, params, &res)
	if err != nil {
		return []entity.Behavior{}, err
	}
	behaviors := []entity.Behavior{}
	for _, u := range res.([]interface{}) {
		behavior := entity.Behavior{}
		b, _ := json.Marshal(u)
		json.Unmarshal(b, &behavior)
		behaviors = append(behaviors, behavior)
	}
	return behaviors, nil
}
