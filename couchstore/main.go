package couchstore

import (
	"reflect"

	geojson "github.com/paulmach/go.geojson"
	"github.com/pkg/errors"
	gocb "github.com/couchbase/gocb"
)

// CouchServer is an interface that any struct can implement to act as
// a couchbase database connection.
type CouchServer interface {
	Close()
	GetDocument(string, interface{}) error
	GetDocuments(*gocb.N1qlQuery, interface{}) error
	UpsertDocument(string, interface{}, uint32) error
	DeleteDocument(string) error
	DeleteDocuments(gocb.N1qlQuery) error
	ExecuteSpatialQuery(string, string, []float64) ([]SpatialDocument, error)
}

// CouchConfig is a structure that represents the allowed values for getting a bucket.
type CouchConfig struct {
	ConnectionString string
	BucketName       string
	BucketPassword   string
}

// Interface that piggybacks the gocb internal errors.
// example: http://stackoverflow.com/questions/30931951/how-check-error-codes-with-couchbase-gocb
type gocbError interface {
	KeyNotFound() bool
}

// gocbKeyNotFound is a helper function that returns true if the error returned from
// gocb is a key not found error.
func gocbKeyNotFound(err error) bool {
	if se, ok := err.(gocbError); ok {
		return se.KeyNotFound()
	}

	return false
}

// DBServer is the implementation of the couchstore interface.
type DBServer struct {
	bucket *gocb.Bucket
}

// NewDBServer returns a reference to the configured server.
func NewDBServer(config CouchConfig) (CouchServer, error) {
	// Validate that the password is something.
	if config.BucketPassword == "" {
		return nil, errors.Wrap(errors.New("password is required to access couchbase buckets"), "config.BucketPassword")
	}

	// Connect to the specified couchbase cluster.
	cluster, err := gocb.Connect(config.ConnectionString)
	if err != nil {
		return nil, errors.Wrap(err, "gocb.Connect")
	}

	// Create a new server object.
	server := DBServer{}

	bucket, err := cluster.OpenBucket(config.BucketName, config.BucketPassword)
	if err != nil {
		return nil, errors.Wrap(err, "cluster.OpenBucket")
	}

	// Assign the bucket to the server.
	server.bucket = bucket

	return &server, nil
}

// Close cleans up the database connection by closeing the bucket.  We will
// ignore the error since we expect this to be used with a defer statement.
func (db *DBServer) Close() {
	_ = db.bucket.Close()
}

// GetDocument returns a document for the provided key.
func (db *DBServer) GetDocument(key string, result interface{}) error {
	// Code copied from MGO code at http://bazaar.launchpad.net/+branch/mgo/v2/view/head:/session.go#L2769
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr {
		panic("result argument must be an address")
	}

	_, err := db.bucket.Get(key, result)
	if err != nil {
		if gocbKeyNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "bucket.Get")
	}

	return nil
}

// GetDocuments returns a list of documents for the provided query.  The results need
// to be cast to the expected return type.
func (db *DBServer) GetDocuments(query *gocb.N1qlQuery, result interface{}) error {
	// Code copied from MGO code at http://bazaar.launchpad.net/+branch/mgo/v2/view/head:/session.go#L2769
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		panic("result argument must be a slice address")
	}

	// Execute the query against the database.
	resultSet, err := db.bucket.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return errors.Wrap(err, "bucket.ExecuteN1q1Query")
	}

	// Ensure that the result set is closed.
	defer resultSet.Close()

	// Code copied from MGO code at http://bazaar.launchpad.net/+branch/mgo/v2/view/head:/session.go#L2769
	slicev := resultv.Elem()
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0
	for {
		if slicev.Len() == i {
			elemp := reflect.New(elemt)
			if !resultSet.Next(elemp.Interface()) {
				break
			}
			slicev = reflect.Append(slicev, elemp.Elem())
			slicev = slicev.Slice(0, slicev.Cap())
		} else {
			if !resultSet.Next(slicev.Index(i).Addr().Interface()) {
				break
			}
		}
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))

	return nil
}

// UpsertDocument performs an insert/update on the specified key with the provided data.
// Note that updates will completely replace the document, so all existing data needs to be provided.
func (db *DBServer) UpsertDocument(key string, document interface{}, expires uint32) error {
	// Perform an upsert with the provided values.
	_, err := db.bucket.Upsert(key, document, expires)
	if err != nil {
		return errors.Wrap(err, "bucket.Upsert")
	}

	return nil
}

// DeleteDocument deletes the document with the provided key.
func (db *DBServer) DeleteDocument(key string) error {
	return nil
}

// DeleteDocuments deletes all documents that match the provided query.
func (db *DBServer) DeleteDocuments(query gocb.N1qlQuery) error {
	return nil
}

// ExecuteSpatialQuery performs a bounding box query with the provided paramters.
func (db *DBServer) ExecuteSpatialQuery(design, view string, bounds []float64) ([]SpatialDocument, error) {
	// Create a new spatial query using the calculated bounding box.
	spatialQuery := gocb.NewSpatialQuery(design, view).Bbox(bounds)

	// Execute the spatial query against the database.
	resultSet, err := db.bucket.ExecuteSpatialQuery(spatialQuery)
	if err != nil {
		return nil, errors.Wrap(err, "bucket.ExecuteSpatialQuery")
	}

	// Ensure that the result set is closed.
	defer resultSet.Close()

	// List of documents that will be returned to the caller.
	documents := []SpatialDocument{}

	var document spatialViewResult
	for resultSet.Next(&document) {
		// Assign the value to prevent array pointer issues.
		d := document.SpatialDocument

		documents = append(documents, d)
	}

	return documents, nil
}

// viewResultDocument is the structure of documents returned by view queries.
type spatialViewResult struct {
	Key interface{}
	ID  string
	SpatialDocument
}

// SpatialDocument is a special structure for data returned from spatial queries.
type SpatialDocument struct {
	Value    interface{}
	Geometry geojson.Geometry
}
