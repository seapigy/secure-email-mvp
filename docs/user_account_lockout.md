# User Account Lockout (Micro-Iteration 4.6)

## Overview

**Micro-Iteration 4.6** implements temporary account lockout after a configurable number of failed login attempts to mitigate brute-force attacks. This feature enhances security by preventing unauthorized access through repeated password guessing attempts.

## Key Features

- **Configurable Thresholds**: Set maximum failed attempts, lockout duration, and attempt window
- **Automatic Lockout**: Accounts are automatically locked after exceeding the failed attempt threshold
- **Time-Based Reset**: Failed attempts reset after the configured time window expires
- **Automatic Unlock**: Lockouts automatically expire after the configured duration
- **Manual Unlock**: Admin or user can manually unlock accounts via API
- **Comprehensive Logging**: All lockout events are logged for audit purposes
- **Status Monitoring**: Real-time status checking and statistics

## Configuration

### Environment Variables

The lockout system is configured through environment variables:

```bash
# Enable/disable account lockout (1 = enabled, 0 = disabled)
LOGIN_RATE_LIMIT_ENABLED=1

# Maximum failed login attempts before lockout (default: 5)
LOGIN_MAX_ATTEMPTS=5

# Lockout duration in minutes (default: 30)
LOGIN_LOCKOUT_MINUTES=30

# Time window in minutes for counting failed attempts (default: 15)
LOGIN_ATTEMPT_WINDOW_MINUTES=15
```

### Default Configuration

If no environment variables are set, the system uses these defaults:

- **Max Attempts**: 5 failed attempts
- **Lockout Duration**: 30 minutes
- **Attempt Window**: 15 minutes
- **Enabled**: true

## Database Schema

### Users Table Extensions

The lockout system adds the following fields to the `users` table:

```sql
-- Migration: migrate_add_login_lockout.sql
ALTER TABLE users ADD COLUMN failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN last_failed_login TIMESTAMP;
ALTER TABLE users ADD COLUMN account_locked_until TIMESTAMP;
```

### Field Descriptions

- **`failed_login_attempts`**: Count of consecutive failed login attempts within the attempt window
- **`last_failed_login`**: Timestamp of the most recent failed login attempt
- **`account_locked_until`**: Timestamp until which the account is locked (NULL if not locked)

## Backend Implementation

### Lockout Service

The core functionality is implemented in `pkg/lockout/lockout.go`:

```go
type UserLockoutService struct {
    db     *sql.DB
    config *LockoutConfig
}
```

### Key Methods

#### CheckUserLockout(email string) (bool, *time.Time, error)
Checks if a user account is currently locked out.

#### IncrementUserFailedAttempt(email string) error
Increments the failed attempt count and applies lockout if threshold is reached.

#### ResetUserFailedAttempts(email string) error
Resets failed attempts after successful login.

#### GetUserLockoutStatus(email string) (*UserLockoutStatus, error)
Returns detailed lockout status information.

### Integration with Login Flow

The lockout system is integrated into the login handler (`cmd/api/login_handler.go`):

1. **Pre-Authentication Check**: Verify account is not locked before processing login
2. **Failed Authentication**: Increment failed attempts and check for lockout trigger
3. **Successful Authentication**: Reset failed attempts

## API Endpoints

### 1. Lockout Status

**Endpoint**: `GET /api/auth/lockout/status?email={email}`

**Description**: Get the current lockout status for a user account.

**Response**:
```json
{
  "email": "user@example.com",
  "is_locked_out": true,
  "failed_attempts": 5,
  "max_attempts": 5,
  "remaining_attempts": 0,
  "last_failed_login": "2024-01-15T10:30:00Z",
  "lockout_until": "2024-01-15T11:00:00Z",
  "lockout_remaining": "25m30s",
  "attempt_window": "15m0s",
  "is_within_window": true
}
```

### 2. Account Unlock

**Endpoint**: `POST /api/auth/lockout/unlock`

**Description**: Manually unlock a user account.

**Request**:
```json
{
  "email": "user@example.com"
}
```

**Response**:
```json
{
  "email": "user@example.com",
  "unlocked": true,
  "message": "Account successfully unlocked",
  "timestamp": "2024-01-15T10:35:00Z"
}
```

### 3. Lockout Configuration

**Endpoint**: `GET /api/auth/lockout/config`

**Description**: Get the current lockout configuration.

**Response**:
```json
{
  "max_attempts": 5,
  "lockout_duration": "30m0s",
  "attempt_window": "15m0s",
  "enabled": true,
  "lockout_duration_minutes": 30,
  "attempt_window_minutes": 15
}
```

### 4. Lockout Statistics

**Endpoint**: `GET /api/auth/lockout/stats`

**Description**: Get system-wide lockout statistics (admin only).

**Response**:
```json
{
  "currently_locked_accounts": 3,
  "accounts_with_failed_attempts": 12,
  "recent_lockout_events_24h": 8,
  "timestamp": "2024-01-15T10:35:00Z"
}
```

## Error Responses

### Account Locked (429 Too Many Requests)

When an account is locked, login attempts return:

```json
{
  "error": "Account temporarily locked due to repeated failed login attempts. Please try again later.",
  "code": "account_locked"
}
```

### Invalid Credentials (401 Unauthorized)

Standard invalid credentials response (before lockout):

```json
{
  "error": "Invalid credentials"
}
```

## Security Considerations

### 1. Brute Force Protection

- Prevents automated password guessing attacks
- Configurable thresholds allow for security vs. usability balance
- Time-based windows prevent permanent lockouts

### 2. Information Disclosure

- Generic error messages don't reveal lockout status
- No indication of whether an account exists
- Lockout status only available through dedicated endpoints

### 3. Audit Logging

- All lockout events are logged with timestamps
- Failed attempt increments are tracked
- Account unlock operations are recorded

### 4. Graceful Degradation

- System continues to function if lockout service fails
- Database errors don't prevent login processing
- Fallback behavior allows access on service failure

## Testing

### Unit Tests

Comprehensive unit tests are available in `pkg/lockout/lockout_test.go`:

```bash
go test ./pkg/lockout -v
```

### Integration Tests

Run the integration test script:

```bash
# Windows PowerShell
.\scripts\test_user_account_lockout.ps1

# With custom parameters
.\scripts\test_user_account_lockout.ps1 -ApiBase "http://localhost:8080" -TestEmail "test@example.com"
```

### Test Scenarios

1. **Successful Login**: Verify failed attempts reset
2. **Failed Login Attempts**: Verify counter increments
3. **Account Lockout**: Verify lockout after threshold
4. **Lockout Status**: Verify status endpoint functionality
5. **Account Unlock**: Verify manual unlock capability
6. **Login After Unlock**: Verify successful login after unlock
7. **Attempt Window**: Verify attempts reset after window expires
8. **Configuration**: Verify environment variable loading
9. **Statistics**: Verify system-wide statistics
10. **Error Handling**: Verify graceful error handling

## Monitoring and Logging

### Log Messages

The system logs various events:

```
User account locked: user@example.com (failed attempts: 5)
User failed attempts reset: user@example.com
User lockout cleared: user@example.com
Failed to check user lockout for user@example.com: database error
```

### Metrics

Track these key metrics:

- **Lockout Events**: Number of accounts locked per time period
- **Failed Attempts**: Distribution of failed login attempts
- **Unlock Operations**: Manual unlock frequency
- **Service Errors**: Lockout service failure rates

## Troubleshooting

### Common Issues

1. **Accounts Not Locking**
   - Check `LOGIN_RATE_LIMIT_ENABLED` environment variable
   - Verify database migration was applied
   - Check application logs for errors

2. **Lockouts Not Expiring**
   - Verify system clock accuracy
   - Check for database connection issues
   - Review lockout duration configuration

3. **Failed Attempts Not Resetting**
   - Check attempt window configuration
   - Verify successful login flow
   - Review database transaction handling

### Debug Mode

Enable debug logging by setting:

```bash
DEBUG=true
```

This provides detailed information about lockout operations.

## Future Enhancements

### Planned Features

1. **IP-Based Lockout**: Lock accounts based on IP address patterns
2. **Progressive Delays**: Increase lockout duration with repeated violations
3. **Notification System**: Alert users when accounts are locked
4. **Admin Dashboard**: Web interface for lockout management
5. **Analytics**: Advanced reporting and trend analysis

### Configuration Enhancements

1. **Per-User Settings**: Allow users to configure their own lockout preferences
2. **Risk-Based Adjustments**: Dynamic thresholds based on user behavior
3. **Whitelist Support**: Exclude certain accounts from lockout
4. **Integration APIs**: Webhook support for external systems

## References

- [Micro-Iteration 4.6 Specification](../docs/micro-iteration-4.6-summary.md)
- [Security Best Practices](../SECURITY.md)
- [API Documentation](../docs/api/)
- [Testing Guide](../scripts/test_user_account_lockout.ps1)

