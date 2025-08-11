# Notification System Implementation

## Overview

The Secure Email MVP notification system provides comprehensive access monitoring and alerting capabilities. It tracks all email access attempts (successful, failed, and blocked) and sends notifications to senders based on their preferences.

## Features

### Core Functionality
- **Access Event Tracking**: Records all email access attempts with metadata
- **Multi-Channel Notifications**: Email and SMS notification support
- **Configurable Preferences**: Granular control over notification settings
- **Access History**: Complete audit trail of access events
- **Device Detection**: Automatic device type identification
- **Geolocation Integration**: Location-based access tracking

### Notification Types
- **Success Notifications**: When emails are successfully accessed
- **Failure Notifications**: When access attempts fail
- **Blocked Notifications**: When access is blocked by security measures

### Metadata Captured
- IP address and geolocation (city, country)
- Device type (Desktop, Mobile, Tablet)
- User agent information
- Timestamp and event details
- Failure reasons (when applicable)

## Backend Implementation

### Database Schema

#### notification_preferences Table
```sql
CREATE TABLE notification_preferences (
    user_id TEXT PRIMARY KEY,
    email_notifications BOOLEAN DEFAULT TRUE,
    sms_notifications BOOLEAN DEFAULT FALSE,
    notify_on_success BOOLEAN DEFAULT TRUE,
    notify_on_failure BOOLEAN DEFAULT TRUE,
    notify_on_blocked BOOLEAN DEFAULT TRUE,
    include_geolocation BOOLEAN DEFAULT TRUE,
    include_device_info BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

#### access_events Table
```sql
CREATE TABLE access_events (
    event_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    event_type TEXT NOT NULL CHECK (event_type IN ('success', 'failure', 'blocked')),
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    country TEXT,
    city TEXT,
    device_type TEXT,
    failure_reason TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

### Core Components

#### NotificationService (`pkg/notification/notification.go`)
- **RecordAccessEvent**: Records access events in the database
- **GetNotificationPreferences**: Retrieves user notification preferences
- **UpdateNotificationPreferences**: Updates user notification settings
- **SendNotification**: Sends notifications based on preferences
- **GetAccessEventHistory**: Retrieves access event history

#### Notification Handlers (`cmd/api/notification_handlers.go`)
- **getNotificationPreferencesHandler**: GET /api/notifications/preferences
- **updateNotificationPreferencesHandler**: PUT /api/notifications/preferences
- **getAccessEventHistoryHandler**: GET /api/notifications/history
- **recordAccessEvent**: Helper function for recording events

### Integration Points

#### Email Access Flow
The notification system is integrated into the email access flow in `cmd/api/view_email_handler.go`:

1. **IP Lockout Check**: Records blocked events when IP is locked out
2. **Geolocation Verification**: Records blocked events for location restrictions
3. **Successful Access**: Records success events when email is accessed
4. **Security Failures**: Records failure events for various security checks

#### Event Recording
```go
// Record successful access event
if err := srv.recordAccessEvent(r.Context(), emailID, senderID, notification.AccessEventTypeSuccess, r, ""); err != nil {
    log.Printf("Failed to record successful access event: %v", err)
}

// Record blocked access event
if err := srv.recordAccessEvent(r.Context(), emailID, senderID, notification.AccessEventTypeBlocked, r, "IP lockout"); err != nil {
    log.Printf("Failed to record blocked access event: %v", err)
}
```

## Frontend Implementation

### API Integration (`src/lib/api.ts`)

#### Interfaces
```typescript
export interface NotificationPreferences {
  user_id: string;
  email_notifications: boolean;
  sms_notifications: boolean;
  notify_on_success: boolean;
  notify_on_failure: boolean;
  notify_on_blocked: boolean;
  include_geolocation: boolean;
  include_device_info: boolean;
  created_at: string;
  updated_at: string;
}

export interface AccessEvent {
  event_id: string;
  email_id: string;
  user_id: string;
  event_type: 'success' | 'failure' | 'blocked';
  ip_address: string;
  user_agent: string;
  country?: string;
  city?: string;
  device_type?: string;
  failure_reason?: string;
  timestamp: string;
}
```

#### API Functions
```typescript
export const getNotificationPreferences = async (): Promise<NotificationPreferences>
export const updateNotificationPreferences = async (preferences: Partial<NotificationPreferences>): Promise<NotificationPreferences>
export const getAccessEventHistory = async (limit?: number): Promise<AccessEvent[]>
```

### UI Components

#### NotificationPreferences Component (`src/components/secure/NotificationPreferences.tsx`)
- **Dual Tab Interface**: Preferences and Access History
- **Toggle Controls**: Easy preference management
- **Real-time Updates**: Immediate preference changes
- **Event Visualization**: Clear access event display

#### Features
- **Notification Channels**: Email and SMS toggles
- **Event Types**: Success, failure, and blocked notifications
- **Information Control**: Geolocation and device info toggles
- **Access History**: Complete event timeline with metadata

## Security Considerations

### Privacy Protection
- **Generic Error Messages**: No sensitive information in notifications
- **Configurable Metadata**: Users control what information is included
- **Secure Storage**: All data encrypted and properly secured

### Data Handling
- **Minimal Data Collection**: Only necessary metadata is captured
- **User Control**: Complete control over notification preferences
- **Audit Trail**: Comprehensive logging for security monitoring

### Notification Content
- **Generic Subjects**: "Secure Email Access Notification"
- **No Sensitive Data**: No email content or personal information
- **Location Privacy**: Optional geolocation inclusion

## Testing

### Backend Tests (`pkg/notification/notification_test.go`)
- **Service Tests**: Notification service functionality
- **Preference Tests**: User preference management
- **Event Tests**: Access event recording and retrieval
- **Integration Tests**: End-to-end notification workflows

### Frontend Tests
- **Component Tests**: UI component functionality
- **API Tests**: API integration and error handling
- **User Experience Tests**: Preference management workflows

## Configuration

### Environment Variables
```bash
# Email Service Configuration (for production)
SENDGRID_API_KEY=your_sendgrid_api_key
SENDGRID_FROM_EMAIL=noreply@yourdomain.com

# SMS Service Configuration (for production)
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_twilio_auth_token
TWILIO_FROM_NUMBER=+1234567890
```

### Default Settings
- **Email Notifications**: Enabled by default
- **SMS Notifications**: Disabled by default
- **Success Notifications**: Enabled by default
- **Failure Notifications**: Enabled by default
- **Blocked Notifications**: Enabled by default
- **Geolocation**: Included by default
- **Device Info**: Included by default

## API Endpoints

### GET /api/notifications/preferences
Retrieves current notification preferences for the authenticated user.

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "user-123",
    "email_notifications": true,
    "sms_notifications": false,
    "notify_on_success": true,
    "notify_on_failure": true,
    "notify_on_blocked": true,
    "include_geolocation": true,
    "include_device_info": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### PUT /api/notifications/preferences
Updates notification preferences for the authenticated user.

**Request:**
```json
{
  "email_notifications": true,
  "sms_notifications": false,
  "notify_on_success": true,
  "notify_on_failure": true,
  "notify_on_blocked": true,
  "include_geolocation": true,
  "include_device_info": true
}
```

### GET /api/notifications/history
Retrieves access event history for the authenticated user.

**Query Parameters:**
- `limit` (optional): Number of events to return (default: 50, max: 100)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "event_id": "event-123",
      "email_id": "email-456",
      "user_id": "user-123",
      "event_type": "success",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "country": "US",
      "city": "New York",
      "device_type": "Desktop",
      "failure_reason": null,
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ]
}
```

## Usage Examples

### Setting Up Notifications
1. Click the notification bell icon in the header
2. Configure notification channels (email/SMS)
3. Choose which events to receive notifications for
4. Configure what information to include
5. Save preferences

### Viewing Access History
1. Open notification preferences
2. Switch to "Access History" tab
3. View recent access events with metadata
4. Monitor for suspicious activity

### Customizing Notifications
- **Email Only**: Disable SMS, enable email notifications
- **Security Focus**: Enable only failure and blocked notifications
- **Privacy Focus**: Disable geolocation and device information
- **Minimal**: Disable all notifications except critical failures

## Future Enhancements

### Planned Features
- **Real-time Notifications**: WebSocket-based live updates
- **Advanced Filtering**: Filter access events by type, location, device
- **Notification Templates**: Customizable notification content
- **Bulk Operations**: Manage multiple notification settings
- **Analytics Dashboard**: Access pattern analysis and insights

### Integration Opportunities
- **Slack Integration**: Send notifications to Slack channels
- **Webhook Support**: Custom webhook notifications
- **Mobile App**: Push notifications for mobile devices
- **Advanced Analytics**: Machine learning-based anomaly detection

## Troubleshooting

### Common Issues
- **Notifications Not Sending**: Check email/SMS service configuration
- **Missing Events**: Verify database connectivity and permissions
- **UI Not Loading**: Check API endpoint availability and authentication
- **Preferences Not Saving**: Verify user authentication and permissions

### Debug Information
- **Backend Logs**: Check server logs for notification errors
- **Database Queries**: Verify event recording in access_events table
- **API Responses**: Check network tab for API call failures
- **User Permissions**: Ensure user has proper authentication

## Conclusion

The notification system provides comprehensive access monitoring and alerting capabilities for the Secure Email MVP. It balances security monitoring with user privacy, offering granular control over notification preferences while maintaining a complete audit trail of access events.

The implementation follows security best practices, provides excellent user experience, and integrates seamlessly with existing security features. The modular design allows for easy extension and customization as requirements evolve.
