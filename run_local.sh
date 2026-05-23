#!/bin/bash

# Default to sqlite if no argument is provided
MODE=${1:-sqlite}

# Shared variables
export PORT="8080"
export API_TOKEN="swordfish"

if [ "$MODE" = "firestore" ]; then
    echo "=> Configuring for local Firestore Emulator..."
    export FIRESTORE_EMULATOR_HOST="localhost:8085"
    export GOOGLE_CLOUD_PROJECT="demo-project"
else
    echo "=> Configuring for local SQLite..."
    unset FIRESTORE_EMULATOR_HOST
    unset GOOGLE_CLOUD_PROJECT
fi

echo "=> Starting the microservice..."
go run cmd/main.go

