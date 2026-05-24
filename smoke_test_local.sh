#!/bin/bash
# A simple locally executed smoke test for the standard EventServer API

API_URL="http://localhost:8080"
TOKEN="swordfish" # Matches default fallback in main.go

echo "================================================="
echo "Starting Smoke Test against $API_URL"
echo "================================================="

# 1. Create a new event (Upsert)
echo -e "\n=> [POST] Creating 'Non-Farm Payrolls' Event..."
curl -s -X POST "$API_URL/events/ingest" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
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

# 2. Create another event (GDP)
echo -e "\n=> [POST] Creating 'GDP Growth Rate' Event..."
curl -s -X POST "$API_URL/events/ingest" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
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

# 3. Retrieve events by Period
echo -e "\n=> [GET] Retrieving events for July 2024..."
# Adjust the query parameters based on exactly how your request parsing is configured in the handler
curl -s -X GET "$API_URL/events/retrieve?from=2024-07-01T00:00:00Z&to=2024-07-31T23:59:59Z" \
  -H "Authorization: Bearer $TOKEN" | jq || echo "Failed or jq not installed."

echo -e "\n================================================="
echo "✅ Smoke test complete!"
echo "================================================="
