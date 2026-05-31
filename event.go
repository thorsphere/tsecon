// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// Import standard library packages and lpstats for utility functions.
import (
	"crypto/sha256" // sha256 for hashing event details to generate a unique document ID for NoSQL databases
	"encoding/hex"  // hex for encoding the hash output as a hexadecimal string
	"fmt"           // fmt for string formatting
	"time"          // time for handling event timestamps and periods

	"github.com/thorsphere/lpstats" // lpstats for utility functions to compare float pointers and format them as strings
)

// EconomicEvent represents a single calendar event with its details.
// Thought: Implement a separate package for country codes and use it here for better type safety and validation.
type Event struct {
	ID       int64       `json:"id"`       // Unique identifier for the event, e.g., a database primary key or a UUID. It is expected to be set by the database.
	Name     string      `json:"name"`     // Name of the economic event, e.g., "Non-Farm Payrolls", "GDP Growth Rate"
	Time     time.Time   `json:"time"`     // Date and time of the event in UTC, when the data is released or expected to be released
	Country  string      `json:"country"`  // ISO 3166-1 alpha-2 two-letter country code
	Actual   *float64    `json:"actual"`   // Pointer, because it can be nil if the value is not yet released
	Estimate *float64    `json:"estimate"` // Pointer, because it can be nil if the value is not yet released
	Previous *float64    `json:"previous"` // Pointer, because it can be nil if the value is not yet released
	Unit     string      `json:"unit"`     // Unit of measurement for the values, e.g., "%", "K", "M", "B"
	Impact   ImpactLevel `json:"impact"`   // Impact level of the event
	Source   string      `json:"source"`   // Source of the data, e.g., "Bloomberg", "Reuters", "Official Government Website"
}

// Period represents a time period with a start and end date.
type Period struct {
	From time.Time // Start date and time of the period
	To   time.Time // End date and time of the period
}

// NewPeriodForDate creates a Period that spans the entire given date (00:00:00 to 23:59:59) in UTC.
func NewPeriodForDate(date time.Time) *Period {
	// Normalize the input date to the start of the day in UTC
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	// The end of the day is calculated by adding one day to the start of the day and
	// then subtracting one nanosecond to get the last moment of the specified date.
	// Calculate the end of the day by adding 24 hours and subtracting a nanosecond to get 23:59:59.999999999
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Nanosecond)
	// Return a new Period struct with the calculated start and end times for the specified date
	return &Period{From: startOfDay, To: endOfDay}
}

// String returns a formatted string representation of the Event.
// Todo: Use tstable package to print as a table
func (ev Event) String() string {
	t := fmt.Sprintf("%d: %s (%s) at %s - Impact: %s", ev.ID, ev.Name, ev.Country, ev.Time.Format(time.RFC3339), ev.Impact.String())
	t += "\n  Actual: " + lpstats.FmtFloatPtr(ev.Actual) + " " + ev.Unit
	t += "\n  Estimate: " + lpstats.FmtFloatPtr(ev.Estimate) + " " + ev.Unit
	t += "\n  Previous: " + lpstats.FmtFloatPtr(ev.Previous) + " " + ev.Unit
	t += "\n  Source: " + ev.Source
	return t
}

// NearEqual compares two Event instances for near-equality, taking into account all fields including
// the pointer fields for Actual, Estimate, and Previous. It does not compare the ID field,
// as it is expected to be set by the database and may not be the same for two events that are otherwise identical.
// It returns true if all fields are equal, and false otherwise.
func (ev Event) NearEqual(other Event) bool {
	maxDiff := 0.001 // Define a maximum difference for comparing float values, adjust as needed
	return ev.Name == other.Name &&
		ev.Time.Equal(other.Time) &&
		ev.Country == other.Country &&
		lpstats.NearEqualFloatPtr(ev.Actual, other.Actual, maxDiff) &&
		lpstats.NearEqualFloatPtr(ev.Estimate, other.Estimate, maxDiff) &&
		lpstats.NearEqualFloatPtr(ev.Previous, other.Previous, maxDiff) &&
		ev.Unit == other.Unit &&
		ev.Impact == other.Impact &&
		ev.Source == other.Source
}

// GenerateDocID creates a deterministic, safe, unique document ID for NoSQL databases (like Firestore)
// based on the event's Time, Country, and Name.
func (ev Event) GenerateDocID() string {
	// Concatenate the fields that make an event unique
	key := fmt.Sprintf("%s|%s|%s", ev.Time.UTC().Format(time.RFC3339), ev.Country, ev.Name)
	// Hash the string to ensure it is URL-safe and of predictable length
	hash := sha256.Sum256([]byte(key))
	// Encode the hash as a hexadecimal string to use as the document ID
	return hex.EncodeToString(hash[:])
}
