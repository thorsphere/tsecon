// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tseventserver_test

// Import standard library packages and third-party packages for context management, testing, and custom error handling.
import (
	"context" // context for managing request contexts and timeouts
	"os"      // os for accessing environment variables to determine Firestore configuration
	"testing" // testing for writing unit tests in Go
	"time"    // time for working with time and dates in tests

	// lpstats for working with statistics
	"github.com/thorsphere/lpstats"
	"github.com/thorsphere/tserr"         // tserr for custom error handling in tests, allowing for better error messages and context when tests fail.
	"github.com/thorsphere/tseventserver" // tseventserver is the package being tested, which contains the FirestoreEventRepository implementation.
	"github.com/thorsphere/tstrading"     // tstrading for economic events
)

// tip is the help text shown when Firestore connectivity isn't set up.
const tip = `Firestore emulator not found.

To run these tests, you need a Firestore emulator. Start it with Docker:

    docker run -d --name firestore-emulator \
        -p 8081:8081 \
        google/cloud-sdk:emulators \
        gcloud beta emulators firestore start --host-port="0.0.0.0:8081"

Then set the environment variable and run the tests:

    export FIRESTORE_EMULATOR_HOST=localhost:8081
    go test -v ./...

Or use the convenience script:

    ./test_firestore.sh

Alternatively, to test against a real Firestore project, set GOOGLE_CLOUD_PROJECT
and ensure your application default credentials are available.
`

var (
	// Define some sample events for testing purposes
	evNfp *tstrading.Event = &tstrading.Event{
		Name:        "Non-Farm Payrolls",
		Description: "The non-farm payrolls index measures the payrolls of non-farm payrolls.",
		Time:        time.Date(2024, 7, 5, 8, 30, 0, 0, time.UTC),
		Country:     "US",
		Currency:    nil,
		Actual:      lpstats.PtrFloat(200.0),
		Estimate:    lpstats.PtrFloat(180.0),
		Previous:    lpstats.PtrFloat(150.0),
		Unit:        lpstats.PtrStr("K"),
		Precision:   0,
		Change:      lpstats.PtrFloat(50.0),
		ChangePct:   lpstats.PtrFloat(33.3),
		Surprise:    lpstats.PtrFloat(-20.0),
		SurprisePct: lpstats.PtrFloat(-11.1),
		Impact:      tstrading.ImpactHigh,
		Source:      "Bureau of Labor Statistics",
	}
	evCc *tstrading.Event = &tstrading.Event{
		Name:        "Consumer Credit",
		Description: "Consumer credit measures the credit of consumers.",
		Time:        time.Date(2026, 6, 5, 8, 30, 0, 0, time.UTC),
		Country:     "US",
		Currency:    lpstats.PtrStr("USD"),
		Actual:      lpstats.PtrFloat(20.73),
		Estimate:    lpstats.PtrFloat(17.80),
		Previous:    lpstats.PtrFloat(22.23),
		Unit:        lpstats.PtrStr("B"),
		Precision:   2,
		Change:      lpstats.PtrFloat(-1.5),
		ChangePct:   lpstats.PtrFloat(-6.7),
		Surprise:    lpstats.PtrFloat(-2.9),
		SurprisePct: lpstats.PtrFloat(-16.3),
		Impact:      tstrading.ImpactHigh,
		Source:      "Bureau of Labor Statistics",
	}
	evGdp24 *tstrading.Event = &tstrading.Event{
		Name:        "GDP Growth Rate",
		Description: "GDP growth rate measures the growth rate of the gross domestic product (GDP).",
		Time:        time.Date(2024, 7, 10, 8, 30, 0, 0, time.UTC),
		Country:     "US",
		Currency:    nil,
		Actual:      lpstats.PtrFloat(3.5),
		Estimate:    lpstats.PtrFloat(3.0),
		Previous:    lpstats.PtrFloat(2.8),
		Unit:        lpstats.PtrStr("%"),
		Precision:   1,
		Change:      lpstats.PtrFloat(0.7),
		ChangePct:   lpstats.PtrFloat(25.0),
		Surprise:    lpstats.PtrFloat(-0.5),
		SurprisePct: lpstats.PtrFloat(-16.7),
		Impact:      tstrading.ImpactMedium,
		Source:      "Bureau of Economic Analysis",
	}
	evGdp30 *tstrading.Event = &tstrading.Event{
		Name:        "GDP Growth Rate",
		Description: "GDP growth rate measures the growth rate of the gross domestic product (GDP).",
		Time:        time.Date(2030, 7, 10, 8, 30, 0, 0, time.UTC),
		Country:     "US",
		Actual:      nil,
		Estimate:    nil,
		Previous:    nil,
		Unit:        lpstats.PtrStr("%"),
		Precision:   1,
		Change:      nil,
		ChangePct:   nil,
		Surprise:    nil,
		SurprisePct: nil,
		Impact:      tstrading.ImpactLow,
		Source:      "Bureau of Economic Analysis",
	}
	// Define a slice of events for testing purposes
	evs []*tstrading.Event = []*tstrading.Event{
		evNfp,
		evCc,
		evGdp24,
		evGdp30,
	}
	// Define a sample period for testing purposes
	per *tstrading.Period = &tstrading.Period{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, 7, 31, 23, 59, 59, 0, time.UTC),
	}
)

// setupFirestore is a helper function that initializes a FirestoreEventRepository for testing purposes.
func setupFirestore(t *testing.T) (*tseventserver.FirestoreEventRepository, func()) {
	// Check environment variables to determine if we should run Firestore tests against an emulator or a real Firestore instance.
	emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	// If neither the emulator host nor the project ID is set, skip the Firestore tests to avoid failures due to missing configuration.
	if emulatorHost == "" && projectID == "" {
		t.Fatal(tip)
	}
	// If the emulator host is set, but project ID is not, we can default to a dummy value
	// since the emulator does not require a real project ID.
	if emulatorHost != "" && projectID == "" {
		projectID = "demo-project"
	}
	// Create a context for initializing the FirestoreEventRepository, which will be used to manage timeouts and cancellation if needed.
	ctx := context.Background()
	// Create a new instance of FirestoreEventRepository using the setup context and project ID.
	repo, err := tseventserver.NewFirestoreEventRepository(ctx, projectID)
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "NewFirestoreEventRepository", Fn: projectID, Err: err}))
	}
	// Define a cleanup function that will be called at the end of the test to close the FirestoreEventRepository and release any resources.
	cleanup := func() {
		repo.Close()
	}
	// Return the initialized FirestoreEventRepository and the cleanup function to be used in the test.
	return repo, cleanup
}

// TestFirestoreCloseNil1 tests the Close method when called on a nil pointer.
func TestFirestoreCloseNil1(t *testing.T) {
	var repo *tseventserver.FirestoreEventRepository = nil
	if err := repo.Close(); err == nil {
		t.Fatal(tserr.NilFailed("Close"))
	}
}

// TestFirestoreCloseNil2 tests the Close method on an uninitialized pointer (nil client).
func TestFirestoreCloseNil2(t *testing.T) {
	var repo *tseventserver.FirestoreEventRepository = &tseventserver.FirestoreEventRepository{}
	if err := repo.Close(); err == nil {
		t.Fatal(tserr.NilFailed("Close"))
	}
}

// TestFirestoreStoreNil1 tests Store on a nil repository pointer.
func TestFirestoreStoreNil1(t *testing.T) {
	var repo *tseventserver.FirestoreEventRepository = nil
	if err := repo.Store(context.Background(), evNfp); err == nil {
		t.Fatal(tserr.NilFailed("Store"))
	}
}

// TestFirestoreStoreNil2 tests Store on an uninitialized repository.
func TestFirestoreStoreNil2(t *testing.T) {
	var repo *tseventserver.FirestoreEventRepository = &tseventserver.FirestoreEventRepository{}
	if err := repo.Store(context.Background(), evNfp); err == nil {
		t.Fatal(tserr.NilFailed("Store"))
	}
}

// TestFirestoreStoreNil3 tests Store when the event is nil.
func TestFirestoreStoreNil3(t *testing.T) {
	var repo *tseventserver.FirestoreEventRepository = &tseventserver.FirestoreEventRepository{}
	if err := repo.Store(context.Background(), nil); err == nil {
		t.Fatal(tserr.NilFailed("Store"))
	}
}

// TestFirestoreGetByPeriodNil1 tests GetByPeriod on a nil repository pointer.
func TestFirestoreGetByPeriodNil1(t *testing.T) {
	var repo *tseventserver.FirestoreEventRepository = nil
	if _, err := repo.GetByPeriod(context.Background(), per); err == nil {
		t.Fatal(tserr.NilFailed("GetByPeriod"))
	}
}

// TestFirestoreGetByPeriodNil2 tests GetByPeriod on an uninitialized repository.
func TestFirestoreGetByPeriodNil2(t *testing.T) {
	var repo *tseventserver.FirestoreEventRepository = &tseventserver.FirestoreEventRepository{}
	if _, err := repo.GetByPeriod(context.Background(), per); err == nil {
		t.Fatal(tserr.NilFailed("GetByPeriod"))
	}
}

// TestFirestoreGetByPeriodNil3 tests GetByPeriod when the passed period is nil.
func TestFirestoreGetByPeriodNil3(t *testing.T) {
	var repo *tseventserver.FirestoreEventRepository = &tseventserver.FirestoreEventRepository{}
	if _, err := repo.GetByPeriod(context.Background(), nil); err == nil {
		t.Fatal(tserr.NilFailed("GetByPeriod"))
	}
}

// TestFirestoreStoreAndGetByPeriod is a functional test that stores an event
// and then retrieves it to ensure Firestore data mapping and queries work.
func TestFirestoreStoreAndGetByPeriod(t *testing.T) {
	// Set up the FirestoreEventRepository for testing and ensure cleanup after the test completes.
	repo, cleanup := setupFirestore(t)
	// Defer the cleanup function to ensure that resources are released after the test, even if it fails.
	defer cleanup()
	// Define a sample event for testing purposes
	ev := evGdp24
	// 1. Store the event
	if err := repo.Store(context.Background(), ev); err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Store", Fn: "Firestore", Err: err}))
	}
	// 2. Retrieve events by the date of the stored event
	// We create a period that encompasses the date of the stored event to ensure
	// that it will be included in the results when we query by period.
	period := tstrading.NewPeriodForDate(ev.Time)
	// We use GetByPeriod instead of GetByDate to align with the current repository implementation and
	// to ensure that we are testing the correct retrieval method.
	events, err := repo.GetByPeriod(context.Background(), period)
	// Check for errors during retrieval and fail the test if any occur, providing context about the operation that failed.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "GetByPeriod", Fn: "Firestore", Err: err}))
	}
	// 3. Verify it was correctly retrieved and mapped
	// We check if the retrieved events include the stored event by comparing them using NearEqual,
	// which allows for a more flexible comparison that accounts for potential differences
	// in how data is stored and retrieved from Firestore.
	found := false
	// Iterate through the retrieved events to check if any of them match the stored event using NearEqual for comparison.
	for _, retrievedEv := range events {
		// Use NearEqual to compare the stored event with each retrieved event,
		// which allows for a more flexible comparison that accounts for potential differences in
		// how data is stored and retrieved from Firestore.
		if ev.NearEqual(&retrievedEv) {
			// If a match is found, set found to true and break out of the loop.
			found = true
			break
		}
	}
	// If no matching event was found in the retrieved results, fail the test and provide context about the failure.
	if !found {
		t.Fatal(tserr.NotFound(ev.Name))
	}
}

// TestFirestoreGetByPeriodEmpty tests querying an empty time period.
func TestFirestoreGetByPeriodEmpty(t *testing.T) {
	// Set up the FirestoreEventRepository for testing and ensure cleanup after the test completes.
	repo, cleanup := setupFirestore(t)
	// Defer the cleanup function to ensure that resources are released after the test, even if it fails.
	defer cleanup()
	// Define a period far in the future where no events should exist
	p := &tstrading.Period{
		From: time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC),
	}
	// Attempt to retrieve events for the defined period, which should return an empty result
	// set since no events are expected to exist in that time frame.
	events, err := repo.GetByPeriod(context.Background(), p)
	// Check for errors during retrieval and fail the test if any occur, providing context about the operation that failed.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "GetByPeriod", Fn: "Firestore", Err: err}))
	}
	// Verify that no events were retrieved for the empty period, which is the expected outcome.
	if len(events) != 0 {
		t.Fatal(tserr.EqualInt(&tserr.EqualIntArgs{
			Var:    "len(events)",
			Actual: int64(len(events)),
			Want:   0,
		}))
	}
}
