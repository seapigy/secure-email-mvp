# Secure Email System MVP

A web-based secure email system with end-to-end encryption, built with Go and React.

<!-- Webhook test - triggering deployment -->

## Features

- End-to-end encryption (AES-256-GCM)
- Secure link-based delivery
- Password-based access verification
- Modern, responsive UI with dark/light modes
- TOTP authentication
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

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd secure-email-mvp
   ```

2. Install dependencies:
   ```bash
   # Backend
   go mod tidy

   # Frontend
   cd src
   npm install
   ```

3. Set up environment variables:
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

4. Set up the database:
   ```bash
   # Create database directory
   sudo mkdir -p /var/db
   sudo chown $USER:$USER /var/db
   
   # Apply schema
   sqlite3 /var/db/secure-email.db < schema/users.sql
   ```

5. Generate JWT secret:
   ```bash
   # Generate a secure 32-byte secret
   openssl rand -base64 32
   # Add this to your .env file as JWT_SECRET
   ```

6. Run the development servers:
   ```bash
   # Backend API
   go run cmd/api/main.go

   # Frontend (in another terminal)
   cd src
   npm run dev
   ```

## API Setup

The backend API provides authentication endpoints:

### Login API
- **Endpoint**: `POST /api/auth/login`
- **Authentication**: Password (Argon2) + TOTP (6-digit)
- **Response**: JWT token for subsequent requests
- **Security**: Rate limiting, secure headers, TLS 1.3

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
See `docs/api/login.md` for detailed API documentation.

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
│   └── users.sql     # Database schema
├── src/              # Frontend source
│   ├── components/   # React components
│   ├── styles/       # CSS and Tailwind
│   └── lib/          # Frontend utilities
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

## Development Status

- ✅ **Micro-Iteration 1**: Infrastructure Setup (Oracle Cloud, Cloudflare, SQLite)
- ✅ **Micro-Iteration 2**: Login API (Password + TOTP authentication)
- 🔄 **Micro-Iteration 3**: Login UI and Design System (Next)

## License

MIT License 