// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon, tsfio and tserr
import (
	"os"            // os for file and directory operations
	"path/filepath" // filepath for constructing file paths
	"testing"       // testing for writing test cases
	"time"

	"github.com/thorsphere/tsecon" // tsecon for the package being tested
	"github.com/thorsphere/tserr"  // tserr for custom error handling
	"github.com/thorsphere/tsfio"  // tsfio for file I/O utilities
)

// The testcases use these tokens
const (
	testprefix string = "tsecon_*"  // mostly used as prefix for temporary files or directories
	testdbname string = "events.db" // the name of the SQLite database file used in tests
)

// tmpDB creates a new SQLite database file in the specified temporary directory with the name testdbname.
// It also tests the creation of a new SQLiteEventRepository with the temporary database file.
// In case of an error during the creation of the repository, the execution stops.
func tmpDB(t *testing.T) (*tsecon.SQLiteEventRepository, tsfio.Filename) {
	// Panic if t is nil
	if t == nil {
		panic("nil pointer")
	}
	// Create the temporary directory
	dn, err := os.MkdirTemp("", testprefix)
	// Stop execution in case of an error
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "create temp dir", Fn: dn, Err: err}))
	}
	// Create the filename for the SQLite database in the temporary directory
	fn := tsfio.Filename(filepath.Join(dn, testdbname))
	// Test creating a new SQLiteEventRepository with the temporary directory
	repo, err := tsecon.NewSQLiteEventRepository(fn)
	// Stop execution in case of an error
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "NewSQLiteEventRepository", Fn: string(fn), Err: err}))
	}
	// Check if the repository was created successfully
	if repo == nil {
		t.Fatal(tserr.NilPtr())
	}
	// Return the filename of the SQLite database
	return repo, fn
}

func rmDB(t *testing.T, repo *tsecon.SQLiteEventRepository, fn tsfio.Filename) {
	if err := repo.Close(); err != nil {
		// Stop execution in case of an error
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Close", Fn: string(fn), Err: err}))
	}
	// Retrieve the directory name from the filename
	dn := tsfio.Directory(filepath.Dir(string(fn)))
	// Clean up the temporary directory
	if err := tsfio.RemoveDir(dn); err != nil {
		// Stop execution in case of an error
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "RemoveDir", Fn: string(dn), Err: err}))
	}
}

// TestCloseNil tests the Close method of the SQLiteEventRepository when called on a nil pointer.
// It expects an error to be returned, and if no error is returned, the test fails.
func TestCloseNil1(t *testing.T) {
	// Test closing a nil SQLiteEventRepository
	var repo *tsecon.SQLiteEventRepository = nil
	// Attempt to close the nil repository and check for the expected error
	err := repo.Close()
	// Check if the error is not nil
	if err == nil {
		// If no error is returned, the test fails with a message indicating that a nil pointer was expected to fail.
		t.Fatal(tserr.NilFailed("Close"))
	}
}

// TestGetByDateNil1 tests the GetByDate method of the SQLiteEventRepository when called on a nil pointer.
// It expects an error to be returned, and if no error is returned, the test fails.
func TestGetByDateNil1(t *testing.T) {
	// Test retrieving events by date from a nil SQLiteEventRepository
	var repo *tsecon.SQLiteEventRepository = nil
	// Attempt to retrieve events by date from the nil repository and check for the expected error
	_, err := repo.GetByDate(time.Now().UTC())
	// Check if the error is not nil
	if err == nil {
		// If no error is returned, the test fails with a message indicating that a nil pointer was expected to fail.
		t.Fatal(tserr.NilFailed("GetByDate"))
	}
}

// TestCloseNil2 tests the Close method of the SQLiteEventRepository
// when called on a non-nil pointer that has not been initialized.
// It expects an error to be returned, and if no error is returned, the test fails.
func TestCloseNil2(t *testing.T) {
	// Test closing a nil SQLiteEventRepository
	var repo *tsecon.SQLiteEventRepository = &tsecon.SQLiteEventRepository{}
	// Attempt to close the nil repository and check for the expected error
	err := repo.Close()
	// Check if the error is not nil
	if err == nil {
		// If no error is returned, the test fails with a message indicating that a nil pointer was expected to fail.
		t.Fatal(tserr.NilFailed("Close"))
	}
}

// TestGetByDateNil2 tests the GetByDate method of the SQLiteEventRepository
// when called on a non-nil pointer that has not been initialized.
// It expects an error to be returned, and if no error is returned, the test fails.
func TestGetByDateNil2(t *testing.T) {
	// Test retrieving events by date from a nil SQLiteEventRepository
	var repo *tsecon.SQLiteEventRepository = &tsecon.SQLiteEventRepository{}
	// Attempt to retrieve events by date from the nil repository and check for the expected error
	_, err := repo.GetByDate(time.Now().UTC())
	// Check if the error is not nil
	if err == nil {
		// If no error is returned, the test fails with a message indicating that a nil pointer was expected to fail.
		t.Fatal(tserr.NilFailed("GetByDate"))
	}
}

// TestStoreNil1 tests the Store method of the SQLiteEventRepository when called on a nil pointer.
// It expects an error to be returned, and if no error is returned, the test fails.
func TestStoreNil1(t *testing.T) {
	// Test storing an event in a nil SQLiteEventRepository
	var repo *tsecon.SQLiteEventRepository = nil
	// Attempt to store an event in the nil repository and check for the expected error
	if err := repo.Store(&evNfp); err == nil {
		// If no error is returned, the test fails with a message indicating that a nil pointer was expected to fail.
		t.Fatal(tserr.NilFailed("Store"))
	}
}

// TestStoreNil2 tests the Store method of the SQLiteEventRepository when called on a db nil pointer.
// It expects an error to be returned, and if no error is returned, the test fails.
func TestStoreNil2(t *testing.T) {
	// Test storing an event in a nil SQLiteEventRepository
	var repo *tsecon.SQLiteEventRepository = &tsecon.SQLiteEventRepository{}
	// Attempt to store an event in the repository with a db nil pointer and check for the expected error
	if err := repo.Store(&evNfp); err == nil {
		// If no error is returned, the test fails with a message indicating that a nil pointer was expected to fail.
		t.Fatal(tserr.NilFailed("Store"))
	}
}

// TestStoreNil3 tests the Store method of the SQLiteEventRepository when called with a nil event.
// It expects an error to be returned, and if no error is returned, the test fails.
func TestStoreNil3(t *testing.T) {
	// Create a new SQLiteEventRepository with the temporary directory
	repo, fn := tmpDB(t)
	// Attempt to store a nil event in the repository and check for the expected error
	if err := repo.Store(nil); err == nil {
		// If no error is returned, the test fails with a message indicating that a nil pointer was expected to fail.
		t.Fatal(tserr.NilFailed("Store"))
	}
	// Clean up the temporary database and directory
	rmDB(t, repo, fn)
}

// TestNewSQLiteEventRepository tests the NewSQLiteEventRepository function of the tsecon package.
// It creates a temporary directory, creates a new SQLiteEventRepository with the temporary directory,
// checks if the repository was created successfully, and then closes the repository and
// cleans up the temporary directory. In case of any error during these steps,
// the execution stops and an error message is logged.
func TestNewSQLiteEventRepository(t *testing.T) {
	// Create a new SQLiteEventRepository with the temporary directory
	repo, fn := tmpDB(t)
	// Remove the temporary database and directory
	rmDB(t, repo, fn)
}

// TestStoreAndGetByDate tests the Store and GetByDate methods of the SQLiteEventRepository.
// It creates a temporary database, stores a sample event, retrieves events by the date of the stored event,
// checks if the retrieved events match the stored event, and then cleans up the temporary database and directory.
// In case of any error during these steps, the execution stops and an error message is logged.
func TestStoreAndGetByDate(t *testing.T) {
	// Create a new SQLiteEventRepository with the temporary directory
	repo, fn := tmpDB(t)
	// Define a sample event for testing purposes
	ev := evGdp24
	// Store a sample event in the repository
	if err := repo.Store(&ev); err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Store", Fn: string(fn), Err: err}))
	}
	// Retrieve events by the date of the stored event
	events, err := repo.GetByDate(ev.Time.UTC())
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "GetByDate", Fn: string(fn), Err: err}))
	}
	// Check if the retrieved events match the stored event
	if len(events) != 1 {
		t.Fatal(tserr.Equal(&tserr.EqualArgs{
			Var:    "len(events)",
			Actual: int64(len(events)),
			Want:   int64(1),
		}))
	}
	// Compare the retrieved event with the original event and check if they are equal
	if !ev.Equal(events[0]) {
		t.Fatal(tserr.EqualStr(&tserr.EqualStrArgs{
			Var:    "events[0]",
			Actual: events[0].String(),
			Want:   ev.String(),
		}))
	}
	// Clean up the temporary database and directory
	rmDB(t, repo, fn)
}
