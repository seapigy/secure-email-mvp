package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// SignupRequestV2 represents the JSON request body for the new signup endpoint
type SignupRequestV2 struct {
	Plan        string `json:"plan"`         // "free", "paid", or "company"
	Email       string `json:"email"`        // User's email address
	Password    string `json:"password"`     // User's password
	CompanyCode string `json:"company_code"` // Optional company code for company plans
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
		if req.Plan == "" || req.Email == "" || req.Password == "" {
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

		// Store user in database
		err = createUserV2(db, userID, req.Email, hashedPassword, req.Plan, req.CompanyCode)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				http.Error(w, `{"error":"User already exists"}`, http.StatusBadRequest)
				return
			}
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
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

// createUserV2 inserts a new user into the database with the new schema
// Uses parameterized queries to prevent SQL injection
func createUserV2(db *sql.DB, userID, email, hashedPassword, plan, companyCode string) error {
	query := `
		INSERT INTO users (
			id, 
			email, 
			password_hash, 
			plan, 
			company_code, 
			status, 
			created_at, 
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := db.Exec(query, userID, email, hashedPassword, plan, companyCode, "pending_verification", now, now)
	return err
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
