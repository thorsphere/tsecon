// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// Import standard library packages
import (
	"fmt"  // fmt
	"time" // time
)

// Period represents a time period with a start and end date.
type Period struct {
	From time.Time // Start date and time of the period
	To   time.Time // End date and time of the period
}

// formatFloatPointer is a helper function to format *float64 values, handling nil pointers gracefully.
func fmtFloatPtr(value *float64) string {
	// If the pointer is nil, return "N/A" to indicate that the value is not available.
	if value == nil {
		return "N/A"
	}
	// If the pointer is not nil, format the float value to two decimal places and return it as a string.
	return fmt.Sprintf("%.1f", *value)
}

// floatPtrEqual is a helper function to compare two *float64 values, handling nil pointers gracefully.
func floatPtrEqual(a, b *float64) bool {
	// If both pointers are nil, they are considered equal
	if a == nil && b == nil {
		return true
	}
	// If one pointer is nil and the other is not, they are not equal
	if a == nil || b == nil {
		return false
	}
	// If both pointers are not nil, compare the values they point to
	return *a == *b
}
