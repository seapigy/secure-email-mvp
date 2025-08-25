package monitoring

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/models"
)

// MonitoringRepository defines the interface for monitoring data operations
type MonitoringRepository interface {
	// Event operations
	LogEvent(event *models.MonitoringEvent) error
	GetEventsByType(eventType string, limit int) ([]*models.MonitoringEvent, error)
	GetEventsBySource(source string, limit int) ([]*models.MonitoringEvent, error)
	GetEventsByDateRange(startDate, endDate time.Time) ([]*models.MonitoringEvent, error)
	GetEventsBySeverity(severity string, limit int) ([]*models.MonitoringEvent, error)
	
	// Metrics operations
	SaveMetricsSummary(summary *models.MetricsSummary) error
	GetMetricsSummary(metricName, timeBucket string, startDate, endDate time.Time) ([]*models.MetricsSummary, error)
	GetRealTimeMetrics() (*models.RealTimeMetrics, error)
	
	// Stream operations
	GetRecentEvents(limit int) ([]*models.MonitoringEvent, error)
	GetEventCountByType(eventType string, duration time.Duration) (int64, error)
	GetErrorRate(duration time.Duration) (float64, error)
	GetAverageLatency(duration time.Duration) (float64, error)
}

// SQLiteMonitoringRepository implements MonitoringRepository using SQLite
type SQLiteMonitoringRepository struct {
	db *sql.DB
}

// NewSQLiteMonitoringRepository creates a new SQLite-based monitoring repository
func NewSQLiteMonitoringRepository(db *sql.DB) *SQLiteMonitoringRepository {
	return &SQLiteMonitoringRepository{db: db}
}

// LogEvent saves a monitoring event to the database
func (r *SQLiteMonitoringRepository) LogEvent(event *models.MonitoringEvent) error {
	log.Printf("Logging monitoring event: %s from %s", event.EventType, *event.Source)
	
	query := `
		INSERT INTO monitoring_events (
			event_type, event_subtype, timestamp, metadata, severity, source,
			user_id, session_id, ip_address, user_agent, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query,
		event.EventType, event.EventSubtype, event.Timestamp, event.Metadata,
		event.Severity, event.Source, event.UserID, event.SessionID,
		event.IPAddress, event.UserAgent, event.CreatedAt,
	)
	
	if err != nil {
		log.Printf("Failed to log monitoring event: %v", err)
		return fmt.Errorf("failed to log monitoring event: %w", err)
	}
	
	return nil
}

// GetEventsByType retrieves events by type with limit
func (r *SQLiteMonitoringRepository) GetEventsByType(eventType string, limit int) ([]*models.MonitoringEvent, error) {
	query := `
		SELECT id, event_type, event_subtype, timestamp, metadata, severity, source,
		       user_id, session_id, ip_address, user_agent, created_at
		FROM monitoring_events 
		WHERE event_type = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by type: %w", err)
	}
	defer rows.Close()
	
	var events []*models.MonitoringEvent
	for rows.Next() {
		var event models.MonitoringEvent
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EventSubtype, &event.Timestamp,
			&event.Metadata, &event.Severity, &event.Source, &event.UserID,
			&event.SessionID, &event.IPAddress, &event.UserAgent, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}
	
	return events, nil
}

// GetEventsBySource retrieves events by source with limit
func (r *SQLiteMonitoringRepository) GetEventsBySource(source string, limit int) ([]*models.MonitoringEvent, error) {
	query := `
		SELECT id, event_type, event_subtype, timestamp, metadata, severity, source,
		       user_id, session_id, ip_address, user_agent, created_at
		FROM monitoring_events 
		WHERE source = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, source, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by source: %w", err)
	}
	defer rows.Close()
	
	var events []*models.MonitoringEvent
	for rows.Next() {
		var event models.MonitoringEvent
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EventSubtype, &event.Timestamp,
			&event.Metadata, &event.Severity, &event.Source, &event.UserID,
			&event.SessionID, &event.IPAddress, &event.UserAgent, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}
	
	return events, nil
}

// GetEventsByDateRange retrieves events within a date range
func (r *SQLiteMonitoringRepository) GetEventsByDateRange(startDate, endDate time.Time) ([]*models.MonitoringEvent, error) {
	query := `
		SELECT id, event_type, event_subtype, timestamp, metadata, severity, source,
		       user_id, session_id, ip_address, user_agent, created_at
		FROM monitoring_events 
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
	`
	
	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by date range: %w", err)
	}
	defer rows.Close()
	
	var events []*models.MonitoringEvent
	for rows.Next() {
		var event models.MonitoringEvent
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EventSubtype, &event.Timestamp,
			&event.Metadata, &event.Severity, &event.Source, &event.UserID,
			&event.SessionID, &event.IPAddress, &event.UserAgent, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}
	
	return events, nil
}

// GetEventsBySeverity retrieves events by severity with limit
func (r *SQLiteMonitoringRepository) GetEventsBySeverity(severity string, limit int) ([]*models.MonitoringEvent, error) {
	query := `
		SELECT id, event_type, event_subtype, timestamp, metadata, severity, source,
		       user_id, session_id, ip_address, user_agent, created_at
		FROM monitoring_events 
		WHERE severity = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, severity, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by severity: %w", err)
	}
	defer rows.Close()
	
	var events []*models.MonitoringEvent
	for rows.Next() {
		var event models.MonitoringEvent
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EventSubtype, &event.Timestamp,
			&event.Metadata, &event.Severity, &event.Source, &event.UserID,
			&event.SessionID, &event.IPAddress, &event.UserAgent, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}
	
	return events, nil
}

// SaveMetricsSummary saves a metrics summary to the database
func (r *SQLiteMonitoringRepository) SaveMetricsSummary(summary *models.MetricsSummary) error {
	query := `
		INSERT OR REPLACE INTO monitoring_metrics_summary (
			metric_name, metric_value, metric_unit, time_bucket, bucket_start, bucket_end, source, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query,
		summary.MetricName, summary.MetricValue, summary.MetricUnit,
		summary.TimeBucket, summary.BucketStart, summary.BucketEnd,
		summary.Source, summary.CreatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to save metrics summary: %w", err)
	}
	
	return nil
}

// GetMetricsSummary retrieves metrics summary data
func (r *SQLiteMonitoringRepository) GetMetricsSummary(metricName, timeBucket string, startDate, endDate time.Time) ([]*models.MetricsSummary, error) {
	query := `
		SELECT id, metric_name, metric_value, metric_unit, time_bucket, bucket_start, bucket_end, source, created_at
		FROM monitoring_metrics_summary 
		WHERE metric_name = ? AND time_bucket = ? AND bucket_start BETWEEN ? AND ?
		ORDER BY bucket_start ASC
	`
	
	rows, err := r.db.Query(query, metricName, timeBucket, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics summary: %w", err)
	}
	defer rows.Close()
	
	var summaries []*models.MetricsSummary
	for rows.Next() {
		var summary models.MetricsSummary
		err := rows.Scan(
			&summary.ID, &summary.MetricName, &summary.MetricValue, &summary.MetricUnit,
			&summary.TimeBucket, &summary.BucketStart, &summary.BucketEnd,
			&summary.Source, &summary.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metrics summary: %w", err)
		}
		summaries = append(summaries, &summary)
	}
	
	return summaries, nil
}

// GetRealTimeMetrics retrieves current real-time metrics
func (r *SQLiteMonitoringRepository) GetRealTimeMetrics() (*models.RealTimeMetrics, error) {
	// Get request count for last minute
	requestCount, err := r.GetEventCountByType("api.request", time.Minute)
	if err != nil {
		return nil, err
	}
	
	// Get error rate for last minute
	errorRate, err := r.GetErrorRate(time.Minute)
	if err != nil {
		return nil, err
	}
	
	// Get average latency for last minute
	avgLatency, err := r.GetAverageLatency(time.Minute)
	if err != nil {
		return nil, err
	}
	
	// Get DLP scans for last minute
	dlpScans, err := r.GetEventCountByType("dlp.scan", time.Minute)
	if err != nil {
		return nil, err
	}
	
	// Get watermarking operations for last minute
	watermarkingOps, err := r.GetEventCountByType("watermarking.apply", time.Minute)
	if err != nil {
		return nil, err
	}
	
	// Get security alerts for last minute
	securityAlerts, err := r.GetEventCountByType("security.alert", time.Minute)
	if err != nil {
		return nil, err
	}
	
	// Get source breakdown
	sourceBreakdown, err := r.getSourceBreakdown(time.Minute)
	if err != nil {
		return nil, err
	}
	
	// Get error breakdown
	errorBreakdown, err := r.getErrorBreakdown(time.Minute)
	if err != nil {
		return nil, err
	}
	
	// Estimate active sessions (simplified - count unique session IDs in last 5 minutes)
	activeSessions, err := r.getActiveSessions(5 * time.Minute)
	if err != nil {
		return nil, err
	}
	
	return &models.RealTimeMetrics{
		RequestCount:     requestCount,
		ErrorRate:        errorRate,
		AverageLatency:   avgLatency,
		ActiveSessions:   activeSessions,
		DLPScans:         dlpScans,
		WatermarkingOps:  watermarkingOps,
		SecurityAlerts:   securityAlerts,
		LastUpdated:      time.Now(),
		SourceBreakdown:  sourceBreakdown,
		ErrorBreakdown:   errorBreakdown,
	}, nil
}

// GetRecentEvents retrieves recent events for streaming
func (r *SQLiteMonitoringRepository) GetRecentEvents(limit int) ([]*models.MonitoringEvent, error) {
	query := `
		SELECT id, event_type, event_subtype, timestamp, metadata, severity, source,
		       user_id, session_id, ip_address, user_agent, created_at
		FROM monitoring_events 
		ORDER BY timestamp DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent events: %w", err)
	}
	defer rows.Close()
	
	var events []*models.MonitoringEvent
	for rows.Next() {
		var event models.MonitoringEvent
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EventSubtype, &event.Timestamp,
			&event.Metadata, &event.Severity, &event.Source, &event.UserID,
			&event.SessionID, &event.IPAddress, &event.UserAgent, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}
	
	return events, nil
}

// GetEventCountByType gets count of events by type within duration
func (r *SQLiteMonitoringRepository) GetEventCountByType(eventType string, duration time.Duration) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM monitoring_events 
		WHERE event_type = ? AND timestamp >= datetime('now', '-' || ? || ' seconds')
	`
	
	var count int64
	err := r.db.QueryRow(query, eventType, int(duration.Seconds())).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get event count by type: %w", err)
	}
	
	return count, nil
}

// GetErrorRate gets error rate within duration
func (r *SQLiteMonitoringRepository) GetErrorRate(duration time.Duration) (float64, error) {
	query := `
		SELECT 
			CASE 
				WHEN COUNT(*) = 0 THEN 0.0
				ELSE (COUNT(CASE WHEN severity IN ('error', 'critical') THEN 1 END) * 100.0 / COUNT(*))
			END as error_rate
		FROM monitoring_events 
		WHERE timestamp >= datetime('now', '-' || ? || ' seconds')
	`
	
	var errorRate float64
	err := r.db.QueryRow(query, int(duration.Seconds())).Scan(&errorRate)
	if err != nil {
		return 0, fmt.Errorf("failed to get error rate: %w", err)
	}
	
	return errorRate, nil
}

// GetAverageLatency gets average latency within duration
func (r *SQLiteMonitoringRepository) GetAverageLatency(duration time.Duration) (float64, error) {
	query := `
		SELECT AVG(CAST(json_extract(metadata, '$.latency_ms') AS REAL))
		FROM monitoring_events 
		WHERE event_type = 'api.request' 
			AND timestamp >= datetime('now', '-' || ? || ' seconds')
			AND json_extract(metadata, '$.latency_ms') IS NOT NULL
	`
	
	var avgLatency sql.NullFloat64
	err := r.db.QueryRow(query, int(duration.Seconds())).Scan(&avgLatency)
	if err != nil {
		return 0, fmt.Errorf("failed to get average latency: %w", err)
	}
	
	if avgLatency.Valid {
		return avgLatency.Float64, nil
	}
	return 0, nil
}

// getSourceBreakdown gets breakdown of events by source
func (r *SQLiteMonitoringRepository) getSourceBreakdown(duration time.Duration) (map[string]int64, error) {
	query := `
		SELECT source, COUNT(*) as count
		FROM monitoring_events 
		WHERE timestamp >= datetime('now', '-' || ? || ' seconds')
			AND source IS NOT NULL
		GROUP BY source
	`
	
	rows, err := r.db.Query(query, int(duration.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to get source breakdown: %w", err)
	}
	defer rows.Close()
	
	breakdown := make(map[string]int64)
	for rows.Next() {
		var source string
		var count int64
		err := rows.Scan(&source, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source breakdown: %w", err)
		}
		breakdown[source] = count
	}
	
	return breakdown, nil
}

// getErrorBreakdown gets breakdown of errors by type
func (r *SQLiteMonitoringRepository) getErrorBreakdown(duration time.Duration) (map[string]int64, error) {
	query := `
		SELECT event_type, COUNT(*) as count
		FROM monitoring_events 
		WHERE severity IN ('error', 'critical')
			AND timestamp >= datetime('now', '-' || ? || ' seconds')
		GROUP BY event_type
	`
	
	rows, err := r.db.Query(query, int(duration.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to get error breakdown: %w", err)
	}
	defer rows.Close()
	
	breakdown := make(map[string]int64)
	for rows.Next() {
		var eventType string
		var count int64
		err := rows.Scan(&eventType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan error breakdown: %w", err)
		}
		breakdown[eventType] = count
	}
	
	return breakdown, nil
}

// getActiveSessions gets count of active sessions
func (r *SQLiteMonitoringRepository) getActiveSessions(duration time.Duration) (int64, error) {
	query := `
		SELECT COUNT(DISTINCT session_id)
		FROM monitoring_events 
		WHERE session_id IS NOT NULL
			AND timestamp >= datetime('now', '-' || ? || ' seconds')
	`
	
	var count int64
	err := r.db.QueryRow(query, int(duration.Seconds())).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active sessions: %w", err)
	}
	
	return count, nil
}
