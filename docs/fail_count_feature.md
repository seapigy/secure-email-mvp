# Fail Count Feature - Auto-Delete After Failed Access Attempts

## Overview

The fail count feature automatically deletes emails after a specified number of failed access attempts. This provides an additional layer of security by preventing brute force attacks and unauthorized access attempts.

## Database Schema Changes

### New Column Added

The following column was added to the `emails` table:

```sql
-- Track failed access attempts
ALTER TABLE emails ADD COLUMN fail_count INTEGER DEFAULT 0;
```

### Index Added

```sql
-- Performance index for fail count functionality
CREATE INDEX IF NOT EXISTS idx_emails_fail_count ON emails(fail_count);
```

## API Endpoint Behavior

### POST /api/email/get

#### Status Codes

- **200 OK**: Email retrieved successfully
- **400 Bad Request**: Invalid request format
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Access denied (failed attempt)
- **404 Not Found**: Email not found
- **410 Gone**: Email deleted due to too many failed attempts
- **500 Internal Server Error**: Server error

#### Fail Count Logic

1. **Check fail count**: If `fail_count >= FAIL_ATTEMPT_LIMIT` (3), return 410 Gone
2. **Failed access**: Increment `fail_count` and check if limit reached
3. **Successful access**: Reset `fail_count = 0`

#### Failed Access Handling

When an access attempt fails (wrong sender, decryption failure, etc.):

1. Increment `fail_count` atomically
2. If `fail_count >= FAIL_ATTEMPT_LIMIT`:
   - Delete the R2 blob permanently
   - Delete the email metadata row from SQLite
   - Return 410 Gone with error message
3. Log the failed attempt with timestamp, client IP, and reason

## Configuration

### FAIL_ATTEMPT_LIMIT Constant

- **Default**: 3
- **Location**: `cmd/api/get_email_handler.go`
- **Description**: Maximum failed attempts before email deletion
- **Usage**: Can be easily adjusted by changing the constant value

```go
// FAIL_ATTEMPT_LIMIT is the maximum number of failed attempts before email deletion
const FAIL_ATTEMPT_LIMIT = 3
```

## Logging

### Failed Attempt Logs

```
Failed access attempt for email <email_id>: <reason> (IP: <client_ip>)
```

### Deletion Logs

```
Fail limit reached for email <email_id>: <count>/<limit> attempts
Successfully deleted email blob <blob_id> from R2
Deleted email <email_id> due to too many failed attempts
```

### Error Logs

```
Failed to delete email from R2: <error>
Failed to delete email metadata: <error>
```

## Testing

### Unit Tests

Run the unit tests:

```bash
go test ./cmd/api -v -run TestHandleFailedAccess
go test ./cmd/api -v -run TestGetEmailHandlerFailCount
```

### Integration Tests

Run the PowerShell integration test:

```powershell
.\scripts\test_fail_count.ps1 -ApiHost "http://localhost:8080"
```

### Manual Testing

1. **Create email**:
   ```bash
   curl -X POST http://localhost:8080/api/email/send \
     -H "Content-Type: application/json" \
     -d '{
       "recipient": "test@example.com",
       "subject": "Test",
       "body": "Secret"
     }'
   ```

2. **Simulate failed attempts**:
   ```bash
   curl -X POST http://localhost:8080/api/email/get \
     -H "Content-Type: application/json" \
     -d '{
       "email_id": "email-uuid"
     }'
   # Should return 403 Forbidden for first 2 attempts, then 410 Gone
   ```

3. **Verify deletion**:
   ```bash
   curl -X POST http://localhost:8080/api/email/get \
     -H "Content-Type: application/json" \
     -d '{
       "email_id": "email-uuid"
     }'
   # Should return 404 Not Found after deletion
   ```

## Security Considerations

### Atomic Operations

All database operations use transactions to ensure atomicity:

- Fail count increments are atomic
- Deletion operations are atomic
- No race conditions between concurrent access attempts

### R2 Storage Cleanup

- Blobs are permanently deleted from R2 storage
- No recovery possible after deletion
- Database metadata is completely removed

### Audit Trail

- All failed attempts are logged with timestamp and IP
- Deletion actions are logged
- No persistent record of deleted emails (for security)

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

- Fail count tracking adds minimal database overhead
- Deletion check is fast (indexed query)
- R2 deletion is asynchronous and doesn't block response

### Monitoring

Monitor these metrics:

- Failed access attempts per email
- Deletion events per time period
- R2 deletion success/failure rates

## Implementation Details

### Key Functions

- `handleFailedAccess()`: Increments fail count and handles deletion
- `EmailDeletedError`: Special error type for deletion events
- `FAIL_ATTEMPT_LIMIT`: Configurable constant for limit

### Database Operations

- Uses SQL transactions for atomicity
- Deletes entire email row (not just marking)
- Resets fail count on successful access

### R2 Integration

- Uses existing `storage.DeleteBlob()` function
- Handles R2 configuration errors gracefully
- Continues with database deletion even if R2 fails
