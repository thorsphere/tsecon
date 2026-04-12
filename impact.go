// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// ImpactLevel represents the expected market impact of an economic event.
type ImpactLevel int

// Define constants for the different impact levels.
const (
	ImpactLow    ImpactLevel = iota // iota starts at 0, so ImpactLow = 0
	ImpactMedium                    // ImpactMedium = 1
	ImpactHigh                      // ImpactHigh = 2
)

// String returns a string representation of the ImpactLevel.
func (i ImpactLevel) String() string {
	switch i {
	case ImpactLow:
		return "low" // Return "low" for ImpactLow
	case ImpactMedium:
		return "medium" // Return "medium" for ImpactMedium
	case ImpactHigh:
		return "high" // Return "high" for ImpactHigh
	default:
		return "unknown" // Return "unknown" for any undefined impact levels
	}
}
