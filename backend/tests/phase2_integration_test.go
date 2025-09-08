package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"securemail-backend/auth"

	_ "github.com/mattn/go-sqlite3"
)

// Test database setup for Phase 2
func setupPhase2TestDB(t *testing.T) *sql.DB {
	// Create in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations
	migrations := []string{
		// Phase 1 migration
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
		// Phase 3 tables for compatibility
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

func TestEmailVerificationFlow(t *testing.T) {
	db := setupPhase2TestDB(t)
	defer db.Close()

	// Test 1: Signup generates verification code
	t.Run("SignupGeneratesVerificationCode", func(t *testing.T) {
		reqBody := map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
			"password": "testpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.SignupHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		// Check that user was created with verification code
		var verificationCode string
		var expiresAt time.Time
		err := db.QueryRow(`
			SELECT verification_code, verification_code_expires_at 
			FROM users 
			WHERE email = 'test@example.com'
		`).Scan(&verificationCode, &expiresAt)

		if err != nil {
			t.Fatalf("Failed to query verification code: %v", err)
		}

		if verificationCode == "" {
			t.Error("Verification code should not be empty")
		}

		if time.Until(expiresAt) < 25*time.Minute || time.Until(expiresAt) > 30*time.Minute {
			t.Errorf("Verification code expiry should be ~30 minutes, got %v", time.Until(expiresAt))
		}
	})

	// Test 2: Email verification with valid code
	t.Run("EmailVerificationWithValidCode", func(t *testing.T) {
		// First create a user with a known verification code
		hashedCode := auth.HashToken("123456")
		expiresAt := time.Now().Add(30 * time.Minute)

		_, err := db.Exec(`
			INSERT INTO users (id, username, email, hashed_password, verification_code, verification_code_expires_at, email_verified)
			VALUES ('test-id', 'testuser2', 'test2@example.com', 'hashedpass', ?, ?, FALSE)
		`, hashedCode, expiresAt)
		if err != nil {
			t.Fatalf("Failed to insert test user: %v", err)
		}

		reqBody := map[string]string{
			"email": "test2@example.com",
			"code":  "123456",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.VerifyEmailHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check that email is now verified
		var emailVerified bool
		var verificationCode sql.NullString
		err = db.QueryRow(`
			SELECT email_verified, verification_code 
			FROM users 
			WHERE email = 'test2@example.com'
		`).Scan(&emailVerified, &verificationCode)

		if err != nil {
			t.Fatalf("Failed to query email verification status: %v", err)
		}

		if !emailVerified {
			t.Error("Email should be verified")
		}

		if verificationCode.Valid {
			t.Error("Verification code should be cleared after successful verification")
		}
	})

	// Test 3: Email verification with invalid code
	t.Run("EmailVerificationWithInvalidCode", func(t *testing.T) {
		reqBody := map[string]string{
			"email": "test2@example.com",
			"code":  "invalid",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.VerifyEmailHandler(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	// Test 4: Resend verification code
	t.Run("ResendVerificationCode", func(t *testing.T) {
		reqBody := map[string]string{
			"email": "test@example.com",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/resend-verification", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ResendVerificationHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestMFAFlow(t *testing.T) {
	db := setupPhase2TestDB(t)
	defer db.Close()

	// Create a verified user for MFA testing
	hashedPassword, _ := auth.HashPassword("testpassword123")
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, hashed_password, email_verified, mfa_enabled)
		VALUES ('mfa-test-id', 'mfauser', 'mfa@example.com', ?, TRUE, FALSE)
	`, hashedPassword)
	if err != nil {
		t.Fatalf("Failed to insert MFA test user: %v", err)
	}

	// Test 1: Setup MFA
	t.Run("SetupMFA", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/auth/setup-mfa", nil)
		req = req.WithContext(auth.ContextWithUserID(req.Context(), "mfa-test-id"))
		w := httptest.NewRecorder()

		auth.SetupMFAHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["secret"] == "" {
			t.Error("TOTP secret should not be empty")
		}

		if response["qr_code_url"] == "" {
			t.Error("QR code URL should not be empty")
		}

		backupCodes, ok := response["backup_codes"].([]interface{})
		if !ok || len(backupCodes) != 10 {
			t.Errorf("Expected 10 backup codes, got %d", len(backupCodes))
		}

		// Check that MFA is enabled in database
		var mfaEnabled bool
		err := db.QueryRow(`
			SELECT mfa_enabled 
			FROM users 
			WHERE id = 'mfa-test-id'
		`).Scan(&mfaEnabled)

		if err != nil {
			t.Fatalf("Failed to query MFA status: %v", err)
		}

		if !mfaEnabled {
			t.Error("MFA should be enabled")
		}
	})

	// Test 2: Login requires MFA when enabled
	t.Run("LoginRequiresMFA", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "mfa@example.com",
			"password": "testpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.LoginHandler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 (MFA required), got %d", w.Code)
		}

		body, _ := io.ReadAll(w.Body)
		if string(body) != "mfa_required\n" {
			t.Errorf("Expected 'mfa_required', got %s", string(body))
		}
	})
}

func TestLoginWithEmailVerification(t *testing.T) {
	db := setupPhase2TestDB(t)
	defer db.Close()

	// Test 1: Login fails for unverified email
	t.Run("LoginFailsForUnverifiedEmail", func(t *testing.T) {
		hashedPassword, _ := auth.HashPassword("testpassword123")
		_, err := db.Exec(`
			INSERT INTO users (id, username, email, hashed_password, email_verified)
			VALUES ('unverified-id', 'unverified', 'unverified@example.com', ?, FALSE)
		`, hashedPassword)
		if err != nil {
			t.Fatalf("Failed to insert unverified user: %v", err)
		}

		reqBody := map[string]string{
			"email":    "unverified@example.com",
			"password": "testpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.LoginHandler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 (email not verified), got %d", w.Code)
		}

		body, _ := io.ReadAll(w.Body)
		if string(body) != "email_not_verified\n" {
			t.Errorf("Expected 'email_not_verified', got %s", string(body))
		}
	})

	// Test 2: Login succeeds for verified email without MFA
	t.Run("LoginSucceedsForVerifiedEmail", func(t *testing.T) {
		hashedPassword, _ := auth.HashPassword("testpassword123")
		_, err := db.Exec(`
			INSERT INTO users (id, username, email, hashed_password, email_verified, mfa_enabled)
			VALUES ('verified-id', 'verified', 'verified@example.com', ?, TRUE, FALSE)
		`, hashedPassword)
		if err != nil {
			t.Fatalf("Failed to insert verified user: %v", err)
		}

		reqBody := map[string]string{
			"email":    "verified@example.com",
			"password": "testpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.LoginHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["token"] == "" {
			t.Error("Login token should not be empty")
		}
	})
}
