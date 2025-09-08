package auth

// DO NOT EDIT EXISTING CODE - new file added
// Trial expiration warning system for Premium/Enterprise placeholder subscriptions

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// CheckTrialExpiration checks if user's trial is expiring soon
func CheckTrialExpiration(userID string) (*TrialWarning, error) {
	var subscriptionID string
	var plan string
	var endDate time.Time
	var status string

	err := DB.QueryRow(`
		SELECT s.id, s.plan, s.end_date, s.status
		FROM subscriptions s
		WHERE s.user_id = ? AND s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1
	`, userID).Scan(&subscriptionID, &plan, &endDate, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			// No active subscription found
			return nil, nil
		}
		return nil, err
	}

	// Check if trial expires within 7 days
	now := time.Now()
	daysUntilExpiry := int(endDate.Sub(now).Hours() / 24)

	if daysUntilExpiry <= 7 && daysUntilExpiry > 0 {
		return &TrialWarning{
			SubscriptionID: subscriptionID,
			Plan:           plan,
			DaysRemaining:  daysUntilExpiry,
			ExpiryDate:     endDate.Format(time.RFC3339),
			WarningLevel:   getWarningLevel(daysUntilExpiry),
		}, nil
	}

	return nil, nil
}

// TrialWarning represents a trial expiration warning
type TrialWarning struct {
	SubscriptionID string `json:"subscription_id"`
	Plan           string `json:"plan"`
	DaysRemaining  int    `json:"days_remaining"`
	ExpiryDate     string `json:"expiry_date"`
	WarningLevel   string `json:"warning_level"` // info, warning, critical
}

// getWarningLevel determines the warning level based on days remaining
func getWarningLevel(daysRemaining int) string {
	switch {
	case daysRemaining <= 1:
		return "critical"
	case daysRemaining <= 3:
		return "warning"
	default:
		return "info"
	}
}

// TrialWarningHandler returns trial expiration warning for authenticated user
func TrialWarningHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	warning, err := CheckTrialExpiration(userID)
	if err != nil {
		log.Printf("ERROR checking trial expiration: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"has_warning": warning != nil,
	}

	if warning != nil {
		response["warning"] = warning
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExtendTrialHandler extends trial for testing purposes (placeholder implementation)
func ExtendTrialHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// For placeholder implementation, extend trial by 30 days
	_, err := DB.Exec(`
		UPDATE subscriptions 
		SET end_date = datetime(end_date, '+30 days'), updated_at = ?
		WHERE user_id = ? AND status = 'active'
	`, time.Now(), userID)

	if err != nil {
		log.Printf("ERROR extending trial: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Trial extended by 30 days",
	})
}
