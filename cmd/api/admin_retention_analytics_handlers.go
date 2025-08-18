package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/email"
)

// AdminRetentionAnalyticsResponse represents the response for retention analytics
type AdminRetentionAnalyticsResponse struct {
	Analytics   *email.RetentionAnalytics `json:"analytics"`
	Filters     email.AnalyticsFilters    `json:"filters"`
	GeneratedAt time.Time                 `json:"generated_at"`
}

// AdminRetentionAnalyticsSummaryResponse represents a summary response for analytics
type AdminRetentionAnalyticsSummaryResponse struct {
	OverallStats email.OverallRetentionStats `json:"overall_stats"`
	TopUsers     []email.UserRetentionStats  `json:"top_users"`
	RecentTrends []email.DailyRetentionStats `json:"recent_trends"`
	GeneratedAt  time.Time                   `json:"generated_at"`
}

// AdminRetentionNotificationsResponse represents the response for retention notifications
type AdminRetentionNotificationsResponse struct {
	Notifications []RetentionNotificationInfo `json:"notifications"`
	TotalCount    int                         `json:"total_count"`
	Limit         int                         `json:"limit"`
	Offset        int                         `json:"offset"`
	HasMore       bool                        `json:"has_more"`
}

// RetentionNotificationInfo represents retention notification information
type RetentionNotificationInfo struct {
	NotificationID   string     `json:"notification_id"`
	EmailID          string     `json:"email_id"`
	SenderID         string     `json:"sender_id"`
	NotificationType string     `json:"notification_type"`
	SentAt           *time.Time `json:"sent_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// adminRetentionAnalyticsHandler handles GET /api/admin/email/retention-analytics
// Returns comprehensive retention analytics with filtering
func (srv *Server) adminRetentionAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionAnalyticsHandler started - Admin Retention Analytics (Micro-Iteration 4.25)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention analytics requested by user: %s", userID)

	// Parse query parameters for filters
	filters := email.AnalyticsFilters{}

	if userIDFilter := r.URL.Query().Get("user_id"); userIDFilter != "" {
		filters.UserID = userIDFilter
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters.EndDate = endDate
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = status
	}

	// Get analytics service
	analyticsService := email.NewRetentionAnalyticsService(srv.db)

	// Get comprehensive analytics
	analytics, err := analyticsService.GetRetentionAnalytics(r.Context(), filters)
	if err != nil {
		log.Printf("Failed to get retention analytics: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve analytics"}`))
		return
	}

	response := AdminRetentionAnalyticsResponse{
		Analytics:   analytics,
		Filters:     filters,
		GeneratedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminRetentionAnalyticsSummaryHandler handles GET /api/admin/email/retention-analytics-summary
// Returns a summary of retention analytics for dashboard display
func (srv *Server) adminRetentionAnalyticsSummaryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionAnalyticsSummaryHandler started - Admin Retention Analytics Summary (Micro-Iteration 4.25)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention analytics summary requested by user: %s", userID)

	// Get analytics service
	analyticsService := email.NewRetentionAnalyticsService(srv.db)

	// Get analytics with default filters (last 30 days)
	filters := email.AnalyticsFilters{
		StartDate: time.Now().AddDate(0, 0, -30),
		EndDate:   time.Now(),
	}

	analytics, err := analyticsService.GetRetentionAnalytics(r.Context(), filters)
	if err != nil {
		log.Printf("Failed to get retention analytics summary: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve analytics summary"}`))
		return
	}

	// Get top 10 users by email count
	topUsers := analytics.UserStats
	if len(topUsers) > 10 {
		topUsers = topUsers[:10]
	}

	// Get recent trends (last 7 days)
	recentTrends := analytics.RetentionTrends.DailyStats
	if len(recentTrends) > 7 {
		recentTrends = recentTrends[:7]
	}

	response := AdminRetentionAnalyticsSummaryResponse{
		OverallStats: analytics.OverallStats,
		TopUsers:     topUsers,
		RecentTrends: recentTrends,
		GeneratedAt:  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminRetentionNotificationsHandler handles GET /api/admin/email/retention-notifications
// Returns retention notification history with filtering and pagination
func (srv *Server) adminRetentionNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionNotificationsHandler started - Admin Retention Notifications (Micro-Iteration 4.25)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention notifications requested by user: %s", userID)

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	notificationType := r.URL.Query().Get("notification_type")
	senderIDFilter := r.URL.Query().Get("sender_id")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

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

	// Build query
	query := `
		SELECT notification_id, email_id, sender_id, notification_type, sent_at, created_at
		FROM retention_notifications
		WHERE 1=1
	`
	args := []interface{}{}

	if notificationType != "" {
		query += " AND notification_type = ?"
		args = append(args, notificationType)
	}

	if senderIDFilter != "" {
		query += " AND sender_id = ?"
		args = append(args, senderIDFilter)
	}

	if startDateStr != "" {
		query += " AND created_at >= ?"
		args = append(args, startDateStr)
	}

	if endDateStr != "" {
		query += " AND created_at <= ?"
		args = append(args, endDateStr)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM retention_notifications WHERE 1=1"
	countArgs := args
	if notificationType != "" {
		countQuery += " AND notification_type = ?"
	}
	if senderIDFilter != "" {
		countQuery += " AND sender_id = ?"
	}
	if startDateStr != "" {
		countQuery += " AND created_at >= ?"
	}
	if endDateStr != "" {
		countQuery += " AND created_at <= ?"
	}

	var totalCount int
	err := srv.db.QueryRowContext(r.Context(), countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Failed to get total count: %v", err)
		totalCount = 0
	}

	// Get notifications with pagination
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := srv.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		log.Printf("Failed to query retention notifications: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve notifications"}`))
		return
	}
	defer rows.Close()

	var notifications []RetentionNotificationInfo
	for rows.Next() {
		var notification RetentionNotificationInfo
		var sentAtStr sql.NullString

		err := rows.Scan(
			&notification.NotificationID, &notification.EmailID, &notification.SenderID,
			&notification.NotificationType, &sentAtStr, &notification.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning notification row: %v", err)
			continue
		}

		if sentAtStr.Valid {
			if sentAt, err := time.Parse("2006-01-02 15:04:05", sentAtStr.String); err == nil {
				notification.SentAt = &sentAt
			}
		}

		notifications = append(notifications, notification)
	}

	// Determine if there are more results
	hasMore := (offset + limit) < totalCount

	response := AdminRetentionNotificationsResponse{
		Notifications: notifications,
		TotalCount:    totalCount,
		Limit:         limit,
		Offset:        offset,
		HasMore:       hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminRetentionNotificationPreferencesHandler handles GET /api/admin/email/retention-notification-preferences
// Returns retention notification preferences for a user
func (srv *Server) adminRetentionNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionNotificationPreferencesHandler started - Admin Retention Notification Preferences (Micro-Iteration 4.25)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention notification preferences requested by user: %s", userID)

	// Get target user ID from query parameter
	targetUserID := r.URL.Query().Get("user_id")
	if targetUserID == "" {
		targetUserID = userID // Default to requesting user
	}

	// Query notification preferences
	query := `
		SELECT user_id, enable_expiration_notifications, enable_cleanup_notifications,
		       expiration_advance_notice_hours, notification_email_template, created_at, updated_at
		FROM retention_notification_preferences
		WHERE user_id = ?
	`

	var prefs struct {
		UserID                        string    `json:"user_id"`
		EnableExpirationNotifications bool      `json:"enable_expiration_notifications"`
		EnableCleanupNotifications    bool      `json:"enable_cleanup_notifications"`
		ExpirationAdvanceNoticeHours  int       `json:"expiration_advance_notice_hours"`
		NotificationEmailTemplate     string    `json:"notification_email_template"`
		CreatedAt                     time.Time `json:"created_at"`
		UpdatedAt                     time.Time `json:"updated_at"`
	}

	err := srv.db.QueryRowContext(r.Context(), query, targetUserID).Scan(
		&prefs.UserID, &prefs.EnableExpirationNotifications, &prefs.EnableCleanupNotifications,
		&prefs.ExpirationAdvanceNoticeHours, &prefs.NotificationEmailTemplate,
		&prefs.CreatedAt, &prefs.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return default preferences if none exist
			prefs = struct {
				UserID                        string    `json:"user_id"`
				EnableExpirationNotifications bool      `json:"enable_expiration_notifications"`
				EnableCleanupNotifications    bool      `json:"enable_cleanup_notifications"`
				ExpirationAdvanceNoticeHours  int       `json:"expiration_advance_notice_hours"`
				NotificationEmailTemplate     string    `json:"notification_email_template"`
				CreatedAt                     time.Time `json:"created_at"`
				UpdatedAt                     time.Time `json:"updated_at"`
			}{
				UserID:                        targetUserID,
				EnableExpirationNotifications: true,
				EnableCleanupNotifications:    false,
				ExpirationAdvanceNoticeHours:  24,
				NotificationEmailTemplate:     "default",
				CreatedAt:                     time.Now(),
				UpdatedAt:                     time.Now(),
			}
		} else {
			log.Printf("Failed to get notification preferences: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to retrieve preferences"}`))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(prefs)
}

// adminRetentionNotificationPreferencesUpdateHandler handles PUT /api/admin/email/retention-notification-preferences
// Updates retention notification preferences for a user
func (srv *Server) adminRetentionNotificationPreferencesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionNotificationPreferencesUpdateHandler started - Admin Update Retention Notification Preferences (Micro-Iteration 4.25)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin update retention notification preferences requested by user: %s", userID)

	// Get target user ID from query parameter
	targetUserID := r.URL.Query().Get("user_id")
	if targetUserID == "" {
		targetUserID = userID // Default to requesting user
	}

	// Parse request body
	var prefs struct {
		EnableExpirationNotifications *bool   `json:"enable_expiration_notifications"`
		EnableCleanupNotifications    *bool   `json:"enable_cleanup_notifications"`
		ExpirationAdvanceNoticeHours  *int    `json:"expiration_advance_notice_hours"`
		NotificationEmailTemplate     *string `json:"notification_email_template"`
	}

	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request body"}`))
		return
	}

	// Build update query
	query := `
		INSERT OR REPLACE INTO retention_notification_preferences (
			user_id, enable_expiration_notifications, enable_cleanup_notifications,
			expiration_advance_notice_hours, notification_email_template, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	// Get current preferences to merge with updates
	currentPrefs := struct {
		EnableExpirationNotifications bool
		EnableCleanupNotifications    bool
		ExpirationAdvanceNoticeHours  int
		NotificationEmailTemplate     string
	}{
		EnableExpirationNotifications: true,
		EnableCleanupNotifications:    false,
		ExpirationAdvanceNoticeHours:  24,
		NotificationEmailTemplate:     "default",
	}

	// Query current preferences
	currentQuery := `
		SELECT enable_expiration_notifications, enable_cleanup_notifications,
		       expiration_advance_notice_hours, notification_email_template
		FROM retention_notification_preferences
		WHERE user_id = ?
	`

	err := srv.db.QueryRowContext(r.Context(), currentQuery, targetUserID).Scan(
		&currentPrefs.EnableExpirationNotifications, &currentPrefs.EnableCleanupNotifications,
		&currentPrefs.ExpirationAdvanceNoticeHours, &currentPrefs.NotificationEmailTemplate,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Failed to get current preferences: %v", err)
	}

	// Apply updates
	if prefs.EnableExpirationNotifications != nil {
		currentPrefs.EnableExpirationNotifications = *prefs.EnableExpirationNotifications
	}
	if prefs.EnableCleanupNotifications != nil {
		currentPrefs.EnableCleanupNotifications = *prefs.EnableCleanupNotifications
	}
	if prefs.ExpirationAdvanceNoticeHours != nil {
		currentPrefs.ExpirationAdvanceNoticeHours = *prefs.ExpirationAdvanceNoticeHours
	}
	if prefs.NotificationEmailTemplate != nil {
		currentPrefs.NotificationEmailTemplate = *prefs.NotificationEmailTemplate
	}

	// Update preferences
	now := time.Now()
	_, err = srv.db.ExecContext(r.Context(), query,
		targetUserID, currentPrefs.EnableExpirationNotifications, currentPrefs.EnableCleanupNotifications,
		currentPrefs.ExpirationAdvanceNoticeHours, currentPrefs.NotificationEmailTemplate,
		now, now,
	)

	if err != nil {
		log.Printf("Failed to update notification preferences: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to update preferences"}`))
		return
	}

	response := map[string]interface{}{
		"message":     "Notification preferences updated successfully",
		"user_id":     targetUserID,
		"preferences": currentPrefs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
