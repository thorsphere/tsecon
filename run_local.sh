#!/bin/bash

# Copyright (c) 2026 thorsphere.
# All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
# that can be found in the LICENSE file.

set -euo pipefail

# ------------------------------------------------------------------
# Configuration
# ------------------------------------------------------------------
PORT="${PORT:-8080}"
API_TOKEN="${API_TOKEN:-swordfish}"
EMULATOR_PORT="${EMULATOR_PORT:-8085}"
PROJECT_ID="demo-project"

API_URL="http://localhost:${PORT}"

# ------------------------------------------------------------------
# Helper: start Firestore emulator if not already listening
# ------------------------------------------------------------------
start_emulator() {
    if curl -s "localhost:${EMULATOR_PORT}" &> /dev/null; then
        echo "=> Firestore emulator already running on port ${EMULATOR_PORT}"
        return 0
    fi

    echo "=> Starting Firestore emulator via Docker on port ${EMULATOR_PORT}..."
    docker run -d --name firestore-emulator \
        -p "${EMULATOR_PORT}:${EMULATOR_PORT}" \
        google/cloud-sdk:emulators \
        gcloud beta emulators firestore start --host-port="0.0.0.0:${EMULATOR_PORT}"

    # Wait until the emulator accepts connections
    for i in $(seq 1 15); do
        if curl -s "localhost:${EMULATOR_PORT}" &> /dev/null; then
            echo "=> Emulator is ready."
            return 0
        fi
        sleep 1
    done

    echo "ERROR: Emulator did not start in time." >&2
    exit 1
}

# ------------------------------------------------------------------
# Cleanup on exit / interrupt
# ------------------------------------------------------------------
cleanup() {
    echo ""
    echo "=> Shutting down..."
    # Kill the Go server if we started it
    if [[ -n "${SERVER_PID:-}" ]]; then
        kill "${SERVER_PID}" 2>/dev/null || true
        wait "${SERVER_PID}" 2>/dev/null || true
    fi
    # Stop and remove the Docker container if we started it
    if docker ps -q -f name=firestore-emulator | grep -q .; then
        docker stop firestore-emulator >/dev/null 2>&1 || true
        docker rm firestore-emulator >/dev/null 2>&1 || true
    fi
    echo "=> Done."
}
trap cleanup EXIT INT TERM

# ------------------------------------------------------------------
# Main
# ------------------------------------------------------------------
echo "================================================="
echo "Local Development Runner"
echo "================================================="

# 1. Start emulator
start_emulator
export FIRESTORE_EMULATOR_HOST="localhost:${EMULATOR_PORT}"
export GOOGLE_CLOUD_PROJECT="${PROJECT_ID}"

# 2. Start the microservice (background)
echo "=> Starting EventServer on port ${PORT}..."
go run cmd/main.go &
SERVER_PID=$!

# Wait until the server is reachable
for i in $(seq 1 15); do
    if curl -s "${API_URL}/events/retrieve" \
        -H "Authorization: Bearer ${API_TOKEN}" &> /dev/null; then
        echo "=> Server is ready."
        break
    fi
    sleep 1
done

# 3. Run the smoke test
echo ""
echo "================================================="
echo "Running Smoke Test against ${API_URL}"
echo "================================================="

# ---- POST Non-Farm Payrolls ----
echo ""
echo "=> [POST] Creating 'Non-Farm Payrolls' Event..."
curl -s -X POST "${API_URL}/events/ingest" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${API_TOKEN}" \
    -d '[
        {
            "name": "Non-Farm Payrolls",
            "time": "2024-07-05T08:30:00Z",
            "country": "US",
            "actual": 200.0,
            "estimate": 180.0,
            "previous": 150.0,
            "unit": "K",
            "impact": 3,
            "source": "Bureau of Labor Statistics"
        }
    ]' | jq || echo "Failed or jq not installed, raw output above."

# ---- POST GDP Growth Rate ----
echo ""
echo "=> [POST] Creating 'GDP Growth Rate' Event..."
curl -s -X POST "${API_URL}/events/ingest" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${API_TOKEN}" \
    -d '[{
        "name": "GDP Growth Rate",
        "time": "2024-07-10T08:30:00Z",
        "country": "US",
        "actual": 3.5,
        "estimate": 3.0,
        "previous": 2.8,
        "unit": "%",
        "impact": 2,
        "source": "Bureau of Economic Analysis"
    }]' | jq || echo "Failed or jq not installed, raw output above."

# ---- GET events for July 2024 ----
echo ""
echo "=> [GET] Retrieving events for July 2024..."
curl -s -X GET "${API_URL}/events/retrieve?from=2024-07-01T00:00:00Z&to=2024-07-31T23:59:59Z" \
    -H "Authorization: Bearer ${API_TOKEN}" | jq || echo "Failed or jq not installed."

echo ""
echo "================================================="
echo "Smoke test complete!"
echo "================================================="
echo ""
echo "Server is still running. Press Ctrl+C to stop."
# Keep the script alive so the server stays up until you Ctrl+C
wait "${SERVER_PID}"
