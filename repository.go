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
		impact TEXT,
		source TEXT
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
