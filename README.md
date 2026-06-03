# tseventserver

[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/thorsphere/tseventserver)](https://pkg.go.dev/mod/github.com/thorsphere/tseventserver)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/thorsphere/tseventserver)

[![Go Report Card](https://goreportcard.com/badge/github.com/thorsphere/tseventserver)](https://goreportcard.com/report/github.com/thorsphere/tseventserver)
[![CodeFactor](https://www.codefactor.io/repository/github/thorsphere/tseventserver/badge)](https://www.codefactor.io/repository/github/thorsphere/tseventserver)
![OSS Lifecycle](https://img.shields.io/osslifecycle/thorsphere/tseventserver)
![Libraries.io dependency status for GitHub repo](https://img.shields.io/librariesio/github/thorsphere/tseventserver)

![GitHub release (latest by date)](https://img.shields.io/github/v/release/thorsphere/tseventserver)
![GitHub last commit](https://img.shields.io/github/last-commit/thorsphere/tseventserver)
![GitHub commit activity](https://img.shields.io/github/commit-activity/m/thorsphere/tseventserver)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/thorsphere/tseventserver)
![GitHub Top Language](https://img.shields.io/github/languages/top/thorsphere/tseventserver)
![GitHub](https://img.shields.io/github/license/thorsphere/tseventserver)

---

**tseventserver** is a Go package for ingesting and querying economic calendar events.  
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
git clone https://github.com/your-org/tseventserver.git
cd tseventserver

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
```

With this (using `run_local.sh` for the all‑in‑one approach, and also showing the manual way):

```markdown
### Local Development with the Firestore Emulator

**Option A – One‑command runner (recommended)**

The `run_local.sh` script starts the Firestore emulator (via Docker), launches
the server, and runs a quick smoke test — all in one terminal:

```bash
./run_local.sh
```

Press `Ctrl+C` to stop everything and clean up.

**Option B – Manual steps**

1. Start the Firestore emulator (requires Docker or `gcloud`).  
   The easiest way is with Docker:

   ```bash
   docker run -d --name firestore-emulator -p 8085:8085 \
     google/cloud-sdk:emulators \
     gcloud beta emulators firestore start --host-port=0.0.0.0:8085
   ```

2. Set the required environment variables and run the server:

   ```bash
   export FIRESTORE_EMULATOR_HOST=localhost:8085
   export GOOGLE_CLOUD_PROJECT=demo-project
   go run ./cmd/main.go
   ```

3. In another terminal, run the smoke test against the local server:

   ```bash
   ./smoke_test.sh http://localhost:8080
   ```

> **Note:** The emulator does not persist data between restarts.
```

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

The included `deploy.sh` script handles the complete deployment. It reads
configuration from environment variables (with sensible defaults for some).

**Required variables**

| Variable | Purpose |
|----------|---------|
| `PROJECT_ID` | Your Google Cloud project ID |
| `SERVICE_NAME` | Cloud Run service name (e.g., `eventserver`) |
| `API_BEARER_KEY` | The token used to secure your API (set via `API_TOKEN` on the server) |
| `REGION` | (optional) Deployment region, defaults to `us-east4` |

**Example**

```bash
export PROJECT_ID=thorsphere-trading-stg
export SERVICE_NAME=eventserver
export API_BEARER_KEY=your-super-secret-token
./deploy.sh
```

The script sets `GOOGLE_CLOUD_PROJECT` and `API_TOKEN` automatically,
ensures the `eventdb` Firestore database exists, and prints the service URL
at the end.

**Automated deployments** are handled by the GitHub Actions workflow
(`.github/workflows/deploy.yml`), which uses the same `deploy.sh` script.

Once deployed, verify with the smoke test:

```bash
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
