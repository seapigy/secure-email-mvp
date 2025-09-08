package auth

// DO NOT EDIT EXISTING CODE - new file added
// Signup handler (Go). Assumes a global DB variable is set by main application.
//
// Required env:
// - DATABASE_URL used by main to open DB and set auth.DB

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DB variable is declared in session.go

type signupRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	AccountType string `json:"account_type,omitempty"`
}

type safeUserResp struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	AccountType string `json:"account_type"`
	CreatedAt   string `json:"created_at"`
}

// POST /api/auth/signup
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Basic JSON decode
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	// Validate username characters (alphanumeric + . _ -)
	for _, ch := range req.Username {
		if !(ch == '.' || ch == '_' || ch == '-' || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
			http.Error(w, "invalid username", http.StatusBadRequest)
			return
		}
	}

	// Hash password
	hashed, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("ERROR hashing password: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Default account type and validate
	if req.AccountType == "" {
		req.AccountType = "free"
	}
	
	// Validate account type
	if req.AccountType != "free" && req.AccountType != "premium" && req.AccountType != "enterprise" {
		http.Error(w, "invalid account type", http.StatusBadRequest)
		return
	}

	// For Premium/Enterprise signup, create placeholder subscription
	// This allows users to sign up without requiring live Stripe keys
	var subscriptionID string
	if req.AccountType == "premium" || req.AccountType == "enterprise" {
		subscriptionID = uuid.New().String()
	}

	// Insert user
	id := uuid.New().String()
	now := time.Now().UTC()

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("ERROR begin tx: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}
	defer tx.Rollback()

	// Check duplicate email or username+domain uniqueness is handled in SQL indexes; here we proactively check email
	var exists int
	err = tx.QueryRow("SELECT 1 FROM users WHERE email = ? LIMIT 1", req.Email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		// For Oracle/Postgres use $1 param; your DB layer must adapt to driver
		// TODO: use a proper query placeholder depending on driver
	}

	// Generate verification code
	verificationCode, err := GenerateRandomToken(6) // 6-character code
	if err != nil {
		log.Printf("ERROR generating verification code: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	hashedVerificationCode := HashToken(verificationCode)
	verificationExpiresAt := now.Add(30 * time.Minute) // 30-minute expiry

	// Try insert (use DB-specific placeholder as necessary)
	_, err = tx.Exec(
		`INSERT INTO users (id, username, email, hashed_password, account_type, account_type_new, account_status, verification_code, verification_code_expires_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 'pending_verification', ?, ?, ?, ?)`,
		id, req.Username, req.Email, hashed, req.AccountType, req.AccountType, hashedVerificationCode, verificationExpiresAt, now, now,
	)
	if err != nil {
		// return nice error for duplicate
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			http.Error(w, "username or email already exists", http.StatusConflict)
			return
		}
		log.Printf("ERROR inserting user: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Create placeholder subscription for Premium/Enterprise users
	if subscriptionID != "" {
		_, err = tx.Exec(`
			INSERT INTO subscriptions (id, user_id, status, plan, start_date, end_date, created_at, updated_at)
			VALUES (?, ?, 'active', ?, ?, ?, ?, ?)
		`, subscriptionID, id, req.AccountType, now, now.AddDate(0, 1, 0), now, now)
		if err != nil {
			log.Printf("ERROR creating placeholder subscription: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("ERROR commit: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Create default inbox and welcome message
	if err := CreateDefaultInbox(id); err != nil {
		log.Printf("ERROR creating default inbox: %v", err)
		// Don't fail signup if inbox creation fails, just log it
	}

	// Minimal safe response
	resp := safeUserResp{
		ID:          id,
		Username:    req.Username,
		Email:       req.Email,
		AccountType: req.AccountType,
		CreatedAt:   now.Format(time.RFC3339),
	}

	// TODO: Send email with verification code
	// For now, log the code (in production, this would be sent via email service)
	log.Printf("INFO verification_code_generated user_id=%s code=%s", id, verificationCode)

	// Log event (non-sensitive)
	log.Printf("INFO user_created id=%s account_type=%s", id, req.AccountType)
	
	// Log analytics event
	LogAnalyticsEvent(id, EventUserSignup, map[string]interface{}{
		"account_type": req.AccountType,
		"success":      true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
