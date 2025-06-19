#!/bin/bash
set -e

# setup_api.sh: Setup script for Secure Email API on Oracle Cloud VM1
# This script installs dependencies and configures the API environment

echo "=== Secure Email API Setup ==="

# Update system
echo "Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install Go 1.23 if not already installed
if ! command -v go &> /dev/null; then
    echo "Installing Go 1.23..."
    wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    source ~/.bashrc
    rm go1.23.0.linux-amd64.tar.gz
else
    echo "Go is already installed: $(go version)"
fi

# Install SQLite3
echo "Installing SQLite3..."
sudo apt install -y sqlite3

# Create application directory
echo "Creating application directory..."
sudo mkdir -p /opt/secure-email-mvp
sudo chown $USER:$USER /opt/secure-email-mvp

# Create database directory
echo "Setting up database..."
sudo mkdir -p /var/db
sudo chown $USER:$USER /var/db

# Create log directory
echo "Setting up logging..."
sudo mkdir -p /var/log
sudo touch /var/log/api.log
sudo chown $USER:$USER /var/log/api.log

# Copy application files (assuming they're in the current directory)
echo "Copying application files..."
cp -r . /opt/secure-email-mvp/
cd /opt/secure-email-mvp

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy

# Apply database schema
echo "Setting up database schema..."
sqlite3 /var/db/secure-email.db < schema/users.sql

# Generate JWT secret if not exists
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cp env.example .env
    
    # Generate JWT secret
    JWT_SECRET=$(openssl rand -base64 32)
    sed -i "s/your_32_byte_jwt_secret_here_generate_with_openssl_rand_base64_32/$JWT_SECRET/" .env
    
    echo "Generated JWT secret and updated .env file"
    echo "Please edit .env file with your Cloudflare R2 credentials"
fi

# Create systemd service
echo "Creating systemd service..."
sudo tee /etc/systemd/system/secure-email-api.service > /dev/null <<EOF
[Unit]
Description=Secure Email API
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/opt/secure-email-mvp
Environment=PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ExecStart=/usr/local/go/bin/go run cmd/api/main.go
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
echo "Enabling and starting API service..."
sudo systemctl daemon-reload
sudo systemctl enable secure-email-api
sudo systemctl start secure-email-api

# Check service status
echo "Checking service status..."
sudo systemctl status secure-email-api --no-pager

# Test API
echo "Testing API endpoint..."
sleep 5
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/auth/login || echo "000"

echo ""
echo "=== Setup Complete ==="
echo "API is running on port 8080"
echo "Service: secure-email-api"
echo "Logs: /var/log/api.log"
echo "Database: /var/db/secure-email.db"
echo ""
echo "Next steps:"
echo "1. Edit /opt/secure-email-mvp/.env with your Cloudflare R2 credentials"
echo "2. Restart the service: sudo systemctl restart secure-email-api"
echo "3. Test the API: curl -X POST http://localhost:8080/api/auth/login"
echo "4. Configure Cloudflare proxy for api.securesystem.email" 