// =============================================================================
// SECURE EMAIL MVP - NOTIFICATION HANDLERS
// =============================================================================
// Handlers for notification preferences and access event history.
// Enhanced with delivery frequency controls and rate limiting (Micro-Iteration 4.18).
// =============================================================================

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/notification"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// NotificationPreferencesRequest represents the request to update notification preferences
type NotificationPreferencesRequest struct {
	EmailNotifications        bool   `json:"email_notifications"`
	SMSNotifications          bool   `json:"sms_notifications"`
	NotifyOnSuccess           bool   `json:"notify_on_success"`
	NotifyOnFailure           bool   `json:"notify_on_failure"`
	NotifyOnBlocked           bool   `json:"notify_on_blocked"`
	IncludeGeolocation        bool   `json:"include_geolocation"`
	IncludeDeviceInfo         bool   `json:"include_device_info"`
	DeliveryFrequency         string `json:"delivery_frequency"`
	ThresholdAttempts         int    `json:"threshold_attempts"`
	RateLimitWindowMinutes    int    `json:"rate_limit_window_minutes"`
	RateLimitMaxNotifications int    `json:"rate_limit_max_notifications"`
}

// EmailNotificationPreferencesRequest represents the request to update per-email notification preferences
type EmailNotificationPreferencesRequest struct {
	DeliveryFrequency         string `json:"delivery_frequency"`
	ThresholdAttempts         int    `json:"threshold_attempts"`
	RateLimitWindowMinutes    int    `json:"rate_limit_window_minutes"`
	RateLimitMaxNotifications int    `json:"rate_limit_max_notifications"`
	InheritGlobalSettings     bool   `json:"inherit_global_settings"`
}

// NotificationStatsResponse represents notification statistics
type NotificationStatsResponse struct {
	TotalEvents       int            `json:"total_events"`
	SuppressedEvents  int            `json:"suppressed_events"`
	SuppressionStats  map[string]int `json:"suppression_stats"`
	DeliveryFrequency string         `json:"delivery_frequency"`
	RateLimitInfo     struct {
		WindowMinutes    int `json:"window_minutes"`
		MaxNotifications int `json:"max_notifications"`
	} `json:"rate_limit_info"`
}

// getNotificationPreferencesHandler handles GET /api/notifications/preferences
func (srv *Server) getNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get user preferences
	prefs, err := notificationService.GetNotificationPreferences(ctx, userID)
	if err != nil {
		log.Printf("Failed to get notification preferences: %v", err)
		http.Error(w, "Failed to get notification preferences", http.StatusInternalServerError)
		return
	}

	// Return preferences
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// updateNotificationPreferencesHandler handles PUT /api/notifications/preferences
func (srv *Server) updateNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Parse request body
	var req NotificationPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate delivery frequency
	if !isValidDeliveryFrequency(req.DeliveryFrequency) {
		http.Error(w, "Invalid delivery frequency", http.StatusBadRequest)
		return
	}

	// Validate threshold attempts
	if req.ThresholdAttempts < 1 || req.ThresholdAttempts > 10 {
		http.Error(w, "Threshold attempts must be between 1 and 10", http.StatusBadRequest)
		return
	}

	// Validate rate limit settings
	if req.RateLimitWindowMinutes < 1 || req.RateLimitWindowMinutes > 1440 {
		http.Error(w, "Rate limit window must be between 1 and 1440 minutes", http.StatusBadRequest)
		return
	}

	if req.RateLimitMaxNotifications < 1 || req.RateLimitMaxNotifications > 100 {
		http.Error(w, "Rate limit max notifications must be between 1 and 100", http.StatusBadRequest)
		return
	}

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get current preferences to preserve timestamps
	currentPrefs, err := notificationService.GetNotificationPreferences(ctx, userID)
	if err != nil {
		log.Printf("Failed to get current notification preferences: %v", err)
		http.Error(w, "Failed to get current preferences", http.StatusInternalServerError)
		return
	}

	// Update preferences
	updatedPrefs := &notification.NotificationPreferences{
		UserID:                    userID,
		EmailNotifications:        req.EmailNotifications,
		SMSNotifications:          req.SMSNotifications,
		NotifyOnSuccess:           req.NotifyOnSuccess,
		NotifyOnFailure:           req.NotifyOnFailure,
		NotifyOnBlocked:           req.NotifyOnBlocked,
		IncludeGeolocation:        req.IncludeGeolocation,
		IncludeDeviceInfo:         req.IncludeDeviceInfo,
		DeliveryFrequency:         notification.DeliveryFrequency(req.DeliveryFrequency),
		ThresholdAttempts:         req.ThresholdAttempts,
		RateLimitWindowMinutes:    req.RateLimitWindowMinutes,
		RateLimitMaxNotifications: req.RateLimitMaxNotifications,
		CreatedAt:                 currentPrefs.CreatedAt,
		UpdatedAt:                 time.Now(),
	}

	// Save updated preferences
	if err := notificationService.UpdateNotificationPreferences(ctx, updatedPrefs); err != nil {
		log.Printf("Failed to update notification preferences: %v", err)
		http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	// Return updated preferences
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPrefs)
}

// getEmailNotificationPreferencesHandler handles GET /api/notifications/email/{emailID}/preferences
func (srv *Server) getEmailNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Extract email ID from URL
	vars := mux.Vars(r)
	emailID := vars["emailID"]

	// Verify user owns this email
	if err := srv.verifyEmailOwnership(ctx, emailID, userID); err != nil {
		http.Error(w, "Email not found or access denied", http.StatusNotFound)
		return
	}

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get email-specific preferences
	prefs, err := notificationService.GetEmailNotificationPreferences(ctx, emailID)
	if err != nil {
		log.Printf("Failed to get email notification preferences: %v", err)
		http.Error(w, "Failed to get email notification preferences", http.StatusInternalServerError)
		return
	}

	// If no email-specific preferences exist, return default structure
	if prefs == nil {
		prefs = &notification.EmailNotificationPreferences{
			EmailID:                   emailID,
			UserID:                    userID,
			DeliveryFrequency:         notification.DeliveryFrequencyImmediate,
			ThresholdAttempts:         3,
			RateLimitWindowMinutes:    15,
			RateLimitMaxNotifications: 5,
			InheritGlobalSettings:     true,
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		}
	}

	// Return preferences
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// updateEmailNotificationPreferencesHandler handles PUT /api/notifications/email/{emailID}/preferences
func (srv *Server) updateEmailNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Extract email ID from URL
	vars := mux.Vars(r)
	emailID := vars["emailID"]

	// Verify user owns this email
	if err := srv.verifyEmailOwnership(ctx, emailID, userID); err != nil {
		http.Error(w, "Email not found or access denied", http.StatusNotFound)
		return
	}

	// Parse request body
	var req EmailNotificationPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate delivery frequency
	if !isValidDeliveryFrequency(req.DeliveryFrequency) {
		http.Error(w, "Invalid delivery frequency", http.StatusBadRequest)
		return
	}

	// Validate threshold attempts
	if req.ThresholdAttempts < 1 || req.ThresholdAttempts > 10 {
		http.Error(w, "Threshold attempts must be between 1 and 10", http.StatusBadRequest)
		return
	}

	// Validate rate limit settings
	if req.RateLimitWindowMinutes < 1 || req.RateLimitWindowMinutes > 1440 {
		http.Error(w, "Rate limit window must be between 1 and 1440 minutes", http.StatusBadRequest)
		return
	}

	if req.RateLimitMaxNotifications < 1 || req.RateLimitMaxNotifications > 100 {
		http.Error(w, "Rate limit max notifications must be between 1 and 100", http.StatusBadRequest)
		return
	}

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get current preferences to preserve timestamps
	currentPrefs, err := notificationService.GetEmailNotificationPreferences(ctx, emailID)
	if err != nil {
		log.Printf("Failed to get current email notification preferences: %v", err)
		http.Error(w, "Failed to get current preferences", http.StatusInternalServerError)
		return
	}

	// Update preferences
	updatedPrefs := &notification.EmailNotificationPreferences{
		EmailID:                   emailID,
		UserID:                    userID,
		DeliveryFrequency:         notification.DeliveryFrequency(req.DeliveryFrequency),
		ThresholdAttempts:         req.ThresholdAttempts,
		RateLimitWindowMinutes:    req.RateLimitWindowMinutes,
		RateLimitMaxNotifications: req.RateLimitMaxNotifications,
		InheritGlobalSettings:     req.InheritGlobalSettings,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	// If current preferences exist, preserve creation timestamp
	if currentPrefs != nil {
		updatedPrefs.CreatedAt = currentPrefs.CreatedAt
	}

	// Save updated preferences
	if err := notificationService.UpdateEmailNotificationPreferences(ctx, updatedPrefs); err != nil {
		log.Printf("Failed to update email notification preferences: %v", err)
		http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	// Return updated preferences
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPrefs)
}

// getAccessEventHistoryHandler handles GET /api/notifications/history
func (srv *Server) getAccessEventHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 200 {
			limit = parsedLimit
		}
	}

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get access event history
	events, err := notificationService.GetAccessEventHistory(ctx, userID, limit)
	if err != nil {
		log.Printf("Failed to get access event history: %v", err)
		http.Error(w, "Failed to get access event history", http.StatusInternalServerError)
		return
	}

	// Return events
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// getNotificationSuppressionsHandler handles GET /api/notifications/suppressions
func (srv *Server) getNotificationSuppressionsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 200 {
			limit = parsedLimit
		}
	}

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get notification suppressions
	suppressions, err := notificationService.GetNotificationSuppressions(ctx, userID, limit)
	if err != nil {
		log.Printf("Failed to get notification suppressions: %v", err)
		http.Error(w, "Failed to get notification suppressions", http.StatusInternalServerError)
		return
	}

	// Return suppressions
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppressions)
}

// getNotificationStatsHandler handles GET /api/notifications/stats
func (srv *Server) getNotificationStatsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get user preferences for current settings
	prefs, err := notificationService.GetNotificationPreferences(ctx, userID)
	if err != nil {
		log.Printf("Failed to get notification preferences: %v", err)
		http.Error(w, "Failed to get notification preferences", http.StatusInternalServerError)
		return
	}

	// Get suppression statistics
	suppressionStats, err := notificationService.GetSuppressionStats(ctx, userID)
	if err != nil {
		log.Printf("Failed to get suppression stats: %v", err)
		http.Error(w, "Failed to get suppression statistics", http.StatusInternalServerError)
		return
	}

	// Calculate total suppressed events
	totalSuppressed := 0
	for _, count := range suppressionStats {
		totalSuppressed += count
	}

	// Get total events count
	var totalEvents int
	query := `SELECT COUNT(*) FROM access_events WHERE user_id = ?`
	err = srv.db.QueryRowContext(ctx, query, userID).Scan(&totalEvents)
	if err != nil {
		log.Printf("Failed to get total events count: %v", err)
		totalEvents = 0
	}

	// Build response
	response := NotificationStatsResponse{
		TotalEvents:       totalEvents,
		SuppressedEvents:  totalSuppressed,
		SuppressionStats:  suppressionStats,
		DeliveryFrequency: string(prefs.DeliveryFrequency),
	}
	response.RateLimitInfo.WindowMinutes = prefs.RateLimitWindowMinutes
	response.RateLimitInfo.MaxNotifications = prefs.RateLimitMaxNotifications

	// Return statistics
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// recordAccessEvent is a helper function to record access events and send notifications
func (srv *Server) recordAccessEvent(ctx context.Context, emailID, userID, ipAddress, userAgent, country, city, deviceType, failureReason string, eventType notification.AccessEventType) error {
	// Create access event
	event := &notification.AccessEvent{
		EventID:       uuid.New().String(),
		EmailID:       emailID,
		UserID:        userID,
		EventType:     eventType,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Country:       country,
		City:          city,
		DeviceType:    deviceType,
		FailureReason: failureReason,
		Timestamp:     time.Now(),
	}

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Record the access event
	if err := notificationService.RecordAccessEvent(ctx, event); err != nil {
		log.Printf("Failed to record access event: %v", err)
		return err
	}

	// Process suspicious access detection (Micro-Iteration 4.18)
	if srv.suspiciousService != nil {
		if err := srv.suspiciousService.ProcessAccessEvent(ctx, emailID, userID, ipAddress, userAgent, country, city, deviceType, failureReason, string(eventType)); err != nil {
			log.Printf("Failed to process suspicious access detection: %v", err)
			// Don't return error as the access event was already recorded
		}
	}

	// Get user notification preferences
	prefs, err := notificationService.GetNotificationPreferences(ctx, userID)
	if err != nil {
		log.Printf("Failed to get notification preferences: %v", err)
		// Continue without sending notification
		return nil
	}

	// Send notification (this will handle delivery controls and rate limiting)
	if err := notificationService.SendNotification(ctx, event, prefs); err != nil {
		log.Printf("Failed to send notification: %v", err)
		// Don't return error as the access event was already recorded
	}

	return nil
}

// verifyEmailOwnership verifies that a user owns a specific email
func (srv *Server) verifyEmailOwnership(ctx context.Context, emailID, userID string) error {
	query := `SELECT COUNT(*) FROM emails WHERE email_id = ? AND sender_id = ?`
	var count int
	err := srv.db.QueryRowContext(ctx, query, emailID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// isValidDeliveryFrequency validates delivery frequency values
func isValidDeliveryFrequency(frequency string) bool {
	validFrequencies := []string{
		string(notification.DeliveryFrequencyImmediate),
		string(notification.DeliveryFrequencyDailyDigest),
		string(notification.DeliveryFrequencyFirstAttemptOnly),
		string(notification.DeliveryFrequencyThresholdTrigger),
	}

	for _, valid := range validFrequencies {
		if frequency == valid {
			return true
		}
	}
	return false
}

// getDailyDigestHistoryHandler handles GET /api/notifications/digest/history
func (srv *Server) getDailyDigestHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get digest history
	deliveries, err := notificationService.GetDailyDigestHistory(ctx, userID, limit)
	if err != nil {
		log.Printf("Failed to get daily digest history: %v", err)
		http.Error(w, "Failed to get digest history", http.StatusInternalServerError)
		return
	}

	// Return digest history
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deliveries)
}

// generateDailyDigestHandler handles POST /api/notifications/digest/generate
func (srv *Server) generateDailyDigestHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	ctx := r.Context()

	// Parse request body
	var req struct {
		Date string `json:"date"` // Optional date in YYYY-MM-DD format
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Determine digest date
	var digestDate time.Time
	if req.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		digestDate = parsedDate
	} else {
		// Default to yesterday
		digestDate = time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	}

	// Get notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Generate digest summary
	summary, err := notificationService.GenerateDailyDigest(ctx, userID, digestDate)
	if err != nil {
		log.Printf("Failed to generate daily digest: %v", err)
		http.Error(w, "Failed to generate digest", http.StatusInternalServerError)
		return
	}

	// Return digest summary
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
