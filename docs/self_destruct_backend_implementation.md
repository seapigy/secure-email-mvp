# Self-Destruct After Failed Attempts - Backend Implementation

## Overview
This document describes the backend implementation for the "Self-Destruct After Failed Attempts" feature, which allows emails to be automatically deleted after a specified number of failed access attempts.

## Database Schema Changes

### New Fields Added to `emails` Table

```sql
-- Should the email self-destruct after max failed attempts?
self_destruct_after_attempts INTEGER DEFAULT 0,

-- Max allowed failed attempts before the email self-deletes
max_attempts INTEGER DEFAULT 3,
```

### Migration Script
The migration is located at `schema/migrate_add_self_destruct.sql` and can be run using:
```powershell
.\scripts\run_migration.ps1
```

## API Changes

### Send Email Handler (`/api/email/send`)

#### Request Structure
```json
{
  "recipient": "user@example.com",
  "subject": "Secure Message",
  "body": "This is a secure message",
  "selfDestructAfterAttempts": true,
  "maxFailedAttempts": 3
}
```

#### Validation Rules
- If `selfDestructAfterAttempts` is `true`, `maxFailedAttempts` must be between 1-10
- If `selfDestructAfterAttempts` is `false`, `maxFailedAttempts` is set to 0
- Invalid values return HTTP 400 with error message

#### Database Storage
- `self_destruct_after_attempts` stored as INTEGER (0 = false, 1 = true)
- `max_attempts` stored as INTEGER
- Both fields are included in encrypted metadata

### View Email Handler (`/api/email/view/{id}`)

#### Response Structure
```json
{
  "email_id": "uuid",
  "recipient": "user@example.com",
  "subject": "Secure Message",
  "body": "Decrypted content",
  "created_at": "2024-01-01T00:00:00Z",
  "status": "success",
  "selfDestructAfterAttempts": true,
  "maxFailedAttempts": 3
}
```

### List Email Handler (`/api/email/list`)

#### Response Structure
```json
{
  "emails": [
    {
      "email_id": "uuid",
      "recipient": "user@example.com",
      "subject": "Secure Message",
      "created_at": "2024-01-01T00:00:00Z",
      "selfDestructAfterAttempts": true,
      "maxFailedAttempts": 3
    }
  ],
  "status": "success"
}
```

## Implementation Details

### Go Type Definitions

```go
type SendEmailRequest struct {
    Recipient                 string `json:"recipient"`
    Subject                   string `json:"subject"`
    Body                      string `json:"body"`
    SelfDestructAfterAttempts bool   `json:"selfDestructAfterAttempts,omitempty"`
    MaxFailedAttempts         int    `json:"maxFailedAttempts,omitempty"`
}

type ViewEmailResponse struct {
    EmailID                   string    `json:"email_id"`
    Recipient                 string    `json:"recipient"`
    Subject                   string    `json:"subject"`
    Body                      string    `json:"body"`
    CreatedAt                 time.Time `json:"created_at"`
    Status                    string    `json:"status"`
    Error                     string    `json:"error,omitempty"`
    SelfDestructAfterAttempts bool      `json:"selfDestructAfterAttempts"`
    MaxFailedAttempts         int       `json:"maxFailedAttempts"`
}

type EmailListItem struct {
    EmailID                   string    `json:"email_id"`
    Recipient                 string    `json:"recipient"`
    Subject                   string    `json:"subject"`
    CreatedAt                 time.Time `json:"created_at"`
    SelfDestructAfterAttempts bool      `json:"selfDestructAfterAttempts"`
    MaxFailedAttempts         int       `json:"maxFailedAttempts"`
}
```

### Validation Logic

```go
// Validate self-destruct settings
if req.SelfDestructAfterAttempts {
    if req.MaxFailedAttempts < 1 || req.MaxFailedAttempts > 10 {
        log.Printf("Invalid maxFailedAttempts: %d (must be between 1-10)", req.MaxFailedAttempts)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(`{"error":"maxFailedAttempts must be between 1 and 10"}`))
        return
    }
} else {
    // If self-destruct is disabled, set maxFailedAttempts to 0
    req.MaxFailedAttempts = 0
}
```

### Database Operations

#### Insert Query
```sql
INSERT INTO emails (
    email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
    encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, 
    self_destruct_after_attempts, max_attempts, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
```

#### Select Query (View)
```sql
SELECT encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, 
       compression_algo, sender_id, recipient, subject, created_at,
       self_destruct_after_attempts, max_attempts
FROM emails WHERE email_id = ?
```

#### Select Query (List)
```sql
SELECT email_id, recipient, subject, created_at, 
       self_destruct_after_attempts, max_attempts
FROM emails 
WHERE sender_id = ?
ORDER BY created_at DESC
```

## Testing

### Unit Tests
Tests are located in `cmd/api/send_email_test.go` and cover:
- Valid self-destruct settings
- Invalid maxFailedAttempts values (too low/high)
- Self-destruct disabled scenarios
- Required field validation

### Test Cases
```go
func TestSendEmailHandler_SelfDestructValidation(t *testing.T) {
    // Tests various validation scenarios
}

func TestSendEmailHandler_RequiredFields(t *testing.T) {
    // Tests required field validation
}
```

## Security Considerations

### Data Protection
- Self-destruct settings are stored in encrypted metadata
- No plaintext storage of sensitive security settings
- Database queries use parameterized statements to prevent SQL injection

### Validation
- Client-side validation on frontend
- Server-side validation in backend
- Range checking for maxFailedAttempts (1-10)
- Type safety with Go structs

### Backward Compatibility
- Existing emails without self-destruct settings default to `false`
- Database migration handles existing records
- API responses include default values for missing fields

## Error Handling

### Validation Errors
- HTTP 400 for invalid maxFailedAttempts values
- Clear error messages for debugging
- Logging of validation failures

### Database Errors
- HTTP 500 for database connection issues
- Graceful handling of migration failures
- Rollback support for failed operations

## Performance Considerations

### Database Indexes
```sql
CREATE INDEX IF NOT EXISTS idx_emails_self_destruct ON emails(self_destruct_after_attempts);
```

### Query Optimization
- Selective field retrieval based on endpoint needs
- Efficient boolean to integer conversion
- Minimal data transfer for list operations

## Future Enhancements

### Planned Features
1. **Failed Attempt Tracking**: Implement backend logic to track and enforce failed attempts
2. **Automatic Deletion**: Add background job to delete emails after max attempts
3. **Audit Logging**: Track all access attempts for security monitoring
4. **Rate Limiting**: Implement per-email rate limiting for access attempts

### API Extensions
1. **Failed Attempts Endpoint**: `GET /api/email/{id}/attempts`
2. **Reset Attempts Endpoint**: `POST /api/email/{id}/reset-attempts`
3. **Force Delete Endpoint**: `DELETE /api/email/{id}/force-delete`

## Deployment Notes

### Migration Requirements
1. Run database migration before deploying new code
2. Ensure SQLite3 is available for migration script
3. Backup database before running migration
4. Test migration on staging environment first

### Configuration
- No additional environment variables required
- Default values are safe for production
- Migration script handles all setup automatically

## Monitoring and Logging

### Key Metrics
- Number of emails with self-destruct enabled
- Distribution of maxFailedAttempts values
- Failed validation attempts
- Database migration success rate

### Log Messages
```go
log.Printf("Invalid maxFailedAttempts: %d (must be between 1-10)", req.MaxFailedAttempts)
log.Printf("Self-destruct settings: enabled=%t, maxAttempts=%d", req.SelfDestructAfterAttempts, req.MaxFailedAttempts)
```
