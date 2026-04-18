// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon, tsfio and tserr
import (
	"os"            // os for file and directory operations
	"path/filepath" // filepath for constructing file paths
	"testing"       // testing for writing test cases

	"github.com/thorsphere/tsecon" // tsecon for the package being tested
	"github.com/thorsphere/tserr"  // tserr for custom error handling
	"github.com/thorsphere/tsfio"  // tsfio for file I/O utilities
)

// The testcases use these tokens
const (
	testprefix string = "tsecon_*"  // mostly used as prefix for temporary files or directories
	testdbname string = "events.db" // the name of the SQLite database file used in tests
)

// tmpDir creates a new temporary directory in the default directory for temporary files
// with the prefix testprefix and a random string to the end. In case of an error
// the execution stops.
func tmpDir(t *testing.T) tsfio.Directory {
	// Panic if t is nil
	if t == nil {
		panic("nil pointer")
	}
	// Create the temporary directory
	d, err := os.MkdirTemp("", testprefix)
	// Stop execution in case of an error
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "create temp dir", Fn: d, Err: err}))
	}
	// Return the temporary Directory
	return tsfio.Directory(d)
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

// TestNewSQLiteEventRepository tests the NewSQLiteEventRepository function of the tsecon package.
// It creates a temporary directory, creates a new SQLiteEventRepository with the temporary directory,
// checks if the repository was created successfully, and then closes the repository and
// cleans up the temporary directory. In case of any error during these steps,
// the execution stops and an error message is logged.
func TestNewSQLiteEventRepository(t *testing.T) {
	// Create a temporary directory for testing the SQLiteEventRepository
	dn := tmpDir(t)
	// Create the filename for the SQLite database in the temporary directory
	fn := tsfio.Filename(filepath.Join(string(dn), testdbname))
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
	// Close the repository
	if err := repo.Close(); err != nil {
		// Stop execution in case of an error
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Close", Fn: string(fn), Err: err}))
	}
	// Clean up the temporary directory
	if err := tsfio.RemoveDir(dn); err != nil {
		// Stop execution in case of an error
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "RemoveDir", Fn: string(dn), Err: err}))
	}
}
