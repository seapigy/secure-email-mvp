package testbypass

import (
	"database/sql"
	"log"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/security"
)

// SeedTestUser creates a test user account when test bypass is enabled
func SeedTestUser(db *sql.DB, config *Config) error {
	if !config.Enabled {
		return nil
	}

	log.Printf("[TEST_BYPASS] 🔧 Test bypass mode enabled - seeding test user account")

	// Check if test user already exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", config.TestEmail).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("[TEST_BYPASS] ✅ Test user already exists: %s", config.TestEmail)
		return nil
	}

	// Generate password hash
	passwordHash, err := auth.HashPassword(config.TestPassword, config.TestEmail)
	if err != nil {
		return err
	}

	// Generate TOTP secret
	totpSecret, err := auth.GenerateTOTPSecret()
	if err != nil {
		return err
	}

	// Generate fallback token
	fallbackToken, err := security.GenerateSecureToken(64)
	if err != nil {
		return err
	}
	fallbackExpiration := time.Now().Add(24 * time.Hour)

	// Insert test user
	_, err = db.Exec(`
		INSERT INTO users (
			id, email, password, password_hash, totp_secret, 
			fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		config.TestUserID,
		config.TestEmail,
		"", // Empty password field (we use password_hash)
		passwordHash,
		totpSecret,
		config.TestEmail, // Use same email as fallback
		fallbackToken,
		true, // Mark as confirmed
		fallbackExpiration,
		time.Now(),
	)

	if err != nil {
		return err
	}

	log.Printf("[TEST_BYPASS] ✅ Test user created successfully:")
	log.Printf("[TEST_BYPASS]   Email: %s", config.TestEmail)
	log.Printf("[TEST_BYPASS]   Password: %s", config.TestPassword)
	log.Printf("[TEST_BYPASS]   User ID: %s", config.TestUserID)
	log.Printf("[TEST_BYPASS]   TOTP Secret: %s", totpSecret)
	log.Printf("[TEST_BYPASS] ⚠️  WARNING: This is for testing only - do not use in production!")

	return nil
}
