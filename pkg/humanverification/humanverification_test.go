package humanverification

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMockHumanVerificationService_VerifyResponse(t *testing.T) {
	mockSvc := NewMockHumanVerificationService()
	ctx := context.Background()
	
	t.Run("Success by default", func(t *testing.T) {
		success, err := mockSvc.VerifyResponse(ctx, "test-email", "test-token", VerificationTypeProofOfWork)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !success {
			t.Error("Expected verification to succeed by default")
		}
	})
	
	t.Run("Failure when configured", func(t *testing.T) {
		mockSvc.SetVerificationResult(true) // shouldFail = true
		success, err := mockSvc.VerifyResponse(ctx, "test-email", "test-token", VerificationTypeProofOfWork)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if success {
			t.Error("Expected verification to fail when configured")
		}
	})
	
	t.Run("Specific token result", func(t *testing.T) {
		mockSvc.SetVerificationResult(false) // Global success
		mockSvc.SetSpecificVerification("specific-token", false) // This token should fail
		
		// Test specific token
		success, err := mockSvc.VerifyResponse(ctx, "test-email", "specific-token", VerificationTypeProofOfWork)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if success {
			t.Error("Expected specific token to fail")
		}
		
		// Test other token (should succeed)
		success, err = mockSvc.VerifyResponse(ctx, "test-email", "other-token", VerificationTypeProofOfWork)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !success {
			t.Error("Expected other token to succeed")
		}
	})
}

func TestMockHumanVerificationService_GenerateChallenge(t *testing.T) {
	mockSvc := NewMockHumanVerificationService()
	ctx := context.Background()
	
	challenge, err := mockSvc.GenerateChallenge(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if challenge.ID == "" {
		t.Error("Expected challenge ID to be set")
	}
	
	if challenge.Prefix == "" {
		t.Error("Expected challenge prefix to be set")
	}
	
	if challenge.Target == "" {
		t.Error("Expected challenge target to be set")
	}
	
	if challenge.MaxNonce <= 0 {
		t.Error("Expected challenge max nonce to be positive")
	}
	
	// Verify the challenge was stored
	challenges := mockSvc.challenges
	if _, exists := challenges[challenge.ID]; !exists {
		t.Error("Expected challenge to be stored")
	}
}

func TestMockHumanVerificationService_LogVerification(t *testing.T) {
	mockSvc := NewMockHumanVerificationService()
	ctx := context.Background()
	
	logEntry := &VerificationLog{
		EmailID:         "test-email",
		IPAddress:       "192.168.1.1",
		UserAgent:       "Test-Agent/1.0",
		VerificationType: VerificationTypeProofOfWork,
		ChallengeID:     "test-challenge",
		Result:          "success",
	}
	
	err := mockSvc.LogVerification(ctx, logEntry)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	logs := mockSvc.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}
	
	logged := logs[0]
	if logged.EmailID != logEntry.EmailID {
		t.Errorf("Expected email ID %s, got %s", logEntry.EmailID, logged.EmailID)
	}
	
	if logged.IPAddress != logEntry.IPAddress {
		t.Errorf("Expected IP address %s, got %s", logEntry.IPAddress, logged.IPAddress)
	}
	
	if logged.Result != logEntry.Result {
		t.Errorf("Expected result %s, got %s", logEntry.Result, logged.Result)
	}
	
	if logged.ID == "" {
		t.Error("Expected log entry ID to be auto-generated")
	}
	
	if logged.Timestamp.IsZero() {
		t.Error("Expected log entry timestamp to be auto-generated")
	}
}

func TestMockHumanVerificationService_GetVerificationStats(t *testing.T) {
	mockSvc := NewMockHumanVerificationService()
	ctx := context.Background()
	
	// Add some test logs
	logs := []*VerificationLog{
		{
			EmailID:         "test-email",
			IPAddress:       "192.168.1.1",
			Result:          "success",
			Timestamp:       time.Now(),
		},
		{
			EmailID:         "test-email",
			IPAddress:       "192.168.1.1",
			Result:          "failure",
			Timestamp:       time.Now(),
		},
		{
			EmailID:         "other-email",
			IPAddress:       "192.168.1.2",
			Result:          "success",
			Timestamp:       time.Now(),
		},
		{
			EmailID:         "test-email",
			IPAddress:       "192.168.1.1",
			Result:          "success",
			Timestamp:       time.Now().Add(-2 * time.Hour), // Old log
		},
	}
	
	for _, log := range logs {
		mockSvc.LogVerification(ctx, log)
	}
	
	// Test stats for test-email
	stats, err := mockSvc.GetVerificationStats(ctx, "test-email", "192.168.1.1", time.Hour)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if stats.TotalAttempts != 2 { // Only recent logs within 1 hour
		t.Errorf("Expected 2 total attempts, got %d", stats.TotalAttempts)
	}
	
	if stats.SuccessAttempts != 1 {
		t.Errorf("Expected 1 success attempt, got %d", stats.SuccessAttempts)
	}
	
	if stats.FailedAttempts != 1 {
		t.Errorf("Expected 1 failed attempt, got %d", stats.FailedAttempts)
	}
	
	if stats.FailureRate != 0.5 {
		t.Errorf("Expected 0.5 failure rate, got %f", stats.FailureRate)
	}
}

func TestMockHumanVerificationService_ClearLogs(t *testing.T) {
	mockSvc := NewMockHumanVerificationService()
	ctx := context.Background()
	
	// Add a log entry
	logEntry := &VerificationLog{
		EmailID: "test-email",
		Result:  "success",
	}
	mockSvc.LogVerification(ctx, logEntry)
	
	// Verify log was added
	if len(mockSvc.GetLogs()) != 1 {
		t.Error("Expected 1 log entry before clearing")
	}
	
	// Clear logs
	mockSvc.ClearLogs()
	
	// Verify logs were cleared
	if len(mockSvc.GetLogs()) != 0 {
		t.Error("Expected 0 log entries after clearing")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if !config.Enabled {
		t.Error("Expected human verification to be enabled by default")
	}
	
	if config.VerificationType != "proof_of_work" {
		t.Errorf("Expected default verification type 'proof_of_work', got %s", config.VerificationType)
	}
	
	if config.ProofOfWorkDifficulty != 4 {
		t.Errorf("Expected default difficulty 4, got %d", config.ProofOfWorkDifficulty)
	}
	
	if config.MaxNonce != 1000000 {
		t.Errorf("Expected default max nonce 1000000, got %d", config.MaxNonce)
	}
	
	if config.FailureThreshold != 5 {
		t.Errorf("Expected default failure threshold 5, got %d", config.FailureThreshold)
	}
	
	if config.BanDuration != 15*time.Minute {
		t.Errorf("Expected default ban duration 15 minutes, got %v", config.BanDuration)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Test with environment variables
	originalEnv := make(map[string]string)
	for _, key := range []string{
		"HUMAN_VERIFICATION_ENABLED",
		"HUMAN_VERIFICATION_TYPE",
		"CAPTCHA_SECRET_KEY",
		"PROOF_OF_WORK_DIFFICULTY",
		"HUMAN_VERIFICATION_FAILURE_THRESHOLD",
	} {
		originalEnv[key] = os.Getenv(key)
	}
	
	// Set test environment variables
	os.Setenv("HUMAN_VERIFICATION_ENABLED", "false")
	os.Setenv("HUMAN_VERIFICATION_TYPE", "captcha")
	os.Setenv("CAPTCHA_SECRET_KEY", "test-secret")
	os.Setenv("PROOF_OF_WORK_DIFFICULTY", "6")
	os.Setenv("HUMAN_VERIFICATION_FAILURE_THRESHOLD", "10")
	
	// Restore environment variables after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()
	
	config := LoadConfigFromEnv()
	
	if config.Enabled {
		t.Error("Expected human verification to be disabled from env")
	}
	
	if config.VerificationType != "captcha" {
		t.Errorf("Expected verification type 'captcha', got %s", config.VerificationType)
	}
	
	if config.CAPTCHASecretKey != "test-secret" {
		t.Errorf("Expected CAPTCHA secret key 'test-secret', got %s", config.CAPTCHASecretKey)
	}
	
	if config.ProofOfWorkDifficulty != 6 {
		t.Errorf("Expected difficulty 6, got %d", config.ProofOfWorkDifficulty)
	}
	
	if config.FailureThreshold != 10 {
		t.Errorf("Expected failure threshold 10, got %d", config.FailureThreshold)
	}
}

func TestVerificationTypeConstants(t *testing.T) {
	if VerificationTypeCAPTCHA != "captcha" {
		t.Errorf("Expected VerificationTypeCAPTCHA to be 'captcha', got %s", VerificationTypeCAPTCHA)
	}
	
	if VerificationTypeProofOfWork != "proof_of_work" {
		t.Errorf("Expected VerificationTypeProofOfWork to be 'proof_of_work', got %s", VerificationTypeProofOfWork)
	}
}

func TestChallengeStructure(t *testing.T) {
	challenge := &Challenge{
		ID:       "test-id",
		Prefix:   "test-prefix",
		Target:   "0000",
		MaxNonce: 1000000,
	}
	
	if challenge.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", challenge.ID)
	}
	
	if challenge.Prefix != "test-prefix" {
		t.Errorf("Expected prefix 'test-prefix', got %s", challenge.Prefix)
	}
	
	if challenge.Target != "0000" {
		t.Errorf("Expected target '0000', got %s", challenge.Target)
	}
	
	if challenge.MaxNonce != 1000000 {
		t.Errorf("Expected max nonce 1000000, got %d", challenge.MaxNonce)
	}
}

func TestVerificationLogStructure(t *testing.T) {
	now := time.Now()
	log := &VerificationLog{
		ID:              "test-id",
		EmailID:         "test-email",
		IPAddress:       "192.168.1.1",
		UserAgent:       "Test-Agent/1.0",
		VerificationType: VerificationTypeProofOfWork,
		ChallengeID:     "test-challenge",
		Result:          "success",
		Timestamp:       now,
		Details:         "test details",
	}
	
	if log.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", log.ID)
	}
	
	if log.EmailID != "test-email" {
		t.Errorf("Expected email ID 'test-email', got %s", log.EmailID)
	}
	
	if log.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP address '192.168.1.1', got %s", log.IPAddress)
	}
	
	if log.Result != "success" {
		t.Errorf("Expected result 'success', got %s", log.Result)
	}
	
	if !log.Timestamp.Equal(now) {
		t.Errorf("Expected timestamp %v, got %v", now, log.Timestamp)
	}
}
