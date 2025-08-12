# Notification Delivery Controls & Rate Limiting

## Overview

Micro-Iteration 4.18 enhances the Access Notification System with sophisticated delivery controls and rate limiting capabilities. This feature allows senders to control the frequency and conditions of notifications, reducing noise while maintaining strong security awareness.

## Features

### Delivery Frequency Controls

The system supports four delivery frequency modes:

1. **Immediate** - Send notification for every access attempt
2. **Daily Digest** - Send a daily summary of all access events
3. **First Attempt Only** - Send notification only for the first access attempt from each IP
4. **Threshold Trigger** - Send notification only after multiple failed attempts

### Rate Limiting

Prevents excessive notifications for repeated accesses from the same IP/device within a configurable time frame:

- **Time Window**: Configurable window (1-1440 minutes)
- **Max Notifications**: Configurable limit (1-100 notifications per window)
- **Per-IP Tracking**: Rate limiting is applied per email/IP combination

### Preference Storage

- **Global Settings**: User-wide notification preferences
- **Per-Email Settings**: Individual email notification preferences
- **Inheritance**: Per-email settings can inherit global settings or override them

### Audit Logging

- **Suppression Tracking**: All suppressed notifications are logged with reasons
- **Transparency**: Users can view suppression statistics and history
- **Compliance**: Full audit trail for security and compliance requirements

## API Endpoints

### Global Notification Preferences

#### GET /api/notifications/preferences
Retrieve global notification preferences for the authenticated user.

**Response:**
```json
{
  "user_id": "user123",
  "email_notifications": true,
  "sms_notifications": false,
  "notify_on_success": true,
  "notify_on_failure": true,
  "notify_on_blocked": true,
  "include_geolocation": true,
  "include_device_info": true,
  "delivery_frequency": "immediate",
  "threshold_attempts": 3,
  "rate_limit_window_minutes": 15,
  "rate_limit_max_notifications": 5,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### PUT /api/notifications/preferences
Update global notification preferences.

**Request Body:**
```json
{
  "email_notifications": true,
  "sms_notifications": false,
  "notify_on_success": true,
  "notify_on_failure": true,
  "notify_on_blocked": true,
  "include_geolocation": true,
  "include_device_info": true,
  "delivery_frequency": "threshold_trigger",
  "threshold_attempts": 5,
  "rate_limit_window_minutes": 30,
  "rate_limit_max_notifications": 10
}
```

### Per-Email Notification Preferences

#### GET /api/notifications/email/{emailID}/preferences
Retrieve notification preferences for a specific email.

**Response:**
```json
{
  "email_id": "email123",
  "user_id": "user123",
  "delivery_frequency": "first_attempt_only",
  "threshold_attempts": 3,
  "rate_limit_window_minutes": 15,
  "rate_limit_max_notifications": 5,
  "inherit_global_settings": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### PUT /api/notifications/email/{emailID}/preferences
Update notification preferences for a specific email.

**Request Body:**
```json
{
  "delivery_frequency": "daily_digest",
  "threshold_attempts": 3,
  "rate_limit_window_minutes": 60,
  "rate_limit_max_notifications": 3,
  "inherit_global_settings": false
}
```

### Notification Statistics

#### GET /api/notifications/stats
Retrieve notification statistics and suppression information.

**Response:**
```json
{
  "total_events": 150,
  "suppressed_events": 45,
  "suppression_stats": {
    "rate_limited": 20,
    "frequency_controlled": 15,
    "threshold_not_met": 8,
    "first_attempt_only": 2
  },
  "delivery_frequency": "immediate",
  "rate_limit_info": {
    "window_minutes": 15,
    "max_notifications": 5
  }
}
```

### Suppression History

#### GET /api/notifications/suppressions
Retrieve suppressed notification history.

**Query Parameters:**
- `limit` (optional): Number of records to return (default: 50, max: 200)

**Response:**
```json
[
  {
    "suppression_id": "supp123",
    "email_id": "email123",
    "user_id": "user123",
    "event_id": "event456",
    "suppression_reason": "rate_limited",
    "suppressed_at": "2024-01-01T12:00:00Z",
    "original_event_type": "success",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "country": "US",
    "city": "New York",
    "device_type": "Desktop",
    "failure_reason": null
  }
]
```

## Database Schema

### Enhanced Tables

#### notification_preferences
```sql
ALTER TABLE notification_preferences ADD COLUMN delivery_frequency TEXT DEFAULT 'immediate' CHECK (delivery_frequency IN ('immediate', 'daily_digest', 'first_attempt_only', 'threshold_trigger'));
ALTER TABLE notification_preferences ADD COLUMN threshold_attempts INTEGER DEFAULT 3;
ALTER TABLE notification_preferences ADD COLUMN rate_limit_window_minutes INTEGER DEFAULT 15;
ALTER TABLE notification_preferences ADD COLUMN rate_limit_max_notifications INTEGER DEFAULT 5;
```

#### email_notification_preferences
```sql
CREATE TABLE email_notification_preferences (
    email_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    delivery_frequency TEXT DEFAULT 'immediate' CHECK (delivery_frequency IN ('immediate', 'daily_digest', 'first_attempt_only', 'threshold_trigger')),
    threshold_attempts INTEGER DEFAULT 3,
    rate_limit_window_minutes INTEGER DEFAULT 15,
    rate_limit_max_notifications INTEGER DEFAULT 5,
    inherit_global_settings BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

#### notification_suppressions
```sql
CREATE TABLE notification_suppressions (
    suppression_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    suppression_reason TEXT NOT NULL CHECK (suppression_reason IN ('rate_limited', 'frequency_controlled', 'threshold_not_met', 'first_attempt_only')),
    suppressed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    original_event_type TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    country TEXT,
    city TEXT,
    device_type TEXT,
    failure_reason TEXT,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES access_events(event_id) ON DELETE CASCADE
);
```

#### notification_rate_limits
```sql
CREATE TABLE notification_rate_limits (
    rate_limit_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    notification_count INTEGER DEFAULT 1,
    window_start DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_notification_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

## Delivery Logic

### Decision Flow

1. **Event Recording**: All access events are recorded regardless of notification settings
2. **Preference Check**: Check if notifications are enabled for the event type
3. **Delivery Frequency Check**: Apply delivery frequency rules
4. **Rate Limiting Check**: Apply rate limiting rules
5. **Notification Decision**: Send notification or record suppression

### Delivery Frequency Rules

#### Immediate
- Send notification for every access attempt
- No suppression based on frequency

#### Daily Digest
- Suppress immediate notifications
- Events are tracked for daily summary
- Summary sent by separate process

#### First Attempt Only
- Check if this is the first attempt from this IP for this email
- Send notification only for first attempt
- Suppress subsequent attempts from same IP

#### Threshold Trigger
- Only applies to failure events
- Count total failures for this email
- Send notification only after threshold is reached

### Rate Limiting Rules

1. **Window Management**: Clean up expired rate limit records
2. **Count Check**: Check current notification count for email/IP combination
3. **Limit Enforcement**: Suppress if count exceeds limit
4. **Record Update**: Update count and timestamp for sent notifications

## Frontend Integration

### Components

#### NotificationDeliveryControls
React component for managing notification delivery settings:

```tsx
<NotificationDeliveryControls
  emailId="email123" // Optional - for per-email settings
  onSettingsChange={(settings) => {
    console.log('Settings updated:', settings);
  }}
/>
```

#### NotificationStats
React component for displaying notification statistics:

```tsx
<NotificationStats />
```

### Usage Examples

#### Global Settings
```tsx
import NotificationDeliveryControls from '@/components/secure/NotificationDeliveryControls';

function SettingsPage() {
  return (
    <div>
      <h2>Global Notification Settings</h2>
      <NotificationDeliveryControls />
    </div>
  );
}
```

#### Per-Email Settings
```tsx
function EmailDetailPage({ emailId }) {
  return (
    <div>
      <h2>Email Notification Settings</h2>
      <NotificationDeliveryControls emailId={emailId} />
    </div>
  );
}
```

#### Statistics Dashboard
```tsx
import NotificationStats from '@/components/secure/NotificationStats';

function DashboardPage() {
  return (
    <div>
      <h2>Notification Statistics</h2>
      <NotificationStats />
    </div>
  );
}
```

## Testing

### Unit Tests

Comprehensive test coverage for all delivery logic:

```bash
go test ./pkg/notification -v
```

### Test Scenarios

1. **Immediate Delivery**: Verify notifications sent for all events
2. **First Attempt Only**: Verify only first attempt triggers notification
3. **Threshold Trigger**: Verify notification only after threshold reached
4. **Rate Limiting**: Verify notifications limited within time window
5. **Per-Email Preferences**: Verify email-specific settings override global
6. **Suppression Logging**: Verify all suppressions are recorded

### Integration Tests

Test complete notification flow:

```bash
go test ./cmd/api -run TestNotificationDeliveryControls
```

## Security Considerations

### Privacy Protection

- **Metadata Stripping**: Sensitive metadata not stored in suppressions unless required
- **Retention Policies**: Suppression logs have configurable retention periods
- **Access Control**: Users can only view their own notification data

### Rate Limiting Security

- **IP-Based**: Rate limiting prevents abuse from single IP addresses
- **Configurable Limits**: Users can adjust limits based on their needs
- **Cleanup**: Automatic cleanup of expired rate limit records

### Audit Compliance

- **Full Logging**: All suppression decisions are logged with reasons
- **Transparency**: Users can view suppression statistics and history
- **Compliance**: Audit trail supports security and compliance requirements

## Configuration

### Environment Variables

```bash
# Notification delivery controls
NOTIFICATION_DEFAULT_FREQUENCY=immediate
NOTIFICATION_DEFAULT_THRESHOLD=3
NOTIFICATION_DEFAULT_RATE_LIMIT_WINDOW=15
NOTIFICATION_DEFAULT_RATE_LIMIT_MAX=5

# Suppression logging
NOTIFICATION_SUPPRESSION_RETENTION_DAYS=30
NOTIFICATION_ENABLE_SUPPRESSION_LOGGING=true
```

### Default Values

- **Delivery Frequency**: `immediate`
- **Threshold Attempts**: `3`
- **Rate Limit Window**: `15 minutes`
- **Rate Limit Max**: `5 notifications`
- **Suppression Retention**: `30 days`

## Migration Guide

### Database Migration

Apply the notification delivery controls migration:

```bash
# Apply migration
sqlite3 /var/db/secure-email.db < schema/migrate_add_notification_delivery_controls.sql
```

### Frontend Updates

1. Install new components:
```bash
npm install
```

2. Import and use components:
```tsx
import NotificationDeliveryControls from '@/components/secure/NotificationDeliveryControls';
import NotificationStats from '@/components/secure/NotificationStats';
```

3. Update existing notification settings pages to include new controls

### Backend Updates

1. Update notification service calls to use new delivery logic
2. Add new API endpoints to routing
3. Update existing handlers to support new parameters

## Troubleshooting

### Common Issues

#### Notifications Not Being Sent
1. Check delivery frequency settings
2. Verify rate limiting configuration
3. Review suppression logs for reasons

#### High Suppression Rate
1. Review rate limiting settings
2. Check delivery frequency configuration
3. Analyze suppression statistics

#### Performance Issues
1. Monitor rate limit table size
2. Check cleanup job execution
3. Review database indexes

### Debug Information

Enable debug logging:

```bash
export NOTIFICATION_DEBUG=true
```

View suppression logs:

```bash
# Query suppression statistics
sqlite3 /var/db/secure-email.db "SELECT suppression_reason, COUNT(*) FROM notification_suppressions GROUP BY suppression_reason;"
```

## Future Enhancements

### Planned Features

1. **Advanced Rate Limiting**: Device fingerprinting and behavioral analysis
2. **Smart Suppression**: Machine learning-based suppression decisions
3. **Notification Channels**: Support for additional notification channels
4. **Real-time Analytics**: Live notification delivery analytics
5. **Custom Rules**: User-defined notification rules and conditions

### API Extensions

1. **Bulk Operations**: Bulk update notification preferences
2. **Templates**: Predefined notification preference templates
3. **Scheduling**: Time-based notification delivery rules
4. **Escalation**: Automatic escalation for critical events

## Support

For questions or issues with notification delivery controls:

1. Check the troubleshooting section
2. Review suppression logs and statistics
3. Verify configuration settings
4. Contact support with detailed error information

## Changelog

### Version 4.18.0
- Initial implementation of notification delivery controls
- Added rate limiting functionality
- Implemented suppression tracking and logging
- Created frontend components for settings management
- Added comprehensive test coverage
- Updated API documentation
