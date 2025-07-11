#!/bin/bash

# vm_directory_check.sh: Generate complete directory tree and file listing for VM analysis
set -e

echo "=== Secure Email MVP VM Directory Analysis ==="
echo "Timestamp: $(date)"
echo "VM Hostname: $(hostname)"
echo "VM IP: $(hostname -I)"
echo ""

# Change to project directory
cd /home/opc/secure-email-mvp

echo "=== Current Directory ==="
pwd
echo ""

echo "=== Directory Tree ==="
tree -a -I '.git' || find . -type f -name ".*" -o -type f | head -50

echo ""
echo "=== File Permissions and Sizes ==="
ls -la

echo ""
echo "=== Go Module Information ==="
if [ -f "go.mod" ]; then
    cat go.mod
else
    echo "go.mod not found"
fi

echo ""
echo "=== Environment File (without secrets) ==="
if [ -f ".env" ]; then
    echo "File exists, size: $(wc -c < .env) bytes"
    echo "First few lines (sanitized):"
    head -5 .env | sed 's/=.*/=***HIDDEN***/'
else
    echo ".env file not found"
fi

echo ""
echo "=== Database Status ==="
DB_PATH="${SQLITE_DB:-/var/db/secure-email.db}"
if [ -f "$DB_PATH" ]; then
    echo "Database exists: $DB_PATH"
    echo "Database size: $(wc -c < "$DB_PATH") bytes"
    echo "Database tables:"
    sqlite3 "$DB_PATH" ".tables" 2>/dev/null || echo "Could not read database"
else
    echo "Database not found: $DB_PATH"
fi

echo ""
echo "=== Process Status ==="
ps aux | grep api-server | grep -v grep || echo "No api-server process running"

echo ""
echo "=== Network Status ==="
netstat -tlnp | grep :8080 || echo "No process listening on port 8080"

echo ""
echo "=== Disk Space ==="
df -h /home/opc

echo ""
echo "=== Memory Usage ==="
free -h

echo ""
echo "=== Analysis Complete ===" 