# tseventserver

[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/thorsphere/tseventserver)](https://pkg.go.dev/mod/github.com/thorsphere/tseventserver)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/thorsphere/tseventserver)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/thorsphere/tseventserver)
![GitHub Top Language](https://img.shields.io/github/languages/top/thorsphere/tseventserver)
[![CodeFactor](https://www.codefactor.io/repository/github/thorsphere/tseventserver/badge)](https://www.codefactor.io/repository/github/thorsphere/tseventserver)
![OSS Lifecycle](https://img.shields.io/osslifecycle/thorsphere/tseventserver)

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
- **Docker** (for running the emulator via script)
- (Optional) **gcloud CLI** for deployment to Cloud Run

### Local Development with the Firestore Emulator

**Option A – One‑command runner (recommended)**

The `run_local.sh` script starts the Firestore emulator (via Docker), launches
the server, and runs a quick smoke test — all in one terminal:

```bash
./run_local.sh
```

Press `Ctrl+C` to stop everything and clean up.

**Option B – Manual steps**

1. Start the Firestore emulator:

   ```bash
   docker run -d --name firestore-emulator -p 8081:8081 \
     google/cloud-sdk:emulators \
     gcloud beta emulators firestore start --host-port=0.0.0.0:8081
   ```

2. Set the required environment variables and run the server:

   ```bash
   export FIRESTORE_EMULATOR_HOST=localhost:8081
   export GOOGLE_CLOUD_PROJECT=demo-project
   go run ./cmd/main.go
   ```

3. In another terminal, run the smoke test against the local server:

   ```bash
   ./smoke_test.sh http://localhost:8080
   ```

> **Note:** The emulator does not persist data between restarts.

### Build & Run the Binary

```bash
go build -o eventserver ./cmd/main.go
API_TOKEN=swordfish ./eventserver
```

The server listens on port `8080` by default (`PORT` env var overrides it).

---

## Running Tests

The unit tests (nil-pointer checks, event formatting) run without any setup:

```bash
go test -v ./...
```

The Firestore integration tests (`TestFirestoreStoreAndGetByPeriod` and
`TestFirestoreGetByPeriodEmpty`) require a running emulator and will fail with
clear instructions if one is not found. Use the convenience script to run them:

```bash
./test_firestore.sh
```

This starts a Firestore emulator container, runs all tests (including integration), and
cleans up afterwards. If you prefer to run the emulator manually:

```bash
docker run -d --name firestore-emulator -p 8081:8081 \
    google/cloud-sdk:emulators \
    gcloud beta emulators firestore start --host-port=0.0.0.0:8081

export FIRESTORE_EMULATOR_HOST=localhost:8081
go test -v ./...

docker rm -f firestore-emulator
```

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
    "name": "Non-Farm Payrolls",
    "description": "The non-farm payrolls index measures the number of jobs added or lost in the economy.",
    "time": "2026-05-15T12:30:00Z",
    "country": "US",
    "actual": 272.0,
    "estimate": 180.0,
    "previous": 236.0,
    "unit": "K",
    "precision": 1,
    "change": 36.0,
    "change_pct": 0.15,
    "surprise": 92.0,
    "surprise_pct": 0.51,
    "impact": 3,
    "source": "Bureau of Labor Statistics"
  }
]
```

**Event fields**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Name of the economic event, e.g., "Non-Farm Payrolls", "GDP Growth Rate" |
| `description` | string | no | Description of the economic event |
| `time` | string | yes | Date and time of the event in RFC 3339 UTC format |
| `country` | string | yes | ISO 3166-1 alpha-2 two-letter country code |
| `currency` | string | no | Currency of the values, e.g., "USD", "EUR" |
| `actual` | number | no | Actual value released |
| `estimate` | number | no | Estimated/forecasted value |
| `previous` | number | no | Previous period's value |
| `unit` | string | no | Unit of measurement, e.g., "%", "K", "M", "B" |
| `precision` | int | no | Number of decimal places for rounding |
| `change` | number | no | Change from previous value |
| `change_pct` | number | no | Percentage change from previous value |
| `surprise` | number | no | Difference between actual and estimate |
| `surprise_pct` | number | no | Percentage surprise relative to estimate |
| `impact` | int | no | Impact level of the event (1-3) |
| `source` | string | yes | Source of the data, e.g., "Bureau of Labor Statistics" |

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

## Documentation & Resources

- [Go Package Documentation](https://pkg.go.dev/github.com/thorsphere/tseventserver) — Complete API reference
- [Open Source Insights](https://deps.dev/go/github.com%2Fthorsphere%2Ftseventserver) — Dependency analysis

---

## ⚖️ License & Commercial Usage

Copyright (c) 2026 thorsphere. All rights reserved.

This project is licensed under the **Functional Source License v1.1 (FSL-1.1-ALv2)**. 

* The use, modification, and redistribution of this Go package is completely free for private, educational, non-commercial, and internal purposes. 
* If you are a company or institution looking to use this package in a commercial product, service, or business environment, you must secure a commercial license.
* Each version of this software automatically converts to the fully open-source Apache License, Version 2.0 on the second anniversary of its release.

For full details, please see the [LICENSE](LICENSE) file.

### 💼 Commercial Licensing & Inquiries

To purchase a commercial license or discuss support options, please reach out directly:

* 📩 **Contact:** business at thorsphere dot com
* 💬 **Response Time:** Usually within a couple of business days.

*Please include your company name and a brief overview of your use case so I can provide the right licensing details.*
