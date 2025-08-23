package email

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"secure-email-mvp/pkg/pqc"
)

// =============================================================================
// SES HANDLER UNIT TESTS - SIMPLIFIED VERSION
// =============================================================================
// Tests cover core functionality without complex mocking
// =============================================================================

// TestRetryLogic tests exponential backoff and jitter calculation
func TestRetryLogic(t *testing.T) {
	handler := &SESHandler{
		retryConfig: &SESRetryConfig{
			MaxRetries:        3,
			BaseDelay:         1 * time.Second,
			MaxDelay:          30 * time.Second,
			BackoffMultiplier: 2.0,
			JitterEnabled:     true,
		},
	}

	// Test delay calculation
	delays := make([]time.Duration, 4)
	for i := 0; i < 4; i++ {
		delays[i] = handler.calculateRetryDelay(i)
	}

	// Verify exponential backoff
	if delays[1] <= delays[0] {
		t.Error("Expected exponential backoff, delay should increase")
	}

	if delays[2] <= delays[1] {
		t.Error("Expected exponential backoff, delay should increase")
	}

	// Verify jitter is applied (delays should be different on multiple runs)
	delay1 := handler.calculateRetryDelay(1)
	delay2 := handler.calculateRetryDelay(1)

	// With jitter enabled, delays should be different
	if delay1 == delay2 {
		t.Error("Expected jitter to produce different delays")
	}
}

// TestRetryableErrorDetection tests retryable vs non-retryable error detection
func TestRetryableErrorDetection(t *testing.T) {
	handler := &SESHandler{}

	// Test retryable errors
	retryableErrors := []string{
		"connection refused",
		"connection timeout",
		"rate limit exceeded",
		"quota exceeded",
		"service unavailable",
		"temporary failure",
	}

	for _, errMsg := range retryableErrors {
		if !handler.isRetryableError(&mockError{msg: errMsg}) {
			t.Errorf("Expected error '%s' to be retryable", errMsg)
		}
	}

	// Test non-retryable errors
	nonRetryableErrors := []string{
		"authentication failed",
		"invalid recipient",
		"permission denied",
		"malformed request",
	}

	for _, errMsg := range nonRetryableErrors {
		if handler.isRetryableError(&mockError{msg: errMsg}) {
			t.Errorf("Expected error '%s' to be non-retryable", errMsg)
		}
	}
}

// TestQuotaHandling tests quota checking and updating
func TestQuotaHandling(t *testing.T) {
	handler := &SESHandler{
		config: &SESConfig{
			RateLimit: 14,
		},
		quotaInfo: &SESQuotaInfo{
			DailyQuota: 1000,
			SentToday:  500,
			LastCheck:  time.Now(),
		},
	}

	// Test quota check when under limit
	err := handler.checkQuota(context.Background())
	if err != nil {
		t.Errorf("Expected quota check to pass, got error: %v", err)
	}

	// Test quota check when over limit
	handler.quotaInfo.SentToday = 1000
	err = handler.checkQuota(context.Background())
	if err == nil {
		t.Error("Expected quota check to fail when over limit")
	}

	// Test quota update
	handler.quotaInfo.SentToday = 500
	handler.updateQuota(context.Background())
	if handler.quotaInfo.SentToday != 501 {
		t.Errorf("Expected sent count to be 501, got %d", handler.quotaInfo.SentToday)
	}
}

// TestLoadSESConfigFromEnv tests environment variable loading
func TestLoadSESConfigFromEnv(t *testing.T) {
	// Test with default values
	config := LoadSESConfigFromEnv()

	if config.SMTPHost != "email-smtp.us-east-1.amazonaws.com" {
		t.Errorf("Expected default SMTP host, got %s", config.SMTPHost)
	}

	if config.SMTPPort != 587 {
		t.Errorf("Expected default SMTP port 587, got %d", config.SMTPPort)
	}

	if config.Region != "us-east-1" {
		t.Errorf("Expected default region us-east-1, got %s", config.Region)
	}

	if config.RateLimit != 14 {
		t.Errorf("Expected default rate limit 14, got %d", config.RateLimit)
	}

	if !config.SandboxMode {
		t.Error("Expected default sandbox mode to be true")
	}
}

// TestEmailContentPreparation tests email content preparation
func TestEmailContentPreparation(t *testing.T) {
	handler := &SESHandler{
		config: &SESConfig{
			DefaultSender: "noreply@securesystem.email",
		},
	}

	recipient := "test@recipient.com"
	subject := "Test Subject"
	body := "Test email body"

	content := handler.prepareEmailContent(recipient, subject, body)

	// Verify content contains required fields
	if !containsString(content, "From: noreply@securesystem.email") {
		t.Error("Expected From header in email content")
	}

	if !containsString(content, "To: test@recipient.com") {
		t.Error("Expected To header in email content")
	}

	if !containsString(content, "Subject: Test Subject") {
		t.Error("Expected Subject header in email content")
	}

	if !containsString(content, "Test email body") {
		t.Error("Expected email body in content")
	}

	if !containsString(content, "Sent via Secure Email MVP with PQC + KT validation") {
		t.Error("Expected signature in email content")
	}
}

// TestPQCValidationStructure tests PQC data structure validation
func TestPQCValidationStructure(t *testing.T) {
	// Test valid PQC data structure
	validPQCData := &pqc.HybridEncryptedData{
		KyberCiphertext: []byte("valid_kyber_ciphertext"),
		KyberLevel:      768,
		AES256GCMData: &pqc.SymmetricEncryptedData{
			Ciphertext: []byte("valid_ciphertext"),
			Nonce:      []byte("valid_nonce"),
			AuthTag:    []byte("valid_auth_tag"),
			Algorithm:  "AES-256-GCM",
		},
		HybridMode: true,
		Version:    "1.0",
	}

	pqcDataBytes, err := json.Marshal(validPQCData)
	if err != nil {
		t.Fatalf("Failed to marshal valid PQC data: %v", err)
	}

	// Test that the data can be unmarshaled back
	var unmarshaledData pqc.HybridEncryptedData
	err = json.Unmarshal(pqcDataBytes, &unmarshaledData)
	if err != nil {
		t.Fatalf("Failed to unmarshal PQC data: %v", err)
	}

	// Verify the structure
	if len(unmarshaledData.KyberCiphertext) == 0 {
		t.Error("Expected Kyber ciphertext to be present")
	}

	if unmarshaledData.AES256GCMData == nil {
		t.Error("Expected AES-256-GCM data to be present")
	}

	if len(unmarshaledData.AES256GCMData.Nonce) == 0 {
		t.Error("Expected AES nonce to be present")
	}

	if len(unmarshaledData.AES256GCMData.AuthTag) == 0 {
		t.Error("Expected AES auth tag to be present")
	}

	if unmarshaledData.KyberLevel != 768 {
		t.Errorf("Expected Kyber level 768, got %d", unmarshaledData.KyberLevel)
	}
}

// TestDefaultRetryConfig tests default retry configuration
func TestDefaultRetryConfig(t *testing.T) {
	config := defaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", config.MaxRetries)
	}

	if config.BaseDelay != 1*time.Second {
		t.Errorf("Expected base delay 1s, got %v", config.BaseDelay)
	}

	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected max delay 30s, got %v", config.MaxDelay)
	}

	if config.BackoffMultiplier != 2.0 {
		t.Errorf("Expected backoff multiplier 2.0, got %f", config.BackoffMultiplier)
	}

	if !config.JitterEnabled {
		t.Error("Expected jitter to be enabled by default")
	}
}

// Helper functions

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests for performance validation

func BenchmarkRetryDelayCalculation(b *testing.B) {
	handler := &SESHandler{
		retryConfig: defaultRetryConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.calculateRetryDelay(i % 4)
	}
}

func BenchmarkEmailContentPreparation(b *testing.B) {
	handler := &SESHandler{
		config: &SESConfig{
			DefaultSender: "noreply@securesystem.email",
		},
	}

	recipient := "test@recipient.com"
	subject := "Test Subject"
	body := "Test email body"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.prepareEmailContent(recipient, subject, body)
	}
}

func BenchmarkPQCDataSerialization(b *testing.B) {
	validPQCData := &pqc.HybridEncryptedData{
		KyberCiphertext: []byte("valid_kyber_ciphertext"),
		KyberLevel:      768,
		AES256GCMData: &pqc.SymmetricEncryptedData{
			Ciphertext: []byte("valid_ciphertext"),
			Nonce:      []byte("valid_nonce"),
			AuthTag:    []byte("valid_auth_tag"),
			Algorithm:  "AES-256-GCM",
		},
		HybridMode: true,
		Version:    "1.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(validPQCData)
	}
}
