package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Test database setup
// var testDB *sql.DB // Reserved for future tests

func setupTestDB() (*sql.DB, error) {
	// Use in-memory SQLite for tests
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Run migrations
	migrationSQL := `
	-- Users table (core)
	CREATE TABLE IF NOT EXISTS users (
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
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_domain ON users(username, domain);

	-- Sessions table
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		device_info TEXT NULL,
		ip_address TEXT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
	`

	_, err = db.Exec(migrationSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func cleanupTestDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

// Mock crypto functions for testing
func mockHashPassword(password string) (string, error) {
	// Simple mock - in real tests you'd use the actual Argon2id implementation
	return fmt.Sprintf("hashed_%s", password), nil
}

func mockVerifyPassword(password, hash string) (bool, error) {
	// Simple mock - in real tests you'd use the actual Argon2id implementation
	expected := fmt.Sprintf("hashed_%s", password)
	return hash == expected, nil
}

func mockGenerateRandomToken(_ int) (string, error) {
	return "test_token_12345", nil
}

func mockHashToken(token string) string {
	return fmt.Sprintf("hashed_%s", token)
}

// Test signup success
func TestSignupSuccess(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	defer cleanupTestDB(db)

	// Set up test environment
	os.Setenv("DATABASE_URL", ":memory:")
	defer os.Unsetenv("DATABASE_URL")

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock signup handler
		var req struct {
			Username    string `json:"username"`
			Email       string `json:"email"`
			Password    string `json:"password"`
			AccountType string `json:"account_type,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Hash password
		hashed, err := mockHashPassword(req.Password)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Insert user
		id := uuid.New().String()
		now := time.Now().UTC()

		_, err = db.Exec(
			`INSERT INTO users (id, username, email, hashed_password, account_type, account_status, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, 'pending_verification', ?, ?)`,
			id, req.Username, req.Email, hashed, req.AccountType, now, now,
		)
		if err != nil {
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}

		// Return response
		resp := map[string]interface{}{
			"id":           id,
			"username":     req.Username,
			"email":        req.Email,
			"account_type": req.AccountType,
			"created_at":   now.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Test signup request
	signupData := map[string]string{
		"username":     "testuser",
		"email":        "test@example.com",
		"password":     "testpassword123",
		"account_type": "free",
	}

	jsonData, _ := json.Marshal(signupData)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// Check database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "test@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}

	// Check password is hashed
	var hashedPassword string
	err = db.QueryRow("SELECT hashed_password FROM users WHERE email = ?", "test@example.com").Scan(&hashedPassword)
	if err != nil {
		t.Fatalf("Failed to query password: %v", err)
	}
	if hashedPassword == "testpassword123" {
		t.Error("Password was not hashed")
	}
}

// Test duplicate signup
func TestSignupDuplicate(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	defer cleanupTestDB(db)

	// Insert existing user
	id := uuid.New().String()
	now := time.Now().UTC()
	_, err = db.Exec(
		`INSERT INTO users (id, username, email, hashed_password, account_type, account_status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 'pending_verification', ?, ?)`,
		id, "testuser", "test@example.com", "hashed_password", "free", now, now,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Check for duplicate
		var exists int
		err := db.QueryRow("SELECT 1 FROM users WHERE email = ? LIMIT 1", req.Email).Scan(&exists)
		if err == nil {
			http.Error(w, "username or email already exists", http.StatusConflict)
			return
		}

		// If not duplicate, proceed with signup
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Test duplicate signup
	signupData := map[string]string{
		"username": "testuser2",
		"email":    "test@example.com", // Same email
		"password": "testpassword123",
	}

	jsonData, _ := json.Marshal(signupData)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", resp.StatusCode)
	}
}

// Test login success
func TestLoginSuccess(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	defer cleanupTestDB(db)

	// Insert test user
	id := uuid.New().String()
	now := time.Now().UTC()
	hashedPassword, _ := mockHashPassword("testpassword123")
	_, err = db.Exec(
		`INSERT INTO users (id, username, email, hashed_password, account_type, account_status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 'pending_verification', ?, ?)`,
		id, "testuser", "test@example.com", hashedPassword, "free", now, now,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Lookup user
		var userID string
		var storedHash string
		var totpConfigured bool
		err := db.QueryRow("SELECT id, hashed_password, totp_configured FROM users WHERE email = ? LIMIT 1", req.Email).Scan(&userID, &storedHash, &totpConfigured)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		ok, err := mockVerifyPassword(req.Password, storedHash)
		if err != nil || !ok {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		// Generate token
		rawToken, _ := mockGenerateRandomToken(32)
		hashedToken := mockHashToken(rawToken)
		expiresAt := time.Now().Add(24 * time.Hour).UTC()

		// Store session
		_, err = db.Exec(`INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`,
			uuid.New().String(), userID, hashedToken, expiresAt, time.Now().UTC(),
		)
		if err != nil {
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}

		// Return response
		resp := map[string]interface{}{
			"token":      rawToken,
			"expires_at": expiresAt.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Test login request
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "testpassword123",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Check database has session
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE user_id = ?", id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 session, got %d", count)
	}
}

// Test login with invalid credentials
func TestLoginInvalidCredentials(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	defer cleanupTestDB(db)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Lookup user (will fail for non-existent user)
		var userID string
		var storedHash string
		var totpConfigured bool
		err := db.QueryRow("SELECT id, hashed_password, totp_configured FROM users WHERE email = ? LIMIT 1", req.Email).Scan(&userID, &storedHash, &totpConfigured)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
	}))
	defer server.Close()

	// Test login with invalid credentials
	loginData := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "wrongpassword",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}
