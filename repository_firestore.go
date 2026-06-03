// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tseventserver

// Import standard library packages and packages for context management, Firestore client, and custom error handling.
import (
	"context" // context for managing request contexts and timeouts
	"time"    // time for working with time and dates

	"cloud.google.com/go/firestore"  // firestore for interacting with Google Cloud Firestore
	"github.com/thorsphere/tserr"    // tserr for custom error handling
	"google.golang.org/api/iterator" // iterator for handling Firestore query results
)

const (
	dbID    = "eventdb"         // dbID is the name of the Firestore collection where economic events will be stored. This constant can be modified if you want to use a different collection name.
	colPath = "economic_events" // path is the Firestore collection path where economic events will be stored. This constant can be modified if you want to use a different collection name.
)

// FirestoreEventRepository is an implementation of the EventRepository interface that uses Google Cloud Firestore as the storage backend.
type FirestoreEventRepository struct {
	client *firestore.Client // client is the Firestore client used to interact with the Firestore database.
}

// NewFirestoreEventRepository creates a new instance of FirestoreEventRepository with the provided Google Cloud project ID.
func NewFirestoreEventRepository(ctx context.Context, projectID string) (*FirestoreEventRepository, error) {
	// Initialize a new Firestore client using the provided context and project ID.
	// If the client creation fails, return an error wrapped with tserr for better error handling.
	client, err := firestore.NewClientWithDatabase(ctx, projectID, dbID)
	if err != nil {
		// Wrap the error using tserr to provide additional context about the operation that failed.
		return nil, tserr.Op(&tserr.OpArgs{Op: "firestore.NewClient", Fn: "FirestoreEventRepository", Err: err})
	}
	// Attempt to connect to the Firestore database to ensure that the client is properly initialized and can communicate with the database.
	// This is especially important when using the Firestore emulator during local development, as it helps catch configuration issues early.
	pctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	// Perform a simple query to check if the Firestore client can successfully connect to the database.
	// We attempt to retrieve at least one document from the "economic_events" collection.
	_, err = client.Collection(colPath).Limit(1).Documents(pctx).GetAll()
	// If the query fails, it indicates that there is an issue with the Firestore client or the database connection,
	// so we close the client and return an error.
	if err != nil {
		client.Close()
		return nil, tserr.Op(&tserr.OpArgs{Op: "Ping", Fn: "FirestoreEventRepository", Err: err})
	}

	// If the client is successfully created, return a new instance of FirestoreEventRepository with the initialized client.
	return &FirestoreEventRepository{client: client}, nil
}

// Close closes the connection to the Firestore database and releases any associated resources.
func (r *FirestoreEventRepository) Close() error {
	// Check if the repository instance is nil
	if r == nil {
		return tserr.NilPtr()
	}
	// Check if the Firestore client is nil
	if r.client == nil {
		return tserr.NilPtr()
	}
	// Attempt to close the Firestore client and handle any errors that occur
	if err := r.client.Close(); err != nil {
		return tserr.Op(&tserr.OpArgs{
			Op:  "client.Close",
			Fn:  "FirestoreEventRepository",
			Err: err,
		})
	}
	// Return nil to indicate success
	return nil
}

// Store saves a new economic event to the FirestoreEventRepository.
// It generates a deterministic document ID to seamlessly perform an upsert.
func (r *FirestoreEventRepository) Store(ctx context.Context, event *Event) error {
	// Validate input parameters and return appropriate errors if any of them are nil.
	if r == nil {
		return tserr.NilPtr()
	}
	if r.client == nil {
		return tserr.NilPtr()
	}
	if event == nil {
		return tserr.NilPtr()
	}
	// Generate the deterministic ID (acts as your composite unique constraint)
	docID := event.GenerateDocID()
	// firestore.MergeAll performs the Upsert.
	// If the doc exists, it updates differing fields. If not, it creates it.
	_, err := r.client.Collection(colPath).Doc(docID).Set(ctx, event)
	if err != nil {
		return tserr.Op(&tserr.OpArgs{Op: "firestore.Set", Fn: event.Name, Err: err})
	}
	// Return nil to indicate success
	return nil
}

// GetByPeriod retrieves economic events that occurred within a specified time period.
func (r *FirestoreEventRepository) GetByPeriod(ctx context.Context, period *Period) ([]Event, error) {
	// Validate input parameters and return appropriate errors if any of them are nil.
	if r == nil {
		return nil, tserr.NilPtr()
	}
	if r.client == nil {
		return nil, tserr.NilPtr()
	}
	if period == nil {
		return nil, tserr.NilPtr()
	}
	// Initialize an empty slice to hold the retrieved events.
	var events []Event
	// Query the Firestore collection "economic_events" for documents where the "Time" field is between the specified period.
	// Firestore queries are inclusive, so we use >= for the start and <= for the end of the period.
	// Note: Firestore will look for the struct field name "Time" by default, unless you provide firestore tags.
	iter := r.client.Collection(colPath).
		Where("Time", ">=", period.From.UTC()).
		Where("Time", "<=", period.To.UTC()).
		Documents(ctx)
		// Iterate over the query results and map each Firestore document back to an Event struct.
	for {
		// Next() retrieves the next document in the query results.
		doc, err := iter.Next()
		// It returns an error when there are no more documents (iterator.Done) or if there is an issue with the query.
		if err == iterator.Done {
			break // Reached the end of the query results
		}
		if err != nil {
			return nil, tserr.Op(&tserr.OpArgs{Op: "iter.Next", Fn: "FirestoreEventRepository", Err: err})
		}
		// Create a new Event struct to hold the data from the Firestore document.
		var ev Event
		// DataTo automatically maps the Firestore document fields back to your Event struct
		if err := doc.DataTo(&ev); err != nil {
			return nil, tserr.Op(&tserr.OpArgs{Op: "doc.DataTo", Fn: ev.Name, Err: err})
		}
		// Append the mapped Event to the slice of events to be returned.
		events = append(events, ev)
	}
	// Return the slice of events that match the specified time period and nil for the error.
	return events, nil
}
