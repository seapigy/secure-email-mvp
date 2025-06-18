#!/bin/bash

echo "Testing system connections..."

# Test SQLite
echo "Testing SQLite..."
if command -v sqlite3 &> /dev/null; then
    sqlite3 --version
else
    echo "SQLite not found"
fi

# Test Go
echo "Testing Go..."
if command -v go &> /dev/null; then
    go version
else
    echo "Go not found"
fi

# Test Node.js
echo "Testing Node.js..."
if command -v node &> /dev/null; then
    node --version
    npm --version
else
    echo "Node.js not found"
fi

# Test database connection
echo "Testing database connection..."
if [ -f "data/secure_email.db" ]; then
    sqlite3 data/secure_email.db "SELECT 1;" &> /dev/null
    if [ $? -eq 0 ]; then
        echo "Database connection successful"
    else
        echo "Database connection failed"
    fi
else
    echo "Database file not found"
fi

# Test SSL certificates
echo "Testing SSL certificates..."
if [ -f "certs/cert.pem" ] && [ -f "certs/key.pem" ]; then
    echo "SSL certificates found"
else
    echo "SSL certificates not found"
fi

echo "Connection tests completed" 