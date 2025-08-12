package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/suspicious"

	"github.com/gorilla/mux"
)

// SuspiciousActivityResponse represents the response for suspicious activity queries
type SuspiciousActivityResponse struct {
	EmailID              string                             `json:"email_id"`
	SuspiciousFlag       bool                               `json:"suspicious_flag"`
	SuspiciousFlagSetAt  *time.Time                         `json:"suspicious_flag_set_at,omitempty"`
	DetectionEvents      []suspicious.SuspiciousAccessEvent `json:"detection_events,omitempty"`
	TotalDetections      int                                `json:"total_detections"`
	UnresolvedDetections int                                `json:"unresolved_detections"`
}

// ClearSuspiciousFlagRequest represents the request to clear a suspicious flag
type ClearSuspiciousFlagRequest struct {
	ResolutionNotes string `json:"resolution_notes,omitempty"`
}

// ResolveDetectionRequest represents the request to resolve a detection event
type ResolveDetectionRequest struct {
	ResolutionNotes string `json:"resolution_notes"`
}

// UpdateUserPreferencesRequest represents the request to update user preferences
type UpdateUserPreferencesRequest struct {
	EnableSuspiciousDetection      bool                `json:"enable_suspicious_detection"`
	NotifyOnSuspiciousActivity     bool                `json:"notify_on_suspicious_activity"`
	AutoFlagSuspiciousEmails       bool                `json:"auto_flag_suspicious_emails"`
	MinimumSeverityForNotification suspicious.Severity `json:"minimum_severity_for_notification"`
}

// getSuspiciousActivityHandler handles GET /api/suspicious/activity/{email_id}
// Returns suspicious activity information for a specific email
func (srv *Server) getSuspiciousActivityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["email_id"]

	// Get user ID from context (set by JWT middleware)
	userID := r.Context().Value("user_id").(string)

	// Verify email ownership
	if err := srv.verifyEmailOwnership(r.Context(), emailID, userID); err != nil {
		log.Printf("Unauthorized access attempt to suspicious activity: user %s, email %s", userID, emailID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// Get suspicious activity information
	response, err := srv.getSuspiciousActivityInfo(r.Context(), emailID)
	if err != nil {
		log.Printf("Failed to get suspicious activity for email %s: %v", emailID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve suspicious activity"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getSuspiciousActivityInfo gets suspicious activity information for an email
func (srv *Server) getSuspiciousActivityInfo(ctx context.Context, emailID string) (*SuspiciousActivityResponse, error) {
	// Get suspicious flag status
	var suspiciousFlag bool
	var suspiciousFlagSetAt *time.Time

	query := `
		SELECT suspicious_flag, suspicious_flag_set_at 
		FROM emails 
		WHERE email_id = ?
	`
	err := srv.db.QueryRowContext(ctx, query, emailID).Scan(&suspiciousFlag, &suspiciousFlagSetAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get suspicious flag status: %w", err)
	}

	// Get detection events
	detectionEvents, err := srv.suspiciousService.GetSuspiciousAccessEvents(ctx, emailID, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get detection events: %w", err)
	}

	// Count unresolved detections
	unresolvedCount := 0
	for _, event := range detectionEvents {
		if event.ResolvedAt == nil {
			unresolvedCount++
		}
	}

	response := &SuspiciousActivityResponse{
		EmailID:              emailID,
		SuspiciousFlag:       suspiciousFlag,
		SuspiciousFlagSetAt:  suspiciousFlagSetAt,
		DetectionEvents:      detectionEvents,
		TotalDetections:      len(detectionEvents),
		UnresolvedDetections: unresolvedCount,
	}

	return response, nil
}

// clearSuspiciousFlagHandler handles POST /api/suspicious/clear-flag/{email_id}
// Clears the suspicious flag on an email
func (srv *Server) clearSuspiciousFlagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["email_id"]

	// Get user ID from context (set by JWT middleware)
	userID := r.Context().Value("user_id").(string)

	// Verify email ownership
	if err := srv.verifyEmailOwnership(r.Context(), emailID, userID); err != nil {
		log.Printf("Unauthorized attempt to clear suspicious flag: user %s, email %s", userID, emailID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// Parse request body
	var req ClearSuspiciousFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request format"}`))
		return
	}

	// Clear the suspicious flag
	if err := srv.suspiciousService.ClearSuspiciousFlag(r.Context(), emailID, userID); err != nil {
		log.Printf("Failed to clear suspicious flag for email %s: %v", emailID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to clear suspicious flag"}`))
		return
	}

	// If resolution notes provided, resolve all unresolved detection events
	if req.ResolutionNotes != "" {
		detectionEvents, err := srv.suspiciousService.GetSuspiciousAccessEvents(r.Context(), emailID, 100)
		if err != nil {
			log.Printf("Failed to get detection events for resolution: %v", err)
		} else {
			for _, event := range detectionEvents {
				if event.ResolvedAt == nil {
					if err := srv.suspiciousService.ResolveDetectionEvent(r.Context(), event.DetectionID, userID, req.ResolutionNotes); err != nil {
						log.Printf("Failed to resolve detection event %s: %v", event.DetectionID, err)
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Suspicious flag cleared successfully"}`))
}

// resolveDetectionHandler handles POST /api/suspicious/resolve/{detection_id}
// Resolves a specific detection event
func (srv *Server) resolveDetectionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	detectionID := vars["detection_id"]

	// Get user ID from context (set by JWT middleware)
	userID := r.Context().Value("user_id").(string)

	// Parse request body
	var req ResolveDetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request format"}`))
		return
	}

	// Validate resolution notes
	if req.ResolutionNotes == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Resolution notes are required"}`))
		return
	}

	// Get the detection event to verify ownership
	var emailID string
	query := `SELECT email_id FROM suspicious_access_events WHERE detection_id = ?`
	err := srv.db.QueryRowContext(r.Context(), query, detectionID).Scan(&emailID)
	if err != nil {
		log.Printf("Detection event not found: %s", detectionID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Detection event not found"}`))
		return
	}

	// Verify email ownership
	if err := srv.verifyEmailOwnership(r.Context(), emailID, userID); err != nil {
		log.Printf("Unauthorized attempt to resolve detection: user %s, detection %s", userID, detectionID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// Resolve the detection event
	if err := srv.suspiciousService.ResolveDetectionEvent(r.Context(), detectionID, userID, req.ResolutionNotes); err != nil {
		log.Printf("Failed to resolve detection event %s: %v", detectionID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to resolve detection event"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Detection event resolved successfully"}`))
}

// getUserPreferencesHandler handles GET /api/suspicious/preferences
// Returns user preferences for suspicious activity detection
func (srv *Server) getUserPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID := r.Context().Value("user_id").(string)

	// Get user preferences
	prefs, err := srv.suspiciousService.GetUserPreferences(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get user preferences for user %s: %v", userID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve user preferences"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// updateUserPreferencesHandler handles PUT /api/suspicious/preferences
// Updates user preferences for suspicious activity detection
func (srv *Server) updateUserPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID := r.Context().Value("user_id").(string)

	// Parse request body
	var req UpdateUserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request format"}`))
		return
	}

	// Update user preferences
	query := `
		INSERT OR REPLACE INTO user_suspicious_activity_preferences (
			user_id, enable_suspicious_detection, notify_on_suspicious_activity,
			auto_flag_suspicious_emails, minimum_severity_for_notification, updated_at
		) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := srv.db.ExecContext(r.Context(), query, userID, req.EnableSuspiciousDetection,
		req.NotifyOnSuspiciousActivity, req.AutoFlagSuspiciousEmails, req.MinimumSeverityForNotification)
	if err != nil {
		log.Printf("Failed to update user preferences for user %s: %v", userID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to update user preferences"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"User preferences updated successfully"}`))
}

// getDetectionRulesHandler handles GET /api/suspicious/rules
// Returns all detection rules (admin only)
func (srv *Server) getDetectionRulesHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin (you may need to implement admin role checking)
	// For now, we'll allow all authenticated users to view rules
	// In production, you should implement proper role-based access control

	// Get detection rules
	rules, err := srv.suspiciousService.GetEnabledDetectionRules(r.Context())
	if err != nil {
		log.Printf("Failed to get detection rules: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve detection rules"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// getUserSuspiciousEmailsHandler handles GET /api/suspicious/emails
// Returns list of user's emails with suspicious flags
func (srv *Server) getUserSuspiciousEmailsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID := r.Context().Value("user_id").(string)

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get suspicious emails for the user
	query := `
		SELECT email_id, subject, suspicious_flag, suspicious_flag_set_at, created_at
		FROM emails 
		WHERE sender_id = ? AND suspicious_flag = TRUE
		ORDER BY suspicious_flag_set_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := srv.db.QueryContext(r.Context(), query, userID, pageSize, offset)
	if err != nil {
		log.Printf("Failed to get suspicious emails for user %s: %v", userID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve suspicious emails"}`))
		return
	}
	defer rows.Close()

	var suspiciousEmails []map[string]interface{}
	for rows.Next() {
		var emailID, subject string
		var suspiciousFlag bool
		var suspiciousFlagSetAt *time.Time
		var createdAt time.Time

		err := rows.Scan(&emailID, &subject, &suspiciousFlag, &suspiciousFlagSetAt, &createdAt)
		if err != nil {
			log.Printf("Failed to scan suspicious email: %v", err)
			continue
		}

		email := map[string]interface{}{
			"email_id":               emailID,
			"subject":                subject,
			"suspicious_flag":        suspiciousFlag,
			"suspicious_flag_set_at": suspiciousFlagSetAt,
			"created_at":             createdAt,
		}

		suspiciousEmails = append(suspiciousEmails, email)
	}

	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM emails WHERE sender_id = ? AND suspicious_flag = TRUE`
	err = srv.db.QueryRowContext(r.Context(), countQuery, userID).Scan(&totalCount)
	if err != nil {
		log.Printf("Failed to count suspicious emails: %v", err)
	}

	response := map[string]interface{}{
		"suspicious_emails": suspiciousEmails,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
