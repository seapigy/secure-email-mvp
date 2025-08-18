# Admin Authentication System Documentation

## Overview

The Admin Authentication System provides secure, production-ready authentication for the Secure Email MVP Admin Dashboard. It implements a bootstrap-first approach where only the root admin (Christopher Pigusch) can initially set up the system, with extensible support for additional admins.

## Architecture

### Core Components

1. **Admin Authentication Service** (`pkg/admin/auth.go`)
   - Handles admin user management, authentication, and session management
   - Implements strong password hashing with Argon2id
   - Supports TOTP-based 2FA
   - Manages admin sessions with secure tokens

2. **Admin Middleware** (`pkg/admin/middleware.go`)
   - Provides role-based access control (RBAC)
   - Validates admin sessions
   - Enforces admin permissions

3. **Admin HTTP Handlers** (`cmd/api/admin_auth_handlers.go`)
   - RESTful API endpoints for admin operations
   - Handles setup, login, logout, and session validation
   - Provides audit logging capabilities

4. **Database Schema** (`schema/migrate_add_admin_users.sql`)
   - Admin users table with role management
   - Admin sessions table for secure session tracking
   - Admin audit logs for comprehensive logging
   - Admin invitation keys for secure admin creation

## Security Features

### Password Security
- **Minimum 16 characters** with complexity requirements:
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one digit
  - At least one special character
- **Argon2id hashing** with secure salt generation
- **Account lockout** after 5 failed attempts (30-minute lockout)

### Session Management
- **Secure session tokens** (UUID-based)
- **30-minute session expiration**
- **Automatic session cleanup**
- **IP address and user agent tracking**

### Two-Factor Authentication (2FA)
- **TOTP support** (Google Authenticator, Authy)
- **Hardware token support** (YubiKey)
- **Secure TOTP secret generation**
- **Optional 2FA enforcement**

### Audit Logging
- **Comprehensive action logging** for all admin operations
- **UUID-only identifiers** for privacy
- **IP address and user agent tracking**
- **Success/failure status tracking**
- **JSON-formatted details**

## API Endpoints

### Bootstrap Endpoints

#### `GET /admin/check-setup`
Checks if admin setup is required.

**Response:**
```json
{
  "setup_required": true,
  "root_admin_email": "cpigusch@gmail.com"
}
```

#### `POST /admin/setup`
Creates the initial root admin account.

**Request:**
```json
{
  "email": "cpigusch@gmail.com",
  "password": "SecureAdminPassword123!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Root admin created successfully",
  "admin_id": "uuid-here"
}
```

### Authentication Endpoints

#### `POST /admin/login`
Authenticates an admin user.

**Request:**
```json
{
  "email": "cpigusch@gmail.com",
  "password": "SecureAdminPassword123!",
  "totp_code": "123456"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "session_token": "uuid-here",
  "admin": {
    "id": "uuid-here",
    "email": "cpigusch@gmail.com",
    "role": "root_admin",
    "totp_enabled": false,
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### `POST /admin/logout`
Logs out an admin user.

**Request:**
```json
{
  "session_token": "uuid-here"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Logout successful"
}
```

#### `GET /admin/session`
Validates an admin session.

**Headers:**
```
Authorization: Bearer <session_token>
```

**Response:**
```json
{
  "success": true,
  "message": "Session valid",
  "admin": {
    "id": "uuid-here",
    "email": "cpigusch@gmail.com",
    "role": "root_admin",
    "totp_enabled": false,
    "is_active": true
  }
}
```

### Audit Endpoints

#### `GET /admin/audit-logs`
Retrieves admin audit logs.

**Headers:**
```
Authorization: Bearer <session_token>
```

**Query Parameters:**
- `limit` (optional): Number of logs to retrieve (default: 100)

**Response:**
```json
{
  "success": true,
  "logs": [
    {
      "id": "uuid-here",
      "admin_id": "uuid-here",
      "action": "admin_login_success",
      "resource_type": "admin_users",
      "resource_id": "uuid-here",
      "details": "{\"email\":\"cpigusch@gmail.com\",\"session_id\":\"uuid-here\"}",
      "ip_address": "127.0.0.1",
      "user_agent": "Mozilla/5.0...",
      "success": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

## Database Schema

### Admin Users Table
```sql
CREATE TABLE admin_users (
    id TEXT PRIMARY KEY,                    -- UUID for admin user
    email TEXT NOT NULL UNIQUE,             -- Admin email address
    password_hash TEXT NOT NULL,            -- Argon2id hash of password
    totp_secret TEXT,                       -- TOTP secret for 2FA
    totp_enabled BOOLEAN DEFAULT FALSE,     -- Whether 2FA is enabled
    role TEXT NOT NULL DEFAULT 'root_admin', -- Role: 'root_admin', 'full_admin', 'read_only_admin'
    is_active BOOLEAN DEFAULT TRUE,         -- Whether admin account is active
    last_login DATETIME,                    -- Last successful login timestamp
    failed_login_attempts INTEGER DEFAULT 0, -- Number of consecutive failed login attempts
    locked_until DATETIME,                  -- Account lockout until timestamp
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,                        -- UUID of admin who created this account
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL
);
```

### Admin Sessions Table
```sql
CREATE TABLE admin_sessions (
    id TEXT PRIMARY KEY,                    -- UUID for session
    admin_id TEXT NOT NULL,                 -- UUID of admin user
    session_token TEXT NOT NULL UNIQUE,     -- JWT session token
    refresh_token TEXT NOT NULL UNIQUE,     -- Refresh token for session renewal
    expires_at DATETIME NOT NULL,           -- When session expires
    ip_address TEXT,                        -- IP address where session was created
    user_agent TEXT,                        -- User agent string
    is_active BOOLEAN DEFAULT TRUE,         -- Whether session is active
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admin_users(id) ON DELETE CASCADE
);
```

### Admin Audit Logs Table
```sql
CREATE TABLE admin_audit_logs (
    id TEXT PRIMARY KEY,                    -- UUID for audit log entry
    admin_id TEXT,                          -- UUID of admin who performed action
    action TEXT NOT NULL,                   -- Action performed
    resource_type TEXT,                     -- Type of resource affected
    resource_id TEXT,                       -- UUID of resource affected
    details TEXT,                           -- JSON details of the action
    ip_address TEXT,                        -- IP address of admin
    user_agent TEXT,                        -- User agent string
    success BOOLEAN NOT NULL,               -- Whether action was successful
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admin_users(id) ON DELETE SET NULL
);
```

## Environment Variables

### Required
- `ROOT_ADMIN_EMAIL`: Email address for the root admin (default: "cpigusch@gmail.com")

### Optional
- `JWT_SECRET`: Secret for JWT token signing (if using JWT-based sessions)

## Usage Examples

### Frontend Integration

#### Check Setup Status
```javascript
const checkSetup = async () => {
  const response = await fetch('/admin/check-setup');
  const data = await response.json();
  
  if (data.setup_required) {
    // Show setup form
    showSetupForm();
  } else {
    // Show login form
    showLoginForm();
  }
};
```

#### Admin Login
```javascript
const login = async (email, password, totpCode = '') => {
  const response = await fetch('/admin/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email,
      password,
      totp_code: totpCode
    })
  });
  
  const data = await response.json();
  
  if (data.success) {
    // Store session token
    localStorage.setItem('adminSessionToken', data.session_token);
    // Redirect to dashboard
    window.location.href = '/admin/dashboard';
  } else {
    // Show error
    showError(data.message);
  }
};
```

#### Validate Session
```javascript
const validateSession = async () => {
  const token = localStorage.getItem('adminSessionToken');
  
  if (!token) {
    return false;
  }
  
  const response = await fetch('/admin/session', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  if (response.ok) {
    const data = await response.json();
    return data.success;
  }
  
  return false;
};
```

## Security Considerations

### Password Policy
- Enforces strong password requirements
- Uses Argon2id for secure hashing
- Implements account lockout protection

### Session Security
- Short session duration (30 minutes)
- Secure token generation
- Automatic session cleanup
- IP and user agent tracking

### Audit Trail
- Comprehensive logging of all admin actions
- UUID-only identifiers for privacy
- Detailed action tracking
- Success/failure status

### Rate Limiting
- Built-in protection against brute force attacks
- Account lockout after failed attempts
- IP-based rate limiting (can be extended)

## Testing

### Manual Testing
Use the provided test script:
```powershell
.\test_admin_auth.ps1
```

### Automated Testing
The system includes comprehensive unit tests for:
- Password validation and hashing
- TOTP generation and validation
- Session management
- Audit logging

## Deployment

### Database Migration
The admin authentication system automatically applies its database migration on startup:

```sql
-- Migration is applied automatically from:
-- schema/migrate_add_admin_users.sql
```

### Environment Setup
1. Set the `ROOT_ADMIN_EMAIL` environment variable
2. Ensure the database is accessible
3. Start the API server
4. Access `/admin/check-setup` to verify setup status

### Production Considerations
- Use HTTPS in production
- Configure proper CORS settings
- Set up monitoring for failed login attempts
- Implement backup and recovery procedures
- Regular audit log review

## Troubleshooting

### Common Issues

#### "Admin already exists" Error
- This occurs when trying to create a root admin when one already exists
- Use `/admin/check-setup` to verify current status
- Only one root admin can exist in the system

#### "Invalid credentials" Error
- Verify email and password are correct
- Check if account is locked due to failed attempts
- Ensure TOTP code is correct if 2FA is enabled

#### Session Validation Failures
- Check if session token is valid and not expired
- Verify Authorization header format: `Bearer <token>`
- Ensure session is still active in database

### Debugging
- Check application logs for detailed error messages
- Review audit logs for authentication attempts
- Verify database connectivity and schema
- Test with the provided PowerShell script

## Future Enhancements

### Planned Features
- Admin invitation system for additional admins
- Role-based permissions for different admin levels
- Enhanced audit reporting and analytics
- Integration with external authentication providers
- Advanced session management with refresh tokens

### Extensibility
The system is designed to be easily extensible:
- Modular architecture with clear separation of concerns
- Configurable security policies
- Pluggable authentication methods
- Comprehensive API for frontend integration
