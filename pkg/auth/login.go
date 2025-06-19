package auth

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"fmt"
	"regexp"
	"time"

	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

var emailRegex = regexp.MustCompile(`^[^@]+@securesystem.email$`)

// ValidateEmail checks if email matches securesystem.email domain
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidatePassword checks length (8–128 characters)
func ValidatePassword(password string) bool {
	return len(password) >= 8 && len(password) <= 128
}

// ValidateTOTP checks 6-digit format
func ValidateTOTP(code string) bool {
	return len(code) == 6 && regexp.MustCompile(`^\d{6}$`).MatchString(code)
}

// HashPassword creates Argon2 hash with email as salt
func HashPassword(password, email string) (string, error) {
	// Hash password using Argon2 with email as salt
	hash := argon2.IDKey([]byte(password), []byte(email), 1, 64*1024, 4, 32)
	return string(hash), nil
}

// GenerateTOTPSecret creates a new base32 TOTP secret
func GenerateTOTPSecret() (string, error) {
	// Generate 20 random bytes for TOTP secret
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %v", err)
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

// Authenticate verifies credentials and returns JWT
func Authenticate(db *sql.DB, email, password, totpCode string) (string, string, error) {
	// Validate inputs
	if !ValidateEmail(email) {
		return "", "", fmt.Errorf("invalid email format")
	}
	if !ValidatePassword(password) {
		return "", "", fmt.Errorf("invalid password length")
	}
	if !ValidateTOTP(totpCode) {
		return "", "", fmt.Errorf("invalid TOTP format")
	}

	// Query user
	var user struct {
		ID           string
		PasswordHash string
		TOTPSecret   string
	}
	err := db.QueryRow("SELECT id, password_hash, totp_secret FROM users WHERE email = ?", email).Scan(&user.ID, &user.PasswordHash, &user.TOTPSecret)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("user not found")
		}
		return "", "", fmt.Errorf("database error: %v", err)
	}

	// Verify password with Argon2
	expectedHash := []byte(user.PasswordHash)
	actualHash := argon2.IDKey([]byte(password), []byte(email), 1, 64*1024, 4, 32)
	if !bytes.Equal(expectedHash, actualHash) {
		return "", "", fmt.Errorf("invalid password")
	}

	// Verify TOTP
	if !totp.Validate(totpCode, user.TOTPSecret) {
		return "", "", fmt.Errorf("invalid TOTP code")
	}

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

// CreateUser creates a new user with hashed password and TOTP secret
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

// ValidateJWT validates and parses JWT token
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
