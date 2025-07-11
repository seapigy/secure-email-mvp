#!/bin/bash

# vm_startup.sh: Start the backend server on Linux VM with detailed logging
set -e

echo "=== Secure Email MVP Backend Startup ==="
echo "Timestamp: $(date)"
echo "User: $(whoami)"
echo "Working directory: $(pwd)"

# Check if we're in the right directory
if [ ! -f "api-server" ]; then
    echo "ERROR: api-server binary not found in current directory"
    echo "Please run this script from /home/opc/secure-email-mvp/"
    exit 1
fi

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "ERROR: .env file not found"
    echo "Please create .env file with your Cloudflare R2 credentials"
    exit 1
fi

# Display environment variables (without secrets)
echo "=== Environment Check ==="
echo "CLOUDFLARE_R2_BUCKET: ${CLOUDFLARE_R2_BUCKET:-'NOT SET'}"
echo "CLOUDFLARE_R2_ENDPOINT: ${CLOUDFLARE_R2_ENDPOINT:-'NOT SET'}"
echo "SQLITE_DB: ${SQLITE_DB:-'NOT SET'}"
echo "LOG_FILE: ${LOG_FILE:-'NOT SET'}"

# Check file permissions
echo "=== File Permissions ==="
ls -la api-server
ls -la .env
ls -la schema/users.sql

# Check database directory
echo "=== Database Directory ==="
DB_DIR=$(dirname "${SQLITE_DB:-/var/db/secure-email.db}")
echo "Database directory: $DB_DIR"
ls -la "$DB_DIR" 2>/dev/null || echo "Database directory does not exist"

# Start the server with detailed logging
echo "=== Starting API Server ==="
echo "Server will log all startup details..."
echo "Press Ctrl+C to stop the server"
echo ""

# Run the server and capture all output
./api-server 2>&1 | tee /home/opc/api-startup.log

echo ""
echo "=== Server Stopped ==="
echo "Check /home/opc/api-startup.log for detailed logs" 