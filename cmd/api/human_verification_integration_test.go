package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/humanverification"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

// solveProofOfWorkChallenge solves a proof-of-work challenge for testing
func solveProofOfWorkChallenge(challenge *humanverification.Challenge) string {
	prefix := challenge.Prefix
	target := challenge.Target

	// Try nonces until we find one that produces a hash starting with the target
	for nonce := int64(0); nonce <= challenge.MaxNonce; nonce++ {
		data := fmt.Sprintf("%s%d", prefix, nonce)
		hash := sha256.Sum256([]byte(data))
		hashHex := hex.EncodeToString(hash[:])

		if strings.HasPrefix(hashHex, target) {
			return fmt.Sprintf("%s:%d", challenge.ID, nonce)
		}
	}

	// If no solution found, return a dummy solution (this shouldn't happen in tests)
	return fmt.Sprintf("%s:0", challenge.ID)
}

func TestHumanVerificationIntegration(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply human verification migration
	migration := `
		CREATE TABLE IF NOT EXISTS emails (
			email_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS human_verification_logs (
			id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			verification_type TEXT NOT NULL,
			challenge_id TEXT,
			result TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			details TEXT
		);
		
		CREATE INDEX IF NOT EXISTS idx_human_verification_logs_email_ip ON human_verification_logs(email_id, ip_address);
		CREATE INDEX IF NOT EXISTS idx_human_verification_logs_timestamp ON human_verification_logs(timestamp);
	`

	if _, err := db.Exec(migration); err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Create test email
	emailID := "test-email-123"
	userID := "test-user-456"
	_, err = db.Exec("INSERT INTO emails (email_id, user_id) VALUES (?, ?)", emailID, userID)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Initialize human verification service with mock config
	config := &humanverification.Config{
		Enabled:               true,
		VerificationType:      "proof_of_work",
		ProofOfWorkDifficulty: 2, // Lower difficulty for testing
		MaxNonce:              1000,
		FailureThreshold:      3,
		BanDuration:           5 * time.Minute,
	}

	verificationSvc := humanverification.NewHumanVerificationService(db, config)

	t.Run("Generate and verify proof-of-work challenge", func(t *testing.T) {
		ctx := context.Background()

		// Generate challenge
		challenge, err := verificationSvc.GenerateChallenge(ctx)
		if err != nil {
			t.Fatalf("Failed to generate challenge: %v", err)
		}

		if challenge.ID == "" {
			t.Error("Challenge ID should not be empty")
		}

		if challenge.Prefix == "" {
			t.Error("Challenge prefix should not be empty")
		}

		if challenge.Target == "" {
			t.Error("Challenge target should not be empty")
		}

		if challenge.MaxNonce <= 0 {
			t.Error("Challenge max nonce should be positive")
		}

		// For testing, we'll solve the proof-of-work challenge
		// In a real scenario, the client would solve this
		solution := solveProofOfWorkChallenge(challenge)

		// Verify the solution
		success, err := verificationSvc.VerifyResponse(ctx, emailID, solution, humanverification.VerificationTypeProofOfWork)
		if err != nil {
			t.Fatalf("Failed to verify solution: %v", err)
		}

		if !success {
			t.Error("Verification should succeed with valid solution")
		}
	})

	t.Run("Log verification attempts", func(t *testing.T) {
		ctx := context.Background()

		// Log a successful verification
		logEntry := &humanverification.VerificationLog{
			EmailID:          emailID,
			IPAddress:        "192.168.1.1",
			UserAgent:        "Test-Agent/1.0",
			VerificationType: humanverification.VerificationTypeProofOfWork,
			ChallengeID:      "test-challenge",
			Result:           "success",
		}

		err := verificationSvc.LogVerification(ctx, logEntry)
		if err != nil {
			t.Fatalf("Failed to log verification: %v", err)
		}

		// Log a failed verification
		logEntry2 := &humanverification.VerificationLog{
			EmailID:          emailID,
			IPAddress:        "192.168.1.1",
			UserAgent:        "Test-Agent/1.0",
			VerificationType: humanverification.VerificationTypeProofOfWork,
			ChallengeID:      "test-challenge-2",
			Result:           "failure",
		}

		err = verificationSvc.LogVerification(ctx, logEntry2)
		if err != nil {
			t.Fatalf("Failed to log verification: %v", err)
		}
	})

	t.Run("Get verification statistics", func(t *testing.T) {
		ctx := context.Background()

		stats, err := verificationSvc.GetVerificationStats(ctx, emailID, "192.168.1.1", time.Hour)
		if err != nil {
			t.Fatalf("Failed to get verification stats: %v", err)
		}

		if stats.TotalAttempts < 2 {
			t.Errorf("Expected at least 2 total attempts, got %d", stats.TotalAttempts)
		}

		if stats.SuccessAttempts < 1 {
			t.Errorf("Expected at least 1 success attempt, got %d", stats.SuccessAttempts)
		}

		if stats.FailedAttempts < 1 {
			t.Errorf("Expected at least 1 failed attempt, got %d", stats.FailedAttempts)
		}

		if stats.FailureRate <= 0 {
			t.Error("Expected positive failure rate")
		}
	})
}

func TestHumanVerificationWithServer(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply migrations
	migration := `
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			totp_secret TEXT NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS emails (
			email_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS human_verification_logs (
			id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			verification_type TEXT NOT NULL,
			challenge_id TEXT,
			result TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			details TEXT
		);
		
		CREATE INDEX IF NOT EXISTS idx_human_verification_logs_email_ip ON human_verification_logs(email_id, ip_address);
		CREATE INDEX IF NOT EXISTS idx_human_verification_logs_timestamp ON human_verification_logs(timestamp);
	`

	if _, err := db.Exec(migration); err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Create test user and email
	userID := "test-user-123"
	emailID := "test-email-456"

	_, err = db.Exec("INSERT INTO users (user_id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		userID, "test@example.com", "hashed_password", "totp_secret")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, user_id) VALUES (?, ?)", emailID, userID)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Initialize human verification service
	config := &humanverification.Config{
		Enabled:               true,
		VerificationType:      "proof_of_work",
		ProofOfWorkDifficulty: 2,
		MaxNonce:              1000,
		FailureThreshold:      3,
		BanDuration:           5 * time.Minute,
	}

	verificationSvc := humanverification.NewHumanVerificationService(db, config)

	// Generate JWT token for testing
	jwtToken, err := auth.GenerateJWT(userID)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	t.Run("Generate challenge endpoint", func(t *testing.T) {
		// Create request
		req, err := http.NewRequest("GET", "/api/human-verification/challenge", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Create response recorder
		rr := httptest.NewRecorder()

		// Create middleware and call handler
		middleware := NewHumanVerificationMiddleware(verificationSvc, config)
		middleware.GenerateChallengeHandler(rr, req)

		// Check response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		// Parse response
		var response map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify response fields
		if response["type"] != "proof_of_work" {
			t.Error("Response should contain type = proof_of_work")
		}

		if response["challenge"] == nil {
			t.Error("Response should contain challenge")
		}
	})

	t.Run("Trust device endpoint with human verification", func(t *testing.T) {
		// Create request without verification token (should return challenge)
		req, err := http.NewRequest("POST", "/api/email/"+emailID+"/trust-device", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("User-Agent", "Test-Agent/1.0")
		req.RemoteAddr = "192.168.1.100:12345"

		// Create response recorder
		rr := httptest.NewRecorder()

		// Create middleware and call handler
		middleware := NewHumanVerificationMiddleware(verificationSvc, config)

		// Create a router to properly set up URL variables
		router := mux.NewRouter()
		router.Handle("/api/email/{id}/trust-device",
			middleware.RequireHumanVerification(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})))

		router.ServeHTTP(rr, req)

		// Check response (should be 401 with challenge)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}

		// Parse response
		var response map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify challenge response
		if response["verification_required"] != true {
			t.Error("Response should contain verification_required = true")
		}

		if response["verification_type"] != "proof_of_work" {
			t.Error("Response should contain verification_type = proof_of_work")
		}

		if response["challenge"] == nil {
			t.Error("Response should contain challenge")
		}
	})

	t.Run("Trust device endpoint with valid verification", func(t *testing.T) {
		ctx := context.Background()

		// Generate a challenge first
		challenge, err := verificationSvc.GenerateChallenge(ctx)
		if err != nil {
			t.Fatalf("Failed to generate challenge: %v", err)
		}

		// Solve the proof-of-work challenge
		solution := solveProofOfWorkChallenge(challenge)

		// Create request with verification token
		req, err := http.NewRequest("POST", "/api/email/"+emailID+"/trust-device", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("User-Agent", "Test-Agent/1.0")
		req.RemoteAddr = "192.168.1.100:12345"
		req.Header.Set("X-Human-Verification-Token", solution)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Create middleware and call handler
		middleware := NewHumanVerificationMiddleware(verificationSvc, config)

		// Create a router to properly set up URL variables
		router := mux.NewRouter()
		router.Handle("/api/email/{id}/trust-device",
			middleware.RequireHumanVerification(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})))

		router.ServeHTTP(rr, req)

		// Check response (should be 200)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	})
}

func TestHumanVerificationDisabled(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Initialize human verification service with disabled config
	config := &humanverification.Config{
		Enabled: false, // Disabled
	}

	verificationSvc := humanverification.NewHumanVerificationService(db, config)

	t.Run("Verification should be bypassed when disabled", func(t *testing.T) {
		ctx := context.Background()

		// Try to verify without token (should succeed when disabled)
		success, err := verificationSvc.VerifyResponse(ctx, "test-email", "", humanverification.VerificationTypeProofOfWork)
		if err != nil {
			t.Fatalf("Failed to verify: %v", err)
		}

		if !success {
			t.Error("Verification should succeed when disabled")
		}
	})

	t.Run("Challenge endpoint should return error when disabled", func(t *testing.T) {
		// Create request
		req, err := http.NewRequest("GET", "/api/human-verification/challenge", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Create response recorder
		rr := httptest.NewRecorder()

		// Create middleware and call handler
		middleware := NewHumanVerificationMiddleware(verificationSvc, config)
		middleware.GenerateChallengeHandler(rr, req)

		// Check response (should be 503)
		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 503, got %d", rr.Code)
		}
	})
}
