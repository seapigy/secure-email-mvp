#!/bin/bash

# Production Deployment Script for Secure Email Backend
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEPLOYMENT_DIR="/opt/secure-email"
BACKUP_DIR="/opt/secure-email/backups"
LOG_FILE="/opt/secure-email/logs/deploy.log"

# Create necessary directories
mkdir -p "$BACKUP_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    error "Please run as root or with sudo"
fi

log "Starting Secure Email Backend deployment..."

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    error "Docker is not running"
fi

# Check if docker-compose is available
if ! command -v docker-compose >/dev/null 2>&1; then
    error "docker-compose is not installed"
fi

# Backup current deployment
log "Creating backup of current deployment..."
if [ -d "$DEPLOYMENT_DIR" ]; then
    BACKUP_NAME="backup_$(date +%Y%m%d_%H%M%S)"
    tar -czf "$BACKUP_DIR/$BACKUP_NAME.tar.gz" -C "$DEPLOYMENT_DIR" . 2>/dev/null || warning "Could not create backup"
    log "Backup created: $BACKUP_NAME.tar.gz"
fi

# Stop current services
log "Stopping current services..."
cd "$DEPLOYMENT_DIR" || error "Deployment directory not found"
docker-compose -f docker-compose.prod.yml down || warning "Could not stop services"

# Pull latest changes
log "Pulling latest changes from repository..."
git fetch origin main
git reset --hard origin/main

# Update environment variables if needed
if [ ! -f .env ]; then
    error "Environment file .env not found"
fi

# Run database migrations
log "Running database migrations..."
chmod +x deployment/database/migrate.sh
./deployment/database/migrate.sh

# Build and start services
log "Building and starting services..."
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d --build

# Wait for services to be ready
log "Waiting for services to be ready..."
sleep 30

# Health check
log "Performing health check..."
MAX_ATTEMPTS=10
ATTEMPT=1

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if curl -f -s http://localhost:8080/health >/dev/null 2>&1; then
        success "Health check passed"
        break
    else
        warning "Health check attempt $ATTEMPT failed, retrying in 10 seconds..."
        sleep 10
        ATTEMPT=$((ATTEMPT + 1))
    fi
done

if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    error "Health check failed after $MAX_ATTEMPTS attempts"
fi

# Test API endpoint
log "Testing API endpoint..."
if curl -f -s -X POST http://localhost:8080/api/signup \
    -H "Content-Type: application/json" \
    -d '{"email":"test@securesystem.email","password":"testpass123","tier":"free"}' >/dev/null 2>&1; then
    success "API endpoint test passed"
else
    warning "API endpoint test failed"
fi

# Clean up old Docker images
log "Cleaning up old Docker images..."
docker image prune -f

# Clean up old backups (keep last 7 days)
log "Cleaning up old backups..."
find "$BACKUP_DIR" -name "backup_*.tar.gz" -mtime +7 -delete

# Set up log rotation
log "Setting up log rotation..."
cat > /etc/logrotate.d/secure-email << EOF
$LOG_FILE {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 root root
}
EOF

success "Deployment completed successfully!"

# Display service status
log "Service status:"
docker-compose -f docker-compose.prod.yml ps

log "Deployment log saved to: $LOG_FILE"
