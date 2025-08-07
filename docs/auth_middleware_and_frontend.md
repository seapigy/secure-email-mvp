# Auth Middleware and Frontend Token Hooks

## Overview

This document describes the finalized authentication middleware and frontend token management system for the Secure Email MVP. The system provides secure JWT-based authentication with automatic token refresh and comprehensive frontend integration.

## Backend Components

### EnhancedJWTMiddleware

The `EnhancedJWTMiddleware` validates JWT access tokens and injects user information into the request context.

#### Features:
- **Token Validation**: Validates JWT access tokens from Authorization header
- **Context Injection**: Injects `user_id` and `email` into request context
- **Error Handling**: Returns 401 for missing or invalid tokens
- **Security**: Uses HMAC-SHA256 signature verification

#### Usage:
```go
// Apply middleware to protected routes
r.Handle("/protected", jwtMiddleware(http.HandlerFunc(protectedHandler))).Methods("GET")

// Access user info in handlers
userID, email, ok := GetUserFromContext(r)
if !ok {
    http.Error(w, "User not found in context", http.StatusInternalServerError)
    return
}
```

#### Helper Functions:
```go
// Get user information from context
func GetUserFromContext(r *http.Request) (string, string, bool)

// Get email from context
func GetEmailFromContext(r *http.Request) (string, bool)

// Get user ID from context
func GetUserIDFromContext(r *http.Request) (string, bool)
```

### GET /api/auth/me Endpoint

Returns current user information from the JWT token.

#### Request:
```http
GET /api/auth/me
Authorization: Bearer <access_token>
```

#### Response:
```json
{
  "user_id": "user-uuid",
  "email": "user@securesystem.email"
}
```

#### Error Responses:
- `401 Unauthorized`: Missing or invalid token
- `500 Internal Server Error`: User context not found

## Frontend Components

### Token Storage (`TokenStorage`)

Manages JWT tokens in sessionStorage with automatic expiry tracking.

#### Methods:
```javascript
// Store tokens
TokenStorage.setTokens(accessToken, refreshToken)

// Retrieve tokens
const accessToken = TokenStorage.getAccessToken()
const refreshToken = TokenStorage.getRefreshToken()

// Check expiry
const isExpired = TokenStorage.isTokenExpired()

// Clear tokens
TokenStorage.clearTokens()
```

### Authentication API (`AuthAPI`)

Provides authentication-related API calls with automatic error handling.

#### Methods:
```javascript
// Login user
const response = await AuthAPI.login(email, password, totpCode)
// Returns: { access_token, refresh_token, user_id, email }

// Refresh access token
const response = await AuthAPI.refreshToken()
// Returns: { access_token, token_type, expires_in }

// Logout user
await AuthAPI.logout()

// Get current user
const user = await AuthAPI.getCurrentUser()
// Returns: { user_id, email }
```

### Authenticated Fetch (`authFetch`)

Wrapper for fetch that automatically handles token refresh and authentication.

#### Features:
- **Automatic Refresh**: Refreshes tokens before expiry
- **Retry Logic**: Retries failed requests with new tokens
- **Error Handling**: Redirects to login on authentication failure
- **Header Management**: Automatically adds Authorization header

#### Usage:
```javascript
// Make authenticated API calls
const response = await authFetch('/api/email/list')
const emails = await response.json()

// With custom options
const response = await authFetch('/api/email/send', {
  method: 'POST',
  body: JSON.stringify(emailData)
})
```

### Authentication Utilities (`AuthUtils`)

Helper functions for authentication state management.

#### Methods:
```javascript
// Check if user is authenticated
const isAuth = AuthUtils.isAuthenticated()

// Get user info from token (no API call)
const user = AuthUtils.getUserFromToken()

// Require authentication (redirect if not)
AuthUtils.requireAuth()
```

## Integration Examples

### Login Flow
```javascript
import { AuthAPI } from './lib/auth.js'

async function handleLogin(email, password, totpCode) {
  try {
    const response = await AuthAPI.login(email, password, totpCode)
    console.log('Login successful:', response.user_id)
    // Redirect to inbox or dashboard
    window.location.href = '/inbox'
  } catch (error) {
    console.error('Login failed:', error.message)
    // Show error to user
  }
}
```

### Protected Component
```javascript
import { AuthUtils, authFetch } from './lib/auth.js'

// Check authentication on component mount
if (!AuthUtils.isAuthenticated()) {
  window.location.href = '/login'
  return
}

// Make authenticated API calls
async function loadEmails() {
  try {
    const response = await authFetch('/api/email/list')
    const emails = await response.json()
    // Update UI with emails
  } catch (error) {
    console.error('Failed to load emails:', error)
  }
}
```

### Logout Flow
```javascript
import { AuthAPI } from './lib/auth.js'

async function handleLogout() {
  try {
    await AuthAPI.logout()
    // Redirect to login
    window.location.href = '/login'
  } catch (error) {
    console.error('Logout error:', error)
    // Still redirect to login
    window.location.href = '/login'
  }
}
```

## Security Features

### Token Security
- **Session Storage**: Tokens stored in sessionStorage (cleared on browser close)
- **Automatic Refresh**: Tokens refreshed 5 minutes before expiry
- **Revocation**: Refresh tokens revoked on logout
- **Secure Headers**: Authorization header with Bearer token

### Error Handling
- **Graceful Degradation**: Redirects to login on authentication failure
- **Retry Logic**: Automatically retries failed requests with fresh tokens
- **Error Logging**: Comprehensive error logging for debugging

### Best Practices
- **Token Validation**: All tokens validated on server side
- **Context Injection**: User info injected into request context
- **Secure Storage**: Tokens never stored in localStorage
- **Automatic Cleanup**: Expired tokens automatically cleaned up

## Testing

### Backend Tests
```bash
# Run middleware tests
go test ./cmd/api -v -run TestMeHandler

# Run session management tests
go test ./cmd/api -v -run TestSessionManagement
```

### Frontend Tests
```javascript
// Test token storage
TokenStorage.setTokens('test-token', 'test-refresh')
expect(TokenStorage.getAccessToken()).toBe('test-token')

// Test authentication check
expect(AuthUtils.isAuthenticated()).toBe(true)
```

## Deployment Notes

### Environment Variables
```bash
# Required for JWT signing
JWT_SECRET=your_32_byte_jwt_secret_here

# Optional: Separate secrets for access and refresh tokens
JWT_ACCESS_SECRET=your_access_token_secret_here
JWT_REFRESH_SECRET=your_refresh_token_secret_here
```

### Frontend Configuration
- Ensure `/api` base URL is correct for your deployment
- Configure token refresh threshold as needed
- Set up proper error handling for production

### Security Considerations
- Use HTTPS in production
- Set secure JWT secrets
- Monitor authentication failures
- Implement rate limiting on auth endpoints
- Regular token cleanup for expired refresh tokens

## API Reference

### Authentication Endpoints

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/auth/login` | POST | Login with email/password/TOTP | No |
| `/api/auth/refresh` | POST | Refresh access token | No |
| `/api/auth/logout` | POST | Logout and revoke token | No |
| `/api/auth/me` | GET | Get current user info | Yes |

### Response Formats

#### Login Response
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

#### Me Response
```json
{
  "user_id": "user-uuid",
  "email": "user@securesystem.email"
}
```

#### Error Response
```json
{
  "error": "Error message description"
}
```

## Future Enhancements

1. **Token Rotation**: Implement automatic refresh token rotation
2. **Device Management**: Track and manage multiple device sessions
3. **Audit Logging**: Comprehensive authentication audit trail
4. **Multi-Factor**: Additional authentication factors for sensitive operations
5. **Session Timeout**: Configurable session timeout settings 