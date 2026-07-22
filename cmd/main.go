// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package main

// Import necessary packages for context handling, OS signal processing, time management,
// and the tseventserver, tserr, and tslog packages from the thorsphere project.
import (
	"context"   // context
	"fmt"       // fmt
	"net/http"  // net/http
	"os"        // os
	"os/signal" // signal
	"syscall"   // syscall
	"time"      // time

	"github.com/thorsphere/tseventserver" // tseventserver
	"github.com/thorsphere/tserr"  		  // tserr
	"github.com/thorsphere/tslog"  		  // tslog
)

const (
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
	// Initialize the Firestore-backed repository.
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	// If FIRESTORE_EMULATOR_HOST is set, we use the Firestore emulator for local testing.
	// In that case a real project ID is not required, so we default to a dummy value.
	if emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); emulatorHost != "" {
		if projectID == "" {
			projectID = "demo-project"
		}
		tslog.Info(fmt.Sprintf("Initializing FirestoreEventRepository with emulator (Host: %s, Project: %s).", emulatorHost, projectID))
	} else if projectID != "" {
		tslog.Info(fmt.Sprintf("Initializing FirestoreEventRepository (Project: %s).", projectID))
	} else {
		// Neither GOOGLE_CLOUD_PROJECT nor FIRESTORE_EMULATOR_HOST is set —
		// we cannot proceed without a backend.
		tslog.Error(tserr.NotSet("FIRESTORE_EMULATOR_HOST and GOOGLE_CLOUD_PROJECT"))
		os.Exit(1)
	}
	// Initialize the FirestoreEventRepository with the project ID.
	repo, err := tseventserver.NewFirestoreEventRepository(context.Background(), projectID)
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
	api := tseventserver.NewEventServer(repo, tok)
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
