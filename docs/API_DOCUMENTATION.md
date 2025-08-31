# API Documentation

## Overview

This document provides comprehensive documentation for the Secure Email API, including all endpoints, authentication methods, request/response formats, and security considerations.

## Base URL

```
Development: http://localhost:8080
Production: https://api.securesystem.email
```

## Authentication

### JWT Token Authentication

All API endpoints require authentication via JWT tokens, except for public endpoints.

**Header Format:**
```
Authorization: Bearer <jwt_token>
```

**Token Storage:**
- Tokens are stored in `localStorage` as `authToken`
- Tokens expire after 24 hours
- Automatic token refresh is handled by the frontend

## Error Handling

### Standard Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error details",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

### Common Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `UNAUTHORIZED` | Invalid or missing authentication | 401 |
| `FORBIDDEN` | Insufficient permissions | 403 |
| `NOT_FOUND` | Resource not found | 404 |
| `VALIDATION_ERROR` | Invalid request data | 400 |
| `RATE_LIMITED` | Too many requests | 429 |
| `INTERNAL_ERROR` | Server error | 500 |

## Core Endpoints

### Email Management

#### Send Secure Email

**Endpoint:** `POST /api/secure-email/send`

**Description:** Send a secure email with optional security features.

**Request Body:**
```typescript
interface SendEmailRequest {
  recipient: string;                    // Required: Email address
  subject: string;                      // Required: Email subject (max 200 chars)
  body: string;                         // Required: Email body (max 10,000 chars)
  
  // Password Protection
  password?: string;                    // Optional: Password for email access
  requirePasswordForEveryEmail?: boolean; // Optional: Require password for every email
  passwordPerEmail?: boolean;           // Optional: Maximum security - password per email
  
  // Self-Destruct Settings
  selfDestructAfterAttempts?: boolean;  // Optional: Enable self-destruct after failed attempts
  maxFailedAttempts?: number;           // Optional: Maximum failed attempts (1-10)
  
  // Geolocation Settings
  geolocationLock?: boolean;            // Optional: Enable geolocation lock
  geoVerificationType?: string;         // Optional: 'none' | 'country' | 'city' | 'city_country'
  geoCity?: string;                     // Optional: Allowed city name
  geoCountry?: string;                  // Optional: Allowed country name
  allowedCountries?: string[];          // Optional: List of allowed countries
  
  // Time-based Settings
  timeLock?: boolean;                   // Optional: Enable time lock
  unlockAfter?: string;                 // Optional: ISO 8601 timestamp for unlock
  
  // Auto-Destruct Settings
  autoDestruct?: boolean;               // Optional: Enable auto-destruct
  destructAfterViews?: number;          // Optional: Number of views before destruction (1-100)
  
  // Read Once
  readOnce?: boolean;                   // Optional: Enable read-once mode
  
  // Remote Revoke
  remoteRevoke?: boolean;               // Optional: Enable remote revocation
  
  // Decoy Message
  decoyMessage?: boolean;               // Optional: Enable decoy message
  decoySecret?: string;                 // Optional: Secret for decoy message (4-50 chars)
  
  // Metadata and Alerts
  stripMetadata?: boolean;              // Optional: Remove identifying information
  tamperAlerts?: boolean;               // Optional: Enable tamper alerts
  
  // Fingerprint Hash
  generateFingerprintHash?: boolean;    // Optional: Generate fingerprint hash
  fingerprintHash?: string;             // Optional: Custom fingerprint hash
}
```

**Response:**
```typescript
interface SendEmailResponse {
  status: 'success' | 'error';
  blob_id: string;                      // Unique blob identifier
  secure_link_url: string;              // Secure link for email access
  message?: string;                     // Success/error message
  code?: string;                        // Error code if applicable
}
```

**Example Request:**
```json
{
  "recipient": "user@example.com",
  "subject": "Secure Document",
  "body": "Please find the secure document attached.",
  "password": "SecurePass123!",
  "geolocationLock": true,
  "geoVerificationType": "country",
  "geoCountry": "United States",
  "autoDestruct": true,
  "destructAfterViews": 3,
  "readOnce": true,
  "stripMetadata": true,
  "tamperAlerts": true
}
```

**Example Response:**
```json
{
  "status": "success",
  "blob_id": "blob_abc123def456",
  "secure_link_url": "https://securesystem.email/v/abc123def456"
}
```

#### Get Secure Link Status

**Endpoint:** `GET /api/secure-email/link/{linkId}`

**Description:** Get the status and metadata of a secure link.

**Response:**
```typescript
interface SecureLinkResponse {
  link_id: string;
  status: 'active' | 'expired' | 'destroyed' | 'blocked';
  created_at: string;
  expires_at?: string;
  view_count: number;
  max_views?: number;
  security_settings: {
    password_protected: boolean;
    geolocation_locked: boolean;
    time_locked: boolean;
    read_once: boolean;
    auto_destruct: boolean;
  };
}
```

### Authentication

#### User Login

**Endpoint:** `POST /api/auth/login`

**Request Body:**
```typescript
interface LoginRequest {
  email: string;                        // User email
  password: string;                     // User password
  totp_code?: string;                   // Optional: TOTP code for 2FA
}
```

**Response:**
```typescript
interface LoginResponse {
  status: 'success' | 'error';
  token: string;                        // JWT token
  user: {
    id: string;
    email: string;
    name: string;
    role: 'user' | 'admin';
    mfa_enabled: boolean;
  };
  requires_totp?: boolean;              // If TOTP is required
  message?: string;                     // Error message if applicable
}
```

#### Get Current User

**Endpoint:** `GET /api/auth/user`

**Description:** Get current user information.

**Response:**
```typescript
interface UserResponse {
  id: string;
  email: string;
  name: string;
  role: 'user' | 'admin';
  mfa_enabled: boolean;
  created_at: string;
  last_login: string;
}
```

#### User Logout

**Endpoint:** `POST /api/auth/logout`

**Description:** Logout current user and invalidate token.

**Response:**
```json
{
  "status": "success",
  "message": "Logged out successfully"
}
```

### File Management

#### Upload Attachment

**Endpoint:** `POST /api/attachments/upload`

**Description:** Upload file attachments for secure emails.

**Request:** Multipart form data
- `file`: File to upload (max 10MB)
- `email_id`: Associated email ID

**Response:**
```typescript
interface UploadResponse {
  url: string;                          // File URL
  filename: string;                     // Original filename
  size: number;                         // File size in bytes
  mime_type: string;                    // File MIME type
}
```

### Security Features

#### Validate Email Address

**Endpoint:** `POST /api/validation/email`

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "valid": true,
  "message": "Email address is valid"
}
```

#### Health Check

**Endpoint:** `GET /api/health`

**Description:** Check API health and status.

**Response:**
```typescript
interface HealthCheckResponse {
  status: 'healthy' | 'degraded' | 'unhealthy';
  timestamp: string;
  version: string;
  services: {
    database: 'healthy' | 'degraded' | 'unhealthy';
    encryption: 'healthy' | 'degraded' | 'unhealthy';
    storage: 'healthy' | 'degraded' | 'unhealthy';
  };
}
```

### Notification Management

#### Get Notification Preferences

**Endpoint:** `GET /api/notifications/preferences`

**Response:**
```typescript
interface NotificationPreferences {
  email_notifications: boolean;
  sms_notifications: boolean;
  push_notifications: boolean;
  security_alerts: boolean;
  marketing_emails: boolean;
}
```

#### Update Notification Preferences

**Endpoint:** `PUT /api/notifications/preferences`

**Request Body:**
```typescript
interface UpdateNotificationPreferences {
  email_notifications?: boolean;
  sms_notifications?: boolean;
  push_notifications?: boolean;
  security_alerts?: boolean;
  marketing_emails?: boolean;
}
```

### Access Event History

#### Get Access Event History

**Endpoint:** `GET /api/security/access-events`

**Query Parameters:**
- `limit`: Number of events to return (default: 50, max: 100)
- `offset`: Number of events to skip (default: 0)
- `event_type`: Filter by event type
- `severity`: Filter by severity level
- `start_date`: Start date filter (ISO 8601)
- `end_date`: End date filter (ISO 8601)

**Response:**
```typescript
interface AccessEventHistory {
  events: Array<{
    event_id: string;
    email_id: string;
    user_id: string;
    ip_address: string;
    user_agent: string;
    timestamp: string;
    event_type: 'view' | 'failed_access' | 'security_violation';
    severity: 'low' | 'medium' | 'high' | 'critical';
    country: string;
    city: string;
    device_type: string;
    browser: string;
    os: string;
    is_mobile: boolean;
    is_tor: boolean;
    is_vpn: boolean;
    is_proxy: boolean;
    threat_intel_match: boolean;
    auto_blocked: boolean;
  }>;
  total_count: number;
  has_more: boolean;
}
```

## Security Considerations

### Rate Limiting

- **Authentication endpoints**: 5 requests per minute
- **Email sending**: 10 requests per minute
- **File uploads**: 20 requests per minute
- **General endpoints**: 100 requests per minute

### Input Validation

All inputs are validated and sanitized:

- **Email addresses**: RFC 5321 compliant validation
- **Passwords**: Minimum 6 characters, complexity requirements
- **File uploads**: Type and size restrictions
- **Geolocation data**: Format validation and sanitization
- **Timestamps**: ISO 8601 format validation

### Security Headers

The API includes the following security headers:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

### Encryption

- **Data in transit**: TLS 1.3 encryption
- **Data at rest**: AES-256 encryption
- **Passwords**: Argon2id hashing
- **JWT tokens**: RS256 signing

## SDK Usage

### JavaScript/TypeScript

```typescript
import { 
  sendSecureEmail, 
  loginUser, 
  getCurrentUser,
  uploadAttachment 
} from '@/lib/api';

// Send secure email
const response = await sendSecureEmail({
  recipient: 'user@example.com',
  subject: 'Secure Document',
  body: 'Please find the secure document attached.',
  password: 'SecurePass123!',
  geolocationLock: true,
  geoCountry: 'United States'
});

// Login user
const loginResponse = await loginUser({
  email: 'user@example.com',
  password: 'password123'
});

// Upload attachment
const uploadResponse = await uploadAttachment(file);
```

### Error Handling

```typescript
import { isApiError, getErrorMessage } from '@/lib/api';

try {
  const response = await sendSecureEmail(emailData);
  console.log('Email sent:', response.secure_link_url);
} catch (error) {
  if (isApiError(error)) {
    console.error('API Error:', getErrorMessage(error));
  } else {
    console.error('Network Error:', error.message);
  }
}
```

## Testing

### Test Endpoints

For testing purposes, the following endpoints are available:

- **Test Email**: `POST /api/test/send-email`
- **Test Authentication**: `POST /api/test/auth`
- **Test File Upload**: `POST /api/test/upload`

### Mock Data

Test endpoints accept a `test_mode` parameter that returns mock data:

```json
{
  "test_mode": true,
  "recipient": "test@example.com",
  "subject": "Test Email",
  "body": "This is a test email"
}
```

## Versioning

API versioning is handled through the URL path:

- **Current version**: `/api/v1/`
- **Legacy version**: `/api/v0/` (deprecated)

## Support

For API support and questions:

- **Documentation**: https://docs.securesystem.email
- **Support Email**: api-support@securesystem.email
- **Status Page**: https://status.securesystem.email
