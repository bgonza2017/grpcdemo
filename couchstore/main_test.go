package couchstore

import (
	"os"
	"testing"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (

	// Store a package level reference to the database connection.
	dbServer CouchServer
)


// Feature represents a geojson feature.
type Feature struct {
	Type     string   `xml:"-" csv:"-" json:"type,omitempty"`
	Geometry Geometry `xml:"-" csv:"-" json:"geometry,omitempty"`
}

// Grid represents a grid location with its boundary data.
type Grid struct {
	GridName string  `xml:"-" csv:"-" json:"gridName,omitempty"`
	Boundary Feature `xml:"-" csv:"-" json:"boundary,omitempty"`
}

// Geometry is the actual geometry for a geojson object.
type Geometry struct {
	Type        string      `xml:"-" csv:"-" json:"type,omitempty"`
	Coordinates interface{} `xml:"-" csv:"-" json:"coordinates,omitempty"`
}

// DeliveryTradeArea represents the store data that contains boundary data and
// other location based data.
type DeliveryTradeArea struct {
	Type        string  `xml:"-" csv:"-" json:"type,omitempty"` // Database record type.
	StoreNumber string  `xml:"-" csv:"-" json:"storeNumber,omitempty"`
	Enabled     *bool   `xml:"-" csv:"-" json:"enabled,omitempty"`
	UpdatedAt   int64   `xml:"-" csv:"-" json:"updatedAt,omitempty"`
	UpdatedBy   string  `xml:"-" csv:"-" json:"updatedBy,omitempty"`
	Boundary    Feature `xml:"-" csv:"-" json:"boundary,omitempty"`
	Grids       []Grid  `xml:"-" csv:"-" json:"grids,omitempty"`
}

// newDBServer is a helper function to create a database connection with the provided config.
// NOTE - a defer close should be used at the end of each test.
func newDBServer(config CouchConfig) error {
	var err error

	// Setup the db server for the testing.
	dbServer, err = NewDBServer(config)
	if err != nil {
		return errors.Wrap(err, "unable to connect to database")
	}

	return nil
}

func TestStart(t *testing.T) {
	// List of invalid configs to test against.
	flagConfigs := []map[string]string{
		// No values passed.
		{"CBPASSWORD": "", "cbconnstr": "", "cbbucket": ""},

		// Invalid connection string.
		{"CBPASSWORD": "p8ssw0rd", "cbconnstr": "httz://invalid.com", "cbbucket": "store"},

		// Invalid bucket name.
		{"CBPASSWORD": "p8ssw0rd", "cbconnstr": "localhost", "cbbucket": "invalid"},

		// Invalid bucket name.
		{"CBPASSWORD": "p8ssw0rd", "cbconnstr": "localhost", "cbbucket": "invalid"},

		// Invalid bucket password.
		{"CBPASSWORD": "invalid", "cbconnstr": "localhost", "cbbucket": "store"},
	}

	for _, config := range flagConfigs {
		os.Setenv("CBPASSWORD", config["CBPASSWORD"])
		viper.Set("cbconnstr", config["cbconnstr"])
		viper.Set("cbbucket", config["cbbucket"])

	}

	// Test a valid config to ensure the database connection works.
	os.Setenv("CBPASSWORD", "p8ssw0rd")
	viper.Set("cbconnstr", "localhost")
	viper.Set("cbbucket", "store")
	viper.Set("isTesting", true)

}

func TestCouchBase(t *testing.T) {

	var err error

	// Create a couchstore config.
	cbConfig := CouchConfig{
		ConnectionString: "localhost",
		BucketName:       "store",
		BucketPassword:   "p8ssw0rd",
	}

	// Using the provided config get a reference to the couchbase server.
	dbServer, err = NewDBServer(cbConfig)
	if err != nil {
		fmt.Printf("========= cb error ===========\n")
		return
	}

		fmt.Printf("========= cb connected ===========\n")
	// Ensure that the server connection is cleaned up.
	defer dbServer.Close()

	number := "005050"

	// Build the record key based on the store number provided.
	key := fmt.Sprintf("%s:%s", "sg", number)

	// Query the document from the database.
	var document DeliveryTradeArea
	err = dbServer.GetDocument(key, &document)	

	if err != nil {
		fmt.Printf("========= dbServer GetDocument error ===========\n")
		return
	}

	fmt.Printf("%v", document)
	fmt.Printf("\n\n========= dbServer GetDocument success ===========\n")
}

