#!/bin/bash

# Copyright (c) 2026 thorsphere.
# All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
# that can be found in the LICENSE file.

# Exit immediately if any command fails
set -e

# Port the Firestore emulator will listen on (mapped to localhost)
PORT=8081
# Name for the Docker container (used to identify and clean up)
CONTAINER_NAME="firestore-emulator"

# Cleanup trap: ensures the emulator container is stopped and removed
# when the script exits, whether tests pass or fail.
cleanup() {
    echo "Stopping and removing container (${CONTAINER_NAME}) ..."
    docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1
}
trap cleanup EXIT

# Start the Firestore emulator inside a Docker container.
# The container is detached (-d) so the script can continue.
# Port ${PORT} inside the container is mapped to ${PORT} on localhost.
echo "Starting Firestore emulator in container on port ${PORT} ..."
docker run -d --name "${CONTAINER_NAME}" \
    -p "${PORT}:${PORT}" \
    google/cloud-sdk:emulators \
    gcloud beta emulators firestore start --host-port="0.0.0.0:${PORT}"

# Poll localhost up to 15 seconds until the emulator accepts connections.
# Firestore emulators don't expose a health-check endpoint, but they do
# bind to the port once fully initialised.
echo "Waiting for emulator to be ready..."
for i in $(seq 1 15); do
    if curl -s "http://localhost:${PORT}" >/dev/null 2>&1; then
        echo "Emulator is ready."
        break
    fi
    if [ $i -eq 10 ]; then
        echo "Emulator failed to start within 10 seconds."
        exit 1
    fi
    sleep 1
done

# The Firestore Go SDK reads this variable at runtime and automatically
# routes all Firestore requests to the emulator instead of production.
export FIRESTORE_EMULATOR_HOST="localhost:${PORT}"

# Run all tests in the module with verbose output.
# The emulator container will be cleaned up automatically afterwards
# via the trap above.
echo "Running tests..."
go test -v ./...