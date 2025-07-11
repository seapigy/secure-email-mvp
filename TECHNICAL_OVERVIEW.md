# Secure Email MVP - Technical Overview for ChatGPT

## 📁 PROJECT STRUCTURE

```
/home/opc/secure-email-mvp/
├── cmd/api/main.go                    # Go backend server entry point
├── pkg/auth/                          # Authentication package
│   ├── login.go                       # Login handler
│   ├── signup.go                      # Signup handler  
│   ├── verify_totp.go                 # TOTP verification
│   └── *_test.go                      # Test files
├── src/                               # React frontend
│   ├── App.jsx                        # Main app component
│   ├── components/
│   │   ├── AuthCard.jsx               # Login/signup form
│   │   ├── Inbox.jsx                  # Email inbox component
│   │   └── OnboardingModal.jsx        # TOTP setup modal
│   ├── lib/api.js                     # API client configuration
│   └── main.jsx                       # React entry point
├── schema.sql                         # Database schema
├── env.example                        # Environment template
├── package.json                       # Frontend dependencies
├── go.mod                             # Backend dependencies
└── dist/                              # Built frontend (Netlify deployment)
```

**Key File Sizes & Dates:**
- Database: `/var/db/secure-email.db` (73KB, created Jul 7 23:12)
- Backend binary: `/tmp/api-server` (compiled from Go source)
- Frontend build: `dist/` (Vite build output for Netlify)

## 🧠 COMPONENT OVERVIEW

### Backend Architecture (Go 1.23)
- **Framework**: Gorilla Mux for routing
- **Database**: SQLite with go-sqlite3 driver
- **Authentication**: JWT tokens + TOTP (Google Authenticator)
- **Password Hashing**: Argon2 (via golang.org/x/crypto)
- **Rate Limiting**: In-memory IP-based rate limiting (10 req/min)
- **CORS**: Configured for localhost:3000 and Netlify domain

### Frontend Architecture (React 18)
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **Routing**: React Router DOM
- **Notifications**: React Toastify
- **Icons**: Heroicons

### Data Flow
1. **User Registration**: Frontend → `/api/auth/signup` → SQLite users table
2. **TOTP Setup**: QR code generation → Google Authenticator app
3. **Login**: Frontend → `/api/auth/login` → JWT token generation
4. **Email Storage**: Encrypted content → Cloudflare R2 (planned)
5. **Access Control**: IP-based rate limiting + geolocation restrictions

## 📊 DATABASE

**Location**: `/var/db/secure-email.db` (73KB)

### Schema (SQLite)
```sql
-- Users table (0 records currently)
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    totp_secret TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Emails table (0 records currently)  
CREATE TABLE emails (
    id TEXT PRIMARY KEY,
    sender_id TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    subject TEXT,
    encrypted_content TEXT NOT NULL,
    access_password_hash TEXT,
    geolocation_circles TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(id)
);

-- Access attempts tracking
CREATE TABLE access_attempts (
    id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    attempt_count INTEGER DEFAULT 0,
    last_attempt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id)
);

-- Folder organization
CREATE TABLE folders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Email-folder mapping
CREATE TABLE email_folders (
    email_id TEXT NOT NULL,
    folder_id TEXT NOT NULL,
    PRIMARY KEY (email_id, folder_id),
    FOREIGN KEY (email_id) REFERENCES emails(id),
    FOREIGN KEY (folder_id) REFERENCES folders(id)
);
```

**Indexes**: Email lookups, sender/recipient queries, expiration tracking

## 🌐 NETWORKING & INFRASTRUCTURE

### Backend Server
- **Port**: 8080 (listening on all interfaces)
- **Process**: `/tmp/api-server` (PID 74326)
- **Status**: ✅ Running and accessible
- **Logs**: `/home/opc/api.log`

### Frontend Deployment
- **Platform**: Netlify
- **Domain**: `secure-email-mvp.netlify.app`
- **Build**: Vite production build in `dist/`
- **API Base URL**: `https://api.securesystem.email`

### DNS Configuration
- **Domain**: `securesystem.email`
- **API Subdomain**: `api.securesystem.email` → Oracle Cloud VM
- **Frontend**: `secure-email-mvp.netlify.app`
- **Nameservers**: Name.com → Cloudflare

## 🔐 SECURITY

### Authentication Flow
1. **Registration**: Username → `username@securesystem.email` + password
2. **TOTP Setup**: QR code generation for Google Authenticator
3. **Login**: Email + password + TOTP code
4. **JWT Token**: 32-byte secret (needs generation)
5. **Session**: Token stored in sessionStorage

### Password Security
- **Hashing**: Argon2 (golang.org/x/crypto/argon2)
- **Validation**: 8-128 characters required
- **Email Restriction**: Only `@securesystem.email` domains

### Rate Limiting
- **Requests**: 10 per minute per IP
- **Storage**: In-memory sync.Map
- **Reset**: Automatic after 60 seconds

### Security Headers
- Content-Security-Policy: `default-src 'self'`
- Strict-Transport-Security: `max-age=31536000`
- X-Frame-Options: `DENY`
- X-Content-Type-Options: `nosniff`
- Referrer-Policy: `strict-origin-when-cross-origin`

## ☁️ CLOUD & DEPLOYMENT

### Oracle Cloud VM
- **OS**: Oracle Linux Server 9.6
- **Architecture**: ARM64 (linux/arm64)
- **Public IP**: 129.146.68.127
- **Go Version**: 1.23.9 (Red Hat)
- **User**: opc (SSH key authentication)

### Installed Packages
- **Go**: 1.23.9 (Red Hat build)
- **SQLite**: 3.44.2
- **System**: Oracle Linux 9.6 with dnf package manager

### Environment Configuration
```bash
# Required .env file (not yet created)
CLOUDFLARE_R2_ACCESS_KEY=your_r2_access_key_here
CLOUDFLARE_R2_SECRET_KEY=your_r2_secret_key_here
CLOUDFLARE_R2_BUCKET=secure-email-blobs
CLOUDFLARE_R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
API_HOST=api.securesystem.email
API_PORT=8080
SQLITE_DB=/var/db/secure-email.db
JWT_SECRET=your_32_byte_jwt_secret_here
LOG_FILE=/var/log/api.log
RATE_LIMIT_REQUESTS=10
RATE_LIMIT_WINDOW=60
DEBUG=false
```

### Netlify Configuration
- **Build Command**: `npm run build`
- **Publish Directory**: `dist`
- **Environment Variables**: `REACT_APP_API_HOST`

## 📦 STORAGE

### Current Storage Locations
- **Project**: `/home/opc/secure-email-mvp/` (51 files, 18 directories)
- **Database**: `/var/db/secure-email.db` (73KB)
- **Logs**: `/home/opc/api.log`
- **Binary**: `/tmp/api-server`

### Planned Storage (R2)
- **Bucket**: `secure-email-blobs`
- **Purpose**: Encrypted email content storage
- **Status**: ⏳ Not yet configured
- **Access**: Via Cloudflare R2 API keys

## ⚙️ BACKEND STATUS

### Server Status
- **Process**: ✅ Running (PID 74326)
- **Port**: ✅ Listening on :8080
- **Database**: ✅ Connected and initialized
- **Logs**: Minimal output - "Starting API on :8080"

### API Endpoints
- `POST /api/auth/login` - User authentication
- `POST /api/auth/signup` - User registration  
- `POST /api/auth/verify-totp` - TOTP verification

### Dependencies (Go)
```go
github.com/dgrijalva/jwt-go v3.2.0
github.com/google/uuid v1.6.0
github.com/gorilla/mux v1.8.1
github.com/joho/godotenv v1.5.1
github.com/mattn/go-sqlite3 v1.14.22
github.com/pquerna/otp v1.4.0
github.com/rs/cors v1.11.1
golang.org/x/crypto v0.21.0
```

## 🧹 OPTIONAL CLEANUP

### Unused Files
- `go1.23.0.windows-amd64.msi` - Windows Go installer (not needed on Linux)
- `vm_connection/` - SSH key files (could be archived)
- `keys/` - Empty directory
- `tests/` - Shell test scripts (could be moved to scripts/)

### Development Files
- `*.test.jsx` - Test files (development only)
- `*.test.go` - Go test files (development only)
- `env.example` - Template (should be copied to .env)

## 📌 REMINDERS FOR CHATGPT CONTEXT

### Technology Stack
- **Frontend**: React 18 + Vite + Tailwind CSS
- **Backend**: Go 1.23 + Gorilla Mux + SQLite
- **Authentication**: JWT + TOTP (Google Authenticator)
- **Deployment**: Netlify (frontend) + Oracle Cloud (backend)

### Domain Restrictions
- **Registration**: Only `@securesystem.email` addresses
- **Username**: 3-50 chars, alphanumeric + ._%+-
- **Password**: 8-128 characters

### Infrastructure
- **DNS**: Cloudflare (proxy) → Name.com (nameservers)
- **Backend**: Oracle Cloud VM (ARM64, Oracle Linux 9.6)
- **Frontend**: Netlify (CDN + hosting)
- **Storage**: Cloudflare R2 (planned for encrypted emails)

### Current Focus
- ✅ Backend API server running
- ✅ Database initialized
- ✅ Frontend deployed to Netlify
- ⏳ Cloudflare R2 configuration pending
- ⏳ Email encryption/storage implementation
- ⏳ Geolocation access controls
- ⏳ Production environment variables

### Security Notes
- JWT secret needs generation (32 bytes)
- R2 API keys need configuration
- Rate limiting is basic (in-memory)
- No SSL termination configured yet
- Database is local SQLite (not production-ready for scale) 