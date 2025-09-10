package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"secure-email/internal/auth"
	"secure-email/internal/email"
	"secure-email/internal/tokens"

	_ "github.com/go-sql-driver/mysql"
)

// SignupRequest represents the signup request payload
type SignupRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Tier         string `json:"tier"`
	CustomDomain string `json:"custom_domain,omitempty"`
}

// SignupResponse represents the signup response payload
type SignupResponse struct {
	Status               string `json:"status"`
	Message              string `json:"message"`
	RecoveryToken        string `json:"recovery_token"`
	RecoveryTokenQRData  string `json:"recovery_token_qr_data_uri,omitempty"`
}

// SignupHandler handles user signup requests
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	// Validate inputs
	if err := validateSignupRequest(req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get database connection
	db, err := getDBConnection()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Generate verification token
	verificationToken, err := tokens.GenerateSecureToken()
	if err != nil {
		log.Printf("Verification token generation error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Hash verification token
	verificationTokenHash, err := tokens.HashToken(verificationToken)
	if err != nil {
		log.Printf("Verification token hashing error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Generate recovery token
	recoveryToken, err := tokens.GenerateSecureToken()
	if err != nil {
		log.Printf("Recovery token generation error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Hash recovery token
	recoveryTokenHash, err := tokens.HashToken(recoveryToken)
	if err != nil {
		log.Printf("Recovery token hashing error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Calculate expiration times
	verificationExp := time.Now().Add(time.Duration(getEnvInt("VERIFICATION_TOKEN_EXP_HOURS", 24)) * time.Hour)
	recoveryExp := time.Now().Add(time.Duration(getEnvInt("RECOVERY_TOKEN_EXP_DAYS", 7)) * 24 * time.Hour)

	// Insert user into database
	userID, err := insertUser(ctx, db, req, passwordHash, verificationTokenHash, verificationExp, recoveryTokenHash, recoveryExp)
	if err != nil {
		log.Printf("User insertion error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Create verification link
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	verificationLink := fmt.Sprintf("%s/verify-email?uid=%d&token=%s", frontendURL, userID, verificationToken)

	// Send verification email
	if err := email.SendVerificationEmail(req.Email, "Verify Your Email Address", verificationToken, verificationLink); err != nil {
		log.Printf("Email sending error: %v", err)
		// Don't fail the signup if email fails, but log it
	}

	// Generate QR code for recovery token (optional)
	qrDataURI := generateQRCodeDataURI(recoveryToken)

	// Create response
	response := SignupResponse{
		Status:              "ok",
		Message:             "Signup successful. A verification email has been sent. Save your recovery token securely.",
		RecoveryToken:       recoveryToken,
		RecoveryTokenQRData: qrDataURI,
	}

	// Send response
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Response encoding error: %v", err)
	}
}

// validateSignupRequest validates the signup request
func validateSignupRequest(req SignupRequest) error {
	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Validate tier
	validTiers := map[string]bool{
		"free":       true,
		"premium":    true,
		"enterprise": true,
	}
	if !validTiers[req.Tier] {
		return fmt.Errorf("invalid tier. Must be 'free', 'premium', or 'enterprise'")
	}

	// Validate free tier email domain
	if req.Tier == "free" && !strings.HasSuffix(req.Email, "@securesystem.email") {
		return fmt.Errorf("free tier accounts must use @securesystem.email domain")
	}

	// Validate enterprise tier (basic validation for now)
	if req.Tier == "enterprise" {
		// For now, just validate email format - enterprise policy will be enforced in later iterations
		if !emailRegex.MatchString(req.Email) {
			return fmt.Errorf("invalid email format for enterprise account")
		}
	}

	return nil
}

// insertUser inserts a new user into the database
func insertUser(ctx context.Context, db *sql.DB, req SignupRequest, passwordHash, verificationTokenHash string, verificationExp time.Time, recoveryTokenHash string, recoveryExp time.Time) (int64, error) {
	query := `
		INSERT INTO users (
			email, email_verified, verification_token_hash, verification_exp,
			password_hash, reset_token_hash, reset_token_exp, tier, custom_domain
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.ExecContext(ctx, query,
		req.Email, false, verificationTokenHash, verificationExp,
		passwordHash, recoveryTokenHash, recoveryExp, req.Tier, req.CustomDomain,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}

	return userID, nil
}

// getDBConnection returns a database connection
func getDBConnection() (*sql.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN environment variable not set")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// getEnvInt gets an environment variable as int with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// generateQRCodeDataURI generates a QR code data URI for the recovery token
func generateQRCodeDataURI(token string) string {
	// For now, return a placeholder. In production, you would use a QR code library
	// like github.com/skip2/go-qrcode to generate an actual QR code
	return fmt.Sprintf("data:image/png;base64,placeholder-qr-code-for-%s", token[:8])
}
