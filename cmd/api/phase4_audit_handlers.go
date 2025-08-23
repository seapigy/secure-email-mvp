package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/audit"
	"secure-email-mvp/pkg/quota"
	"secure-email-mvp/pkg/retry"
)

// =============================================================================
// PHASE 4 AUDIT & RELIABILITY API HANDLERS
// =============================================================================

// AuditHandler handles audit log queries and management
func (srv *Server) AuditHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("AuditHandler called - Method: %s", r.Method)

	switch r.Method {
	case http.MethodGet:
		srv.handleGetAuditEvents(w, r)
	case http.MethodPost:
		srv.handleCreateAuditEvent(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetAuditEvents handles GET requests to retrieve audit events
func (srv *Server) handleGetAuditEvents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := audit.AuditLogFilter{}
	page := 1
	pageSize := 100

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filter.UserIDs = []string{userID}
	}

	// Parse time range
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.DateFrom = &startTime
		}
	}
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.DateTo = &endTime
		}
	}

	// Parse pagination
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
			pageSize = ps
		}
	}

	// Create audit service
	auditService := audit.NewAuditService(srv.db)

	// Query events
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := auditService.QueryEvents(ctx, filter, page, pageSize)
	if err != nil {
		log.Printf("Error querying audit events: %v", err)
		http.Error(w, "Failed to retrieve audit events", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  result,
	})
}

// handleCreateAuditEvent handles POST requests to create audit events
func (srv *Server) handleCreateAuditEvent(w http.ResponseWriter, r *http.Request) {
	var event audit.AuditEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("Error decoding audit event: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if event.EventType == "" {
		http.Error(w, "Event type is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	clientIP := srv.getClientIP(r)
	if event.IPAddress == nil {
		event.IPAddress = &clientIP
	}
	userAgent := r.UserAgent()
	if event.UserAgent == nil {
		event.UserAgent = &userAgent
	}
	if event.Details == nil {
		event.Details = make(map[string]interface{})
	}
	if event.Outcome == "" {
		event.Outcome = audit.OutcomeSuccess
	}
	if event.Severity == "" {
		event.Severity = audit.SeverityInfo
	}

	// Create audit service
	auditService := audit.NewAuditService(srv.db)

	// Log event
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := auditService.RecordEvent(ctx, &event); err != nil {
		log.Printf("Error logging audit event: %v", err)
		http.Error(w, "Failed to log audit event", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"event_id": event.LogID,
		"message":  "Audit event logged successfully",
	})
}

// AuditSummaryHandler handles audit summary and statistics
func (srv *Server) AuditSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create audit service
	auditService := audit.NewAuditService(srv.db)

	// Get event types
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	eventTypes, err := auditService.GetEventTypes(ctx)
	if err != nil {
		log.Printf("Error getting event types: %v", err)
		http.Error(w, "Failed to retrieve audit summary", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"event_types": eventTypes,
		"message":     "Audit summary retrieved successfully",
	})
}

// QuotaHandler handles quota management and checking
func (srv *Server) QuotaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("QuotaHandler called - Method: %s", r.Method)

	switch r.Method {
	case http.MethodGet:
		srv.handleGetQuotaUsage(w, r)
	case http.MethodPost:
		srv.handleConsumeQuota(w, r)
	case http.MethodPut:
		srv.handleUpdateQuota(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetQuotaUsage handles GET requests to retrieve quota usage
func (srv *Server) handleGetQuotaUsage(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	if entityType == "" || entityID == "" {
		http.Error(w, "Entity type and entity ID are required", http.StatusBadRequest)
		return
	}

	// Create quota service
	quotaService := quota.NewQuotaService(srv.db)

	// Get quota usage
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	usage, err := quotaService.GetQuotaUsage(ctx, entityType, entityID)
	if err != nil {
		log.Printf("Error getting quota usage: %v", err)
		http.Error(w, "Failed to retrieve quota usage", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"usage":   usage,
	})
}

// handleConsumeQuota handles POST requests to consume quota
func (srv *Server) handleConsumeQuota(w http.ResponseWriter, r *http.Request) {
	var req quota.QuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding quota request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.EntityType == "" || req.EntityID == "" || req.QuotaType == "" || req.Amount <= 0 {
		http.Error(w, "Entity type, entity ID, quota type, and positive amount are required", http.StatusBadRequest)
		return
	}

	// Create quota service
	quotaService := quota.NewQuotaService(srv.db)

	// Consume quota
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	check, err := quotaService.ConsumeQuota(ctx, &req)
	if err != nil {
		log.Printf("Error consuming quota: %v", err)

		// If quota exceeded, return 429 Too Many Requests
		if !check.Allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Quota exceeded",
				"quota":   check,
			})
			return
		}

		http.Error(w, "Failed to consume quota", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"quota":   check,
		"message": "Quota consumed successfully",
	})
}

// handleUpdateQuota handles PUT requests to update quota limits
func (srv *Server) handleUpdateQuota(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QuotaID string `json:"quota_id"`
		Limit   int64  `json:"limit"`
		Period  string `json:"period"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding quota update request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.QuotaID == "" || req.Limit <= 0 || req.Period == "" {
		http.Error(w, "Quota ID, positive limit, and period are required", http.StatusBadRequest)
		return
	}

	// Create quota service
	quotaService := quota.NewQuotaService(srv.db)

	// Update quota
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := quotaService.UpdateQuota(ctx, req.QuotaID, req.Limit, req.Period); err != nil {
		log.Printf("Error updating quota: %v", err)
		http.Error(w, "Failed to update quota", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Quota updated successfully",
	})
}

// RetryTaskHandler handles retry task management
func (srv *Server) RetryTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("RetryTaskHandler called - Method: %s", r.Method)

	switch r.Method {
	case http.MethodGet:
		srv.handleGetRetryTaskStatus(w, r)
	case http.MethodPost:
		srv.handleScheduleRetryTask(w, r)
	case http.MethodDelete:
		srv.handleCancelRetryTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetRetryTaskStatus handles GET requests to retrieve retry task status
func (srv *Server) handleGetRetryTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Create retry service
	retryService := retry.NewRetryService(srv.db)

	// Get task status
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	task, err := retryService.GetTaskStatus(ctx, taskID)
	if err != nil {
		log.Printf("Error getting retry task status: %v", err)
		http.Error(w, "Failed to retrieve task status", http.StatusInternalServerError)
		return
	}

	// Get task attempts
	attempts, err := retryService.GetTaskAttempts(ctx, taskID)
	if err != nil {
		log.Printf("Error getting retry task attempts: %v", err)
		// Continue without attempts
		attempts = []*retry.RetryAttempt{}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"task":     task,
		"attempts": attempts,
	})
}

// handleScheduleRetryTask handles POST requests to schedule retry tasks
func (srv *Server) handleScheduleRetryTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskType    string                 `json:"task_type"`
		EntityID    string                 `json:"entity_id"`
		Payload     map[string]interface{} `json:"payload"`
		MaxAttempts int                    `json:"max_attempts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding retry task request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.TaskType == "" || req.EntityID == "" {
		http.Error(w, "Task type and entity ID are required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.MaxAttempts <= 0 {
		req.MaxAttempts = 3
	}
	if req.Payload == nil {
		req.Payload = make(map[string]interface{})
	}

	// Get retry configuration
	configs := retry.GetDefaultRetryConfigs()
	config, exists := configs[req.TaskType]
	if !exists {
		config = &retry.RetryConfig{
			TaskType:          req.TaskType,
			MaxAttempts:       req.MaxAttempts,
			InitialDelay:      time.Second,
			MaxDelay:          time.Minute * 5,
			BackoffMultiplier: 2.0,
			EnableJitter:      true,
		}
	} else {
		config.MaxAttempts = req.MaxAttempts
	}

	// Create retry service
	retryService := retry.NewRetryService(srv.db)

	// Schedule task
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	task, err := retryService.ScheduleTask(ctx, req.TaskType, req.EntityID, req.Payload, config)
	if err != nil {
		log.Printf("Error scheduling retry task: %v", err)
		http.Error(w, "Failed to schedule retry task", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"task":    task,
		"message": "Retry task scheduled successfully",
	})
}

// handleCancelRetryTask handles DELETE requests to cancel retry tasks
func (srv *Server) handleCancelRetryTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Create retry service
	retryService := retry.NewRetryService(srv.db)

	// Cancel task
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := retryService.CancelTask(ctx, taskID); err != nil {
		log.Printf("Error cancelling retry task: %v", err)
		http.Error(w, "Failed to cancel retry task", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Retry task cancelled successfully",
	})
}

// SystemHealthHandler handles system health and performance metrics
func (srv *Server) SystemHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return basic system health information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(time.Now().Add(-time.Hour)), // Placeholder
		"version":   "1.0.0",
		"components": map[string]string{
			"database": "healthy",
			"audit":    "healthy",
			"retry":    "healthy",
			"quota":    "healthy",
		},
	})
}
