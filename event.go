// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// Import standard library packages
import "time" // time

// EconomicEvent represents a single calendar event with its details.
// Todo: Implement a separate package for country codes and use it here for better type safety and validation.
type Event struct {
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
	t := ev.Name + " (" + ev.Country + ") at " + ev.Time.Format(time.RFC3339) + " - Impact: " + ev.Impact.String()
	t += "\n  Actual: " + fmtFloatPtr(ev.Actual) + " " + ev.Unit
	t += "\n  Estimate: " + fmtFloatPtr(ev.Estimate) + " " + ev.Unit
	t += "\n  Previous: " + fmtFloatPtr(ev.Previous) + " " + ev.Unit
	return t
}
