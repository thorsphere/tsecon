// Package tseventserver provides the core types and interfaces for economic calendar events,
// including an HTTP‑based ingestion and retrieval server, and pluggable storage backends
// (SQLite and Google Cloud Firestore).
//
// # Key types
//
//   - [Event] represents a single economic release (e.g., GDP, CPI) with its actual,
//     estimate, previous, and impact fields.
//   - [EventRepository] is the storage interface implemented by [SQLiteEventRepository]
//     and [FirestoreEventRepository].
//   - [EventServer] exposes authenticated HTTP endpoints:
//     /events/ingest (POST) and /events/retrieve (GET).
//
// # Typical usage
//
// Create a repository, then start the server:
//
//	repo, err := tseventserver.NewSQLiteEventRepository("events.db")
//	// ... error handling ...
//	defer repo.Close()
//	api := tseventserver.NewEventServer(repo, "your-api-token")
//	http.ListenAndServe(":8080", api)
//
// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tseventserver

// Import necessary packages for JSON handling, HTTP server functionality,
// and custom error and logging utilities from the thorsphere project.
import (
	"encoding/json" // encoding/json for JSON encoding and decoding
	"fmt"           // fmt for string formatting
	"net/http"      // net/http for HTTP server and client functionality
	"time"          // time for handling timestamps and durations

	"github.com/thorsphere/tserr" 		// tserr for custom error handling
	"github.com/thorsphere/tslog" 		// tslog for logging utilities
	"github.com/thorsphere/tstrading"	// tstrading for economic events
)

// EventServer is a struct that represents the event server responsible for handling event ingestion.
type EventServer struct {
	repo EventRepository // EventRepository is an interface that defines methods for storing and retrieving events from a data source.
	mux  *http.ServeMux  // HTTP request multiplexer to route incoming requests to the appropriate handlers
	tok  string          // Authentication token for securing the API

}

// NewEventServer creates a new instance of EventServer with the provided EventRepository.
func NewEventServer(repo EventRepository, tok string) *EventServer {
	// Initialize the EventServer struct with the provided repository and token, and set up the HTTP request multiplexer.
	s := &EventServer{repo: repo, mux: http.NewServeMux(), tok: tok}
	// Register the ingestHandler to process events at the "/events/ingest" endpoint.
	s.mux.HandleFunc("/events/ingest", s.requireAuth(s.ingestHandler))
	// Register the retrieveHandler to fetch events at the "/events/retrieve" endpoint.
	s.mux.HandleFunc("/events/retrieve", s.requireAuth(s.retrieveHandler))
	// Return the fully configured EventServer instance, ready to be started.
	return s
}

// requireAuth is a middleware function that wraps HTTP handlers to enforce authentication using a bearer token.
// It checks the Authorization header of incoming requests against the expected token and returns a 401 Unauthorized response
// if the token is invalid or missing.
func (s *EventServer) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	// Return a new handler function that wraps the original handler with authentication logic.
	return func(w http.ResponseWriter, r *http.Request) {
		// Construct the expected Authorization header value using the provided token.
		expectedHeader := "Bearer " + s.tok
		// Check if the incoming request's Authorization header matches the expected value.
		if r.Header.Get("Authorization") != expectedHeader {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// If the token is valid, call the next handler in the chain to process the request.
		next(w, r)
	}
}

// ServeHTTP makes EventServer implement the http.Handler interface,
// allowing it to easily route incoming requests to its internal multiplexer.
// This handles HTTP requests natively and is especially useful for testing.
func (s *EventServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ingestHandler handles incoming HTTP requests to ingest events.
// It expects a POST request with a JSON body containing an array of events.
// The handler decodes the JSON, stores each event in the repository,
// and returns an appropriate HTTP response based on the success or failure of the operation.
func (s *EventServer) ingestHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests for event ingestion
	if r.Method != http.MethodPost {
		http.Error(w, tserr.MethodNotAllowed(&tserr.MethodNotAllowedArgs{Method: r.Method, Resource: "EventServer"}).Error(), http.StatusMethodNotAllowed)
		return
	}
	// Decode the JSON body into a slice of Event structs.
	// If there is an error during decoding, return a bad request error.
	var events []tstrading.Event
	// Create a new JSON decoder for the request body and decode the JSON into the events slice.
	decoder := json.NewDecoder(r.Body)
	// If there is an error during decoding, return a bad request error with the error message.
	if err := decoder.Decode(&events); err != nil {
		http.Error(w, tserr.InvalidJson(err).Error(), http.StatusBadRequest)
		return
	}
	// Iterate over the decoded events and store each event in the repository.
	for _, ev := range events {
		// Attempt to store the event in the repository. If there is an error,
		// log the error and continue with the next event.
		if err := s.repo.Store(r.Context(), &ev); err != nil {
			tslog.Error(err)
			// If there is an error storing the event, return a 500 Internal Server Error with a detailed error message.
			http.Error(w, tserr.Op(&tserr.OpArgs{Op: "repo.Store", Fn: ev.Name, Err: err}).Error(), http.StatusInternalServerError)
			return
		}
	}
	// If all events are processed successfully, return a success response.
	w.WriteHeader(http.StatusOK)
	// Write a success message to the response body.
	w.Write([]byte(`{"message": "Successfully ingested events"}`))
	// Close the request body to free up resources.
	r.Body.Close()
}

// retrieveHandler handles incoming HTTP GET requests to fetch events.
// It expects query parameters 'from' and 'to' in RFC3339 format.
func (s *EventServer) retrieveHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests for retrieving events
	if r.Method != http.MethodGet {
		http.Error(w, tserr.MethodNotAllowed(&tserr.MethodNotAllowedArgs{Method: r.Method, Resource: "EventServer"}).Error(), http.StatusMethodNotAllowed)
		return
	}
	// Extract the 'from' and 'to' timestamp parameters from the URL query string
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	// Initialize a Period pointer to define the time range for the database query
	var period *tstrading.Period
	if fromStr != "" && toStr != "" {
		// Parse the provided query strings into time.Time objects using RFC3339 format
		from, err1 := time.Parse(time.RFC3339, fromStr)
		to, err2 := time.Parse(time.RFC3339, toStr)
		// If both timestamps are valid, assign them to the period
		if err1 == nil && err2 == nil {
			period = &tstrading.Period{From: from, To: to}
		} else { // If parsing fails for either timestamp, return a 400 Bad Request error
			http.Error(w, tserr.InvalidTimestampFormat(fmt.Errorf("%w %w", err1, err2)).Error(), http.StatusBadRequest)
			return
		}
	} else { // If no specific period is provided, default to retrieving events for the current day
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		period = &tstrading.Period{From: startOfDay, To: startOfDay.AddDate(0, 0, 1)}
	}
	// Fetch the economic events from the repository for the determined time period
	events, err := s.repo.GetByPeriod(r.Context(), period)
	// Handle any unexpected database or repository errors by returning a 500 status
	if err != nil {
		tslog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Buffer the entire JSON in memory first
	jsonBytes, err := json.Marshal(events)
	if err != nil {
		tslog.Error(err)
		// Safe to send http.Error because no headers have been sent yet!
		http.Error(w, tserr.Op(&tserr.OpArgs{Op: "json.Marshal", Fn: "events", Err: err}).Error(), http.StatusInternalServerError)
		return
	}
	// Now send the headers and the buffered bytes
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
