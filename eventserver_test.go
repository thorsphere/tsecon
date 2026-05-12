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

// setupTestServer starts the server and registers the teardown/cleanup automatically.
// It returns the server URL and the initialized repository.
func setupTestServer(t *testing.T) (string, tsecon.EventRepository) {
	repo, fn := tmpDB(t)
	server := tsecon.NewEventServer(repo)

	// httptest.NewServer automatically binds to a random free OS port (127.0.0.1:0)
	// and starts the server immediately. No sleeps or error channels required!
	ts := httptest.NewServer(server)

	// Register automated cleanup. t.Cleanup runs automatically when the test using it finishes.
	t.Cleanup(func() {
		ts.Close() // Immediately shuts down the httptest server
		rmDB(t, repo, fn)
	})

	// ts.URL contains the actual active URL with the random port (e.g., http://127.0.0.1:54932)
	return ts.URL, repo
}

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
		expectedStatus int
	}{
		{
			name:           "Valid Ingestion",
			method:         http.MethodPost,
			payload:        validPayload, // Use the valid JSON payload prepared above
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Invalid JSON",
			method: http.MethodPost,
			// Missing closing bracket to make it invalid JSON
			payload:        `[{"name": "GDP Growth", "date": 2024-07-10 }`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong HTTP Method",
			method:         http.MethodGet,
			payload:        `[]`, // Empty array, but method is GET which is not allowed
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}
	// Iterate over the test cases and execute them
	for _, tc := range tests {
		// Use t.Run to run each test case as a subtest, which provides better reporting and isolation.
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request with the specified method, URL, and payload.
			req, err := http.NewRequest(tc.method, baseURL+"/api/ingest", bytes.NewBufferString(tc.payload))
			if err != nil {
				// If there is an error creating the request, fail the test with a detailed error message.
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "http.NewRequest", Fn: tc.name, Err: err}).Error())
			}
			// Set the Content-Type header to application/json since the payload is JSON.
			req.Header.Set("Content-Type", "application/json")
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
		expectedStatus int
	}{
		{
			name:   "Valid timeframe with results",
			method: http.MethodGet,
			// Timeframe that encompasses evNfp and evGdp24 from the mock data
			endpoint:       "/api/retrieve?from=2024-01-01T00:00:00Z&to=2025-12-31T23:59:59Z",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing timeframe (defaults to today)",
			method:         http.MethodGet,
			endpoint:       "/api/retrieve",
			expectedStatus: http.StatusOK, // Expected to succeed and return an empty map/null
		},
		{
			name:           "Invalid timeframe format",
			method:         http.MethodGet,
			endpoint:       "/api/retrieve?from=not-a-date&to=also-not-a-date",
			expectedStatus: http.StatusBadRequest, // Invalid timestamps should result in a 400 Bad Request
		},
		{
			name:           "Wrong HTTP Method",
			method:         http.MethodPost,
			endpoint:       "/api/retrieve?from=2024-01-01T00:00:00Z&to=2025-12-31T23:59:59Z",
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
