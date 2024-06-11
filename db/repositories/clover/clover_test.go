package repositories_clover

import (
	"fmt"
	"os"

	clover "github.com/ostafen/clover/v2"
)

// setup initializes and sets up the clover database using bbolt under the hood in a temporary dir.
// Additionally, it automatically creates collections for the necessary models.
func setup() (*clover.DB, string) {
	path := tempfile()

	db, err := clover.Open(path)
	if err != nil {
		fmt.Println(err)
		panic("failed to connect to database")
	}

	//Create collections
	db.CreateCollection("peer_info")
	db.CreateCollection("machine")
	db.CreateCollection("free_resources")
	db.CreateCollection("available_resources")
	db.CreateCollection("services")
	db.CreateCollection("service_resource_requirements")
	db.CreateCollection("libp_2_p_info")
	db.CreateCollection("machine_uuid")
	db.CreateCollection("connection")
	db.CreateCollection("elastic_token")
	db.CreateCollection("log_bin_auth")
	db.CreateCollection("deployment_request_flat")
	db.CreateCollection("request_tracker")
	db.CreateCollection("virtual_machine")

	return db, path
}

// teardown closes the GORM database connection after tests.
func teardown(db *clover.DB, path string) {
	// close the clover database
	db.Close()
	os.RemoveAll(path)
}

// tempfile returns a temporary file path.
func tempfile() string {
	dir, err := os.MkdirTemp("", "nunet-test-")
	if err != nil {
		panic(err)
	}
	return dir
}
