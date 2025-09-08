package auth

// DO NOT EDIT EXISTING CODE - new file added
// Privacy-friendly analytics and event logging system

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AnalyticsEvent represents an anonymized analytics event
type AnalyticsEvent struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	UserHash  string                 `json:"user_hash"` // SHA256 hash of user ID for privacy
	Metadata  map[string]interface{} `json:"metadata"`  // JSON metadata (no personal data)
	Timestamp time.Time              `json:"timestamp"`
}

// LogAnalyticsEvent logs an anonymized analytics event
func LogAnalyticsEvent(userID, eventType string, metadata map[string]interface{}) {
	// Create anonymized user hash
	userHash := hashUserID(userID)
	
	// Ensure metadata contains no personal information
	safeMetadata := sanitizeMetadata(metadata)
	
	metadataJSON, err := json.Marshal(safeMetadata)
	if err != nil {
		log.Printf("ERROR marshaling analytics metadata: %v", err)
		return
	}

	// Insert analytics event
	_, err = DB.Exec(`
		INSERT INTO analytics_events (id, event_type, user_hash, metadata, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, generateEventID(), eventType, userHash, string(metadataJSON), time.Now())

	if err != nil {
		log.Printf("ERROR logging analytics event: %v", err)
	}
}

// hashUserID creates a privacy-preserving hash of the user ID
func hashUserID(userID string) string {
	hash := sha256.Sum256([]byte(userID + "analytics_salt"))
	return hex.EncodeToString(hash[:])
}

// sanitizeMetadata removes any potentially personal information from metadata
func sanitizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	safe := make(map[string]interface{})
	
	// Allowed metadata fields (no personal data)
	allowedFields := map[string]bool{
		"account_type":     true,
		"feature_used":     true,
		"action_type":      true,
		"success":          true,
		"error_code":       true,
		"session_duration": true,
		"device_type":      true, // generic: mobile, desktop, tablet
		"country_code":     true, // ISO country code only
		"plan_type":        true,
		"trial_days_left":  true,
	}

	for key, value := range metadata {
		if allowedFields[key] {
			safe[key] = value
		}
	}

	return safe
}

// generateEventID creates a unique event ID
func generateEventID() string {
	return uuid.New().String()
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// AnalyticsHandler returns anonymized analytics data (admin only)
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is admin (placeholder - implement proper admin check)
	var accountType string
	err := DB.QueryRow("SELECT account_type_new FROM users WHERE id = ?", userID).Scan(&accountType)
	if err != nil || accountType != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Get analytics events (last 30 days)
	rows, err := DB.Query(`
		SELECT event_type, user_hash, metadata, timestamp
		FROM analytics_events
		WHERE timestamp >= datetime('now', '-30 days')
		ORDER BY timestamp DESC
		LIMIT 1000
	`)
	if err != nil {
		log.Printf("ERROR getting analytics: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []AnalyticsEvent
	for rows.Next() {
		var event AnalyticsEvent
		var metadataJSON string

		err := rows.Scan(&event.EventType, &event.UserHash, &metadataJSON, &event.Timestamp)
		if err != nil {
			continue
		}

		// Parse metadata
		var metadata map[string]interface{}
		json.Unmarshal([]byte(metadataJSON), &metadata)
		event.Metadata = metadata
		events = append(events, event)
	}

	// Calculate summary statistics
	summary := calculateAnalyticsSummary(events)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events":  events,
		"summary": summary,
	})
}

// calculateAnalyticsSummary calculates anonymized summary statistics
func calculateAnalyticsSummary(events []AnalyticsEvent) map[string]interface{} {
	eventCounts := make(map[string]int)
	accountTypes := make(map[string]int)
	
	for _, event := range events {
		eventCounts[event.EventType]++
		
		// Extract account type from metadata if available
		if accountType, ok := event.Metadata["account_type"].(string); ok {
			accountTypes[accountType]++
		}
	}

	return map[string]interface{}{
		"total_events":    len(events),
		"event_types":     eventCounts,
		"account_types":   accountTypes,
		"date_range":      "last_30_days",
	}
}

// Common analytics event types
const (
	EventUserSignup        = "user_signup"
	EventUserLogin         = "user_login"
	EventEmailSent         = "email_sent"
	EventEmailReceived     = "email_received"
	EventFolderCreated     = "folder_created"
	EventTrialWarning      = "trial_warning"
	EventAccountUpgrade    = "account_upgrade"
	EventAccountDowngrade  = "account_downgrade"
	EventDomainAdded       = "domain_added"
	EventDomainVerified    = "domain_verified"
	EventOrgCreated        = "organization_created"
	EventOrgUserAdded      = "organization_user_added"
	EventMfaSetup          = "mfa_setup"
	EventMfaUsed           = "mfa_used"
)
