#!/bin/bash

# Create necessary directories
mkdir -p data
mkdir -p logs
mkdir -p certs

# Initialize Go module
go mod init secure-email-mvp
go mod tidy

# Install frontend dependencies
cd src
npm install
cd ..

# Create SQLite database
sqlite3 data/secure_email.db < schema.sql

# Set up environment
cp .env.example .env

# Create self-signed certificate for development
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes -subj "/CN=localhost"

echo "Setup completed successfully!" 