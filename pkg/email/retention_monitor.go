package email

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// RetentionEvent represents a real-time retention event
type RetentionEvent struct {
	ID           int64           `json:"id"`
	EventType    string          `json:"event_type"`    // "policy_evaluation", "email_deletion", "email_archival", "policy_change"
	EventData    json.RawMessage `json:"event_data"`    // JSON event data
	UserID       *string         `json:"user_id,omitempty"`
	PolicyID     *int64          `json:"policy_id,omitempty"`
	EventTime    time.Time       `json:"event_time"`
	Processed    bool            `json:"processed"`
}

// RealtimeMetrics represents live retention metrics for a specific scope
type RealtimeMetrics struct {
	MetricType   string    `json:"metric_type"`   // "user", "domain", "policy", "global"
	MetricKey    string    `json:"metric_key"`    // user_id, domain, policy_id, or "global"
	
	// Live metrics
	ActiveEmailsCount      int     `json:"active_emails_count"`
	ArchivedEmailsCount    int     `json:"archived_emails_count"`
	DeletedEmailsCount     int     `json:"deleted_emails_count"`
	TotalStorageBytes      int64   `json:"total_storage_bytes"`
	CompressedStorageBytes int64   `json:"compressed_storage_bytes"`
	
	// Policy performance metrics
	PolicyEvaluationsCount int     `json:"policy_evaluations_count"`
	PolicyMatchesCount     int     `json:"policy_matches_count"`
	PolicyApplicationsCount int    `json:"policy_applications_count"`
	AvgMatchScore          float64 `json:"avg_match_score"`
	AvgImpactScore         float64 `json:"avg_impact_score"`
	
	// Archival load metrics
	ArchivalOperationsCount int     `json:"archival_operations_count"`
	AvgArchivalDurationMs   int     `json:"avg_archival_duration_ms"`
	ArchivalSuccessRate     float64 `json:"archival_success_rate"`
	
	LastUpdated time.Time `json:"last_updated"`
}

// RetentionMonitorService provides real-time monitoring of retention operations
type RetentionMonitorService struct {
	db           *sql.DB
	metricsCache map[string]*RealtimeMetrics // In-memory cache: key = "type:key"
	cacheMutex   sync.RWMutex
	eventChannel chan *RetentionEvent // Channel for real-time events
}

// NewRetentionMonitorService creates a new retention monitor service
func NewRetentionMonitorService(db *sql.DB) *RetentionMonitorService {
	return &RetentionMonitorService{
		db:           db,
		metricsCache: make(map[string]*RealtimeMetrics),
		eventChannel: make(chan *RetentionEvent, 1000), // Buffer for 1000 events
	}
}

// Start begins the real-time monitoring service
func (rms *RetentionMonitorService) Start(ctx context.Context) {
	log.Println("Starting Retention Monitor Service...")
	
	// Initialize metrics cache from database
	if err := rms.initializeMetricsCache(ctx); err != nil {
		log.Printf("Failed to initialize metrics cache: %v", err)
	}
	
	// Start event processing goroutine
	go rms.processEvents(ctx)
	
	// Start metrics update goroutine
	go rms.updateMetricsPeriodically(ctx)
	
	log.Println("Retention Monitor Service started successfully")
}

// Stop gracefully stops the monitoring service
func (rms *RetentionMonitorService) Stop() {
	log.Println("Stopping Retention Monitor Service...")
	close(rms.eventChannel)
}

// RecordPolicyEvaluation records a policy evaluation event in real-time
func (rms *RetentionMonitorService) RecordPolicyEvaluation(ctx context.Context, emailID string, policyID int64, result string, matchScore int, reasons []string, impactScore float64, storageSavings int64, archivalLoadImpact float64) error {
	// Create event data
	eventData := map[string]interface{}{
		"email_id":              emailID,
		"policy_id":             policyID,
		"result":                result,
		"match_score":           matchScore,
		"reasons":               reasons,
		"impact_score":          impactScore,
		"storage_savings_bytes": storageSavings,
		"archival_load_impact":  archivalLoadImpact,
	}
	
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	
	// Create retention event
	event := &RetentionEvent{
		EventType: "policy_evaluation",
		EventData: eventJSON,
		PolicyID:  &policyID,
		EventTime: time.Now(),
		Processed: false,
	}
	
	// Store event in database
	if err := rms.storeEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}
	
	// Send to event channel for real-time processing
	select {
	case rms.eventChannel <- event:
	default:
		log.Printf("Warning: Event channel full, dropping policy evaluation event")
	}
	
	// Update metrics immediately
	rms.updatePolicyMetrics(policyID, result, matchScore, impactScore, storageSavings, archivalLoadImpact)
	
	return nil
}

// RecordEmailDeletion records an email deletion event
func (rms *RetentionMonitorService) RecordEmailDeletion(ctx context.Context, emailID string, userID string, reason string, storageBytes int64) error {
	eventData := map[string]interface{}{
		"email_id":      emailID,
		"user_id":       userID,
		"reason":        reason,
		"storage_bytes": storageBytes,
	}
	
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	
	event := &RetentionEvent{
		EventType: "email_deletion",
		EventData: eventJSON,
		UserID:    &userID,
		EventTime: time.Now(),
		Processed: false,
	}
	
	if err := rms.storeEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}
	
	select {
	case rms.eventChannel <- event:
	default:
		log.Printf("Warning: Event channel full, dropping email deletion event")
	}
	
	// Update metrics
	rms.updateDeletionMetrics(userID, storageBytes)
	
	return nil
}

// RecordEmailArchival records an email archival event
func (rms *RetentionMonitorService) RecordEmailArchival(ctx context.Context, emailID string, userID string, originalSize int64, compressedSize int64, durationMs int, success bool) error {
	eventData := map[string]interface{}{
		"email_id":       emailID,
		"user_id":        userID,
		"original_size":  originalSize,
		"compressed_size": compressedSize,
		"duration_ms":    durationMs,
		"success":        success,
	}
	
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	
	event := &RetentionEvent{
		EventType: "email_archival",
		EventData: eventJSON,
		UserID:    &userID,
		EventTime: time.Now(),
		Processed: false,
	}
	
	if err := rms.storeEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}
	
	select {
	case rms.eventChannel <- event:
	default:
		log.Printf("Warning: Event channel full, dropping email archival event")
	}
	
	// Update metrics
	rms.updateArchivalMetrics(userID, originalSize, compressedSize, durationMs, success)
	
	return nil
}

// GetRealtimeMetrics retrieves real-time metrics for a specific scope
func (rms *RetentionMonitorService) GetRealtimeMetrics(ctx context.Context, metricType, metricKey string) (*RealtimeMetrics, error) {
	rms.cacheMutex.RLock()
	defer rms.cacheMutex.RUnlock()
	
	cacheKey := fmt.Sprintf("%s:%s", metricType, metricKey)
	metrics, exists := rms.metricsCache[cacheKey]
	
	if !exists {
		// Load from database
		return rms.loadMetricsFromDB(ctx, metricType, metricKey)
	}
	
	return metrics, nil
}

// GetGlobalMetrics retrieves global real-time metrics
func (rms *RetentionMonitorService) GetGlobalMetrics(ctx context.Context) (*RealtimeMetrics, error) {
	return rms.GetRealtimeMetrics(ctx, "global", "global")
}

// GetUserMetrics retrieves real-time metrics for a specific user
func (rms *RetentionMonitorService) GetUserMetrics(ctx context.Context, userID string) (*RealtimeMetrics, error) {
	return rms.GetRealtimeMetrics(ctx, "user", userID)
}

// GetDomainMetrics retrieves real-time metrics for a specific domain
func (rms *RetentionMonitorService) GetDomainMetrics(ctx context.Context, domain string) (*RealtimeMetrics, error) {
	return rms.GetRealtimeMetrics(ctx, "domain", domain)
}

// GetPolicyMetrics retrieves real-time metrics for a specific policy
func (rms *RetentionMonitorService) GetPolicyMetrics(ctx context.Context, policyID int64) (*RealtimeMetrics, error) {
	return rms.GetRealtimeMetrics(ctx, "policy", fmt.Sprintf("%d", policyID))
}

// GetUnprocessedEvents retrieves unprocessed events for WebSocket/SSE streaming
func (rms *RetentionMonitorService) GetUnprocessedEvents(ctx context.Context, limit int) ([]*RetentionEvent, error) {
	query := `
		SELECT id, event_type, event_data, user_id, policy_id, event_timestamp, processed
		FROM retention_events
		WHERE processed = 0
		ORDER BY event_timestamp ASC
		LIMIT ?
	`
	
	rows, err := rms.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unprocessed events: %w", err)
	}
	defer rows.Close()
	
	var events []*RetentionEvent
	for rows.Next() {
		var event RetentionEvent
		var userID, policyID sql.NullString
		
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EventData, &userID, &policyID,
			&event.EventTime, &event.Processed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		
		if userID.Valid {
			event.UserID = &userID.String
		}
		if policyID.Valid {
			if id, err := parsePolicyID(policyID.String); err == nil {
				event.PolicyID = &id
			}
		}
		
		events = append(events, &event)
	}
	
	return events, nil
}

// MarkEventProcessed marks an event as processed
func (rms *RetentionMonitorService) MarkEventProcessed(ctx context.Context, eventID int64) error {
	query := `UPDATE retention_events SET processed = 1 WHERE id = ?`
	_, err := rms.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}
	return nil
}

// initializeMetricsCache loads initial metrics from database
func (rms *RetentionMonitorService) initializeMetricsCache(ctx context.Context) error {
	query := `
		SELECT metric_type, metric_key, active_emails_count, archived_emails_count,
		       deleted_emails_count, total_storage_bytes, compressed_storage_bytes,
		       policy_evaluations_count, policy_matches_count, policy_applications_count,
		       avg_match_score, avg_impact_score, archival_operations_count,
		       avg_archival_duration_ms, archival_success_rate, last_updated
		FROM realtime_retention_metrics
	`
	
	rows, err := rms.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()
	
	rms.cacheMutex.Lock()
	defer rms.cacheMutex.Unlock()
	
	for rows.Next() {
		var metrics RealtimeMetrics
		err := rows.Scan(
			&metrics.MetricType, &metrics.MetricKey, &metrics.ActiveEmailsCount,
			&metrics.ArchivedEmailsCount, &metrics.DeletedEmailsCount,
			&metrics.TotalStorageBytes, &metrics.CompressedStorageBytes,
			&metrics.PolicyEvaluationsCount, &metrics.PolicyMatchesCount,
			&metrics.PolicyApplicationsCount, &metrics.AvgMatchScore,
			&metrics.AvgImpactScore, &metrics.ArchivalOperationsCount,
			&metrics.AvgArchivalDurationMs, &metrics.ArchivalSuccessRate,
			&metrics.LastUpdated,
		)
		if err != nil {
			return fmt.Errorf("failed to scan metrics: %w", err)
		}
		
		cacheKey := fmt.Sprintf("%s:%s", metrics.MetricType, metrics.MetricKey)
		rms.metricsCache[cacheKey] = &metrics
	}
	
	log.Printf("Initialized metrics cache with %d entries", len(rms.metricsCache))
	return nil
}

// processEvents processes events from the event channel
func (rms *RetentionMonitorService) processEvents(ctx context.Context) {
	for {
		select {
		case event, ok := <-rms.eventChannel:
			if !ok {
				return // Channel closed
			}
			
			if err := rms.processEvent(ctx, event); err != nil {
				log.Printf("Failed to process event: %v", err)
			}
			
		case <-ctx.Done():
			return
		}
	}
}

// processEvent processes a single event
func (rms *RetentionMonitorService) processEvent(ctx context.Context, event *RetentionEvent) error {
	switch event.EventType {
	case "policy_evaluation":
		return rms.processPolicyEvaluationEvent(ctx, event)
	case "email_deletion":
		return rms.processEmailDeletionEvent(ctx, event)
	case "email_archival":
		return rms.processEmailArchivalEvent(ctx, event)
	case "policy_change":
		return rms.processPolicyChangeEvent(ctx, event)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
		return nil
	}
}

// updateMetricsPeriodically updates metrics cache periodically
func (rms *RetentionMonitorService) updateMetricsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := rms.refreshMetricsCache(ctx); err != nil {
				log.Printf("Failed to refresh metrics cache: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Helper methods for event processing and metrics updates
func (rms *RetentionMonitorService) storeEvent(ctx context.Context, event *RetentionEvent) error {
	query := `
		INSERT INTO retention_events (event_type, event_data, user_id, policy_id, event_timestamp, processed)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	var policyID interface{}
	if event.PolicyID != nil {
		policyID = *event.PolicyID
	}
	
	_, err := rms.db.ExecContext(ctx, query,
		event.EventType, event.EventData, event.UserID, policyID, event.EventTime, event.Processed,
	)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	
	return nil
}

func (rms *RetentionMonitorService) loadMetricsFromDB(ctx context.Context, metricType, metricKey string) (*RealtimeMetrics, error) {
	query := `
		SELECT metric_type, metric_key, active_emails_count, archived_emails_count,
		       deleted_emails_count, total_storage_bytes, compressed_storage_bytes,
		       policy_evaluations_count, policy_matches_count, policy_applications_count,
		       avg_match_score, avg_impact_score, archival_operations_count,
		       avg_archival_duration_ms, archival_success_rate, last_updated
		FROM realtime_retention_metrics
		WHERE metric_type = ? AND metric_key = ?
	`
	
	var metrics RealtimeMetrics
	err := rms.db.QueryRowContext(ctx, query, metricType, metricKey).Scan(
		&metrics.MetricType, &metrics.MetricKey, &metrics.ActiveEmailsCount,
		&metrics.ArchivedEmailsCount, &metrics.DeletedEmailsCount,
		&metrics.TotalStorageBytes, &metrics.CompressedStorageBytes,
		&metrics.PolicyEvaluationsCount, &metrics.PolicyMatchesCount,
		&metrics.PolicyApplicationsCount, &metrics.AvgMatchScore,
		&metrics.AvgImpactScore, &metrics.ArchivalOperationsCount,
		&metrics.AvgArchivalDurationMs, &metrics.ArchivalSuccessRate,
		&metrics.LastUpdated,
	)
	
	if err == sql.ErrNoRows {
		// Create new metrics entry
		metrics = RealtimeMetrics{
			MetricType: metricType,
			MetricKey:  metricKey,
			LastUpdated: time.Now(),
		}
		if err := rms.createMetricsEntry(ctx, &metrics); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to load metrics: %w", err)
	}
	
	// Cache the metrics
	rms.cacheMutex.Lock()
	cacheKey := fmt.Sprintf("%s:%s", metricType, metricKey)
	rms.metricsCache[cacheKey] = &metrics
	rms.cacheMutex.Unlock()
	
	return &metrics, nil
}

func (rms *RetentionMonitorService) createMetricsEntry(ctx context.Context, metrics *RealtimeMetrics) error {
	query := `
		INSERT INTO realtime_retention_metrics (
			metric_type, metric_key, active_emails_count, archived_emails_count,
			deleted_emails_count, total_storage_bytes, compressed_storage_bytes,
			policy_evaluations_count, policy_matches_count, policy_applications_count,
			avg_match_score, avg_impact_score, archival_operations_count,
			avg_archival_duration_ms, archival_success_rate, last_updated, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := rms.db.ExecContext(ctx, query,
		metrics.MetricType, metrics.MetricKey, metrics.ActiveEmailsCount,
		metrics.ArchivedEmailsCount, metrics.DeletedEmailsCount,
		metrics.TotalStorageBytes, metrics.CompressedStorageBytes,
		metrics.PolicyEvaluationsCount, metrics.PolicyMatchesCount,
		metrics.PolicyApplicationsCount, metrics.AvgMatchScore,
		metrics.AvgImpactScore, metrics.ArchivalOperationsCount,
		metrics.AvgArchivalDurationMs, metrics.ArchivalSuccessRate,
		metrics.LastUpdated, time.Now(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to create metrics entry: %w", err)
	}
	
	return nil
}

func (rms *RetentionMonitorService) refreshMetricsCache(ctx context.Context) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled during cache refresh: %w", ctx.Err())
	default:
	}
	
	// This would refresh the cache with the latest data from the database
	// For now, we'll just update the last_updated timestamp
	rms.cacheMutex.Lock()
	defer rms.cacheMutex.Unlock()
	
	for _, metrics := range rms.metricsCache {
		// Check for context cancellation during iteration
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during cache refresh: %w", ctx.Err())
		default:
		}
		metrics.LastUpdated = time.Now()
	}
	
	log.Printf("Refreshed metrics cache with %d entries", len(rms.metricsCache))
	return nil
}

// Event processing methods
func (rms *RetentionMonitorService) processPolicyEvaluationEvent(ctx context.Context, event *RetentionEvent) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled during policy evaluation processing: %w", ctx.Err())
	default:
	}
	
	// Parse event data
	var eventData map[string]interface{}
	if err := json.Unmarshal(event.EventData, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}
	
	// Update policy metrics
	policyID := eventData["policy_id"].(float64)
	result := eventData["result"].(string)
	matchScore := int(eventData["match_score"].(float64))
	impactScore := eventData["impact_score"].(float64)
	storageSavings := int64(eventData["storage_savings_bytes"].(float64))
	archivalLoadImpact := eventData["archival_load_impact"].(float64)
	
	rms.updatePolicyMetrics(int64(policyID), result, matchScore, impactScore, storageSavings, archivalLoadImpact)
	
	log.Printf("Processed policy evaluation event for policy %d, result: %s", int64(policyID), result)
	return nil
}

func (rms *RetentionMonitorService) processEmailDeletionEvent(ctx context.Context, event *RetentionEvent) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled during email deletion processing: %w", ctx.Err())
	default:
	}
	
	var eventData map[string]interface{}
	if err := json.Unmarshal(event.EventData, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}
	
	userID := eventData["user_id"].(string)
	storageBytes := int64(eventData["storage_bytes"].(float64))
	
	rms.updateDeletionMetrics(userID, storageBytes)
	
	log.Printf("Processed email deletion event for user %s, storage freed: %d bytes", userID, storageBytes)
	return nil
}

func (rms *RetentionMonitorService) processEmailArchivalEvent(ctx context.Context, event *RetentionEvent) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled during email archival processing: %w", ctx.Err())
	default:
	}
	
	var eventData map[string]interface{}
	if err := json.Unmarshal(event.EventData, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}
	
	userID := eventData["user_id"].(string)
	originalSize := int64(eventData["original_size"].(float64))
	compressedSize := int64(eventData["compressed_size"].(float64))
	durationMs := int(eventData["duration_ms"].(float64))
	success := eventData["success"].(bool)
	
	rms.updateArchivalMetrics(userID, originalSize, compressedSize, durationMs, success)
	
	log.Printf("Processed email archival event for user %s, original: %d bytes, compressed: %d bytes, duration: %d ms, success: %t", 
		userID, originalSize, compressedSize, durationMs, success)
	return nil
}

func (rms *RetentionMonitorService) processPolicyChangeEvent(ctx context.Context, event *RetentionEvent) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled during policy change processing: %w", ctx.Err())
	default:
	}
	
	// Parse event data
	var eventData map[string]interface{}
	if err := json.Unmarshal(event.EventData, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}
	
	// Extract policy change information
	policyID := eventData["policy_id"].(float64)
	changeType := eventData["change_type"].(string) // "created", "updated", "deleted"
	oldSettings := eventData["old_settings"]
	newSettings := eventData["new_settings"]
	
	// Log policy change for audit purposes
	log.Printf("Policy change event: policy %d, change type: %s, old settings: %v, new settings: %v", 
		int64(policyID), changeType, oldSettings, newSettings)
	
	// Update policy metrics cache to reflect changes
	rms.cacheMutex.Lock()
	defer rms.cacheMutex.Unlock()
	
	cacheKey := fmt.Sprintf("policy:%d", int64(policyID))
	
	switch changeType {
	case "created":
		// Initialize new policy metrics
		if _, exists := rms.metricsCache[cacheKey]; !exists {
			rms.metricsCache[cacheKey] = &RealtimeMetrics{
				MetricType: "policy",
				MetricKey:  fmt.Sprintf("%d", int64(policyID)),
				LastUpdated: time.Now(),
			}
		}
	case "updated":
		// Update existing policy metrics timestamp
		if metrics, exists := rms.metricsCache[cacheKey]; exists {
			metrics.LastUpdated = time.Now()
		}
	case "deleted":
		// Remove policy metrics from cache
		delete(rms.metricsCache, cacheKey)
	default:
		log.Printf("Unknown policy change type: %s", changeType)
	}
	
	return nil
}

// Metrics update methods
func (rms *RetentionMonitorService) updatePolicyMetrics(policyID int64, result string, matchScore int, impactScore float64, storageSavings int64, archivalLoadImpact float64) {
	rms.cacheMutex.Lock()
	defer rms.cacheMutex.Unlock()
	
	cacheKey := fmt.Sprintf("policy:%d", policyID)
	metrics, exists := rms.metricsCache[cacheKey]
	
	if !exists {
		metrics = &RealtimeMetrics{
			MetricType: "policy",
			MetricKey:  fmt.Sprintf("%d", policyID),
			LastUpdated: time.Now(),
		}
		rms.metricsCache[cacheKey] = metrics
	}
	
	metrics.PolicyEvaluationsCount++
	metrics.AvgMatchScore = (metrics.AvgMatchScore*float64(metrics.PolicyEvaluationsCount-1) + float64(matchScore)) / float64(metrics.PolicyEvaluationsCount)
	metrics.AvgImpactScore = (metrics.AvgImpactScore*float64(metrics.PolicyEvaluationsCount-1) + impactScore) / float64(metrics.PolicyEvaluationsCount)
	
	// Use storageSavings for storage efficiency tracking
	if storageSavings > 0 {
		metrics.TotalStorageBytes -= storageSavings
		log.Printf("Policy %d achieved storage savings of %d bytes", policyID, storageSavings)
	}
	
	// Use archivalLoadImpact for performance monitoring
	if archivalLoadImpact > 0 {
		log.Printf("Policy %d archival load impact: %.2f", policyID, archivalLoadImpact)
	}
	
	metrics.LastUpdated = time.Now()
	
	// Use switch statement for better performance
	switch result {
	case "matched":
		metrics.PolicyMatchesCount++
	case "applied":
		metrics.PolicyApplicationsCount++
	case "rejected":
		// Track rejected policies for analysis
		log.Printf("Policy %d evaluation rejected", policyID)
	default:
		log.Printf("Unknown policy evaluation result: %s", result)
	}
	
	// Update global metrics
	globalKey := "global:global"
	globalMetrics, exists := rms.metricsCache[globalKey]
	if exists {
		globalMetrics.PolicyEvaluationsCount++
		globalMetrics.AvgMatchScore = (globalMetrics.AvgMatchScore*float64(globalMetrics.PolicyEvaluationsCount-1) + float64(matchScore)) / float64(globalMetrics.PolicyEvaluationsCount)
		globalMetrics.AvgImpactScore = (globalMetrics.AvgImpactScore*float64(globalMetrics.PolicyEvaluationsCount-1) + impactScore) / float64(globalMetrics.PolicyEvaluationsCount)
		globalMetrics.LastUpdated = time.Now()
	}
}

func (rms *RetentionMonitorService) updateDeletionMetrics(userID string, storageBytes int64) {
	rms.cacheMutex.Lock()
	defer rms.cacheMutex.Unlock()
	
	// Update user metrics
	userKey := fmt.Sprintf("user:%s", userID)
	userMetrics, exists := rms.metricsCache[userKey]
	if !exists {
		userMetrics = &RealtimeMetrics{
			MetricType: "user",
			MetricKey:  userID,
			LastUpdated: time.Now(),
		}
		rms.metricsCache[userKey] = userMetrics
	}
	
	userMetrics.DeletedEmailsCount++
	userMetrics.TotalStorageBytes -= storageBytes
	userMetrics.LastUpdated = time.Now()
	
	// Update global metrics
	globalKey := "global:global"
	globalMetrics, exists := rms.metricsCache[globalKey]
	if exists {
		globalMetrics.DeletedEmailsCount++
		globalMetrics.TotalStorageBytes -= storageBytes
		globalMetrics.LastUpdated = time.Now()
	}
}

func (rms *RetentionMonitorService) updateArchivalMetrics(userID string, originalSize int64, compressedSize int64, durationMs int, success bool) {
	rms.cacheMutex.Lock()
	defer rms.cacheMutex.Unlock()
	
	// Update user metrics
	userKey := fmt.Sprintf("user:%s", userID)
	userMetrics, exists := rms.metricsCache[userKey]
	if !exists {
		userMetrics = &RealtimeMetrics{
			MetricType: "user",
			MetricKey:  userID,
			LastUpdated: time.Now(),
		}
		rms.metricsCache[userKey] = userMetrics
	}
	
	userMetrics.ArchivedEmailsCount++
	userMetrics.TotalStorageBytes += originalSize
	userMetrics.CompressedStorageBytes += compressedSize
	userMetrics.ArchivalOperationsCount++
	
	// Update average archival duration
	userMetrics.AvgArchivalDurationMs = (userMetrics.AvgArchivalDurationMs*(userMetrics.ArchivalOperationsCount-1) + durationMs) / userMetrics.ArchivalOperationsCount
	
	// Update success rate
	if success {
		userMetrics.ArchivalSuccessRate = float64(userMetrics.ArchivalOperationsCount) / float64(userMetrics.ArchivalOperationsCount)
	}
	
	userMetrics.LastUpdated = time.Now()
	
	// Update global metrics
	globalKey := "global:global"
	globalMetrics, exists := rms.metricsCache[globalKey]
	if exists {
		globalMetrics.ArchivedEmailsCount++
		globalMetrics.TotalStorageBytes += originalSize
		globalMetrics.CompressedStorageBytes += compressedSize
		globalMetrics.ArchivalOperationsCount++
		globalMetrics.AvgArchivalDurationMs = (globalMetrics.AvgArchivalDurationMs*(globalMetrics.ArchivalOperationsCount-1) + durationMs) / globalMetrics.ArchivalOperationsCount
		if success {
			globalMetrics.ArchivalSuccessRate = float64(globalMetrics.ArchivalOperationsCount) / float64(globalMetrics.ArchivalOperationsCount)
		}
		globalMetrics.LastUpdated = time.Now()
	}
}

// Helper function to parse policy ID
func parsePolicyID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}






