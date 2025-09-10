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
		// Additional tables for complete functionality
		`CREATE TABLE IF NOT EXISTS domains (
			id TEXT PRIMARY KEY,
			domain_name TEXT NOT NULL UNIQUE,
			user_id TEXT NOT NULL,
			verified BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			admin_user_id TEXT NOT NULL,
			domain TEXT NULL,
			max_users INTEGER DEFAULT 100,
			settings_json TEXT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (admin_user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS organization_members (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			invited_by TEXT NULL,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS mailbox_folders (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			folder_type TEXT NOT NULL DEFAULT 'custom',
			parent_folder_id TEXT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS email_messages (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			folder_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			thread_id TEXT NULL,
			from_address TEXT NOT NULL,
			to_addresses TEXT NOT NULL,
			cc_addresses TEXT NULL,
			bcc_addresses TEXT NULL,
			subject TEXT NOT NULL,
			body_encrypted BLOB NOT NULL,
			body_type TEXT NOT NULL DEFAULT 'text/plain',
			attachments_encrypted BLOB NULL,
			headers_encrypted BLOB NULL,
			size_bytes INTEGER NOT NULL DEFAULT 0,
			is_read BOOLEAN DEFAULT FALSE,
			is_important BOOLEAN DEFAULT FALSE,
			is_starred BOOLEAN DEFAULT FALSE,
			received_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (folder_id) REFERENCES mailbox_folders(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS email_labels (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			color TEXT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS message_labels (
			message_id TEXT NOT NULL,
			label_id TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (message_id, label_id),
			FOREIGN KEY (message_id) REFERENCES email_messages(id) ON DELETE CASCADE,
			FOREIGN KEY (label_id) REFERENCES email_labels(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS email_drafts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			to_addresses TEXT NULL,
			cc_addresses TEXT NULL,
			bcc_addresses TEXT NULL,
			subject TEXT NULL,
			body_text TEXT NULL,
			body_html TEXT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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
