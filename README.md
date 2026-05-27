# tsecon

**tsecon** is a Go package for ingesting and querying economic calendar events.  
It provides an HTTP API backed by Google Cloud Firestore, with a token‑authenticated  
retrieval endpoint.

---

## Features

- **Ingestion** – `POST /events/ingest` accepts an array of economic events (JSON) and  
  upserts them into Firestore.
- **Retrieval** – `GET /events/retrieve?from=…&to=…` returns events within a time range,  
  secured by a bearer token.
- **Google Cloud Firestore** – production‑ready storage with automatic backend  
  detection via `GOOGLE_CLOUD_PROJECT` and `FIRESTORE_EMULATOR_HOST`.
- **Tested** – unit tests with high code coverage

---

## Getting Started

### Prerequisites

- **Go 1.26+**
- A **Firestore** database (or the **Firestore emulator** for local development)
- (Optional) **gcloud CLI** for deployment to Cloud Run

### Local Development with the Firestore Emulator

```bash
# Clone the repository
git clone https://github.com/your-org/tsecon.git
cd tsecon

# Start the Firestore emulator (in a separate terminal)
gcloud emulators firestore start --host-port=localhost:8081

# Set the emulator host and run the server
export FIRESTORE_EMULATOR_HOST=localhost:8081
go run ./cmd/main.go
```

Once the server is running, you can verify it with the local smoke test:

```bash
./smoke_test_local.sh
```

> **Note:** The emulator does not persist data between restarts.

### Build & Run the Binary

```bash
go build -o eventserver ./cmd/main.go
API_TOKEN=swordfish ./eventserver
```

The server listens on port `8080` by default (`PORT` env var overrides it).

---

## API Reference

### Authentication

Every request requires the `Authorization` header:

```
Authorization: Bearer <API_TOKEN>
```

The token is set via the `API_TOKEN` environment variable.  
If omitted, the default token is `swordfish` (local development only).

### Endpoints

#### `POST /events/ingest`

Ingest one or more economic events. Existing events (same time, country, name)
are updated automatically.

**Request body** – JSON array of event objects:

```json
[
  {
    "name": "US Core Inflation Rate",
    "time": "2026-05-15T12:30:00Z",
    "country": "US",
    "actual": 3.1,
    "estimate": 3.2,
    "previous": 3.3,
    "unit": "%",
    "impact": 3,
    "source": "US Bureau of Labor Statistics"
  }
]
```

**Responses**

| Status | Meaning |
|--------|---------|
| `200`  | Events ingested successfully |
| `400`  | Invalid JSON body |
| `401`  | Missing or invalid token |
| `500`  | Storage error |

#### `GET /events/retrieve`

Retrieve events within a time window.

**Query parameters**

| Param  | Required | Format           | Example                  |
|--------|----------|------------------|--------------------------|
| `from` | yes*     | RFC 3339 (UTC)   | `2026-05-01T00:00:00Z`   |
| `to`   | yes*     | RFC 3339 (UTC)   | `2026-05-31T23:59:59Z`   |

\* When `from` and `to` are omitted the endpoint returns events for the current UTC day.

**Example**

```bash
curl -H "Authorization: Bearer swordfish" \
  "http://localhost:8080/events/retrieve?from=2026-05-01T00:00:00Z&to=2026-05-31T23:59:59Z"
```

**Responses**

| Status | Meaning |
|--------|---------|
| `200`  | JSON array of events (may be empty) |
| `400`  | Invalid timestamp format |
| `401`  | Missing or invalid token |
| `500`  | Storage error |

---

## Deployment to Cloud Run

### Prerequisites

- A Firestore database named `eventdb` must exist in your Google Cloud project.
- The Cloud Run service account must have the **Datastore User** role to read and write to Firestore.

### Deploy

Copy and customise the deployment script, then run it:

```bash
cp deploy.sh.example deploy.sh
# Edit PROJECT_ID, REGION, SERVICE_NAME, and API_BEARER_KEY inside deploy.sh
chmod +x deploy.sh
./deploy.sh
```

The script sets `GOOGLE_CLOUD_PROJECT` and `API_TOKEN` environment variables
automatically. The server will detect Cloud Run and connect to Firestore.

Once deployed, you can verify the service with the smoke test script:

```bash
cp smoke_test.sh.example smoke_test.sh
# Replace the placeholder token inside smoke_test.sh
chmod +x smoke_test.sh
./smoke_test.sh https://your-service.run.app
```

---

## SQLite Support (Archived)

The SQLite‑backed `EventRepository` was used for prototyping and local
development before the project moved to a Firestore‑only backend.  
It has been removed from the main branch to keep the binary lean.

If you need the SQLite implementation for reference or local testing,
it is available in two places:

- **GitHub Gist** — [SQLite EventRepository for tsecon](https://gist.github.com/thorsphere/10bafc5eb7d4984bfa68402483530f4a)  
  Contains `repository_sqlite.go` and `repository_sqlite_test.go` with the
  full implementation and test suite.

- **Git tag `v1.0.0`** — The last release that includes the complete SQLite
  backend alongside the Firestore implementation.

---

## License

GNU Affero General Public License v3.0 – see [LICENSE](LICENSE) for details.
