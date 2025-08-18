package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"

	_ "modernc.org/sqlite"
)

// TestSendEmailTrackingFields tests that tracking fields are properly returned in send email response
func TestSendEmailTrackingFields(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-send-email-tracking-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schemas
	emailsSchema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read emails schema: %v", err)
	}

	usersSchema, err := os.ReadFile("../../schema/users.sql")
	if err != nil {
		t.Fatalf("Failed to read users schema: %v", err)
	}

	_, err = db.Exec(string(usersSchema))
	if err != nil {
		t.Fatalf("Failed to apply users schema: %v", err)
	}

	_, err = db.Exec(string(emailsSchema))
	if err != nil {
		t.Fatalf("Failed to apply emails schema: %v", err)
	}

	// Create test user
	testUserID := "test-user-tracking"
	_, err = db.Exec(`INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)`,
		testUserID, "test@example.com", "hashed-password", "test-totp-secret")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")
	
	// Set test environment to skip R2 uploads
	os.Setenv("TEST_MODE", "1")
	defer os.Unsetenv("TEST_MODE")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Test cases
	testCases := []struct {
		name           string
		requestBody    SendEmailRequest
		expectedFields map[string]interface{}
	}{
		{
			name: "Basic email without tracking",
			requestBody: SendEmailRequest{
				Recipient: "recipient@test.com",
				Subject:   "Test Subject",
				Body:      "Test body content",
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": nil,
				"access_count":    0,
				"max_attempts":    nil,
			},
		},
		{
			name: "Email with burn after read",
			requestBody: SendEmailRequest{
				Recipient:     "recipient@test.com",
				Subject:       "Test Subject",
				Body:          "Test body content",
				BurnAfterRead: true,
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": true,
				"access_count":    0,
				"max_attempts":    nil,
			},
		},
		{
			name: "Email with self destruct",
			requestBody: SendEmailRequest{
				Recipient:                 "recipient@test.com",
				Subject:                   "Test Subject",
				Body:                      "Test body content",
				SelfDestructAfterAttempts: true,
				MaxFailedAttempts:         5,
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": nil,
				"access_count":    0,
				"max_attempts":    5,
			},
		},
		{
			name: "Email with both burn after read and self destruct",
			requestBody: SendEmailRequest{
				Recipient:                 "recipient@test.com",
				Subject:                   "Test Subject",
				Body:                      "Test body content",
				BurnAfterRead:             true,
				SelfDestructAfterAttempts: true,
				MaxFailedAttempts:         3,
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": true,
				"access_count":    0,
				"max_attempts":    3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			requestBody, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("POST", "/api/email/send", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tokenString)

			// Create response recorder
			w := httptest.NewRecorder()

			// Set up router
			router := mux.NewRouter()
			router.Handle("/api/email/send", jwtMiddleware(http.HandlerFunc(srv.sendEmailHandler))).Methods("POST")
			router.ServeHTTP(w, req)

			// Check response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Parse response
			var response SendEmailResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Verify tracking fields
			if response.BurnAfterRead != nil && *response.BurnAfterRead != tc.expectedFields["burn_after_read"] {
				t.Errorf("Expected burn_after_read %v, got %v", tc.expectedFields["burn_after_read"], *response.BurnAfterRead)
			}
			if response.BurnAfterRead == nil && tc.expectedFields["burn_after_read"] != nil {
				t.Errorf("Expected burn_after_read %v, got nil", tc.expectedFields["burn_after_read"])
			}

			if response.AccessCount != nil && *response.AccessCount != tc.expectedFields["access_count"] {
				t.Errorf("Expected access_count %v, got %v", tc.expectedFields["access_count"], *response.AccessCount)
			}

			if response.MaxAttempts != nil && *response.MaxAttempts != tc.expectedFields["max_attempts"] {
				t.Errorf("Expected max_attempts %v, got %v", tc.expectedFields["max_attempts"], *response.MaxAttempts)
			}
			if response.MaxAttempts == nil && tc.expectedFields["max_attempts"] != nil {
				t.Errorf("Expected max_attempts %v, got nil", tc.expectedFields["max_attempts"])
			}

			// Verify basic response fields
			if response.Status != "success" {
				t.Errorf("Expected status 'success', got '%s'", response.Status)
			}
			if response.BlobID == "" {
				t.Error("Expected blob_id to be present")
			}
		})
	}
}

// TestEmailDetailTrackingFields tests that tracking fields are properly returned in email detail response
func TestEmailDetailTrackingFields(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-email-detail-tracking-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schemas
	emailsSchema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read emails schema: %v", err)
	}

	usersSchema, err := os.ReadFile("../../schema/users.sql")
	if err != nil {
		t.Fatalf("Failed to read users schema: %v", err)
	}

	_, err = db.Exec(string(usersSchema))
	if err != nil {
		t.Fatalf("Failed to apply users schema: %v", err)
	}

	_, err = db.Exec(string(emailsSchema))
	if err != nil {
		t.Fatalf("Failed to apply emails schema: %v", err)
	}

	// Create test user
	testUserID := "test-user-detail-tracking"
	_, err = db.Exec(`INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)`,
		testUserID, "test@example.com", "hashed-password", "test-totp-secret")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")
	
	// Set test environment to skip R2 uploads
	os.Setenv("TEST_MODE", "1")
	defer os.Unsetenv("TEST_MODE")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Test cases
	testCases := []struct {
		name           string
		emailData      map[string]interface{}
		expectedFields map[string]interface{}
	}{
		{
			name: "Basic email without tracking",
			emailData: map[string]interface{}{
				"email_id":        "email-basic",
				"burn_after_read": 0,
				"access_count":    0,
				"max_attempts":    0,
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": nil,
				"access_count":    0,
				"max_attempts":    nil,
			},
		},
		{
			name: "Email with burn after read",
			emailData: map[string]interface{}{
				"email_id":        "email-burn",
				"burn_after_read": 1,
				"access_count":    0,
				"max_attempts":    0,
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": true,
				"access_count":    0,
				"max_attempts":    nil,
			},
		},
		{
			name: "Email with access count",
			emailData: map[string]interface{}{
				"email_id":        "email-accessed",
				"burn_after_read": 0,
				"access_count":    3,
				"max_attempts":    0,
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": nil,
				"access_count":    3,
				"max_attempts":    nil,
			},
		},
		{
			name: "Email with max attempts",
			emailData: map[string]interface{}{
				"email_id":        "email-max-attempts",
				"burn_after_read": 0,
				"access_count":    0,
				"max_attempts":    5,
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": nil,
				"access_count":    0,
				"max_attempts":    5,
			},
		},
		{
			name: "Email with all tracking fields",
			emailData: map[string]interface{}{
				"email_id":        "email-all-tracking",
				"burn_after_read": 1,
				"access_count":    2,
				"max_attempts":    3,
			},
			expectedFields: map[string]interface{}{
				"burn_after_read": true,
				"access_count":    2,
				"max_attempts":    3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test email
			emailID := tc.emailData["email_id"].(string)
			_, err = db.Exec(`
				INSERT INTO emails (
					email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
					encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at,
					burn_after_read, access_count, max_attempts
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				emailID, testUserID, "recipient@test.com", "Test Subject", "blob-url",
				"dGVzdC1rZXk=", "dGVzdC1ub25jZQ==", "dGVzdC1hdXRoLXRhZw==", "gzip", "dGVzdC1oYXNo", time.Now(),
				tc.emailData["burn_after_read"], tc.emailData["access_count"], tc.emailData["max_attempts"],
			)
			if err != nil {
				t.Fatalf("Failed to insert test email: %v", err)
			}

			// Create request
			req := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)

			// Create response recorder
			w := httptest.NewRecorder()

			// Set up router with URL parameters
			router := mux.NewRouter()
			router.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
			router.ServeHTTP(w, req)

			// Check response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Parse response
			var response EmailDetailResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Verify tracking fields
			if response.BurnAfterRead != nil && *response.BurnAfterRead != tc.expectedFields["burn_after_read"] {
				t.Errorf("Expected burn_after_read %v, got %v", tc.expectedFields["burn_after_read"], *response.BurnAfterRead)
			}
			if response.BurnAfterRead == nil && tc.expectedFields["burn_after_read"] != nil {
				t.Errorf("Expected burn_after_read %v, got nil", tc.expectedFields["burn_after_read"])
			}

			if response.AccessCount != nil && *response.AccessCount != tc.expectedFields["access_count"] {
				t.Errorf("Expected access_count %v, got %v", tc.expectedFields["access_count"], *response.AccessCount)
			}

			if response.MaxAttempts != nil && *response.MaxAttempts != tc.expectedFields["max_attempts"] {
				t.Errorf("Expected max_attempts %v, got %v", tc.expectedFields["max_attempts"], *response.MaxAttempts)
			}
			if response.MaxAttempts == nil && tc.expectedFields["max_attempts"] != nil {
				t.Errorf("Expected max_attempts %v, got nil", tc.expectedFields["max_attempts"])
			}

			// Clean up test email
			db.Exec("DELETE FROM emails WHERE email_id = ?", emailID)
		})
	}
}

// TestTrackingFieldsPermission tests that tracking fields are only returned to the sender
func TestTrackingFieldsPermission(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-tracking-permission-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schemas
	emailsSchema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read emails schema: %v", err)
	}

	usersSchema, err := os.ReadFile("../../schema/users.sql")
	if err != nil {
		t.Fatalf("Failed to read users schema: %v", err)
	}

	_, err = db.Exec(string(usersSchema))
	if err != nil {
		t.Fatalf("Failed to apply users schema: %v", err)
	}

	_, err = db.Exec(string(emailsSchema))
	if err != nil {
		t.Fatalf("Failed to apply emails schema: %v", err)
	}

	// Create test users
	senderID := "sender-user"
	unauthorizedUserID := "unauthorized-user"

		_, err = db.Exec(`INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)`,
		senderID, "sender@example.com", "hashed-password", "test-totp-secret")
	if err != nil {
		t.Fatalf("Failed to create sender user: %v", err)
	}
	
	_, err = db.Exec(`INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)`,
		unauthorizedUserID, "unauthorized@example.com", "hashed-password", "test-totp-secret")
	if err != nil {
		t.Fatalf("Failed to create unauthorized user: %v", err)
	}

	// Create test email with tracking fields
	emailID := "email-permission-test"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at,
			burn_after_read, access_count, max_attempts
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "recipient@test.com", "Test Subject", "blob-url",
		"dGVzdC1rZXk=", "dGVzdC1ub25jZQ==", "dGVzdC1hdXRoLXRhZw==", "gzip", "dGVzdC1oYXNo", time.Now(),
		1, 2, 3, // burn_after_read=1, access_count=2, max_attempts=3
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")
	
	// Set test environment to skip R2 uploads
	os.Setenv("TEST_MODE", "1")
	defer os.Unsetenv("TEST_MODE")

	// Test unauthorized access
	t.Run("Unauthorized access should return 403", func(t *testing.T) {
		// Generate JWT token for unauthorized user
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": unauthorizedUserID,
			"email":   "unauthorized@example.com",
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
			"iat":     time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
		if err != nil {
			t.Fatalf("Failed to generate JWT token: %v", err)
		}

		// Create request
		req := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		// Create response recorder
		w := httptest.NewRecorder()

		// Set up router with URL parameters
		router := mux.NewRouter()
		router.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
		router.ServeHTTP(w, req)

		// Check response
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}
	})

	// Test authorized access
	t.Run("Authorized access should return tracking fields", func(t *testing.T) {
		// Generate JWT token for sender
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": senderID,
			"email":   "sender@example.com",
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
			"iat":     time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
		if err != nil {
			t.Fatalf("Failed to generate JWT token: %v", err)
		}

		// Create request
		req := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		// Create response recorder
		w := httptest.NewRecorder()

		// Set up router with URL parameters
		router := mux.NewRouter()
		router.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
		router.ServeHTTP(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())
			return
		}

		// Parse response
		var response EmailDetailResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify tracking fields are present for authorized user
		if response.BurnAfterRead == nil || !*response.BurnAfterRead {
			t.Error("Expected burn_after_read to be true for authorized user")
		}
		if response.AccessCount == nil || *response.AccessCount != 2 {
			t.Errorf("Expected access_count to be 2, got %v", response.AccessCount)
		}
		if response.MaxAttempts == nil || *response.MaxAttempts != 3 {
			t.Errorf("Expected max_attempts to be 3, got %v", response.MaxAttempts)
		}
	})
}
