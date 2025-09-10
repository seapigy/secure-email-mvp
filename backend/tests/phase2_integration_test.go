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

// Test database setup - using shared setupPhase2TestDB function from test_setup.go

func TestEmailVerificationFlow(t *testing.T) {
	db := setupPhase2TestDB(t)
	defer db.Close()

	// Test 1: Signup generates verification code
	t.Run("SignupGeneratesVerificationCode", func(t *testing.T) {
		reqBody := map[string]string{
			"username":       "testuser",
			"email":          "test@example.com",
			"password":       "TestPassword123!",
			"fallback_email": "backup@example.com",
			"account_type":   "free",
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
	// TODO: Implement ResendVerificationHandler
	// t.Run("ResendVerificationCode", func(t *testing.T) {
	//	reqBody := map[string]string{
	//		"email": "test@example.com",
	//	}
	//	jsonBody, _ := json.Marshal(reqBody)
	//
	//	req := httptest.NewRequest("POST", "/api/auth/resend-verification", bytes.NewBuffer(jsonBody))
	//	req.Header.Set("Content-Type", "application/json")
	//	w := httptest.NewRecorder()
	//
	//	auth.ResendVerificationHandler(w, req)
	//
	//	if w.Code != http.StatusOK {
	//		t.Errorf("Expected status 200, got %d", w.Code)
	//	}
	// })
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
