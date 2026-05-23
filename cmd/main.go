// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package main

// Import necessary packages for context handling, OS signal processing, time management,
// and the tsecon, tserr, and tslog packages from the thorsphere project.
import (
	"context"   // context
	"fmt"       // fmt
	"net/http"  // net/http
	"os"        // os
	"os/signal" // signal
	"syscall"   // syscall
	"time"      // time

	"github.com/thorsphere/tsecon" // tsecon
	"github.com/thorsphere/tserr"  // tserr
	"github.com/thorsphere/tslog"  // tslog
)

const (
	// Define the path to the SQLite database file. This constant can be modified to point to a different location or filename as needed.
	dbPath = "events.db"
	// Define the default port for the server to listen on. This can be overridden by setting the PORT environment variable, which is useful for deployment in environments like Cloud Run that expect the application to listen on a specific port.
	srvPort = "8080"
)

// The main function serves as the entry point for the application.
// It initializes the event repository, starts the event server,
// and handles graceful shutdown on interrupt signals.
func main() {
	// Define the address and port on which the EventServer will listen for incoming HTTP requests.
	// Let Cloud Run decide the port, fallback to 8080 for local development
	port := os.Getenv("PORT")
	if port == "" {
		port = srvPort
	}
	// Construct the server address by combining the host (empty string for all interfaces) and the port.
	serverAddr := ":" + port
	// Read API_TOKEN from environment for basic protection
	tok := os.Getenv("API_TOKEN")
	// If the API_TOKEN environment variable is not set, log a warning and use a default token for local development.
	if tok == "" {
		tslog.Warn("Environment variable 'API_TOKEN' is empty. Using default 'swordfish' for local development.")
		tok = "swordfish"
	}
	// Initialize your repository based on the environment.
	var repo tsecon.EventRepository
	var err error
	// Check if GOOGLE_CLOUD_PROJECT is set to determine if we are running in Google Cloud and should use Firestore.
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	// Check if FIRESTORE_EMULATOR_HOST is set to determine if we should use the Firestore emulator for testing.
	emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	// If either the project ID or the emulator host is set, we will attempt to initialize the FirestoreEventRepository.
	// If neither is set, we will fall back to using the SQLiteEventRepository for local development.
	if projectID != "" || emulatorHost != "" {
		// If the emulator host is set, we will use the emulator for testing.
		// If the project ID is not set, we can default to a dummy value since the emulator does not require a real project ID.
		if projectID == "" {
			projectID = "demo-project"
		}
		// Log the initialization of the FirestoreEventRepository, including whether we are using the emulator or a real Firestore instance.
		tslog.Info(fmt.Sprintf("Initializing FirestoreEventRepository (Project: %s, Emulator: %v).", projectID, emulatorHost != ""))
		// Attempt to initialize the FirestoreEventRepository with the provided project ID and context.
		repo, err = tsecon.NewFirestoreEventRepository(context.Background(), projectID)
	} else {
		// If neither the project ID nor the emulator host is set, log that we are falling back
		// to using the SQLiteEventRepository for local development.
		tslog.Info("FIRESTORE_EMULATOR_HOST / GOOGLE_CLOUD_PROJECT not set. Falling back to local SQLiteEventRepository.")
		// Attempt to initialize the SQLiteEventRepository with the specified database path.
		repo, err = tsecon.NewSQLiteEventRepository(dbPath)
	}
	// If there was an error initializing the repository, log the error and exit the application with a non-zero status code to indicate failure.
	if err != nil {
		// Log the error using tslog and exit the application with a non-zero status code to indicate failure.
		tslog.Error(tserr.Op(&tserr.OpArgs{Op: "Initialize Event Repository", Fn: "main", Err: err}))
		os.Exit(1)
	}
	// Ensure that the repository is properly closed when the main function exits,
	// which will release any resources associated with the repository, such as database connections.
	defer repo.Close()
	// Create a new EventServer instance with the initialized repository.
	api := tsecon.NewEventServer(repo, tok)
	// Configure the HTTP server with the EventServer as the handler, and set reasonable timeouts for read and write operations.
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      api,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	// Start the server in a separate goroutine
	go func() {
		tslog.Info(fmt.Sprintf("Starting EventServer on %v", serverAddr))
		if err := srv.ListenAndServe(); err != nil {
			tslog.Error(tserr.Op(&tserr.OpArgs{Op: "ListenAndServe", Fn: serverAddr, Err: err}))
		}
	}()
	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit // Blocks here until you press Ctrl+C in the terminal
	// Log that we have received an interrupt signal and are beginning the shutdown process.
	tslog.Info("\nShutting down server gracefully...")
	// Create a deadline to wait for currently active requests to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Attempt to gracefully shut down the server, allowing active connections to complete
	if err := srv.Shutdown(ctx); err != nil {
		tslog.Error(tserr.Op(&tserr.OpArgs{Op: "Shutdown", Fn: serverAddr, Err: err}))
	}
	// Log that the server has successfully exited
	tslog.Info("Server successfully exited")
}
