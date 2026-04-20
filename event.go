// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// Import standard library packages
import (
	"fmt"  // fmt
	"time" // time
)

// EconomicEvent represents a single calendar event with its details.
// Todo: Implement a separate package for country codes and use it here for better type safety and validation.
type Event struct {
	ID       int64       `json:"id"`       // Unique identifier for the event, e.g., a database primary key or a UUID. It is expected to be set by the database.
	Name     string      `json:"name"`     // Name of the economic event, e.g., "Non-Farm Payrolls", "GDP Growth Rate"
	Time     time.Time   `json:"date"`     // Date and time of the event in UTC, when the data is released or expected to be released
	Country  string      `json:"country"`  // ISO 3166-1 alpha-2 two-letter country code
	Actual   *float64    `json:"actual"`   // Pointer, because it can be nil if the value is not yet released
	Estimate *float64    `json:"estimate"` // Pointer, because it can be nil if the value is not yet released
	Previous *float64    `json:"previous"` // Pointer, because it can be nil if the value is not yet released
	Unit     string      `json:"unit"`     // Unit of measurement for the values, e.g., "%", "K", "M", "B"
	Impact   ImpactLevel `json:"impact"`   // Impact level of the event
	Source   string      `json:"source"`   // Source of the data, e.g., "Bloomberg", "Reuters", "Official Government Website"
}

// String returns a formatted string representation of the Event.
// Todo: Use tstable package to print as a table
// Todo: Add a method to compare actual vs estimate and previous, and return a string indicating if it's better, worse, or as expected.
// Todo: Add a helper function to print formatted *float64 values, handling nil pointers gracefully.
func (ev Event) String() string {
	t := fmt.Sprintf("%d: %s (%s) at %s - Impact: %s", ev.ID, ev.Name, ev.Country, ev.Time.Format(time.RFC3339), ev.Impact.String())
	t += "\n  Actual: " + fmtFloatPtr(ev.Actual) + " " + ev.Unit
	t += "\n  Estimate: " + fmtFloatPtr(ev.Estimate) + " " + ev.Unit
	t += "\n  Previous: " + fmtFloatPtr(ev.Previous) + " " + ev.Unit
	t += "\n  Source: " + ev.Source
	return t
}

// Equal compares two Event instances for equality, taking into account all fields including
// the pointer fields for Actual, Estimate, and Previous. It does not compare the ID field,
// as it is expected to be set by the database and may not be the same for two events that are otherwise identical.
// It returns true if all fields are equal, and false otherwise.
func (ev Event) Equal(other Event) bool {
	return ev.Name == other.Name &&
		ev.Time.Equal(other.Time) &&
		ev.Country == other.Country &&
		floatPtrEqual(ev.Actual, other.Actual) &&
		floatPtrEqual(ev.Estimate, other.Estimate) &&
		floatPtrEqual(ev.Previous, other.Previous) &&
		ev.Unit == other.Unit &&
		ev.Impact == other.Impact &&
		ev.Source == other.Source
}
