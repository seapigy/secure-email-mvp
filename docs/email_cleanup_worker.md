# Email Cleanup Worker Documentation

## Overview

The Email Cleanup Worker is an automated background process that periodically deletes expired and consumed (burn-after-read) emails from the Secure Email MVP system. This ensures data security, reduces storage usage, and maintains compliance with email retention policies.

## Architecture

### Components

1. **EmailCleanupWorker** (`pkg/cleanup/worker.go`)
   - Main worker implementation
   - Handles database queries and R2 storage operations
   - Manages cleanup intervals and graceful shutdown

2. **Admin Handlers** (`cmd/api/admin_handlers.go`)
   - `/admin/email-retention-stats` - Get cleanup statistics
   - `/admin/manual-cleanup` - Trigger manual cleanup

3. **Standalone Worker** (`cmd/workers/email_cleanup_worker.go`)
   - Independent executable for running the worker
   - Can be deployed as a separate service

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `EMAIL_CLEANUP_INTERVAL_MINUTES` | 15 | Minutes between cleanup cycles |
| `SQLITE_DB` | `/var/db/secure-email.db` | Database path |
| `CLOUDFLARE_R2_ACCESS_KEY` | Required | R2 access key |
| `CLOUDFLARE_R2_SECRET_KEY` | Required | R2 secret key |
| `CLOUDFLARE_R2_BUCKET` | Required | R2 bucket name |
| `CLOUDFLARE_R2_ENDPOINT` | Required | R2 endpoint URL |

### Example Configuration

```bash
# .env file
EMAIL_CLEANUP_INTERVAL_MINUTES=5
SQLITE_DB=/var/db/secure-email.db
CLOUDFLARE_R2_ACCESS_KEY=your_access_key
CLOUDFLARE_R2_SECRET_KEY=your_secret_key
CLOUDFLARE_R2_BUCKET=your_bucket
CLOUDFLARE_R2_ENDPOINT=https://your_account_id.r2.cloudflarestorage.com
```

## Worker Logic

### Deletion Criteria

The worker deletes emails that meet **either** of the following conditions:

1. **Expired Emails**: `expires_at <= NOW()` AND `encrypted_blob_url IS NOT NULL`
2. **Burn-After-Read Emails**: `burn_after_read = 1 AND access_count > 0` AND `encrypted_blob_url IS NOT NULL`

### Cleanup Process

1. **Query for Deletable Emails**
   ```sql
   SELECT email_id, encrypted_blob_url, expires_at, burn_after_read, access_count
   FROM emails 
   WHERE (
       (expires_at IS NOT NULL AND expires_at <= datetime('now')) OR
       (burn_after_read = 1 AND access_count > 0)
   ) AND encrypted_blob_url IS NOT NULL
   ORDER BY created_at ASC
   ```

2. **For Each Email**:
   - Delete encrypted blob from R2 storage
   - Soft-delete from database (NULL out encryption fields)
   - Log success/failure for audit

3. **Soft Delete Strategy**
   - Preserves email metadata (sender, recipient, timestamps)
   - Removes encryption data (blob URL, keys, nonce, auth tag, hash)
   - Maintains audit trail while ensuring content is unrecoverable

### Safety Features

- **Minimum Interval**: 1 minute minimum between cleanups
- **Graceful Shutdown**: Handles SIGINT/SIGTERM signals
- **Error Handling**: Continues processing even if individual deletions fail
- **Dry Run Mode**: Available via admin API for testing
- **Comprehensive Logging**: All operations logged with timestamps

## Deployment Options

### Option 1: Integrated with API Server

The worker can be integrated into the main API server:

```go
// In main.go
worker, err := cleanup.NewEmailCleanupWorkerWithDB(db, 15)
if err != nil {
    log.Fatal("Failed to create cleanup worker:", err)
}
worker.Start()
defer worker.Stop()
```

### Option 2: Standalone Service

Run as a separate process:

```bash
# Build standalone worker
go build -o email-cleanup-worker cmd/workers/email_cleanup_worker.go

# Run with environment variables
EMAIL_CLEANUP_INTERVAL_MINUTES=5 ./email-cleanup-worker
```

### Option 3: Systemd Service

Create `/etc/systemd/system/email-cleanup-worker.service`:

```ini
[Unit]
Description=Secure Email Cleanup Worker
After=network.target

[Service]
Type=simple
User=secure-email
WorkingDirectory=/opt/secure-email-mvp
Environment=EMAIL_CLEANUP_INTERVAL_MINUTES=5
Environment=SQLITE_DB=/var/db/secure-email.db
ExecStart=/opt/secure-email-mvp/email-cleanup-worker
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Manual Operations

### Running Manual Cleanup

Via API:
```bash
# Dry run (no actual deletion)
curl -X POST http://localhost:8080/admin/manual-cleanup \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"dry_run": true}'

# Actual cleanup
curl -X POST http://localhost:8080/admin/manual-cleanup \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"dry_run": false}'
```

Via Go code:
```go
worker, err := cleanup.NewEmailCleanupWorkerWithDB(db, 15)
if err != nil {
    log.Fatal(err)
}
err = worker.RunCleanupOnce()
if err != nil {
    log.Printf("Cleanup failed: %v", err)
}
```

### Getting Statistics

```bash
curl -X GET http://localhost:8080/admin/email-retention-stats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Response:
```json
{
  "stats": {
    "expired_emails": 5,
    "burn_after_read_emails": 2,
    "total_emails_with_content": 150,
    "cleanup_interval_minutes": 15,
    "next_cleanup_run": "2024-01-15T10:30:00Z"
  },
  "summary": {
    "emails_pending_deletion": 7,
    "total_emails": 200,
    "soft_deleted_emails": 50
  }
}
```

## Monitoring and Logging

### Log Format

```
2024-01-15 10:15:00 Starting email cleanup process...
2024-01-15 10:15:00 Found 3 emails to delete
2024-01-15 10:15:01 Successfully deleted email blob abc123 from R2
2024-01-15 10:15:01 Successfully deleted email expired-email-1 (reason: expired)
2024-01-15 10:15:02 Successfully deleted email burn-email-1 (reason: burn-after-read)
2024-01-15 10:15:02 Cleanup completed in 2.1s: 3 successful, 0 failed
```

### Key Metrics to Monitor

- **Cleanup Frequency**: Should match configured interval
- **Success Rate**: Should be close to 100%
- **Processing Time**: Should be under 30 seconds for normal loads
- **Emails Pending Deletion**: Should not grow indefinitely
- **R2 Storage Usage**: Should decrease over time

### Alert Conditions

- Cleanup failures > 10% of attempts
- Processing time > 60 seconds
- R2 deletion failures
- Database connection errors
- Worker process not running

## Security Considerations

### Data Protection

- **Soft Delete**: Preserves audit trail while removing content
- **R2 Deletion**: Permanently removes encrypted blobs
- **No Recovery**: Deleted emails cannot be recovered
- **Access Control**: Admin endpoints require JWT authentication

### Operational Safety

- **Dry Run Mode**: Test cleanup without actual deletion
- **Minimum Intervals**: Prevents excessive resource usage
- **Error Handling**: Continues processing despite individual failures
- **Graceful Shutdown**: Prevents data corruption

### Compliance

- **Audit Trail**: All deletions logged with timestamps
- **Metadata Preservation**: Sender/recipient info retained for compliance
- **Configurable Retention**: Adjustable cleanup intervals
- **Manual Override**: Admin can trigger immediate cleanup

## Troubleshooting

### Common Issues

1. **Worker Not Starting**
   - Check environment variables
   - Verify database connectivity
   - Ensure R2 credentials are valid

2. **Cleanup Not Running**
   - Check logs for errors
   - Verify interval configuration
   - Ensure worker process is running

3. **R2 Deletion Failures**
   - Check R2 credentials
   - Verify bucket permissions
   - Check network connectivity

4. **Database Errors**
   - Verify database schema
   - Check disk space
   - Ensure proper permissions

### Debug Commands

```bash
# Check worker status
ps aux | grep email-cleanup-worker

# View recent logs
tail -f /var/log/secure-email/cleanup.log

# Test database connection
sqlite3 /var/db/secure-email.db "SELECT COUNT(*) FROM emails;"

# Test R2 connectivity
curl -H "Authorization: AWS4-HMAC-SHA256 ..." \
  https://your-bucket.your-account.r2.cloudflarestorage.com/
```

## Performance Considerations

### Optimization Tips

- **Batch Processing**: Process emails in batches for large datasets
- **Index Optimization**: Ensure proper database indexes
- **Connection Pooling**: Reuse database connections
- **Parallel Processing**: Consider parallel R2 deletions (with caution)

### Resource Usage

- **Memory**: ~10-50MB depending on batch size
- **CPU**: Minimal, mostly I/O bound
- **Network**: R2 API calls for blob deletion
- **Database**: Read queries for email discovery

### Scaling Considerations

- **Multiple Workers**: Can run multiple workers with different intervals
- **Sharding**: Partition emails by date or user for parallel processing
- **Load Balancing**: Distribute cleanup across multiple instances

## Future Enhancements

### Planned Features

1. **Metrics Dashboard**: Real-time cleanup statistics
2. **Advanced Filtering**: Custom deletion criteria
3. **Bulk Operations**: Batch email management
4. **Retention Policies**: Configurable retention rules
5. **Backup Integration**: Pre-deletion backup options

### API Extensions

- `/admin/cleanup-config` - Configure cleanup parameters
- `/admin/cleanup-history` - View cleanup history
- `/admin/email-bulk-delete` - Bulk email deletion
- `/admin/retention-policies` - Manage retention rules














