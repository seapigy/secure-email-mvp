#!/bin/bash
set -e

# test_connect.sh: Verify infrastructure setup
source .env

echo "Testing VM connectivity..."
ping -c 4 $API_HOST

echo "Testing SQLite..."
sqlite3 --version

echo "Testing Cloudflare R2..."
curl -I "https://$CLOUDFLARE_R2_BUCKET.r2.cloudflarestorage.com"

echo "Testing Nominatim API..."
curl -s "$NOMINATIM_URL?lat=40.7128&lon=-74.0060&format=json" | grep -q "New York"

echo "All tests passed!" 