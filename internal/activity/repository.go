// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE activity.

package activity

import (
	"context"
	"fmt"
	"strings"

	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Repository encapsulates the logic to access users from the data source.
type Repository interface {
	// Get returns the activity with the specified activity ID.
	Get(ctx context.Context, id string, fields []string) (entity.Activity, error)
	// Count returns the number of activities.
	Count(ctx context.Context) (int, error)
	// Query returns the list of activities with the given offset and limit.
	Query(ctx context.Context, offset, limit int) ([]interface{}, error)
	// Create saves a new activity in the storage.
	Create(ctx context.Context, activity entity.Activity) error
	// Update updates the whole activity with given ID in the storage.
	Update(ctx context.Context, key string, activity entity.Activity) error
	// Patch patches a sub entry in the activity with given ID in the storage.
	Patch(ctx context.Context, key, path string, val interface{}) error
	// Delete removes the activity with given ID from the storage.
	Delete(ctx context.Context, id string) error
	// Delete an activity given its kind, username and target.
	DeleteWith(ctx context.Context, kind, username, target string) error
}

// repository persists users in database.
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new activity repository.
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the activity with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string, fields []string) (
	entity.Activity, error) {

	var err error
	var activity entity.Activity

	// if only some fields are wanted from the whole document.
	if len(fields) > 0 {
		err = r.db.Lookup(ctx, id, fields, &activity)
	} else {
		err = r.db.Get(ctx, id, &activity)
	}

	return activity, err
}

// Create saves a new activity record in the database.
// It returns the ID of the newly inserted activity record.
func (r repository) Create(ctx context.Context,
	activity entity.Activity) error {
	return r.db.Create(ctx, activity.ID, &activity)
}

// Update saves the changes to an activity in the database.
func (r repository) Update(ctx context.Context, id string,
	activity entity.Activity) error {
	return r.db.Update(ctx, id, &activity)
}

// Patch performs a sub doc update to an activity in the database.
func (r repository) Patch(ctx context.Context, key, path string,
	val interface{}) error {
	return r.db.Patch(ctx, key, path, val)
}

// Delete deletes an activity with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	return r.db.Delete(ctx, id)
}

// Count returns the number of the activity records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	params := make(map[string]interface{}, 1)
	params["docType"] = "activity"

	statement :=
		"SELECT RAW COUNT(*) AS count FROM `" + r.db.Bucket.Name() + "` " +
			"WHERE `type`=$docType"

	err := r.db.Count(ctx, statement, params, &count)
	return count, err
}

// Query retrieves the activity records with the specified offset and limit
// from the database.
func (r repository) Query(ctx context.Context, offset, limit int) (
	[]interface{}, error) {
	statement :=
		`
	SELECT {
		"type": activity.kind,
		"author": {
			"username": activity.username,
			"member_since": (
				SELECT RAW u.member_since FROM` + " `sfw` " +
			`u USE KEYS activity.username)[0]},
		"comment": f.body,
		"timestamp": activity.timestamp}.*,
		(CASE WHEN activity.kind = "follow" THEN
		{"target": activity.target } ELSE
		{"file": {
			"hash": f.sha256,
			"tags": f.tags,
			"filename": f.submissions[0].filename,
			"class": f.ml.pe.predicted_class,
			"multiav": CONCAT ( TOSTRING ( ARRAY_COUNT (
				array_flatten(array i.infected
				for i in OBJECT_VALUES(f.multiav.last_scan)
				when i.infected=true end, 1))), "/",
				TOSTRING(OBJECT_LENGTH(f.multiav.last_scan)))}} END).*
	  FROM` + " `sfw` " + `activity
	  LEFT JOIN` + " `sfw` " + `f ON KEYS activity.target
	  WHERE activity.type = 'activity'
	  ORDER BY activity.timestamp DESC
	  OFFSET ` + fmt.Sprintf("%d", offset) + " LIMIT " + fmt.Sprintf("%d", limit)

	var activities interface{}
	err := r.db.Query(ctx, statement, nil, &activities)
	if err != nil {
		return nil, err
	}
	return activities.([]interface{}), nil
}

// Delete an activity given its kind, username and target.
func (r repository) DeleteWith(ctx context.Context, kind, username,
	target string) error {

	var result interface{}
	params := make(map[string]interface{}, 1)
	params["kind"] = kind
	params["username"] = strings.ToLower(username)
	params["target"] = target
	query := r.db.N1QLQuery[dbcontext.DeleteActivity]
	return r.db.Query(ctx, query, params, &result)
}
