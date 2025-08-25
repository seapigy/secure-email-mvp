package models

import (
	"encoding/json"
	"time"
)

// MonitoringEvent represents a system monitoring event
type MonitoringEvent struct {
	ID          int64           `json:"id" db:"id"`
	EventType   string          `json:"event_type" db:"event_type"`
	EventSubtype *string        `json:"event_subtype,omitempty" db:"event_subtype"`
	Timestamp   time.Time       `json:"timestamp" db:"timestamp"`
	Metadata    *string         `json:"metadata,omitempty" db:"metadata"` // JSON string
	Severity    string          `json:"severity" db:"severity"`
	Source      *string         `json:"source,omitempty" db:"source"`
	UserID      *string         `json:"user_id,omitempty" db:"user_id"`
	SessionID   *string         `json:"session_id,omitempty" db:"session_id"`
	IPAddress   *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   *string         `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// MetricsSummary represents aggregated metrics data
type MetricsSummary struct {
	ID          int64     `json:"id" db:"id"`
	MetricName  string    `json:"metric_name" db:"metric_name"`
	MetricValue float64   `json:"metric_value" db:"metric_value"`
	MetricUnit  *string   `json:"metric_unit,omitempty" db:"metric_unit"`
	TimeBucket  string    `json:"time_bucket" db:"time_bucket"`
	BucketStart time.Time `json:"bucket_start" db:"bucket_start"`
	BucketEnd   time.Time `json:"bucket_end" db:"bucket_end"`
	Source      *string   `json:"source,omitempty" db:"source"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// RealTimeMetrics represents current system metrics
type RealTimeMetrics struct {
	RequestCount     int64             `json:"request_count"`
	ErrorRate        float64           `json:"error_rate"`
	AverageLatency   float64           `json:"average_latency"`
	ActiveSessions   int64             `json:"active_sessions"`
	DLPScans         int64             `json:"dlp_scans"`
	WatermarkingOps  int64             `json:"watermarking_ops"`
	SecurityAlerts   int64             `json:"security_alerts"`
	LastUpdated      time.Time         `json:"last_updated"`
	SourceBreakdown  map[string]int64  `json:"source_breakdown"`
	ErrorBreakdown   map[string]int64  `json:"error_breakdown"`
}

// MonitoringConfig represents monitoring system configuration
type MonitoringConfig struct {
	RetentionDays           int     `json:"retention_days"`
	Enabled                 bool    `json:"enabled"`
	SampleRate              float64 `json:"sample_rate"`
	AlertThresholdErrorRate float64 `json:"alert_threshold_error_rate"`
	AlertThresholdLatencyMs int     `json:"alert_threshold_latency_ms"`
}

// StreamEvent represents a real-time streaming event
type StreamEvent struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// EventMetadata represents common event metadata
type EventMetadata struct {
	Endpoint        *string  `json:"endpoint,omitempty"`
	Method          *string  `json:"method,omitempty"`
	StatusCode      *int     `json:"status_code,omitempty"`
	LatencyMs       *float64 `json:"latency_ms,omitempty"`
	ProcessingTime  *float64 `json:"processing_time_ms,omitempty"`
	ContentType     *string  `json:"content_type,omitempty"`
	ScanResult      *string  `json:"scan_result,omitempty"`
	WatermarkType   *string  `json:"watermark_type,omitempty"`
	ErrorCount      *int     `json:"error_count,omitempty"`
	Attempts        *int     `json:"attempts,omitempty"`
	Version         *string  `json:"version,omitempty"`
	Port            *int     `json:"port,omitempty"`
	Migration       *string  `json:"migration,omitempty"`
}

// SetMetadata sets the metadata field from a struct
func (e *MonitoringEvent) SetMetadata(metadata interface{}) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	metadataStr := string(data)
	e.Metadata = &metadataStr
	return nil
}

// GetMetadata unmarshals the metadata field into a struct
func (e *MonitoringEvent) GetMetadata(target interface{}) error {
	if e.Metadata == nil {
		return nil
	}
	return json.Unmarshal([]byte(*e.Metadata), target)
}

// CreateEvent creates a new monitoring event with default values
func CreateEvent(eventType, severity, source string) *MonitoringEvent {
	return &MonitoringEvent{
		EventType: eventType,
		Severity:  severity,
		Source:    &source,
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}
}

// CreateAPIRequestEvent creates an API request monitoring event
func CreateAPIRequestEvent(endpoint, method string, statusCode int, latencyMs float64) *MonitoringEvent {
	event := CreateEvent("api.request", "info", "api")
	event.EventSubtype = stringPtr("endpoint_call")
	
	metadata := EventMetadata{
		Endpoint:   &endpoint,
		Method:     &method,
		StatusCode: &statusCode,
		LatencyMs:  &latencyMs,
	}
	
	// Set severity based on status code
	if statusCode >= 400 {
		event.Severity = "warning"
	}
	if statusCode >= 500 {
		event.Severity = "error"
	}
	
	event.SetMetadata(metadata)
	return event
}

// CreateDLPScanEvent creates a DLP scan monitoring event
func CreateDLPScanEvent(contentType, scanResult string, processingTimeMs float64) *MonitoringEvent {
	event := CreateEvent("dlp.scan", "info", "dlp")
	event.EventSubtype = stringPtr("content_analysis")
	
	metadata := EventMetadata{
		ContentType:    &contentType,
		ScanResult:     &scanResult,
		ProcessingTime: &processingTimeMs,
	}
	
	event.SetMetadata(metadata)
	return event
}

// CreateWatermarkingEvent creates a watermarking operation monitoring event
func CreateWatermarkingEvent(watermarkType, contentType string, processingTimeMs float64) *MonitoringEvent {
	event := CreateEvent("watermarking.apply", "info", "watermarking")
	event.EventSubtype = stringPtr(watermarkType + "_watermark")
	
	metadata := EventMetadata{
		WatermarkType:  &watermarkType,
		ContentType:    &contentType,
		ProcessingTime: &processingTimeMs,
	}
	
	event.SetMetadata(metadata)
	return event
}

// CreateSecurityAlertEvent creates a security alert monitoring event
func CreateSecurityAlertEvent(alertType string, metadata EventMetadata) *MonitoringEvent {
	event := CreateEvent("security.alert", "warning", "security")
	event.EventSubtype = &alertType
	
	event.SetMetadata(metadata)
	return event
}

// CreateSystemEvent creates a system-level monitoring event
func CreateSystemEvent(eventType, subtype string, metadata EventMetadata) *MonitoringEvent {
	event := CreateEvent(eventType, "info", "system")
	event.EventSubtype = &subtype
	
	event.SetMetadata(metadata)
	return event
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
