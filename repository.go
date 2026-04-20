// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// Import standard library packages and third-party packages for database/sql, time, and custom error handling and file I/O utilities.
import (
	"database/sql" // database/sql for interacting with the SQLite database
	"time"         // time for handling event timestamps

	_ "github.com/ncruces/go-sqlite3/driver" // SQLite driver for database/sql
	"github.com/thorsphere/tserr"            // tserr for custom error handling
	"github.com/thorsphere/tsfio"            // tsfio for file I/O utilities
)

// This file defines the EventRepository interface and its SQLite implementation
// for managing economic events in a storage system.
const (
	// dbParam defines the connection parameters for the SQLite database,
	// enabling Write-Ahead Logging (WAL) and setting a timeout for database operations.
	dbParam = "?_journal=WAL&_timeout=5000"
	// dbCreateTableQuery is the SQL query used to create the economic_events table
	// if it does not already exist. The table includes columns for event details such as
	// name, time, country, actual/estimate/previous values, unit, impact, and source.
	dbCreateTableQuery = `
	CREATE TABLE IF NOT EXISTS economic_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		event_time DATETIME NOT NULL,
		country TEXT NOT NULL,
		actual REAL,
		estimate REAL,
		previous REAL,
		unit TEXT,
		impact INT,
		source TEXT,
        UNIQUE(event_time, country, name)
	);`
)

// EventRepository defines the interface for managing economic events in a storage system.
type EventRepository interface {
	Store(event *Event) error                    // Store saves a new economic event to the repository. If it already exists, it should update the existing record.
	GetByDate(date time.Time) ([]Event, error)   // GetByDate retrieves economic events that occurred on a specific date.
	GetByPeriod(period *Period) ([]Event, error) // GetByPeriod retrieves economic events that occurred within a specified time period.
	Close() error                                // Close releases any resources held by the repository, such as database connections.
}

// SQLiteEventRepository is an implementation of the EventRepository interface
// that uses SQLite as the underlying storage mechanism.
type SQLiteEventRepository struct {
	fn tsfio.Filename // The filename of the SQLite database file
	db *sql.DB        // The database connection pool for interacting with the SQLite database
}

// GetByDate retrieves economic events that occurred on a specific date from the SQLiteEventRepository.
// It checks for nil pointers and handles database queries and errors appropriately,
// returning a slice of Event structs that match the specified date.
func (r *SQLiteEventRepository) GetByDate(date time.Time) ([]Event, error) {
	// Check if the repository instance is nil
	if r == nil {
		return nil, tserr.NilPtr()
	}
	// Check if the database connection is nil
	if r.db == nil {
		return nil, tserr.NilPtr()
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Nanosecond)

	// Prepare the SQL statement for retrieving events by date
	stmt := `
    SELECT id, name, event_time, country, actual, estimate, previous, unit, impact, source
    FROM economic_events
    WHERE event_time >= ? AND event_time <= ?
    `
	// Execute the SQL statement with the date parameter
	rows, err := r.db.Query(stmt, startOfDay.UTC().Format(time.RFC3339), endOfDay.UTC().Format(time.RFC3339))
	// Handle any errors that occur while executing the query
	if err != nil {
		return nil, tserr.Op(&tserr.OpArgs{
			Op:  "db.Query",
			Err: err,
		})
	}

	// Iterate over the rows returned by the query and scan the event details into Event structs.
	var events []Event
	for rows.Next() {
		// Create a new Event struct to hold the details of the current row
		var event Event
		// Scan the columns of the current row into the fields of the Event struct
		if err := rows.Scan(&event.ID, &event.Name, &event.Time, &event.Country, &event.Actual, &event.Estimate, &event.Previous, &event.Unit, &event.Impact, &event.Source); err != nil {
			// If there is an error scanning the row, close the rows and return the error
			rows.Close()
			return nil, tserr.Op(&tserr.OpArgs{
				Op:  "rows.Scan",
				Err: err,
			})
		}
		// Append the scanned event to the list of events to return
		events = append(events, event)
	}
	// Check for any errors that occurred during iteration over the rows
	if err := rows.Close(); err != nil {
		return nil, tserr.Op(&tserr.OpArgs{
			Op:  "rows.Close",
			Err: err,
		})
	}
	// Return the list of events that match the specified date and nil for the error
	return events, nil
}

// Store saves a new economic event to the SQLiteEventRepository.
// If an event with the same name, time, and country already exists,
// it updates the existing record with the new details.
func (r *SQLiteEventRepository) Store(event *Event) error {
	// Check if the repository instance is nil
	if r == nil {
		return tserr.NilPtr()
	}
	// Check if the database connection is nil
	if r.db == nil {
		return tserr.NilPtr()
	}
	// Check if the event instance is nil
	if event == nil {
		return tserr.NilPtr()
	}
	// Prepare the SQL statement for inserting or updating an event
	stmt := `
    INSERT INTO economic_events (name, event_time, country, actual, estimate, previous, unit, impact, source)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(event_time, country, name) DO UPDATE SET
        actual=excluded.actual,
        estimate=excluded.estimate,
        previous=excluded.previous,
        unit=excluded.unit,
        impact=excluded.impact,
        source=excluded.source;
    `
	// Execute the SQL statement with the event details
	_, err := r.db.Exec(stmt, event.Name, event.Time.UTC().Format(time.RFC3339), event.Country, event.Actual, event.Estimate, event.Previous, event.Unit, event.Impact, event.Source)
	if err != nil {
		return tserr.Op(&tserr.OpArgs{
			Op:  "db.Exec",
			Err: err,
		})
	}
	// Return nil to indicate success
	return nil
}

// Close closes the database connection of the SQLiteEventRepository.
func (r *SQLiteEventRepository) Close() error {
	// Check if the repository instance is nil
	if r == nil {
		return tserr.NilPtr()
	}
	// Check if the database connection is nil
	if r.db == nil {
		return tserr.NilPtr()
	}
	// Attempt to close the database connection and handle any errors that occur
	if err := r.db.Close(); err != nil {
		return tserr.Op(&tserr.OpArgs{
			Op:  "db.Close",
			Err: err,
		})
	}
	// If the connection is closed successfully, return nil to indicate success
	return nil
}

// NewSQLiteEventRepository initializes a new SQLiteEventRepository with the given filename fn.
func NewSQLiteEventRepository(fn tsfio.Filename) (*SQLiteEventRepository, error) {
	// Check if the file is accessible and writable
	if err := tsfio.CheckFile(fn); err != nil {
		return nil, err
	}
	// Construct the SQLite connection string with appropriate parameters
	source := "file:" + string(fn) + dbParam
	// Open the SQLite database connection
	db, err := sql.Open("sqlite3", source)
	// Handle any errors that occur while opening the database
	if err != nil {
		return nil, tserr.Op(&tserr.OpArgs{
			Op:  "sql.Open",
			Err: err,
		})
	}
	// Force connection initialization and testing
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, tserr.Op(&tserr.OpArgs{
			Op:  "db.Ping",
			Err: err,
		})
	}
	// Configure the database connection pool for optimal performance with SQLite
	db.SetMaxOpenConns(1)    // SQLite does not support concurrent writes, so we limit to 1 connection
	db.SetMaxIdleConns(1)    // Keep one idle connection for reuse
	db.SetConnMaxLifetime(0) // Connections can be reused indefinitely
	// Create the events table if it does not exist
	if _, err := db.Exec(dbCreateTableQuery); err != nil {
		return nil, tserr.Op(&tserr.OpArgs{
			Op:  "db.Exec",
			Err: err,
		})
	}
	// Return the initialized repository instance
	return &SQLiteEventRepository{fn: fn, db: db}, nil
}
