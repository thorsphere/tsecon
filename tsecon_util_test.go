// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon, tserr and tsfio
import (
	"context" // context for managing request-scoped values and cancellation
	// os for file and directory operations
	// filepath for constructing file paths
	// testing for writing test cases
	"github.com/thorsphere/tsecon" // tsecon for the package being tested
	"github.com/thorsphere/tserr"  // tserr for custom error handling
	// tsfio for file input/output operations, including handling golden files
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
