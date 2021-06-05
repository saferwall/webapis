package db

import (
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
	err = mgr.CreatePrimaryIndex(bucketName, &gocb.CreatePrimaryQueryIndexOptions{
		IgnoreIfExists: true,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		Bucket:     bucket,
		Cluster:    cluster,
		Collection: collection,
	}, nil
}
