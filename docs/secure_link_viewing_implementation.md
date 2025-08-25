# Secure Link Viewing Implementation

## Overview

This document describes the implementation of the public secure link viewing functionality for external recipients. The system allows external users to access secure emails through a public URL without requiring authentication, while enforcing security measures like password protection, MFA, and geolocation restrictions.

## Architecture

### Backend Components

#### 1. Public Secure Link Handlers (`cmd/api/public_secure_link_handlers.go`)

**Endpoints:**
- `GET /v/{linkID}` - Public secure link metadata access
- `POST /v/{linkID}/validate` - Security validation (password, MFA, geo)
- `POST /v/{linkID}/content` - Secure email content retrieval

**Key Features:**
- Rate limiting and brute-force protection
- Comprehensive audit logging
- Security validation with decoy messages
- IP address tracking and geolocation validation

#### 2. Security Validation Flow

The security validation process follows a multi-step approach:

1. **Password Validation** - Argon2id password verification
2. **MFA Validation** - TOTP or Email code verification
3. **Geolocation Validation** - IP-based location verification
4. **Access Control** - Read-once, auto-destruct, and time-lock enforcement

#### 3. Audit Logging

Every external access is logged with:
- Link ID and recipient IP
- User agent and timestamp
- Security validation attempts
- SES transaction ID linkage
- Suspicious activity flagging

### Frontend Components

#### 1. SecureEmailViewer (`src/components/external/SecureEmailViewer.tsx`)

**Features:**
- Landing page for `/v/{linkID}` URLs
- Displays sender information and security features
- Handles security validation flow
- Shows secure email content after validation
- Respects read-once and auto-destruct settings

**State Management:**
- Loading states for metadata and content
- Error handling with user-friendly messages
- Security validation step tracking
- Modal integration for security checks

#### 2. SecurityValidationModal (`src/components/external/SecurityValidationModal.tsx`)

**Features:**
- Multi-step security validation UI
- Password input with attempt tracking
- MFA code input (TOTP/Email)
- Geolocation confirmation
- Decoy message display on failures

**Validation Steps:**
1. **Password Step** - Password input with validation
2. **MFA Step** - Verification code input
3. **Geo Step** - Location confirmation
4. **Complete Step** - Success confirmation

## API Endpoints

### GET /v/{linkID}

**Purpose:** Retrieve secure link metadata for public access

**Response:**
```json
{
  "link_id": "secure-link-123",
  "subject": "Secure Message Subject",
  "sender_email": "sender@example.com",
  "sender_name": "John Doe",
  "security_settings": {
    "require_password": true,
    "require_mfa": true,
    "mfa_type": "email",
    "geolocation_restriction": true,
    "allowed_countries": ["United States"],
    "allowed_cities": ["New York"],
    "time_lock": false,
    "read_once": true,
    "auto_destruct": false,
    "max_access_attempts": 5,
    "current_attempts": 0
  },
  "status": "active",
  "message": "Link is accessible"
}
```

### POST /v/{linkID}/validate

**Purpose:** Validate security requirements for secure link access

**Request:**
```json
{
  "link_id": "secure-link-123",
  "password": "userpassword",
  "mfa_code": "123456",
  "mfa_type": "email",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

**Response:**
```json
{
  "success": true,
  "validated": true,
  "requires_mfa": false,
  "requires_geo": false,
  "error": null,
  "error_code": null,
  "decoy_message": null
}
```

### POST /v/{linkID}/content

**Purpose:** Retrieve secure email content after successful validation

**Request:**
```json
{
  "link_id": "secure-link-123",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

**Response:**
```json
{
  "link_id": "secure-link-123",
  "subject": "Secure Message Subject",
  "body": "<html>Secure email content...</html>",
  "sender_email": "sender@example.com",
  "sender_name": "John Doe",
  "read_once": true,
  "auto_destruct": false
}
```

## Security Features

### 1. Password Protection
- Argon2id hashing for password verification
- Configurable password requirements
- Brute-force protection with attempt limiting

### 2. Multi-Factor Authentication
- TOTP (Time-based One-Time Password)
- Email verification codes
- Configurable MFA types per link

### 3. Geolocation Restrictions
- Country-level restrictions
- City-level restrictions
- IP-based geolocation validation
- Mock geolocation service for testing

### 4. Access Control
- **Read-once**: Email content destroyed after first viewing
- **Auto-destruct**: Link destroyed after failed attempts
- **Time-lock**: Access restricted until specific time
- **Expiration**: Automatic link expiration

### 5. Audit Logging
- Comprehensive access logging
- SES transaction ID tracking
- Suspicious activity detection
- IP address and user agent tracking

## Error Handling

### Error Codes
- `LINK_NOT_FOUND` - Invalid or non-existent link
- `LINK_EXPIRED` - Link has expired
- `LINK_REVOKED` - Link has been revoked
- `LINK_DESTROYED` - Link destroyed due to failed attempts
- `ACCESS_DENIED` - Security validation failed
- `RATE_LIMITED` - Too many attempts
- `GEO_RESTRICTED` - Location not allowed

### Decoy Messages
- Fake error messages for security
- Prevents information leakage
- Configurable per security level

## Testing

### Unit Tests
- Email service validation (`pkg/securelinks/email/service_test.go`)
- Security validation logic
- Error handling scenarios

### Integration Tests
- End-to-end secure link flow (`tests/test_secure_link_viewing.ps1`)
- Public endpoint testing
- Security validation testing
- Auto-destruct behavior testing
- Audit logging verification

### Test Scenarios
1. **Valid Link Access** - Successful security validation
2. **Invalid Link Access** - 404 for non-existent links
3. **Expired Link Access** - 410 for expired links
4. **Security Validation** - Password, MFA, and geo validation
5. **Auto-Destruct** - Link destruction after failed attempts
6. **Audit Logging** - Verification of access logging

## Configuration

### Environment Variables
```bash
# Base URL for secure links
BASE_URL=http://localhost:8080

# SES configuration
AWS_SES_REGION=us-east-1
AWS_SES_ACCESS_KEY_ID=your-access-key
AWS_SES_SECRET_ACCESS_KEY=your-secret-key

# Rate limiting
MAX_ACCESS_ATTEMPTS=5
RATE_LIMIT_WINDOW=300
```

### Database Schema
The implementation uses existing tables:
- `secure_links` - Link metadata and security settings
- `link_audit_log` - Access logging and audit trail
- `ses_transactions` - Email delivery tracking

## Performance Considerations

### Caching
- Link metadata caching for frequently accessed links
- Security validation result caching
- Geolocation data caching

### Rate Limiting
- IP-based rate limiting
- Per-link attempt limiting
- Global rate limiting for public endpoints

### Database Optimization
- Indexed queries on link_id
- Efficient audit log queries
- Connection pooling

## Security Considerations

### Input Validation
- Strict validation of all inputs
- SQL injection prevention
- XSS protection in email content

### Access Control
- No authentication required for public access
- Security validation enforced per link
- Audit trail for all access attempts

### Data Protection
- Encrypted storage of sensitive data
- Secure transmission of content
- Proper cleanup of expired/destroyed links

## Future Enhancements

### Planned Features
1. **Reply Functionality** - Allow external users to reply securely
2. **Attachment Support** - Secure file attachments
3. **Advanced Analytics** - Detailed access analytics
4. **Mobile Optimization** - Enhanced mobile experience
5. **API Rate Limiting** - Advanced rate limiting strategies

### Scalability Improvements
1. **CDN Integration** - Content delivery network
2. **Load Balancing** - Multiple server instances
3. **Database Sharding** - Horizontal scaling
4. **Microservices** - Service decomposition

## Monitoring and Alerting

### Key Metrics
- Public link access rates
- Security validation success rates
- Failed access attempts
- Geographic access patterns
- Performance metrics

### Alerts
- Suspicious access patterns
- High failure rates
- Geographic anomalies
- Performance degradation
- Security incidents

## Conclusion

The secure link viewing implementation provides a robust, secure, and user-friendly way for external recipients to access secure emails. The system balances security with usability, providing comprehensive audit logging and protection against various attack vectors while maintaining a smooth user experience.

The implementation is production-ready with comprehensive testing, proper error handling, and scalability considerations for future growth.
