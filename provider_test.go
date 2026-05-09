// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public Licence v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon, tsfio and tserr
import (
	"context" // context
	"testing" // testing

	"github.com/thorsphere/tsecon" // tsecon
	"github.com/thorsphere/tserr"  // tserr
	"github.com/thorsphere/tsfio"  // tsfio
)

// TestFetchEvents tests the FetchEvents function by using the Mock provider
// to fetch events for a specified period and comparing the output to a golden file.
func TestFetchEvents(t *testing.T) {
	// Create a new instance of the Mock provider
	p := &Mock{}
	// Use the FetchEvents function to fetch events for the specified period using the Mock provider
	events, err := tsecon.FetchEvents(context.Background(), p, per)
	// If there is an error fetching events, fail the test with the error message
	if err != nil {
		t.Fatal(err)
	}
	// Use the PrintEvents function to get a formatted string representation of the fetched events
	out := tsecon.PrintEvents(events)
	// Compare the output to a golden file using the EvalGoldenFile function from the tsfio package,
	// and if there is an error, fail the test with the error message
	if e := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "fetchevents", Data: out}); e != nil {
		t.Fatal(e)
	}
}

// TestFetchEventsNilProvider tests the FetchEvents function with a nil provider and expects an error.
func TestFetchEventsNilProvider(t *testing.T) {
	// Use the FetchEvents function with a nil provider and a valid period, and expect an error
	_, err := tsecon.FetchEvents(context.Background(), nil, per)
	// If there is no error, fail the test with a message indicating that a nil provider should have caused an error
	if err == nil {
		t.Fatal(tserr.NilExpected("FetchEvents for nil provider"))
	}
}

// TestFetchEventsNilPeriod tests the FetchEvents function with a nil period and expects an error.
func TestFetchEventsNilPeriod(t *testing.T) {
	// Create a new instance of the Mock provider
	p := &Mock{}
	// Use the FetchEvents function with a valid provider and a nil period, and expect an error
	_, err := tsecon.FetchEvents(context.Background(), p, nil)
	if err == nil {
		t.Fatal(tserr.NilExpected("FetchEvents for nil period"))
	}
}

// TestFetchEventsProviderError tests the FetchEvents function with a provider
// that returns an error and expects the error to be propagated.
func TestFetchEventsProviderError(t *testing.T) {
	// Create a new instance of the MockErr provider, which returns an error when GetEvents is called
	p := &MockErr{}
	// Use the FetchEvents function with the MockErr provider and a valid period, and expect an error
	events, err := tsecon.FetchEvents(context.Background(), p, per)
	// If there is no error, fail the test with a message indicating that an error was expected
	if err == nil {
		t.Fatal(tserr.NilFailed("FetchEvents"))
	}
	// If the error message is correct, the test passes, and we can also check that no events were returned
	l := int64(len(events))
	if l != 0 {
		t.Fatal(tserr.Equal(&tserr.EqualArgs{Var: "Length of events", Actual: l, Want: 0}))
	}
}
