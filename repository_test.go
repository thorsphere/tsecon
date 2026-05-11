// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon, tsfio and tserr
import (
	"context" // context for handling request contexts
	"testing" // testing for writing test cases
	"time"    // time for handling event timestamps

	"github.com/thorsphere/tsecon" // tsecon for the package being tested
	"github.com/thorsphere/tserr"  // tserr for custom error handling
	"github.com/thorsphere/tsfio"  // tsfio for file I/O utilities
)

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
	_, err := repo.GetByDate(context.Background(), time.Now().UTC())
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
	_, err := repo.GetByDate(context.Background(), time.Now().UTC())
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
	if err := repo.Store(context.Background(), &evNfp); err == nil {
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
	if err := repo.Store(context.Background(), &evNfp); err == nil {
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
	if err := repo.Store(context.Background(), nil); err == nil {
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

// testGetByPeriod tests the GetByPeriod method of the SQLiteEventRepository.
// It creates a temporary database, stores sample events, retrieves events by a defined period,
// compares the retrieved events to a golden file, and then cleans up the temporary database and directory.
// In case of any error during these steps, the execution stops and an error message is logged.
func testGetByPeriod(t *testing.T, p *tsecon.Period) string {
	// Create a new SQLiteEventRepository with the temporary directory
	repo, fn := tmpDB(t)
	// Iterate over each event in the sample events slice and store it in the repository,
	// checking for errors during the storage process.
	for _, ev := range evs {
		// Store each event in the repository and check for errors
		if err := repo.Store(context.Background(), &ev); err != nil {
			t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Store", Fn: string(fn), Err: err}))
		}
	}
	// Retrieve events by the defined period
	events, err := repo.GetByPeriod(context.Background(), p)
	// Check for errors during the retrieval process, and if an error occurs, stop execution and log the error message.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "GetByPeriod", Fn: string(fn), Err: err}))
	}
	// Remove the temporary database and directory
	rmDB(t, repo, fn)
	// Return the formatted string representation of the retrieved events
	return tsecon.PrintEvents(events)
}

// TestGetByPeriod1 tests the GetByPeriod method of the SQLiteEventRepository by comparing the output to a golden file.
func TestGetByPeriod1(t *testing.T) {
	s := testGetByPeriod(t, per)
	// Compare the output to a golden file using the EvalGoldenFile function from the tsfio package,
	// and if there is an error, fail the test with the error message.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "getbyperiod", Data: s}); err != nil {
		t.Fatal(err)
	}
}

// TestGetByPeriod2 tests the GetByPeriod method of the SQLiteEventRepository with a period
// that does not include any events, and expects an empty string as output.
// If the output is not empty, the test fails.
func TestGetByPeriod2(t *testing.T) {
	// Define a period that does not include any of the sample events
	p := &tsecon.Period{
		From: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
	}
	// Call the testGetByPeriod function with the defined period and store the output in variable s
	s := testGetByPeriod(t, p)
	// Check if the output string s is not empty, which would indicate that events were retrieved
	// when none were expected.
	if s != "" {
		t.Fatal(tserr.EqualStr(&tserr.EqualStrArgs{
			Var:    "s",
			Actual: s,
			Want:   "",
		}))
	}
}

// TestGetByPeriod3 tests the GetByPeriod method of the SQLiteEventRepository with a period
// that does not include any events, and expects an empty string as output.
// If the output is not empty, the test fails.
func TestGetByPeriod3(t *testing.T) {
	// Define a period that does not include any of the sample events
	p := &tsecon.Period{
		From: time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC),
	}
	// Call the testGetByPeriod function with the defined period and store the output in variable s
	s := testGetByPeriod(t, p)
	// Check if the output string s is not empty, which would indicate that events were retrieved
	// when none were expected.
	if s != "" {
		t.Fatal(tserr.EqualStr(&tserr.EqualStrArgs{
			Var:    "s",
			Actual: s,
			Want:   "",
		}))
	}
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
	if err := repo.Store(context.Background(), &ev); err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Store", Fn: string(fn), Err: err}))
	}
	// Retrieve events by the date of the stored event
	events, err := repo.GetByDate(context.Background(), ev.Time.UTC())
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
	if !ev.NearEqual(events[0]) {
		t.Fatal(tserr.EqualStr(&tserr.EqualStrArgs{
			Var:    "events[0]",
			Actual: events[0].String(),
			Want:   ev.String(),
		}))
	}
	// Clean up the temporary database and directory
	rmDB(t, repo, fn)
}
