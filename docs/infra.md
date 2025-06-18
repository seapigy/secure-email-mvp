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
├── data/             # Database and GeoLite2
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