// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public Licence v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon, tsfio and tserr
import (
	"strings"
	"testing" // testing

	// time
	"github.com/thorsphere/tsecon" // tsecon
	"github.com/thorsphere/tserr"
	"github.com/thorsphere/tsfio" // tsfio
)

// TestEvents tests the String method of the Event struct by comparing the output to a golden file.
func TestEvents(t *testing.T) {
	// Create a formatted string representation of the sample events using the String method of the Event struct
	var out strings.Builder
	// Iterate over each event in the sample events slice and append its string representation to the output string
	for _, ev := range evs {
		out.WriteString(ev.String() + "\n")
	}
	// Compare the output to a golden file using the EvalGoldenFile function from the tsfio package,
	// and if there is an error, fail the test with the error message
	if e := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "events", Data: out.String()}); e != nil {
		t.Fatal(e)
	}
}

// TestPrintEvents tests the PrintEvents function by comparing its output to a golden file.
func TestPrintEvents(t *testing.T) {
	// Use the PrintEvents function to get a formatted string representation of the sample events
	out := tsecon.PrintEvents(evs)
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
	s := tsecon.PrintEvents([]tsecon.Event{})
	// If there is an error printing events, fail the test with the error message
	if s != "" {
		t.Fatal(tserr.EqualStr(&tserr.EqualStrArgs{Var: "PrintEvents", Actual: s, Want: ""}))
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
