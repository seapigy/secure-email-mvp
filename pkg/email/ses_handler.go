package email

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"secure-email-mvp/pkg/audit"
	"secure-email-mvp/pkg/e2e"
	"secure-email-mvp/pkg/pqc"

	"github.com/google/uuid"
)

// =============================================================================
// SECURE EMAIL MVP - SES HANDLER WITH PQC + KT ENFORCEMENT
// =============================================================================
// This package implements security and reliability checks before mail handoff
// to Amazon SES, including:
// - Post-Quantum Cryptography (PQC) encryption layer validation
// - Key Transparency (KT) verification checks
// - SES handoff with transaction ID capture
// - Retry logic with exponential backoff and quota handling
// - Comprehensive audit logging for compliance
// =============================================================================

// SESConfig holds SES configuration
type SESConfig struct {
	SMTPHost      string `json:"smtp_host"`
	SMTPPort      int    `json:"smtp_port"`
	SMTPUsername  string `json:"smtp_username"`
	SMTPPassword  string `json:"smtp_password"`
	Region        string `json:"region"`
	RateLimit     int    `json:"rate_limit"`
	SandboxMode   bool   `json:"sandbox_mode"`
	DefaultSender string `json:"default_sender"`
}

// SESValidationResult represents the result of PQC + KT validation
type SESValidationResult struct {
	Valid        bool   `json:"valid"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	PQCValid     bool   `json:"pqc_valid"`
	KTValid      bool   `json:"kt_valid"`
	ValidationID string `json:"validation_id"`
}

// SESTransaction represents a successful SES transaction
type SESTransaction struct {
	TransactionID string    `json:"transaction_id"`
	MessageID     string    `json:"message_id"`
	SenderID      string    `json:"sender_id"`
	Recipient     string    `json:"recipient"`
	BlobID        string    `json:"blob_id"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
	RetryCount    int       `json:"retry_count"`
}

// SESRetryConfig holds retry configuration
type SESRetryConfig struct {
	MaxRetries        int           `json:"max_retries"`
	BaseDelay         time.Duration `json:"base_delay"`
	MaxDelay          time.Duration `json:"max_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	JitterEnabled     bool          `json:"jitter_enabled"`
}

// SESQuotaInfo holds SES quota information
type SESQuotaInfo struct {
	DailyQuota     int       `json:"daily_quota"`
	SentToday      int       `json:"sent_today"`
	RateLimit      int       `json:"rate_limit"`
	QuotaResetTime time.Time `json:"quota_reset_time"`
	LastCheck      time.Time `json:"last_check"`
}

// SESHandler provides SES integration with security enforcement
type SESHandler struct {
	config       *SESConfig
	retryConfig  *SESRetryConfig
	quotaInfo    *SESQuotaInfo
	db           *sql.DB
	auditService *audit.AuditService
	pqcService   *pqc.PQCService
	ktService    *e2e.KeyTransparency
	mu           sync.RWMutex
}

// NewSESHandler creates a new SES handler
func NewSESHandler(config *SESConfig, db *sql.DB, auditService *audit.AuditService, pqcService *pqc.PQCService, ktService *e2e.KeyTransparency) *SESHandler {
	return &SESHandler{
		config:       config,
		retryConfig:  defaultRetryConfig(),
		quotaInfo:    &SESQuotaInfo{},
		db:           db,
		auditService: auditService,
		pqcService:   pqcService,
		ktService:    ktService,
	}
}

// defaultRetryConfig returns default retry configuration
func defaultRetryConfig() *SESRetryConfig {
	return &SESRetryConfig{
		MaxRetries:        3,
		BaseDelay:         1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterEnabled:     true,
	}
}

// LoadSESConfigFromEnv loads SES configuration from environment variables
func LoadSESConfigFromEnv() *SESConfig {
	port, _ := strconv.Atoi(getEnvOrDefault("SES_SMTP_PORT", "587"))
	rateLimit, _ := strconv.Atoi(getEnvOrDefault("SES_SMTP_RATE_LIMIT", "14"))
	sandboxMode := getEnvOrDefault("SES_SMTP_SANDBOX_MODE", "true") == "true"

	return &SESConfig{
		SMTPHost:      getEnvOrDefault("SES_SMTP_HOST", "email-smtp.us-east-1.amazonaws.com"),
		SMTPPort:      port,
		SMTPUsername:  os.Getenv("SES_SMTP_USERNAME"),
		SMTPPassword:  os.Getenv("SES_SMTP_PASSWORD"),
		Region:        getEnvOrDefault("SES_SMTP_REGION", "us-east-1"),
		RateLimit:     rateLimit,
		SandboxMode:   sandboxMode,
		DefaultSender: getEnvOrDefault("SES_DEFAULT_SENDER", "noreply@securesystem.email"),
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ValidatePQCAndKT validates PQC encryption and Key Transparency before SES handoff
func (h *SESHandler) ValidatePQCAndKT(ctx context.Context, emailID, senderID, recipient string, encryptedData []byte) (*SESValidationResult, error) {
	validationID := uuid.New().String()
	result := &SESValidationResult{
		Valid:        false,
		ValidationID: validationID,
	}

	// Step 1: Validate PQC encryption
	pqcValid, pqcErr := h.validatePQCEncryption(ctx, emailID, encryptedData)
	result.PQCValid = pqcValid
	if pqcErr != nil {
		result.ErrorCode = "PQC_VALIDATION_FAILED"
		result.ErrorMessage = fmt.Sprintf("PQC validation failed: %v", pqcErr)
		h.logValidationFailure(ctx, emailID, senderID, "PQC_VALIDATION_FAILED", pqcErr.Error(), validationID)
		return result, pqcErr
	}

	// Step 2: Validate Key Transparency
	ktValid, ktErr := h.validateKeyTransparency(ctx, emailID, senderID, recipient)
	result.KTValid = ktValid
	if ktErr != nil {
		result.ErrorCode = "KT_VALIDATION_FAILED"
		result.ErrorMessage = fmt.Sprintf("Key Transparency validation failed: %v", ktErr)
		h.logValidationFailure(ctx, emailID, senderID, "KT_VALIDATION_FAILED", ktErr.Error(), validationID)
		return result, ktErr
	}

	// Step 3: Both validations passed
	result.Valid = true
	h.logValidationSuccess(ctx, emailID, senderID, validationID)

	return result, nil
}

// validatePQCEncryption validates that the email content is properly encrypted with PQC
func (h *SESHandler) validatePQCEncryption(ctx context.Context, emailID string, encryptedData []byte) (bool, error) {
	if h.pqcService == nil {
		return false, fmt.Errorf("PQC service not available")
	}

	// Parse the encrypted data to validate PQC structure
	var hybridData pqc.HybridEncryptedData
	if err := json.Unmarshal(encryptedData, &hybridData); err != nil {
		return false, fmt.Errorf("failed to parse PQC encrypted data: %w", err)
	}

	// Validate PQC components
	if len(hybridData.KyberCiphertext) == 0 {
		return false, fmt.Errorf("missing Kyber ciphertext")
	}

	if hybridData.AES256GCMData == nil {
		return false, fmt.Errorf("missing AES-256-GCM data")
	}

	if len(hybridData.AES256GCMData.Nonce) == 0 {
		return false, fmt.Errorf("missing AES nonce")
	}

	if len(hybridData.AES256GCMData.AuthTag) == 0 {
		return false, fmt.Errorf("missing AES auth tag")
	}

	// Validate Kyber security level
	if hybridData.KyberLevel != 512 && hybridData.KyberLevel != 768 && hybridData.KyberLevel != 1024 {
		return false, fmt.Errorf("invalid Kyber security level: %d", hybridData.KyberLevel)
	}

	// Log PQC validation success
	log.Printf("✅ PQC validation successful for email %s: Kyber-%d, hybrid mode: %t",
		emailID, hybridData.KyberLevel, hybridData.HybridMode)

	return true, nil
}

// validateKeyTransparency validates Key Transparency for the email
func (h *SESHandler) validateKeyTransparency(ctx context.Context, emailID, senderID, recipient string) (bool, error) {
	if h.ktService == nil {
		return false, fmt.Errorf("Key Transparency service not available")
	}

	// Verify sender's public key in KT
	auditResult, err := h.ktService.VerifyPublicKey(senderID, "sender_public_key", "pqc")
	if err != nil {
		return false, fmt.Errorf("failed to verify sender public key in KT: %w", err)
	}

	if !auditResult.Valid {
		return false, fmt.Errorf("sender public key verification failed: %s", auditResult.ErrorMsg)
	}

	// Try to verify recipient's public key in KT (if they're a registered user)
	recipientAuditResult, err := h.ktService.VerifyPublicKey(recipient, "recipient_public_key", "pqc")
	if err != nil {
		// Recipient might not be registered, which is acceptable for external emails
		log.Printf("ℹ️ Recipient %s not found in KT (external email): %v", recipient, err)
	} else if !recipientAuditResult.Valid {
		log.Printf("⚠️ Recipient public key verification failed: %s", recipientAuditResult.ErrorMsg)
		// Don't fail for recipient verification issues with external emails
	}

	log.Printf("✅ Key Transparency validation successful for email %s: sender=%s, recipient=%s",
		emailID, senderID, recipient)

	return true, nil
}

// SendEmailViaSES sends email via SES with retry logic and quota handling
func (h *SESHandler) SendEmailViaSES(ctx context.Context, emailID, senderID, recipient, subject, body string) (*SESTransaction, error) {
	// Step 1: Check quota before sending
	if err := h.checkQuota(ctx); err != nil {
		return nil, fmt.Errorf("quota check failed: %w", err)
	}

	// Step 2: Prepare email content
	emailContent := h.prepareEmailContent(recipient, subject, body)

	// Step 3: Send with retry logic
	var transaction *SESTransaction
	var lastErr error

	for attempt := 0; attempt <= h.retryConfig.MaxRetries; attempt++ {
		// Calculate delay with exponential backoff and jitter
		delay := h.calculateRetryDelay(attempt)
		if attempt > 0 {
			log.Printf("🔄 Retry attempt %d/%d for email %s, delay: %v",
				attempt, h.retryConfig.MaxRetries, emailID, delay)
			time.Sleep(delay)
		}

		// Send email
		transaction, lastErr = h.sendEmailAttempt(ctx, emailID, senderID, recipient, emailContent, attempt)
		if lastErr == nil {
			break // Success
		}

		// Check if error is retryable
		if !h.isRetryableError(lastErr) {
			log.Printf("❌ Non-retryable error for email %s: %v", emailID, lastErr)
			break
		}

		log.Printf("⚠️ Retryable error for email %s (attempt %d): %v", emailID, attempt, lastErr)
	}

	// Step 4: Handle final result
	if lastErr != nil {
		h.logSendFailure(ctx, emailID, senderID, recipient, lastErr.Error())
		return nil, fmt.Errorf("failed to send email after %d attempts: %w", h.retryConfig.MaxRetries+1, lastErr)
	}

	// Step 5: Update quota and log success
	h.updateQuota(ctx)
	h.logSendSuccess(ctx, emailID, senderID, recipient, transaction)

	return transaction, nil
}

// sendEmailAttempt performs a single email send attempt
func (h *SESHandler) sendEmailAttempt(ctx context.Context, emailID, senderID, recipient, emailContent string, attempt int) (*SESTransaction, error) {
	// Create SMTP client
	auth := smtp.PlainAuth("", h.config.SMTPUsername, h.config.SMTPPassword, h.config.SMTPHost)

	// Send email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", h.config.SMTPHost, h.config.SMTPPort),
		auth,
		h.config.DefaultSender,
		[]string{recipient},
		[]byte(emailContent),
	)

	if err != nil {
		return nil, err
	}

	// Generate transaction ID (SES doesn't return one, so we create our own)
	transactionID := fmt.Sprintf("ses_%s_%d", uuid.New().String()[:8], time.Now().Unix())

	transaction := &SESTransaction{
		TransactionID: transactionID,
		MessageID:     emailID,
		SenderID:      senderID,
		Recipient:     recipient,
		BlobID:        emailID, // Using emailID as blobID for now
		Timestamp:     time.Now(),
		Status:        "sent",
		RetryCount:    attempt,
	}

	// Store transaction in database
	if err := h.storeSESTransaction(ctx, transaction); err != nil {
		log.Printf("⚠️ Failed to store SES transaction: %v", err)
		// Don't fail the send operation for storage errors
	}

	return transaction, nil
}

// prepareEmailContent prepares the email content for sending
func (h *SESHandler) prepareEmailContent(recipient, subject, body string) string {
	content := fmt.Sprintf(`From: %s
To: %s
Subject: %s
Date: %s
Content-Type: text/plain; charset=UTF-8
MIME-Version: 1.0

%s

---
Sent via Secure Email MVP with PQC + KT validation
`, h.config.DefaultSender, recipient, subject, time.Now().Format(time.RFC1123Z), body)

	return content
}

// calculateRetryDelay calculates delay for retry with exponential backoff and jitter
func (h *SESHandler) calculateRetryDelay(attempt int) time.Duration {
	if attempt == 0 {
		return 0
	}

	// Exponential backoff
	delay := h.retryConfig.BaseDelay * time.Duration(math.Pow(h.retryConfig.BackoffMultiplier, float64(attempt-1)))

	// Cap at max delay
	if delay > h.retryConfig.MaxDelay {
		delay = h.retryConfig.MaxDelay
	}

	// Add jitter if enabled
	if h.retryConfig.JitterEnabled {
		jitter := delay / 4 // 25% jitter
		jitterAmount, _ := rand.Int(rand.Reader, big.NewInt(int64(jitter)))
		delay += time.Duration(jitterAmount.Int64())
	}

	return delay
}

// isRetryableError determines if an error is retryable
func (h *SESHandler) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Retryable errors
	retryableErrors := []string{
		"connection refused",
		"connection timeout",
		"network unreachable",
		"temporary failure",
		"rate limit exceeded",
		"quota exceeded",
		"service unavailable",
		"timeout",
		"temporary",
		"retry",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryable) {
			return true
		}
	}

	return false
}

// checkQuota checks if we can send more emails
func (h *SESHandler) checkQuota(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Update quota info if needed
	if time.Since(h.quotaInfo.LastCheck) > 5*time.Minute {
		if err := h.updateQuotaInfo(ctx); err != nil {
			log.Printf("⚠️ Failed to update quota info: %v", err)
			// Continue with cached info
		}
	}

	// Check daily quota
	if h.quotaInfo.SentToday >= h.quotaInfo.DailyQuota {
		return fmt.Errorf("daily quota exceeded: %d/%d", h.quotaInfo.SentToday, h.quotaInfo.DailyQuota)
	}

	// Check rate limit (simplified - in production, use proper rate limiting)
	// For now, we'll just log a warning if we're approaching the limit
	if h.quotaInfo.SentToday > h.quotaInfo.DailyQuota*9/10 {
		log.Printf("⚠️ Approaching daily quota: %d/%d", h.quotaInfo.SentToday, h.quotaInfo.DailyQuota)
	}

	return nil
}

// updateQuotaInfo updates quota information from database
func (h *SESHandler) updateQuotaInfo(ctx context.Context) error {
	// Get today's sent count from database
	var sentToday int
	query := `
		SELECT COUNT(*) FROM ses_transactions 
		WHERE DATE(timestamp) = DATE('now') AND status = 'sent'
	`
	err := h.db.QueryRowContext(ctx, query).Scan(&sentToday)
	if err != nil {
		return fmt.Errorf("failed to get sent count: %w", err)
	}

	h.quotaInfo.SentToday = sentToday
	h.quotaInfo.LastCheck = time.Now()
	h.quotaInfo.DailyQuota = h.config.RateLimit * 24 * 60 * 60 // Convert rate limit to daily quota
	h.quotaInfo.QuotaResetTime = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)

	return nil
}

// updateQuota increments the sent count
func (h *SESHandler) updateQuota(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.quotaInfo.SentToday++
}

// storeSESTransaction stores SES transaction in database
func (h *SESHandler) storeSESTransaction(ctx context.Context, transaction *SESTransaction) error {
	query := `
		INSERT INTO ses_transactions (
			transaction_id, message_id, sender_id, recipient, blob_id, 
			timestamp, status, retry_count, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.ExecContext(ctx, query,
		transaction.TransactionID,
		transaction.MessageID,
		transaction.SenderID,
		transaction.Recipient,
		transaction.BlobID,
		transaction.Timestamp,
		transaction.Status,
		transaction.RetryCount,
		time.Now(),
	)

	return err
}

// logValidationSuccess logs successful validation
func (h *SESHandler) logValidationSuccess(ctx context.Context, emailID, senderID, validationID string) {
	if h.auditService != nil {
		event := &audit.AuditEvent{
			LogID:          uuid.New().String(),
			Timestamp:      time.Now(),
			EventType:      audit.EventTypeSystemEvent,
			UserID:         &senderID,
			RelatedEmailID: &emailID,
			Outcome:        audit.OutcomeSuccess,
			Severity:       audit.SeverityInfo,
			Details: map[string]interface{}{
				"validation_id":   validationID,
				"validation_type": "PQC_KT_ENFORCEMENT",
				"pqc_valid":       true,
				"kt_valid":        true,
			},
		}
		h.auditService.RecordEvent(ctx, event)
	}
}

// logValidationFailure logs validation failure
func (h *SESHandler) logValidationFailure(ctx context.Context, emailID, senderID, errorCode, errorMessage, validationID string) {
	if h.auditService != nil {
		event := &audit.AuditEvent{
			LogID:          uuid.New().String(),
			Timestamp:      time.Now(),
			EventType:      audit.EventTypeSystemEvent,
			UserID:         &senderID,
			RelatedEmailID: &emailID,
			Outcome:        audit.OutcomeFailure,
			Severity:       audit.SeverityError,
			Details: map[string]interface{}{
				"validation_id":   validationID,
				"validation_type": "PQC_KT_ENFORCEMENT",
				"error_code":      errorCode,
				"error_message":   errorMessage,
			},
		}
		h.auditService.RecordEvent(ctx, event)
	}
}

// logSendSuccess logs successful email send
func (h *SESHandler) logSendSuccess(ctx context.Context, emailID, senderID, recipient string, transaction *SESTransaction) {
	if h.auditService != nil {
		event := &audit.AuditEvent{
			LogID:          uuid.New().String(),
			Timestamp:      time.Now(),
			EventType:      audit.EventTypeEmailCreation,
			UserID:         &senderID,
			RelatedEmailID: &emailID,
			Outcome:        audit.OutcomeSuccess,
			Severity:       audit.SeverityInfo,
			Details: map[string]interface{}{
				"ses_transaction_id": transaction.TransactionID,
				"recipient":          recipient,
				"retry_count":        transaction.RetryCount,
				"status":             transaction.Status,
			},
		}
		h.auditService.RecordEvent(ctx, event)
	}
}

// logSendFailure logs email send failure
func (h *SESHandler) logSendFailure(ctx context.Context, emailID, senderID, recipient, errorMessage string) {
	if h.auditService != nil {
		event := &audit.AuditEvent{
			LogID:          uuid.New().String(),
			Timestamp:      time.Now(),
			EventType:      audit.EventTypeEmailCreation,
			UserID:         &senderID,
			RelatedEmailID: &emailID,
			Outcome:        audit.OutcomeFailure,
			Severity:       audit.SeverityError,
			Details: map[string]interface{}{
				"recipient":     recipient,
				"error_message": errorMessage,
			},
		}
		h.auditService.RecordEvent(ctx, event)
	}
}

// GetSESTransaction retrieves SES transaction by ID
func (h *SESHandler) GetSESTransaction(ctx context.Context, transactionID string) (*SESTransaction, error) {
	query := `
		SELECT transaction_id, message_id, sender_id, recipient, blob_id, 
		       timestamp, status, retry_count
		FROM ses_transactions 
		WHERE transaction_id = ?
	`

	var transaction SESTransaction
	err := h.db.QueryRowContext(ctx, query, transactionID).Scan(
		&transaction.TransactionID,
		&transaction.MessageID,
		&transaction.SenderID,
		&transaction.Recipient,
		&transaction.BlobID,
		&transaction.Timestamp,
		&transaction.Status,
		&transaction.RetryCount,
	)

	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// GetQuotaInfo returns current quota information
func (h *SESHandler) GetQuotaInfo(ctx context.Context) (*SESQuotaInfo, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Update quota info if needed
	if time.Since(h.quotaInfo.LastCheck) > 5*time.Minute {
		if err := h.updateQuotaInfo(ctx); err != nil {
			return nil, err
		}
	}

	return h.quotaInfo, nil
}
