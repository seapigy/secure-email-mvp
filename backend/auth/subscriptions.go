package auth

// DO NOT EDIT EXISTING CODE - new file added
// Subscription management handlers: upgrade, downgrade, and billing management

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
)

type upgradeAccountRequest struct {
	Plan string `json:"plan"` // premium or enterprise
}

type upgradeAccountResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ClientSecret string `json:"client_secret,omitempty"` // For Stripe payment intent
}

type downgradeAccountRequest struct {
	Plan string `json:"plan"` // free or premium
}

type downgradeAccountResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// encryptSubscriptionData encrypts sensitive subscription data using AES-256-GCM
func encryptSubscriptionData(data string) (string, error) {
	return encryptTOTPSecret(data) // Reuse the same encryption function
}

// decryptSubscriptionData decrypts sensitive subscription data using AES-256-GCM
func decryptSubscriptionData(encryptedData string) (string, error) {
	return decryptTOTPSecret(encryptedData) // Reuse the same decryption function
}

// POST /api/auth/upgrade-account
func UpgradeAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req upgradeAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate plan
	if req.Plan != "premium" && req.Plan != "enterprise" {
		http.Error(w, "invalid plan", http.StatusBadRequest)
		return
	}

	// Get current user and subscription
	var currentPlan string
	var stripeCustomerID sql.NullString
	err := DB.QueryRow(`
		SELECT u.account_type, s.stripe_customer_id 
		FROM users u 
		LEFT JOIN subscriptions s ON u.id = s.user_id AND s.status = 'active'
		WHERE u.id = ?
	`, userID).Scan(&currentPlan, &stripeCustomerID)

	if err != nil {
		log.Printf("ERROR getting user subscription: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Check if already on this plan or higher
	if (currentPlan == "premium" && req.Plan == "premium") ||
		(currentPlan == "enterprise" && (req.Plan == "premium" || req.Plan == "enterprise")) {
		http.Error(w, "already on this plan or higher", http.StatusConflict)
		return
	}

	// Initialize Stripe (in production, use real API key)
	stripeAPIKey := os.Getenv("STRIPE_SECRET_KEY")
	var customerID string
	var encryptedCustomerID string

	if stripeAPIKey == "" {
		// For testing/development, skip Stripe integration
		log.Printf("INFO: Skipping Stripe integration - no API key provided")
		customerID = "test_customer_" + userID // Use test customer ID
		encryptedCustomerID = customerID       // No encryption needed for test
	} else {
		stripe.Key = stripeAPIKey

		if stripeCustomerID.Valid {
			// Use existing customer
			decryptedCustomerID, err := decryptSubscriptionData(stripeCustomerID.String)
			if err != nil {
				log.Printf("ERROR decrypting customer ID: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			customerID = decryptedCustomerID
		} else {
			// Create new Stripe customer
			userEmail, err := getUserEmail(userID)
			if err != nil {
				log.Printf("ERROR getting user email: %v", err)
				http.Error(w, "database error", http.StatusServiceUnavailable)
				return
			}

			params := &stripe.CustomerParams{
				Email: stripe.String(userEmail),
			}
			customer, err := customer.New(params)
			if err != nil {
				log.Printf("ERROR creating Stripe customer: %v", err)
				http.Error(w, "payment service error", http.StatusServiceUnavailable)
				return
			}
			customerID = customer.ID
		}

		// Encrypt customer ID for storage
		encryptedCustomerID, err = encryptSubscriptionData(customerID)
		if err != nil {
			log.Printf("ERROR encrypting customer ID: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	// Create or update subscription in database
	subscriptionID := uuid.New().String()
	now := time.Now().UTC()

	// Determine subscription end date based on plan
	var endDate time.Time
	switch req.Plan {
	case "premium":
		endDate = now.AddDate(0, 1, 0) // 1 month
	case "enterprise":
		endDate = now.AddDate(1, 0, 0) // 1 year
	default:
		http.Error(w, "invalid plan", http.StatusBadRequest)
		return
	}

	_, err = DB.Exec(`
		INSERT INTO subscriptions (id, user_id, stripe_customer_id, status, plan, start_date, end_date, created_at, updated_at)
		VALUES (?, ?, ?, 'active', ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			stripe_customer_id = excluded.stripe_customer_id,
			status = 'active',
			plan = excluded.plan,
			start_date = excluded.start_date,
			end_date = excluded.end_date,
			updated_at = excluded.updated_at
	`, subscriptionID, userID, encryptedCustomerID, req.Plan, now, endDate, now, now)

	if err != nil {
		log.Printf("ERROR creating subscription: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Update user account type
	_, err = DB.Exec(`
		UPDATE users 
		SET account_type = ?, updated_at = ?
		WHERE id = ?
	`, req.Plan, now, userID)

	if err != nil {
		log.Printf("ERROR updating user account type: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log successful upgrade (non-sensitive)
	log.Printf("INFO account_upgraded user_id=%s plan=%s", userID, req.Plan)

	resp := upgradeAccountResponse{
		Success: true,
		Message: "Account upgraded successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/auth/downgrade-account
func DowngradeAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req downgradeAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate plan
	if req.Plan != "free" && req.Plan != "premium" {
		http.Error(w, "invalid plan", http.StatusBadRequest)
		return
	}

	// Get current subscription
	var currentPlan string
	var subscriptionID string
	err := DB.QueryRow(`
		SELECT u.account_type, s.id
		FROM users u 
		LEFT JOIN subscriptions s ON u.id = s.user_id AND s.status = 'active'
		WHERE u.id = ?
	`, userID).Scan(&currentPlan, &subscriptionID)

	if err != nil {
		log.Printf("ERROR getting user subscription: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Check if already on this plan or lower
	if (currentPlan == "free" && req.Plan == "free") ||
		(currentPlan == "premium" && req.Plan == "premium") {
		http.Error(w, "already on this plan or lower", http.StatusConflict)
		return
	}

	// Cancel current subscription
	if subscriptionID != "" {
		_, err = DB.Exec(`
			UPDATE subscriptions 
			SET status = 'canceled', updated_at = ?
			WHERE id = ?
		`, time.Now().UTC(), subscriptionID)

		if err != nil {
			log.Printf("ERROR canceling subscription: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}
	}

	// Update user account type
	_, err = DB.Exec(`
		UPDATE users 
		SET account_type = ?, updated_at = ?
		WHERE id = ?
	`, req.Plan, time.Now().UTC(), userID)

	if err != nil {
		log.Printf("ERROR updating user account type: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log successful downgrade (non-sensitive)
	log.Printf("INFO account_downgraded user_id=%s plan=%s", userID, req.Plan)

	resp := downgradeAccountResponse{
		Success: true,
		Message: "Account downgraded successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Helper function to get user email
func getUserEmail(userID string) (string, error) {
	var email string
	err := DB.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&email)
	return email, err
}
