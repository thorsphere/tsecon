// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package main

// Import necessary packages for context handling, OS signal processing, time management,
// and the tsecon, tserr, and tslog packages from the thorsphere project.
import (
	"context" // context
	"fmt"
	"net/http"
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
	// Define the address and port on which the EventServer will listen for incoming HTTP requests.
	// This can be changed to use a different port or bind to a specific IP address.
	serverAddr = ":8080"
)

// The main function serves as the entry point for the application.
// It initializes the event repository, starts the event server,
// and handles graceful shutdown on interrupt signals.
func main() {
	// Initialize your repository.
	repo, err := tsecon.NewSQLiteEventRepository(dbPath)
	if err != nil {
		tslog.Error(tserr.Op(&tserr.OpArgs{Op: "New SQLite Event Repository", Fn: dbPath, Err: err}))
		os.Exit(1)
	}
	// Create a new EventServer instance with the initialized repository.
	api := tsecon.NewEventServer(repo)
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
