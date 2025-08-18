package auth

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var emailRegex = regexp.MustCompile(`^[^@]+@securesystem\.email$`)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidateEmail checks if email matches securesystem.email domain for tenant isolation and phishing prevention.
// For development/testing, also allows example.com domain.
func ValidateEmail(email string) bool {
	if emailRegex.MatchString(email) {
		return true
	}
	// Allow example.com for development/testing
	if strings.HasSuffix(email, "@example.com") {
		return true
	}
	return false
}

// ValidatePassword checks length (8–128 characters) for password policy enforcement.
func ValidatePassword(password string) bool {
	return len(password) >= 8 && len(password) <= 128
}

// ValidateTOTP checks 6-digit format for TOTP code validity.
func ValidateTOTP(code string) bool {
	return len(code) == 6 && regexp.MustCompile(`^\d{6}$`).MatchString(code)
}

// HashPassword creates Argon2 hash with email as salt for strong password security and user-specific salting.
func HashPassword(password, email string) (string, error) {
	// Use the new configuration-based hashing
	hash, err := hashPasswordWithConfig(password, email, GlobalAuthConfig.Argon2)
	if err != nil {
		return "", err
	}

	// Return as string for backward compatibility
	return string(hash), nil
}

// GenerateTOTPSecret creates a new base32 TOTP secret for 2FA setup.
func GenerateTOTPSecret() (string, error) {
	// Use the new configuration-based TOTP secret generation
	return generateTOTPSecretWithConfig(GlobalAuthConfig.TOTP)
}

// Authenticate verifies credentials (email, password, TOTP) and returns a JWT if successful. Enforces all security checks.
func Authenticate(db *sql.DB, email, password, totpCode string) (string, string, error) {
	// DIAGNOSTIC: Log authentication attempt
	log.Printf("[AUTH_DEBUG] Authentication attempt for email: %s", email)

	// Validate inputs
	if !ValidateEmail(email) {
		log.Printf("[AUTH_DEBUG] Email validation failed for: %s", email)
		return "", "", fmt.Errorf("invalid email format")
	}
	if !ValidatePassword(password) {
		log.Printf("[AUTH_DEBUG] Password validation failed for: %s", email)
		return "", "", fmt.Errorf("invalid password length")
	}
	if !ValidateTOTP(totpCode) {
		log.Printf("[AUTH_DEBUG] TOTP validation failed for: %s", email)
		return "", "", fmt.Errorf("invalid TOTP format")
	}

	// Query user - handle both schema versions
	var user struct {
		ID           string
		PasswordHash string
		TOTPSecret   string
	}

	// Try to query with password_hash column first (full schema)
	err := db.QueryRow("SELECT id, password_hash, totp_secret FROM users WHERE email = ?", email).Scan(&user.ID, &user.PasswordHash, &user.TOTPSecret)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("user not found")
		}
		// If password_hash column doesn't exist, try with password column (simple schema)
		err = db.QueryRow("SELECT id, password, totp_secret FROM users WHERE email = ?", email).Scan(&user.ID, &user.PasswordHash, &user.TOTPSecret)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("[AUTH_DEBUG] User not found: %s", email)
				return "", "", fmt.Errorf("user not found")
			}
			log.Printf("[AUTH_DEBUG] Database error for %s: %v", email, err)
			return "", "", fmt.Errorf("database error: %v", err)
		}
	}

	// DIAGNOSTIC: Log user found and TOTP details
	log.Printf("[AUTH_DEBUG] User found - ID: %s, Email: %s", user.ID, email)
	log.Printf("[AUTH_DEBUG] TOTP Secret (base32): %s", user.TOTPSecret)
	log.Printf("[AUTH_DEBUG] TOTP Code being validated: %s", totpCode)
	log.Printf("[AUTH_DEBUG] Current timestamp: %v", time.Now())

	// DIAGNOSTIC: Log Argon2 parameters and email normalization
	normalizedEmail := normalizeEmail(email)
	log.Printf("[AUTH_DEBUG] Email normalization - Original: '%s', Normalized: '%s'", email, normalizedEmail)
	log.Printf("[AUTH_DEBUG] Argon2 Parameters - Memory: %d, Iterations: %d, Parallelism: %d, KeyLength: %d",
		GlobalAuthConfig.Argon2.Memory, GlobalAuthConfig.Argon2.Iterations,
		GlobalAuthConfig.Argon2.Parallelism, GlobalAuthConfig.Argon2.KeyLength)

	// Verify password with Argon2 using new configuration
	expectedHash := []byte(user.PasswordHash)
	actualHash, err := hashPasswordWithConfig(password, email, GlobalAuthConfig.Argon2)
	if err != nil {
		log.Printf("[AUTH_DEBUG] Password hashing error: %v", err)
		return "", "", fmt.Errorf("password verification error")
	}

	// DIAGNOSTIC: Log hash comparison details
	logHashComparison(expectedHash, actualHash, email)

	passwordValid := compareHashes(expectedHash, actualHash)

	// Backward compatibility: try old hashing method if new method fails
	if !passwordValid && !GlobalAuthConfig.UseNewFlow {
		log.Printf("[AUTH_DEBUG] New hash method failed, trying old method for user: %s", email)

		// Try old Argon2 parameters (hardcoded values)
		oldHash := argon2.IDKey([]byte(password), []byte(normalizedEmail), 1, 64*1024, 4, 32)
		passwordValid = compareHashes(expectedHash, oldHash)

		if passwordValid {
			log.Printf("[AUTH_DEBUG] Old hash method succeeded, user should be migrated: %s", email)
			// TODO: Implement password rehashing on next login
		}
	}

	if !passwordValid {
		log.Printf("[AUTH_DEBUG] Password verification FAILED for user: %s", email)
		return "", "", fmt.Errorf("invalid password")
	}

	log.Printf("[AUTH_DEBUG] Password verification SUCCESS for user: %s", email)

	// DIAGNOSTIC: Log TOTP validation details
	logTOTPValidation(totpCode, user.TOTPSecret, GlobalAuthConfig.TOTP)

	// Verify TOTP using new configuration with time skew tolerance
	totpValid := validateTOTPWithConfig(totpCode, user.TOTPSecret, GlobalAuthConfig.TOTP)

	if !totpValid {
		log.Printf("[AUTH_DEBUG] TOTP verification FAILED for user: %s", email)
		return "", "", fmt.Errorf("invalid TOTP code")
	}

	log.Printf("[AUTH_DEBUG] TOTP verification SUCCESS for user: %s", email)

	// Generate JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", "", fmt.Errorf("JWT_SECRET not configured")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("JWT signing error: %v", err)
	}

	return tokenString, user.ID, nil
}

// CreateUser creates a new user with hashed password and TOTP secret, ensuring uniqueness and secure storage.
func CreateUser(db *sql.DB, email, password string) (string, string, error) {
	// Validate inputs
	if !ValidateEmail(email) {
		return "", "", fmt.Errorf("invalid email format")
	}
	if !ValidatePassword(password) {
		return "", "", fmt.Errorf("invalid password length")
	}

	// Check if user already exists
	var existingID string
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&existingID)
	if err == nil {
		return "", "", fmt.Errorf("user already exists")
	} else if err != sql.ErrNoRows {
		return "", "", fmt.Errorf("database error: %v", err)
	}

	// Generate user ID
	userID := uuid.New().String()

	// Hash password using Argon2 with email as salt
	passwordHash, err := HashPassword(password, email)
	if err != nil {
		return "", "", fmt.Errorf("password hashing error: %v", err)
	}

	// Generate TOTP secret
	totpSecret, err := GenerateTOTPSecret()
	if err != nil {
		return "", "", fmt.Errorf("TOTP secret generation error: %v", err)
	}

	// Insert user
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		userID, email, passwordHash, totpSecret)
	if err != nil {
		return "", "", fmt.Errorf("database insert error: %v", err)
	}

	return userID, totpSecret, nil
}

// ValidateJWT validates and parses JWT token for protected endpoints.
func ValidateJWT(tokenString string) (string, string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", "", fmt.Errorf("JWT_SECRET not configured")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", "", fmt.Errorf("JWT parsing error: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", "", fmt.Errorf("invalid user_id in JWT")
		}
		email, ok := claims["email"].(string)
		if !ok {
			return "", "", fmt.Errorf("invalid email in JWT")
		}
		return userID, email, nil
	}

	return "", "", fmt.Errorf("invalid JWT token")
}
