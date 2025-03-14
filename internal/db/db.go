// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	gocb "github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocb/v2/search"
	"github.com/saferwall/saferwall-api/internal/query-parser/gen"
)

const (
	// Duration to wait until memd connections have been established with
	// the server and are ready.
	timeout = 30 * time.Second
)

var (
	// ErrDocumentNotFound is returned when the doc does not exist in the DB.
	ErrDocumentNotFound = errors.New("document not found")
	ErrSubDocNotFound   = gocb.ErrPathNotFound
)

// DB represents the database connection.
type DB struct {
	Bucket       *gocb.Bucket
	Cluster      *gocb.Cluster
	Collection   *gocb.Collection
	N1QLQuery    map[n1qlQuery]string
	FTSIndexName string
}

// Open opens a connection to the database.
func Open(server, username, password, bucketName, ftsIndexName string) (*DB, error) {

	// Get a couchbase cluster instance.
	cluster, err := gocb.Connect(
		server,
		gocb.ClusterOptions{
			Username: username,
			Password: password,
		})
	if err != nil {
		return nil, err
	}

	// Get a bucket reference.
	bucket := cluster.Bucket(bucketName)

	// We wait until the bucket is definitely connected and setup.
	err = bucket.WaitUntilReady(timeout, nil)
	if err != nil {
		return nil, err
	}

	// Get a collection reference.
	collection := bucket.DefaultCollection()

	// Create primary indexe.
	mgr := cluster.QueryIndexes()
	err = mgr.CreatePrimaryIndex(bucketName,
		&gocb.CreatePrimaryQueryIndexOptions{IgnoreIfExists: true})
	if err != nil {
		return nil, err
	}

	return &DB{
		Bucket:       bucket,
		Cluster:      cluster,
		Collection:   collection,
		FTSIndexName: ftsIndexName,
	}, nil
}

// Exists checks weather a document exists in the DB.
func (db *DB) Exists(ctx context.Context, key string, docExists *bool) error {
	existsResult, err := db.Collection.Exists(key, &gocb.ExistsOptions{})
	if err != nil {
		return err
	}

	*docExists = existsResult.Exists()
	return nil
}

// Query executes a N1QL query.
func (db *DB) Query(ctx context.Context, statement string,
	args map[string]interface{}, val *interface{}) error {

	results, err := db.Cluster.Query(statement, &gocb.QueryOptions{
		NamedParameters: args, Adhoc: true})
	if err != nil {
		return err
	}

	rows := []interface{}{}
	for results.Next() {
		var row interface{}
		err := results.Row(&row)
		if err != nil {
			return err
		}
		rows = append(rows, row)
	}

	err = results.Err()
	if err != nil {
		return err
	}
	*val = rows
	return nil
}

// Get retrieves the document using its key.
func (db *DB) Get(ctx context.Context, key string, model interface{}) error {

	// Performs a fetch operation against the collection.
	getResult, err := db.Collection.Get(key, &gocb.GetOptions{})
	if errors.Is(err, gocb.ErrDocumentNotFound) {
		return ErrDocumentNotFound
	}
	if err != nil {
		return err
	}

	// Assigns the value of the result into the valuePtr using default decoding.
	err = getResult.Content(&model)
	if err != nil {
		return err
	}

	return nil
}

// Create saves a new document into the collection.
func (db *DB) Create(ctx context.Context, key string, val interface{}) error {
	_, err := db.Collection.Insert(key, val, &gocb.InsertOptions{})
	return err
}

// Update updates a document in the collection.
func (db *DB) Update(ctx context.Context, key string, val interface{}) error {
	_, err := db.Collection.Replace(key, val, &gocb.ReplaceOptions{})
	return err
}

// Patch performs a sub document in the collection. Sub documents operations
// may be quicker and more network-efficient than full-document operations.
func (db *DB) Patch(ctx context.Context, key string, path string,
	val interface{}) error {

	mops := []gocb.MutateInSpec{
		gocb.UpsertSpec(path, val, &gocb.UpsertSpecOptions{}),
	}
	_, err := db.Collection.MutateIn(key, mops,
		&gocb.MutateInOptions{Timeout: 10050 * time.Millisecond})
	return err
}

// Delete removes a document from the collection.
func (db *DB) Delete(ctx context.Context, key string) error {
	_, err := db.Collection.Remove(key, &gocb.RemoveOptions{})
	return err
}

// Count retrieves the total number of documents.
func (db *DB) Count(ctx context.Context, statement string,
	args map[string]interface{}, val *int) error {

	results, err := db.Cluster.Query(statement, &gocb.QueryOptions{
		NamedParameters: args, Adhoc: true})
	if err != nil {
		return err
	}

	var row float64
	err = results.One(&row)
	if err != nil {
		return err
	}

	count := int(row)
	*val = count
	return nil
}

// Lookup query the document for certain path(s); these path(s) are then returned.
func (db *DB) Lookup(ctx context.Context, key string, paths []string,
	val interface{}) error {

	ops := []gocb.LookupInSpec{}
	getSpecOptions := gocb.GetSpecOptions{}

	for _, path := range paths {
		ops = append(ops, gocb.GetSpec(path, &getSpecOptions))
	}
	getResult, err := db.Collection.LookupIn(key, ops, &gocb.LookupInOptions{})
	if err != nil {
		return err
	}

	for i, path := range paths {
		var content interface{}
		err = getResult.ContentAt(uint(i), &content)
		if err != nil {
			return err
		}

		m := make(map[string]interface{})
		keys := strings.Split(path, ".")
		if len(keys) > 0 {
			m[keys[len(keys)-1]] = content
			for idx := len(keys) - 2; idx >= 0; idx-- {
				mn := make(map[string]interface{})
				mn[keys[idx]] = m
				m = mn
			}
		} else {
			m[path] = content
		}

		x, err := json.Marshal(m)
		if err != nil {
			return err
		}
		err = json.Unmarshal(x, &val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) Search(ctx context.Context, stringQuery string, page uint32, perPage uint32, sortBy string, order string, val *interface{}, totalHits *uint64) error {

	query, err := gen.Generate(stringQuery,
		gen.Config{
			"first_seen": {
				Type: gen.DATE,
			},
			"fs": {
				Type:  gen.DATE,
				Field: "first_seen",
			},
			"last_scanned": {
				Type: gen.DATE,
			},
			"ls": {
				Type:  gen.DATE,
				Field: "last_scanned",
			},
			"extension": {
				Field: "file_extension",
			},
			"type": {
				Field: "file_format",
			},
			"size": {
				Type: gen.NUMBER,
			},
			"name": {
				Field: "submissions.filename",
			},
			"positives": {
				Type:  gen.NUMBER,
				Field: "multiav.last_scan.stats.positives",
			},
			"avast": {
				Field: "multiav.last_scan.detections.avast.output",
			},
			"avira": {
				Field: "multiav.last_scan.detections.avira.output",
			},
			"bitdefender": {
				Field: "multiav.last_scan.detections.bitdefender.output",
			},
			"clamav": {
				Field: "multiav.last_scan.detections.clamav.output",
			},
			"comodo": {
				Field: "multiav.last_scan.detections.comodo.output",
			},
			"drweb": {
				Field: "multiav.last_scan.detections.drweb.output",
			},
			"eset": {
				Field: "multiav.last_scan.detections.eset.output",
			},
			"kaspersky": {
				Field: "multiav.last_scan.detections.kaspersky.output",
			},
			"mcafee": {
				Field: "multiav.last_scan.detections.mcafee.output",
			},
			"sophos": {
				Field: "multiav.last_scan.detections.sophos.output",
			},
			"symantec": {
				Field: "multiav.last_scan.detections.symantec.output",
			},
			"trendmicro": {
				Field: "multiav.last_scan.detections.trendmicro.output",
			},
			"windefender": {
				Field: "multiav.last_scan.detections.windefender.output",
			},
			"fsecure": {
				Field: "multiav.last_scan.detections.fsecure.output",
			},
			"engines": {
				FieldGroup: []string{
					"multiav.last_scan.detections.avast.output",
					"multiav.last_scan.detections.avira.output",
					"multiav.last_scan.detections.bitdefender.output",
					"multiav.last_scan.detections.clamav.output",
					"multiav.last_scan.detections.comodo.output",
					"multiav.last_scan.detections.drweb.output",
					"multiav.last_scan.detections.eset.output",
					"multiav.last_scan.detections.kaspersky.output",
					"multiav.last_scan.detections.mcafee.output",
					"multiav.last_scan.detections.sophos.output",
					"multiav.last_scan.detections.symantec.output",
					"multiav.last_scan.detections.trendmicro.output",
					"multiav.last_scan.detections.windefender.output",
					"multiav.last_scan.detections.fsecure.output",
				},
			},
		},
	)
	if err != nil {
		return err
	}

	searchOptions := gocb.SearchOptions{
		Limit: perPage,
		Skip:  perPage * (page - 1),
		Fields: []string{"size", "file_extension", "file_format", "first_seen", "last_scanned", "tags.packer", "tags.pe",
			"tags.avira", "tags.eset", "tags.windefender", "submissions.filename", "classification",
			"multiav.last_scan.stats.positives", "multiav.last_scan.stats.engines_count",
		},
	}

	if sortBy != "" {
		searchOptions.Sort =
			[]search.Sort{search.NewSearchSortField(sortBy).Descending(order == "desc" || order == "")}
	}

	result, err := db.Cluster.SearchQuery(
		db.FTSIndexName, query,
		&searchOptions,
	)
	if err != nil {
		return err
	}

	rows := []interface{}{}
	for result.Next() {
		row := result.Row()
		docID := row.ID

		var fields map[string]interface{}
		err := row.Fields(&fields)
		if err != nil {
			return err
		}
		fields["id"] = docID
		fields["class"] = fields["classification"]
		fields["name"] = fields["submissions.filename"]
		fields["multiav"] = map[string]interface{}{
			"hits":  fields["multiav.last_scan.stats.positives"],
			"total": fields["multiav.last_scan.stats.engines_count"],
		}
		delete(fields, "multiav.last_scan.stats.positives")
		delete(fields, "multiav.last_scan.stats.engines_count")
		delete(fields, "submissions.filename")
		delete(fields, "classification")
		unflattenFields(fields)
		rows = append(rows, fields)
	}

	meta, err := result.MetaData()
	if err != nil {
		return err
	}

	*totalHits = meta.Metrics.TotalRows

	// always check for errors after iterating
	err = result.Err()
	if err != nil {
		return err
	}

	*val = rows
	return nil

}
