# Micro-Iteration 4.28: Real-Time Retention Monitoring & Adaptive Policy Enforcement

## Overview

Micro-Iteration 4.28 introduces advanced real-time monitoring capabilities and intelligent adaptive policy enforcement to the Secure Email MVP. This iteration builds upon the successful insights and recommendations system from 4.27, adding live monitoring, event streaming, and automated policy adjustments with comprehensive safety controls.

## Key Features Implemented

### 1. Real-Time Monitoring Layer

#### RetentionMonitorService
- **Live Event Processing**: Processes policy evaluation events in real-time rather than waiting for scheduled rollups
- **In-Memory Cache**: Maintains active retention metrics (per user/domain/status) for instant access
- **Event Stream Integration**: Captures new deletions, archival operations, and policy overrides instantly
- **WebSocket/SSE Ready**: Supports push notifications for Admin UI with event streaming capabilities

#### Real-Time Metrics
- **Global Metrics**: System-wide retention statistics and performance indicators
- **User-Specific Metrics**: Individual user retention patterns and storage usage
- **Domain Metrics**: Domain-level retention analysis and trends
- **Policy Metrics**: Real-time policy performance and effectiveness tracking

#### Event Processing
- **Policy Evaluation Events**: Real-time tracking of policy matches and applications
- **Email Deletion Events**: Live monitoring of email cleanup operations
- **Email Archival Events**: Tracking of archival operations and compression ratios
- **Policy Change Events**: Monitoring of adaptive policy adjustments

### 2. Adaptive Policy Enforcement

#### AdaptiveRetentionEngine
- **Dynamic Policy Adjustment**: Adjusts policy parameters on-the-fly based on live metrics
- **Load-Based Optimization**: Shortens retention days when archival load is high
- **Storage Optimization**: Extends retention days when archival space is under-utilized
- **Safety Thresholds**: Implements cooldown periods and maximum change limits to prevent oscillations

#### Adaptive Recommendations
- **High Load Detection**: Automatically reduces retention days when archival load exceeds 80%
- **Low Utilization Optimization**: Increases retention days when load is below 20% and storage is high
- **Policy Utilization Analysis**: Switches policies to archival mode when underutilized
- **Risk Assessment**: Categorizes changes as low, medium, or high risk

#### Safety Controls
- **Cooldown Periods**: Prevents rapid policy changes with configurable cooldown days
- **Maximum Change Limits**: Restricts percentage changes to prevent drastic adjustments
- **Admin Approval**: Requires explicit admin opt-in for each policy's adaptive features
- **Impact Analysis**: Calculates expected storage savings and archival load impact

### 3. Enhanced Admin APIs

#### Real-Time Monitoring Endpoints
- `GET /api/admin/email/retention-realtime` - Retrieve live retention metrics
- `GET /api/admin/email/retention-realtime?metric_type=user&metric_key=user@example.com` - User-specific metrics
- `GET /api/admin/email/retention-realtime?metric_type=domain&metric_key=example.com` - Domain-specific metrics
- `GET /api/admin/email/retention-realtime?metric_type=policy&metric_key=1` - Policy-specific metrics

#### Adaptive Policy Management
- `GET /api/admin/email/adaptive-policy-changes` - View history of automatic adjustments
- `POST /api/admin/email/adaptive-policy/enable` - Enable adaptive enforcement per policy
- `POST /api/admin/email/adaptive-policy/disable` - Disable adaptive enforcement per policy
- `POST /api/admin/email/adaptive-policy/apply` - Apply adaptive changes with preview mode
- `GET /api/admin/email/policy-performance?policy_id=1` - Analyze policy performance
- `POST /api/admin/email/adaptive-policy/generate-recommendations` - Generate new recommendations

### 4. Configuration & Controls

#### Environment Variables
```bash
# Real-Time Monitoring
ENABLE_REALTIME_RETENTION_MONITORING=true
ENABLE_ADAPTIVE_POLICY_ENFORCEMENT=false

# Adaptive Policy Controls
ADAPTIVE_POLICY_COOLDOWN_DAYS=7
ADAPTIVE_POLICY_MAX_CHANGE_PERCENT=20
```

#### Policy-Level Configuration
- **Adaptive Enabled**: Per-policy toggle for adaptive adjustments
- **Max Change Percentage**: Maximum allowed percentage change (default: 20%)
- **Cooldown Days**: Days to wait between changes (default: 7)
- **Admin Approval Required**: Whether admin approval is needed (default: true)
- **Safety Thresholds**: Min/max retention days and storage impact limits

## Technical Implementation

### New Services

#### RetentionMonitorService (`pkg/email/retention_monitor.go`)
```go
type RetentionMonitorService struct {
    db           *sql.DB
    metricsCache map[string]*RealtimeMetrics
    cacheMutex   sync.RWMutex
    eventChannel chan *RetentionEvent
}

// Key methods:
- Start(ctx context.Context) - Begin real-time monitoring
- RecordPolicyEvaluation(...) - Log policy evaluation events
- RecordEmailDeletion(...) - Log email deletion events
- RecordEmailArchival(...) - Log email archival events
- GetRealtimeMetrics(ctx, metricType, metricKey) - Retrieve live metrics
- GetUnprocessedEvents(ctx, limit) - Get events for WebSocket/SSE
```

#### AdaptiveRetentionEngine (`pkg/email/adaptive_retention.go`)
```go
type AdaptiveRetentionEngine struct {
    db                    *sql.DB
    monitorService        *RetentionMonitorService
    policyEngine          *RetentionPolicyEngine
    adaptiveEnabled       bool
    maxChangePercentage   float64
    cooldownDays          int
    requiresAdminApproval bool
}

// Key methods:
- Start(ctx context.Context) - Begin adaptive analysis
- AnalyzePolicyPerformance(ctx, policyID) - Analyze policy effectiveness
- GenerateAdaptiveRecommendations(ctx) - Create adaptive recommendations
- ApplyAdaptiveChange(ctx, changeID, appliedBy, preview) - Apply changes
- EnableAdaptivePolicy(ctx, policyID, config) - Enable adaptive features
- DisableAdaptivePolicy(ctx, policyID) - Disable adaptive features
```

### Database Schema Enhancements

#### New Tables
```sql
-- Real-time retention metrics cache
CREATE TABLE realtime_retention_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_type TEXT NOT NULL, -- "user", "domain", "policy", "global"
    metric_key TEXT NOT NULL, -- user_id, domain, policy_id, or "global"
    active_emails_count INTEGER DEFAULT 0,
    archived_emails_count INTEGER DEFAULT 0,
    deleted_emails_count INTEGER DEFAULT 0,
    total_storage_bytes INTEGER DEFAULT 0,
    compressed_storage_bytes INTEGER DEFAULT 0,
    policy_evaluations_count INTEGER DEFAULT 0,
    policy_matches_count INTEGER DEFAULT 0,
    policy_applications_count INTEGER DEFAULT 0,
    avg_match_score REAL DEFAULT 0.0,
    avg_impact_score REAL DEFAULT 0.0,
    archival_operations_count INTEGER DEFAULT 0,
    avg_archival_duration_ms INTEGER DEFAULT 0,
    archival_success_rate REAL DEFAULT 0.0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(metric_type, metric_key)
);

-- Adaptive policy changes tracking
CREATE TABLE adaptive_policy_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    policy_id INTEGER NOT NULL,
    change_type TEXT NOT NULL, -- "retention_days", "archive_instead", "archive_retention_days"
    old_value TEXT NOT NULL,
    new_value TEXT NOT NULL,
    change_reason TEXT NOT NULL,
    change_percentage REAL DEFAULT 0.0,
    expected_storage_savings_bytes INTEGER DEFAULT 0,
    expected_archival_load_impact REAL DEFAULT 0.0,
    risk_assessment TEXT DEFAULT "low",
    cooldown_until TIMESTAMP,
    requires_admin_approval BOOLEAN DEFAULT 0,
    status TEXT DEFAULT "pending", -- "pending", "approved", "applied", "rejected"
    applied_at TIMESTAMP,
    applied_by TEXT,
    applied_result TEXT,
    FOREIGN KEY (policy_id) REFERENCES retention_policies(id)
);

-- Real-time event stream
CREATE TABLE retention_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL, -- "policy_evaluation", "email_deletion", "email_archival", "policy_change"
    event_data TEXT NOT NULL, -- JSON event data
    user_id TEXT,
    policy_id INTEGER,
    event_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN DEFAULT 0,
    FOREIGN KEY (policy_id) REFERENCES retention_policies(id)
);

-- Adaptive policy configuration
CREATE TABLE adaptive_policy_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    policy_id INTEGER NOT NULL,
    adaptive_enabled BOOLEAN DEFAULT 0,
    max_change_percentage REAL DEFAULT 20.0,
    cooldown_days INTEGER DEFAULT 7,
    requires_admin_approval BOOLEAN DEFAULT 1,
    min_retention_days INTEGER DEFAULT 1,
    max_retention_days INTEGER DEFAULT 365,
    min_archive_retention_days INTEGER DEFAULT 30,
    max_archive_retention_days INTEGER DEFAULT 2555,
    max_storage_impact_bytes INTEGER DEFAULT 1073741824,
    max_archival_load_impact REAL DEFAULT 0.5,
    UNIQUE(policy_id)
);
```

### API Endpoints

#### Real-Time Metrics Endpoints
- `GET /api/admin/email/retention-realtime` - Get global real-time metrics
- `GET /api/admin/email/retention-realtime?metric_type=user&metric_key=user@example.com` - Get user metrics
- `GET /api/admin/email/retention-realtime?metric_type=domain&metric_key=example.com` - Get domain metrics
- `GET /api/admin/email/retention-realtime?metric_type=policy&metric_key=1` - Get policy metrics

#### Adaptive Policy Endpoints
- `GET /api/admin/email/adaptive-policy-changes` - Get adaptive policy changes with filtering
- `POST /api/admin/email/adaptive-policy/enable` - Enable adaptive policy with configuration
- `POST /api/admin/email/adaptive-policy/disable` - Disable adaptive policy
- `POST /api/admin/email/adaptive-policy/apply` - Apply adaptive change with preview mode
- `GET /api/admin/email/policy-performance?policy_id=1` - Analyze policy performance
- `POST /api/admin/email/adaptive-policy/generate-recommendations` - Generate recommendations

#### Query Parameters
- **Filtering**: `metric_type`, `metric_key`, `policy_id`, `status`, `change_type`
- **Pagination**: `limit`, `offset`
- **Time Range**: `start_date`, `end_date` (YYYY-MM-DD format)

### Background Processing

#### Real-Time Event Processing
- **Event Channel**: Buffered channel for real-time event processing
- **Metrics Cache**: In-memory cache with periodic database synchronization
- **Event Types**: Policy evaluation, email deletion, email archival, policy change
- **Processing Frequency**: Real-time with 30-second cache refresh intervals

#### Adaptive Analysis
- **Periodic Analysis**: Runs every hour to generate adaptive recommendations
- **Performance Analysis**: Analyzes policy effectiveness and archival load
- **Recommendation Generation**: Creates adaptive changes based on performance metrics
- **Safety Validation**: Ensures changes meet safety thresholds and cooldown requirements

## Configuration Options

### Environment Variables
```bash
# Enable/disable features
ENABLE_REALTIME_RETENTION_MONITORING=true
ENABLE_ADAPTIVE_POLICY_ENFORCEMENT=false

# Adaptive policy controls
ADAPTIVE_POLICY_COOLDOWN_DAYS=7
ADAPTIVE_POLICY_MAX_CHANGE_PERCENT=20
```

### Policy-Level Configuration
```json
{
  "policy_id": 1,
  "adaptive_enabled": true,
  "max_change_percentage": 15.0,
  "cooldown_days": 5,
  "requires_admin_approval": true,
  "min_retention_days": 1,
  "max_retention_days": 365,
  "min_archive_retention_days": 30,
  "max_archive_retention_days": 2555,
  "max_storage_impact_bytes": 1073741824,
  "max_archival_load_impact": 0.5
}
```

## Testing & Validation

### Test Script
- **Comprehensive Testing**: `scripts/test_realtime_adaptive_features.ps1`
- **Endpoint Validation**: Tests all new API endpoints
- **Data Validation**: Validates input validation and error handling
- **Performance Testing**: Tests concurrent request handling
- **Integration Testing**: Tests adaptive change workflow

### Test Coverage
- **Real-Time Metrics**: Global, user, domain, and policy metrics
- **Adaptive Policy Changes**: Change history, filtering, and pagination
- **Policy Control**: Enable/disable adaptive features
- **Policy Performance**: Performance analysis and metrics
- **Recommendation Generation**: Adaptive recommendation creation
- **Change Application**: Preview and apply adaptive changes
- **Data Validation**: Invalid input handling
- **Performance**: Concurrent request handling

## Security Considerations

### Privacy Protection
- **Metadata Only**: Real-time monitoring analyzes metadata, not email content
- **User Consent**: Adaptive features require explicit admin opt-in per policy
- **Audit Logging**: All adaptive changes are logged with full audit trail
- **Access Control**: All endpoints require admin authentication

### Safety Controls
- **Cooldown Periods**: Prevents rapid policy oscillations
- **Maximum Changes**: Limits percentage changes to prevent drastic adjustments
- **Admin Approval**: Requires admin approval for adaptive changes by default
- **Impact Analysis**: Calculates and validates expected impact before changes
- **Rollback Capability**: Changes can be previewed before application

## Performance Optimization

### Caching Strategy
- **In-Memory Cache**: Real-time metrics cached in memory for instant access
- **Periodic Sync**: Cache synchronized with database every 30 seconds
- **Lazy Loading**: Metrics loaded on-demand for new scopes
- **Cleanup Triggers**: Automatic cleanup of old events and changes

### Database Optimization
- **Indexed Queries**: Optimized indexes for real-time queries
- **Batch Operations**: Efficient batch processing for event handling
- **Connection Pooling**: Reuses database connections for better performance
- **Query Optimization**: Optimized queries for large datasets

## Monitoring & Observability

### Metrics Collection
- **Real-Time Metrics**: Live retention statistics and performance indicators
- **Policy Performance**: Policy effectiveness and archival load analysis
- **Adaptive Changes**: Tracking of adaptive policy adjustments
- **Event Processing**: Event processing rates and success metrics

### Logging
- **Event Logging**: Comprehensive logging of all retention events
- **Adaptive Changes**: Detailed logging of adaptive policy changes
- **Performance Metrics**: Logging of performance analysis results
- **Error Handling**: Detailed error logging with context

## Future Enhancements

### Planned Features
- **WebSocket Integration**: Real-time event streaming to Admin UI
- **Machine Learning**: Advanced ML-based policy optimization
- **Predictive Analytics**: Predictive policy performance analysis
- **Advanced Visualization**: Interactive dashboards for real-time metrics
- **Multi-Tenant Support**: Enhanced support for multi-tenant deployments

### Performance Improvements
- **Redis Integration**: Redis-based caching for better performance
- **Event Streaming**: Kafka/RabbitMQ integration for event processing
- **Horizontal Scaling**: Support for horizontal scaling of monitoring services
- **Advanced Analytics**: More sophisticated analytics and reporting

## Migration Guide

### Database Migration
```sql
-- Run the migration script
-- schema/migrate_add_realtime_monitoring_adaptive_policies.sql
```

### Configuration Updates
1. Add new environment variables to `.env`
2. Configure adaptive policy settings per policy
3. Enable real-time monitoring features
4. Test adaptive policy enforcement in staging

### Service Updates
1. Update API server with new endpoints
2. Initialize new monitoring and adaptive services
3. Configure event processing and caching
4. Test all new functionality

## Conclusion

Micro-Iteration 4.28 successfully implements real-time retention monitoring and adaptive policy enforcement, providing administrators with live insights into retention operations and intelligent policy optimization. The system maintains strong safety controls while enabling automated policy adjustments based on real-time performance metrics.

The implementation includes comprehensive testing, documentation, and configuration options, ensuring a production-ready system that can be safely deployed and monitored. The modular design allows for future enhancements and integrations while maintaining backward compatibility with existing features.



