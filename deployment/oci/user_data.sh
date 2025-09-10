#!/bin/bash

# Secure Email Backend Setup Script
set -e

# Update system
apt-get update
apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
usermod -aG docker ubuntu

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Install additional tools
apt-get install -y git curl wget unzip

# Create application directory
mkdir -p /opt/secure-email
cd /opt/secure-email

# Clone the repository (will be updated by CI/CD)
git clone https://github.com/seapigy/secure-email-mvp.git .

# Create production environment file
cat > .env << EOF
# Database Configuration
DB_DSN=mysql://secureuser:${DB_PASSWORD}@tcp(${db_host}:${db_port})/securesystem?parseTime=true

# Email Configuration
SMTP_HOST=${SMTP_HOST}
SMTP_PORT=${SMTP_PORT}
SMTP_USER=${SMTP_USER}
SMTP_PASS=${SMTP_PASS}
EMAIL_FROM=${EMAIL_FROM}
FRONTEND_URL=${FRONTEND_URL}

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

# Create systemd service for the backend
cat > /etc/systemd/system/secure-email-backend.service << EOF
[Unit]
Description=Secure Email Backend
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/secure-email
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
systemctl enable secure-email-backend.service

# Create log rotation
cat > /etc/logrotate.d/secure-email << EOF
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

# Set up firewall
ufw --force enable
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp

# Create monitoring script
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
echo "*/5 * * * * /opt/secure-email/monitor.sh" | crontab -

# Create backup script
cat > /opt/secure-email/backup.sh << 'EOF'
#!/bin/bash

# Database backup script
BACKUP_DIR="/opt/secure-email/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup database
docker exec securechat-email-db-1 mysqldump -u secureuser -p"${DB_PASSWORD}" securesystem > "$BACKUP_DIR/db_backup_$DATE.sql"

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name "db_backup_*.sql" -mtime +7 -delete
EOF

chmod +x /opt/secure-email/backup.sh

# Add daily backup cron job
echo "0 2 * * * /opt/secure-email/backup.sh" | crontab -

# Wait for database to be ready
echo "Waiting for database to be ready..."
sleep 60

# Start the backend service
systemctl start secure-email-backend.service

echo "Secure Email Backend setup completed!"
