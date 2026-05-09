// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon, tserr and tsfio
import (
	"context"       // context for managing request-scoped values and cancellation
	"os"            // os for file and directory operations
	"path/filepath" // filepath for constructing file paths
	"testing"       // testing for writing test cases
	"time"          // time for working with time and dates

	"github.com/thorsphere/tsecon" // tsecon for the package being tested
	"github.com/thorsphere/tserr"  // tserr for custom error handling
	"github.com/thorsphere/tsfio"  // tsfio for file input/output operations, including handling golden files
)

// The testcases use these tokens
const (
	testprefix string = "tsecon_*"  // mostly used as prefix for temporary files or directories
	testdbname string = "events.db" // the name of the SQLite database file used in tests
)

var (
	// Define some sample events for testing purposes
	evNfp tsecon.Event = tsecon.Event{
		Name:     "Non-Farm Payrolls",
		Time:     time.Date(2024, 7, 5, 8, 30, 0, 0, time.UTC),
		Country:  "US",
		Actual:   new(200.0),
		Estimate: new(180.0),
		Previous: new(150.0),
		Unit:     "K",
		Impact:   tsecon.ImpactHigh,
		Source:   "Bureau of Labor Statistics",
	}
	evGdp24 tsecon.Event = tsecon.Event{
		Name:     "GDP Growth Rate",
		Time:     time.Date(2024, 7, 10, 8, 30, 0, 0, time.UTC),
		Country:  "US",
		Actual:   new(3.5),
		Estimate: new(3.0),
		Previous: new(2.8),
		Unit:     "%",
		Impact:   tsecon.ImpactMedium,
		Source:   "Bureau of Economic Analysis",
	}
	evGdp30 tsecon.Event = tsecon.Event{
		Name:     "GDP Growth Rate",
		Time:     time.Date(2030, 7, 10, 8, 30, 0, 0, time.UTC),
		Country:  "US",
		Actual:   nil,
		Estimate: nil,
		Previous: nil,
		Unit:     "%",
		Impact:   tsecon.ImpactLow,
		Source:   "Bureau of Economic Analysis",
	}
	// Define a slice of events for testing purposes
	evs []tsecon.Event = []tsecon.Event{
		evNfp,
		evGdp24,
		evGdp30,
	}
	// Define a sample period for testing purposes
	per *tsecon.Period = &tsecon.Period{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, 7, 31, 23, 59, 59, 0, time.UTC),
	}
)

// Mock is a mock implementation of the Provider interface for testing purposes.
type Mock struct{}

// GetEvents returns a slice of Event for the specified date range, filtering the events based on the provided period.
func (p *Mock) GetEvents(ctx context.Context, period *tsecon.Period) ([]tsecon.Event, error) {
	// Check if the provider or period is nil, and return an error if so
	if (p == nil) || (period == nil) {
		return nil, tserr.NilPtr()
	}
	// Filter events based on the provided period and return the matching events
	evlist := []tsecon.Event{}
	for _, event := range evs {
		// Check if the event's time is within the specified period, and if so, add it to the list of events to return
		if event.Time.After(period.From) && event.Time.Before(period.To) {
			// If the event is within the period, append it to the list of events to return
			evlist = append(evlist, event)
		}
	}
	// Return the list of events that match the specified period and nil for the error
	return evlist, nil
}

// MockErr is a mock implementation of the Provider interface that returns an error when GetEvents is called.
type MockErr struct{}

// GetEvents returns an error indicating that the operation is forbidden,
// simulating a failure scenario for testing purposes.
func (p *MockErr) GetEvents(ctx context.Context, period *tsecon.Period) ([]tsecon.Event, error) {
	return nil, tserr.Forbidden("MockErr")
}

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
