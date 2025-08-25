# Iteration 9: Real-Time Monitoring & Dashboards

## Overview

Iteration 9 implements a comprehensive real-time monitoring and dashboard system for the Secure Email MVP. This system provides live metrics, system health monitoring, and operational insights through both backend APIs and frontend dashboards.

## Architecture

### Backend Components

#### 1. Monitoring Service (`pkg/securelinks/monitoring/service.go`)

The core monitoring service that:
- Collects and aggregates metrics in real-time
- Manages in-memory counters with thread-safe operations
- Handles Server-Sent Events (SSE) for real-time streaming
- Implements retention policies and cleanup
- Provides system health assessment

**Key Features:**
- **In-Memory Counters**: Fast access to current metrics
- **Database Persistence**: Long-term storage of monitoring events
- **SSE Streaming**: Real-time updates to frontend dashboards
- **Background Aggregation**: Automatic metrics calculation
- **Retention Management**: Automatic cleanup of old data

#### 2. Monitoring Repository (`pkg/securelinks/monitoring/repository.go`)

Database layer for monitoring data:
- Stores monitoring events in `monitoring_events` table
- Calculates real-time metrics from event data
- Implements retention policies
- Provides efficient querying for dashboards

**Database Schema:**
```sql
-- monitoring_events table
CREATE TABLE monitoring_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    event_subtype TEXT,
    timestamp DATETIME NOT NULL,
    metadata TEXT, -- JSON metadata
    severity TEXT NOT NULL,
    source TEXT NOT NULL,
    user_id TEXT,
    session_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- monitoring_metrics_summary table
CREATE TABLE monitoring_metrics_summary (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    metric_unit TEXT,
    timestamp DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. Monitoring Models (`pkg/models/monitoring.go`)

Go structs for monitoring data:
- `MonitoringEvent`: Individual monitoring events
- `MetricsSummary`: Aggregated metrics
- `RealTimeMetrics`: Current system metrics
- `MonitoringConfig`: Configuration settings
- `StreamEvent`: SSE event structure

#### 4. HTTP Endpoints

**GET /api/metrics**
- Returns current system metrics
- Response includes request count, error rate, latency, active sessions
- Used by dashboard for initial load and polling fallback

**GET /api/metrics/stream**
- Server-Sent Events endpoint for real-time updates
- Streams monitoring events as they occur
- Supports multiple concurrent subscribers

**GET /api/metrics/health**
- Returns system health status
- Includes overall status, message, and key metrics
- Used for health checks and alerting

#### 5. Middleware Integration

**Monitoring Middleware** (`cmd/api/main.go`)
- Automatically logs all API requests
- Captures request details (path, method, status, latency)
- Extracts user and session information
- Integrates with monitoring service

### Frontend Components

#### 1. Monitoring Dashboard (`src/components/dashboard/MonitoringDashboard.tsx`)

Comprehensive dashboard with:
- **Real-time Metrics Display**: Current system statistics
- **System Health Status**: Overall system health indicator
- **Tabbed Interface**: Overview, Security, Performance, Analytics
- **Live Updates**: SSE integration for real-time data
- **Responsive Design**: Works on all screen sizes

**Key Sections:**
- **Overview**: Request trends, error rates, general metrics
- **Security**: DLP scans, watermarking operations, security alerts
- **Performance**: Response latency, system throughput
- **Analytics**: Traffic breakdown, error analysis

#### 2. Metrics Chart Component (`src/components/dashboard/MetricsChart.tsx`)

Reusable chart component using Recharts:
- **Multiple Chart Types**: Line, bar, area, pie charts
- **Real-time Data**: Updates automatically with new data
- **Sample Data**: Fallback data for development
- **Customizable**: Colors, labels, grid options

#### 3. Monitoring Store (`src/stores/monitoringStore.ts`)

Zustand store for monitoring state:
- **SSE Connection Management**: Handles real-time connections
- **Event Subscription**: Pub/sub pattern for components
- **State Management**: Centralized monitoring state
- **Error Handling**: Connection failures and reconnection

## API Reference

### Metrics Endpoint

**GET /api/metrics**

Returns current system metrics in JSON format.

**Response:**
```json
{
  "status": "success",
  "timestamp": "2024-01-15T10:30:00Z",
  "metrics": {
    "request_count": 1250,
    "error_rate": 2.5,
    "average_latency": 45.2,
    "active_sessions": 23,
    "dlp_scans": 89,
    "watermarking_ops": 34,
    "security_alerts": 2,
    "last_updated": "2024-01-15T10:30:00Z",
    "source_breakdown": {
      "api": 1200,
      "web": 50
    },
    "error_breakdown": {
      "400": 15,
      "500": 5
    }
  }
}
```

### Health Endpoint

**GET /api/metrics/health**

Returns system health status.

**Response:**
```json
{
  "status": "success",
  "timestamp": "2024-01-15T10:30:00Z",
  "health": {
    "status": "healthy",
    "message": "All systems operational",
    "timestamp": "2024-01-15T10:30:00Z",
    "error_rate": 2.5,
    "avg_latency": 45.2,
    "request_count": 1250,
    "active_sessions": 23
  }
}
```

### Stream Endpoint

**GET /api/metrics/stream**

Server-Sent Events endpoint for real-time updates.

**Event Format:**
```
event: metrics_update
data: {"type": "metrics_update", "data": {...}, "timestamp": "..."}

event: health_update
data: {"type": "health_update", "data": {...}, "timestamp": "..."}
```

## Service Integration

### DLP Service Integration

The DLP service logs monitoring events for:
- Content scans (success/failure)
- Violation detection
- Processing time
- Action taken (allow/warn/block)

**Event Example:**
```go
event := models.CreateDLPScanEvent(
    req.ContentType,
    scanResult,
    processingTimeMs,
)
monitoringService.LogEvent(event)
```

### Watermarking Service Integration

The watermarking service logs monitoring events for:
- Watermark applications
- Processing time
- Success/failure status
- Watermark type and content type

**Event Example:**
```go
event := models.CreateWatermarkingEvent(
    req.WatermarkType,
    req.ContentType,
    processingTimeMs,
)
monitoringService.LogEvent(event)
```

## Frontend Usage

### Setting Up the Dashboard

1. **Import Components:**
```typescript
import MonitoringDashboard from '@/components/dashboard/MonitoringDashboard';
import { useMonitoringStore } from '@/stores/monitoringStore';
```

2. **Add Route:**
```typescript
<Route path="/monitoring" element={<MonitoringDashboard />} />
```

3. **Add Navigation:**
```typescript
{
  name: 'Monitoring',
  href: '/monitoring',
  icon: ChartBarIcon,
}
```

### Using the Monitoring Store

```typescript
// Subscribe to real-time updates
const { subscribeToStream, unsubscribeFromStream } = useMonitoringStore();

useEffect(() => {
  const clientId = `dashboard_${Date.now()}`;
  
  const handleEvent = (event) => {
    if (event.type === 'metrics_update') {
      setMetrics(event.data);
    }
  };
  
  subscribeToStream(clientId, handleEvent);
  
  return () => unsubscribeFromStream(clientId);
}, []);
```

### Custom Charts

```typescript
import MetricsChart from '@/components/dashboard/MetricsChart';

<MetricsChart 
  title="Request Trends"
  data={requestData}
  type="line"
  height={300}
  colors={['#3b82f6', '#ef4444']}
/>
```

## Configuration

### Monitoring Configuration

```go
monitoringConfig := &models.MonitoringConfig{
    RetentionDays:           30,    // Keep events for 30 days
    Enabled:                 true,  // Enable monitoring
    SampleRate:              1.0,   // Sample 100% of events
    AlertThresholdErrorRate: 0.05,  // Alert if error rate > 5%
    AlertThresholdLatencyMs: 1000,  // Alert if latency > 1s
}
```

### Database Configuration

The monitoring system uses the existing SQLite database with new tables:
- `monitoring_events`: Individual monitoring events
- `monitoring_metrics_summary`: Aggregated metrics
- Indexes for efficient querying
- Retention triggers for automatic cleanup

## Testing

### Integration Tests

Run the comprehensive test suite:
```powershell
.\tests\test_iteration9_monitoring.ps1
```

**Test Coverage:**
- Basic endpoint functionality
- JSON response validation
- SSE connectivity
- Concurrent request handling
- Performance testing
- Error handling
- CORS headers

### Manual Testing

1. **Start the backend:**
```bash
go run ./cmd/api
```

2. **Access the dashboard:**
```
http://localhost:8080/monitoring
```

3. **Test real-time updates:**
- Open browser dev tools
- Monitor SSE connection
- Make API requests
- Watch dashboard updates

## Performance Considerations

### Backend Performance

- **In-Memory Counters**: Fast access to current metrics
- **Database Indexing**: Efficient querying of historical data
- **Background Processing**: Non-blocking metrics aggregation
- **Connection Pooling**: Efficient SSE connection management

### Frontend Performance

- **SSE Streaming**: Real-time updates without polling
- **Component Optimization**: Efficient re-rendering
- **Chart Performance**: Optimized chart rendering
- **Memory Management**: Automatic cleanup of old data

## Security

### Data Protection

- **Event Sanitization**: All events are sanitized before storage
- **Access Control**: Monitoring endpoints require authentication
- **Data Retention**: Automatic cleanup of old monitoring data
- **Audit Trail**: All monitoring events are logged

### Privacy

- **User Anonymization**: User IDs are optional and can be anonymized
- **IP Address Handling**: IP addresses are stored but can be masked
- **Session Tracking**: Session IDs are used for correlation only
- **Data Minimization**: Only necessary data is collected

## Monitoring and Alerting

### Built-in Alerts

The system automatically detects:
- High error rates (>5%)
- High latency (>1s)
- System health degradation
- Connection failures

### Custom Alerts

Extend the monitoring system with custom alerts:
```go
// Create custom alert
event := models.CreateSecurityAlertEvent("custom_alert", metadata)
monitoringService.LogEvent(event)
```

## Troubleshooting

### Common Issues

1. **SSE Connection Fails**
   - Check CORS configuration
   - Verify endpoint is accessible
   - Check browser console for errors

2. **No Real-time Updates**
   - Verify SSE connection is established
   - Check monitoring service is running
   - Verify events are being logged

3. **High Memory Usage**
   - Check retention policy settings
   - Monitor event volume
   - Adjust cleanup frequency

### Debug Mode

Enable debug logging:
```go
log.SetLevel(log.DebugLevel)
```

### Database Queries

Check monitoring data directly:
```sql
-- View recent events
SELECT * FROM monitoring_events ORDER BY timestamp DESC LIMIT 10;

-- Check metrics summary
SELECT * FROM monitoring_metrics_summary ORDER BY timestamp DESC LIMIT 5;
```

## Future Enhancements

### Planned Features

1. **Advanced Analytics**
   - Machine learning for anomaly detection
   - Predictive analytics
   - Custom dashboards

2. **Enhanced Alerting**
   - Email/SMS notifications
   - Webhook integrations
   - Escalation policies

3. **Performance Optimization**
   - Redis caching layer
   - Database partitioning
   - CDN integration

4. **Compliance Features**
   - GDPR compliance tools
   - Data export capabilities
   - Audit log enhancements

## Conclusion

The real-time monitoring system provides comprehensive visibility into the Secure Email MVP's operations. With live dashboards, automated metrics collection, and seamless integration with existing services, it enables proactive monitoring and rapid issue resolution.

The system is designed to scale with the application and can be extended with additional monitoring capabilities as needed.
