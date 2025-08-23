package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
)

func TestPasswordResetHandlers(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			reset_token TEXT,
			reset_token_expiration DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert test user
	_, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	t.Run("Test Initiate Password Reset", func(t *testing.T) {
		handler := initiatePasswordResetHandler(db)
		reqBody := map[string]string{"email": "test@example.com"}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/password-reset/initiate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify reset token was generated
		var resetToken sql.NullString
		err := db.QueryRow("SELECT reset_token FROM users WHERE email = ?", "test@example.com").Scan(&resetToken)
		if err != nil {
			t.Fatalf("Failed to query reset token: %v", err)
		}

		if !resetToken.Valid {
			t.Error("Reset token was not generated")
		}
	})

	t.Run("Test Complete Password Reset", func(t *testing.T) {
		// First generate a reset token
		handler := initiatePasswordResetHandler(db)
		reqBody := map[string]string{"email": "test@example.com"}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/password-reset/initiate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Get the generated reset token
		var resetToken string
		err := db.QueryRow("SELECT reset_token FROM users WHERE email = ?", "test@example.com").Scan(&resetToken)
		if err != nil {
			t.Fatalf("Failed to get reset token: %v", err)
		}

		// Test password reset completion
		handler = passwordResetHandler(db)
		reqBody = map[string]string{
			"email":        "test@example.com",
			"new_password": "NewSecurePassword123!",
			"reset_token":  resetToken,
		}
		jsonBody, _ = json.Marshal(reqBody)

		req = httptest.NewRequest("POST", "/api/auth/password-reset/complete", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify reset token was cleared
		var clearedToken sql.NullString
		err = db.QueryRow("SELECT reset_token FROM users WHERE email = ?", "test@example.com").Scan(&clearedToken)
		if err != nil {
			t.Fatalf("Failed to query cleared reset token: %v", err)
		}

		if clearedToken.Valid {
			t.Error("Reset token was not cleared after password reset")
		}
	})
}
