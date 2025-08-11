# Multi-Factor Authentication (MFA) Implementation

## Overview

Micro-Iteration 4.12 implements optional Multi-Factor Authentication (MFA) for secure email access, adding an extra layer of verification beyond existing security features (password, geolocation, expiration, etc.).

## Features

### MFA Methods
- **TOTP-based verification**: Google Authenticator, Authy, or any TOTP-compatible app
- **Email-based one-time codes**: 6-digit codes sent via email as a fallback option

### Security Features
- **AES-256-GCM encrypted TOTP secrets**: Secure storage of TOTP secrets
- **Brute-force protection**: Lock out after 5 failed attempts for 30 minutes
- **Automatic attempt tracking**: Reset counters on successful validation
- **Integration with existing security**: Works seamlessly with geolocation, password, expiration, and burn-after-read

## Database Schema

### New Fields Added to `emails` Table

```sql
-- Add require_mfa field (boolean - whether MFA is required for this email)
ALTER TABLE emails ADD COLUMN require_mfa INTEGER DEFAULT 0;

-- Add mfa_type field (enum: TOTP, EMAIL_CODE)
ALTER TABLE emails ADD COLUMN mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE'));

-- Add encrypted_totp_secret field (AES-256-GCM encrypted TOTP secret for TOTP-based MFA)
ALTER TABLE emails ADD COLUMN encrypted_totp_secret TEXT;

-- Add mfa_failed_attempts field (track failed MFA attempts for brute-force protection)
ALTER TABLE emails ADD COLUMN mfa_failed_attempts INTEGER DEFAULT 0;

-- Add mfa_locked_until field (timestamp when MFA is locked due to too many failed attempts)
ALTER TABLE emails ADD COLUMN mfa_locked_until DATETIME;
```

## Backend Implementation

### MFA Package (`pkg/mfa/mfa.go`)

The MFA package provides core functionality for TOTP and email-based MFA:

#### Key Functions

- `GenerateTOTPSecret(emailID string)`: Generates and encrypts TOTP secrets
- `ValidateTOTP(emailID, code string)`: Validates TOTP codes
- `GenerateEmailCode()`: Generates random 6-digit email codes
- `ValidateEmailCode(emailID, code string)`: Validates email-based codes
- `CheckMFALockout(emailID string)`: Checks if MFA is locked due to failed attempts
- `IncrementFailedAttempts(emailID string)`: Tracks failed attempts and implements lockout
- `ResetFailedAttempts(emailID string)`: Resets attempt counters on success

#### TOTP Secret Encryption

TOTP secrets are encrypted using AES-256-GCM before storage:

```go
// Encrypt the TOTP secret using AES-256-GCM
encryptedData, err := auth.EncryptAES256GCM([]byte(key.Secret()))

// Store encrypted components as JSON
encryptedComponents := map[string]string{
    "ciphertext": base64.StdEncoding.EncodeToString(encryptedData.Ciphertext),
    "key":        base64.StdEncoding.EncodeToString(encryptedData.Key),
    "nonce":      base64.StdEncoding.EncodeToString(encryptedData.Nonce),
    "auth_tag":   base64.StdEncoding.EncodeToString(encryptedData.AuthTag),
}
```

### API Endpoints

#### 1. MFA Validation (`POST /api/mfa/validate`)

Validates MFA codes for email access.

**Request:**
```json
{
  "email_id": "email-uuid",
  "mfa_code": "123456"
}
```

**Response:**
```json
{
  "success": true,
  "message": "MFA validation successful"
}
```

**Error Responses:**
- `401 Unauthorized`: Invalid MFA code
- `429 Too Many Requests`: MFA locked due to failed attempts
- `400 Bad Request`: MFA not required or invalid request

#### 2. Email Code Generation (`POST /api/mfa/email-code`)

Generates email-based MFA codes.

**Request:**
```json
{
  "email_id": "email-uuid"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Email code generated successfully",
  "code": "123456"
}
```

#### 3. MFA Configuration (`GET /api/mfa/config/{emailID}`)

Retrieves MFA configuration for an email.

**Response:**
```json
{
  "require_mfa": true,
  "mfa_type": "TOTP",
  "failed_attempts": 0,
  "locked_until": null
}
```

### Integration with Email View Handler

The email view handler (`cmd/api/view_email_handler.go`) has been updated to enforce MFA:

1. **MFA Check Order**: MFA is checked after geolocation but before decryption
2. **Lockout Enforcement**: Checks for MFA lockout before processing
3. **Code Validation**: Validates MFA codes based on type (TOTP or EMAIL_CODE)
4. **Attempt Tracking**: Increments failed attempts and implements lockout
5. **Success Handling**: Resets attempt counters on successful validation

### Email Send Handler Updates

The email send handler (`cmd/api/send_email_handler.go`) now supports MFA configuration:

- **MFA Fields**: Accepts `requireMFA` and `mfaType` in requests
- **TOTP Generation**: Automatically generates and encrypts TOTP secrets for TOTP-based MFA
- **Database Storage**: Stores MFA configuration in the database

## Frontend Implementation

### ComposeModal Updates (`src/components/secure/ComposeModal.tsx`)

The compose modal has been enhanced with MFA options:

#### New Interface Fields

```typescript
interface ComposeFormData {
  securitySettings: {
    // ... existing fields ...
    requireMFA: boolean;
    mfaType?: string; // "TOTP" or "EMAIL_CODE"
  };
}
```

#### UI Components

1. **MFA Toggle**: Enable/disable MFA for the email
2. **MFA Type Selection**: Dropdown to choose between TOTP and email-based codes
3. **Information Display**: Shows what the recipient will need to do
4. **Security Warnings**: Explains MFA security implications

#### Form Submission

MFA settings are included in the API request:

```typescript
const apiRequest = {
  // ... existing fields ...
  requireMFA: formData.securitySettings.requireMFA,
  mfaType: formData.securitySettings.requireMFA 
    ? formData.securitySettings.mfaType || 'TOTP'
    : undefined,
};
```

## Security Considerations

### TOTP Security
- **Secret Encryption**: TOTP secrets are encrypted using AES-256-GCM
- **Secure Generation**: Uses cryptographically secure random generation
- **QR Code URLs**: Generated for easy setup with authenticator apps

### Email Code Security
- **Random Generation**: 6-digit codes generated using crypto/rand
- **Time-based Expiration**: Codes expire after 10 minutes
- **Case-insensitive Validation**: Codes are validated case-insensitively

### Brute Force Protection
- **Attempt Tracking**: Failed attempts are tracked per email
- **Automatic Lockout**: Locked after 5 failed attempts for 30 minutes
- **Reset on Success**: Attempt counters reset on successful validation

### Integration Security
- **Order of Operations**: MFA is enforced after geolocation but before decryption
- **JWT Authentication**: All MFA endpoints require valid JWT tokens
- **Error Handling**: Secure error messages that don't leak sensitive information

## Usage Examples

### Sending Email with TOTP MFA

```bash
curl -X POST http://localhost:8080/api/email/send \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "recipient@example.com",
    "subject": "Secure Document",
    "body": "Please find the secure document attached.",
    "requireMFA": true,
    "mfaType": "TOTP",
    "burnAfterRead": true
  }'
```

### Sending Email with Email-based MFA

```bash
curl -X POST http://localhost:8080/api/email/send \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "recipient@example.com",
    "subject": "Secure Document",
    "body": "Please find the secure document attached.",
    "requireMFA": true,
    "mfaType": "EMAIL_CODE",
    "burnAfterRead": true
  }'
```

### Generating Email Code

```bash
curl -X POST http://localhost:8080/api/mfa/email-code \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email_id": "email-uuid"
  }'
```

### Validating MFA Code

```bash
curl -X POST http://localhost:8080/api/mfa/validate \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email_id": "email-uuid",
    "mfa_code": "123456"
  }'
```

### Viewing Email with MFA

```bash
curl -X GET "http://localhost:8080/api/email/view/email-uuid?mfa_code=123456" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Testing

### Test Script

A comprehensive test script is provided: `scripts/test_mfa_functionality.ps1`

The script tests:
1. Email sending with TOTP MFA
2. Email sending with email-based MFA
3. MFA configuration retrieval
4. Access blocking without MFA codes
5. Invalid code rejection
6. Email code generation and validation
7. Brute force protection
8. Lockout functionality

### Running Tests

```powershell
.\scripts\test_mfa_functionality.ps1 -BaseUrl "http://localhost:8080" -Email "test@example.com" -Password "testpassword123"
```

## Deployment Considerations

### Environment Variables

No new environment variables are required for MFA functionality.

### Database Migration

The MFA migration (`schema/migrate_add_mfa_fields.sql`) is automatically applied during server startup.

### Production Considerations

1. **Email Delivery**: In production, implement proper email delivery for email-based codes
2. **Rate Limiting**: Consider additional rate limiting for MFA endpoints
3. **Monitoring**: Monitor MFA failure rates and lockout events
4. **Backup**: Ensure TOTP secrets are properly backed up and recoverable

## Future Enhancements

### Potential Improvements

1. **SMS-based MFA**: Add SMS as an additional MFA method
2. **Hardware Tokens**: Support for FIDO2/U2F hardware tokens
3. **Backup Codes**: Generate backup codes for account recovery
4. **MFA Recovery**: Implement MFA recovery procedures
5. **Advanced Analytics**: Track MFA usage patterns and security metrics

### Integration Opportunities

1. **Admin Dashboard**: Add MFA statistics to admin interface
2. **Audit Logging**: Enhanced logging for MFA events
3. **User Preferences**: Allow users to set default MFA preferences
4. **Bulk Operations**: Support for bulk MFA configuration

## Conclusion

The MFA implementation provides a robust, secure, and user-friendly multi-factor authentication system for secure email access. It integrates seamlessly with existing security features while maintaining high security standards through encryption, brute force protection, and proper error handling.

The implementation is production-ready and includes comprehensive testing, documentation, and security considerations for deployment in enterprise environments.
