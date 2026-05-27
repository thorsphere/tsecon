// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon and tserr
import (
	"bytes" // context
	"context"
	"encoding/json"     // json
	"net/http"          // http
	"net/http/httptest" // httptest
	"testing"           // testing

	"github.com/thorsphere/tsecon" // tsecon
	"github.com/thorsphere/tserr"  // tserr
)

const (
	tok = "all-your-base-are-belong-to-us" // Dummy token for testing authentication
)

// mockRepo is an in-memory implementation of tsecon.EventRepository for testing.
type mockRepo struct {
	events map[string]*tsecon.Event // map of events by ID
}

// newMockRepo creates a new mockRepo instance.
func newMockRepo() *mockRepo {
	// Create a new mockRepo instance
	return &mockRepo{events: make(map[string]*tsecon.Event)}
}

// Store adds an event to the mockRepo's events map.
func (m *mockRepo) Store(ctx context.Context, ev *tsecon.Event) error {
	// Check if the repository instance is nil
	if m == nil {
		return tserr.NilPtr()
	}
	// Check if the event is nil
	if ev == nil {
		return tserr.NilPtr()
	}
	// Add the event to the mockRepo's events map
	m.events[ev.GenerateDocID()] = ev
	// Return nil to indicate success
	return nil
}

// GetByPeriod retrieves events from the mockRepo's events map based on the provided period.
func (m *mockRepo) GetByPeriod(ctx context.Context, period *tsecon.Period) ([]tsecon.Event, error) {
	// Check if the repository instance is nil
	if m == nil {
		return nil, tserr.NilPtr()
	}
	// Check if the period is nil
	if period == nil {
		return nil, tserr.NilPtr()
	}
	// Create a slice of events that match the period
	var result []tsecon.Event
	// Iterate over the events in the mockRepo's events map
	for _, ev := range m.events {
		// Check if the event's time is within the specified period
		if !ev.Time.Before(period.From) && !ev.Time.After(period.To) {
			// If the event is within the period, append it to the result slice
			result = append(result, *ev)
		}
	}
	// If the result slice is empty, return an empty slice
	if result == nil {
		result = []tsecon.Event{}
	}
	// Return the slice of events that match the specified period
	return result, nil
}

// Close does nothing in the mockRepo implementation.
func (m *mockRepo) Close() error {
	return nil
}

// setupTestServer sets up a test server and returns the base URL and the mockRepo instance.
func setupTestServer(t *testing.T) (string, tsecon.EventRepository) {
	// Use t.Helper() to call t.Fatal if any of the following steps fail
	t.Helper()
	// Create a new mockRepo instance
	repo := newMockRepo()
	// Create a new EventServer instance with the mockRepo
	server := tsecon.NewEventServer(repo, tok)
	// Create a new HTTP test server with the EventServer
	ts := httptest.NewServer(server)
	// Close the test server to clean up resources
	t.Cleanup(func() { ts.Close() })
	// Return the base URL and the mockRepo instance
	return ts.URL, repo
}

// TestIngestHandler tests the ingestHandler of the EventServer for various scenarios.
func TestIngestHandler(t *testing.T) {
	// Setup test server and get the base URL for requests
	baseURL, _ := setupTestServer(t)
	// Prepare valid test data (using the same events as in TestEventRepository)
	validPayloadBytes, err := json.Marshal(evs)
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "json.Marshal", Fn: "evs", Err: err}).Error())
	}
	// Convert the byte slice to a string for use in the test cases
	validPayload := string(validPayloadBytes)
	// Define test cases for different scenarios
	tests := []struct {
		name           string
		method         string
		payload        string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid Ingestion",
			method:         http.MethodPost,
			payload:        validPayload, // Use the valid JSON payload prepared above
			authHeader:     "Bearer " + tok,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized Access",
			method:         http.MethodPost,
			payload:        validPayload, // Payload is valid, but auth header is incorrect
			authHeader:     "Bearer " + "totally-wrong-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Unauthorized Access (Missing token)",
			method:         http.MethodPost,
			payload:        validPayload, // Payload is valid, but auth header is incorrect
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Invalid JSON",
			method: http.MethodPost,
			// Missing closing bracket to make it invalid JSON
			payload:        `[{"name": "GDP Growth", "date": 2024-07-10 }`,
			authHeader:     "Bearer " + tok,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong HTTP Method",
			method:         http.MethodGet,
			payload:        `[]`, // Empty array, but method is GET which is not allowed
			authHeader:     "Bearer " + tok,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}
	// Iterate over the test cases and execute them
	for _, tc := range tests {
		// Use t.Run to run each test case as a subtest, which provides better reporting and isolation.
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request with the specified method, URL, and payload.
			req, err := http.NewRequest(tc.method, baseURL+"/events/ingest", bytes.NewBufferString(tc.payload))
			if err != nil {
				// If there is an error creating the request, fail the test with a detailed error message.
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "http.NewRequest", Fn: tc.name, Err: err}).Error())
			}
			// Set the Content-Type header to application/json since the payload is JSON.
			req.Header.Set("Content-Type", "application/json")
			// Set the Authorization header if it's provided.
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			// Send the HTTP request using the default HTTP client and capture the response.
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				// If there is an error sending the request, fail the test with a detailed error message.
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "http.DefaultClient.Do", Fn: tc.name, Err: err}).Error())
			}
			// Defer the closing of the response body to ensure resources are freed after the test case is done.
			defer resp.Body.Close()
			// Check if the response status code matches the expected status code for the test case.
			if resp.StatusCode != tc.expectedStatus {
				t.Error(tserr.StatusNotMatching(&tserr.StatusNotMatchingArgs{Expected: tc.expectedStatus, Actual: resp.StatusCode}))
			}
		})
	}
}

// TestRetrieveHandler tests the retrieveHandler of the EventServer for various scenarios.
// It checks for valid retrieval, missing timeframe, invalid timeframe format, and wrong HTTP method.
func TestRetrieveHandler(t *testing.T) {
	// Setup test server and get the base URL and repository instance
	baseURL, repo := setupTestServer(t)
	// Seed the database with our sample events before running retrieve tests
	for _, ev := range evs {
		e := ev // create a local copy to pass a stable pointer
		if err := repo.Store(context.Background(), &e); err != nil {
			t.Fatal(tserr.Op(&tserr.OpArgs{Op: "repo.Store", Fn: ev.Name, Err: err}).Error())
		}
	}
	// Define test cases for different retrieve scenarios
	tests := []struct {
		name           string
		method         string
		endpoint       string
		authHeader     string
		expectedStatus int
	}{
		{
			name:   "Valid timeframe with results",
			method: http.MethodGet,
			// Timeframe that encompasses evNfp and evGdp24 from the mock data
			endpoint:       "/events/retrieve?from=2024-01-01T00:00:00Z&to=2025-12-31T23:59:59Z",
			authHeader:     "Bearer " + tok,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized Access",
			method:         http.MethodGet,
			endpoint:       "/events/retrieve?from=2024-01-01T00:00:00Z&to=2025-12-31T23:59:59Z",
			authHeader:     "Bearer " + "totally-wrong-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing timeframe (defaults to today)",
			method:         http.MethodGet,
			endpoint:       "/events/retrieve",
			authHeader:     "Bearer " + tok,
			expectedStatus: http.StatusOK, // Expected to succeed and return an empty map/null
		},
		{
			name:           "Invalid timeframe format",
			method:         http.MethodGet,
			endpoint:       "/events/retrieve?from=not-a-date&to=also-not-a-date",
			authHeader:     "Bearer " + tok,
			expectedStatus: http.StatusBadRequest, // Invalid timestamps should result in a 400 Bad Request
		},
		{
			name:           "Wrong HTTP Method",
			method:         http.MethodPost,
			endpoint:       "/events/retrieve?from=2024-01-01T00:00:00Z&to=2025-12-31T23:59:59Z",
			authHeader:     "Bearer " + tok,
			expectedStatus: http.StatusMethodNotAllowed, // POST is not allowed on the retrieve endpoint
		},
	}

	// Iterate over the test cases and execute them
	for _, tc := range tests {
		// Use t.Run to run each test case as a subtest, which provides better reporting and isolation.
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request without a body (nil) because it's a GET request
			req, err := http.NewRequest(tc.method, baseURL+tc.endpoint, nil)
			if err != nil {
				// If there is an error creating the request, fail the test with a detailed error message.
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "http.NewRequest", Fn: tc.name, Err: err}).Error())
			}
			// Set the Authorization header if it's provided.
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			// Send the HTTP request using the default HTTP client
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				// If there is an error sending the request, fail the test with a detailed error message.
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "http.DefaultClient.Do", Fn: tc.name, Err: err}).Error())
			}
			// Defer the closing of the response body to ensure resources are freed after the test case is done.
			defer resp.Body.Close()

			// Check if the response status code matches the expected status code
			if resp.StatusCode != tc.expectedStatus {
				t.Error(tserr.StatusNotMatching(&tserr.StatusNotMatchingArgs{Expected: tc.expectedStatus, Actual: resp.StatusCode}))
			}
		})
	}
}
