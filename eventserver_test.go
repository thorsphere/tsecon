// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsecon_test

// Import standard library packages, tsecon and tserr
import (
	"bytes"
	"context" // context
	"encoding/json"
	"net/http" // http
	"testing"  // testing
	"time"     // time

	"github.com/thorsphere/tsecon" // tsecon
	"github.com/thorsphere/tserr"  // tserr
)

func TestIngestHandler(t *testing.T) {
	// Create a new SQLiteEventRepository with the temporary directory
	repo, fn := tmpDB(t)
	// Create a new EventServer instance
	server := tsecon.NewEventServer(repo)
	// Create a channel to capture server errors
	errChan := make(chan error, 1)
	// Start the server in a separate goroutine
	go func() {
		if err := server.ListenAndServe(":8080"); err != nil && err != http.ErrServerClosed {
			// Send the error to the error channel
			errChan <- err
		}
	}()
	// Give the server a small moment to bind to the port
	time.Sleep(3000 * time.Millisecond)
	// Create a channel to capture server errors
	select {
	case err := <-errChan: // If the server failed to start, report the error and fail the test
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "server.ListenAndServe", Err: err}))
	default:
		payload, err := json.Marshal(evs)
		if err != nil {
			t.Fatal(tserr.Op(&tserr.OpArgs{Op: "json.Marshal", Fn: "events", Err: err}).Error())
		}
		// Perform an HTTP POST request to the server's ingest endpoint
		resp, err := http.Post("http://localhost:8080/api/ingest", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(tserr.Op(&tserr.OpArgs{Op: "http.Post", Fn: "events", Err: err}).Error())
		}
		// Close the response body when the function exits
		resp.Body.Close()
		// Verify that the server responded with an HTTP 200 OK status
		if resp.StatusCode != http.StatusOK {
			t.Error(tserr.StatusNotMatching(&tserr.StatusNotMatchingArgs{Expected: http.StatusOK, Actual: resp.StatusCode}))

		}
		// Create a context with a timeout for the shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// Defer the shutdown context
		defer cancel()
		// Close the server
		if err := server.Shutdown(ctx); err != nil {
			t.Fatal(tserr.Op(&tserr.OpArgs{Op: "server.Shutdown", Err: err}))
		}
	}
	// Remove the temporary database and directory
	rmDB(t, repo, fn)
}
