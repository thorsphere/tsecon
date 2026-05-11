// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon

// Import necessary packages for JSON handling, HTTP server functionality,
// and custom error and logging utilities from the thorsphere project.
import (
	"context"       // context
	"encoding/json" // json
	"fmt"
	"net/http" // http
	"time"

	"github.com/thorsphere/tserr" // tserr
	"github.com/thorsphere/tslog" // tslog
)

// EventServer is a struct that represents the event server responsible for handling event ingestion.
type EventServer struct {
	repo   EventRepository
	server *http.Server
}

// NewEventServer creates a new instance of EventServer with the provided EventRepository.
func NewEventServer(repo EventRepository) *EventServer {
	// Initialize a new HTTP ServeMux to act as the request router.
	mux := http.NewServeMux()
	// Instantiate the EventServer with the injected EventRepository dependency.
	s := &EventServer{repo: repo}
	// Register the IngestHandler to process events at the "/api/ingest" endpoint.
	mux.HandleFunc("/api/ingest", s.IngestHandler)
	// Register the RetrieveHandler to fetch events at the "/api/events" endpoint.
	mux.HandleFunc("/api/retrieve", s.RetrieveHandler)
	// Initialize the embedded HTTP server with the configured request multiplexer.
	s.server = &http.Server{Handler: mux}
	// Return the fully configured EventServer instance, ready to be started.
	return s
}

// ListenAndServe starts the HTTP server on the specified address.
func (s *EventServer) ListenAndServe(addr string) error {
	// Set the TCP network address (e.g., ":8080") on which the server should listen.
	s.server.Addr = addr
	// Start listening for incoming HTTP requests. This method is blocking and
	// will run until the server is closed or encounters a fatal error.
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server, waiting for active connections to finish.
func (s *EventServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ServeHTTP makes EventServer implement the http.Handler interface,
// allowing it to easily route incoming requests to its internal multiplexer.
// This handles HTTP requests natively and is especially useful for testing.
func (s *EventServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.Handler.ServeHTTP(w, r)
}

// IngestHandler handles incoming HTTP requests to ingest events.
// It expects a POST request with a JSON body containing an array of events.
// The handler decodes the JSON, stores each event in the repository,
// and returns an appropriate HTTP response based on the success or failure of the operation.
func (s *EventServer) IngestHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests for event ingestion
	if r.Method != http.MethodPost {
		http.Error(w, tserr.MethodNotAllowed(&tserr.MethodNotAllowedArgs{Method: r.Method, Resource: "EventServer"}).Error(), http.StatusMethodNotAllowed)
		return
	}
	// Decode the JSON body into a slice of Event structs.
	// If there is an error during decoding, return a bad request error.
	var events []Event
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
		if err := s.repo.Store(&ev); err != nil {
			tslog.Error(err)
		}
	}
	// If all events are processed successfully, return a success response.
	w.WriteHeader(http.StatusOK)
	// Write a success message to the response body.
	w.Write([]byte("Successfully ingested events\n"))
	// Close the request body to free up resources.
	r.Body.Close()
}

// RetrieveHandler handles incoming HTTP GET requests to fetch events.
// It expects query parameters 'from' and 'to' in RFC3339 format.
func (s *EventServer) RetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests for retrieving events
	if r.Method != http.MethodGet {
		http.Error(w, tserr.MethodNotAllowed(&tserr.MethodNotAllowedArgs{Method: r.Method, Resource: "EventServer"}).Error(), http.StatusMethodNotAllowed)
		return
	}
	// Extract the 'from' and 'to' timestamp parameters from the URL query string
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	// Initialize a Period pointer to define the time range for the database query
	var period *Period
	if fromStr != "" && toStr != "" {
		// Parse the provided query strings into time.Time objects using RFC3339 format
		from, err1 := time.Parse(time.RFC3339, fromStr)
		to, err2 := time.Parse(time.RFC3339, toStr)
		// If both timestamps are valid, assign them to the period
		if err1 == nil && err2 == nil {
			period = &Period{From: from, To: to}
		} else { // If parsing fails for either timestamp, return a 400 Bad Request error
			http.Error(w, tserr.InvalidTimestampFormat(fmt.Errorf("%w %w", err1, err2)).Error(), http.StatusBadRequest)
			return
		}
	} else { // If no specific period is provided, default to retrieving events for the current day
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		period = &Period{From: startOfDay, To: startOfDay.AddDate(0, 0, 1)}
	}
	// Fetch the economic events from the repository for the determined time period
	events, err := s.repo.GetByPeriod(period)
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
