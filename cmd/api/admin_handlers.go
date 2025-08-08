package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/cleanup"
)

// EmailRetentionStats represents statistics about email retention
type EmailRetentionStats struct {
	ExpiredEmails          int        `json:"expired_emails"`
	BurnAfterReadEmails    int        `json:"burn_after_read_emails"`
	TotalEmailsWithContent int        `json:"total_emails_with_content"`
	CleanupIntervalMinutes int        `json:"cleanup_interval_minutes"`
	LastCleanupRun         *time.Time `json:"last_cleanup_run,omitempty"`
	NextCleanupRun         *time.Time `json:"next_cleanup_run,omitempty"`
}

// adminEmailRetentionStatsHandler handles GET /admin/email-retention-stats
// Returns counts of expired, deleted, and burn-after-read emails
func (srv *Server) adminEmailRetentionStatsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminEmailRetentionStatsHandler started")

	// Check if user is admin (simple check for demo - in production, use proper admin authentication)
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// For demo purposes, allow any authenticated user to access admin stats
	// In production, check if user has admin role
	log.Printf("Admin stats requested by user: %s", userID)

	// Get cleanup interval from environment
	cleanupIntervalStr := os.Getenv("EMAIL_CLEANUP_INTERVAL_MINUTES")
	if cleanupIntervalStr == "" {
		cleanupIntervalStr = "15" // Default
	}
	cleanupInterval, err := strconv.Atoi(cleanupIntervalStr)
	if err != nil {
		cleanupInterval = 15 // Default fallback
	}

	// Count expired emails
	var expiredCount int
	err = srv.db.QueryRow(`
		SELECT COUNT(*) FROM emails 
		WHERE expires_at IS NOT NULL AND expires_at <= datetime('now') AND encrypted_blob_url IS NOT NULL
	`).Scan(&expiredCount)
	if err != nil {
		log.Printf("Failed to count expired emails: %v", err)
		expiredCount = 0
	}

	// Count burn-after-read emails that have been accessed
	var burnAfterReadCount int
	err = srv.db.QueryRow(`
		SELECT COUNT(*) FROM emails 
		WHERE burn_after_read = 1 AND access_count > 0 AND encrypted_blob_url IS NOT NULL
	`).Scan(&burnAfterReadCount)
	if err != nil {
		log.Printf("Failed to count burn-after-read emails: %v", err)
		burnAfterReadCount = 0
	}

	// Count total emails with content
	var totalWithContent int
	err = srv.db.QueryRow(`
		SELECT COUNT(*) FROM emails WHERE encrypted_blob_url IS NOT NULL
	`).Scan(&totalWithContent)
	if err != nil {
		log.Printf("Failed to count total emails: %v", err)
		totalWithContent = 0
	}

	// Count emails that have been soft-deleted (no content)
	var softDeletedCount int
	err = srv.db.QueryRow(`
		SELECT COUNT(*) FROM emails WHERE encrypted_blob_url IS NULL
	`).Scan(&softDeletedCount)
	if err != nil {
		log.Printf("Failed to count soft-deleted emails: %v", err)
		softDeletedCount = 0
	}

	// Calculate cleanup timing
	now := time.Now()
	nextCleanup := now.Add(time.Duration(cleanupInterval) * time.Minute)

	stats := EmailRetentionStats{
		ExpiredEmails:          expiredCount,
		BurnAfterReadEmails:    burnAfterReadCount,
		TotalEmailsWithContent: totalWithContent,
		CleanupIntervalMinutes: cleanupInterval,
		NextCleanupRun:         &nextCleanup,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": stats,
		"summary": map[string]interface{}{
			"emails_pending_deletion": expiredCount + burnAfterReadCount,
			"total_emails":            totalWithContent + softDeletedCount,
			"soft_deleted_emails":     softDeletedCount,
		},
	})
}

// ManualCleanupRequest represents a request to run cleanup manually
type ManualCleanupRequest struct {
	DryRun bool `json:"dry_run,omitempty"`
}

// adminManualCleanupHandler handles POST /admin/manual-cleanup
// Allows manual execution of the email cleanup process
func (srv *Server) adminManualCleanupHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminManualCleanupHandler started")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Manual cleanup requested by user: %s", userID)

	// Parse request body
	var req ManualCleanupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request body"}`))
		return
	}

	// Create a temporary worker for manual cleanup
	cleanupInterval := 15 // Default interval for manual cleanup
	worker, err := cleanup.NewEmailCleanupWorkerWithDB(srv.db, cleanupInterval)
	if err != nil {
		log.Printf("Failed to create cleanup worker: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to initialize cleanup worker"}`))
		return
	}

	// Get stats before cleanup
	beforeStats, err := worker.GetCleanupStats()
	if err != nil {
		log.Printf("Failed to get before stats: %v", err)
		beforeStats = make(map[string]interface{})
	}

	// Perform cleanup
	var result map[string]interface{}
	if req.DryRun {
		log.Printf("Performing dry run cleanup...")
		result = map[string]interface{}{
			"dry_run":      true,
			"message":      "Dry run completed - no emails were actually deleted",
			"stats_before": beforeStats,
		}
	} else {
		log.Printf("Performing manual cleanup...")
		err = worker.RunCleanupOnce()
		if err != nil {
			log.Printf("Manual cleanup failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Cleanup failed"}`))
			return
		}

		// Get stats after cleanup
		afterStats, err := worker.GetCleanupStats()
		if err != nil {
			log.Printf("Failed to get after stats: %v", err)
			afterStats = make(map[string]interface{})
		}

		result = map[string]interface{}{
			"dry_run":      false,
			"message":      "Manual cleanup completed successfully",
			"stats_before": beforeStats,
			"stats_after":  afterStats,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
