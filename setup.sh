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
sudo apt-get install -y sqlite3 unzip ufw

# Install Go v1.23
echo "Installing Go v1.23..."
if ! command -v go &> /dev/null || ! go version | grep -q "go1.23"; then
    wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    source ~/.bashrc
    rm go1.23.0.linux-amd64.tar.gz
fi

# Create required directories
echo "Creating required directories..."
mkdir -p data logs certs

# Download GeoLite2 City database
echo "Downloading GeoLite2 City database..."
GEO_KEY=${GEO_KEY:-"YOUR_GEO_LICENSE_KEY"}
if [ ! -f data/GeoLite2-City.mmdb ]; then
    wget "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=$GEO_KEY&suffix=tar.gz" -O geo.tar.gz
    tar -xzf geo.tar.gz
    mv GeoLite2-City_*/GeoLite2-City.mmdb data/GeoLite2-City.mmdb
    rm -rf GeoLite2-City_* geo.tar.gz
fi

# Configure UFW firewall
echo "Configuring UFW firewall..."
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp  # SSH
sudo ufw allow 80/tcp  # HTTP
sudo ufw allow 443/tcp # HTTPS
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

echo "Setup completed at $(date)"
echo "Please check $LOG_FILE for details" 