// =============================================================================
// SECURE EMAIL MVP - AUDIT LOG HTTP HANDLERS
// =============================================================================
// Package main provides HTTP handlers for the audit log system.
// =============================================================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"secure-email-mvp/pkg/audit"

	"github.com/gorilla/mux"
)

// AuditLogQueryRequest represents a request to query audit logs
type AuditLogQueryRequest struct {
	Filter   audit.AuditLogFilter `json:"filter"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// ExportRequest represents a request to create an export
type ExportRequest struct {
	ExportType string               `json:"export_type"`
	Filter     audit.AuditLogFilter `json:"filter"`
}

// SavedFilterRequest represents a request to save a filter
type SavedFilterRequest struct {
	FilterName string               `json:"filter_name"`
	Filter     audit.AuditLogFilter `json:"filter"`
	IsDefault  bool                 `json:"is_default"`
}

// RetentionPolicyRequest represents a request to update retention policy
type RetentionPolicyRequest struct {
	RetentionDays int  `json:"retention_days"`
	AutoPurge     bool `json:"auto_purge"`
}

// getAuditLogsHandler handles GET /api/audit/logs
func (srv *Server) getAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	// Build filter from query parameters
	filter := audit.AuditLogFilter{}

	// Date range
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if parsed, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filter.DateFrom = &parsed
		}
	}
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if parsed, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filter.DateTo = &parsed
		}
	}

	// Event types
	if eventTypes := r.URL.Query().Get("event_types"); eventTypes != "" {
		eventTypeStrings := splitAndTrim(eventTypes, ",")
		filter.EventTypes = make([]audit.EventType, len(eventTypeStrings))
		for i, et := range eventTypeStrings {
			filter.EventTypes[i] = audit.EventType(et)
		}
	}

	// Outcomes
	if outcomes := r.URL.Query().Get("outcomes"); outcomes != "" {
		outcomeStrings := splitAndTrim(outcomes, ",")
		filter.Outcomes = make([]audit.Outcome, len(outcomeStrings))
		for i, o := range outcomeStrings {
			filter.Outcomes[i] = audit.Outcome(o)
		}
	}

	// Severities
	if severities := r.URL.Query().Get("severities"); severities != "" {
		severityStrings := splitAndTrim(severities, ",")
		filter.Severities = make([]audit.Severity, len(severityStrings))
		for i, s := range severityStrings {
			filter.Severities[i] = audit.Severity(s)
		}
	}

	// User IDs (for admin access)
	if userIDs := r.URL.Query().Get("user_ids"); userIDs != "" {
		filter.UserIDs = splitAndTrim(userIDs, ",")
	} else {
		// For regular users, only show their own events
		filter.UserIDs = []string{userID}
	}

	// Other filters
	if ipAddresses := r.URL.Query().Get("ip_addresses"); ipAddresses != "" {
		filter.IPAddresses = splitAndTrim(ipAddresses, ",")
	}
	if relatedEmailIDs := r.URL.Query().Get("related_email_ids"); relatedEmailIDs != "" {
		filter.RelatedEmailIDs = splitAndTrim(relatedEmailIDs, ",")
	}
	if sessionIDs := r.URL.Query().Get("session_ids"); sessionIDs != "" {
		filter.SessionIDs = splitAndTrim(sessionIDs, ",")
	}
	if countries := r.URL.Query().Get("countries"); countries != "" {
		filter.Countries = splitAndTrim(countries, ",")
	}
	if deviceTypes := r.URL.Query().Get("device_types"); deviceTypes != "" {
		filter.DeviceTypes = splitAndTrim(deviceTypes, ",")
	}
	if searchTerm := r.URL.Query().Get("search_term"); searchTerm != "" {
		filter.SearchTerm = &searchTerm
	}

	// Query audit logs
	result, err := srv.auditService.QueryEvents(r.Context(), filter, page, pageSize)
	if err != nil {
		log.Printf("Failed to query audit logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// getAuditEventTypesHandler handles GET /api/audit/event-types
func (srv *Server) getAuditEventTypesHandler(w http.ResponseWriter, r *http.Request) {
	eventTypes, err := srv.auditService.GetEventTypes(r.Context())
	if err != nil {
		log.Printf("Failed to get event types: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event_types": eventTypes,
	})
}

// getUserAuditEventsHandler handles GET /api/audit/user-events
func (srv *Server) getUserAuditEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse limit parameter
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	// Get user events
	events, err := srv.auditService.GetUserEvents(r.Context(), userID, limit)
	if err != nil {
		log.Printf("Failed to get user events: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// createExportHandler handles POST /api/audit/exports
func (srv *Server) createExportHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate export type
	if req.ExportType != "csv" && req.ExportType != "json" {
		http.Error(w, "Invalid export type. Must be 'csv' or 'json'", http.StatusBadRequest)
		return
	}

	// For regular users, only allow their own events
	if len(req.Filter.UserIDs) == 0 {
		req.Filter.UserIDs = []string{userID}
	} else {
		// Ensure user can only export their own events
		hasOwnEvents := false
		for _, uid := range req.Filter.UserIDs {
			if uid == userID {
				hasOwnEvents = true
				break
			}
		}
		if !hasOwnEvents {
			http.Error(w, "Can only export your own events", http.StatusForbidden)
			return
		}
	}

	// Create export request
	export, err := srv.exportService.CreateExportRequest(r.Context(), userID, req.ExportType, req.Filter)
	if err != nil {
		log.Printf("Failed to create export request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Process export in background
	go func() {
		if err := srv.exportService.ProcessExport(context.Background(), export.ExportID); err != nil {
			log.Printf("Failed to process export %s: %v", export.ExportID, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(export)
}

// getExportsHandler handles GET /api/audit/exports
func (srv *Server) getExportsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse limit parameter
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Get user exports
	exports, err := srv.exportService.GetUserExports(r.Context(), userID, limit)
	if err != nil {
		log.Printf("Failed to get user exports: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exports": exports,
		"total":   len(exports),
	})
}

// getExportHandler handles GET /api/audit/exports/{id}
func (srv *Server) getExportHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get export ID from URL
	vars := mux.Vars(r)
	exportID := vars["id"]

	// Get export request
	export, err := srv.exportService.GetExportRequest(r.Context(), exportID)
	if err != nil {
		log.Printf("Failed to get export request: %v", err)
		http.Error(w, "Export not found", http.StatusNotFound)
		return
	}

	// Check ownership
	if export.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(export)
}

// downloadExportHandler handles GET /api/audit/exports/{id}/download
func (srv *Server) downloadExportHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get export ID from URL
	vars := mux.Vars(r)
	exportID := vars["id"]

	// Get export request
	export, err := srv.exportService.GetExportRequest(r.Context(), exportID)
	if err != nil {
		log.Printf("Failed to get export request: %v", err)
		http.Error(w, "Export not found", http.StatusNotFound)
		return
	}

	// Check ownership
	if export.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if export is completed
	if export.Status != "completed" || export.FilePath == nil {
		http.Error(w, "Export not ready", http.StatusBadRequest)
		return
	}

	// Serve file
	http.ServeFile(w, r, *export.FilePath)
}

// deleteExportHandler handles DELETE /api/audit/exports/{id}
func (srv *Server) deleteExportHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get export ID from URL
	vars := mux.Vars(r)
	exportID := vars["id"]

	// Get export request to check ownership
	export, err := srv.exportService.GetExportRequest(r.Context(), exportID)
	if err != nil {
		log.Printf("Failed to get export request: %v", err)
		http.Error(w, "Export not found", http.StatusNotFound)
		return
	}

	// Check ownership
	if export.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete export
	if err := srv.exportService.DeleteExport(r.Context(), exportID); err != nil {
		log.Printf("Failed to delete export: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getRetentionPoliciesHandler handles GET /api/audit/retention-policies
func (srv *Server) getRetentionPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	policies, err := srv.auditService.GetRetentionPolicies(r.Context())
	if err != nil {
		log.Printf("Failed to get retention policies: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"policies": policies,
	})
}

// updateRetentionPolicyHandler handles PUT /api/audit/retention-policies/{event_type}
func (srv *Server) updateRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	// Get event type from URL
	vars := mux.Vars(r)
	eventType := vars["event_type"]

	// Parse request body
	var req RetentionPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate retention days
	if req.RetentionDays < 1 || req.RetentionDays > 3650 { // Max 10 years
		http.Error(w, "Retention days must be between 1 and 3650", http.StatusBadRequest)
		return
	}

	// Update retention policy
	policy := audit.RetentionPolicy{
		RetentionID:   fmt.Sprintf("retention_%s", eventType),
		EventType:     eventType,
		RetentionDays: req.RetentionDays,
		AutoPurge:     req.AutoPurge,
	}

	if err := srv.auditService.UpdateRetentionPolicy(r.Context(), policy); err != nil {
		log.Printf("Failed to update retention policy: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// purgeExpiredLogsHandler handles POST /api/audit/purge-expired
func (srv *Server) purgeExpiredLogsHandler(w http.ResponseWriter, r *http.Request) {
	if err := srv.auditService.PurgeExpiredLogs(r.Context()); err != nil {
		log.Printf("Failed to purge expired logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// cleanupExpiredExportsHandler handles POST /api/audit/cleanup-exports
func (srv *Server) cleanupExpiredExportsHandler(w http.ResponseWriter, r *http.Request) {
	if err := srv.exportService.CleanupExpiredExports(r.Context()); err != nil {
		log.Printf("Failed to cleanup expired exports: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to split and trim strings
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}












