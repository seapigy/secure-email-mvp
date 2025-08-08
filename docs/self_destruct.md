# Self-Destruct After Failed Access Attempts

## Overview

The self-destruct feature automatically deletes emails after a specified number of failed access attempts. This provides an additional layer of security by preventing brute force attacks and unauthorized access attempts.

## Database Schema Changes

### New Columns Added

The following columns were added to the `emails` table:

```sql
-- Track if email has been self-destructed
ALTER TABLE emails ADD COLUMN self_destructed INTEGER DEFAULT 0;

-- Timestamp when email was deleted
ALTER TABLE emails ADD COLUMN deleted_at DATETIME;

-- Track failed access attempts
ALTER TABLE emails ADD COLUMN failed_access_attempts INTEGER DEFAULT 0;
```

### Indexes Added

```sql
-- Performance indexes for self-destruct functionality
CREATE INDEX IF NOT EXISTS idx_emails_self_destructed ON emails(self_destructed);
CREATE INDEX IF NOT EXISTS idx_emails_deleted_at ON emails(deleted_at);
CREATE INDEX IF NOT EXISTS idx_emails_failed_access_attempts ON emails(failed_access_attempts);
```

## API Endpoint Behavior

### GET /api/email/view/{id}

#### Status Codes

- **200 OK**: Email retrieved successfully
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Access denied (failed attempt)
- **404 Not Found**: Email not found
- **410 Gone**: Email has been self-destructed or expired
- **500 Internal Server Error**: Server error

#### Self-Destruct Logic

1. **Check if already self-destructed**: If `self_destructed = 1`, return 410 Gone
2. **Check expiration**: If `expires_at` is in the past, return 410 Gone
3. **Check failed attempts threshold**: If `failed_access_attempts >= max_attempts`:
   - Delete R2 blob
   - Mark email as self-destructed in database
   - Return 410 Gone
4. **Successful access**: Reset `failed_access_attempts = 0`

#### Failed Access Handling

When an access attempt fails (wrong password, unauthorized access, etc.):

1. Increment `failed_access_attempts` atomically
2. If `failed_access_attempts >= max_attempts`:
   - Delete the R2 blob permanently
   - Set `self_destructed = 1`
   - Set `deleted_at = CURRENT_TIMESTAMP`
   - Clear sensitive fields (`encrypted_blob_url`, `encrypted_key`, etc.)
3. Log the failed attempt and self-destruct action

## Environment Variables

### DEFAULT_MAX_FAILED_ATTEMPTS

- **Default**: 3
- **Description**: Default maximum failed attempts before self-destruct
- **Usage**: Can be overridden by per-email `max_attempts` field

### SIMULATE_SELF_DESTRUCT

- **Default**: 0 (disabled)
- **Values**: 0 (disabled) or 1 (enabled)
- **Description**: Enables test endpoint for simulating failed attempts
- **Security**: Only available when explicitly enabled

## Test Endpoint (Simulation Only)

### POST /test/self-destruct

**Available only when `SIMULATE_SELF_DESTRUCT=1`**

#### Request Body

```json
{
  "email_id": "uuid-of-email",
  "action": "increment_failed" | "reset"
}
```

#### Response

```json
{
  "email_id": "uuid-of-email",
  "failed_attempts": 2,
  "max_attempts": 3,
  "self_destructed": false
}
```

#### Actions

- **increment_failed**: Simulates a failed access attempt
- **reset**: Resets failed attempts counter to 0

## Configuration

### Per-Email Settings

When sending an email, you can configure self-destruct behavior:

```json
{
  "recipient": "user@example.com",
  "subject": "Secure Message",
  "body": "Secret content",
  "selfDestructAfterAttempts": true,
  "maxFailedAttempts": 3
}
```

### Default Behavior

- **selfDestructAfterAttempts**: false (disabled by default)
- **maxFailedAttempts**: 3 (uses `DEFAULT_MAX_FAILED_ATTEMPTS` if not specified)

## Logging

### Failed Attempt Logs

```
Failed attempt for email <email_id>: X/Y
```

### Self-Destruct Logs

```
Self-destruct triggered for email <email_id>, deleting blob <blobID>
Successfully deleted self-destructed email blob <blobID> from R2
Successfully marked email <email_id> as self-destructed in database
```

### Error Logs

```
Failed to delete self-destructed email from R2: <error>
Failed to mark email as self-destructed in database: <error>
```

## Testing

### Unit Tests

Run the unit tests:

```bash
go test ./cmd/api -v -run TestHandleFailedAccessAttempt
go test ./cmd/api -v -run TestViewEmailHandlerSelfDestruct
go test ./cmd/api -v -run TestTestSelfDestructHandler
```

### Integration Tests

Run the PowerShell integration test:

```powershell
# Enable simulation mode
$env:SIMULATE_SELF_DESTRUCT = "1"

# Run test script
.\scripts\test_self_destruct.ps1 -ApiHost "http://localhost:8080" -EnableSimulation

# Disable simulation
Remove-Item Env:SIMULATE_SELF_DESTRUCT
```

### Manual Testing

1. **Create email with self-destruct enabled**:
   ```bash
   curl -X POST http://localhost:8080/api/email/send \
     -H "Content-Type: application/json" \
     -d '{
       "recipient": "test@example.com",
       "subject": "Test",
       "body": "Secret",
       "selfDestructAfterAttempts": true,
       "maxFailedAttempts": 3
     }'
   ```

2. **Simulate failed attempts** (with simulation enabled):
   ```bash
   curl -X POST http://localhost:8080/test/self-destruct \
     -H "Content-Type: application/json" \
     -d '{
       "email_id": "email-uuid",
       "action": "increment_failed"
     }'
   ```

3. **Verify self-destruct**:
   ```bash
   curl -X GET http://localhost:8080/api/email/view/email-uuid
   # Should return 410 Gone
   ```

## Safety and Revert

### Production Safety

- **Default disabled**: Self-destruct is disabled by default (`selfDestructAfterAttempts: false`)
- **Conservative defaults**: Default max attempts is 3
- **Simulation disabled**: Test endpoints only available when explicitly enabled

### Reverting Test Changes

1. **Disable simulation**:
   ```bash
   unset SIMULATE_SELF_DESTRUCT
   ```

2. **Reset failed attempts** (if needed):
   ```bash
   curl -X POST http://localhost:8080/test/self-destruct \
     -H "Content-Type: application/json" \
     -d '{
       "email_id": "email-uuid",
       "action": "reset"
     }'
   ```

3. **Restart server** to ensure clean state

### Database Cleanup

If you need to reset the database state:

```sql
-- Reset failed attempts for all emails
UPDATE emails SET failed_access_attempts = 0;

-- Reset self-destructed emails (use with caution)
UPDATE emails SET self_destructed = 0, deleted_at = NULL;
```

## Security Considerations

### Atomic Operations

All database operations use transactions to ensure atomicity:

- Failed attempt increments are atomic
- Self-destruct operations are atomic
- No race conditions between concurrent access attempts

### R2 Storage Cleanup

- Blobs are permanently deleted from R2 storage
- No recovery possible after self-destruct
- Database metadata is cleared for security

### Audit Trail

- All failed attempts are logged
- Self-destruct actions are logged
- Database retains `deleted_at` timestamp for audit

## Error Handling

### Graceful Degradation

- If R2 deletion fails, database is still updated
- If database update fails, operation is rolled back
- Failed operations are logged but don't crash the system

### Concurrency Safety

- Uses database transactions for atomicity
- Handles concurrent access attempts safely
- Prevents double-deletion scenarios

## Performance Impact

### Minimal Overhead

- Failed attempt tracking adds minimal database overhead
- Self-destruct check is fast (indexed query)
- R2 deletion is asynchronous and doesn't block response

### Monitoring

Monitor these metrics:

- Failed access attempts per email
- Self-destruct events per time period
- R2 deletion success/failure rates
