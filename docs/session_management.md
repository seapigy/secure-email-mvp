# Secure Session Management with JWT

## Overview

This document describes the implementation of secure JWT-based session management for the Secure Email MVP. The system provides short-lived access tokens and long-lived refresh tokens with proper security controls.

## Architecture

### Token Types

1. **Access Tokens (15 minutes)**
   - Short-lived JWT tokens for API access
   - Contains user_id, email, and expiration
   - Signed with HMAC-SHA256
   - No database storage required

2. **Refresh Tokens (7 days)**
   - Long-lived JWT tokens for token renewal
   - Contains user_id and token_id
   - Stored securely in database with hashed values
   - Can be revoked individually

### Security Features

- **Token Rotation**: Refresh tokens are validated against database
- **Revocation**: Tokens can be revoked to prevent reuse
- **Expiration**: Automatic expiration prevents indefinite access
- **Secure Storage**: Refresh tokens are hashed before database storage
- **Rate Limiting**: Login and refresh endpoints are rate-limited

## Database Schema

### Refresh Tokens Table

```sql
CREATE TABLE refresh_tokens (
    id TEXT PRIMARY KEY,                    -- UUID of the refresh token
    user_id TEXT NOT NULL,                 -- Associated user ID
    token_hash TEXT NOT NULL,              -- Hashed refresh token for security
    expires_at TIMESTAMP NOT NULL,         -- When the refresh token expires
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_revoked BOOLEAN DEFAULT FALSE,      -- Whether token has been revoked
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## API Endpoints

### POST /api/auth/login

Authenticates user and returns token pair.

**Request:**
```json
{
  "email": "user@securesystem.email",
  "password": "userpassword",
  "totp_code": "123456"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user_id": "user-uuid",
  "email": "user@securesystem.email"
}
```

### POST /api/auth/refresh

Refreshes access token using valid refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

### POST /api/auth/logout

Revokes refresh token to prevent reuse.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "message": "Successfully logged out"
}
```

## Environment Variables

```bash
# Required for JWT signing
JWT_SECRET=your_32_byte_jwt_secret_here

# Optional: Separate secrets for access and refresh tokens
JWT_ACCESS_SECRET=your_access_token_secret_here
JWT_REFRESH_SECRET=your_refresh_token_secret_here
```

## Implementation Details

### Session Manager

The `SessionManager` class handles all token operations:

```go
type SessionManager struct {
    accessTokenSecret  string
    refreshTokenSecret string
    accessTokenExpiry  time.Duration
    refreshTokenExpiry time.Duration
}
```

### Key Methods

- `GenerateTokenPair(userID, email, db)` - Creates access and refresh tokens
- `ValidateAccessToken(tokenString)` - Validates access token
- `ValidateRefreshToken(tokenString, db)` - Validates refresh token
- `RevokeRefreshToken(tokenString, db)` - Revokes refresh token
- `RevokeAllUserTokens(userID, db)` - Revokes all user tokens

### JWT Middleware

The `EnhancedJWTMiddleware` validates access tokens and injects user context:

```go
func EnhancedJWTMiddleware(db *sql.DB) func(http.Handler) http.Handler
```

## Security Considerations

### Token Security

1. **Access Tokens**
   - Short expiration (15 minutes) limits exposure
   - No database storage reduces attack surface
   - HMAC-SHA256 signing prevents tampering

2. **Refresh Tokens**
   - Hashed before database storage
   - Can be revoked individually
   - Database validation prevents replay attacks
   - Automatic cleanup of expired tokens

### Best Practices

1. **Token Storage**
   - Store access tokens in memory (sessionStorage)
   - Store refresh tokens securely (httpOnly cookies in production)
   - Never store tokens in localStorage

2. **Token Rotation**
   - Implement automatic refresh before expiration
   - Handle refresh failures gracefully
   - Redirect to login on refresh failure

3. **Error Handling**
   - Don't leak token validity information
   - Log security events for monitoring
   - Implement proper error responses

## Usage Examples

### Frontend Integration

```javascript
// Login
const response = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@securesystem.email',
    password: 'password',
    totp_code: '123456'
  })
});

const { access_token, refresh_token } = await response.json();

// Store tokens
sessionStorage.setItem('access_token', access_token);
sessionStorage.setItem('refresh_token', refresh_token);

// Use access token for API calls
const apiResponse = await fetch('/api/email/list', {
  headers: {
    'Authorization': `Bearer ${access_token}`
  }
});
```

### Automatic Token Refresh

```javascript
// Refresh token before expiration
async function refreshToken() {
  const refresh_token = sessionStorage.getItem('refresh_token');
  
  const response = await fetch('/api/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token })
  });
  
  if (response.ok) {
    const { access_token } = await response.json();
    sessionStorage.setItem('access_token', access_token);
  } else {
    // Redirect to login
    window.location.href = '/login';
  }
}
```

## Testing

Run the comprehensive test suite:

```bash
go test ./cmd/api -v -run TestSessionManagement
```

## Deployment Notes

1. **Environment Variables**: Set secure JWT secrets in production
2. **Database Migration**: Apply refresh_tokens schema
3. **Rate Limiting**: Configure appropriate limits for auth endpoints
4. **Monitoring**: Set up alerts for failed authentication attempts
5. **Cleanup**: Schedule periodic cleanup of expired tokens

## Future Enhancements

1. **Token Rotation**: Implement automatic refresh token rotation
2. **Device Tracking**: Track and manage multiple device sessions
3. **Geolocation**: Add location-based access controls
4. **Audit Logging**: Comprehensive session audit trail
5. **Multi-Factor**: Additional authentication factors for sensitive operations 