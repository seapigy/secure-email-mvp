# Infrastructure Documentation

## System Architecture

### Components

1. **Frontend (Netlify)**
   - React 18 application
   - Tailwind CSS for styling
   - Vite for build tooling
   - Environment variables for configuration

2. **Backend (Oracle Cloud Free Tier)**
   - Go 1.23 API server
   - SQLite database
   - Cloudflare R2 for encrypted storage

3. **Security**
   - TLS 1.3 for transport security
   - AES-256-GCM for encryption
   - Argon2 for password hashing
   - ECDSA for digital signatures

### Directory Structure

```
.
├── cmd/
│   └── api/          # Backend entry point
├── pkg/              # Backend packages
│   ├── auth/         # Authentication
│   ├── email/        # Email handling
│   └── geolocation/  # Location services
├── src/              # Frontend source
│   ├── components/   # React components
│   ├── styles/       # CSS and Tailwind
│   └── lib/          # Frontend utilities
├── docs/             # Documentation
├── tests/            # Test files
├── data/             # Database
├── logs/             # Application logs
└── certs/            # SSL certificates
```

### Development Setup

1. **Prerequisites**
   - Node.js v20
   - Go v1.23
   - Git
   - SQLite3

2. **Environment Variables**
   - Copy `.env.example` to `.env`
   - Update values for your environment

3. **Database**
   - SQLite database in `data/secure_email.db`
   - Schema in `schema.sql`

4. **Security**
   - Development certificates in `certs/`
   - Production certificates from Cloudflare

### Deployment

1. **Frontend**
   - Build with `npm run build`
   - Deploy to Netlify
   - Configure environment variables

2. **Backend**
   - Build with `go build`
   - Deploy to Oracle Cloud
   - Configure firewall rules

3. **Storage**
   - Configure Cloudflare R2
   - Set up CORS policies
   - Configure access keys

### Monitoring

1. **Logs**
   - Application logs in `logs/`
   - Error tracking
   - Access logs

2. **Security**
   - Failed login attempts
   - Access violations
   - System health

### Backup

1. **Database**
   - Daily SQLite backups
   - Retention: 30 days

2. **Storage**
   - R2 versioning enabled
   - Retention: 30 days

### Maintenance

1. **Updates**
   - Weekly dependency updates
   - Monthly security patches
   - Quarterly major updates

2. **Cleanup**
   - Expired secure links
   - Old logs
   - Temporary files

# Infrastructure Setup Guide

## Oracle Cloud
1. Sign up for Free Tier: https://www.oracle.com/cloud/free/
2. Create 2 VMs (Ubuntu 24.04, 1 GB RAM):
   - VM1: Backend API (name: api-vm, ports 22, 80, 443)
   - VM2: SQLite DB (name: db-vm, port 22)
   ```bash
   oci compute instance launch --compartment-id <id> --shape VM.Standard.E2.1.Micro --image-id <ubuntu-id> --subnet-id <subnet-id> --ssh-authorized-keys-file ~/.ssh/id_rsa.pub
   ```
   Note public IPs for .env

## Cloudflare
1. Add domain (e.g., securesystem.email)
2. Create A record: api.securesystem.email -> VM1 IP
3. Enable TLS 1.3, strict mode
4. Create R2 bucket secure-email-blobs, generate API keys, add to .env

## Geolocation
1. Uses HTML5 Geolocation API (browser-based lat/long, Micro-Iterations 13, 18)
2. Reverse geocoding via OpenStreetMap Nominatim: https://nominatim.openstreetmap.org/reverse
3. Rate limit: 1 request/second, cache results (GDPR-compliant)
4. Fallback: ipapi.co for IP-based location (free, 1,000 requests/day, Micro-Iteration 12)

## Run Setup
1. Copy .env.example to .env, fill keys
2. Run setup.sh on both VMs:
   ```bash
   chmod +x setup.sh
   ./setup.sh
   ```
3. Test connectivity:
   ```bash
   ./test_connect.sh
   ```

## GitHub
1. Create repo (e.g., secure-email-mvp)
2. Push initial files:
   ```bash
   git add .
   git commit -m "Initial infrastructure setup"
   git push origin main
   ``` 