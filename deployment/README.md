# Secure Email Backend - Production Deployment Guide

This guide provides comprehensive instructions for deploying the Secure Email Backend to Oracle Cloud Infrastructure with full production setup.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Cloudflare    │    │   Oracle Cloud  │    │   MySQL DB      │
│   (DNS + SSL)   │───▶│   (Backend)     │───▶│   (OCI MySQL)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📋 Prerequisites

- Oracle Cloud Infrastructure account
- GitHub repository with code
- Domain name configured with Cloudflare
- SSH key pair for server access

## 🚀 Quick Start

### 1. Initial Server Setup

```bash
# Connect to your Oracle Cloud instance
ssh ubuntu@your-server-ip

# Run the setup script
sudo bash /path/to/deployment/setup.sh
```

### 2. Configure Environment

```bash
# Edit the environment file
sudo nano /opt/secure-email/.env

# Update the following variables:
DB_DSN=mysql://secureuser:YOUR_DB_PASSWORD@tcp(YOUR_DB_HOST:3306)/securesystem?parseTime=true
SMTP_HOST=your-smtp-host
SMTP_USER=your-smtp-user
SMTP_PASS=your-smtp-password
EMAIL_FROM=no-reply@yourdomain.com
FRONTEND_URL=https://yourdomain.com
```

### 3. Deploy Application

```bash
# Run the deployment script
cd /opt/secure-email
sudo ./deployment/deploy.sh
```

### 4. Test Production

```bash
# Run production tests
sudo ./deployment/test-production.sh
```

## 🔧 Detailed Setup Instructions

### Oracle Cloud Infrastructure Setup

1. **Create Infrastructure with Terraform**
   ```bash
   cd deployment/oci
   terraform init
   terraform plan
   terraform apply
   ```

2. **Configure OCI CLI**
   ```bash
   oci setup config
   ```

3. **Set up GitHub Secrets**
   - `OCI_CONFIG`: Base64 encoded OCI config file
   - `OCI_PRIVATE_KEY`: Base64 encoded private key
   - `OCI_INSTANCE_ID`: Your compute instance ID
   - `OCI_COMPARTMENT_ID`: Your compartment ID

### Database Setup

1. **MySQL Configuration**
   - Automatic migration on startup
   - Secure user with limited privileges
   - Audit logging enabled
   - Backup automation

2. **Database Security**
   - No root remote access
   - Application-specific user
   - Encrypted connections
   - Regular backups

### Security Configuration

1. **Run Security Hardening**
   ```bash
   sudo ./deployment/security/hardening.sh
   ```

2. **Configure Firewall**
   - Only ports 22, 80, 443 open
   - Fail2ban for intrusion prevention
   - Rate limiting on API endpoints

3. **SSL/TLS Setup**
   - Cloudflare handles SSL termination
   - Security headers configured
   - HSTS enabled

## 📊 Monitoring and Maintenance

### Health Monitoring

```bash
# Manual health check
sudo ./deployment/monitoring/healthcheck.sh

# Automated monitoring (runs every 5 minutes)
crontab -l
```

### Log Management

```bash
# View application logs
docker-compose -f docker-compose.prod.yml logs -f app

# View system logs
tail -f /opt/secure-email/logs/health.log
tail -f /opt/secure-email/logs/deploy.log
```

### Backup Management

```bash
# Manual backup
sudo ./deployment/database/backup.sh

# Automated backups (daily at 2 AM)
crontab -l
```

## 🔄 CI/CD Pipeline

The GitHub Actions pipeline automatically:

1. **On Push to Main:**
   - Runs tests
   - Builds Docker image
   - Pushes to GitHub Container Registry
   - Deploys to Oracle Cloud
   - Runs database migrations
   - Performs health checks

2. **Zero-Downtime Deployment:**
   - Rolling updates
   - Health checks before traffic routing
   - Automatic rollback on failure

## 🧪 Testing

### API Endpoints

- **Health Check:** `GET /health`
- **Signup:** `POST /api/signup`

### Test Signup

```bash
curl -X POST https://yourdomain.com/api/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@securesystem.email",
    "password": "password123",
    "tier": "free"
  }'
```

### Expected Response

```json
{
  "status": "ok",
  "message": "Signup successful. A verification email has been sent. Save your recovery token securely.",
  "recovery_token": "USER-VISUAL-TOKEN-HERE",
  "recovery_token_qr_data_uri": "data:image/png;base64,...."
}
```

## 🛠️ Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check database status
   docker ps | grep mysql
   
   # Check database logs
   docker logs securechat-email-db-1
   ```

2. **Backend Not Starting**
   ```bash
   # Check backend logs
   docker-compose -f docker-compose.prod.yml logs app
   
   # Restart service
   sudo systemctl restart secure-email-backend.service
   ```

3. **SSL Issues**
   ```bash
   # Check SSL certificate
   openssl x509 -in /opt/secure-email/deployment/nginx/ssl/cert.pem -text -noout
   ```

### Log Locations

- Application logs: `/opt/secure-email/logs/`
- System logs: `/var/log/`
- Docker logs: `docker-compose logs`

## 📈 Performance Optimization

### Database Optimization

- Indexed columns for fast queries
- Connection pooling
- Query optimization
- Regular maintenance

### Application Optimization

- Docker multi-stage builds
- Nginx caching
- Gzip compression
- Resource limits

## 🔒 Security Best Practices

1. **Regular Updates**
   - System packages
   - Docker images
   - Dependencies

2. **Monitoring**
   - Failed login attempts
   - Unusual traffic patterns
   - Resource usage

3. **Backup Strategy**
   - Daily database backups
   - Configuration backups
   - Off-site storage

## 📞 Support

For issues or questions:

1. Check logs first
2. Run health checks
3. Review this documentation
4. Check GitHub Issues

## 🎯 Success Criteria

After deployment, you should be able to:

- ✅ Access API at `https://yourdomain.com/api/signup`
- ✅ Create new users via API
- ✅ See users in production database
- ✅ Receive verification emails
- ✅ Monitor system health
- ✅ Deploy updates via `git push`

## 📝 Maintenance Schedule

- **Daily:** Health checks, log rotation
- **Weekly:** Security updates, backup verification
- **Monthly:** Performance review, security audit
- **Quarterly:** Full system backup, disaster recovery test
