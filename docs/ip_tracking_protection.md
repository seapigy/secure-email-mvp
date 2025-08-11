# IP-Based Tracking & Lockout for Email Access Attempts

## Overview

**Micro-Iteration 4.13** implements IP-based tracking and lockout functionality for secure email access attempts. This feature tracks failed access attempts by IP address and implements timed lockouts, providing an additional layer of security that works alongside the existing per-email brute-force protection.

## Key Features

- **IP-Based Tracking**: Failed attempts are tracked by client IP address
- **Configurable Limits**: Default 5 failed attempts within 15 minutes before lockout
- **Timed Lockouts**: Default 30-minute lockout duration
- **Generic Error Messages**: Security-focused responses that don't reveal lockout status
- **Automatic Reset**: Failed attempts reset on successful access
- **Automatic Cleanup**: Old records automatically removed after 24 hours
- **Multi-Layer Integration**: Works with MFA, geolocation, and per-email brute-force protection

## Database Schema

### New Table: `ip_access_attempts`

```sql
CREATE TABLE ip_access_attempts (
    ip_address TEXT PRIMARY KEY,
    failed_attempts INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    lockout_until TIMESTAMP NULL
);

-- Indexes for performance
CREATE INDEX idx_ip_access_attempts_last_attempt ON ip_access_attempts(last_attempt_at);
CREATE INDEX idx_ip_access_attempts_lockout ON ip_access_attempts(lockout_until);
```

### Field Details

- **`ip_address`**: Client IP address (primary key)
- **`failed_attempts`**: Count of consecutive failed attempts from this IP
- **`last_attempt_at`**: Timestamp of the last attempt (for cleanup and window tracking)
- **`lockout_until`**: Timestamp until which this IP is locked out (NULL if not locked)

## Backend Implementation

### IP Tracking Service

The `pkg/iptracking` package provides the core functionality:

```go
// Initialize IP tracking service
ipTracking := iptracking.NewIPTrackingService(db)

// Check if IP is locked out
locked, err := ipTracking.CheckIPLockout(clientIP)
if locked {
    // Return generic "Access denied" response
}

// Increment failed attempts
err = ipTracking.IncrementFailedAttempt(clientIP)

// Reset failed attempts on successful access
err = ipTracking.ResetFailedAttempts(clientIP)
```

### Configuration

Default configuration settings:

```go
type IPTrackingConfig struct {
    MaxAttempts           5                    // Maximum failed attempts before lockout
    LockoutDuration       30 * time.Minute     // Lockout duration
    CleanupOlderThan      24 * time.Hour       // Cleanup records older than this
    AttemptWindowDuration 15 * time.Minute     // Time window for counting attempts
}
```

### Integration Points

#### 1. Email Access Flow

The IP tracking is integrated into the email access flow in `view_email_handler.go`:

```go
// 1. Authentication Check
// 2. IP-Based Lockout Check (Micro-Iteration 4.13)
// 3. Geolocation Check (if restrictions set)
// 4. Per-Email Brute-Force Lockout Check (Micro-Iteration 4.12)
// 5. MFA Check (if enabled)
// 6. Password Check (if required)
// 7. Email Decryption
```

#### 2. Security Failure Handling

IP tracking is triggered by all types of security failures:

- **MFA Failures**: Invalid TOTP or email codes
- **Geolocation Failures**: Location restrictions not met
- **Password Failures**: Invalid passwords (when implemented)
- **Authentication Failures**: Invalid tokens or unauthorized access

#### 3. Success Handling

Failed attempts are automatically reset on successful access:

```go
// Reset IP tracking failed attempts on successful access
if err := ipTracking.ResetFailedAttempts(clientIP); err != nil {
    log.Printf("Failed to reset IP attempts on successful access: %v", err)
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
    err = ipTracking.clearIPLockout(ipAddress)
}
```

### 3. Server-Side Only

All lockout and attempt tracking is stored server-side only, preventing client-side manipulation.

### 4. Comprehensive Logging

All IP tracking events are logged for monitoring and analysis:

```go
log.Printf("IP %s is locked out due to repeated failed attempts", clientIP)
log.Printf("Failed to increment IP attempts for MFA failure: %v", err)
```

### 5. Automatic Cleanup

Old IP records are automatically cleaned up to prevent unbounded growth:

```go
// Cleanup old records on server startup
if err := ipTrackingService.CleanupOldRecords(); err != nil {
    log.Printf("Warning: Failed to cleanup old IP records: %v", err)
}
```

## Configuration

### Default Settings

- **Max Attempts**: 5 failed attempts within 15 minutes
- **Lockout Duration**: 30 minutes
- **Attempt Window**: 15 minutes (attempts outside this window reset the counter)
- **Cleanup Duration**: 24 hours (records older than this are automatically removed)
- **Reset on Success**: Automatic

### Environment Variables

Currently using database defaults, but can be extended to support environment variable configuration:

```go
// Future enhancement: Configurable via environment variables
maxAttempts := os.Getenv("IP_TRACKING_MAX_ATTEMPTS")
lockoutDuration := os.Getenv("IP_TRACKING_LOCKOUT_DURATION")
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

### IP Lockout Response

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
go test ./pkg/iptracking -v
```

Tests cover:
- Lockout checking (no lockout, active lockout, expired lockout)
- Failed attempt incrementing and resetting
- Status retrieval and error handling
- Cleanup functionality
- Configuration validation

### Integration Tests

```bash
./scripts/test_ip_tracking.ps1
```

Tests cover:
- MFA failure scenarios
- Geolocation failure scenarios
- IP lockout enforcement
- Success scenarios
- Generic error messages

## Usage Examples

### Example 1: MFA Failure Scenario

1. **User attempts to access email with MFA enabled**
2. **User provides invalid MFA code**
3. **System increments IP failed attempt count**
4. **After 5 failed attempts within 15 minutes, IP is locked out for 30 minutes**
5. **User receives generic "Access denied" message**

### Example 2: Geolocation Failure Scenario

1. **User attempts to access location-restricted email**
2. **User's location doesn't match restrictions**
3. **System increments IP failed attempt count**
4. **After 5 failed attempts within 15 minutes, IP is locked out for 30 minutes**
5. **User receives generic "Access denied" message**

### Example 3: Successful Access

1. **User provides valid credentials/access**
2. **System resets IP failed attempt count to 0**
3. **Email is accessible normally**
4. **IP lockout is cleared if previously active**

### Example 4: Attempt Window Reset

1. **User has 3 failed attempts**
2. **User waits 20 minutes (outside 15-minute window)**
3. **User attempts access again**
4. **System resets attempt count to 1 (new attempt)**
5. **User can continue attempting access**

## Monitoring and Logging

### Log Messages

The system logs various IP tracking events:

```
IP 192.168.1.100 is locked out due to repeated failed attempts
Failed to increment IP attempts for MFA failure: database error
Failed to reset IP attempts on successful access: database error
Cleaned up 5 old IP access records
```

### Database Monitoring

Monitor the following fields for security analysis:

```sql
-- Check IPs with active lockouts
SELECT ip_address, failed_attempts, lockout_until 
FROM ip_access_attempts 
WHERE lockout_until IS NOT NULL 
  AND lockout_until > datetime('now');

-- Check IPs with high failed attempt counts
SELECT ip_address, failed_attempts, last_attempt_at 
FROM ip_access_attempts 
WHERE failed_attempts > 0 
ORDER BY failed_attempts DESC;

-- Check for suspicious activity (multiple IPs with high attempts)
SELECT COUNT(*) as suspicious_ips
FROM ip_access_attempts 
WHERE failed_attempts >= 3;
```

## Security Considerations

### 1. Information Leakage Prevention

- Generic error messages prevent attackers from determining lockout status
- No indication of remaining attempts or lockout duration
- Consistent response format for all failure types

### 2. Attack Prevention

- Server-side tracking prevents client-side manipulation
- Automatic expiration prevents permanent lockouts
- Per-IP tracking prevents cross-IP attacks
- Attempt window prevents long-term accumulation

### 3. User Experience

- Legitimate users can still access emails after lockout expiration
- Successful access immediately resets failed attempts
- Attempt window allows recovery after temporary issues
- No impact on normal email functionality

### 4. Performance

- Database indexes optimize lockout queries
- Minimal overhead on successful access paths
- Efficient cleanup of expired lockouts and old records
- Automatic cleanup prevents database bloat

### 5. Integration with Existing Security

- Works seamlessly with per-email brute-force protection
- Both systems can trigger independently
- No conflicts between IP-based and email-based lockouts
- Comprehensive security coverage

## Future Enhancements

### Potential Improvements

1. **Environment Variable Configuration**: Configurable via environment variables
2. **Advanced Lockout**: Progressive lockout durations
3. **IP Whitelist**: Exempt certain IPs from tracking
4. **Notification System**: Alert administrators of suspicious activity
5. **Analytics Dashboard**: Visual monitoring of IP tracking attempts

### Configuration Options

1. **Per-System Settings**: System-wide IP tracking configuration
2. **Geographic Restrictions**: Different limits for different regions
3. **Time-Based Rules**: Different limits for different times of day
4. **Custom Messages**: Configurable error messages

## Troubleshooting

### Common Issues

1. **IP Lockout Not Working**
   - Check database migration was applied
   - Verify IP tracking package is imported
   - Check logs for initialization errors

2. **Failed Attempts Not Incrementing**
   - Verify database connection
   - Check for SQL errors in logs
   - Ensure IP address is being captured correctly

3. **Lockout Not Clearing**
   - Check system time synchronization
   - Verify lockout expiration logic
   - Check for database transaction issues

4. **Cleanup Not Working**
   - Verify cleanup is called on server startup
   - Check for database permission issues
   - Monitor cleanup logs

### Debug Information

Enable debug logging to see:
- Failed attempt increments
- Lockout application and clearing
- Database query results
- Error conditions
- Cleanup operations

## Conclusion

The IP-based tracking and lockout feature provides robust security against unauthorized access attempts while maintaining a good user experience. The implementation is secure, performant, and fully integrated with existing security features.

### Key Benefits

- ✅ Prevents brute-force attacks from specific IP addresses
- ✅ Maintains security without revealing system details
- ✅ Integrates seamlessly with existing security layers
- ✅ Provides comprehensive monitoring and logging
- ✅ Automatically handles lockout expiration and reset
- ✅ Prevents database bloat through automatic cleanup
- ✅ Works alongside per-email brute-force protection

The feature successfully meets all acceptance criteria and provides a valuable addition to the Secure Email MVP's security capabilities.
