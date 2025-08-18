# Secure Email MVP API Documentation

## Overview

This document provides comprehensive API documentation for the Secure Email MVP system, including authentication, email management, enterprise features, and the Zero-Knowledge Identity Layer (ZKID).

## Authentication

### Login
- **Endpoint**: `POST /api/auth/login`
- **Description**: Authenticate user with email and password
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "userpassword"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "token": "jwt_token_here",
      "user": {
        "id": "user_uuid",
        "email": "user@example.com",
        "role": "enterprise_user"
      }
    }
  }
  ```

### Signup
- **Endpoint**: `POST /api/auth/signup`
- **Description**: Register new user account
- **Request Body**:
  ```json
  {
    "email": "newuser@example.com",
    "password": "securepassword",
    "fallback_email": "backup@example.com"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "user_id": "new_user_uuid",
      "message": "Account created successfully"
    }
  }
  ```

## Email Management

### Send Secure Email
- **Endpoint**: `POST /api/emails/send`
- **Description**: Send encrypted email with optional security features
- **Headers**: `Authorization: Bearer <jwt_token>`
- **Request Body**:
  ```json
  {
    "to": "recipient@example.com",
    "subject": "Secure Message",
    "content": "This is a secure message",
    "password": "optional_password",
    "expires_in": 3600,
    "burn_after_read": true
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "email_id": "email_uuid",
      "access_url": "https://secure-email.com/access/email_uuid"
    }
  }
  ```

### Read Email
- **Endpoint**: `GET /api/emails/{email_id}`
- **Description**: Read encrypted email content
- **Query Parameters**:
  - `password`: Optional password for protected emails
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "from": "sender@example.com",
      "subject": "Secure Message",
      "content": "Decrypted message content",
      "created_at": "2024-01-01T12:00:00Z",
      "expires_at": "2024-01-01T13:00:00Z"
    }
  }
  ```

## Enterprise Features

### Organization Management

#### Create Organization
- **Endpoint**: `POST /api/admin/organizations`
- **Description**: Create new enterprise organization
- **Headers**: `Authorization: Bearer <admin_token>`
- **Request Body**:
  ```json
  {
    "name": "Acme Corporation",
    "domain": "acme.com",
    "admin_email": "admin@acme.com"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "organization_id": "org_uuid",
      "name": "Acme Corporation",
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
  ```

#### Get Organization Details
- **Endpoint**: `GET /api/admin/organizations/{org_id}`
- **Description**: Get organization information
- **Headers**: `Authorization: Bearer <admin_token>`
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": "org_uuid",
      "name": "Acme Corporation",
      "domain": "acme.com",
      "user_count": 150,
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
  ```

### Compliance Reporting

#### Get Compliance Summary
- **Endpoint**: `GET /api/admin/compliance/summary`
- **Description**: Get compliance summary for organization
- **Headers**: `Authorization: Bearer <admin_token>`
- **Query Parameters**:
  - `organization_id`: Organization UUID
  - `start_date`: Start date (YYYY-MM-DD)
  - `end_date`: End date (YYYY-MM-DD)
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "total_emails": 1250,
      "encrypted_emails": 1250,
      "compliance_rate": 100.0,
      "period": "2024-01-01 to 2024-01-31"
    }
  }
  ```

#### Get Compliance Logs
- **Endpoint**: `GET /api/admin/compliance/logs`
- **Description**: Get detailed compliance logs
- **Headers**: `Authorization: Bearer <admin_token>`
- **Query Parameters**:
  - `organization_id`: Organization UUID
  - `page`: Page number (default: 1)
  - `limit`: Items per page (default: 50)
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "logs": [
        {
          "id": "log_uuid",
          "user_id": "user_uuid",
          "action": "email_sent",
          "timestamp": "2024-01-01T12:00:00Z",
          "details": {
            "email_id": "email_uuid",
            "encrypted": true
          }
        }
      ],
      "pagination": {
        "page": 1,
        "limit": 50,
        "total": 1250,
        "pages": 25
      }
    }
  }
  ```

## Zero-Knowledge Identity Layer (ZKID)

### Internal ZKID Operations

#### Create/Update Email Mapping
- **Endpoint**: `POST /api/zkid/mapping`
- **Description**: Create or update encrypted email mapping
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "user_id": "user_uuid",
    "email": "user@example.com",
    "fallback_email": "backup@example.com"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": "mapping_uuid",
      "user_id": "user_uuid"
    }
  }
  ```

#### Get Email by User ID
- **Endpoint**: `GET /api/zkid/email`
- **Description**: Retrieve encrypted email by user UUID
- **Headers**: `Authorization: Bearer <token>`
- **Query Parameters**:
  - `user_id`: User UUID
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "email": "user@example.com"
    }
  }
  ```

### Admin ZKID Operations (RBAC Protected)

#### Generate Recovery Codes
- **Endpoint**: `GET /api/admin/zkid/recovery-codes`
- **Description**: Generate recovery codes for user (UUID-only)
- **Headers**: `Authorization: Bearer <admin_token>`
- **Required Role**: `system_admin` or `enterprise_admin`
- **Query Parameters**:
  - `user_id`: User UUID
  - `count`: Number of codes to generate (default: 10, max: 50)
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "user_id": "user_uuid",
      "count": 5,
      "codes": [
        "abc123def456",
        "ghi789jkl012",
        "mno345pqr678",
        "stu901vwx234",
        "yz567abc890"
      ]
    }
  }
  ```

#### Revoke Recovery Code
- **Endpoint**: `POST /api/admin/zkid/revoke-code`
- **Description**: Revoke specific recovery code
- **Headers**: `Authorization: Bearer <admin_token>`
- **Required Role**: `system_admin` or `enterprise_admin`
- **Request Body**:
  ```json
  {
    "user_id": "user_uuid",
    "code_id": "code_uuid"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "user_id": "user_uuid",
      "code_id": "code_uuid",
      "revoked": true,
      "message": "Recovery code revoked successfully"
    }
  }
  ```

#### Get ZKID Statistics
- **Endpoint**: `GET /api/admin/zkid/stats`
- **Description**: Get ZKID layer statistics
- **Headers**: `Authorization: Bearer <admin_token>`
- **Required Role**: `system_admin` or `enterprise_admin`
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "enabled": true,
      "total_mappings": 1250,
      "total_recovery_codes": 6250,
      "used_recovery_codes": 1250,
      "available_recovery_codes": 5000
    }
  }
  ```

## PQC Layer (Post-Quantum Cryptography)

### PQC Status
- **Endpoint**: `GET /api/pqc/status`
- **Description**: Get PQC layer status and configuration
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "enabled": true,
      "algorithm": "Kyber-1024",
      "key_rotation_enabled": true,
      "last_rotation": "2024-01-01T12:00:00Z"
    }
  }
  ```

### PQC Key Rotation
- **Endpoint**: `POST /api/admin/pqc/rotate-keys`
- **Description**: Trigger PQC key rotation
- **Headers**: `Authorization: Bearer <admin_token>`
- **Required Role**: `system_admin`
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Key rotation initiated",
      "new_key_id": "new_key_uuid"
    }
  }
  ```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": "Additional error details"
  }
}
```

### Common Error Codes

- `AUTH_REQUIRED`: Authentication required
- `ADMIN_REQUIRED`: Admin privileges required
- `INVALID_REQUEST`: Invalid request format
- `NOT_FOUND`: Resource not found
- `FORBIDDEN`: Access denied
- `INTERNAL_ERROR`: Server error

## Rate Limiting

- **Authentication endpoints**: 5 requests per minute per IP
- **Email endpoints**: 10 requests per minute per user
- **Admin endpoints**: 30 requests per minute per admin
- **ZKID endpoints**: 20 requests per minute per user

## Security Headers

All responses include security headers:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

## Environment Variables

### Required for ZKID
```bash
ZKID_ENABLED=true
ZKID_MASTER_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
ZKID_EMAIL_HASH_PEPPER=deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef
ZKID_RECOVERY_PEPPER=feedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeef
```

### Required for PQC
```bash
PQC_ENABLED=true
PQC_ALGORITHM=Kyber-1024
PQC_KEY_ROTATION_ENABLED=true
```

### Required for Enterprise
```bash
ENABLE_ENTERPRISE_MULTI_TENANCY=true
```

## Testing

### Test Environment
- Base URL: `http://localhost:8080`
- Test database: In-memory SQLite
- Test keys: Provided in test configuration

### Example Test Requests

```bash
# Test ZKID functionality
curl -X POST "http://localhost:8080/api/zkid/mapping" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"user_id":"test-uuid","email":"test@example.com"}'

# Test admin ZKID operations
curl -X GET "http://localhost:8080/api/admin/zkid/stats" \
  -H "Authorization: Bearer <admin_token>"

# Test compliance reporting
curl -X GET "http://localhost:8080/api/admin/compliance/summary?organization_id=test-org" \
  -H "Authorization: Bearer <admin_token>"
```

## Support

For API support and questions:
- Documentation: `/docs/`
- Technical support: `support@company.com`
- Security issues: `security@company.com`
