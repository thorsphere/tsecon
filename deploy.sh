#!/usr/bin/env bash

# Copyright (c) 2026 thorsphere.
# All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
# that can be found in the LICENSE file.

# Exit immediately if a command exits with a non-zero status
set -euo pipefail

# 1. CONFIGURATION
PROJECT_ID="${PROJECT_ID}"
REGION="${REGION:-us-east4}"
SERVICE_NAME="${SERVICE_NAME}"
API_BEARER_KEY="${API_BEARER_KEY:-your-super-secret-api-token}" 

echo "Starting deployment for '$SERVICE_NAME' in region '$REGION'..."

# 2. DEPLOYMENT
gcloud run deploy "$SERVICE_NAME" \
  --source . \
  --region "$REGION" \
  --project "$PROJECT_ID" \
  --allow-unauthenticated \
  --set-env-vars="GOOGLE_CLOUD_PROJECT=$PROJECT_ID,API_TOKEN=$API_BEARER_KEY"

echo "Fetching final service URL..."
# Using the grep/awk workaround to safely extract the URL
SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --project "$PROJECT_ID" \
  | grep -E "^URL:" | awk '{print $2}')

# Ensure Firestore database exists (idempotent operation)
echo "Ensuring Firestore database 'eventdb' exists..."
if ! gcloud firestore databases describe --database=eventdb --project="$PROJECT_ID" &>/dev/null; then
    echo "Database 'eventdb' does not exist. Creating..."
    gcloud firestore databases create \
      --database=eventdb \
      --location="$REGION" \
      --project="$PROJECT_ID" \
      --type=firestore-native
else
    echo " ↳ Database 'eventdb' already exists, skipping."
fi

echo -e "\n✅ Deployment successful! Service URL is: $SERVICE_URL\n"

# 3. NEXT STEPS
echo "To run the smoke test against this deployment, use the following command:"
echo "export API_BEARER_KEY=\"$API_BEARER_KEY\""
echo "./smoke_test.sh.example \"$SERVICE_URL\""
