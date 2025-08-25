package monitoring

import (
	"fmt"
	"log"
	"sync"
	"time"

	"secure-email-mvp/pkg/models"
)

// Service handles real-time monitoring and metrics collection
type Service struct {
	repository MonitoringRepository
	config     *models.MonitoringConfig
	
	// In-memory counters for real-time metrics
	counters struct {
		sync.RWMutex
		requestCount    int64
		errorCount      int64
		dlpScans        int64
		watermarkingOps int64
		securityAlerts  int64
		lastReset       time.Time
	}
	
	// Stream subscribers for real-time updates
	subscribers struct {
		sync.RWMutex
		clients map[string]chan *models.StreamEvent
	}
}

// NewService creates a new monitoring service
func NewService(repository MonitoringRepository, config *models.MonitoringConfig) *Service {
	service := &Service{
		repository: repository,
		config:     config,
	}
	
	// Initialize subscribers map
	service.subscribers.clients = make(map[string]chan *models.StreamEvent)
	
	// Start background tasks
	go service.startMetricsAggregation()
	go service.startRetentionCleanup()
	
	return service
}

// LogEvent logs a monitoring event
func (s *Service) LogEvent(event *models.MonitoringEvent) error {
	if !s.config.Enabled {
		return nil
	}
	
	// Apply sampling if configured
	if s.config.SampleRate < 1.0 {
		if time.Now().UnixNano()%100 < int64(s.config.SampleRate*100) {
			return nil
		}
	}
	
	// Update in-memory counters
	s.updateCounters(event)
	
	// Save to database
	if err := s.repository.LogEvent(event); err != nil {
		log.Printf("Failed to log monitoring event: %v", err)
		return err
	}
	
	// Broadcast to subscribers
	s.broadcastEvent(event)
	
	return nil
}

// LogAPIRequest logs an API request event
func (s *Service) LogAPIRequest(endpoint, method string, statusCode int, latencyMs float64, userID, sessionID, ipAddress, userAgent *string) error {
	event := models.CreateAPIRequestEvent(endpoint, method, statusCode, latencyMs)
	event.UserID = userID
	event.SessionID = sessionID
	event.IPAddress = ipAddress
	event.UserAgent = userAgent
	
	return s.LogEvent(event)
}

// LogDLPScan logs a DLP scan event
func (s *Service) LogDLPScan(contentType, scanResult string, processingTimeMs float64) error {
	event := models.CreateDLPScanEvent(contentType, scanResult, processingTimeMs)
	return s.LogEvent(event)
}

// LogWatermarking logs a watermarking operation event
func (s *Service) LogWatermarking(watermarkType, contentType string, processingTimeMs float64) error {
	event := models.CreateWatermarkingEvent(watermarkType, contentType, processingTimeMs)
	return s.LogEvent(event)
}

// LogSecurityAlert logs a security alert event
func (s *Service) LogSecurityAlert(alertType string, metadata models.EventMetadata) error {
	event := models.CreateSecurityAlertEvent(alertType, metadata)
	return s.LogEvent(event)
}

// GetRealTimeMetrics retrieves current real-time metrics
func (s *Service) GetRealTimeMetrics() (*models.RealTimeMetrics, error) {
	// Get database metrics
	dbMetrics, err := s.repository.GetRealTimeMetrics()
	if err != nil {
		return nil, err
	}
	
	// Combine with in-memory counters for more accurate real-time data
	s.counters.RLock()
	defer s.counters.RUnlock()
	
	// If counters were reset recently, use them; otherwise use database metrics
	if time.Since(s.counters.lastReset) < time.Minute {
		dbMetrics.RequestCount = s.counters.requestCount
		dbMetrics.DLPScans = s.counters.dlpScans
		dbMetrics.WatermarkingOps = s.counters.watermarkingOps
		dbMetrics.SecurityAlerts = s.counters.securityAlerts
	}
	
	return dbMetrics, nil
}

// GetMetricsSummary retrieves metrics summary for charts
func (s *Service) GetMetricsSummary(metricName, timeBucket string, startDate, endDate time.Time) ([]*models.MetricsSummary, error) {
	return s.repository.GetMetricsSummary(metricName, timeBucket, startDate, endDate)
}

// GetRecentEvents retrieves recent events for streaming
func (s *Service) GetRecentEvents(limit int) ([]*models.MonitoringEvent, error) {
	return s.repository.GetRecentEvents(limit)
}

// SubscribeToStream subscribes to real-time monitoring events
func (s *Service) SubscribeToStream(clientID string) (<-chan *models.StreamEvent, error) {
	s.subscribers.Lock()
	defer s.subscribers.Unlock()
	
	// Create channel for this client
	eventChan := make(chan *models.StreamEvent, 100)
	s.subscribers.clients[clientID] = eventChan
	
	log.Printf("Client %s subscribed to monitoring stream", clientID)
	
	return eventChan, nil
}

// UnsubscribeFromStream unsubscribes from real-time monitoring events
func (s *Service) UnsubscribeFromStream(clientID string) {
	s.subscribers.Lock()
	defer s.subscribers.Unlock()
	
	if eventChan, exists := s.subscribers.clients[clientID]; exists {
		close(eventChan)
		delete(s.subscribers.clients, clientID)
		log.Printf("Client %s unsubscribed from monitoring stream", clientID)
	}
}

// GetMonitoringConfig retrieves monitoring configuration
func (s *Service) GetMonitoringConfig() *models.MonitoringConfig {
	return s.config
}

// UpdateMonitoringConfig updates monitoring configuration
func (s *Service) UpdateMonitoringConfig(config *models.MonitoringConfig) error {
	s.config = config
	log.Printf("Monitoring configuration updated: enabled=%v, sample_rate=%f", config.Enabled, config.SampleRate)
	return nil
}

// updateCounters updates in-memory counters based on event
func (s *Service) updateCounters(event *models.MonitoringEvent) {
	s.counters.Lock()
	defer s.counters.Unlock()
	
	switch event.EventType {
	case "api.request":
		s.counters.requestCount++
		if event.Severity == "error" || event.Severity == "critical" {
			s.counters.errorCount++
		}
	case "dlp.scan":
		s.counters.dlpScans++
	case "watermarking.apply":
		s.counters.watermarkingOps++
	case "security.alert":
		s.counters.securityAlerts++
	}
}

// broadcastEvent broadcasts an event to all subscribers
func (s *Service) broadcastEvent(event *models.MonitoringEvent) {
	s.subscribers.RLock()
	defer s.subscribers.RUnlock()
	
	streamEvent := &models.StreamEvent{
		Type:      "monitoring_event",
		Timestamp: time.Now(),
		Data:      event,
	}
	
	for clientID, eventChan := range s.subscribers.clients {
		select {
		case eventChan <- streamEvent:
			// Event sent successfully
		default:
			// Channel is full, remove this subscriber
			log.Printf("Removing subscriber %s due to full channel", clientID)
			go s.UnsubscribeFromStream(clientID)
		}
	}
}

// startMetricsAggregation starts background metrics aggregation
func (s *Service) startMetricsAggregation() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.aggregateMetrics()
	}
}

// aggregateMetrics aggregates metrics and saves summaries
func (s *Service) aggregateMetrics() {
	now := time.Now()
	
	// Get metrics for the last minute
	metrics, err := s.repository.GetRealTimeMetrics()
	if err != nil {
		log.Printf("Failed to get real-time metrics for aggregation: %v", err)
		return
	}
	
	// Save request count summary
	requestSummary := &models.MetricsSummary{
		MetricName:  "request_count",
		MetricValue: float64(metrics.RequestCount),
		MetricUnit:  stringPtr("count"),
		TimeBucket:  "minute",
		BucketStart: now.Add(-time.Minute),
		BucketEnd:   now,
		Source:      stringPtr("api"),
		CreatedAt:   now,
	}
	
	if err := s.repository.SaveMetricsSummary(requestSummary); err != nil {
		log.Printf("Failed to save request count summary: %v", err)
	}
	
	// Save error rate summary
	errorRateSummary := &models.MetricsSummary{
		MetricName:  "error_rate",
		MetricValue: metrics.ErrorRate,
		MetricUnit:  stringPtr("percentage"),
		TimeBucket:  "minute",
		BucketStart: now.Add(-time.Minute),
		BucketEnd:   now,
		Source:      stringPtr("api"),
		CreatedAt:   now,
	}
	
	if err := s.repository.SaveMetricsSummary(errorRateSummary); err != nil {
		log.Printf("Failed to save error rate summary: %v", err)
	}
	
	// Save average latency summary
	latencySummary := &models.MetricsSummary{
		MetricName:  "average_latency",
		MetricValue: metrics.AverageLatency,
		MetricUnit:  stringPtr("milliseconds"),
		TimeBucket:  "minute",
		BucketStart: now.Add(-time.Minute),
		BucketEnd:   now,
		Source:      stringPtr("api"),
		CreatedAt:   now,
	}
	
	if err := s.repository.SaveMetricsSummary(latencySummary); err != nil {
		log.Printf("Failed to save latency summary: %v", err)
	}
	
	// Reset in-memory counters
	s.counters.Lock()
	s.counters.requestCount = 0
	s.counters.errorCount = 0
	s.counters.dlpScans = 0
	s.counters.watermarkingOps = 0
	s.counters.securityAlerts = 0
	s.counters.lastReset = now
	s.counters.Unlock()
	
	log.Printf("Metrics aggregated for %s", now.Format("2006-01-02 15:04:05"))
}

// startRetentionCleanup starts background retention cleanup
func (s *Service) startRetentionCleanup() {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()
	
	for range ticker.C {
		s.cleanupOldEvents()
	}
}

// cleanupOldEvents removes events older than retention period
func (s *Service) cleanupOldEvents() {
	retentionDate := time.Now().AddDate(0, 0, -s.config.RetentionDays)
	
	// Get old events
	oldEvents, err := s.repository.GetEventsByDateRange(time.Time{}, retentionDate)
	if err != nil {
		log.Printf("Failed to get old events for cleanup: %v", err)
		return
	}
	
	if len(oldEvents) > 0 {
		log.Printf("Cleaning up %d events older than %s", len(oldEvents), retentionDate.Format("2006-01-02"))
	}
	
	// Note: The actual cleanup is handled by the database trigger,
	// but we can log the cleanup event
	cleanupEvent := models.CreateSystemEvent("cleanup.complete", "retention_cleanup", models.EventMetadata{
		ErrorCount: &[]int{len(oldEvents)}[0],
	})
	
	if err := s.LogEvent(cleanupEvent); err != nil {
		log.Printf("Failed to log cleanup event: %v", err)
	}
}

// GetSystemHealth returns system health status
func (s *Service) GetSystemHealth() map[string]interface{} {
	metrics, err := s.GetRealTimeMetrics()
	if err != nil {
		return map[string]interface{}{
			"status":    "error",
			"message":   "Failed to get metrics",
			"timestamp": time.Now(),
		}
	}
	
	// Determine health status based on thresholds
	status := "healthy"
	message := "System is operating normally"
	
	if metrics.ErrorRate > s.config.AlertThresholdErrorRate*100 {
		status = "warning"
		message = fmt.Sprintf("High error rate: %.2f%%", metrics.ErrorRate)
	}
	
	if metrics.AverageLatency > float64(s.config.AlertThresholdLatencyMs) {
		status = "warning"
		message = fmt.Sprintf("High latency: %.2fms", metrics.AverageLatency)
	}
	
	return map[string]interface{}{
		"status":         status,
		"message":        message,
		"timestamp":      time.Now(),
		"error_rate":     metrics.ErrorRate,
		"avg_latency":    metrics.AverageLatency,
		"request_count":  metrics.RequestCount,
		"active_sessions": metrics.ActiveSessions,
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
