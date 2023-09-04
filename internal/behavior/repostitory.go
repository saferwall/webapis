// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package behavior

import (
	"context"
	"encoding/json"
	"fmt"

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

	CountAPIs(ctx context.Context, id string) (int, error)
	CountEvents(ctx context.Context, id string) (int, error)
	APIs(ctx context.Context, id string, offset, limit int) (
		interface{}, error)
	Events(ctx context.Context, id string, offset, limit int) (
		interface{}, error)
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
	var res interface{}

	// if only some fields are wanted from the whole document.
	if len(fields) > 0 {
		err = r.db.Lookup(ctx, id, fields, &behavior)
	} else {
		params := make(map[string]interface{}, 1)
		params["behavior_id"] = id
		params["behavior_id_apis"] = id + "::apis"
		params["behavior_id_events"] = id + "::events"
		statement := r.db.N1QLQuery[dbcontext.BehaviorReport]
		err := r.db.Query(ctx, statement, params, &res)
		if err != nil {
			return entity.Behavior{}, err
		}

		behaviors := res.([]interface{})
		b, _ := json.Marshal(behaviors[0])
		json.Unmarshal(b, &behavior)
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

// CountStrings returns the number of strings in a file doc in the database.
func (r repository) CountAPIs(ctx context.Context, id string) (int, error) {
	var count int
	var statement string
	params := make(map[string]interface{}, 1)
	params["id"] = id + "::apis"

	filters, ok := ctx.Value(filtersKey).(map[string][]string)
	if ok {
		statement =
			"SELECT RAW COUNT(api) AS count FROM `" + r.db.Bucket.Name() + "` d" +
				" USE KEYS $id UNNEST d.api_trace as api"
		i := 0
		for k, v := range filters {
			if i == 0 {
				statement += " WHERE"
			} else {
				statement += " AND"
			}
			i++
			statement += fmt.Sprintf(" api.%s IN $%s", k, k)
			params[k] = v
		}
	} else {
		statement =
			"SELECT RAW ARRAY_LENGTH(d.api_trace) AS count FROM `" + r.db.Bucket.Name() + "` d" +
				" USE KEYS $id"
	}

	err := r.db.Count(ctx, statement, params, &count)
	return count, err
}

// CountEvents returns the number of strings in a file doc in the database.
func (r repository) CountEvents(ctx context.Context, id string) (int, error) {
	var count int
	var statement string
	params := make(map[string]interface{}, 1)
	params["id"] = id + "::events"

	filters, ok := ctx.Value(filtersKey).(map[string][]string)
	if ok {
		statement =
			"SELECT RAW COUNT(event) AS count FROM `" + r.db.Bucket.Name() + "` d" +
				" USE KEYS $id UNNEST d.sys_events as event"
		i := 0
		for k, v := range filters {
			if i == 0 {
				statement += " WHERE"
			} else {
				statement += " AND"
			}
			i++
			statement += fmt.Sprintf(" event.%s IN $%s", k, k)
			params[k] = v
		}
	} else {
		statement =
			"SELECT RAW ARRAY_LENGTH(d.sys_events) AS count FROM `" + r.db.Bucket.Name() + "` d" +
				" USE KEYS $id"
	}

	err := r.db.Count(ctx, statement, params, &count)
	return count, err
}

func (r repository) APIs(ctx context.Context, id string, offset,
	limit int) (interface{}, error) {

	var results interface{}
	var statement string

	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["id"] = id + "::apis"

	filters, ok := ctx.Value(filtersKey).(map[string][]string)
	if ok {
		statement =
			"SELECT RAW api FROM `" + r.db.Bucket.Name() + "` d" +
				" USE KEYS $id UNNEST d.api_trace as api"
		i := 0
		for k, v := range filters {
			if i == 0 {
				statement += " WHERE"
			} else {
				statement += " AND"
			}
			i++
			statement += fmt.Sprintf(" api.%s IN $%s", k, k)
			params[k] = v
		}
		statement += " OFFSET $offset LIMIT $limit"
	} else {
		statement =
			"SELECT RAW api_trace[$offset:$limit] FROM `" + r.db.Bucket.Name() +
				"` USE KEYS $id"

	}

	err := r.db.Query(ctx, statement, params, &results)
	if err != nil {
		return nil, err
	}
	if len(results.([]interface{})) == 0 {
		return results, nil
	} else if len(results.([]interface{})) == 1 {
		return results.([]interface{})[0], nil
	}
	return results.([]interface{}), nil
}

func (r repository) Events(ctx context.Context, id string, offset,
	limit int) (interface{}, error) {

	var results interface{}
	var statement string

	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["id"] = id + "::events"

	filters, ok := ctx.Value(filtersKey).(map[string][]string)
	if ok {
		statement =
			"SELECT RAW event FROM `" + r.db.Bucket.Name() + "` d" +
				" USE KEYS $id UNNEST d.sys_events as event"
		i := 0
		for k, v := range filters {
			if i == 0 {
				statement += " WHERE"
			} else {
				statement += " AND"
			}
			i++
			statement += fmt.Sprintf(" event.%s IN $%s", k, k)
			params[k] = v
		}
		statement += " OFFSET $offset LIMIT $limit"
	} else {
		statement =
			"SELECT RAW sys_events[$offset:$limit] FROM `" + r.db.Bucket.Name() +
				"` USE KEYS $id"

	}

	err := r.db.Query(ctx, statement, params, &results)
	if err != nil {
		return nil, err
	}
	if len(results.([]interface{})) == 0 {
		return results, nil
	} else if len(results.([]interface{})) == 1 {
		return results.([]interface{})[0], nil
	}
	return results.([]interface{}), nil
}
