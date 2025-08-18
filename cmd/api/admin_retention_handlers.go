package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/email"
)

// AdminRetentionResponse represents the response for admin retention queries
type AdminRetentionResponse struct {
	Emails     []email.EmailRetentionInfo `json:"emails"`
	TotalCount int                        `json:"total_count"`
	Limit      int                        `json:"limit"`
	Offset     int                        `json:"offset"`
	HasMore    bool                       `json:"has_more"`
}

// AdminRetentionStatsResponse represents the response for retention statistics
type AdminRetentionStatsResponse struct {
	Statistics map[string]interface{} `json:"statistics"`
	Summary    map[string]interface{} `json:"summary"`
}

// SetEmailExpirationRequest represents a request to set email expiration
type SetEmailExpirationRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
}

// adminRetentionQueryHandler handles GET /api/admin/email/retention
// Returns emails pending cleanup with filtering and pagination
func (srv *Server) adminRetentionQueryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionQueryHandler started - Admin Retention Query (Micro-Iteration 4.24)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention query requested by user: %s", userID)

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	userIDFilter := r.URL.Query().Get("user_id")
	statusFilter := r.URL.Query().Get("status")
	startDateFilter := r.URL.Query().Get("start_date")
	endDateFilter := r.URL.Query().Get("end_date")

	// Set defaults and validate
	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build filters
	filters := make(map[string]string)
	if userIDFilter != "" {
		filters["user_id"] = userIDFilter
	}
	if statusFilter != "" {
		filters["status"] = statusFilter
	}
	if startDateFilter != "" {
		filters["start_date"] = startDateFilter
	}
	if endDateFilter != "" {
		filters["end_date"] = endDateFilter
	}

	// Get retention service
	if srv.emailRetentionService == nil {
		log.Printf("Email retention service not initialized")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Retention service not available"}`))
		return
	}

	// Get emails pending cleanup
	emails, err := srv.emailRetentionService.GetEmailsPendingCleanup(r.Context(), filters, limit, offset)
	if err != nil {
		log.Printf("Failed to get emails pending cleanup: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve emails"}`))
		return
	}

	// Get total count for pagination
	totalCount, err := srv.emailRetentionService.GetEmailsPendingCleanupCount(r.Context(), filters)
	if err != nil {
		log.Printf("Failed to get total count: %v", err)
		totalCount = len(emails) // Fallback to current batch size
	}

	// Determine if there are more results
	hasMore := (offset + limit) < totalCount

	response := AdminRetentionResponse{
		Emails:     emails,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminRetentionStatsHandler handles GET /api/admin/email/retention-stats
// Returns comprehensive retention statistics
func (srv *Server) adminRetentionStatsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionStatsHandler started - Admin Retention Statistics (Micro-Iteration 4.24)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention stats requested by user: %s", userID)

	// Get retention service
	if srv.emailRetentionService == nil {
		log.Printf("Email retention service not initialized")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Retention service not available"}`))
		return
	}

	// Get retention statistics
	statistics, err := srv.emailRetentionService.GetRetentionStatistics(r.Context())
	if err != nil {
		log.Printf("Failed to get retention statistics: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve statistics"}`))
		return
	}

	// Create summary
	summary := map[string]interface{}{
		"emails_pending_deletion": statistics["emails_pending_deletion"],
		"total_emails":            statistics["total_emails"],
		"soft_deleted_emails":     statistics["soft_deleted_emails"],
		"emails_expiring_soon":    statistics["emails_expiring_soon"],
		"cleanup_configuration": map[string]interface{}{
			"default_expiration_days": statistics["default_expiration_days"],
			"cleanup_audit_logs":      statistics["cleanup_audit_logs"],
			"enable_notifications":    statistics["enable_notifications"],
			"batch_size":              statistics["batch_size"],
		},
	}

	response := AdminRetentionStatsResponse{
		Statistics: statistics,
		Summary:    summary,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminSetEmailExpirationHandler handles POST /api/admin/email/{email_id}/expiration
// Allows setting expiration time for a specific email
func (srv *Server) adminSetEmailExpirationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminSetEmailExpirationHandler started - Admin Set Email Expiration (Micro-Iteration 4.24)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Extract email ID from URL
	emailID := r.URL.Query().Get("email_id")
	if emailID == "" {
		log.Printf("Email ID not provided")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Email ID is required"}`))
		return
	}

	log.Printf("Admin set expiration requested by user: %s for email: %s", userID, emailID)

	// Parse request body
	var req SetEmailExpirationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request body"}`))
		return
	}

	// Get retention service
	if srv.emailRetentionService == nil {
		log.Printf("Email retention service not initialized")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Retention service not available"}`))
		return
	}

	// Set email expiration
	err := srv.emailRetentionService.SetEmailExpiration(emailID, userID, req.ExpiresAt)
	if err != nil {
		log.Printf("Failed to set email expiration: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to set expiration"}`))
		return
	}

	response := map[string]interface{}{
		"message":    "Email expiration updated successfully",
		"email_id":   emailID,
		"expires_at": req.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminManualRetentionCleanupHandler handles POST /api/admin/email/retention-cleanup
// Allows manual execution of the retention cleanup process
func (srv *Server) adminManualRetentionCleanupHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminManualRetentionCleanupHandler started - Admin Manual Retention Cleanup (Micro-Iteration 4.24)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin manual retention cleanup requested by user: %s", userID)

	// Parse request body for dry run option
	var req struct {
		DryRun bool `json:"dry_run,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to false if no body provided
		req.DryRun = false
	}

	// Get retention service
	if srv.emailRetentionService == nil {
		log.Printf("Email retention service not initialized")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Retention service not available"}`))
		return
	}

	// Get stats before cleanup
	beforeStats, err := srv.emailRetentionService.GetRetentionStatistics(r.Context())
	if err != nil {
		log.Printf("Failed to get before stats: %v", err)
		beforeStats = make(map[string]interface{})
	}

	var result map[string]interface{}

	if req.DryRun {
		log.Printf("Performing dry run retention cleanup...")

		// For dry run, just get the emails that would be deleted
		emails, err := srv.emailRetentionService.GetEmailsPendingCleanup(r.Context(), map[string]string{}, 1000, 0)
		if err != nil {
			log.Printf("Failed to get emails for dry run: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to perform dry run"}`))
			return
		}

		result = map[string]interface{}{
			"dry_run":                 true,
			"message":                 "Dry run completed - no emails were actually deleted",
			"stats_before":            beforeStats,
			"emails_would_be_deleted": len(emails),
			"sample_emails":           emails[:min(len(emails), 10)], // Show first 10 emails
		}
	} else {
		log.Printf("Performing manual retention cleanup...")

		// Perform actual cleanup
		err = srv.emailRetentionService.PerformCleanup(r.Context())
		if err != nil {
			log.Printf("Manual retention cleanup failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Cleanup failed"}`))
			return
		}

		// Get stats after cleanup
		afterStats, err := srv.emailRetentionService.GetRetentionStatistics(r.Context())
		if err != nil {
			log.Printf("Failed to get after stats: %v", err)
			afterStats = make(map[string]interface{})
		}

		// Get cleanup stats
		cleanupStats := srv.emailRetentionService.GetCleanupStats()

		result = map[string]interface{}{
			"dry_run":       false,
			"message":       "Manual retention cleanup completed successfully",
			"stats_before":  beforeStats,
			"stats_after":   afterStats,
			"cleanup_stats": cleanupStats,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
