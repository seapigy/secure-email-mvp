# Secure Email MVP - Technical Overview

## 📁 PROJECT STRUCTURE

```
/home/opc/secure-email-mvp/
├── cmd/api/                           # Backend API server
│   ├── main.go                        # Server entry point with JWT middleware
│   ├── login_handler.go               # Login authentication handler
│   ├── signup_handler.go              # User registration handler
│   ├── verify_totp.go                 # TOTP verification handler
│   ├── fallback_handler.go            # Fallback email confirmation
│   ├── resend_fallback_handler.go     # Resend fallback confirmation
│   ├── logout_handler.go              # User logout handler
│   ├── me_handler.go                  # Get current user info
│   ├── refresh_handler.go             # JWT token refresh
│   ├── send_email_handler.go          # Email sending with encryption
│   ├── get_email_handler.go           # Email retrieval with decryption
│   ├── list_email_handler.go          # List user's emails
│   ├── view_email_handler.go          # View individual email
│   ├── delete_email_handler.go        # Delete email with cleanup
│   ├── health_handler.go              # Health check endpoint
│   ├── jwt_middleware.go              # JWT authentication middleware
│   ├── rate_limit.go                  # Rate limiting middleware
│   └── *_test.go                      # Comprehensive test suites
├── pkg/auth/                          # Authentication & encryption
│   ├── jwt.go                         # JWT token generation/validation
│   ├── encryption.go                  # AES-256-GCM encryption
│   ├── login.go                       # Login handler logic
│   ├── signup.go                      # Signup handler logic
│   ├── verify_totp.go                 # TOTP verification logic
│   ├── session.go                     # Session management
│   ├── fallback.go                    # Fallback email logic
│   └── *_test.go                      # Test files
├── pkg/storage/                       # Cloudflare R2 storage
│   └── r2.go                          # R2 client and operations
├── docs/                              # API documentation
│   ├── jwt_authentication.md          # JWT implementation guide
│   ├── encryption_implementation.md   # Encryption details
│   ├── r2_storage_implementation.md   # R2 storage guide
│   ├── complete_email_flow.md         # End-to-end email flow
│   ├── list_email_endpoint.md         # List endpoint docs
│   ├── view_email_endpoint.md         # View endpoint docs
│   ├── delete_email_endpoint.md       # Delete endpoint docs
│   ├── session_management.md          # Session management guide
│   ├── auth_middleware_and_frontend.md # Auth middleware docs
│   └── infra.md                       # Infrastructure setup
├── schema/                            # Database migrations
│   └── migrate_add_encryption_fields.sql
├── examples/                          # Usage examples
│   ├── encryption_example.go          # Encryption demo
│   └── email_upload_example.go        # R2 upload demo
├── src/                               # React frontend (TypeScript)
│   ├── App.tsx                        # Main app component
│   ├── components/
│   │   ├── auth/                      # Authentication components
│   │   │   ├── LoginForm.tsx          # Login form component
│   │   │   └── SignupForm.tsx         # Signup form component
│   │   ├── email/                     # Email management components
│   │   │   ├── EmailSendForm.tsx      # Secure email composition
│   │   │   └── EmailView.tsx          # Email viewing component
│   │   ├── layout/                    # Layout components
│   │   │   ├── Layout.tsx             # Main layout component
│   │   │   ├── Sidebar.tsx            # Navigation sidebar
│   │   │   └── Header.tsx             # App header
│   │   ├── pages/                     # Page components
│   │   │   ├── Dashboard.tsx          # Dashboard page
│   │   │   ├── Send.tsx               # Send email page
│   │   │   └── View.tsx               # View email page
│   │   ├── secure/                    # Secure email UI components
│   │   │   ├── SecureEmailPage.tsx    # Main secure email interface
│   │   │   ├── EmailInbox.tsx         # Email inbox with filtering
│   │   │   ├── EmailDetail.tsx        # Email detail view
│   │   │   ├── SecuritySettings.tsx   # Security settings panel
│   │   │   ├── ComposeModal.tsx       # Email composition modal
│   │   │   └── UnlockModal.tsx        # Password unlock modal
│   │   └── ui/                        # Reusable UI components
│   │       ├── Button.tsx             # Button component
│   │       ├── Input.tsx              # Input component
│   │       ├── HealthStatusBanner.tsx # Health check banner
│   │       └── Modal.tsx              # Modal component
│   ├── hooks/                         # Custom React hooks
│   │   ├── useAuth.ts                 # Authentication hook
│   │   ├── useTheme.ts                # Theme management hook
│   │   ├── useHealthCheck.ts          # Health check hook
│   │   └── useEmail.ts                # Email operations hook
│   ├── stores/                        # State management (Zustand)
│   │   ├── authStore.ts               # Authentication state
│   │   ├── emailStore.ts              # Email state management
│   │   ├── uiStore.ts                 # UI state management
│   │   └── sessionStore.ts            # Session state for unlocked emails
│   ├── lib/                           # Utility functions
│   │   ├── api.ts                     # API client configuration
│   │   ├── utils.ts                   # General utilities
│   │   └── validation.ts              # Input validation
│   ├── types/                         # TypeScript type definitions
│   │   ├── auth.ts                    # Authentication types
│   │   ├── email.ts                   # Email types
│   │   ├── secureEmail.ts             # Secure email types
│   │   └── api.ts                     # API response types
│   ├── data/                          # Mock data
│   │   └── mockEmails.json            # Mock email data
│   └── main.tsx                       # React entry point
├── schema.sql                         # Database schema
├── env.example                        # Environment template
├── package.json                       # Frontend dependencies
├── go.mod                             # Backend dependencies
└── dist/                              # Built frontend (Netlify deployment)
```

**Key File Sizes & Dates:**
- Database: `/var/db/secure-email.db` (20KB, created recently)
- Backend binary: `/tmp/api-server` (compiled from Go source)
- Frontend build: `dist/` (Vite build output for Netlify)

## 🧠 COMPONENT OVERVIEW

### Backend Architecture (Go 1.23)
- **Framework**: Gorilla Mux for routing
- **Database**: SQLite with modernc.org driver
- **Authentication**: JWT tokens + TOTP (Google Authenticator)
- **Password Hashing**: Argon2 (via golang.org/x/crypto)
- **Encryption**: AES-256-GCM for email content
- **Storage**: Cloudflare R2 for encrypted blobs
- **Rate Limiting**: In-memory IP-based rate limiting (10 req/min)
- **CORS**: Configured for localhost:3000 and Netlify domain
- **Session Management**: JWT-based with refresh tokens
- **Health Monitoring**: `/health` endpoint for connectivity checks

### Frontend Architecture (React 18 + TypeScript)
- **Build Tool**: Vite
- **Styling**: Tailwind CSS with custom design system
- **HTTP Client**: Axios with interceptors
- **Routing**: React Router DOM
- **State Management**: Zustand
- **Notifications**: React Toastify
- **Icons**: Lucide React (recently migrated from Heroicons)
- **Animations**: GSAP and Framer Motion
- **Health Monitoring**: Real-time backend connectivity checks
- **Session Management**: Zustand store for tracking unlocked emails
- **Modal System**: Custom modal components for compose and unlock

### Data Flow
1. **User Registration**: Frontend → `/api/auth/signup` → SQLite users table
2. **TOTP Setup**: QR code generation → Google Authenticator app
3. **Login**: Frontend → `/api/auth/login` → JWT token generation
4. **Health Check**: Frontend → `/health` → Backend status monitoring
5. **Email Sending**: Content → gzip compression → AES-256-GCM encryption → R2 storage
6. **Email Retrieval**: R2 download → decryption → decompression → plaintext
7. **Access Control**: JWT validation + user authorization + IP-based rate limiting
8. **Session Management**: JWT refresh tokens for extended sessions
9. **Secure Email UI**: Mock data loading → Privacy-first interface → Future API integration
10. **Per-Email Password**: Individual password protection with session tracking
11. **Self-Destruct Feature**: Failed attempt tracking and auto-deletion
12. **Compose Interface**: Modern email composition with comprehensive security options

## 📊 DATABASE

**Location**: `/var/db/secure-email.db` (20KB)

### Schema (SQLite)
```sql
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    totp_secret TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Emails table
CREATE TABLE IF NOT EXISTS emails (
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
CREATE TABLE IF NOT EXISTS access_attempts (
    id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    attempt_count INTEGER DEFAULT 0,
    last_attempt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id)
);

-- Folder organization
CREATE TABLE IF NOT EXISTS folders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Email-folder mapping
CREATE TABLE IF NOT EXISTS email_folders (
    email_id TEXT NOT NULL,
    folder_id TEXT NOT NULL,
    PRIMARY KEY (email_id, folder_id),
    FOREIGN KEY (email_id) REFERENCES emails(id),
    FOREIGN KEY (folder_id) REFERENCES folders(id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_emails_sender ON emails(sender_id);
CREATE INDEX IF NOT EXISTS idx_emails_recipient ON emails(recipient_email);
CREATE INDEX IF NOT EXISTS idx_emails_expires ON emails(expires_at);
CREATE INDEX IF NOT EXISTS idx_access_attempts_email ON access_attempts(email_id);
CREATE INDEX IF NOT EXISTS idx_folders_user ON folders(user_id);
```

**Indexes**: Email lookups, sender/recipient queries, expiration tracking

## 🌐 NETWORKING & INFRASTRUCTURE

### Backend Server
- **Port**: 8080 (listening on all interfaces)
- **Process**: `/tmp/api-server` (PID varies)
- **Status**: ✅ Running and accessible
- **Logs**: `/home/opc/api.log`
- **Health Endpoint**: `GET /health` for connectivity monitoring

### Frontend Deployment
- **Platform**: Netlify
- **Domain**: `secure-email-mvp.netlify.app`
- **Build**: Vite production build in `dist/`
- **API Base URL**: `https://api.securesystem.email`
- **Health Monitoring**: Real-time backend status display

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
4. **JWT Token**: 32-byte secret with user context injection
5. **Session**: Token stored in sessionStorage with refresh capability
6. **Fallback Email**: Account recovery with confirmation flow
7. **Health Monitoring**: Continuous backend connectivity validation

### Email Encryption Flow
1. **Content Preparation**: Email subject + body
2. **Compression**: gzip compression to reduce size
3. **Encryption**: AES-256-GCM with random key and nonce
4. **Storage**: Encrypted blob uploaded to Cloudflare R2
5. **Metadata**: Encryption parameters stored in SQLite
6. **Retrieval**: Download → decrypt → decompress → plaintext

### Authorization System
- **JWT Validation**: Token validation on all protected endpoints
- **User Context**: user_id injected into request context
- **Ownership Verification**: Users can only access their own emails
- **Access Logging**: All access attempts logged for audit
- **Session Management**: Refresh tokens for extended sessions

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

### Advanced Security Features
- **Per-Email Password Protection**: Individual password protection for emails
- **Self-Destruct After Failed Attempts**: Auto-delete messages after failed access
- **Session Management**: Track unlocked emails and failed attempts
- **Geolocation Restrictions**: Restrict access by country
- **Time-Based Access**: Set unlock times for messages
- **Auto-Destruct Features**: Messages that self-destruct after viewing
- **Read-Once Mode**: Messages that can only be viewed once
- **Remote Revoke**: Ability to revoke access to sent messages
- **Decoy Messages**: Fake messages to mislead attackers
- **Metadata Stripping**: Remove identifying information
- **Tamper Alerts**: Detect unauthorized access attempts

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
# Required .env file
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
- **Environment Variables**: `VITE_API_HOST`

## 📦 STORAGE

### Current Storage Locations
- **Project**: `/home/opc/secure-email-mvp/` (51 files, 18 directories)
- **Database**: `/var/db/secure-email.db` (20KB)
- **Logs**: `/home/opc/api.log`
- **Binary**: `/tmp/api-server`

### Cloudflare R2 Storage
- **Bucket**: `secure-email-blobs`
- **Purpose**: Encrypted email content storage
- **Status**: ✅ Implemented and tested
- **Access**: Via Cloudflare R2 API keys
- **Operations**: Upload, download, delete encrypted blobs
- **Path Structure**: `emails/{blob_id}` for organized storage

## ⚙️ BACKEND STATUS

### Server Status
- **Process**: ✅ Running
- **Port**: ✅ Listening on :8080
- **Database**: ✅ Connected and initialized
- **Logs**: Minimal output - "Starting API on :8080"
- **Health Endpoint**: ✅ `/health` responding

### API Endpoints
- `POST /api/auth/login` - User authentication
- `POST /api/auth/signup` - User registration  
- `POST /api/auth/verify-totp` - TOTP verification
- `POST /api/auth/logout` - User logout
- `GET /api/auth/me` - Get current user info
- `POST /api/auth/refresh` - Refresh JWT token
- `POST /api/auth/fallback` - Send fallback confirmation
- `GET /api/auth/confirm-fallback` - Confirm fallback email
- `POST /api/auth/resend-fallback` - Resend fallback confirmation
- `POST /api/email/send` - Send encrypted email
- `POST /api/email/get` - Retrieve encrypted email
- `GET /api/email/list` - List user's emails
- `GET /api/email/view/{id}` - View individual email
- `DELETE /api/email/{id}` - Delete email with cleanup
- `GET /health` - Backend health check

### Dependencies (Go)
```go
github.com/aws/aws-sdk-go v1.55.7
github.com/dgrijalva/jwt-go v3.2.0+incompatible
github.com/google/uuid v1.6.0
github.com/gorilla/mux v1.8.1
github.com/joho/godotenv v1.5.1
github.com/pquerna/otp v1.4.0
golang.org/x/crypto v0.21.0
modernc.org/sqlite v1.28.0
```

## 🎨 FRONTEND STATUS

### React Application
- **Framework**: React 18 with TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS with custom design system
- **State Management**: Zustand
- **Routing**: React Router DOM
- **HTTP Client**: Axios with interceptors
- **UI Components**: Custom components with Headless UI
- **Icons**: Lucide React (recently migrated from Heroicons)
- **Animations**: GSAP and Framer Motion

### Key Features
- **Theme Support**: Dark/light mode with system preference detection
- **Responsive Design**: Mobile-first approach
- **Glassmorphic UI**: Modern design with glass effects
- **Toast Notifications**: User feedback for all actions
- **Loading States**: Comprehensive loading indicators
- **Error Handling**: Graceful error handling and display
- **Health Monitoring**: Real-time backend connectivity status
- **Secure Email UI**: Privacy-first design inspired by ProtonMail and Skiff
- **Split-View Layout**: Desktop layout with inbox and detail panels
- **Mobile Responsive**: Single panel layout for mobile devices
- **Compose Modal**: Modern email composition interface
- **Unlock Modal**: Password verification for protected emails

### Component Structure
- **Authentication**: LoginForm, SignupForm with TOTP setup
- **Email Management**: EmailSendForm, EmailView, Dashboard
- **Layout**: Layout, Sidebar, Header components
- **Pages**: Dashboard, Send, View page components
- **Secure Email**: SecureEmailPage, EmailInbox, EmailDetail, SecuritySettings, ComposeModal, UnlockModal
- **UI**: Button, Input, Modal, HealthStatusBanner, and other reusable components
- **Hooks**: useAuth, useTheme, useHealthCheck, useEmail for state management

### Recent Updates
- **Health Check System**: Real-time backend connectivity monitoring
- **Secure Email UI**: Privacy-first design inspired by ProtonMail and Skiff
- **Icon Migration**: Migrated from Heroicons to Lucide React
- **Mock Data Integration**: Comprehensive mock email data for development
- **Advanced Security Features**: Password protection, geolocation restrictions, auto-destruct
- **Split-View Layout**: Desktop layout with inbox and detail panels
- **Mobile Responsive**: Single panel layout for mobile devices
- **Per-Email Password Protection**: Individual password protection for emails
- **Self-Destruct After Failed Attempts**: Auto-delete messages after failed access
- **Compose Modal**: Modern email composition with comprehensive security options
- **Unlock Modal**: Password verification for protected emails
- **Session Management**: Track unlocked emails and failed attempts

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

## 📌 REMINDERS FOR CONTEXT

### Technology Stack
- **Frontend**: React 18 + TypeScript + Vite + Tailwind CSS + Lucide React
- **Backend**: Go 1.23 + Gorilla Mux + SQLite
- **Authentication**: JWT + TOTP (Google Authenticator)
- **Deployment**: Netlify (frontend) + Oracle Cloud (backend)
- **Storage**: Cloudflare R2 (encrypted email content)

### Domain Restrictions
- **Registration**: Only `@securesystem.email` addresses
- **Username**: 3-50 chars, alphanumeric + ._%+-
- **Password**: 8-128 characters

### Infrastructure
- **DNS**: Cloudflare (proxy) → Name.com (nameservers)
- **Backend**: Oracle Cloud VM (ARM64, Oracle Linux 9.6)
- **Frontend**: Netlify (CDN + hosting)
- **Storage**: Cloudflare R2 (encrypted emails)

### Current Focus
- ✅ Backend API server running
- ✅ Database initialized with encryption fields
- ✅ Frontend deployed to Netlify
- ✅ JWT authentication implemented
- ✅ AES-256-GCM encryption implemented
- ✅ Cloudflare R2 storage integrated
- ✅ Complete email CRUD operations
- ✅ Comprehensive test coverage
- ✅ Session management with refresh tokens
- ✅ Fallback email confirmation flow
- ✅ Health check system implemented
- ✅ Secure email UI with privacy-first design
- ✅ Mock data integration for development
- ✅ Split-view inbox layout implementation
- ✅ Per-email password protection
- ✅ Self-destruct after failed attempts feature
- ✅ Compose secure email modal
- ⏳ Geolocation access controls (future enhancement)
- ⏳ Production environment variables (needs configuration)

### Security Notes
- ✅ JWT authentication with user context injection
- ✅ AES-256-GCM encryption for email content
- ✅ User authorization (users can only access their own emails)
- ✅ Session management with refresh tokens
- ✅ Rate limiting is implemented (in-memory)
- ✅ Fallback email confirmation for account recovery
- ✅ Health monitoring for backend connectivity
- ✅ Privacy-first UI design with advanced security features
- ✅ Per-email password protection with session tracking
- ✅ Self-destruct after failed attempts with attempt counting
- ✅ Comprehensive security options in compose interface
- ⏳ SSL termination not configured yet
- ⏳ Database is local SQLite (not production-ready for scale) 