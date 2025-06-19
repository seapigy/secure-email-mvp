#!/bin/bash

# setup.sh: Configure Oracle Cloud VM for secure email system
set -e
LOG_FILE=/var/log/setup.log
exec &>> $LOG_FILE

echo "Starting setup at $(date)"

# Update system
echo "Updating system packages..."
sudo apt-get update -y
sudo apt-get upgrade -y

# Install required packages
echo "Installing required packages..."
sudo apt-get install -y sqlite3 golang-go

# Install Go v1.23
echo "Installing Go v1.23..."
if ! command -v go &> /dev/null || ! go version | grep -q "go1.23"; then
    wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
    source ~/.bashrc
    rm go1.23.0.linux-amd64.tar.gz
fi

# Create required directories
echo "Creating required directories..."
mkdir -p data logs certs

# Configure UFW firewall
echo "Configuring UFW firewall..."
if [ "$(hostname)" == "api-vm" ]; then
  sudo ufw allow 22
  sudo ufw allow 80
  sudo ufw allow 443
else
  sudo ufw allow 22
fi
sudo ufw --force enable

# Disable root SSH
echo "Configuring SSH security..."
sudo sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sudo sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo systemctl restart sshd

# Create self-signed certificate for development
echo "Generating self-signed certificate..."
if [ ! -f certs/cert.pem ] || [ ! -f certs/key.pem ]; then
    openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes -subj "/CN=localhost"
fi

# Initialize SQLite database
echo "Initializing SQLite database..."
if [ ! -f data/secure_email.db ]; then
    sqlite3 data/secure_email.db < schema.sql
fi

# Set up environment
echo "Setting up environment..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "Please update .env with your configuration"
fi

# Note: Geolocation is handled by HTML5 Geolocation API in the browser
# and OpenStreetMap Nominatim for reverse geocoding (no setup needed)

echo "Setup completed at $(date)"
echo "Please check $LOG_FILE for details" 