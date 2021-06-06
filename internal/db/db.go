// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"time"

	gocb "github.com/couchbase/gocb/v2"
)

const (
	// Duration to wait until memd connections have been established with
	// the server and are ready.
	timeout = 5 * time.Second
)

// DB represents the database connection.
type DB struct {
	Bucket     *gocb.Bucket
	Cluster    *gocb.Cluster
	Collection *gocb.Collection
}

// Open opens a connection to the database.
func Open(server, username, password, bucketName string) (*DB, error) {

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
		Bucket:     bucket,
		Cluster:    cluster,
		Collection: collection,
	}, nil
}

// Query executes a N1QL query.
func (db *DB) Query(ctx context.Context, statement string,
	 args map[string]interface{}) (*gocb.QueryResult, error) {

	results, err := db.Cluster.Query(statement, &gocb.QueryOptions{
		NamedParameters: args, Adhoc: true})
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Get retrieves the document using its key.
func (db *DB) Get(ctx context.Context, id string, model interface{}) error {

	// Performs a fetch operation against the collection.
	getResult, err := db.Collection.Get(id, &gocb.GetOptions{})
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
func (db *DB) Create(ctx context.Context, id string, val interface{}) error {
	_, err := db.Collection.Insert(id, val, &gocb.InsertOptions{})
	return err
}

// Update updates a document in the collection.
func (db *DB) Update(ctx context.Context, id string, val interface{}) error {
	_, err := db.Collection.Replace(id, val, &gocb.ReplaceOptions{})
	return err
}

// Delete removes a document from the collection.
func (db *DB) Delete(ctx context.Context, id string) error {
	_, err := db.Collection.Remove(id, &gocb.RemoveOptions{})
	return err
}

// Count retrieves the total number of documents.
func (db *DB) Count(ctx context.Context, docType string,
	 val interface{}) error {

	//val = nil

	params := make(map[string]interface{}, 2)
	params["bucketName"] = db.Bucket.Name()
	params["docType"] = docType

	statement := `SELECT COUNT(*) FROM $bucketname WHERE type=$docType`
	results, err := db.Query(ctx, statement, params)
	if err != nil {
		return err
	}

	var row int
	err = results.One(&row)
	if err != nil {
		return err
	}

	return nil
}
