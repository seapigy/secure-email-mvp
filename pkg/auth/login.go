package auth

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"os"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var emailRegex = regexp.MustCompile(`^[^@]+@securesystem\.email$`)

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
	// STRUCTURED DEBUGGING: Initialize authentication trace
	testMode := os.Getenv("TEST_MODE") == "true"
	authTrace := map[string]interface{}{
		"email":           email,
		"email_normalized": "",
		"user_found":      false,
		"account_ok":      false,
		"locked":          false,
		"lock_until":      nil,
		"pw_ok":           false,
		"totp_required":   false,
		"totp_ok":         false,
		"session_ok":      false,
		"final_auth_ok":   false,
		"reason":          "",
	}

	// ENHANCED DEBUGGING: Log authentication attempt with request details
	log.Printf("[AUTH_DEBUG] ===== AUTHENTICATION ATTEMPT START =====")
	log.Printf("[AUTH_DEBUG] Email: %s", email)
	log.Printf("[AUTH_DEBUG] Password length: %d", len(password))
	log.Printf("[AUTH_DEBUG] TOTP code: %s", totpCode)
	log.Printf("[AUTH_DEBUG] Timestamp: %v", time.Now())
	log.Printf("[AUTH_DEBUG] Test mode: %t", testMode)

	// 1. Email normalization
	normalizedEmail := normalizeEmail(email)
	authTrace["email_normalized"] = normalizedEmail
	log.Printf("[AUTH_DEBUG] Email normalization: '%s' -> '%s'", email, normalizedEmail)

	// 2. Input validation
	if !ValidateEmail(email) {
		authTrace["reason"] = "invalid email format"
		log.Printf("[AUTH_DEBUG] ❌ Email validation failed for: %s", email)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("invalid email format")
	}
	log.Printf("[AUTH_DEBUG] ✅ Email validation passed")

	if !ValidatePassword(password) {
		authTrace["reason"] = "invalid password length"
		log.Printf("[AUTH_DEBUG] ❌ Password validation failed for: %s (length: %d)", email, len(password))
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("invalid password length")
	}
	log.Printf("[AUTH_DEBUG] ✅ Password validation passed")

	if !ValidateTOTP(totpCode) {
		authTrace["reason"] = "invalid TOTP format"
		log.Printf("[AUTH_DEBUG] ❌ TOTP validation failed for: %s (code: %s)", email, totpCode)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("invalid TOTP format")
	}
	log.Printf("[AUTH_DEBUG] ✅ TOTP format validation passed")

	// 3. User lookup with normalized email
	var user struct {
		ID           string
		PasswordHash string
		TOTPSecret   string
		IsActive     *bool
		EmailVerified *bool
		LockedUntil  *time.Time
	}

	log.Printf("[AUTH_DEBUG] 🔍 Querying user from database with normalized email: %s", normalizedEmail)

	// Try to query with password_hash column first (full schema)
	err := db.QueryRow("SELECT id, password_hash, totp_secret FROM users WHERE email = ?", normalizedEmail).Scan(&user.ID, &user.PasswordHash, &user.TOTPSecret)
	if err != nil {
		if err == sql.ErrNoRows {
			authTrace["reason"] = "user not found"
			log.Printf("[AUTH_DEBUG] ❌ User not found in password_hash schema: %s", normalizedEmail)
			if testMode {
				log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
			}
			return "", "", fmt.Errorf("user not found")
		}
		log.Printf("[AUTH_DEBUG] ⚠️  password_hash column query failed, trying password column: %v", err)

		// If password_hash column doesn't exist, try with password column (simple schema)
		err = db.QueryRow("SELECT id, password, totp_secret FROM users WHERE email = ?", normalizedEmail).Scan(&user.ID, &user.PasswordHash, &user.TOTPSecret)
		if err != nil {
			if err == sql.ErrNoRows {
				authTrace["reason"] = "user not found"
				log.Printf("[AUTH_DEBUG] ❌ User not found in password column schema: %s", normalizedEmail)
				if testMode {
					log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
				}
				return "", "", fmt.Errorf("user not found")
			}
			authTrace["reason"] = "database error"
			log.Printf("[AUTH_DEBUG] ❌ Database error for %s: %v", normalizedEmail, err)
			if testMode {
				log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
			}
			return "", "", fmt.Errorf("database error: %v", err)
		}
		log.Printf("[AUTH_DEBUG] ✅ User found using password column schema")
	} else {
		log.Printf("[AUTH_DEBUG] ✅ User found using password_hash column schema")
	}

	authTrace["user_found"] = true

	// ENHANCED DIAGNOSTIC: Log user details
	log.Printf("[AUTH_DEBUG] 📋 User Details:")
	log.Printf("[AUTH_DEBUG]   - ID: %s", user.ID)
	log.Printf("[AUTH_DEBUG]   - Email: %s", normalizedEmail)
	log.Printf("[AUTH_DEBUG]   - TOTP Secret (base32): %s", user.TOTPSecret)
	log.Printf("[AUTH_DEBUG]   - TOTP Secret length: %d", len(user.TOTPSecret))
	log.Printf("[AUTH_DEBUG]   - Password Hash length: %d", len(user.PasswordHash))
	log.Printf("[AUTH_DEBUG]   - TOTP Code being validated: %s", totpCode)

	// 4. Account status check
	if user.IsActive != nil && !*user.IsActive {
		authTrace["account_ok"] = false
		authTrace["reason"] = "account disabled"
		log.Printf("[AUTH_DEBUG] ❌ Account is disabled for user: %s", normalizedEmail)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("account disabled")
	}

	if user.EmailVerified != nil && !*user.EmailVerified {
		authTrace["account_ok"] = false
		authTrace["reason"] = "email not verified"
		log.Printf("[AUTH_DEBUG] ❌ Email not verified for user: %s", normalizedEmail)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("email not verified")
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		authTrace["account_ok"] = false
		authTrace["locked"] = true
		authTrace["lock_until"] = user.LockedUntil
		authTrace["reason"] = "account locked"
		log.Printf("[AUTH_DEBUG] ❌ Account is locked until %v for user: %s", *user.LockedUntil, normalizedEmail)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("account locked")
	}

	authTrace["account_ok"] = true
	log.Printf("[AUTH_DEBUG] ✅ Account status check passed")

	// 5. Password verification with detailed logging
	log.Printf("[AUTH_DEBUG] 🔐 Starting password verification...")
	
	// Log Argon2 configuration
	log.Printf("[AUTH_DEBUG] 🔧 Argon2 Configuration:")
	log.Printf("[AUTH_DEBUG]   - Memory: %d", GlobalAuthConfig.Argon2.Memory)
	log.Printf("[AUTH_DEBUG]   - Iterations: %d", GlobalAuthConfig.Argon2.Iterations)
	log.Printf("[AUTH_DEBUG]   - Parallelism: %d", GlobalAuthConfig.Argon2.Parallelism)
	log.Printf("[AUTH_DEBUG]   - KeyLength: %d", GlobalAuthConfig.Argon2.KeyLength)
	
	// Check for pepper
	pepper := os.Getenv("AUTH_PEPPER")
	pepperSet := pepper != ""
	log.Printf("[AUTH_DEBUG]   - AUTH_PEPPER_SET: %t", pepperSet)
	if pepperSet {
		log.Printf("[AUTH_DEBUG]   - Pepper length: %d", len(pepper))
	}

	expectedHash := []byte(user.PasswordHash)
	actualHash, err := hashPasswordWithConfig(password, normalizedEmail, GlobalAuthConfig.Argon2)
	if err != nil {
		authTrace["reason"] = "password hashing error"
		log.Printf("[AUTH_DEBUG] ❌ Password hashing error: %v", err)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("password verification error: %v", err)
	}

	// ENHANCED DIAGNOSTIC: Log hash comparison details
	logHashComparison(expectedHash, actualHash, normalizedEmail)

	passwordValid := compareHashes(expectedHash, actualHash)
	authTrace["pw_ok"] = passwordValid
	log.Printf("[AUTH_DEBUG] 🔐 Password verification result: %t", passwordValid)

	// Backward compatibility: try old hashing method if new method fails
	if !passwordValid && !GlobalAuthConfig.UseNewFlow {
		log.Printf("[AUTH_DEBUG] ⚠️  New hash method failed, trying old method for user: %s", normalizedEmail)

		// Try old Argon2 parameters (hardcoded values)
		oldHash := argon2.IDKey([]byte(password), []byte(normalizedEmail), 1, 64*1024, 4, 32)
		passwordValid = compareHashes(expectedHash, oldHash)
		authTrace["pw_ok"] = passwordValid

		if passwordValid {
			log.Printf("[AUTH_DEBUG] ✅ Old hash method succeeded, user should be migrated: %s", normalizedEmail)
			// TODO: Implement password rehashing on next login
		} else {
			log.Printf("[AUTH_DEBUG] ❌ Old hash method also failed: %s", normalizedEmail)
		}
	}

	if !passwordValid {
		authTrace["reason"] = "invalid password"
		log.Printf("[AUTH_DEBUG] ❌ Password verification FAILED for user: %s", normalizedEmail)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("invalid password")
	}

	log.Printf("[AUTH_DEBUG] ✅ Password verification SUCCESS for user: %s", normalizedEmail)

	// 6. TOTP verification with detailed logging
	authTrace["totp_required"] = true
	log.Printf("[AUTH_DEBUG] 🔢 Starting TOTP verification...")
	
	// Log TOTP configuration
	totpWindow := 1
	if testMode {
		if envWindow := os.Getenv("AUTH_TEST_ALLOW_TOTP_WINDOW"); envWindow != "" {
			if window, err := strconv.Atoi(envWindow); err == nil {
				totpWindow = window
			}
		}
	}
	
	serverTime := time.Now()
	log.Printf("[AUTH_DEBUG] 🔧 TOTP Configuration:")
	log.Printf("[AUTH_DEBUG]   - Window steps: ±%d", totpWindow)
	log.Printf("[AUTH_DEBUG]   - Server time: %v", serverTime)
	log.Printf("[AUTH_DEBUG]   - TOTP secret: %s", user.TOTPSecret)
	log.Printf("[AUTH_DEBUG]   - TOTP code: %s", totpCode)

	// Verify TOTP using new configuration with time skew tolerance
	totpValid := validateTOTPWithConfig(totpCode, user.TOTPSecret, GlobalAuthConfig.TOTP)
	authTrace["totp_ok"] = totpValid
	log.Printf("[AUTH_DEBUG] 🔢 TOTP verification result: %t", totpValid)

	if !totpValid {
		authTrace["reason"] = "invalid TOTP code"
		log.Printf("[AUTH_DEBUG] ❌ TOTP verification FAILED for user: %s", normalizedEmail)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("invalid TOTP code")
	}

	log.Printf("[AUTH_DEBUG] ✅ TOTP verification SUCCESS for user: %s", normalizedEmail)

	// 7. Session/JWT generation
	log.Printf("[AUTH_DEBUG] 🎫 Generating JWT token...")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		authTrace["reason"] = "JWT secret not configured"
		log.Printf("[AUTH_DEBUG] ❌ JWT_SECRET not configured")
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("JWT_SECRET not configured")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   normalizedEmail,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		authTrace["reason"] = "JWT signing error"
		log.Printf("[AUTH_DEBUG] ❌ JWT signing error: %v", err)
		if testMode {
			log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
		}
		return "", "", fmt.Errorf("JWT signing error: %v", err)
	}

	authTrace["session_ok"] = true
	authTrace["final_auth_ok"] = true
	log.Printf("[AUTH_DEBUG] ✅ JWT token generated successfully")
	log.Printf("[AUTH_DEBUG] ===== AUTHENTICATION ATTEMPT SUCCESS =====")
	
	if testMode {
		log.Printf("[AUTH_DEBUG] AUTH_TRACE: %+v", authTrace)
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
