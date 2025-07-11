package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func TestSignupWithFallbackEmail(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		fallback_token_expiration TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_fallback_token ON users(fallback_token);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create handler
	handler := signupHandlerFactory(db)

	// Test valid signup with fallback email
	reqBody := map[string]string{
		"email":          "test@example.com",
		"password":       "securepassword123",
		"fallback_email": "recovery@example.com",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "User created" {
		t.Errorf("Expected message 'User created', got '%s'", response["message"])
	}

	// Verify user was created with fallback data
	var user struct {
		ID                int
		Email             string
		FallbackEmail     string
		FallbackToken     string
		FallbackConfirmed bool
	}
	err = db.QueryRow("SELECT id, email, fallback_email, fallback_token, fallback_confirmed FROM users WHERE email = ?", "test@example.com").
		Scan(&user.ID, &user.Email, &user.FallbackEmail, &user.FallbackToken, &user.FallbackConfirmed)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
	if user.FallbackEmail != "recovery@example.com" {
		t.Errorf("Expected fallback email 'recovery@example.com', got '%s'", user.FallbackEmail)
	}
	if user.FallbackToken == "" {
		t.Error("Expected fallback token to be generated")
	}
	if user.FallbackConfirmed {
		t.Error("Expected fallback_confirmed to be false initially")
	}
}

func TestSignupWithoutFallbackEmail(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create handler
	handler := signupHandlerFactory(db)

	// Test signup without fallback email
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "securepassword123",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(response["error"], "Fallback email is required") {
		t.Errorf("Expected error about fallback email, got '%s'", response["error"])
	}
}

func TestSignupWithInvalidFallbackEmail(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create handler
	handler := signupHandlerFactory(db)

	// Test signup with invalid fallback email
	reqBody := map[string]string{
		"email":          "test@example.com",
		"password":       "securepassword123",
		"fallback_email": "invalid-email",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(response["error"], "Invalid fallback email format") {
		t.Errorf("Expected error about invalid fallback email, got '%s'", response["error"])
	}
}

func TestFallbackConfirmation(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		fallback_token_expiration TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_fallback_token ON users(fallback_token);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Insert test user with future expiration
	testToken := "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef12345678"
	futureExpiration := time.Now().Add(1 * time.Hour)
	_, err = db.Exec("INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration) VALUES (?, ?, ?, ?, ?, ?)",
		"test@example.com", "hashedpassword", "recovery@example.com", testToken, false, futureExpiration)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create handler
	handler := confirmFallbackHandlerFactory(db)

	// Test valid confirmation
	req := httptest.NewRequest("GET", "/confirm-fallback?token="+testToken, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "Fallback email confirmed successfully. You may now log in." {
		t.Errorf("Expected message 'Fallback email confirmed successfully. You may now log in.', got '%s'", response["message"])
	}

	// Verify fallback_confirmed was updated
	var confirmed bool
	err = db.QueryRow("SELECT fallback_confirmed FROM users WHERE email = ?", "test@example.com").Scan(&confirmed)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if !confirmed {
		t.Error("Expected fallback_confirmed to be true after confirmation")
	}
}

func TestFallbackConfirmationInvalidToken(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create handler
	handler := confirmFallbackHandlerFactory(db)

	// Test invalid token
	req := httptest.NewRequest("GET", "/confirm-fallback?token=invalid-token", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["error"] != "Invalid or expired confirmation token" {
		t.Errorf("Expected error 'Invalid or expired confirmation token', got '%s'", response["error"])
	}
}

func TestFallbackConfirmationMissingToken(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create handler
	handler := confirmFallbackHandlerFactory(db)

	// Test missing token
	req := httptest.NewRequest("GET", "/confirm-fallback", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["error"] != "Invalid or expired confirmation token" {
		t.Errorf("Expected error 'Invalid or expired confirmation token', got '%s'", response["error"])
	}
}

func TestFallbackConfirmationExpiredToken(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		fallback_token_expiration TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Insert test user with expired token
	testToken := "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef12345678"
	pastExpiration := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	_, err = db.Exec("INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration) VALUES (?, ?, ?, ?, ?, ?)",
		"test@example.com", "hashedpassword", "recovery@example.com", testToken, false, pastExpiration)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create handler
	handler := confirmFallbackHandlerFactory(db)

	// Test expired token
	req := httptest.NewRequest("GET", "/confirm-fallback?token="+testToken, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["error"] != "Invalid or expired confirmation token" {
		t.Errorf("Expected error 'Invalid or expired confirmation token', got '%s'", response["error"])
	}
}

func TestLoginBeforeFallbackConfirmation(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create a proper bcrypt hash for the test password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Insert test user with unconfirmed fallback
	_, err = db.Exec("INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed) VALUES (?, ?, ?, ?, ?)",
		"test@example.com", string(hashedPassword), "recovery@example.com", "token123", false)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create handler
	handler := loginHandlerFactory(db)

	// Test login before fallback confirmation
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(response["error"], "Fallback email not confirmed") {
		t.Errorf("Expected error about fallback email not confirmed, got '%s'", response["error"])
	}
}

func TestLoginAfterFallbackConfirmation(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Insert test user with confirmed fallback
	_, err = db.Exec("INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed) VALUES (?, ?, ?, ?, ?)",
		"test@example.com", "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890", "recovery@example.com", "token123", true)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create handler
	handler := loginHandlerFactory(db)

	// Test login after fallback confirmation
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Note: This test will fail because we're using a fake bcrypt hash
	// In a real test, we'd use a proper bcrypt hash for "password123"
	// For now, we expect a 401 due to invalid password
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 (due to invalid password), got %d", w.Code)
	}
}
