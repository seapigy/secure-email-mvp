#!/bin/bash

# test_connect.sh: Verify infrastructure setup
echo "Testing infrastructure components..."

# Function to check command status
check_status() {
    if [ $? -eq 0 ]; then
        echo "✅ $1"
    else
        echo "❌ $1"
    fi
}

# Test Go installation
echo "Testing Go..."
if command -v go &> /dev/null; then
    go version
    check_status "Go installation"
else
    echo "❌ Go not found"
fi

# Test Node.js installation
echo "Testing Node.js..."
if command -v node &> /dev/null; then
    node --version
    npm --version
    check_status "Node.js installation"
else
    echo "❌ Node.js not found"
fi

# Test SQLite
echo "Testing SQLite..."
if command -v sqlite3 &> /dev/null; then
    sqlite3 --version
    check_status "SQLite installation"
else
    echo "❌ SQLite not found"
fi

# Test database connection
echo "Testing database connection..."
if [ -f "data/secure_email.db" ]; then
    sqlite3 data/secure_email.db "SELECT 1;" &> /dev/null
    check_status "Database connection"
else
    echo "❌ Database file not found"
fi

# Test SSL certificates
echo "Testing SSL certificates..."
if [ -f "certs/cert.pem" ] && [ -f "certs/key.pem" ]; then
    echo "✅ SSL certificates found"
else
    echo "❌ SSL certificates not found"
fi

# Test UFW
echo "Testing UFW..."
if command -v ufw &> /dev/null; then
    sudo ufw status | grep -q "Status: active"
    check_status "UFW firewall"
else
    echo "❌ UFW not found"
fi

# Test SSH configuration
echo "Testing SSH configuration..."
if [ -f "/etc/ssh/sshd_config" ]; then
    grep -q "PermitRootLogin no" /etc/ssh/sshd_config
    check_status "SSH root login disabled"
    grep -q "PasswordAuthentication no" /etc/ssh/sshd_config
    check_status "SSH password authentication disabled"
else
    echo "❌ SSH config not found"
fi

# Test environment variables
echo "Testing environment variables..."
if [ -f ".env" ]; then
    echo "✅ .env file found"
    # Check required variables
    required_vars=("CLOUDFLARE_API_KEY" "CLOUDFLARE_R2_ACCESS_KEY" "CLOUDFLARE_R2_SECRET_KEY" "R2_BUCKET_NAME" "GEO_KEY")
    for var in "${required_vars[@]}"; do
        if grep -q "^$var=" .env; then
            echo "✅ $var is set"
        else
            echo "❌ $var is not set"
        fi
    done
else
    echo "❌ .env file not found"
fi

echo "Infrastructure tests completed" 