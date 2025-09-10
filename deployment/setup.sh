#!/bin/bash

# Initial Production Setup Script for Secure Email Backend
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEPLOYMENT_DIR="/opt/secure-email"
REPO_URL="https://github.com/seapigy/secure-email-mvp.git"

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    error "Please run as root"
fi

log "Starting Secure Email Backend production setup..."

# Update system
log "Updating system packages..."
apt-get update
apt-get upgrade -y

# Install required packages
log "Installing required packages..."
apt-get install -y \
    curl \
    wget \
    git \
    unzip \
    htop \
    ufw \
    fail2ban \
    logrotate \
    mysql-client

# Install Docker
log "Installing Docker..."
if ! command -v docker >/dev/null 2>&1; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    usermod -aG docker ubuntu
    rm get-docker.sh
    success "Docker installed"
else
    log "Docker already installed"
fi

# Install Docker Compose
log "Installing Docker Compose..."
if ! command -v docker-compose >/dev/null 2>&1; then
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    success "Docker Compose installed"
else
    log "Docker Compose already installed"
fi

# Create deployment directory
log "Creating deployment directory..."
mkdir -p "$DEPLOYMENT_DIR"
cd "$DEPLOYMENT_DIR"

# Clone repository
log "Cloning repository..."
if [ ! -d ".git" ]; then
    git clone "$REPO_URL" .
    success "Repository cloned"
else
    git pull origin main
    log "Repository updated"
fi

# Create environment file template
log "Creating environment file template..."
if [ ! -f .env ]; then
    cat > .env << 'EOF'
# Database Configuration
DB_DSN=mysql://secureuser:CHANGE_ME@tcp(localhost:3306)/securesystem?parseTime=true

# Email Configuration
SMTP_HOST=CHANGE_ME
SMTP_PORT=587
SMTP_USER=CHANGE_ME
SMTP_PASS=CHANGE_ME
EMAIL_FROM=no-reply@securesystem.email
FRONTEND_URL=https://securemail.example.com

# Token & Security Configuration
RECOVERY_TOKEN_EXP_DAYS=7
VERIFICATION_TOKEN_EXP_HOURS=24

# Argon2 Password Hashing Configuration
ARGON2_MEMORY_KB=131072
ARGON2_ITERATIONS=4
ARGON2_PARALLELISM=4
ARGON2_SALT_LEN=16
ARGON2_KEY_LEN=32

# Server Configuration
PORT=8080
HOST=0.0.0.0

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Environment
ENV=production
DEBUG=false
EOF
    warning "Please edit .env file with your configuration"
fi

# Set up firewall
log "Configuring firewall..."
ufw --force enable
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw deny 3306/tcp   # MySQL (block external access)
success "Firewall configured"

# Set up fail2ban
log "Configuring fail2ban..."
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 10
EOF

systemctl enable fail2ban
systemctl restart fail2ban
success "Fail2ban configured"

# Create systemd service
log "Creating systemd service..."
cat > /etc/systemd/system/secure-email-backend.service << EOF
[Unit]
Description=Secure Email Backend
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$DEPLOYMENT_DIR
ExecStart=/usr/local/bin/docker-compose -f docker-compose.prod.yml up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose.prod.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable secure-email-backend.service
success "Systemd service created"

# Create monitoring script
log "Creating monitoring script..."
cat > /opt/secure-email/monitor.sh << 'EOF'
#!/bin/bash

# Health check script
BACKEND_URL="http://localhost:8080/health"
LOG_FILE="/opt/secure-email/logs/health.log"

# Create logs directory if it doesn't exist
mkdir -p /opt/secure-email/logs

# Check backend health
if curl -f -s "$BACKEND_URL" > /dev/null; then
    echo "$(date): Backend is healthy" >> "$LOG_FILE"
else
    echo "$(date): Backend health check failed" >> "$LOG_FILE"
    # Restart the service
    systemctl restart secure-email-backend.service
fi
EOF

chmod +x /opt/secure-email/monitor.sh

# Add cron job for health monitoring
log "Setting up health monitoring cron job..."
echo "*/5 * * * * /opt/secure-email/monitor.sh" | crontab -

# Create backup script
log "Creating backup script..."
cat > /opt/secure-email/backup.sh << 'EOF'
#!/bin/bash

# Database backup script
BACKUP_DIR="/opt/secure-email/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup database (if MySQL is running locally)
if docker ps | grep -q mysql; then
    docker exec securechat-email-db-1 mysqldump -u secureuser -p"${DB_PASSWORD}" securesystem > "$BACKUP_DIR/db_backup_$DATE.sql"
fi

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name "db_backup_*.sql" -mtime +7 -delete
EOF

chmod +x /opt/secure-email/backup.sh

# Add daily backup cron job
log "Setting up backup cron job..."
echo "0 2 * * * /opt/secure-email/backup.sh" | crontab -

# Set up log rotation
log "Setting up log rotation..."
cat > /etc/logrotate.d/secure-email << 'EOF'
/opt/secure-email/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 ubuntu ubuntu
}
EOF

# Make deployment script executable
chmod +x deployment/deploy.sh
chmod +x deployment/database/migrate.sh

success "Production setup completed!"

log "Next steps:"
log "1. Edit /opt/secure-email/.env with your configuration"
log "2. Run: cd /opt/secure-email && ./deployment/deploy.sh"
log "3. Configure your domain DNS to point to this server"
log "4. Set up SSL certificates in /opt/secure-email/deployment/nginx/ssl/"

warning "Remember to:"
warning "- Change default passwords"
warning "- Configure SMTP settings"
warning "- Set up SSL certificates"
warning "- Update DNS records"
