// =============================================================================
// SECURE EMAIL MVP - NOTIFICATION HANDLERS
// =============================================================================
// Handlers for notification preferences and access event history.
// =============================================================================

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/geolocation"
	"secure-email-mvp/pkg/notification"

	"github.com/google/uuid"
)

// GetNotificationPreferencesRequest represents a request to get notification preferences
type GetNotificationPreferencesRequest struct {
	UserID string `json:"user_id"`
}

// UpdateNotificationPreferencesRequest represents a request to update notification preferences
type UpdateNotificationPreferencesRequest struct {
	UserID             string `json:"user_id"`
	EmailNotifications bool   `json:"email_notifications"`
	SMSNotifications   bool   `json:"sms_notifications"`
	NotifyOnSuccess    bool   `json:"notify_on_success"`
	NotifyOnFailure    bool   `json:"notify_on_failure"`
	NotifyOnBlocked    bool   `json:"notify_on_blocked"`
	IncludeGeolocation bool   `json:"include_geolocation"`
	IncludeDeviceInfo  bool   `json:"include_device_info"`
}

// GetAccessEventHistoryRequest represents a request to get access event history
type GetAccessEventHistoryRequest struct {
	UserID string `json:"user_id"`
	Limit  int    `json:"limit,omitempty"`
}

// NotificationResponse represents a notification API response
type NotificationResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// getNotificationPreferencesHandler handles GET /api/notifications/preferences
func (srv *Server) getNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getNotificationPreferencesHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(NotificationResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	// Create notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get notification preferences
	prefs, err := notificationService.GetNotificationPreferences(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get notification preferences: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NotificationResponse{
			Success: false,
			Message: "Failed to retrieve notification preferences",
		})
		return
	}

	// Return preferences
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Success: true,
		Data:    prefs,
	})
}

// updateNotificationPreferencesHandler handles PUT /api/notifications/preferences
func (srv *Server) updateNotificationPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("updateNotificationPreferencesHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(NotificationResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	// Parse request body
	var req UpdateNotificationPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NotificationResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Ensure user can only update their own preferences
	req.UserID = userID

	// Create notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Convert request to preferences struct
	prefs := &notification.NotificationPreferences{
		UserID:             req.UserID,
		EmailNotifications: req.EmailNotifications,
		SMSNotifications:   req.SMSNotifications,
		NotifyOnSuccess:    req.NotifyOnSuccess,
		NotifyOnFailure:    req.NotifyOnFailure,
		NotifyOnBlocked:    req.NotifyOnBlocked,
		IncludeGeolocation: req.IncludeGeolocation,
		IncludeDeviceInfo:  req.IncludeDeviceInfo,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Update preferences
	if err := notificationService.UpdateNotificationPreferences(r.Context(), prefs); err != nil {
		log.Printf("Failed to update notification preferences: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NotificationResponse{
			Success: false,
			Message: "Failed to update notification preferences",
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Success: true,
		Message: "Notification preferences updated successfully",
		Data:    prefs,
	})
}

// getAccessEventHistoryHandler handles GET /api/notifications/history
func (srv *Server) getAccessEventHistoryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getAccessEventHistoryHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(NotificationResponse{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	// Get limit from query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Create notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get access event history
	events, err := notificationService.GetAccessEventHistory(r.Context(), userID, limit)
	if err != nil {
		log.Printf("Failed to get access event history: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NotificationResponse{
			Success: false,
			Message: "Failed to retrieve access event history",
		})
		return
	}

	// Return events
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Success: true,
		Data:    events,
	})
}

// recordAccessEvent is a helper function to record access events
func (srv *Server) recordAccessEvent(ctx context.Context, emailID, userID string, eventType notification.AccessEventType, r *http.Request, failureReason string) error {
	// Create notification service
	notificationService := notification.NewNotificationService(srv.db)

	// Get client IP and user agent
	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	// Get geolocation information
	var country, city string
	geoService := geolocation.NewGeolocationService()
	if location, err := geoService.GetLocationByIP(clientIP); err == nil {
		country = location.Country
		city = location.City
	}

	// Detect device type
	deviceType := notification.DetectDeviceType(userAgent)

	// Create access event
	event := &notification.AccessEvent{
		EventID:       uuid.New().String(),
		EmailID:       emailID,
		UserID:        userID,
		EventType:     eventType,
		IPAddress:     clientIP,
		UserAgent:     userAgent,
		Country:       country,
		City:          city,
		DeviceType:    deviceType,
		FailureReason: failureReason,
		Timestamp:     time.Now(),
	}

	// Record the event
	if err := notificationService.RecordAccessEvent(ctx, event); err != nil {
		log.Printf("Failed to record access event: %v", err)
		return err
	}

	// Get notification preferences
	prefs, err := notificationService.GetNotificationPreferences(ctx, userID)
	if err != nil {
		log.Printf("Failed to get notification preferences: %v", err)
		return err
	}

	// Send notification if enabled
	if err := notificationService.SendNotification(ctx, event, prefs); err != nil {
		log.Printf("Failed to send notification: %v", err)
		// Don't return error here as the event was already recorded
	}

	return nil
}
