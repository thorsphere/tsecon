// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// Import standard library packages and tserr
import (
	"context" // context

	"github.com/thorsphere/tserr" // tserr
)

// Provider defines the interface for fetching economic events from a data source.
type Provider interface {
	// GetEvents returns a slice of Event for the specified date range.
	GetEvents(ctx context.Context, period *Period) ([]Event, error)
}

// FetchEvents uses the provided Provider to fetch events for the specified period.
// If the provider or period is nil, it returns an error.
// Otherwise, it returns the fetched events and nil for the error.
func FetchEvents(ctx context.Context, p Provider, period *Period) ([]Event, error) {
	// Check if the provider or period is nil, and return an error if so
	if (p == nil) || (period == nil) {
		return nil, tserr.NilPtr()
	}
	// Fetch events using the provider's GetEvents method
	events, err := p.GetEvents(ctx, period)
	// If there is an error fetching events, return the error
	if err != nil {
		return nil, err
	}
	// Return the fetched events and nil for the error
	return events, nil
}

// PrintEvents prints the details of each event in the provided slice of events.
func PrintEvents(events []Event) (string, error) {
	// If there are no events, print a message indicating that no events were found for the specified period.
	if len(events) == 0 {
		return "", tserr.Empty("events")
	} else { // If there are events, iterate over each event and print its details using the String method of the Event struct.
		out := ""
		for _, event := range events {
			out += event.String() + "\n"
		}
		return out, nil
	}
}
