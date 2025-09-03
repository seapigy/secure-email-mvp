package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"secure-email-mvp/pkg/auth"
)

// SignupRequestV2 represents the JSON request body for the new signup endpoint
type SignupRequestV2 struct {
	Plan          string `json:"plan"`           // "free", "paid", or "company"
	Email         string `json:"email"`          // User's email address
	Password      string `json:"password"`       // User's password
	CompanyCode   string `json:"company_code"`   // Optional company code for company plans
	FallbackEmail string `json:"fallback_email"` // Required fallback email address
}

// SignupResponseV2 represents the JSON response for the new signup endpoint
type SignupResponseV2 struct {
	Status   string `json:"status"`
	UserID   string `json:"user_id"`
	NextStep string `json:"next_step"`
}

// signupHandlerV2Factory returns an HTTP handler for the new signup endpoint
// This handler follows strict privacy rules and supports Free, Paid, and Company plans
func signupHandlerV2Factory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse JSON request body
		var req SignupRequestV2
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields (without logging PII)
		if req.Plan == "" || req.Email == "" || req.Password == "" || req.FallbackEmail == "" {
			http.Error(w, `{"error":"Missing required fields"}`, http.StatusBadRequest)
			return
		}

		// Validate plan type
		if !isValidPlan(req.Plan) {
			http.Error(w, `{"error":"Invalid plan type"}`, http.StatusBadRequest)
			return
		}

		// Validate email format
		if !isValidEmailFormat(req.Email) {
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}

		// Validate fallback email format
		if !isValidEmailFormat(req.FallbackEmail) {
			http.Error(w, `{"error":"Invalid fallback email format"}`, http.StatusBadRequest)
			return
		}

		// Validate that fallback email is different from primary email
		if req.FallbackEmail == req.Email {
			http.Error(w, `{"error":"Fallback email must be different from primary email"}`, http.StatusBadRequest)
			return
		}

		// Validate password strength
		if !isValidPasswordStrength(req.Password) {
			http.Error(w, `{"error":"Password does not meet security requirements"}`, http.StatusBadRequest)
			return
		}

		// Validate company code for company plans
		if req.Plan == "company" && req.CompanyCode == "" {
			http.Error(w, `{"error":"Company code required for company plans"}`, http.StatusBadRequest)
			return
		}

		// Generate UUID for user_id
		userID := uuid.New().String()

		// Hash password with Argon2id (no raw storage, no logs)
		hashedPassword, err := hashPasswordWithArgon2id(req.Password)
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Generate TOTP secret for 2FA setup
		totpSecret, err := auth.GenerateTOTPSecret()
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Generate fallback confirmation token
		fallbackToken := auth.GenerateFallbackToken(req.FallbackEmail)
		fallbackExpiration := time.Now().Add(24 * time.Hour) // 24 hour expiration

		// Store user in database with explicit transaction
		log.Printf("[TX_DEBUG] 🔥 ===== V2 TRANSACTION START ===== email: %s", req.Email)

		// Begin explicit transaction
		log.Printf("[TX_DEBUG] 🔍 BEGIN TRANSACTION for V2 email: %s", req.Email)
		tx, err := db.Begin()
		if err != nil {
			log.Printf("[TX_DEBUG] ❌ FAILED TO BEGIN TRANSACTION for V2 email %s: %v", req.Email, err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("[TX_DEBUG] ✅ TRANSACTION BEGUN successfully for V2 email: %s", req.Email)

		// Ensure transaction is rolled back on error
		defer func() {
			if tx != nil {
				log.Printf("[TX_DEBUG] 🔄 ROLLBACK TRIGGERED for V2 email %s (tx not committed)", req.Email)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("[TX_DEBUG] ⚠️ FAILED TO ROLLBACK V2 transaction: %v", rollbackErr)
				} else {
					log.Printf("[TX_DEBUG] ✅ ROLLBACK COMPLETED for V2 email: %s", req.Email)
				}
			} else {
				log.Printf("[TX_DEBUG] ✅ NO ROLLBACK NEEDED - V2 transaction was committed for email: %s", req.Email)
			}
		}()

		// Create user within transaction
		log.Printf("[TX_DEBUG] 🔍 EXECUTING V2 USER INSERT for email: %s", req.Email)
		err = createUserV2InTx(tx, userID, req.Email, hashedPassword, req.Plan, req.CompanyCode, req.FallbackEmail, totpSecret, fallbackToken, fallbackExpiration)
		if err != nil {
			log.Printf("[TX_DEBUG] ❌ V2 DATABASE INSERTION FAILED for email %s: %v", req.Email, err)
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				http.Error(w, `{"error":"User already exists"}`, http.StatusBadRequest)
				return
			}
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("[TX_DEBUG] ✅ V2 USER INSERT SUCCESSFUL for email: %s", req.Email)

		// Commit the transaction
		log.Printf("[TX_DEBUG] 🔍 COMMITTING V2 TRANSACTION for email: %s", req.Email)
		if err = tx.Commit(); err != nil {
			log.Printf("[TX_DEBUG] ❌ FAILED TO COMMIT V2 TRANSACTION for email %s: %v", req.Email, err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Clear the defer rollback since we committed successfully
		log.Printf("[TX_DEBUG] ✅ V2 COMMIT SUCCESSFUL for email: %s", req.Email)
		tx = nil

		log.Printf("[TX_DEBUG] 🔥 ===== V2 TRANSACTION COMMITTED ===== email: %s", req.Email)

		// Enhanced V2 signup logging
		log.Printf("[SIGNUP_V2] ✅ User created successfully: email=%s, plan=%s, company_code=%s, fallback_email=%s",
			req.Email, req.Plan, req.CompanyCode, req.FallbackEmail)

		// POST-INSERT VERIFICATION: Confirm user actually exists in database
		log.Printf("[TX_DEBUG] 🔍 POST-INSERT VERIFICATION for V2 email: %s", req.Email)
		var userIDVerify string
		var userEmailVerify string
		err = db.QueryRow("SELECT id, email FROM users WHERE email = ?", req.Email).Scan(&userIDVerify, &userEmailVerify)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("[TX_DEBUG] ❌ CRITICAL: V2 User not found in database after commit for email: %s", req.Email)
				log.Printf("[TX_DEBUG] ❌ This indicates a serious database issue - V2 transaction committed but user not persisted")
			} else {
				log.Printf("[TX_DEBUG] ⚠️ Failed to verify V2 user in database for email %s: %v", req.Email, err)
			}
		} else {
			log.Printf("[TX_DEBUG] ✅ POST-INSERT VERIFICATION SUCCESSFUL for V2 email: %s", req.Email)
			log.Printf("[TX_DEBUG] ✅ Verified V2 user ID: %s, Email: %s", userIDVerify, userEmailVerify)
		}

		// Return success response
		response := SignupResponseV2{
			Status:   "success",
			UserID:   userID,
			NextStep: "verify_email",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// hashPasswordWithArgon2id hashes a password using Argon2id with secure parameters
// This function ensures no raw passwords are stored or logged
func hashPasswordWithArgon2id(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password with Argon2id
	// Parameters: time=3, memory=64MB, threads=4, keyLen=32
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	// Return encoded hash (salt + hash)
	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=4$%x$%x", salt, hash)
	return encodedHash, nil
}

// createUserV2 inserts a new user into the database with the new V2 schema
// Uses parameterized queries to prevent SQL injection
func createUserV2(db *sql.DB, userID, email, hashedPassword, plan, companyCode, fallbackEmail, totpSecret, fallbackToken string, fallbackExpiration time.Time) error {
	query := `
		INSERT INTO users (
			id, 
			email, 
			password, 
			password_hash, 
			plan, 
			company_code, 
			account_status, 
			fallback_email, 
			fallback_confirmed, 
			totp_secret, 
			totp_configured, 
			fallback_token, 
			fallback_token_expiration, 
			created_at, 
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := db.Exec(query, userID, email, hashedPassword, hashedPassword, plan, companyCode, "pending", fallbackEmail, 0, totpSecret, 0, fallbackToken, fallbackExpiration, now, now)
	return err
}

// createUserV2InTx inserts a new user into the database within a transaction
// Uses parameterized queries to prevent SQL injection
func createUserV2InTx(tx *sql.Tx, userID, email, hashedPassword, plan, companyCode, fallbackEmail, totpSecret, fallbackToken string, fallbackExpiration time.Time) error {
	log.Printf("[TX_DEBUG] 🔍 ===== V2 INSERT OPERATION START ===== email: %s", email)

	// Use the complete V2 schema with all required columns including TOTP and fallback
	query := `
		INSERT INTO users (
			id, 
			email, 
			password, 
			password_hash, 
			plan, 
			company_code, 
			account_status, 
			fallback_email, 
			fallback_confirmed, 
			totp_secret, 
			totp_configured, 
			fallback_token, 
			fallback_token_expiration, 
			created_at, 
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	log.Printf("[TX_DEBUG] 🔍 Executing V2 SQL query: %s", query)
		log.Printf("[TX_DEBUG] 🔍 V2 Query parameters:")
	log.Printf("[TX_DEBUG]   - userID: %s", userID)
	log.Printf("[TX_DEBUG]   - email: %s", email)
	log.Printf("[TX_DEBUG]   - password: [REDACTED]")
	log.Printf("[TX_DEBUG]   - hashedPassword: %s...", hashedPassword[:16])
	log.Printf("[TX_DEBUG]   - plan: %s", plan)
	log.Printf("[TX_DEBUG]   - companyCode: %s", companyCode)
	log.Printf("[TX_DEBUG]   - account_status: pending")
	log.Printf("[TX_DEBUG]   - fallback_email: %s", fallbackEmail)
	log.Printf("[TX_DEBUG]   - fallback_confirmed: 0")
	log.Printf("[TX_DEBUG]   - totp_secret: %s...", totpSecret[:16])
	log.Printf("[TX_DEBUG]   - totp_configured: 0")
	log.Printf("[TX_DEBUG]   - fallback_token: %s...", fallbackToken[:16])
	log.Printf("[TX_DEBUG]   - fallback_token_expiration: %v", fallbackExpiration)
	log.Printf("[TX_DEBUG]   - created_at: %v", time.Now())
	log.Printf("[TX_DEBUG]   - updated_at: %v", time.Now())
	
	now := time.Now()
	log.Printf("[TX_DEBUG] 🔍 Executing tx.Exec() for V2...")
	result, err := tx.Exec(query, userID, email, hashedPassword, hashedPassword, plan, companyCode, "pending", fallbackEmail, 0, totpSecret, 0, fallbackToken, fallbackExpiration, now, now)
	if err != nil {
		log.Printf("[TX_DEBUG] ❌ V2 DATABASE INSERT ERROR for email %s: %v", email, err)
		log.Printf("[TX_DEBUG] ❌ V2 Error type: %T", err)
		log.Printf("[TX_DEBUG] ❌ V2 Full error details: %+v", err)

		// Check for specific constraint violations
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Printf("[TX_DEBUG] ❌ V2 UNIQUE CONSTRAINT VIOLATION: email %s already exists", email)
		} else if strings.Contains(err.Error(), "NOT NULL constraint failed") {
			log.Printf("[TX_DEBUG] ❌ V2 NOT NULL CONSTRAINT VIOLATION: required field is null")
		} else if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			log.Printf("[TX_DEBUG] ❌ V2 FOREIGN KEY CONSTRAINT VIOLATION: referenced table/row not found")
		}

		return fmt.Errorf("V2 insert failed: %v", err)
	}

	// Log rows affected to confirm insertion
	log.Printf("[TX_DEBUG] 🔍 Getting V2 rows affected...")
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[TX_DEBUG] ⚠️ Could not get V2 rows affected: %v", err)
		log.Printf("[TX_DEBUG] ⚠️ V2 Rows affected error type: %T", err)
	} else {
		log.Printf("[TX_DEBUG] ✅ V2 Rows affected: %d", rowsAffected)
		if rowsAffected == 0 {
			log.Printf("[TX_DEBUG] ❌ CRITICAL: V2 Insert succeeded but 0 rows were affected")
			return fmt.Errorf("V2 insert succeeded but no rows were affected")
		}
		if rowsAffected > 1 {
			log.Printf("[TX_DEBUG] ⚠️ WARNING: V2 Insert affected %d rows (expected 1)", rowsAffected)
		}
	}

	log.Printf("[TX_DEBUG] ✅ ===== V2 INSERT OPERATION SUCCESSFUL ===== email: %s", email)
	return nil
}

// isValidPlan validates that the plan type is supported
func isValidPlan(plan string) bool {
	validPlans := map[string]bool{
		"free":    true,
		"paid":    true,
		"company": true,
	}
	return validPlans[plan]
}

// isValidEmailFormat validates email format without logging the email
func isValidEmailFormat(email string) bool {
	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return false
	}

	// Check for minimum length
	if len(email) < 5 {
		return false
	}

	// Check for maximum length
	if len(email) > 254 {
		return false
	}

	return true
}

// isValidPasswordStrength validates password strength without logging the password
func isValidPasswordStrength(password string) bool {
	// Minimum length check
	if len(password) < 8 {
		return false
	}

	// Maximum length check
	if len(password) > 128 {
		return false
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}
