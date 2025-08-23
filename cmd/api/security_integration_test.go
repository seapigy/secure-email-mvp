package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SecurityIntegrationTestSuite tests all security features in isolation and combination
type SecurityIntegrationTestSuite struct {
	server *httptest.Server
	db     *sql.DB
	userID string
	token  string
}

// setupTestSuite initializes the test environment
func setupTestSuite(t *testing.T) *SecurityIntegrationTestSuite {
	// Create temporary database
	dbPath := fmt.Sprintf("./test_security_%s.db", uuid.New().String())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Apply users schema
	usersSchema, err := os.ReadFile("../../schema/users.sql")
	if err != nil {
		t.Fatalf("Failed to read users schema: %v", err)
	}
	if _, err := db.Exec(string(usersSchema)); err != nil {
		t.Fatalf("Failed to apply users schema: %v", err)
	}

	// Apply emails schema
	emailsSchema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read emails schema: %v", err)
	}
	if _, err := db.Exec(string(emailsSchema)); err != nil {
		t.Fatalf("Failed to apply emails schema: %v", err)
	}

	// Create test user
	userID := uuid.New().String()
	email := "test@securesystem.email"
	password := "testpassword123"

	passwordHash, err := auth.HashPassword(password, email)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, totp_secret, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		userID, email, passwordHash, "JBSWY3DPEHPK3PXP",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create server instance
	srv := &Server{
		db:         db,
		rateLimits: &sync.Map{},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add user context for testing
		type testContextKey string
		ctx := context.WithValue(r.Context(), testContextKey("user_id"), userID)
		r = r.WithContext(ctx)

		// Route requests to appropriate handlers
		switch r.URL.Path {
		case "/api/auth/login":
			srv.loginHandler(w, r)
		case "/api/email/send":
			srv.sendEmailHandler(w, r)
		case "/api/email/get":
			srv.getEmailHandler(w, r)
		case "/api/email/list":
			srv.listEmailHandler(w, r)
		case "/admin/manual-cleanup":
			srv.adminManualCleanupHandler(w, r)
		case "/admin/email-retention-stats":
			srv.adminEmailRetentionStatsHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))

	return &SecurityIntegrationTestSuite{
		server: server,
		db:     db,
		userID: userID,
	}
}

// teardownTestSuite cleans up test resources
func (suite *SecurityIntegrationTestSuite) teardownTestSuite() {
	suite.server.Close()
	suite.db.Close()
	// Note: SQLite database cleanup is handled automatically by the OS
	// when the process terminates, so we don't need to manually remove the file
}

// authenticateUser performs login and returns JWT token
func (suite *SecurityIntegrationTestSuite) authenticateUser(t *testing.T) string {
	loginData := map[string]interface{}{
		"email":     "test@securesystem.email",
		"password":  "testpassword123",
		"totp_code": "123456", // Mock TOTP for testing
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", suite.server.URL+"/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Authentication failed: %d", resp.StatusCode)
	}

	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	return loginResp["token"].(string)
}

// TestAuthenticationAndAuthorization tests login and protected endpoint access
func TestAuthenticationAndAuthorization(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.teardownTestSuite()

	// Test valid login
	t.Run("ValidLogin", func(t *testing.T) {
		token := suite.authenticateUser(t)
		if token == "" {
			t.Error("Expected valid JWT token")
		}
		suite.token = token
	})

	// Test invalid credentials
	t.Run("InvalidCredentials", func(t *testing.T) {
		loginData := map[string]interface{}{
			"email":     "test@securesystem.email",
			"password":  "wrongpassword",
			"totp_code": "123456",
		}

		jsonData, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", suite.server.URL+"/api/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	// Test protected endpoint access
	t.Run("ProtectedEndpointAccess", func(t *testing.T) {
		req, _ := http.NewRequest("GET", suite.server.URL+"/api/email/list", nil)
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to access protected endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}

// TestEmailExpiration tests email expiration functionality
func TestEmailExpiration(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.teardownTestSuite()

	suite.token = suite.authenticateUser(t)

	// Test sending email with expiration
	t.Run("SendEmailWithExpiration", func(t *testing.T) {
		expirationTime := time.Now().Add(1 * time.Minute).Format("2006-01-02T15:04:05Z")
		emailData := map[string]interface{}{
			"recipient": "recipient@example.com",
			"subject":   "Test Expired Email",
			"body":      "This email will expire soon",
			"expiresAt": expirationTime,
		}

		jsonData, _ := json.Marshal(emailData)
		req, _ := http.NewRequest("POST", suite.server.URL+"/api/email/send", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send email: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	// Test accessing expired email
	t.Run("AccessExpiredEmail", func(t *testing.T) {
		// Create an expired email in the database
		expiredTime := time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05")
		_, err := suite.db.Exec(`
			INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, sha256_hash, expires_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"expired-email-1", suite.userID, "recipient@example.com", "Expired Email",
			"blob-url", "encrypted-key", "hash", expiredTime,
		)
		if err != nil {
			t.Fatalf("Failed to create expired email: %v", err)
		}

		// Try to access the expired email
		accessData := map[string]interface{}{
			"emailId": "expired-email-1",
		}

		jsonData, _ := json.Marshal(accessData)
		req, _ := http.NewRequest("POST", suite.server.URL+"/api/email/get", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to access expired email: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusGone {
			t.Errorf("Expected 410 Gone, got %d", resp.StatusCode)
		}
	})
}

// TestBurnAfterRead tests burn-after-read functionality
func TestBurnAfterRead(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.teardownTestSuite()

	suite.token = suite.authenticateUser(t)

	// Test sending burn-after-read email
	t.Run("SendBurnAfterReadEmail", func(t *testing.T) {
		emailData := map[string]interface{}{
			"recipient":     "recipient@example.com",
			"subject":       "Test Burn-After-Read Email",
			"body":          "This email will be deleted after reading",
			"burnAfterRead": true,
		}

		jsonData, _ := json.Marshal(emailData)
		req, _ := http.NewRequest("POST", suite.server.URL+"/api/email/send", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send burn-after-read email: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	// Test accessing burn-after-read email
	t.Run("AccessBurnAfterReadEmail", func(t *testing.T) {
		// Create a burn-after-read email in the database
		_, err := suite.db.Exec(`
			INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, sha256_hash, burn_after_read, access_count)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"burn-email-1", suite.userID, "recipient@example.com", "Burn Email",
			"blob-url", "encrypted-key", "hash", 1, 0,
		)
		if err != nil {
			t.Fatalf("Failed to create burn-after-read email: %v", err)
		}

		// Access the email (should succeed)
		accessData := map[string]interface{}{
			"emailId": "burn-email-1",
		}

		jsonData, _ := json.Marshal(accessData)
		req, _ := http.NewRequest("POST", suite.server.URL+"/api/email/get", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to access burn-after-read email: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		// Try to access again (should fail - email deleted)
		resp2, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to access deleted email: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 Not Found, got %d", resp2.StatusCode)
		}
	})
}

// TestFailedAttempts tests failed access attempt handling
func TestFailedAttempts(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.teardownTestSuite()

	suite.token = suite.authenticateUser(t)

	// Test failed access attempts
	t.Run("FailedAccessAttempts", func(t *testing.T) {
		// Create a password-protected email
		_, err := suite.db.Exec(`
			INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, sha256_hash, requires_password, password_hash, fail_count)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"protected-email-1", suite.userID, "recipient@example.com", "Protected Email",
			"blob-url", "encrypted-key", "hash", 1, "password-hash", 0,
		)
		if err != nil {
			t.Fatalf("Failed to create protected email: %v", err)
		}

		// Make failed access attempts
		for i := 0; i < 3; i++ {
			accessData := map[string]interface{}{
				"emailId":  "protected-email-1",
				"password": "wrongpassword",
			}

			jsonData, _ := json.Marshal(accessData)
			req, _ := http.NewRequest("POST", suite.server.URL+"/api/email/get", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make failed attempt %d: %v", i+1, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
				t.Errorf("Expected 401 or 403, got %d", resp.StatusCode)
			}
		}

		// Make one more attempt - should trigger deletion
		accessData := map[string]interface{}{
			"emailId":  "protected-email-1",
			"password": "wrongpassword",
		}

		jsonData, _ := json.Marshal(accessData)
		req, _ := http.NewRequest("POST", suite.server.URL+"/api/email/get", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make final attempt: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusGone {
			t.Errorf("Expected 410 Gone, got %d", resp.StatusCode)
		}
	})
}

// TestCleanupWorker tests the cleanup worker functionality
func TestCleanupWorker(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.teardownTestSuite()

	suite.token = suite.authenticateUser(t)

	// Test manual cleanup trigger
	t.Run("ManualCleanupTrigger", func(t *testing.T) {
		cleanupData := map[string]interface{}{
			"dry_run": true,
		}

		jsonData, _ := json.Marshal(cleanupData)
		req, _ := http.NewRequest("POST", suite.server.URL+"/admin/manual-cleanup", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to trigger manual cleanup: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	// Test cleanup statistics
	t.Run("CleanupStatistics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", suite.server.URL+"/admin/email-retention-stats", nil)
		req.Header.Set("Authorization", "Bearer "+suite.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to get cleanup statistics: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.teardownTestSuite()

	// Test rate limiting enforcement
	t.Run("RateLimitingEnforcement", func(t *testing.T) {
		rateLimitHits := 0
		for i := 0; i < 15; i++ {
			req, _ := http.NewRequest("GET", suite.server.URL+"/health", nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request %d: %v", i+1, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				rateLimitHits++
			}

			time.Sleep(100 * time.Millisecond)
		}

		if rateLimitHits == 0 {
			t.Error("Expected rate limiting to be triggered")
		}
	})
}

// TestConcurrentAccess tests concurrent access handling
func TestConcurrentAccess(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.teardownTestSuite()

	suite.token = suite.authenticateUser(t)

	// Test concurrent email access
	t.Run("ConcurrentEmailAccess", func(t *testing.T) {
		successCount := 0
		errorCount := 0

		// Create multiple concurrent requests
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				req, _ := http.NewRequest("GET", suite.server.URL+"/api/email/list", nil)
				req.Header.Set("Authorization", "Bearer "+suite.token)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					errorCount++
				} else {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						successCount++
					} else {
						errorCount++
					}
				}
				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < 5; i++ {
			<-done
		}

		if successCount == 0 {
			t.Error("Expected at least some successful concurrent requests")
		}
	})
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.teardownTestSuite()

	// Test malformed requests
	t.Run("MalformedRequests", func(t *testing.T) {
		malformedData := map[string]interface{}{
			"invalid_field": "invalid_value",
		}

		jsonData, _ := json.Marshal(malformedData)
		req, _ := http.NewRequest("POST", suite.server.URL+"/api/email/send", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make malformed request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", resp.StatusCode)
		}
	})

	// Test invalid token
	t.Run("InvalidToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", suite.server.URL+"/api/email/list", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request with invalid token: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}
	})
}

// TestSecurityIntegration runs all security integration tests
func TestSecurityIntegration(t *testing.T) {
	t.Run("AuthenticationAndAuthorization", TestAuthenticationAndAuthorization)
	t.Run("EmailExpiration", TestEmailExpiration)
	t.Run("BurnAfterRead", TestBurnAfterRead)
	t.Run("FailedAttempts", TestFailedAttempts)
	t.Run("CleanupWorker", TestCleanupWorker)
	t.Run("RateLimiting", TestRateLimiting)
	t.Run("ConcurrentAccess", TestConcurrentAccess)
	t.Run("EdgeCases", TestEdgeCases)
}
