# Advanced Audit Log & Export System

## Overview

Micro-Iteration 4.20 implements a comprehensive audit logging system for the Secure Email MVP. This system provides detailed tracking of all system events, configurable retention policies, and powerful export capabilities for compliance and analysis purposes.

## Features

### Core Functionality
- **Comprehensive Event Tracking**: Records all system events with detailed metadata
- **Configurable Retention Policies**: Automatic cleanup based on event type and age
- **Advanced Filtering & Search**: Multi-criteria filtering with full-text search
- **Export Capabilities**: CSV and JSON export with background processing
- **Security & Privacy**: Sensitive data masking and role-based access control
- **Real-time Dashboard**: Web-based interface for viewing and managing logs

### Event Types Tracked
- **Email Operations**: Creation, access, deletion, forwarding attempts
- **Authentication Events**: Login attempts, MFA setup, password changes
- **Security Events**: Failed access attempts, blocked requests, geolocation violations
- **System Events**: Read receipts, expiration alerts, cleanup operations
- **API Usage**: API key usage, rate limiting events

### Data Fields Captured
- **Basic Information**: Timestamp, event type, outcome, severity
- **User Context**: User ID, session ID, request ID
- **Network Information**: IP address, user agent, geolocation (country/city)
- **Device Information**: Device type, browser details
- **Related Resources**: Email IDs, API endpoints
- **Detailed Context**: JSON-formatted additional data

## Database Schema

### audit_log Table
```sql
CREATE TABLE audit_log (
    log_id TEXT PRIMARY KEY,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    event_type TEXT NOT NULL,
    user_id TEXT, -- NULL for system events
    ip_address TEXT,
    user_agent TEXT,
    related_email_id TEXT,
    outcome TEXT NOT NULL CHECK (outcome IN ('success', 'failure', 'blocked')),
    details TEXT, -- JSON with additional event details
    severity TEXT DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    session_id TEXT,
    request_id TEXT,
    country TEXT,
    city TEXT,
    device_type TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL,
    FOREIGN KEY (related_email_id) REFERENCES emails(email_id) ON DELETE SET NULL
);
```

### audit_log_retention Table
```sql
CREATE TABLE audit_log_retention (
    retention_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    retention_days INTEGER DEFAULT 90,
    auto_purge BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### audit_log_exports Table
```sql
CREATE TABLE audit_log_exports (
    export_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    export_type TEXT NOT NULL CHECK (export_type IN ('csv', 'json')),
    date_from DATETIME,
    date_to DATETIME,
    event_types TEXT, -- Comma-separated list
    filters TEXT, -- JSON with additional filters
    file_path TEXT,
    file_size INTEGER,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    expires_at DATETIME, -- When the export file should be deleted
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

## API Endpoints

### Query Audit Logs
```
GET /api/audit/logs
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `page_size` (int): Items per page (default: 100, max: 1000)
- `date_from` (ISO 8601): Start date filter
- `date_to` (ISO 8601): End date filter
- `event_types` (comma-separated): Filter by event types
- `outcomes` (comma-separated): Filter by outcomes (success, failure, blocked)
- `severities` (comma-separated): Filter by severities (info, warning, error, critical)
- `user_ids` (comma-separated): Filter by user IDs
- `ip_addresses` (comma-separated): Filter by IP addresses
- `related_email_ids` (comma-separated): Filter by related email IDs
- `search_term` (string): Full-text search in details and user agent

**Response:**
```json
{
  "events": [
    {
      "log_id": "uuid",
      "timestamp": "2024-01-15T10:30:00Z",
      "event_type": "email_creation",
      "user_id": "user123",
      "ip_address": "192.168.1.100",
      "outcome": "success",
      "severity": "info",
      "details": { "email_id": "email456", "subject": "Test" },
      "country": "US",
      "city": "New York"
    }
  ],
  "total": 150,
  "page": 1,
  "page_size": 25,
  "has_more": true
}
```

### Get Event Types
```
GET /api/audit/event-types
```

**Response:**
```json
{
  "event_types": ["email_creation", "email_access", "login_attempt", ...]
}
```

### Get User Events
```
GET /api/audit/user-events?limit=10
```

**Response:**
```json
{
  "events": [...],
  "total": 10
}
```

### Create Export
```
POST /api/audit/exports
```

**Request Body:**
```json
{
  "export_type": "json",
  "filter": {
    "date_from": "2024-01-01T00:00:00Z",
    "date_to": "2024-01-31T23:59:59Z",
    "event_types": ["email_creation", "email_access"]
  }
}
```

**Response:**
```json
{
  "export_id": "uuid",
  "user_id": "user123",
  "export_type": "json",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-16T10:30:00Z"
}
```

### Get Export Status
```
GET /api/audit/exports/{export_id}
```

### Download Export
```
GET /api/audit/exports/{export_id}/download
```

### Delete Export
```
DELETE /api/audit/exports/{export_id}
```

### Get Retention Policies
```
GET /api/audit/retention-policies
```

### Update Retention Policy
```
PUT /api/audit/retention-policies/{event_type}
```

**Request Body:**
```json
{
  "retention_days": 365,
  "auto_purge": true
}
```

### Admin Endpoints
```
POST /api/audit/purge-expired
POST /api/audit/cleanup-exports
```

## Implementation Details

### Backend Services

#### AuditService
- **RecordEvent**: Records new audit events with metadata
- **QueryEvents**: Retrieves events with filtering and pagination
- **GetEventTypes**: Returns all available event types
- **GetUserEvents**: Gets events for a specific user
- **GetRetentionPolicies**: Retrieves retention policy configuration
- **UpdateRetentionPolicy**: Updates retention settings
- **PurgeExpiredLogs**: Removes logs based on retention policies

#### ExportService
- **CreateExportRequest**: Creates new export requests
- **ProcessExport**: Generates export files in background
- **GetExportRequest**: Retrieves export status and metadata
- **GetUserExports**: Lists user's export requests
- **DeleteExport**: Removes export requests and files
- **CleanupExpiredExports**: Removes expired export files

### Frontend Components

#### AuditLogDashboard
- **Event Viewer**: Table with filtering and pagination
- **Export Manager**: Create and manage export requests
- **Retention Settings**: Configure retention policies
- **Summary Cards**: Overview statistics
- **Recent Events**: Quick view of recent activity

### Background Workers

#### Audit Worker (`cmd/audit_worker/main.go`)
- Runs periodic cleanup tasks
- Purges expired logs based on retention policies
- Cleans up expired export files
- Configurable via `AUDIT_CLEANUP_INTERVAL_MINUTES` environment variable

## Configuration

### Environment Variables
```bash
# Audit cleanup interval (minutes)
AUDIT_CLEANUP_INTERVAL_MINUTES=60

# Export file lifetime (hours)
AUDIT_EXPORT_LIFETIME_HOURS=24

# Export directory
AUDIT_EXPORT_DIR=/tmp/audit-exports
```

### Default Retention Policies
- **Email Creation**: 365 days
- **Email Access**: 90 days
- **Email Deletion**: 365 days
- **Login Attempts**: 90 days
- **API Key Usage**: 90 days
- **Read Receipts**: 90 days
- **Expiration Alerts**: 90 days
- **System Events**: 180 days

## Security & Privacy

### Data Protection
- **Sensitive Data Masking**: IP addresses and user IDs are masked in UI
- **Role-Based Access**: Users can only see their own events (admins see all)
- **Secure Export**: Export files are stored securely with expiration
- **Audit Trail**: All audit operations are themselves logged

### Compliance Features
- **Data Retention**: Configurable retention policies for different event types
- **Export Capabilities**: CSV and JSON export for compliance reporting
- **Search & Filter**: Advanced filtering for compliance investigations
- **Tamper Evidence**: Immutable audit trail with cryptographic integrity

## Usage Examples

### Recording Events
```go
// Record email creation event
event := &audit.AuditEvent{
    EventType: audit.EventTypeEmailCreation,
    UserID:    &userID,
    IPAddress: &clientIP,
    Outcome:   audit.OutcomeSuccess,
    Severity:  audit.SeverityInfo,
    Details: map[string]interface{}{
        "email_id": emailID,
        "subject":  subject,
        "recipient": recipient,
    },
}
auditService.RecordEvent(ctx, event)
```

### Querying Events
```go
// Query events with filters
filter := audit.AuditLogFilter{
    DateFrom:   &startDate,
    DateTo:     &endDate,
    EventTypes: []audit.EventType{audit.EventTypeEmailAccess},
    Outcomes:   []audit.Outcome{audit.OutcomeFailure},
}
result, err := auditService.QueryEvents(ctx, filter, 1, 100)
```

### Creating Exports
```go
// Create JSON export
export, err := exportService.CreateExportRequest(ctx, userID, "json", filter)
if err != nil {
    return err
}

// Process export in background
go func() {
    if err := exportService.ProcessExport(context.Background(), export.ExportID); err != nil {
        log.Printf("Export failed: %v", err)
    }
}()
```

## Testing

### Unit Tests
```bash
# Run audit service tests
go test ./pkg/audit -v

# Run export service tests
go test ./pkg/audit -run TestExport -v
```

### Integration Tests
```bash
# Run audit log integration tests
./scripts/test_audit_logs.ps1
```

### Manual Testing
1. Start the API server and audit worker
2. Perform various operations (login, send emails, etc.)
3. View audit logs in the dashboard
4. Create and download exports
5. Test filtering and search functionality

## Monitoring & Maintenance

### Health Checks
- Monitor audit log table size and growth
- Check export file cleanup is working
- Verify retention policies are being applied
- Monitor audit worker logs for errors

### Performance Considerations
- Indexes on frequently queried columns
- Pagination for large result sets
- Background processing for exports
- Regular cleanup of old data

### Troubleshooting

#### Common Issues
1. **Export files not generating**: Check export directory permissions
2. **Audit logs not being recorded**: Verify audit service is initialized
3. **Cleanup not working**: Check audit worker is running
4. **Performance issues**: Review indexes and query optimization

#### Log Analysis
```sql
-- Check audit log growth
SELECT 
    DATE(created_at) as date,
    COUNT(*) as events,
    COUNT(DISTINCT user_id) as users
FROM audit_log 
WHERE created_at >= DATE('now', '-30 days')
GROUP BY DATE(created_at)
ORDER BY date;

-- Check export status
SELECT status, COUNT(*) as count
FROM audit_log_exports
GROUP BY status;

-- Check retention policy compliance
SELECT 
    event_type,
    COUNT(*) as total_events,
    COUNT(CASE WHEN created_at < DATE('now', '-90 days') THEN 1 END) as old_events
FROM audit_log
GROUP BY event_type;
```

## Future Enhancements

### Planned Features
- **Real-time Alerts**: WebSocket notifications for critical events
- **Advanced Analytics**: Event correlation and anomaly detection
- **Custom Dashboards**: User-configurable dashboard layouts
- **API Rate Limiting**: Per-user audit log query limits
- **Data Archival**: Long-term storage for compliance requirements

### Integration Opportunities
- **SIEM Integration**: Export to security information systems
- **Compliance Frameworks**: Built-in compliance reporting
- **Machine Learning**: Anomaly detection and threat analysis
- **Third-party Tools**: Integration with external audit tools





















