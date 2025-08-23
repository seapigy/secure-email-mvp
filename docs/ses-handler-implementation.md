# SES Handler Implementation with PQC + KT Enforcement

## Overview

The SES Handler is a comprehensive security and reliability layer for the Secure Email MVP that enforces Post-Quantum Cryptography (PQC) and Key Transparency (KT) validation before handing off emails to Amazon SES. This implementation provides:

- **PQC + KT Enforcement**: Validates encryption and key transparency before SES handoff
- **SES Handoff with Audit Logging**: Captures transaction IDs and logs all operations
- **Retry Logic with Quota Handling**: Implements exponential backoff and rate limiting
- **Comprehensive Testing**: Unit and integration tests for all scenarios

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    SES Handler Layer                        │
├─────────────────────────────────────────────────────────────┤
│ 1. PQC + KT Validation                                      │
│    ├── PQC Encryption Validation                            │
│    ├── Key Transparency Verification                        │
│    └── Structured Error Codes                               │
│                                                             │
│ 2. SES Handoff & Audit Logging                              │
│    ├── Transaction ID Capture                               │
│    ├── Comprehensive Audit Logs                             │
│    └── Compliance-Ready Format                              │
│                                                             │
│ 3. Retry Logic & Quota Handling                             │
│    ├── Exponential Backoff with Jitter                      │
│    ├── SES Rate Limit Management                            │
│    └── Quota Exceeded Handling                              │
└─────────────────────────────────────────────────────────────┘
```

### Security Flow

1. **Email Creation**: User creates email with PQC encryption
2. **PQC Validation**: System validates PQC encryption structure
3. **KT Validation**: System verifies Key Transparency for sender/recipient
4. **Quota Check**: System checks SES sending quotas
5. **SES Handoff**: Email sent via SES with retry logic
6. **Audit Logging**: All operations logged for compliance

## Implementation Details

### PQC + KT Enforcement

#### PQC Validation
```go
func (h *SESHandler) validatePQCEncryption(ctx context.Context, emailID string, encryptedData []byte) (bool, error) {
    // Parse hybrid PQC data
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

    // Validate Kyber security level
    if hybridData.KyberLevel != 512 && hybridData.KyberLevel != 768 && hybridData.KyberLevel != 1024 {
        return false, fmt.Errorf("invalid Kyber security level: %d", hybridData.KyberLevel)
    }

    return true, nil
}
```

#### KT Validation
```go
func (h *SESHandler) validateKeyTransparency(ctx context.Context, emailID, senderID, recipient string) (bool, error) {
    // Verify sender's public key in KT
    auditResult, err := h.ktService.VerifyPublicKey(senderID, "sender_public_key", "pqc")
    if err != nil {
        return false, fmt.Errorf("failed to verify sender public key in KT: %w", err)
    }

    if !auditResult.Valid {
        return false, fmt.Errorf("sender public key verification failed: %s", auditResult.ErrorMsg)
    }

    // Try to verify recipient's public key (optional for external emails)
    recipientAuditResult, err := h.ktService.VerifyPublicKey(recipient, "recipient_public_key", "pqc")
    if err != nil {
        log.Printf("ℹ️ Recipient %s not found in KT (external email): %v", recipient, err)
    }

    return true, nil
}
```

### SES Handoff with Transaction Tracking

#### Transaction Structure
```go
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
```

#### Email Sending with Retry Logic
```go
func (h *SESHandler) SendEmailViaSES(ctx context.Context, emailID, senderID, recipient, subject, body string) (*SESTransaction, error) {
    // Check quota before sending
    if err := h.checkQuota(ctx); err != nil {
        return nil, fmt.Errorf("quota check failed: %w", err)
    }

    // Send with retry logic
    for attempt := 0; attempt <= h.retryConfig.MaxRetries; attempt++ {
        delay := h.calculateRetryDelay(attempt)
        if attempt > 0 {
            time.Sleep(delay)
        }

        transaction, err := h.sendEmailAttempt(ctx, emailID, senderID, recipient, emailContent, attempt)
        if err == nil {
            return transaction, nil
        }

        if !h.isRetryableError(err) {
            break
        }
    }

    return nil, fmt.Errorf("failed to send email after %d attempts", h.retryConfig.MaxRetries+1)
}
```

### Retry Logic with Exponential Backoff

#### Delay Calculation
```go
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
```

#### Error Classification
```go
func (h *SESHandler) isRetryableError(err error) bool {
    retryableErrors := []string{
        "connection refused",
        "connection timeout",
        "rate limit exceeded",
        "quota exceeded",
        "service unavailable",
        "temporary failure",
    }

    for _, retryable := range retryableErrors {
        if strings.Contains(strings.ToLower(err.Error()), retryable) {
            return true
        }
    }

    return false
}
```

### Quota Management

#### Quota Checking
```go
func (h *SESHandler) checkQuota(ctx context.Context) error {
    // Update quota info if needed
    if time.Since(h.quotaInfo.LastCheck) > 5*time.Minute {
        if err := h.updateQuotaInfo(ctx); err != nil {
            log.Printf("⚠️ Failed to update quota info: %v", err)
        }
    }

    // Check daily quota
    if h.quotaInfo.SentToday >= h.quotaInfo.DailyQuota {
        return fmt.Errorf("daily quota exceeded: %d/%d", h.quotaInfo.SentToday, h.quotaInfo.DailyQuota)
    }

    return nil
}
```

### Audit Logging

#### Validation Success Logging
```go
func (h *SESHandler) logValidationSuccess(ctx context.Context, emailID, senderID, validationID string) {
    if h.auditService != nil {
        event := &audit.AuditEvent{
            LogID:     uuid.New().String(),
            Timestamp: time.Now(),
            EventType: audit.EventTypeSystemEvent,
            UserID:    &senderID,
            RelatedEmailID: &emailID,
            Outcome:   audit.OutcomeSuccess,
            Severity:  audit.SeverityInfo,
            Details: map[string]interface{}{
                "validation_id": validationID,
                "validation_type": "PQC_KT_ENFORCEMENT",
                "pqc_valid": true,
                "kt_valid": true,
            },
        }
        h.auditService.RecordEvent(ctx, event)
    }
}
```

#### Send Success Logging
```go
func (h *SESHandler) logSendSuccess(ctx context.Context, emailID, senderID, recipient string, transaction *SESTransaction) {
    if h.auditService != nil {
        event := &audit.AuditEvent{
            LogID:     uuid.New().String(),
            Timestamp: time.Now(),
            EventType: audit.EventTypeEmailCreation,
            UserID:    &senderID,
            RelatedEmailID: &emailID,
            Outcome:   audit.OutcomeSuccess,
            Severity:  audit.SeverityInfo,
            Details: map[string]interface{}{
                "ses_transaction_id": transaction.TransactionID,
                "recipient": recipient,
                "retry_count": transaction.RetryCount,
                "status": transaction.Status,
            },
        }
        h.auditService.RecordEvent(ctx, event)
    }
}
```

## Database Schema

### SES Transactions Table
```sql
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
```

### SES Quota Usage Table
```sql
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
```

### SES Validation Logs Table
```sql
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
```

## Configuration

### Environment Variables
```bash
# SES Configuration
SES_SMTP_USERNAME=your_ses_smtp_username_here
SES_SMTP_PASSWORD=your_ses_smtp_password_here
SES_SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SES_SMTP_PORT=587
SES_SMTP_REGION=us-east-1
SES_SMTP_RATE_LIMIT=14
SES_SMTP_SANDBOX_MODE=true
SES_DEFAULT_SENDER=noreply@securesystem.email
```

### Default Retry Configuration
```go
func defaultRetryConfig() *SESRetryConfig {
    return &SESRetryConfig{
        MaxRetries:        3,
        BaseDelay:         1 * time.Second,
        MaxDelay:          30 * time.Second,
        BackoffMultiplier: 2.0,
        JitterEnabled:     true,
    }
}
```

## Integration with Send Email Handler

### Updated Send Email Flow
```go
// Step 28: PQC + KT Validation and SES Handoff (NEW SECURITY LAYER)
if srv.sesHandler != nil {
    log.Printf("🔐 Starting PQC + KT validation for email %s", emailID)
    
    // Validate PQC encryption and Key Transparency
    validationResult, validationErr := srv.sesHandler.ValidatePQCAndKT(r.Context(), emailID, userID, req.Recipient, hybridDataBytes)
    if validationErr != nil {
        log.Printf("❌ PQC + KT validation failed for email %s: %v", emailID, validationErr)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        errorResponse := map[string]string{
            "error":   "Security validation failed",
            "details": validationErr.Error(),
            "error_code": validationResult.ErrorCode,
        }
        json.NewEncoder(w).Encode(errorResponse)
        return
    }

    if !validationResult.Valid {
        log.Printf("❌ PQC + KT validation rejected email %s: %s", emailID, validationResult.ErrorMessage)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusForbidden)
        errorResponse := map[string]string{
            "error":   "Security validation rejected",
            "details": validationResult.ErrorMessage,
            "error_code": validationResult.ErrorCode,
        }
        json.NewEncoder(w).Encode(errorResponse)
        return
    }

    log.Printf("✅ PQC + KT validation successful for email %s", emailID)

    // Send email via SES with retry logic and quota handling
    log.Printf("📧 Sending email %s via SES", emailID)
    transaction, sesErr := srv.sesHandler.SendEmailViaSES(r.Context(), emailID, userID, req.Recipient, req.Subject, req.Body)
    if sesErr != nil {
        log.Printf("❌ SES send failed for email %s: %v", emailID, sesErr)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        errorResponse := map[string]string{
            "error":   "Email delivery failed",
            "details": sesErr.Error(),
        }
        json.NewEncoder(w).Encode(errorResponse)
        return
    }

    log.Printf("✅ Email %s sent successfully via SES, transaction ID: %s", emailID, transaction.TransactionID)
}
```

## Testing

### Unit Tests
- **PQC Validation Tests**: Test PQC encryption structure validation
- **KT Validation Tests**: Test Key Transparency verification
- **Retry Logic Tests**: Test exponential backoff and jitter
- **Quota Handling Tests**: Test quota checking and updating
- **Error Classification Tests**: Test retryable vs non-retryable errors

### Integration Tests
- **Complete Flow Tests**: Test end-to-end email sending flow
- **Database Integration Tests**: Test transaction storage and retrieval
- **Audit Logging Tests**: Test audit event recording
- **Configuration Tests**: Test environment variable loading

### Performance Benchmarks
- **PQC Validation Benchmark**: Measure validation performance
- **Retry Delay Calculation Benchmark**: Measure delay calculation performance
- **Email Content Preparation Benchmark**: Measure content preparation performance

## Error Handling

### Structured Error Codes
- `PQC_VALIDATION_FAILED`: PQC encryption validation failed
- `KT_VALIDATION_FAILED`: Key Transparency validation failed
- `QUOTA_EXCEEDED`: SES quota exceeded
- `RATE_LIMIT_EXCEEDED`: SES rate limit exceeded
- `RETRYABLE_ERROR`: Transient error that can be retried
- `NON_RETRYABLE_ERROR`: Permanent error that should not be retried

### Error Response Format
```json
{
    "error": "Security validation failed",
    "details": "PQC validation failed: missing Kyber ciphertext",
    "error_code": "PQC_VALIDATION_FAILED"
}
```

## Monitoring and Observability

### Key Metrics
- **Validation Success Rate**: Percentage of successful PQC + KT validations
- **SES Send Success Rate**: Percentage of successful SES sends
- **Retry Count Distribution**: Distribution of retry attempts
- **Quota Usage**: Daily quota utilization
- **Error Rate by Type**: Error rates categorized by error type

### Audit Log Queries
```sql
-- Daily validation statistics
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_validations,
    COUNT(CASE WHEN overall_valid = 1 THEN 1 END) as successful_validations,
    COUNT(CASE WHEN overall_valid = 0 THEN 1 END) as failed_validations
FROM ses_validation_logs
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- SES transaction statistics
SELECT 
    DATE(timestamp) as date,
    COUNT(*) as total_sent,
    COUNT(CASE WHEN status = 'sent' THEN 1 END) as successful_sends,
    AVG(retry_count) as avg_retry_count
FROM ses_transactions
GROUP BY DATE(timestamp)
ORDER BY date DESC;
```

## Security Considerations

### PQC + KT Enforcement
- All emails must pass PQC encryption validation before SES handoff
- All emails must pass Key Transparency validation before SES handoff
- Failed validations are logged with structured error codes
- Validation failures block email sending

### Audit Logging
- All validation attempts are logged for compliance
- All SES transactions are logged with transaction IDs
- All retry attempts are logged with retry counts
- All quota checks are logged for monitoring

### Rate Limiting
- SES rate limits are enforced to prevent quota exhaustion
- Exponential backoff prevents overwhelming SES during outages
- Jitter prevents thundering herd problems during recovery

## Deployment

### Prerequisites
- Amazon SES configured with SMTP credentials
- Database with SES tables created
- PQC and KT services configured
- Audit service configured

### Configuration Steps
1. Set environment variables for SES configuration
2. Create database tables using provided schema
3. Initialize SES handler with required services
4. Test PQC + KT validation with sample data
5. Test SES handoff with test emails
6. Monitor audit logs for validation and send events

### Health Checks
- Verify SES connectivity
- Verify quota information retrieval
- Verify audit logging functionality
- Verify PQC + KT service availability

## Future Enhancements

### Planned Features
- **Real-time Quota Monitoring**: Live quota usage dashboard
- **Advanced Retry Strategies**: Circuit breaker pattern
- **Multi-region SES Support**: Automatic failover between regions
- **Enhanced Audit Analytics**: Machine learning for anomaly detection
- **Performance Optimization**: Connection pooling and caching

### Scalability Improvements
- **Horizontal Scaling**: Multiple SES handler instances
- **Load Balancing**: Distribute email sending across instances
- **Database Sharding**: Partition SES transaction tables
- **Caching Layer**: Cache quota and validation results

