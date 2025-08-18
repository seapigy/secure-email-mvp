# Session Tokens & One-Time Links

## Overview

Micro-Iteration 4.15 implements per-session access tokens and one-time links for enhanced email security. This feature reduces replay risk and makes stolen links or tokens useless by requiring fresh tokens for each email access session.

## Features

### Core Functionality
- **Per-Session Tokens**: Short-lived (5-minute) tokens required for email access
- **One-Time Links**: Tokens that become invalid after first use
- **Secure Storage**: Argon2id hashed tokens with email ID as salt
- **Automatic Cleanup**: Expired tokens are automatically removed
- **Audit Logging**: Comprehensive logging of token issuance and usage

### Security Benefits
- **Replay Protection**: Stolen tokens become useless quickly
- **Session Isolation**: Each access requires a fresh token
- **One-Time Access**: Highly sensitive emails can be configured for single-use
- **Audit Trail**: Complete tracking of token usage and access patterns

## Architecture

### Database Schema

#### `emails` Table
```sql
ALTER TABLE emails ADD COLUMN one_time_link_only BOOLEAN DEFAULT FALSE;
```

#### `email_sessions` Table
```sql
CREATE TABLE email_sessions (
    session_id TEXT PRIMARY KEY,           -- UUID for the session record
    email_id TEXT NOT NULL,                -- Foreign key to emails table
    token_hash TEXT NOT NULL,              -- Argon2id hash of the session token
    expires_at DATETIME NOT NULL,          -- When the session token expires
    used BOOLEAN DEFAULT FALSE,            -- Whether the token has been used
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    user_agent TEXT,                       -- User-Agent string for audit
    ip_address TEXT,                       -- IP address for audit
    
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);
```

### Service Architecture

#### SessionTokenService Interface
```go
type SessionTokenService interface {
    GenerateSessionToken(emailID string, userAgent, clientIP string) (string, error)
    ValidateSessionToken(emailID, sessionToken string) (bool, error)
    MarkSessionTokenUsed(emailID, sessionToken string) error
    CleanupExpiredSessions() error
    GetSessionInfo(emailID, sessionToken string) (*SessionInfo, error)
    GetActiveSessions(emailID string) ([]*SessionInfo, error)
    RevokeSession(emailID, sessionToken string) error
}
```

## Implementation Details

### Token Generation
1. **High-Entropy Tokens**: 256-bit random tokens using crypto/rand
2. **Secure Hashing**: Argon2id with email ID as salt
3. **Expiration**: 5-minute lifetime for security
4. **Storage**: Only hashed tokens stored in database

### Token Validation
1. **Hash Verification**: Re-hash provided token and compare
2. **Expiration Check**: Verify token hasn't expired
3. **Usage Check**: Ensure token hasn't been used (for one-time links)
4. **Generic Errors**: Prevent information leakage

### One-Time Links
1. **Per-Email Toggle**: `one_time_link_only` boolean flag
2. **Automatic Marking**: Tokens marked as used after successful access
3. **Replay Prevention**: Used tokens cannot be reused

## API Endpoints

### Generate Session Token
```
POST /api/email/{id}/session
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
    "session_token": "abc123...",
    "expires_in": 300,
    "one_time_only": false
}
```

### Email Access with Session Token
```
GET /api/email/{id}?session_token=abc123...
Authorization: Bearer <jwt_token>
```

**Headers:**
```
X-Session-Token: abc123...
```

## Security Flow

### Updated Email Access Flow
1. **JWT Authentication**: Validate user identity
2. **Revoke/Time Checks**: Verify email hasn't been revoked or expired
3. **MFA/Decoy**: Handle multi-factor authentication
4. **Geofencing**: Check location-based restrictions
5. **Trusted Devices**: Verify device fingerprinting
6. **Session Token**: Validate session token (NEW)
7. **Read-Once**: Handle burn-after-read functionality

### Session Token Integration
- **Optional**: Session tokens are optional unless required by email settings
- **Query Parameter**: `?session_token=...`
- **Header Support**: `X-Session-Token: ...`
- **Generic Errors**: All failures return "Email has been revoked or cannot be accessed"

## Configuration

### Environment Variables
- No additional environment variables required
- Token expiration is hardcoded to 5 minutes
- Cleanup runs automatically via database queries

### Email Settings
- `one_time_link_only`: Boolean flag for one-time link mode
- Default: `false` (tokens can be reused until expiration)
- When `true`: Tokens become invalid after first use

## Testing

### Unit Tests
- **Mock Implementation**: `MockSessionTokenService` for testing
- **Token Generation**: Verify high-entropy token creation
- **Validation Logic**: Test expiration and usage checks
- **One-Time Links**: Verify token invalidation after use
- **Cleanup**: Test expired token removal

### Integration Tests
- **API Endpoints**: Test token generation and validation
- **Database Integration**: Verify proper storage and retrieval
- **Error Handling**: Test invalid tokens and edge cases
- **Audit Logging**: Verify comprehensive event tracking

## Security Considerations

### Token Security
- **High Entropy**: 256-bit random tokens
- **Secure Hashing**: Argon2id with email-specific salt
- **Short Lifetime**: 5-minute expiration
- **No Raw Storage**: Only hashed tokens in database

### Information Leakage Prevention
- **Generic Errors**: All failures return same error message
- **No Token Exposure**: Raw tokens never logged
- **Audit Trail**: Comprehensive logging for security analysis

### Replay Attack Prevention
- **One-Time Use**: Tokens can be configured for single use
- **Short Expiration**: 5-minute window limits attack surface
- **Email-Specific**: Tokens tied to specific email IDs

## Monitoring and Maintenance

### Audit Logging
- **Token Issuance**: Log when tokens are generated
- **Token Usage**: Track successful and failed validations
- **One-Time Links**: Monitor usage patterns
- **Security Events**: Flag suspicious activity

### Cleanup Process
- **Automatic Cleanup**: Expired tokens removed via SQL
- **Manual Cleanup**: Admin endpoints for bulk operations
- **Performance**: Indexed queries for efficient cleanup

## Future Enhancements

### Potential Improvements
- **Configurable Expiration**: Allow per-email token lifetime
- **Token Revocation**: Allow manual token invalidation
- **Rate Limiting**: Prevent token generation abuse
- **Advanced Analytics**: Token usage patterns and security insights

### Integration Opportunities
- **Notification System**: Alert on suspicious token usage
- **Geofencing**: Token restrictions based on location
- **Device Fingerprinting**: Token tied to specific devices
- **Multi-Channel**: Token delivery via multiple channels

## Troubleshooting

### Common Issues
1. **Token Expiration**: Tokens expire after 5 minutes
2. **One-Time Links**: Tokens become invalid after first use
3. **Database Cleanup**: Expired tokens automatically removed
4. **Generic Errors**: All failures return same message for security

### Debugging
- **Audit Logs**: Check `audit_log` table for token events
- **Database Queries**: Verify token storage and cleanup
- **Service Logs**: Monitor session token service operations
- **Network Analysis**: Check token transmission and validation

## Conclusion

The session tokens feature provides a robust security layer for email access, significantly reducing the risk of replay attacks and unauthorized access. The implementation follows security best practices with comprehensive audit logging, secure token handling, and flexible configuration options.
