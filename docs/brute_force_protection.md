# Brute-Force Protection for Email Access Attempts

## Overview

**Micro-Iteration 4.12** implements comprehensive brute-force protection for secure email access attempts. This feature prevents unauthorized access by tracking failed attempts and implementing timed lockouts, providing an additional layer of security beyond existing authentication mechanisms.

## Key Features

- **Per-Email Tracking**: Failed attempts are tracked individually for each email
- **Configurable Limits**: Default 3 failed attempts before lockout
- **Timed Lockouts**: Default 15-minute lockout duration
- **Generic Error Messages**: Security-focused responses that don't reveal lockout status
- **Automatic Reset**: Failed attempts reset on successful access
- **Multi-Layer Integration**: Works with MFA, geolocation, and password protection

## Database Schema

### New Fields Added to `emails` Table

```sql
-- Brute-force protection fields
ALTER TABLE emails ADD COLUMN brute_force_failed_attempts INTEGER DEFAULT 0;
ALTER TABLE emails ADD COLUMN brute_force_last_failed_attempt DATETIME;
ALTER TABLE emails ADD COLUMN brute_force_lockout_until DATETIME;
ALTER TABLE emails ADD COLUMN brute_force_max_attempts INTEGER DEFAULT 3;
ALTER TABLE emails ADD COLUMN brute_force_lockout_duration_minutes INTEGER DEFAULT 15;

-- Index for brute-force protection queries
CREATE INDEX IF NOT EXISTS idx_emails_brute_force ON emails(brute_force_failed_attempts, brute_force_lockout_until);
```

### Field Details

- **`brute_force_failed_attempts`**: Count of consecutive failed access attempts
- **`brute_force_last_failed_attempt`**: Timestamp of the last failed attempt
- **`brute_force_lockout_until`**: Timestamp until which access is locked out (NULL if not locked)
- **`brute_force_max_attempts`**: Maximum failed attempts before lockout (default: 3)
- **`brute_force_lockout_duration_minutes`**: Lockout duration in minutes (default: 15)

## Backend Implementation

### Brute-Force Protection Package

The `pkg/bruteforce` package provides the core functionality:

```go
// Initialize brute-force protection
bfProtection := bruteforce.NewBruteForceProtection(db)

// Check if email is locked out
locked, err := bfProtection.CheckLockout(emailID)
if locked {
    // Return generic "Access denied" response
}

// Increment failed attempts
err = bfProtection.IncrementFailedAttempt(emailID)

// Reset failed attempts on successful access
err = bfProtection.ResetFailedAttempts(emailID)
```

### Integration Points

#### 1. Email Access Flow

The brute-force protection is integrated into the email access flow in `view_email_handler.go`:

```go
// 1. Authentication Check
// 2. Geolocation Check (if restrictions set)
// 3. Brute-Force Lockout Check
// 4. MFA Check (if enabled)
// 5. Password Check (if required)
// 6. Email Decryption
```

#### 2. Security Failure Handling

Brute-force protection is triggered by all types of security failures:

- **MFA Failures**: Invalid TOTP or email codes
- **Geolocation Failures**: Location restrictions not met
- **Password Failures**: Invalid passwords (when implemented)
- **Authentication Failures**: Invalid tokens or unauthorized access

#### 3. Success Handling

Failed attempts are automatically reset on successful access:

```go
// Reset brute-force protection failed attempts on successful access
if err := bfProtection.ResetFailedAttempts(emailID); err != nil {
    log.Printf("Failed to reset brute-force attempts on successful access: %v", err)
}
```

## Security Features

### 1. Generic Error Messages

All lockout responses use generic "Access denied" messages to prevent information leakage:

```json
{
  "error": "Access denied"
}
```

### 2. Automatic Lockout Expiration

Lockouts automatically expire after the configured duration:

```go
// If lockout has expired, clear it
if lockoutUntil != nil && time.Now().After(*lockoutUntil) {
    err = bfProtection.clearLockout(emailID)
}
```

### 3. Server-Side Only

All lockout and attempt tracking is stored server-side only, preventing client-side manipulation.

### 4. Comprehensive Logging

All brute-force events are logged for monitoring and analysis:

```go
log.Printf("Email %s is locked out due to brute-force protection", emailID)
log.Printf("Failed to increment brute-force attempts: %v", err)
```

## Configuration

### Default Settings

- **Max Attempts**: 3 failed attempts
- **Lockout Duration**: 15 minutes
- **Reset on Success**: Automatic

### Environment Variables

Currently using database defaults, but can be extended to support environment variable configuration:

```go
// Future enhancement: Configurable via environment variables
maxAttempts := os.Getenv("BRUTE_FORCE_MAX_ATTEMPTS")
lockoutDuration := os.Getenv("BRUTE_FORCE_LOCKOUT_DURATION")
```

## API Behavior

### Successful Access

```http
GET /api/email/view/{emailID}
Authorization: Bearer {token}

Response: 200 OK
{
  "email_id": "...",
  "subject": "...",
  "body": "...",
  "status": "success"
}
```

### Failed Access (Generic Response)

```http
GET /api/email/view/{emailID}
Authorization: Bearer {token}

Response: 403 Forbidden
{
  "error": "Access denied"
}
```

### Lockout Response

```http
GET /api/email/view/{emailID}
Authorization: Bearer {token}

Response: 403 Forbidden
{
  "error": "Access denied"
}
```

## Testing

### Unit Tests

```bash
go test ./pkg/bruteforce -v
```

Tests cover:
- Lockout checking (no lockout, active lockout, expired lockout)
- Failed attempt incrementing
- Attempt resetting
- Status retrieval
- Error handling

### Integration Tests

```bash
./scripts/test_brute_force_protection.ps1
```

Tests cover:
- MFA failure scenarios
- Geolocation failure scenarios
- Lockout enforcement
- Success scenarios
- Generic error messages

## Usage Examples

### Example 1: MFA Failure Scenario

1. **User attempts to access email with MFA enabled**
2. **User provides invalid MFA code**
3. **System increments failed attempt count**
4. **After 3 failed attempts, email is locked out for 15 minutes**
5. **User receives generic "Access denied" message**

### Example 2: Geolocation Failure Scenario

1. **User attempts to access location-restricted email**
2. **User's location doesn't match restrictions**
3. **System increments failed attempt count**
4. **After 3 failed attempts, email is locked out for 15 minutes**
5. **User receives generic "Access denied" message**

### Example 3: Successful Access

1. **User provides valid credentials/access**
2. **System resets failed attempt count to 0**
3. **Email is accessible normally**
4. **Lockout is cleared if previously active**

## Monitoring and Logging

### Log Messages

The system logs various brute-force events:

```
Email test-email-123 is locked out due to brute-force protection
Failed to increment brute-force attempts for geolocation failure: database error
Failed to reset brute-force attempts on successful access: database error
```

### Database Monitoring

Monitor the following fields for security analysis:

```sql
-- Check emails with active lockouts
SELECT email_id, brute_force_failed_attempts, brute_force_lockout_until 
FROM emails 
WHERE brute_force_lockout_until IS NOT NULL 
  AND brute_force_lockout_until > datetime('now');

-- Check emails with high failed attempt counts
SELECT email_id, brute_force_failed_attempts, brute_force_last_failed_attempt 
FROM emails 
WHERE brute_force_failed_attempts > 0 
ORDER BY brute_force_failed_attempts DESC;
```

## Security Considerations

### 1. Information Leakage Prevention

- Generic error messages prevent attackers from determining lockout status
- No indication of remaining attempts or lockout duration
- Consistent response format for all failure types

### 2. Attack Prevention

- Server-side tracking prevents client-side manipulation
- Automatic expiration prevents permanent lockouts
- Per-email tracking prevents cross-email attacks

### 3. User Experience

- Legitimate users can still access emails after lockout expiration
- Successful access immediately resets failed attempts
- No impact on normal email functionality

### 4. Performance

- Database indexes optimize lockout queries
- Minimal overhead on successful access paths
- Efficient cleanup of expired lockouts

## Future Enhancements

### Potential Improvements

1. **Configurable Settings**: Environment variable configuration
2. **Advanced Lockout**: Progressive lockout durations
3. **IP-Based Tracking**: Track attempts by IP address
4. **Notification System**: Alert administrators of suspicious activity
5. **Analytics Dashboard**: Visual monitoring of brute-force attempts

### Configuration Options

1. **Per-Email Settings**: Allow senders to configure limits
2. **Global Settings**: System-wide brute-force protection
3. **Whitelist**: Exempt certain emails from protection
4. **Custom Messages**: Configurable error messages

## Troubleshooting

### Common Issues

1. **Lockout Not Working**
   - Check database migration was applied
   - Verify brute-force package is imported
   - Check logs for initialization errors

2. **Failed Attempts Not Incrementing**
   - Verify database connection
   - Check for SQL errors in logs
   - Ensure email ID exists in database

3. **Lockout Not Clearing**
   - Check system time synchronization
   - Verify lockout expiration logic
   - Check for database transaction issues

### Debug Information

Enable debug logging to see:
- Failed attempt increments
- Lockout application and clearing
- Database query results
- Error conditions

## Conclusion

The brute-force protection feature provides robust security against unauthorized access attempts while maintaining a good user experience. The implementation is secure, performant, and fully integrated with existing security features.

### Key Benefits

- ✅ Prevents brute-force attacks on email access
- ✅ Maintains security without revealing system details
- ✅ Integrates seamlessly with existing security layers
- ✅ Provides comprehensive monitoring and logging
- ✅ Automatically handles lockout expiration and reset

The feature successfully meets all acceptance criteria and provides a valuable addition to the Secure Email MVP's security capabilities.
