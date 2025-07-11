package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func setupResendFallbackTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
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
	return db
}

func TestResendFallbackHandler(t *testing.T) {
	db := setupResendFallbackTestDB(t)
	defer db.Close()

	handler := resendFallbackHandlerFactory(db)

	// Insert a user with unconfirmed fallback
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	futureExpiration := time.Now().Add(1 * time.Hour)
	_, err := db.Exec(`INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration) VALUES (?, ?, ?, ?, ?, ?)`,
		"user1@example.com", string(hashedPassword), "recovery1@example.com", "token1", false, futureExpiration)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Insert a user with confirmed fallback
	_, err = db.Exec(`INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration) VALUES (?, ?, ?, ?, ?, ?)`,
		"user2@example.com", string(hashedPassword), "recovery2@example.com", "token2", true, futureExpiration)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	t.Run("Valid resend case", func(t *testing.T) {
		body := map[string]string{"email": "user1@example.com"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/resend-fallback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if resp["message"] != "Fallback confirmation email sent" {
			t.Errorf("Expected message, got %v", resp)
		}
		// Check that the token was updated
		var newToken string
		err := db.QueryRow("SELECT fallback_token FROM users WHERE email = ?", "user1@example.com").Scan(&newToken)
		if err != nil {
			t.Fatalf("Failed to query token: %v", err)
		}
		if newToken == "token1" || newToken == "" {
			t.Errorf("Expected new token, got %s", newToken)
		}
	})

	t.Run("Already confirmed case", func(t *testing.T) {
		body := map[string]string{"email": "user2@example.com"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/resend-fallback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if resp["error"] != "Fallback already confirmed" {
			t.Errorf("Expected error, got %v", resp)
		}
	})

	t.Run("Invalid email format", func(t *testing.T) {
		body := map[string]string{"email": "not-an-email"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/resend-fallback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if resp["error"] != "Invalid email format" {
			t.Errorf("Expected error, got %v", resp)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		body := map[string]string{"email": "nouser@example.com"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/resend-fallback", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if resp["error"] != "User not found" {
			t.Errorf("Expected error, got %v", resp)
		}
	})

	t.Run("Wrong HTTP method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/resend-fallback", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405, got %d", w.Code)
		}
		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err == nil {
			if resp["error"] != "Method not allowed" {
				t.Errorf("Expected method not allowed error, got %v", resp)
			}
		}
	})
}
