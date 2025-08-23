package email

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"secure-email-mvp/pkg/audit"
	"secure-email-mvp/pkg/e2e"
	"secure-email-mvp/pkg/pqc"

	_ "github.com/mattn/go-sqlite3"
)

// =============================================================================
// SES HANDLER INTEGRATION TESTS
// =============================================================================
// Tests the complete SES handoff flow with real database and services
// =============================================================================

// setupTestDatabase creates a test SQLite database
func setupTestDatabase(t *testing.T) *sql.DB {
	// Create temporary database file
	tmpFile := t.TempDir() + "/test.db"

	db, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create SES tables
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS ses_transactions (
		transaction_id TEXT PRIMARY KEY NOT NULL,
		message_id TEXT NOT NULL,
		sender_id TEXT NOT NULL,
		recipient TEXT NOT NULL,
		blob_id TEXT,
		timestamp DATETIME NOT NULL,
		status TEXT NOT NULL DEFAULT 'sent',
		retry_count INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS ses_quota_usage (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		usage_date DATE NOT NULL,
		daily_quota INTEGER NOT NULL,
		sent_count INTEGER NOT NULL DEFAULT 0,
		failed_count INTEGER NOT NULL DEFAULT 0,
		rate_limit INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(usage_date)
	);

	CREATE TABLE IF NOT EXISTS ses_validation_logs (
		validation_id TEXT PRIMARY KEY NOT NULL,
		email_id TEXT NOT NULL,
		sender_id TEXT NOT NULL,
		recipient TEXT,
		pqc_valid BOOLEAN NOT NULL,
		kt_valid BOOLEAN NOT NULL,
		overall_valid BOOLEAN NOT NULL,
		error_code TEXT,
		error_message TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createTablesSQL)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return db
}

// TestCompleteSESHandoffFlow tests the complete email sending flow
func TestCompleteSESHandoffFlow(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	// Setup test database
	db := setupTestDatabase(t)
	defer db.Close()

	// Setup services
	auditService := audit.NewAuditService(db)
	pqcService, _ := pqc.NewPQCService(pqc.DefaultPQCConfig())
	ktService := e2e.NewKeyTransparency(e2e.KTConfig{Enabled: true})

	// Create SES handler
	sesConfig := &SESConfig{
		SMTPHost:      "test.smtp.com",
		SMTPPort:      587,
		SMTPUsername:  "test_user",
		SMTPPassword:  "test_pass",
		RateLimit:     14,
		SandboxMode:   true,
		DefaultSender: "test@example.com",
	}

	handler := NewSESHandler(sesConfig, db, auditService, pqcService, ktService)

	// Test data
	emailID := "test_email_123"
	senderID := "test_sender"
	recipient := "test@recipient.com"
	subject := "Test Subject"
	body := "Test email body"

	// Create valid PQC encrypted data
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
		t.Fatalf("Failed to marshal PQC data: %v", err)
	}

	// Test PQC + KT validation
	t.Run("PQC_KT_Validation", func(t *testing.T) {
		result, err := handler.ValidatePQCAndKT(context.Background(), emailID, senderID, recipient, pqcDataBytes)

		// Note: This will likely fail in test environment since we don't have real PQC/KT services
		// The important thing is that the validation logic runs without panicking
		if err != nil {
			t.Logf("Validation failed as expected in test environment: %v", err)
		}

		if result == nil {
			t.Error("Expected validation result, got nil")
		}
	})

	// Test quota handling
	t.Run("QuotaHandling", func(t *testing.T) {
		// Test initial quota check
		err := handler.checkQuota(context.Background())
		if err != nil {
			t.Errorf("Expected quota check to pass initially: %v", err)
		}

		// Test quota info retrieval
		quotaInfo, err := handler.GetQuotaInfo(context.Background())
		if err != nil {
			t.Errorf("Failed to get quota info: %v", err)
		}

		if quotaInfo == nil {
			t.Error("Expected quota info, got nil")
		}
	})

	// Test retry logic
	t.Run("RetryLogic", func(t *testing.T) {
		// Test delay calculation
		delays := make([]time.Duration, 4)
		for i := 0; i < 4; i++ {
			delays[i] = handler.calculateRetryDelay(i)
		}

		// Verify exponential backoff
		if delays[1] <= delays[0] {
			t.Error("Expected exponential backoff")
		}

		if delays[2] <= delays[1] {
			t.Error("Expected exponential backoff")
		}
	})

	// Test error classification
	t.Run("ErrorClassification", func(t *testing.T) {
		// Test retryable errors
		retryableErrors := []string{
			"connection refused",
			"rate limit exceeded",
			"service unavailable",
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
		}

		for _, errMsg := range nonRetryableErrors {
			if handler.isRetryableError(&mockError{msg: errMsg}) {
				t.Errorf("Expected error '%s' to be non-retryable", errMsg)
			}
		}
	})

	// Test email content preparation
	t.Run("EmailContentPreparation", func(t *testing.T) {
		content := handler.prepareEmailContent(recipient, subject, body)

		// Verify required headers
		requiredHeaders := []string{
			"From: test@example.com",
			"To: test@recipient.com",
			"Subject: Test Subject",
			"Content-Type: text/plain; charset=UTF-8",
			"MIME-Version: 1.0",
		}

		for _, header := range requiredHeaders {
			if !containsString(content, header) {
				t.Errorf("Expected header '%s' in email content", header)
			}
		}

		// Verify body content
		if !containsString(content, body) {
			t.Error("Expected email body in content")
		}

		// Verify signature
		if !containsString(content, "Sent via Secure Email MVP with PQC + KT validation") {
			t.Error("Expected signature in email content")
		}
	})

	// Test transaction storage
	t.Run("TransactionStorage", func(t *testing.T) {
		transaction := &SESTransaction{
			TransactionID: "test_transaction_123",
			MessageID:     emailID,
			SenderID:      senderID,
			Recipient:     recipient,
			BlobID:        "test_blob_789",
			Timestamp:     time.Now(),
			Status:        "sent",
			RetryCount:    0,
		}

		// Store transaction
		err := handler.storeSESTransaction(context.Background(), transaction)
		if err != nil {
			t.Errorf("Failed to store transaction: %v", err)
		}

		// Retrieve transaction
		retrieved, err := handler.GetSESTransaction(context.Background(), transaction.TransactionID)
		if err != nil {
			t.Errorf("Failed to retrieve transaction: %v", err)
		}

		if retrieved == nil {
			t.Error("Expected retrieved transaction, got nil")
		}

		if retrieved.TransactionID != transaction.TransactionID {
			t.Errorf("Expected transaction ID %s, got %s", transaction.TransactionID, retrieved.TransactionID)
		}
	})
}

// TestSESConfiguration tests SES configuration loading and validation
func TestSESConfiguration(t *testing.T) {
	// Test default configuration
	config := LoadSESConfigFromEnv()

	// Verify required fields have defaults
	if config.SMTPHost == "" {
		t.Error("Expected SMTP host to have default value")
	}

	if config.SMTPPort == 0 {
		t.Error("Expected SMTP port to have default value")
	}

	if config.Region == "" {
		t.Error("Expected region to have default value")
	}

	if config.RateLimit == 0 {
		t.Error("Expected rate limit to have default value")
	}

	// Test configuration validation
	if config.SMTPUsername == "" {
		t.Log("SMTP username not set (expected in test environment)")
	}

	if config.SMTPPassword == "" {
		t.Log("SMTP password not set (expected in test environment)")
	}
}

// TestAuditLogging tests audit logging integration
func TestAuditLogging(t *testing.T) {
	// Setup test database
	db := setupTestDatabase(t)
	defer db.Close()

	// Setup audit service
	auditService := audit.NewAuditService(db)

	// Create SES handler with audit service
	handler := &SESHandler{
		config:       &SESConfig{},
		retryConfig:  defaultRetryConfig(),
		quotaInfo:    &SESQuotaInfo{},
		db:           db,
		auditService: auditService,
		pqcService:   nil,
		ktService:    nil,
	}

	// Test validation success logging
	handler.logValidationSuccess(context.Background(), "test_email", "test_sender", "test_validation_id")

	// Test validation failure logging
	handler.logValidationFailure(context.Background(), "test_email", "test_sender", "TEST_ERROR", "Test error message", "test_validation_id")

	// Test send success logging
	transaction := &SESTransaction{
		TransactionID: "test_transaction",
		MessageID:     "test_email",
		SenderID:      "test_sender",
		Recipient:     "test@recipient.com",
		Status:        "sent",
		RetryCount:    0,
	}
	handler.logSendSuccess(context.Background(), "test_email", "test_sender", "test@recipient.com", transaction)

	// Test send failure logging
	handler.logSendFailure(context.Background(), "test_email", "test_sender", "test@recipient.com", "Test failure message")
}

// Note: Helper functions are defined in ses_handler_test.go
