# Suspicious Access Pattern Detection

## Overview

**Micro-Iteration 4.18** implements an advanced suspicious access pattern detection system for the Secure Email MVP. This feature automatically monitors all email access attempts in real-time, applies configurable detection rules, and flags potentially malicious behavior for sender review.

## Key Features

- **Real-time Access Monitoring**: Inspects all email access attempts (successful/failed) with comprehensive metadata collection
- **Configurable Detection Rules**: Four built-in detection patterns with customizable thresholds and time windows
- **Automatic Flagging**: Emails are automatically flagged when suspicious patterns are detected
- **Detailed Event Logging**: Complete audit trail of all detection events with metadata
- **User Preferences**: Configurable settings for detection sensitivity and notification preferences
- **Resolution Workflow**: Tools for reviewing and resolving flagged detections
- **Frontend Dashboard**: Rich UI for managing suspicious activity and preferences

## Detection Rules

### 1. Multiple Failed Attempts
- **Type**: `multiple_failed_attempts`
- **Default Threshold**: 3 failed attempts within 5 minutes
- **Severity**: High
- **Description**: Flags emails when multiple failed access attempts occur within a short time window

### 2. Unusual Geolocation
- **Type**: `unusual_geolocation`
- **Default Threshold**: 1 access from new location
- **Severity**: Medium
- **Description**: Flags emails when accessed from geolocations not seen in previous successful accesses

### 3. Rapid Multiple IPs
- **Type**: `rapid_multiple_ips`
- **Default Threshold**: 2 different IPs within 10 minutes
- **Severity**: High
- **Description**: Flags emails when accessed from multiple different IP addresses within a short time

### 4. Impossible Travel
- **Type**: `impossible_travel`
- **Default Threshold**: 1 access from different country within 5 minutes
- **Severity**: Critical
- **Description**: Flags emails when access occurs from geographically impossible locations within a very short time

## Database Schema

### New Tables

#### `suspicious_access_events`
Stores detailed detection events with metadata.

```sql
CREATE TABLE suspicious_access_events (
    detection_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    detection_type TEXT NOT NULL,
    detection_rule TEXT NOT NULL,
    severity TEXT NOT NULL,
    triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    resolved_by TEXT,
    resolution_notes TEXT,
    detection_metadata TEXT, -- JSON field
    FOREIGN KEY (email_id) REFERENCES emails(email_id),
    FOREIGN KEY (resolved_by) REFERENCES users(user_id)
);
```

#### `detection_rules`
Stores configurable detection rules.

```sql
CREATE TABLE detection_rules (
    rule_id TEXT PRIMARY KEY,
    rule_name TEXT NOT NULL UNIQUE,
    rule_type TEXT NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    threshold_value INTEGER NOT NULL,
    time_window_minutes INTEGER NOT NULL,
    severity TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### `user_suspicious_activity_preferences`
Stores user preferences for suspicious activity detection.

```sql
CREATE TABLE user_suspicious_activity_preferences (
    user_id TEXT PRIMARY KEY,
    enable_suspicious_detection BOOLEAN DEFAULT TRUE,
    notify_on_suspicious_activity BOOLEAN DEFAULT TRUE,
    auto_flag_suspicious_emails BOOLEAN DEFAULT TRUE,
    minimum_severity_for_notification TEXT DEFAULT 'medium',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);
```

### Enhanced `emails` Table

Added suspicious flag fields to the existing emails table:

```sql
ALTER TABLE emails ADD COLUMN suspicious_flag BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN suspicious_flag_set_at DATETIME;
ALTER TABLE emails ADD COLUMN suspicious_flag_cleared_at DATETIME;
ALTER TABLE emails ADD COLUMN suspicious_flag_cleared_by TEXT;
```

## Backend Implementation

### SuspiciousDetectionService

The core service provides methods for:

- **ProcessAccessEvent**: Analyzes access events and applies detection rules
- **GetUserPreferences**: Retrieves user detection preferences
- **GetEnabledDetectionRules**: Gets active detection rules
- **GetSuspiciousAccessEvents**: Retrieves detection events for an email
- **ClearSuspiciousFlag**: Clears suspicious flags on emails
- **ResolveDetectionEvent**: Marks detection events as resolved

### Integration with Access Flow

The suspicious detection is integrated into the existing access event recording:

```go
// In recordAccessEvent function
if srv.suspiciousService != nil {
    if err := srv.suspiciousService.ProcessAccessEvent(ctx, emailID, userID, ipAddress, userAgent, country, city, deviceType, failureReason, string(eventType)); err != nil {
        log.Printf("Failed to process suspicious access detection: %v", err)
    }
}
```

## API Endpoints

### Suspicious Activity Management

#### GET `/api/suspicious/activity/{email_id}`
Returns suspicious activity information for a specific email.

**Response:**
```json
{
  "email_id": "email123",
  "suspicious_flag": true,
  "suspicious_flag_set_at": "2024-01-15T10:30:00Z",
  "detection_events": [
    {
      "detection_id": "detection123",
      "email_id": "email123",
      "detection_type": "multiple_failed_attempts",
      "detection_rule": "rule_001",
      "severity": "high",
      "triggered_at": "2024-01-15T10:30:00Z",
      "detection_metadata": {
        "failed_attempts": 3,
        "time_window": "5 minutes"
      }
    }
  ],
  "total_detections": 1,
  "unresolved_detections": 1
}
```

#### POST `/api/suspicious/clear-flag/{email_id}`
Clears the suspicious flag on an email.

**Request:**
```json
{
  "resolution_notes": "False positive - legitimate access"
}
```

#### POST `/api/suspicious/resolve/{detection_id}`
Resolves a specific detection event.

**Request:**
```json
{
  "resolution_notes": "False positive - legitimate access"
}
```

### User Preferences

#### GET `/api/suspicious/preferences`
Returns user preferences for suspicious activity detection.

**Response:**
```json
{
  "user_id": "user123",
  "enable_suspicious_detection": true,
  "notify_on_suspicious_activity": true,
  "auto_flag_suspicious_emails": true,
  "minimum_severity_for_notification": "medium"
}
```

#### PUT `/api/suspicious/preferences`
Updates user preferences.

**Request:**
```json
{
  "enable_suspicious_detection": true,
  "notify_on_suspicious_activity": true,
  "auto_flag_suspicious_emails": true,
  "minimum_severity_for_notification": "high"
}
```

### Detection Rules

#### GET `/api/suspicious/rules`
Returns all detection rules.

**Response:**
```json
[
  {
    "rule_id": "rule_001",
    "rule_name": "Multiple Failed Attempts",
    "rule_type": "multiple_failed_attempts",
    "is_enabled": true,
    "threshold_value": 3,
    "time_window_minutes": 5,
    "severity": "high",
    "description": "Flag email if 3 or more failed access attempts within 5 minutes"
  }
]
```

### Suspicious Emails List

#### GET `/api/suspicious/emails`
Returns list of user's emails with suspicious flags.

**Response:**
```json
{
  "suspicious_emails": [
    {
      "email_id": "email123",
      "subject": "Important Document",
      "suspicious_flag": true,
      "suspicious_flag_set_at": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-15T09:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 1,
    "total_pages": 1
  }
}
```

## Frontend Implementation

### SuspiciousAccessDashboard Component

The main dashboard component provides:

- **Overview Cards**: Summary statistics for suspicious emails, detection rules, and system status
- **Suspicious Emails Table**: List of flagged emails with actions
- **Activity Details Dialog**: Detailed view of detection events for a specific email
- **User Preferences Dialog**: Configuration for detection settings
- **Detection Rules Dialog**: View of all detection rules and their status
- **Resolution Workflows**: Tools for clearing flags and resolving detections

### Key Features

- **Real-time Updates**: Automatic refresh of suspicious email list
- **Visual Indicators**: Color-coded severity levels and status indicators
- **Metadata Display**: Rich display of detection metadata with icons
- **Action Buttons**: Quick actions for viewing activity and clearing flags
- **Responsive Design**: Works on desktop and mobile devices

## Configuration

### Environment Variables

No additional environment variables are required. The system uses existing database connections and notification services.

### Default Detection Rules

The system includes four pre-configured detection rules:

1. **Multiple Failed Attempts**: 3 attempts in 5 minutes (High severity)
2. **Unusual Geolocation**: 1 new location (Medium severity)
3. **Rapid Multiple IPs**: 2 IPs in 10 minutes (High severity)
4. **Impossible Travel**: 1 different country in 5 minutes (Critical severity)

### User Preferences Defaults

- **Enable Suspicious Detection**: `true`
- **Notify on Suspicious Activity**: `true`
- **Auto-Flag Suspicious Emails**: `true`
- **Minimum Severity for Notification**: `medium`

## Security & Privacy

### Data Protection

- **No Sensitive Content**: Detection logs never contain email content or sensitive data
- **Anonymized Descriptions**: UI displays generic descriptions for security events
- **Role-Based Access**: Only email owners can view their suspicious activity
- **Audit Trail**: All detection events are logged for compliance

### Privacy Considerations

- **IP Address Handling**: IP addresses are stored but can be hashed/truncated
- **Geolocation Data**: Location data is stored but can be anonymized
- **User Consent**: Users can disable suspicious detection entirely
- **Data Retention**: Detection events follow the same retention policies as other audit logs

## Testing

### Unit Tests

The `pkg/suspicious/suspicious_test.go` file includes comprehensive unit tests for:

- Service initialization and configuration
- User preferences management
- Detection rule retrieval
- Multiple failed attempts detection
- Unusual geolocation detection
- Rapid multiple IPs detection
- Suspicious flag management
- Detection event resolution

### Integration Tests

The `scripts/test_suspicious_detection.ps1` script provides integration testing for:

- User authentication and preferences
- Detection rules retrieval
- Suspicious email listing
- Activity monitoring and flagging
- Flag clearing and event resolution

### Manual Testing

To test the suspicious detection system:

1. **Start the API server** with the new migration applied
2. **Create a test email** using the send endpoint
3. **Simulate suspicious activity** by:
   - Making multiple failed access attempts
   - Accessing from different IP addresses
   - Accessing from different geolocations
4. **Check the suspicious dashboard** for flagged emails
5. **Review detection events** and metadata
6. **Test resolution workflows** by clearing flags and resolving events

## Monitoring & Maintenance

### Performance Considerations

- **Database Indexes**: Proper indexes on detection event tables for efficient queries
- **Background Processing**: Detection analysis runs asynchronously to avoid blocking access
- **Caching**: User preferences and detection rules can be cached for performance
- **Cleanup**: Old detection events are automatically cleaned up based on retention policies

### Logging

The system logs all detection events with appropriate levels:

- **INFO**: Normal detection events and flag operations
- **WARNING**: High-severity detections
- **ERROR**: Detection processing failures

### Metrics

Key metrics to monitor:

- **Detection Rate**: Number of detections per time period
- **False Positive Rate**: Percentage of resolved detections marked as false positives
- **Response Time**: Time to resolve suspicious flags
- **System Performance**: Impact on email access response times

## Troubleshooting

### Common Issues

1. **No Detections Triggered**
   - Check if suspicious detection is enabled in user preferences
   - Verify detection rules are active
   - Ensure access events are being recorded properly

2. **High False Positive Rate**
   - Adjust detection rule thresholds
   - Review and tune geolocation detection logic
   - Consider user-specific patterns and preferences

3. **Performance Issues**
   - Check database indexes on detection tables
   - Monitor query performance for detection analysis
   - Consider caching frequently accessed data

### Debug Mode

Enable debug logging by setting the log level to DEBUG in the application configuration.

## Future Enhancements

### Planned Features

1. **Machine Learning Detection**: Advanced pattern recognition using ML models
2. **Behavioral Analysis**: User behavior profiling for more accurate detection
3. **Real-time Alerts**: Push notifications for critical detections
4. **Advanced Geolocation**: More sophisticated location-based detection
5. **Integration with SIEM**: Export detection events to security information systems

### Customization Options

1. **Custom Detection Rules**: User-defined detection patterns
2. **Adaptive Thresholds**: Dynamic threshold adjustment based on user patterns
3. **Whitelist Management**: IP and location whitelisting
4. **Advanced Notifications**: Custom notification channels and formats

## Conclusion

The Suspicious Access Pattern Detection system provides a comprehensive security layer for the Secure Email MVP, automatically identifying and flagging potentially malicious access patterns while maintaining user privacy and system performance. The system is designed to be configurable, scalable, and maintainable, with extensive testing and documentation to ensure reliable operation in production environments.







