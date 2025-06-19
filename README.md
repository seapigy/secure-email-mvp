# Secure Email System MVP

A web-based secure email system with end-to-end encryption, built with Go and React.

<!-- Webhook test - triggering deployment -->

## Features

- End-to-end encryption (AES-256-GCM)
- Secure link-based delivery
- Password-based access verification
- Modern, responsive UI with dark/light modes
- TOTP authentication with QR code setup
- Unified login/sign-up interface
- Glassmorphic onboarding modal
- Folder-based organization

## Tech Stack

- Backend: Go 1.23
- Frontend: React 18
- Database: SQLite
- Storage: Cloudflare R2
- Hosting: Oracle Cloud Free Tier, Netlify

## Prerequisites

- Node.js v20
- Go v1.23
- Git

## Development Setup

### Frontend Setup
1. Navigate to the frontend directory:
   ```bash
   cd src
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Create environment file:
   ```bash
   echo "REACT_APP_API_HOST=http://localhost:8080" > .env.local
   ```

4. Start development server:
   ```bash
   npm start
   ```

5. Run tests:
   ```bash
   npm test
   ```

### Backend Setup
1. Install Go dependencies:
   ```bash
   go mod tidy
   ```

2. Set up the database:
   ```bash
   # Create database directory
   sudo mkdir -p /var/db
   sudo chown $USER:$USER /var/db
   
   # Apply schema
   sqlite3 /var/db/secure-email.db < schema/users.sql
   ```

3. Generate JWT secret:
   ```bash
   # Generate a secure 32-byte secret
   openssl rand -base64 32
   # Add this to your .env file as JWT_SECRET
   ```

4. Run the development server:
   ```bash
   go run cmd/api/main.go
   ```

5. Run tests:
   ```bash
   go test ./pkg/auth
   ```

## API Setup

The backend API provides authentication endpoints:

### Login API
- **Endpoint**: `POST /api/auth/login`
- **Authentication**: Password (Argon2) + TOTP (6-digit)
- **Response**: JWT token for subsequent requests
- **Security**: Rate limiting, secure headers, TLS 1.3

### Sign-Up API
- **Endpoint**: `POST /api/auth/signup`
- **Input**: Email, password, confirm_password
- **Response**: TOTP QR code and temp_id
- **Validation**: Email format, password match, user count <100

### Verify TOTP API
- **Endpoint**: `POST /api/auth/verify-totp`
- **Input**: temp_id, totp_code
- **Response**: JWT token
- **Process**: Creates user after TOTP validation

### Testing the API
```bash
# Run the test suite
chmod +x tests/login_test.sh
./tests/login_test.sh

# Manual testing with curl
curl -X POST https://api.securesystem.email/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@securesystem.email",
    "password": "securepassword123",
    "totp_code": "123456"
  }'
```

### API Documentation
See `docs/api/` for detailed API documentation.

## Deployment

### Backend (Oracle Cloud)
1. Follow `docs/infra.md` for Oracle Cloud setup
2. Run `setup.sh` on VM1
3. Deploy API:
   ```bash
   # On VM1
   cd /opt/secure-email-mvp
   go build -o api cmd/api/main.go
   sudo systemctl enable secure-email-api
   sudo systemctl start secure-email-api
   ```

### Frontend (Netlify)
1. Install the Netlify CLI:
   ```bash
   npm install -g netlify-cli
   ```

2. Log in to your Netlify account:
   ```bash
   netlify login
   ```

3. Initialize the site:
   ```bash
   netlify init
   ```

4. Deploy to production:
   ```bash
   netlify deploy --prod
   ```

## Project Structure

```
.
├── cmd/
│   └── api/          # Backend entry point
├── pkg/
│   └── auth/         # Authentication package
├── schema/
│   ├── users.sql     # Database schema
│   └── temp_totp.sql # Temporary TOTP storage
├── src/              # Frontend source
│   ├── components/   # React components
│   ├── styles/       # CSS and Tailwind
│   ├── lib/          # Frontend utilities
│   └── tests/        # Frontend tests
├── docs/             # Documentation
├── tests/            # Test files
└── env.example       # Environment variables template
```

## Security Features

- **TLS 1.3**: Enforced by Cloudflare
- **Rate Limiting**: 10 requests/minute per IP
- **Secure Headers**: HSTS, CSP, X-Frame-Options
- **Password Hashing**: Argon2 with email as salt
- **TOTP Authentication**: 6-digit codes, 30-second window
- **JWT Tokens**: HS256 signed, 24-hour expiration
- **Input Validation**: Email format, password length, TOTP format
- **CORS Protection**: Restricted origins

## Design System

- **Colors**: Primary (#1E40AF), Accent (#F472B6), Success (#34D399), Error (#EF4444)
- **Typography**: Inter font family
- **Animations**: GSAP for smooth transitions
- **Accessibility**: WCAG 2.2 AA compliant
- **Responsive**: Mobile-first design (320px–1440px)

## Development Status

- ✅ **Micro-Iteration 1**: Infrastructure Setup (Oracle Cloud, Cloudflare, SQLite)
- ✅ **Micro-Iteration 2**: Login API (Password + TOTP authentication)
- ✅ **Micro-Iteration 3.1**: Login and Sign-Up UI Redesign (Next)

## Next Steps

**Micro-Iteration 4**: Email Send API with AES-256-GCM encryption, compression, and R2 storage.

## License

MIT License 