package email

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/storage"
)

// RetentionConfig holds configuration for email retention policies
type RetentionConfig struct {
	DefaultExpirationDays int  // Default expiration in days for new emails
	CleanupAuditLogs      bool // Whether to delete audit logs when cleaning up emails
	EnableNotifications   bool // Whether to log notifications for removed emails
	BatchSize             int  // Number of emails to process in each cleanup batch
}

// EmailRetentionService provides comprehensive email retention and cleanup functionality
type EmailRetentionService struct {
	db           *sql.DB
	r2Client     *storage.R2Client
	config       RetentionConfig
	lastCleanup  time.Time
	cleanupStats CleanupStats
	notification *RetentionNotificationService
}

// CleanupStats tracks statistics about cleanup operations
type CleanupStats struct {
	TotalProcessed        int
	ExpiredDeleted        int
	BurnAfterReadDeleted  int
	SelfDestructedDeleted int
	FailedDeletions       int
	LastCleanupTime       time.Time
	AuditLogsDeleted      int
}

// EmailRetentionInfo represents information about an email's retention status
type EmailRetentionInfo struct {
	EmailID         string     `json:"email_id"`
	SenderID        string     `json:"sender_id"`
	Recipient       string     `json:"recipient"`
	Subject         string     `json:"subject"`
	CreatedAt       time.Time  `json:"created_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	BurnAfterRead   bool       `json:"burn_after_read"`
	AccessCount     int        `json:"access_count"`
	SelfDestructed  bool       `json:"self_destructed"`
	Status          string     `json:"status"` // "active", "expired", "burned", "self_destructed"
	DaysUntilExpiry *int       `json:"days_until_expiry,omitempty"`
	PendingDeletion bool       `json:"pending_deletion"`
}

// NewEmailRetentionService creates a new email retention service
func NewEmailRetentionService(db *sql.DB, r2Client *storage.R2Client) *EmailRetentionService {
	config := RetentionConfig{
		DefaultExpirationDays: getDefaultExpirationDays(),
		CleanupAuditLogs:      getCleanupAuditLogs(),
		EnableNotifications:   getEnableNotifications(),
		BatchSize:             getBatchSize(),
	}

	return &EmailRetentionService{
		db:           db,
		r2Client:     r2Client,
		config:       config,
		notification: NewRetentionNotificationService(db),
	}
}

// getDefaultExpirationDays gets the default expiration days from environment
func getDefaultExpirationDays() int {
	daysStr := os.Getenv("DEFAULT_EMAIL_EXPIRATION_DAYS")
	if daysStr == "" {
		return 30 // Default to 30 days
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		return 30 // Default fallback
	}

	return days
}

// getCleanupAuditLogs gets whether to cleanup audit logs from environment
func getCleanupAuditLogs() bool {
	cleanupStr := os.Getenv("CLEANUP_AUDIT_LOGS")
	if cleanupStr == "" {
		return false // Default to keeping audit logs
	}

	cleanup, err := strconv.ParseBool(cleanupStr)
	if err != nil {
		return false // Default fallback
	}

	return cleanup
}

// getEnableNotifications gets whether to enable notifications from environment
func getEnableNotifications() bool {
	notifyStr := os.Getenv("ENABLE_CLEANUP_NOTIFICATIONS")
	if notifyStr == "" {
		return true // Default to enabling notifications
	}

	notify, err := strconv.ParseBool(notifyStr)
	if err != nil {
		return true // Default fallback
	}

	return notify
}

// getBatchSize gets the batch size for cleanup operations from environment
func getBatchSize() int {
	batchStr := os.Getenv("CLEANUP_BATCH_SIZE")
	if batchStr == "" {
		return 100 // Default to 100 emails per batch
	}

	batch, err := strconv.Atoi(batchStr)
	if err != nil || batch <= 0 {
		return 100 // Default fallback
	}

	return batch
}

// GetDefaultExpirationTime returns the default expiration time for new emails
func (ers *EmailRetentionService) GetDefaultExpirationTime() time.Time {
	return time.Now().UTC().AddDate(0, 0, ers.config.DefaultExpirationDays)
}

// SetEmailExpiration sets the expiration time for a specific email
func (ers *EmailRetentionService) SetEmailExpiration(emailID, senderID string, expiresAt *time.Time) error {
	// Verify the email exists and belongs to the sender
	var existingSenderID string
	err := ers.db.QueryRow(`
		SELECT sender_id FROM emails WHERE email_id = ?
	`, emailID).Scan(&existingSenderID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("email not found")
		}
		return fmt.Errorf("database error: %w", err)
	}

	if existingSenderID != senderID {
		return fmt.Errorf("unauthorized: email does not belong to sender")
	}

	// Update the expiration time
	_, err = ers.db.Exec(`
		UPDATE emails SET expires_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE email_id = ?
	`, expiresAt, emailID)

	if err != nil {
		return fmt.Errorf("failed to update expiration: %w", err)
	}

	log.Printf("Updated expiration for email %s to %v", emailID, expiresAt)
	return nil
}

// GetEmailsPendingCleanup retrieves emails that are pending cleanup with pagination and filtering
func (ers *EmailRetentionService) GetEmailsPendingCleanup(ctx context.Context, filters map[string]string, limit, offset int) ([]EmailRetentionInfo, error) {
	query := `
		SELECT email_id, sender_id, recipient, subject, created_at, expires_at, 
		       burn_after_read, access_count, self_destructed
		FROM emails 
		WHERE (
			(expires_at IS NOT NULL AND expires_at <= datetime('now') AND encrypted_blob_url IS NOT NULL) OR
			(burn_after_read = 1 AND access_count > 0 AND encrypted_blob_url IS NOT NULL) OR
			(self_destructed = 1 AND encrypted_blob_url IS NOT NULL)
		)
	`

	args := []interface{}{}
	argIndex := 1

	// Add filters
	if userID, ok := filters["user_id"]; ok && userID != "" {
		query += " AND sender_id = ?"
		args = append(args, userID)
		argIndex++
	}

	if status, ok := filters["status"]; ok && status != "" {
		switch status {
		case "expired":
			query += " AND expires_at IS NOT NULL AND expires_at <= datetime('now')"
		case "burned":
			query += " AND burn_after_read = 1 AND access_count > 0"
		case "self_destructed":
			query += " AND self_destructed = 1"
		}
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += " AND created_at >= ?"
		args = append(args, startDate)
		argIndex++
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += " AND created_at <= ?"
		args = append(args, endDate)
		argIndex++
	}

	query += " ORDER BY created_at ASC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := ers.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query emails pending cleanup: %w", err)
	}
	defer rows.Close()

	var emails []EmailRetentionInfo
	now := time.Now().UTC()

	for rows.Next() {
		var email EmailRetentionInfo
		var expiresAtStr sql.NullString
		var burnAfterRead int
		var selfDestructed int

		err := rows.Scan(
			&email.EmailID, &email.SenderID, &email.Recipient, &email.Subject,
			&email.CreatedAt, &expiresAtStr, &burnAfterRead, &email.AccessCount, &selfDestructed,
		)
		if err != nil {
			log.Printf("Error scanning email row: %v", err)
			continue
		}

		email.BurnAfterRead = burnAfterRead == 1
		email.SelfDestructed = selfDestructed == 1

		// Parse expires_at if present
		if expiresAtStr.Valid {
			if expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr.String); err == nil {
				email.ExpiresAt = &expiresAt
			}
		}

		// Determine status and pending deletion
		email.Status = ers.determineEmailStatus(email, now)
		email.PendingDeletion = ers.isEmailPendingDeletion(email, now)

		// Calculate days until expiry if applicable
		if email.ExpiresAt != nil && email.ExpiresAt.After(now) {
			days := int(email.ExpiresAt.Sub(now).Hours() / 24)
			email.DaysUntilExpiry = &days
		}

		emails = append(emails, email)
	}

	return emails, nil
}

// GetEmailsPendingCleanupCount returns the total count of emails pending cleanup
func (ers *EmailRetentionService) GetEmailsPendingCleanupCount(ctx context.Context, filters map[string]string) (int, error) {
	query := `
		SELECT COUNT(*) FROM emails 
		WHERE (
			(expires_at IS NOT NULL AND expires_at <= datetime('now') AND encrypted_blob_url IS NOT NULL) OR
			(burn_after_read = 1 AND access_count > 0 AND encrypted_blob_url IS NOT NULL) OR
			(self_destructed = 1 AND encrypted_blob_url IS NOT NULL)
		)
	`

	args := []interface{}{}

	// Add filters
	if userID, ok := filters["user_id"]; ok && userID != "" {
		query += " AND sender_id = ?"
		args = append(args, userID)
	}

	if status, ok := filters["status"]; ok && status != "" {
		switch status {
		case "expired":
			query += " AND expires_at IS NOT NULL AND expires_at <= datetime('now')"
		case "burned":
			query += " AND burn_after_read = 1 AND access_count > 0"
		case "self_destructed":
			query += " AND self_destructed = 1"
		}
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += " AND created_at >= ?"
		args = append(args, startDate)
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += " AND created_at <= ?"
		args = append(args, endDate)
	}

	var count int
	err := ers.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count emails pending cleanup: %w", err)
	}

	return count, nil
}

// GetRetentionStatistics returns comprehensive retention statistics
func (ers *EmailRetentionService) GetRetentionStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count expired emails
	var expiredCount int
	err := ers.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM emails 
		WHERE expires_at IS NOT NULL AND expires_at <= datetime('now') AND encrypted_blob_url IS NOT NULL
	`).Scan(&expiredCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count expired emails: %w", err)
	}

	// Count burn-after-read emails that have been accessed
	var burnAfterReadCount int
	err = ers.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM emails 
		WHERE burn_after_read = 1 AND access_count > 0 AND encrypted_blob_url IS NOT NULL
	`).Scan(&burnAfterReadCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count burn-after-read emails: %w", err)
	}

	// Count self-destructed emails
	var selfDestructedCount int
	err = ers.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM emails 
		WHERE self_destructed = 1 AND encrypted_blob_url IS NOT NULL
	`).Scan(&selfDestructedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count self-destructed emails: %w", err)
	}

	// Count total emails with content
	var totalWithContent int
	err = ers.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM emails WHERE encrypted_blob_url IS NOT NULL
	`).Scan(&totalWithContent)
	if err != nil {
		return nil, fmt.Errorf("failed to count total emails: %w", err)
	}

	// Count soft-deleted emails (no content)
	var softDeletedCount int
	err = ers.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM emails WHERE encrypted_blob_url IS NULL
	`).Scan(&softDeletedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count soft-deleted emails: %w", err)
	}

	// Count emails expiring in next 24 hours
	var expiringSoonCount int
	err = ers.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM emails 
		WHERE expires_at IS NOT NULL 
		AND expires_at > datetime('now') 
		AND expires_at <= datetime('now', '+1 day')
		AND encrypted_blob_url IS NOT NULL
	`).Scan(&expiringSoonCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count emails expiring soon: %w", err)
	}

	stats["expired_emails"] = expiredCount
	stats["burn_after_read_emails"] = burnAfterReadCount
	stats["self_destructed_emails"] = selfDestructedCount
	stats["total_emails_with_content"] = totalWithContent
	stats["soft_deleted_emails"] = softDeletedCount
	stats["emails_expiring_soon"] = expiringSoonCount
	stats["emails_pending_deletion"] = expiredCount + burnAfterReadCount + selfDestructedCount
	stats["total_emails"] = totalWithContent + softDeletedCount
	stats["default_expiration_days"] = ers.config.DefaultExpirationDays
	stats["cleanup_audit_logs"] = ers.config.CleanupAuditLogs
	stats["enable_notifications"] = ers.config.EnableNotifications
	stats["batch_size"] = ers.config.BatchSize
	stats["last_cleanup_time"] = ers.lastCleanup
	stats["cleanup_stats"] = ers.cleanupStats

	return stats, nil
}

// PerformCleanup executes the cleanup process with transactional safety
func (ers *EmailRetentionService) PerformCleanup(ctx context.Context) error {
	log.Printf("Starting email retention cleanup process...")
	startTime := time.Now()

	// Get emails to delete in batches
	offset := 0
	totalProcessed := 0
	totalDeleted := 0
	totalSkipped := 0

	for {
		// Get batch of emails to delete
		emails, err := ers.GetEmailsPendingCleanup(ctx, map[string]string{}, ers.config.BatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get emails for cleanup: %w", err)
		}

		if len(emails) == 0 {
			break // No more emails to process
		}

		// Process each email in the batch
		for _, email := range emails {
			if err := ers.deleteEmailWithTransaction(ctx, email); err != nil {
				log.Printf("Failed to delete email %s: %v", email.EmailID, err)
				ers.cleanupStats.FailedDeletions++
				totalSkipped++
			} else {
				totalDeleted++
				ers.updateCleanupStats(email)

				// Send cleanup notification if enabled
				if ers.notification != nil {
					cleanupNotification := &CleanupNotification{
						EmailID:       email.EmailID,
						SenderID:      email.SenderID,
						Recipient:     email.Recipient,
						Subject:       email.Subject,
						CleanupReason: email.Status,
						CleanupTime:   time.Now(),
						Initiator:     "worker",
					}
					if err := ers.notification.SendCleanupNotification(ctx, cleanupNotification); err != nil {
						log.Printf("Failed to send cleanup notification for email %s: %v", email.EmailID, err)
					}
				}

				if ers.config.EnableNotifications {
					log.Printf("Successfully deleted email %s (reason: %s)", email.EmailID, email.Status)
				}
			}
			totalProcessed++
		}

		offset += ers.config.BatchSize

		// Check if we should continue processing
		if len(emails) < ers.config.BatchSize {
			break
		}
	}

	// Update cleanup statistics
	ers.cleanupStats.TotalProcessed += totalProcessed
	ers.lastCleanup = time.Now()

	// Log cleanup operation
	if err := ers.logCleanupOperation(ctx, totalProcessed, totalDeleted, totalSkipped, "worker", startTime); err != nil {
		log.Printf("Failed to log cleanup operation: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("Cleanup completed in %v: %d processed, %d deleted, %d skipped, %d failed",
		duration, totalProcessed, totalDeleted, totalSkipped, ers.cleanupStats.FailedDeletions)

	return nil
}

// deleteEmailWithTransaction deletes a single email with transactional safety
func (ers *EmailRetentionService) deleteEmailWithTransaction(ctx context.Context, email EmailRetentionInfo) error {
	tx, err := ers.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the blob URL for R2 deletion
	var blobURL string
	err = tx.QueryRowContext(ctx, `
		SELECT encrypted_blob_url FROM emails WHERE email_id = ?
	`, email.EmailID).Scan(&blobURL)
	if err != nil {
		return fmt.Errorf("failed to get blob URL: %w", err)
	}

	// Delete from R2 storage first
	if ers.r2Client != nil && blobURL != "" {
		if err := ers.r2Client.DeleteEmail(ctx, blobURL); err != nil {
			log.Printf("Failed to delete email %s from R2: %v", email.EmailID, err)
			// Continue with database cleanup even if R2 deletion fails
		}
	}

	// Soft delete from database (remove encryption data but keep metadata)
	_, err = tx.ExecContext(ctx, `
		UPDATE emails SET 
			encrypted_blob_url = NULL,
			encrypted_key = NULL,
			encryption_nonce = NULL,
			encryption_auth_tag = NULL,
			sha256_hash = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE email_id = ?`,
		email.EmailID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark email as deleted in database: %w", err)
	}

	// Delete associated audit logs if configured
	if ers.config.CleanupAuditLogs {
		auditResult, err := tx.ExecContext(ctx, `
			DELETE FROM email_access_logs WHERE email_id = ?
		`, email.EmailID)
		if err != nil {
			log.Printf("Failed to delete audit logs for email %s: %v", email.EmailID, err)
		} else {
			rowsAffected, _ := auditResult.RowsAffected()
			ers.cleanupStats.AuditLogsDeleted += int(rowsAffected)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// determineEmailStatus determines the current status of an email
func (ers *EmailRetentionService) determineEmailStatus(email EmailRetentionInfo, now time.Time) string {
	if email.SelfDestructed {
		return "self_destructed"
	}

	if email.ExpiresAt != nil && now.After(*email.ExpiresAt) {
		return "expired"
	}

	if email.BurnAfterRead && email.AccessCount > 0 {
		return "burned"
	}

	return "active"
}

// isEmailPendingDeletion determines if an email is pending deletion
func (ers *EmailRetentionService) isEmailPendingDeletion(email EmailRetentionInfo, now time.Time) bool {
	if email.SelfDestructed {
		return true
	}

	if email.ExpiresAt != nil && now.After(*email.ExpiresAt) {
		return true
	}

	if email.BurnAfterRead && email.AccessCount > 0 {
		return true
	}

	return false
}

// updateCleanupStats updates the cleanup statistics based on the deleted email
func (ers *EmailRetentionService) updateCleanupStats(email EmailRetentionInfo) {
	switch email.Status {
	case "expired":
		ers.cleanupStats.ExpiredDeleted++
	case "burned":
		ers.cleanupStats.BurnAfterReadDeleted++
	case "self_destructed":
		ers.cleanupStats.SelfDestructedDeleted++
	}
}

// GetCleanupStats returns the current cleanup statistics
func (ers *EmailRetentionService) GetCleanupStats() CleanupStats {
	return ers.cleanupStats
}

// logCleanupOperation logs a cleanup operation to the cleanup_logs table
func (ers *EmailRetentionService) logCleanupOperation(ctx context.Context, emailsProcessed, emailsDeleted, emailsSkipped int, initiator string, startTime time.Time) error {
	duration := time.Since(startTime)

	query := `
		INSERT INTO cleanup_logs (
			log_id, cleanup_reason, cleanup_time, initiator,
			emails_processed, emails_deleted, emails_skipped, audit_logs_deleted, duration, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	logID := fmt.Sprintf("cleanup_%s_%d", initiator, time.Now().Unix())
	cleanupReason := "batch_cleanup"
	auditLogsDeleted := 0
	if ers.config.CleanupAuditLogs {
		auditLogsDeleted = emailsDeleted // Estimate
	}

	metadata := fmt.Sprintf(`{"batch_size": %d, "cleanup_audit_logs": %t}`, ers.config.BatchSize, ers.config.CleanupAuditLogs)

	_, err := ers.db.ExecContext(ctx, query,
		logID, cleanupReason, time.Now(), initiator,
		emailsProcessed, emailsDeleted, emailsSkipped, auditLogsDeleted,
		duration.String(), metadata,
	)

	return err
}
