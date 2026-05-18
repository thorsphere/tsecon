// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// Import standard library packages and third-party packages for database/sql, time, and custom error handling and file I/O utilities.
import (
	"context" // database/sql for interacting with the SQLite database
)

// EventRepository defines the interface for managing economic events in a storage system.
type EventRepository interface {
	Store(ctx context.Context, event *Event) error                    // Store saves a new economic event to the repository. If it already exists, it should update the existing record.
	GetByPeriod(ctx context.Context, period *Period) ([]Event, error) // GetByPeriod retrieves economic events that occurred within a specified time period.
	Close() error                                                     // Close releases any resources held by the repository, such as database connections.
}
