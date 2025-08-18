# ZKID Single-Admin Dashboard Deployment Guide

## Overview

The ZKID Single-Admin Dashboard is a secure, privacy-focused web interface for monitoring the Zero-Knowledge Identity Layer (ZKID) of the Secure Email MVP system. This dashboard provides comprehensive operational visibility while maintaining zero-knowledge privacy guarantees.

## Features

### ✅ Core Dashboard Components
- **Feature Status Panel**: ZKID layer status, PQC integration, recovery system health
- **Endpoint Health Panel**: Real-time API endpoint performance and availability
- **Recovery Operations Panel**: Recovery code management and activity monitoring
- **Security & Compliance Panel**: Zero-knowledge guarantees and compliance status
- **Performance Metrics Panel**: System performance and latency monitoring
- **Alerts Panel**: Real-time alerting and notification management
- **Logs Panel**: UUID-only operational logs with search and filtering
- **Historical Trends Panel**: Performance analytics and trend analysis

### ✅ Security Features
- **Single-Admin Authentication**: Token-based admin authentication
- **Zero-Knowledge Compliance**: No external email addresses ever displayed
- **UUID-Only Operations**: All admin actions use internal UUIDs
- **Audit Logging**: Comprehensive logging with privacy protection
- **RBAC Integration**: Role-based access control enforcement
- **HTTPS Enforcement**: Secure communication protocols

## Prerequisites

### System Requirements
- **Node.js**: Version 18.0.0 or higher
- **npm/yarn**: Package manager for dependencies
- **Modern Browser**: Chrome, Firefox, Safari, or Edge (latest versions)
- **Network Access**: HTTPS access to ZKID backend API

### Backend Requirements
- **ZKID Layer**: Must be deployed and operational (v4.37+)
- **Admin Endpoints**: All ZKID admin endpoints must be accessible
- **CORS Configuration**: Backend must allow dashboard domain
- **Admin Token**: Valid admin authentication token

## Installation

### 1. Clone and Setup

```bash
# Navigate to the project directory
cd secure-email-mvp

# Install dependencies (if not already installed)
npm install

# Verify the dashboard components are present
ls src/components/admin/
```

### 2. Environment Configuration

Create or update your environment configuration:

```bash
# Create .env file for dashboard configuration
cat > .env << EOF
# ZKID Dashboard Configuration
VITE_API_BASE_URL=http://localhost:8080
VITE_DASHBOARD_TITLE=ZKID Admin Dashboard
VITE_REFRESH_INTERVAL=30000
VITE_ENABLE_WEBSOCKETS=false

# Security Configuration
VITE_ENFORCE_HTTPS=true
VITE_SESSION_TIMEOUT=3600000
VITE_MAX_LOGIN_ATTEMPTS=5
EOF
```

### 3. Build Configuration

Update `vite.config.js` to include dashboard-specific settings:

```javascript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: '0.0.0.0',
    https: process.env.VITE_ENFORCE_HTTPS === 'true'
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          dashboard: ['@heroicons/react', 'axios']
        }
      }
    }
  }
})
```

## Development Setup

### 1. Start Development Server

```bash
# Start the development server
npm run dev

# The dashboard will be available at:
# http://localhost:3000/admin
```

### 2. Access the Dashboard

1. Open your browser and navigate to `http://localhost:3000/admin`
2. Enter your admin token in the login form
3. The dashboard will load with real-time data from the ZKID backend

### 3. Development Features

- **Hot Reload**: Changes to dashboard components will automatically reload
- **Mock Data**: Dashboard uses mock data when backend is unavailable
- **Error Handling**: Comprehensive error handling and user feedback
- **Responsive Design**: Dashboard adapts to different screen sizes

## Production Deployment

### 1. Build for Production

```bash
# Build the production version
npm run build

# The built files will be in the `dist` directory
ls dist/
```

### 2. Web Server Configuration

#### Nginx Configuration

```nginx
server {
    listen 80;
    server_name zkid-dashboard.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name zkid-dashboard.yourdomain.com;

    # SSL Configuration
    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self' https://api.yourdomain.com;" always;

    # Root directory
    root /var/www/zkid-dashboard;
    index index.html;

    # Handle React Router
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy (if needed)
    location /api/ {
        proxy_pass https://api.yourdomain.com;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static assets caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

#### Apache Configuration

```apache
<VirtualHost *:80>
    ServerName zkid-dashboard.yourdomain.com
    Redirect permanent / https://zkid-dashboard.yourdomain.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName zkid-dashboard.yourdomain.com
    DocumentRoot /var/www/zkid-dashboard

    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /path/to/your/certificate.crt
    SSLCertificateKeyFile /path/to/your/private.key

    # Security Headers
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"
    Header always set X-Frame-Options DENY
    Header always set X-Content-Type-Options nosniff
    Header always set X-XSS-Protection "1; mode=block"
    Header always set Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self' https://api.yourdomain.com;"

    # React Router support
    RewriteEngine On
    RewriteCond %{REQUEST_FILENAME} !-f
    RewriteCond %{REQUEST_FILENAME} !-d
    RewriteRule ^(.*)$ /index.html [QSA,L]

    # Static assets caching
    <FilesMatch "\.(js|css|png|jpg|jpeg|gif|ico|svg)$">
        ExpiresActive On
        ExpiresDefault "access plus 1 year"
        Header set Cache-Control "public, immutable"
    </FilesMatch>
</VirtualHost>
```

### 3. Deployment Script

Create a deployment script for automated deployment:

```bash
#!/bin/bash
# deploy-zkid-dashboard.sh

set -e

echo "🚀 Deploying ZKID Dashboard..."

# Build the application
echo "📦 Building application..."
npm run build

# Create deployment directory
DEPLOY_DIR="/var/www/zkid-dashboard"
sudo mkdir -p $DEPLOY_DIR

# Copy built files
echo "📁 Copying files..."
sudo cp -r dist/* $DEPLOY_DIR/

# Set permissions
echo "🔐 Setting permissions..."
sudo chown -R www-data:www-data $DEPLOY_DIR
sudo chmod -R 755 $DEPLOY_DIR

# Restart web server
echo "🔄 Restarting web server..."
sudo systemctl reload nginx

echo "✅ ZKID Dashboard deployed successfully!"
echo "🌐 Dashboard available at: https://zkid-dashboard.yourdomain.com"
```

## Security Configuration

### 1. Admin Token Management

```bash
# Generate a secure admin token
openssl rand -hex 32

# Store securely (example for production)
export ZKID_ADMIN_TOKEN="your_generated_token_here"
```

### 2. Environment Variables

```bash
# Production environment variables
cat > .env.production << EOF
VITE_API_BASE_URL=https://api.yourdomain.com
VITE_DASHBOARD_TITLE=ZKID Admin Dashboard
VITE_REFRESH_INTERVAL=30000
VITE_ENFORCE_HTTPS=true
VITE_SESSION_TIMEOUT=3600000
VITE_MAX_LOGIN_ATTEMPTS=5
EOF
```

### 3. CORS Configuration

Ensure your ZKID backend allows the dashboard domain:

```go
// In your ZKID backend
corsMiddleware := cors.New(cors.Options{
    AllowedOrigins: []string{
        "https://zkid-dashboard.yourdomain.com",
        "http://localhost:3000", // For development
    },
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
})
```

## Monitoring and Maintenance

### 1. Health Checks

```bash
# Check dashboard availability
curl -I https://zkid-dashboard.yourdomain.com

# Check API connectivity
curl -H "Authorization: Bearer $ZKID_ADMIN_TOKEN" \
     https://api.yourdomain.com/api/admin/zkid/stats
```

### 2. Log Monitoring

```bash
# Monitor dashboard access logs
tail -f /var/log/nginx/access.log | grep zkid-dashboard

# Monitor error logs
tail -f /var/log/nginx/error.log | grep zkid-dashboard
```

### 3. Backup and Recovery

```bash
# Backup dashboard configuration
tar -czf zkid-dashboard-backup-$(date +%Y%m%d).tar.gz \
    /var/www/zkid-dashboard \
    /etc/nginx/sites-available/zkid-dashboard

# Restore from backup
tar -xzf zkid-dashboard-backup-20241201.tar.gz -C /
```

## Troubleshooting

### Common Issues

1. **Dashboard not loading**
   - Check browser console for errors
   - Verify API endpoint accessibility
   - Confirm CORS configuration

2. **Authentication failures**
   - Verify admin token format
   - Check token expiration
   - Confirm backend authentication

3. **Performance issues**
   - Monitor API response times
   - Check dashboard refresh intervals
   - Verify network connectivity

### Debug Mode

Enable debug mode for troubleshooting:

```bash
# Set debug environment variable
export VITE_DEBUG_MODE=true

# Rebuild and restart
npm run build
sudo systemctl reload nginx
```

## Compliance and Auditing

### 1. Audit Logging

The dashboard maintains comprehensive audit logs:

- All admin authentication attempts
- Dashboard access and navigation
- API calls and responses
- Error conditions and alerts

### 2. Privacy Compliance

- **Zero-Knowledge Guarantee**: No external emails displayed
- **UUID-Only Operations**: All identifiers are internal UUIDs
- **Data Minimization**: Only necessary data is collected
- **Audit Trail**: Complete logging for compliance

### 3. Security Standards

- **HTTPS Enforcement**: All communications encrypted
- **Content Security Policy**: XSS protection
- **Secure Headers**: Comprehensive security headers
- **Session Management**: Secure token handling

## Support and Documentation

### Resources

- **API Documentation**: `docs/api-documentation.md`
- **ZKID Implementation**: `docs/micro-iteration-4.37-summary.md`
- **Deployment Reports**: `docs/zkid-deployment-execution-report.md`

### Contact

For support and questions:
- Review the deployment documentation
- Check the troubleshooting guide
- Contact the development team

---

**ZKID Dashboard v4.37** - Production Ready • Zero-Knowledge Compliant
