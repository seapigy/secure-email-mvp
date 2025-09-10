# Signup Micro-Iteration Documentation

## Overview

This micro-iteration implements user signup functionality with email verification and recovery token generation. The implementation includes a MySQL database schema, secure password hashing with Argon2id, token generation and hashing, and email verification system.

## Changes Made

### Database Schema
- **File**: `migrations/migrate_create_users.sql`
- **Description**: Creates the `users` table with all required fields for user management, email verification, and recovery tokens.

### Core Handlers
- **File**: `handlers/signup.go`
- **Description**: Implements the signup endpoint (`POST /api/signup`) with validation, password hashing, token generation, and email sending.

### Helper Libraries
- **File**: `internal/auth/argon2.go`
- **Description**: Argon2id password hashing and verification with configurable parameters.
- **File**: `internal/tokens/pqc.go`
- **Description**: Cryptographically secure token generation and SHA-512 hashing.
- **File**: `internal/email/send.go`
- **Description**: Production-ready email sending with SMTP support and retry logic.

### Email Templates
- **Files**: `email_templates/verification.html` and `email_templates/verification.txt`
- **Description**: Professional email templates for user verification with branding and security information.

### Configuration
- **File**: `config/env.example`
- **Description**: Environment variable documentation for database, email, security, and server configuration.

### Testing
- **File**: `tests/signup_test.go`
- **Description**: Comprehensive test suite covering database schema, signup functionality, token generation, password hashing, and Oracle reference validation.

### Application Entry Point
- **File**: `main.go`
- **Description**: Basic HTTP server setup with signup route registration.
- **File**: `go.mod`
- **Description**: Go module definition with required dependencies.

## How to Test Locally

### Prerequisites
1. MySQL database server running
2. Go 1.21+ installed
3. SMTP server configured (for email testing)

### Setup Steps

1. **Install Dependencies**
   ```bash
   go mod tidy
   ```

2. **Configure Environment**
   ```bash
   cp config/env.example .env
   # Edit .env with your database and email settings
   ```

3. **Run Database Migration**
   ```bash
   mysql -u username -p database_name < migrations/migrate_create_users.sql
   ```

4. **Start the Server**
   ```bash
   go run main.go
   ```

5. **Test Signup Endpoint**
   ```bash
   curl -X POST http://localhost:8080/api/signup \
     -H "Content-Type: application/json" \
     -d '{
       "email": "test@securesystem.email",
       "password": "password123",
       "tier": "free"
     }'
   ```

### Running Tests

1. **Set Test Database**
   ```bash
   export TEST_DB_DSN="mysql://user:pass@tcp(localhost:3306)/test_db?parseTime=true"
   ```

2. **Run All Tests**
   ```bash
   go test ./...
   ```

3. **Run Specific Test**
   ```bash
   go test ./tests -v
   ```

## API Endpoints

### POST /api/signup

**Request Body:**
```json
{
  "email": "user@securesystem.email",
  "password": "securepassword",
  "tier": "free|premium|enterprise",
  "custom_domain": "example.com" // optional, for premium/enterprise
}
```

**Response (201 Created):**
```json
{
  "status": "ok",
  "message": "Signup successful. A verification email has been sent. Save your recovery token securely.",
  "recovery_token": "USER-VISUAL-TOKEN-HERE",
  "recovery_token_qr_data_uri": "data:image/png;base64,...."
}
```

**Validation Rules:**
- Email must be valid format
- Password must be at least 8 characters
- Free tier accounts must use `@securesystem.email` domain
- Tier must be one of: `free`, `premium`, `enterprise`

## Security Features

### Password Security
- Argon2id hashing with configurable parameters
- Memory: 128MB (configurable)
- Iterations: 4 (configurable)
- Parallelism: 4 (configurable)
- Salt length: 16 bytes (configurable)
- Key length: 32 bytes (configurable)

### Token Security
- Cryptographically random token generation (32 bytes)
- SHA-512 hashing for storage
- Constant-time comparison for verification
- Configurable expiration times

### Email Security
- Verification tokens expire in 24 hours (configurable)
- Recovery tokens expire in 7 days (configurable)
- SMTP authentication required
- Retry logic with exponential backoff

## Database Schema

The `users` table includes:
- `id`: Primary key (BIGINT AUTO_INCREMENT)
- `email`: Unique email address (VARCHAR(255))
- `email_verified`: Boolean verification status
- `verification_token_hash`: Hashed verification token (CHAR(128))
- `verification_exp`: Token expiration timestamp
- `password_hash`: Argon2id hashed password (CHAR(128))
- `reset_token_hash`: Hashed recovery token (CHAR(128))
- `reset_token_exp`: Recovery token expiration
- `tier`: Account tier (ENUM: free, premium, enterprise)
- `custom_domain`: Custom domain for premium accounts
- `domain_verified`: Domain verification status
- `created_at`, `updated_at`, `last_login`: Timestamps

## Environment Variables

Required environment variables (see `config/env.example` for full list):
- `DB_DSN`: MySQL connection string
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`: Email configuration
- `EMAIL_FROM`: Sender email address
- `FRONTEND_URL`: Frontend URL for verification links
- `VERIFICATION_TOKEN_EXP_HOURS`: Token expiration (default: 24)
- `RECOVERY_TOKEN_EXP_DAYS`: Recovery token expiration (default: 7)

## Testing Coverage

The test suite covers:
- Database schema validation
- Signup endpoint functionality
- Input validation
- Password hashing and verification
- Token generation and hashing
- Response format validation
- Oracle reference detection (ensures clean codebase)

## Next Steps

This micro-iteration provides the foundation for:
1. Email verification endpoint
2. Password reset functionality
3. User login system
4. Domain verification for premium accounts
5. Enterprise account management

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Verify `DB_DSN` environment variable
   - Ensure MySQL server is running
   - Check database credentials

2. **Email Sending Failed**
   - Verify SMTP configuration
   - Check network connectivity
   - Review SMTP server logs

3. **Tests Failing**
   - Ensure `TEST_DB_DSN` is set
   - Run migration on test database
   - Check test database permissions

### Logs

The application logs important events:
- User creation success/failure
- Email sending status
- Database operation errors
- Token generation events

Check application logs for detailed error information.
