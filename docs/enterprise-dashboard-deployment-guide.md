# Single-Admin Enterprise Dashboard Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying the Single-Admin Enterprise Dashboard in production environments. The dashboard provides secure, zero-knowledge monitoring for the entire Secure Email MVP system with enhanced authentication and compliance features.

## Prerequisites

### System Requirements
- **Node.js**: Version 18.0.0 or higher
- **npm**: Version 8.0.0 or higher
- **Web Server**: Nginx or Apache for production serving
- **SSL Certificate**: Valid TLS certificate for HTTPS
- **Database**: Access to the Secure Email MVP database
- **Network**: Secure network access to backend APIs

### Security Requirements
- **Firewall**: Restrict access to admin dashboard endpoints
- **VPN Access**: Secure VPN for admin access (recommended)
- **Monitoring**: System monitoring and alerting setup
- **Backup**: Regular backup procedures for configuration

## Installation

### 1. Clone and Setup
```bash
# Clone the repository
git clone <repository-url>
cd secure-email-mvp

# Install dependencies
npm install

# Verify installation
npm run build
```

### 2. Environment Configuration
Create environment configuration files:

```bash
# Production environment
cp .env.example .env.production
```

Configure the following environment variables:

```env
# API Configuration
VITE_API_BASE_URL=https://api.secure-email-mvp.com
VITE_WS_BASE_URL=wss://api.secure-email-mvp.com

# Security Configuration
VITE_ADMIN_DASHBOARD_ENABLED=true
VITE_MFA_ENABLED=true
VITE_SESSION_TIMEOUT_MINUTES=30
VITE_MAX_FAILED_ATTEMPTS=5
VITE_LOCKOUT_DURATION_MINUTES=15

# Feature Flags
VITE_ZKID_ENABLED=true
VITE_PQC_ENABLED=true
VITE_ENTERPRISE_ENABLED=true

# Monitoring Configuration
VITE_AUTO_REFRESH_ENABLED=true
VITE_REFRESH_INTERVAL_SECONDS=30
VITE_REAL_TIME_UPDATES_ENABLED=true
```

### 3. Build Configuration
Update `vite.config.ts` for production:

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'terser',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          ui: ['@heroicons/react'],
          utils: ['axios']
        }
      }
    }
  },
  server: {
    port: 3000,
    https: true
  }
})
```

## Production Deployment

### 1. Build Process
```bash
# Production build
npm run build

# Verify build output
ls -la dist/
```

### 2. Web Server Configuration

#### Nginx Configuration
Create `/etc/nginx/sites-available/enterprise-dashboard`:

```nginx
server {
    listen 443 ssl http2;
    server_name dashboard.secure-email-mvp.com;

    # SSL Configuration
    ssl_certificate /path/to/ssl/certificate.crt;
    ssl_certificate_key /path/to/ssl/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self' wss:; frame-ancestors 'none';" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Root directory
    root /var/www/enterprise-dashboard;
    index index.html;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Main application
    location / {
        try_files $uri $uri/ /index.html;
        
        # Security headers for application
        add_header X-Frame-Options DENY always;
        add_header X-Content-Type-Options nosniff always;
    }

    # API proxy (if needed)
    location /api/ {
        proxy_pass https://api.secure-email-mvp.com;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Security headers for API
        add_header X-Frame-Options DENY always;
        add_header X-Content-Type-Options nosniff always;
    }

    # Health check endpoint
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }

    # Deny access to sensitive files
    location ~ /\. {
        deny all;
    }

    location ~ \.(env|log|conf)$ {
        deny all;
    }
}

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name dashboard.secure-email-mvp.com;
    return 301 https://$server_name$request_uri;
}
```

#### Apache Configuration
Create `/etc/apache2/sites-available/enterprise-dashboard.conf`:

```apache
<VirtualHost *:443>
    ServerName dashboard.secure-email-mvp.com
    DocumentRoot /var/www/enterprise-dashboard

    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /path/to/ssl/certificate.crt
    SSLCertificateKeyFile /path/to/ssl/private.key
    SSLCertificateChainFile /path/to/ssl/chain.crt

    # Security Headers
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"
    Header always set X-Frame-Options DENY
    Header always set X-Content-Type-Options nosniff
    Header always set X-XSS-Protection "1; mode=block"
    Header always set Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self' wss:; frame-ancestors 'none';"
    Header always set Referrer-Policy "strict-origin-when-cross-origin"

    # Compression
    <IfModule mod_deflate.c>
        AddOutputFilterByType DEFLATE text/plain
        AddOutputFilterByType DEFLATE text/html
        AddOutputFilterByType DEFLATE text/xml
        AddOutputFilterByType DEFLATE text/css
        AddOutputFilterByType DEFLATE application/xml
        AddOutputFilterByType DEFLATE application/xhtml+xml
        AddOutputFilterByType DEFLATE application/rss+xml
        AddOutputFilterByType DEFLATE application/javascript
        AddOutputFilterByType DEFLATE application/x-javascript
    </IfModule>

    # Cache static assets
    <FilesMatch "\.(js|css|png|jpg|jpeg|gif|ico|svg)$">
        ExpiresActive On
        ExpiresDefault "access plus 1 year"
        Header set Cache-Control "public, immutable"
    </FilesMatch>

    # Main application
    <Directory /var/www/enterprise-dashboard>
        Options -Indexes +FollowSymLinks
        AllowOverride All
        Require all granted
        
        # SPA routing
        RewriteEngine On
        RewriteCond %{REQUEST_FILENAME} !-f
        RewriteCond %{REQUEST_FILENAME} !-d
        RewriteRule . /index.html [L]
    </Directory>

    # Health check
    <Location /health>
        SetHandler application/x-httpd-php
        Require all granted
    </Location>

    # Deny access to sensitive files
    <FilesMatch "\.(env|log|conf)$">
        Require all denied
    </FilesMatch>

    <Files ~ "^\.">
        Require all denied
    </Files>
</VirtualHost>

# HTTP to HTTPS redirect
<VirtualHost *:80>
    ServerName dashboard.secure-email-mvp.com
    Redirect permanent / https://dashboard.secure-email-mvp.com/
</VirtualHost>
```

### 3. Deployment Script
Create a deployment script `deploy.sh`:

```bash
#!/bin/bash

# Deployment script for Enterprise Dashboard
set -e

echo "Starting Enterprise Dashboard deployment..."

# Configuration
DEPLOY_PATH="/var/www/enterprise-dashboard"
BACKUP_PATH="/var/backups/enterprise-dashboard"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create backup
echo "Creating backup..."
if [ -d "$DEPLOY_PATH" ]; then
    mkdir -p "$BACKUP_PATH"
    cp -r "$DEPLOY_PATH" "$BACKUP_PATH/backup_$TIMESTAMP"
fi

# Build application
echo "Building application..."
npm run build

# Deploy to production
echo "Deploying to production..."
sudo rm -rf "$DEPLOY_PATH"
sudo mkdir -p "$DEPLOY_PATH"
sudo cp -r dist/* "$DEPLOY_PATH/"

# Set permissions
echo "Setting permissions..."
sudo chown -R www-data:www-data "$DEPLOY_PATH"
sudo chmod -R 755 "$DEPLOY_PATH"

# Restart web server
echo "Restarting web server..."
sudo systemctl reload nginx

# Health check
echo "Performing health check..."
sleep 5
if curl -f -s https://dashboard.secure-email-mvp.com/health > /dev/null; then
    echo "Deployment successful!"
else
    echo "Deployment failed - health check failed"
    exit 1
fi

echo "Enterprise Dashboard deployment completed successfully!"
```

## Security Configuration

### 1. Firewall Configuration
```bash
# Configure UFW firewall
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP (redirect)
sudo ufw allow 443/tcp   # HTTPS
sudo ufw deny 3000/tcp   # Development port
sudo ufw enable
```

### 2. SSL Certificate Management
```bash
# Install Certbot for Let's Encrypt
sudo apt install certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot --nginx -d dashboard.secure-email-mvp.com

# Set up auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### 3. Security Monitoring
Create security monitoring script `security-monitor.sh`:

```bash
#!/bin/bash

# Security monitoring for Enterprise Dashboard
LOG_FILE="/var/log/enterprise-dashboard-security.log"

# Check for failed login attempts
check_failed_logins() {
    local failed_count=$(grep "Failed login attempt" /var/log/auth.log | wc -l)
    if [ $failed_count -gt 10 ]; then
        echo "$(date): High number of failed login attempts: $failed_count" >> "$LOG_FILE"
        # Send alert
        echo "Security Alert: High failed login attempts" | mail -s "Dashboard Security Alert" admin@secure-email-mvp.com
    fi
}

# Check SSL certificate expiry
check_ssl_certificate() {
    local cert_file="/etc/letsencrypt/live/dashboard.secure-email-mvp.com/fullchain.pem"
    local expiry_date=$(openssl x509 -enddate -noout -in "$cert_file" | cut -d= -f2)
    local expiry_epoch=$(date -d "$expiry_date" +%s)
    local current_epoch=$(date +%s)
    local days_until_expiry=$(( (expiry_epoch - current_epoch) / 86400 ))
    
    if [ $days_until_expiry -lt 30 ]; then
        echo "$(date): SSL certificate expires in $days_until_expiry days" >> "$LOG_FILE"
        echo "SSL Certificate Alert: Expires in $days_until_expiry days" | mail -s "SSL Certificate Alert" admin@secure-email-mvp.com
    fi
}

# Check disk space
check_disk_space() {
    local usage=$(df /var/www/enterprise-dashboard | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ $usage -gt 80 ]; then
        echo "$(date): High disk usage: ${usage}%" >> "$LOG_FILE"
        echo "Disk Space Alert: ${usage}% usage" | mail -s "Disk Space Alert" admin@secure-email-mvp.com
    fi
}

# Run all checks
check_failed_logins
check_ssl_certificate
check_disk_space
```

### 4. Backup Configuration
Create backup script `backup.sh`:

```bash
#!/bin/bash

# Backup script for Enterprise Dashboard
BACKUP_DIR="/var/backups/enterprise-dashboard"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup application files
tar -czf "$BACKUP_DIR/app_$TIMESTAMP.tar.gz" -C /var/www enterprise-dashboard

# Backup configuration files
tar -czf "$BACKUP_DIR/config_$TIMESTAMP.tar.gz" \
    /etc/nginx/sites-available/enterprise-dashboard \
    /etc/ssl/certs/dashboard.secure-email-mvp.com.crt \
    /etc/ssl/private/dashboard.secure-email-mvp.com.key

# Clean up old backups (keep last 7 days)
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete

echo "Backup completed: $TIMESTAMP"
```

## Monitoring & Alerting

### 1. System Monitoring
Create system monitoring script `system-monitor.sh`:

```bash
#!/bin/bash

# System monitoring for Enterprise Dashboard
LOG_FILE="/var/log/enterprise-dashboard-system.log"

# Check application health
check_application_health() {
    local response=$(curl -s -o /dev/null -w "%{http_code}" https://dashboard.secure-email-mvp.com/health)
    if [ "$response" != "200" ]; then
        echo "$(date): Application health check failed: HTTP $response" >> "$LOG_FILE"
        echo "Application Health Alert: HTTP $response" | mail -s "Application Health Alert" admin@secure-email-mvp.com
    fi
}

# Check memory usage
check_memory_usage() {
    local memory_usage=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
    if [ $memory_usage -gt 80 ]; then
        echo "$(date): High memory usage: ${memory_usage}%" >> "$LOG_FILE"
        echo "Memory Usage Alert: ${memory_usage}%" | mail -s "Memory Usage Alert" admin@secure-email-mvp.com
    fi
}

# Check CPU usage
check_cpu_usage() {
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    if (( $(echo "$cpu_usage > 80" | bc -l) )); then
        echo "$(date): High CPU usage: ${cpu_usage}%" >> "$LOG_FILE"
        echo "CPU Usage Alert: ${cpu_usage}%" | mail -s "CPU Usage Alert" admin@secure-email-mvp.com
    fi
}

# Run all checks
check_application_health
check_memory_usage
check_cpu_usage
```

### 2. Log Monitoring
Configure log monitoring in `/etc/logrotate.d/enterprise-dashboard`:

```
/var/log/enterprise-dashboard-*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 www-data www-data
    postrotate
        systemctl reload nginx
    endscript
}
```

## Operational Procedures

### 1. Startup Procedure
```bash
# Start Enterprise Dashboard
sudo systemctl start nginx
sudo systemctl enable nginx

# Verify startup
sudo systemctl status nginx
curl -f https://dashboard.secure-email-mvp.com/health
```

### 2. Shutdown Procedure
```bash
# Graceful shutdown
sudo systemctl stop nginx

# Verify shutdown
sudo systemctl status nginx
```

### 3. Maintenance Mode
Create maintenance mode script `maintenance.sh`:

```bash
#!/bin/bash

# Maintenance mode script for Enterprise Dashboard
MAINTENANCE_FILE="/var/www/enterprise-dashboard/maintenance.html"

if [ "$1" = "enable" ]; then
    echo "Enabling maintenance mode..."
    cat > "$MAINTENANCE_FILE" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Maintenance - Enterprise Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .maintenance { background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 8px; padding: 40px; }
    </style>
</head>
<body>
    <div class="maintenance">
        <h1>🛠️ Maintenance in Progress</h1>
        <p>The Enterprise Dashboard is currently undergoing scheduled maintenance.</p>
        <p>We expect to be back online shortly. Thank you for your patience.</p>
    </div>
</body>
</html>
EOF
    echo "Maintenance mode enabled"
elif [ "$1" = "disable" ]; then
    echo "Disabling maintenance mode..."
    rm -f "$MAINTENANCE_FILE"
    echo "Maintenance mode disabled"
else
    echo "Usage: $0 {enable|disable}"
    exit 1
fi
```

### 4. Emergency Procedures
Create emergency procedures document:

```bash
# Emergency contact information
EMERGENCY_CONTACTS="admin@secure-email-mvp.com"

# Emergency shutdown
emergency_shutdown() {
    echo "EMERGENCY SHUTDOWN INITIATED"
    sudo systemctl stop nginx
    sudo ufw deny 443/tcp
    echo "Emergency shutdown completed"
}

# Emergency recovery
emergency_recovery() {
    echo "EMERGENCY RECOVERY INITIATED"
    sudo systemctl start nginx
    sudo ufw allow 443/tcp
    echo "Emergency recovery completed"
}
```

## Troubleshooting

### Common Issues

1. **SSL Certificate Issues**
   ```bash
   # Check certificate validity
   openssl x509 -in /etc/ssl/certs/dashboard.secure-email-mvp.com.crt -text -noout
   
   # Renew certificate
   sudo certbot renew --force-renewal
   ```

2. **Permission Issues**
   ```bash
   # Fix permissions
   sudo chown -R www-data:www-data /var/www/enterprise-dashboard
   sudo chmod -R 755 /var/www/enterprise-dashboard
   ```

3. **Nginx Configuration Issues**
   ```bash
   # Test configuration
   sudo nginx -t
   
   # Reload configuration
   sudo systemctl reload nginx
   ```

4. **Application Health Issues**
   ```bash
   # Check application logs
   sudo tail -f /var/log/nginx/error.log
   
   # Check system resources
   htop
   df -h
   free -h
   ```

### Performance Optimization

1. **Enable Gzip Compression**
2. **Configure Browser Caching**
3. **Optimize Static Assets**
4. **Monitor Resource Usage**

## Conclusion

This deployment guide provides comprehensive instructions for deploying the Single-Admin Enterprise Dashboard in production. Follow all security configurations and monitoring procedures to ensure a secure and reliable deployment.

For additional support or questions, refer to the implementation documentation or contact the development team.
