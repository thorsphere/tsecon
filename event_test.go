// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public Licence v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon, tsfio and tserr
import (
	"testing" // testing
	"time"    // time

	"github.com/thorsphere/tsecon" // tsecon
	"github.com/thorsphere/tserr"
	"github.com/thorsphere/tsfio" // tsfio
)

var (
	// Define some sample events for testing purposes
	evNfp tsecon.Event = tsecon.Event{
		Name:     "Non-Farm Payrolls",
		Time:     time.Date(2024, 7, 5, 8, 30, 0, 0, time.UTC),
		Country:  "US",
		Actual:   ptrFloat64(200),
		Estimate: ptrFloat64(180),
		Previous: ptrFloat64(150),
		Unit:     "K",
		Impact:   tsecon.ImpactHigh,
		Source:   "Bureau of Labor Statistics",
	}
	evGdp24 tsecon.Event = tsecon.Event{
		Name:     "GDP Growth Rate",
		Time:     time.Date(2024, 7, 10, 8, 30, 0, 0, time.UTC),
		Country:  "US",
		Actual:   ptrFloat64(3.5),
		Estimate: ptrFloat64(3.0),
		Previous: ptrFloat64(2.8),
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

// Helper function to create a pointer to a float64 value
func ptrFloat64(value float64) *float64 {
	return &value
}

// TestEvents tests the String method of the Event struct by comparing the output to a golden file.
func TestEvents(t *testing.T) {
	// Create a formatted string representation of the sample events using the String method of the Event struct
	out := ""
	// Iterate over each event in the sample events slice and append its string representation to the output string
	for _, ev := range evs {
		out += ev.String() + "\n"
	}
	// Compare the output to a golden file using the EvalGoldenFile function from the tsfio package,
	// and if there is an error, fail the test with the error message
	if e := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "events", Data: out}); e != nil {
		t.Fatal(e)
	}
}

// TestPrintEvents tests the PrintEvents function by comparing its output to a golden file.
func TestPrintEvents(t *testing.T) {
	// Use the PrintEvents function to get a formatted string representation of the sample events
	out, err := tsecon.PrintEvents(evs)
	// If there is an error printing events, fail the test with the error message
	if err != nil {
		t.Fatal(err)
	}
	// Compare the output to a golden file using the EvalGoldenFile function from the tsfio package,
	// and if there is an error, fail the test with the error message
	if e := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "printevents", Data: out}); e != nil {
		t.Fatal(e)
	}
}

// TestPrintNoEvents tests the PrintEvents function with an empty slice of events and
// expects an error indicating that there are no events to print.
func TestPrintNoEvents(t *testing.T) {
	// Use the PrintEvents function with an empty slice of events
	_, err := tsecon.PrintEvents([]tsecon.Event{})
	// If there is an error printing events, fail the test with the error message
	if err == nil {
		t.Fatal(tserr.NilFailed("PrintEvents"))
	}
}

// TestWrongImpact tests the String method of the ImpactLevel type with an invalid impact level value
// and expects the output to be "unknown".
func TestWrongImpact(t *testing.T) {
	var i tsecon.ImpactLevel = 99 // Invalid impact level
	// The expected output for an invalid impact level should be "unknown"
	expected := "unknown"
	// Get the actual string representation of the impact level using the String method
	actual := i.String()
	// If the actual output does not match the expected output, fail the test with an error message
	// indicating the mismatch
	if actual != expected {
		t.Fatal(tserr.EqualStr(&tserr.EqualStrArgs{Var: "ImpactLevel", Actual: actual, Want: expected}))
	}
}
