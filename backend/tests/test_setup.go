package tests

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"securemail-backend/auth"
)

// setupTestDB creates a test database with all required migrations
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			email TEXT NOT NULL,
			hashed_password TEXT NOT NULL,
			totp_secret_encrypted BLOB,
			totp_configured BOOLEAN DEFAULT FALSE,
			recovery_codes_hashed JSON,
			public_pqc_key TEXT NULL,
			public_sign_key TEXT NULL,
			encrypted_profile_blob BLOB NULL,
			account_type TEXT NOT NULL DEFAULT 'free',
			account_type_new TEXT DEFAULT 'free',
			account_status TEXT NOT NULL DEFAULT 'pending_verification',
			domain TEXT NULL,
			organization_id TEXT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_domain ON users(username, domain)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			device_info TEXT NULL,
			ip_address TEXT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id)`,
		// Phase 2 migrations
		`ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE`,
		`ALTER TABLE users ADD COLUMN verification_code TEXT NULL`,
		`ALTER TABLE users ADD COLUMN verification_code_expires_at TIMESTAMP NULL`,
		`CREATE INDEX IF NOT EXISTS idx_users_verification_code ON users(verification_code)`,
		`ALTER TABLE users ADD COLUMN totp_secret TEXT NULL`,
		`ALTER TABLE users ADD COLUMN mfa_enabled BOOLEAN DEFAULT FALSE`,
		`ALTER TABLE users ADD COLUMN backup_codes_hashed JSON NULL`,
		`CREATE INDEX IF NOT EXISTS idx_users_mfa_enabled ON users(mfa_enabled)`,
		// Migration 0010: Add fallback email and recovery key fields
		`ALTER TABLE users ADD COLUMN fallback_email TEXT NULL`,
		`ALTER TABLE users ADD COLUMN fallback_email_verified BOOLEAN DEFAULT FALSE`,
		`ALTER TABLE users ADD COLUMN recovery_private_key_hashed TEXT NULL`,
		// Phase 3 tables
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			stripe_customer_id TEXT NULL,
			stripe_subscription_id TEXT NULL,
			status TEXT NOT NULL,
			plan TEXT NOT NULL,
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id)`,
		`CREATE TABLE IF NOT EXISTS analytics_events (
			id TEXT PRIMARY KEY,
			event_type TEXT NOT NULL,
			user_hash TEXT NOT NULL,
			metadata TEXT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			t.Fatalf("Failed to run migration: %v", err)
		}
	}

	// Set global DB variable
	auth.DB = db
	return db
}

// setupPhase2TestDB creates a test database for Phase 2 tests
func setupPhase2TestDB(t *testing.T) *sql.DB {
	return setupTestDB(t)
}

// setupPhase3TestDB creates a test database for Phase 3 tests
func setupPhase3TestDB(t *testing.T) *sql.DB {
	return setupTestDB(t)
}

// setupCompletePhase3TestDB creates a test database for complete Phase 3 tests
func setupCompletePhase3TestDB(t *testing.T) *sql.DB {
	return setupTestDB(t)
}
