# Read Receipts & Expiration Alerts

## Overview

Micro-Iteration 4.19 implements real-time read receipts and expiration alerts for secure emails. This feature allows senders to know exactly when their emails have been read for the first time and receive notifications before emails expire or are auto-deleted.

## Features

### Read Receipts
- **First Read Detection**: Automatically detects when an email is read for the first time
- **Read Event Tracking**: Logs all read events with metadata (IP, user agent, timestamp)
- **Sender Notifications**: Sends notifications to the sender when their email is first read
- **Configurable Delivery**: Supports email and SMS delivery methods
- **Privacy Protection**: Only sends metadata, never the actual email content

### Expiration Alerts
- **Reminder Notifications**: Sends alerts X hours before email expiration
- **Final Notifications**: Sends alerts when emails are deleted due to expiration
- **Configurable Timing**: Customizable alert timing per email or globally
- **Multiple Delivery Methods**: Email and SMS delivery support

## Database Schema

### New Tables

#### `read_events`
Tracks individual read events for audit and analytics.

```sql
CREATE TABLE read_events (
    event_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    read_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    country TEXT,
    city TEXT,
    device_type TEXT,
    is_first_read BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (email_id) REFERENCES emails(email_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);
```

#### `read_receipts`
Stores read receipt notification records.

```sql
CREATE TABLE read_receipts (
    receipt_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    recipient_id TEXT NOT NULL,
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    delivery_method TEXT NOT NULL, -- 'email' or 'sms'
    delivery_status TEXT DEFAULT 'pending', -- 'pending', 'sent', 'failed'
    error_message TEXT,
    metadata TEXT, -- JSON with read details
    FOREIGN KEY (email_id) REFERENCES emails(email_id),
    FOREIGN KEY (sender_id) REFERENCES users(user_id),
    FOREIGN KEY (recipient_id) REFERENCES users(user_id)
);
```

#### `expiration_alerts`
Stores expiration alert notification records.

```sql
CREATE TABLE expiration_alerts (
    alert_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    alert_type TEXT NOT NULL, -- 'reminder' or 'final'
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    delivery_method TEXT NOT NULL, -- 'email' or 'sms'
    delivery_status TEXT DEFAULT 'pending', -- 'pending', 'sent', 'failed'
    error_message TEXT,
    metadata TEXT, -- JSON with expiration details
    FOREIGN KEY (email_id) REFERENCES emails(email_id),
    FOREIGN KEY (sender_id) REFERENCES users(user_id)
);
```

#### `read_receipt_preferences`
Stores user preferences for read receipts and expiration alerts.

```sql
CREATE TABLE read_receipt_preferences (
    user_id TEXT PRIMARY KEY,
    enable_read_receipts BOOLEAN DEFAULT TRUE,
    enable_expiration_alerts BOOLEAN DEFAULT TRUE,
    expiration_alert_hours INTEGER DEFAULT 24,
    delivery_methods TEXT DEFAULT 'email,sms', -- Comma-separated list
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);
```

### Updated Tables

#### `emails` Table Additions
```sql
-- Read receipt fields
ALTER TABLE emails ADD COLUMN first_read_at DATETIME;
ALTER TABLE emails ADD COLUMN read_count INTEGER DEFAULT 0;
ALTER TABLE emails ADD COLUMN read_receipt_sent BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN expiration_alert_sent BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN final_expiration_alert_sent BOOLEAN DEFAULT FALSE;

-- Read receipt and expiration alert preferences
ALTER TABLE emails ADD COLUMN enable_read_receipts BOOLEAN DEFAULT TRUE;
ALTER TABLE emails ADD COLUMN enable_expiration_alerts BOOLEAN DEFAULT TRUE;
ALTER TABLE emails ADD COLUMN expiration_alert_hours INTEGER DEFAULT 24;
```

## API Endpoints

### Read Receipt Preferences

#### GET `/api/read-receipts/preferences`
Get user's read receipt and expiration alert preferences.

**Response:**
```json
{
    "user_id": "user123",
    "enable_read_receipts": true,
    "enable_expiration_alerts": true,
    "expiration_alert_hours": 24,
    "delivery_methods": "email,sms",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
}
```

#### PUT `/api/read-receipts/preferences`
Update user's read receipt and expiration alert preferences.

**Request Body:**
```json
{
    "enable_read_receipts": true,
    "enable_expiration_alerts": true,
    "expiration_alert_hours": 48,
    "delivery_methods": "email,sms"
}
```

### Email Read Receipt Information

#### GET `/api/emails/{id}/read-receipts`
Get read receipt and expiration information for a specific email.

**Response:**
```json
{
    "read_receipts": {
        "email_id": "email123",
        "first_read_at": "2024-01-01T12:00:00Z",
        "read_count": 3,
        "read_receipt_sent": true,
        "enable_read_receipts": true
    },
    "expiration": {
        "email_id": "email123",
        "expires_at": "2024-01-08T00:00:00Z",
        "expiration_alert_sent": false,
        "final_expiration_alert_sent": false,
        "enable_expiration_alerts": true,
        "expiration_alert_hours": 24
    }
}
```

#### GET `/api/emails/{id}/read-events`
Get read events for a specific email.

**Query Parameters:**
- `limit` (optional): Maximum number of events to return (default: 50, max: 100)

**Response:**
```json
[
    {
        "event_id": "event123",
        "email_id": "email123",
        "user_id": "user456",
        "read_at": "2024-01-01T12:00:00Z",
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "country": "US",
        "city": "New York",
        "device_type": "desktop",
        "is_first_read": true
    }
]
```

#### PUT `/api/emails/{id}/read-receipt-settings`
Update read receipt and expiration alert settings for a specific email.

**Request Body:**
```json
{
    "enable_read_receipts": true,
    "enable_expiration_alerts": true,
    "expiration_alert_hours": 12
}
```

## Implementation Details

### Read Receipt Service

The `ReadReceiptService` handles all read receipt and expiration alert functionality:

```go
type ReadReceiptService struct {
    db *sql.DB
}
```

#### Key Methods

- `RecordReadEvent(ctx, event)`: Records a read event and handles first read logic
- `SendReadReceipt(ctx, emailID, senderID, recipientID, readEvent)`: Sends read receipt notifications
- `GetReadReceiptPreferences(ctx, userID)`: Gets user preferences
- `UpdateReadReceiptPreferences(ctx, prefs)`: Updates user preferences
- `GetEmailReadReceiptInfo(ctx, emailID)`: Gets read receipt info for an email
- `GetReadEvents(ctx, emailID, limit)`: Gets read events for an email

### Expiration Worker

The `ExpirationWorker` processes expiration alerts periodically:

```go
type ExpirationWorker struct {
    db *sql.DB
}
```

#### Key Methods

- `ProcessExpirationAlerts(ctx)`: Main processing method
- `getEmailsNeedingReminder(ctx)`: Finds emails needing reminder alerts
- `getEmailsNeedingFinalAlert(ctx)`: Finds emails needing final alerts
- `sendExpirationReminder(ctx, email)`: Sends reminder alerts
- `sendFinalExpirationAlert(ctx, email)`: Sends final alerts

### Integration Points

#### Email Access Integration
When a recipient accesses an email via `/api/email/{id}/content`, the system:

1. Records the read event
2. Checks if this is the first read
3. Sends read receipt notification if enabled
4. Updates email read count and first read timestamp

#### Email Sending Integration
When sending emails, the system:

1. Sets default read receipt and expiration alert settings
2. Stores preferences in the emails table
3. Enables tracking for the new email

## Configuration

### Environment Variables

- `EXPIRATION_WORKER_INTERVAL_MINUTES`: Interval for expiration worker processing (default: 15)

### Default Settings

- **Read Receipts**: Enabled by default
- **Expiration Alerts**: Enabled by default
- **Alert Timing**: 24 hours before expiration
- **Delivery Methods**: Email and SMS

## Testing

### Unit Tests

Run the unit tests for the read receipt functionality:

```bash
go test ./pkg/readreceipts/...
```

### Integration Tests

Use the provided PowerShell test script:

```powershell
.\scripts\test_read_receipts.ps1
```

### Manual Testing

1. **Start the API server:**
   ```bash
   go run cmd/api/main.go
   ```

2. **Start the expiration worker:**
   ```bash
   go run cmd/expiration_worker/main.go
   ```

3. **Test read receipts:**
   - Send an email with read receipts enabled
   - Have a recipient access the email
   - Verify read receipt notification is sent

4. **Test expiration alerts:**
   - Send an email with short expiration time
   - Wait for reminder and final alerts
   - Verify alerts are sent to sender

## Security Considerations

### Privacy Protection
- Read receipts only include metadata (IP, user agent, timestamp)
- Never include actual email content in notifications
- Recipient information is masked in notifications

### Access Control
- Only email senders can view read receipt information
- Read events are protected by recipient-based access control
- All endpoints require JWT authentication

### Data Retention
- Read events are stored for audit purposes
- Read receipt and expiration alert records are maintained
- Consider implementing data retention policies

## Future Enhancements

### Planned Features
- **Real-time Notifications**: WebSocket-based real-time updates
- **Advanced Analytics**: Detailed read analytics and insights
- **Custom Notification Templates**: User-defined notification messages
- **Bulk Operations**: Batch processing for multiple emails

### Integration Opportunities
- **Email Service Integration**: Direct integration with email providers
- **SMS Service Integration**: Twilio or other SMS providers
- **Push Notifications**: Mobile app push notifications
- **Webhook Support**: External system notifications

## Troubleshooting

### Common Issues

1. **Read receipts not being sent:**
   - Check if read receipts are enabled for the email
   - Verify sender preferences allow read receipts
   - Check delivery method configuration

2. **Expiration alerts not working:**
   - Ensure expiration worker is running
   - Check email expiration settings
   - Verify alert timing configuration

3. **Database migration errors:**
   - Check if migration has already been applied
   - Verify database permissions
   - Review migration logs

### Debugging

Enable debug logging by setting the log level:

```bash
export LOG_LEVEL=debug
```

Check the application logs for detailed error information and processing status.

