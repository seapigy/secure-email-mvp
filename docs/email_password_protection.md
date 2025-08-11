# Email Password Protection for Secure Email Access

## Overview

**Micro-Iteration 4.14** implements optional per-email password protection functionality for the Secure Email MVP. This feature allows senders to add an additional layer of security by requiring a password to access specific emails, integrating seamlessly with existing security features like MFA, geolocation restrictions, and brute-force protection.

## Key Features

- **Optional Password Protection**: Senders can optionally require a password for email access
- **Strong Password Hashing**: Argon2id with random salt for secure password storage
- **Password Strength Validation**: Enforces minimum requirements and rejects common weak passwords
- **Multi-Layer Integration**: Works alongside MFA, geolocation, and brute-force protection
- **Generic Error Messages**: Security-focused responses that don't reveal system details
- **Automatic Reset**: Failed attempts reset on successful access
- **Brute-Force Protection**: Integrates with existing IP and per-email lockout systems

## Database Schema

### New Fields in `emails` Table

```sql
ALTER TABLE emails ADD COLUMN is_password_protected BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN password_hash TEXT;
ALTER TABLE emails ADD COLUMN password_salt TEXT;

CREATE INDEX IF NOT EXISTS idx_emails_password_protection ON emails(is_password_protected);
```

### Field Details

- **`is_password_protected`**: Boolean flag indicating if email requires password for access
- **`password_hash`**: Argon2id hash of the email password (stored as base64)
- **`password_salt`**: Random salt used for password hashing (stored as base64)

## Backend Implementation

### Email Password Service

The `pkg/emailpassword` package provides the core functionality:

```go
// Initialize email password service
emailPasswordService := emailpassword.NewEmailPasswordService(db)

// Set password for an email
err := emailPasswordService.SetEmailPassword(emailID, rawPassword)

// Check password for an email
valid, err := emailPasswordService.CheckEmailPassword(emailID, providedPassword)

// Clear password protection
err := emailPasswordService.ClearEmailPassword(emailID)

// Validate password strength
err := emailPasswordService.ValidatePasswordStrength(password)
```

### Argon2id Configuration

Default configuration settings:

```go
type Argon2Config struct {
    Memory      64 * 1024  // 64 MiB
    Time        3          // 3 iterations
    Parallelism 2          // 2 threads
    KeyLength   32         // 32 bytes
    SaltLength  16         // 16 bytes
}
```

### Integration Points

#### 1. Email Send Flow

Password protection is integrated into the email send flow in `send_email_handler.go`:

```go
// Validate and process password protection
var isPasswordProtectedInt int = 0
if req.Password != "" {
    // Validate password strength
    emailPasswordService := emailpassword.NewEmailPasswordService(srv.db)
    if err := emailPasswordService.ValidatePasswordStrength(req.Password); err != nil {
        // Return error response
    }
    isPasswordProtectedInt = 1
}

// Store email with password protection flag
// Set password after email creation
if req.Password != "" {
    if err := emailPasswordService.SetEmailPassword(emailID, req.Password); err != nil {
        // Return error response
    }
}
```

#### 2. Email Access Flow

Password protection is integrated into the email access flow in `view_email_handler.go`:

```go
// Security flow order:
// 1. Authentication Check
// 2. IP-Based Lockout Check (Micro-Iteration 4.13)
// 3. Geolocation Check (if restrictions set)
// 4. Per-Email Brute-Force Lockout Check (Micro-Iteration 4.12)
// 5. Password Check (if password-protected) - NEW
// 6. MFA Check (if enabled)
// 7. Email Decryption
```

#### 3. Password Validation

```go
// Check password protection
if isPasswordProtected == 1 {
    // Check if password is provided in request body
    var requestBody map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        // Return password required error
    }

    password, ok := requestBody["password"].(string)
    if !ok || password == "" {
        // Return password required error
    }

    // Validate password
    emailPasswordService := emailpassword.NewEmailPasswordService(srv.db)
    valid, err := emailPasswordService.CheckEmailPassword(emailID, password)
    if err != nil {
        // Return validation error
    }

    if !valid {
        // Increment brute-force and IP tracking attempts
        // Return access denied
    }

    // Reset failed attempts on successful validation
}
```

## Security Features

### 1. Strong Password Hashing

- **Argon2id**: Industry-standard password hashing algorithm
- **Random Salt**: 16-byte random salt for each password
- **Configurable Parameters**: Memory, time, and parallelism settings
- **Constant-Time Comparison**: Prevents timing attacks

### 2. Password Strength Validation

```go
// Password requirements:
// - Minimum 8 characters
// - Maximum 128 characters
// - Not in common weak password list
// - Case-sensitive validation
```

### 3. Generic Error Messages

All password-related responses use generic messages to prevent information leakage:

```json
{
  "error": "Access denied"
}
```

### 4. Brute-Force Integration

Password failures trigger existing brute-force protection:

- **Per-Email Lockout**: Failed attempts tracked per email ID
- **IP-Based Lockout**: Failed attempts tracked per IP address
- **Automatic Reset**: Failed attempts reset on successful access
- **Generic Responses**: No indication of lockout status

### 5. Multi-Layer Security

Password protection works alongside existing security features:

- **MFA**: Password check before MFA validation
- **Geolocation**: Password check after geolocation validation
- **Brute-Force**: Password failures increment brute-force counters
- **IP Tracking**: Password failures increment IP tracking counters

## API Changes

### Send Email Endpoint

**POST** `/api/email/send`

New optional field in request body:

```json
{
  "recipient": "user@example.com",
  "subject": "Secure Email",
  "body": "Email content",
  "password": "optionalpassword123"  // NEW: Optional password for access
}
```

### View Email Endpoint

**POST** `/api/email/view/{emailID}`

New request body format for password-protected emails:

```json
{
  "password": "emailpassword123"
}
```

**Response for password-required emails:**

```json
{
  "error": "Password required",
  "code": "password_required"
}
```

## Configuration

### Default Settings

- **Password Length**: 8-128 characters
- **Hash Algorithm**: Argon2id
- **Memory Cost**: 64 MiB
- **Time Cost**: 3 iterations
- **Parallelism**: 2 threads
- **Salt Length**: 16 bytes
- **Key Length**: 32 bytes

### Environment Variables

Currently using database defaults, but can be extended to support environment variable configuration:

```go
// Future enhancement: Configurable via environment variables
passwordMinLength := os.Getenv("EMAIL_PASSWORD_MIN_LENGTH")
passwordMaxLength := os.Getenv("EMAIL_PASSWORD_MAX_LENGTH")
```

## Testing

### Unit Tests

```bash
go test ./pkg/emailpassword -v
```

Tests cover:
- Password setting and validation
- Salt uniqueness and hash verification
- Password strength validation
- Error handling and edge cases
- Configuration validation

### Integration Tests

```bash
./scripts/test_email_password_protection.ps1
```

Tests cover:
- Password-protected email sending
- Password validation scenarios
- Weak password rejection
- Brute-force protection integration
- Multi-layer security integration

## Usage Examples

### Example 1: Send Password-Protected Email

```bash
curl -X POST http://localhost:8080/api/email/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "recipient@example.com",
    "subject": "Confidential Information",
    "body": "This email contains sensitive information.",
    "password": "securepassword123"
  }'
```

### Example 2: Access Password-Protected Email

```bash
curl -X POST http://localhost:8080/api/email/view/EMAIL_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "password": "securepassword123"
  }'
```

### Example 3: Access Non-Password-Protected Email

```bash
curl -X GET http://localhost:8080/api/email/view/EMAIL_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Security Considerations

### 1. Password Storage

- **No Plaintext Storage**: Passwords are never stored in plaintext
- **Unique Salts**: Each password uses a unique random salt
- **Strong Hashing**: Argon2id provides resistance against GPU/ASIC attacks
- **Configurable Parameters**: Can be adjusted for security vs performance

### 2. Attack Prevention

- **Brute-Force Protection**: Failed attempts trigger lockouts
- **IP Tracking**: Prevents attacks from specific IP addresses
- **Generic Responses**: No information leakage about password status
- **Rate Limiting**: Integrates with existing rate limiting

### 3. User Experience

- **Optional Feature**: Only applies when password is set
- **Clear Error Messages**: Users know when password is required
- **Automatic Reset**: Failed attempts reset on success
- **No Impact on Normal Emails**: Non-protected emails work normally

### 4. Integration

- **Seamless Integration**: Works with all existing security features
- **No Conflicts**: No interference with MFA, geolocation, etc.
- **Consistent Behavior**: Follows same patterns as other security layers
- **Comprehensive Coverage**: All security failures tracked consistently

## Monitoring and Logging

### Log Messages

The system logs various password protection events:

```
Password validation failed: password too short
Password validation successful for email abc123
Failed to set email password: database error
Password required but no password provided for email abc123
```

### Database Monitoring

Monitor the following fields for security analysis:

```sql
-- Check password-protected emails
SELECT email_id, is_password_protected, created_at 
FROM emails 
WHERE is_password_protected = 1;

-- Check for emails with password hashes
SELECT email_id, password_hash IS NOT NULL as has_password
FROM emails 
WHERE is_password_protected = 1;
```

## Troubleshooting

### Common Issues

1. **Password Not Working**
   - Check password strength requirements
   - Verify password is being sent in request body
   - Check for typos or case sensitivity

2. **Weak Password Rejected**
   - Ensure password is at least 8 characters
   - Avoid common passwords like "password", "123456"
   - Use a mix of letters, numbers, and symbols

3. **Access Denied After Multiple Attempts**
   - Wait for lockout period to expire
   - Check if IP is locked out
   - Verify correct password is being used

4. **Migration Issues**
   - Check database migration was applied
   - Verify password protection package is imported
   - Check logs for initialization errors

### Debug Information

Enable debug logging to see:
- Password validation attempts
- Hash generation and verification
- Database query results
- Error conditions
- Security event tracking

## Future Enhancements

### Potential Improvements

1. **Password Policies**: Configurable password requirements
2. **Password History**: Prevent reuse of recent passwords
3. **Password Expiration**: Time-based password expiration
4. **Password Reset**: Self-service password reset functionality
5. **Password Sharing**: Secure password sharing mechanisms

### Configuration Options

1. **Per-System Settings**: System-wide password policy
2. **Per-User Settings**: User-specific password requirements
3. **Password Complexity**: Configurable complexity rules
4. **Password Expiration**: Time-based expiration policies

## Conclusion

The email password protection feature provides robust security for sensitive emails while maintaining a good user experience. The implementation is secure, performant, and fully integrated with existing security features.

### Key Benefits

- ✅ Provides additional layer of security for sensitive emails
- ✅ Uses industry-standard Argon2id password hashing
- ✅ Integrates seamlessly with existing security layers
- ✅ Prevents brute-force attacks through lockout mechanisms
- ✅ Maintains security without revealing system details
- ✅ Provides comprehensive monitoring and logging
- ✅ Automatically handles failed attempts and resets

The feature successfully meets all acceptance criteria and provides a valuable addition to the Secure Email MVP's security capabilities. The implementation is production-ready and provides a solid foundation for future security enhancements.
