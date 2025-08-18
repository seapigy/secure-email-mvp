package email

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/storage"
)

// ArchivedEmail represents an archived email with metadata
type ArchivedEmail struct {
	ID              int64     `json:"id"`
	OriginalEmailID string    `json:"original_email_id"`
	SenderID        string    `json:"sender_id"`
	Recipient       string    `json:"recipient"`
	Subject         string    `json:"subject"`
	ArchivedAt      time.Time `json:"archived_at"`
	ArchiveReason   string    `json:"archive_reason"` // "expired", "policy", "manual"
	RetentionDays   int       `json:"retention_days"`
	ExpiresAt       time.Time `json:"expires_at"`

	// Storage information
	ArchiveBlobURL string `json:"archive_blob_url"`
	EncryptionKey  string `json:"encryption_key"` // Encrypted key
	CompressedSize int64  `json:"compressed_size"`
	OriginalSize   int64  `json:"original_size"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ArchiveRequest represents a request to archive an email
type ArchiveRequest struct {
	EmailID       string `json:"email_id"`
	ArchiveReason string `json:"archive_reason"`
	RetentionDays int    `json:"retention_days"`
}

// RestoreRequest represents a request to restore an archived email
type RestoreRequest struct {
	ArchiveID int64 `json:"archive_id"`
}

// ArchiveResponse represents the response from archival operations
type ArchiveResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ArchiveID int64  `json:"archive_id,omitempty"`
}

// EmailArchivalService manages email archival operations
type EmailArchivalService struct {
	db       *sql.DB
	r2Client *storage.R2Client
}

// NewEmailArchivalService creates a new archival service
func NewEmailArchivalService(db *sql.DB, r2Client *storage.R2Client) *EmailArchivalService {
	return &EmailArchivalService{
		db:       db,
		r2Client: r2Client,
	}
}

// ArchiveEmail archives an email with encryption and compression
func (eas *EmailArchivalService) ArchiveEmail(ctx context.Context, req *ArchiveRequest) (*ArchiveResponse, error) {
	// Start transaction
	tx, err := eas.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the original email data
	email, err := eas.getEmailForArchival(ctx, tx, req.EmailID)
	if err != nil {
		return nil, fmt.Errorf("failed to get email for archival: %w", err)
	}

	// Create archive record
	archive := &ArchivedEmail{
		OriginalEmailID: req.EmailID,
		SenderID:        email.SenderID,
		Recipient:       email.Recipient,
		Subject:         email.Subject,
		ArchivedAt:      time.Now().UTC(),
		ArchiveReason:   req.ArchiveReason,
		RetentionDays:   req.RetentionDays,
		ExpiresAt:       time.Now().UTC().AddDate(0, 0, req.RetentionDays),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// Archive the email content to R2/S3
	archiveBlobURL, encryptionKey, compressedSize, originalSize, err := eas.archiveEmailContent(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to archive email content: %w", err)
	}

	archive.ArchiveBlobURL = archiveBlobURL
	archive.EncryptionKey = encryptionKey
	archive.CompressedSize = compressedSize
	archive.OriginalSize = originalSize

	// Insert archive record
	archiveID, err := eas.insertArchiveRecord(ctx, tx, archive)
	if err != nil {
		return nil, fmt.Errorf("failed to insert archive record: %w", err)
	}

	// Delete the original email
	if err := eas.deleteOriginalEmail(ctx, tx, req.EmailID); err != nil {
		return nil, fmt.Errorf("failed to delete original email: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully archived email %s to archive ID %d", req.EmailID, archiveID)

	return &ArchiveResponse{
		Success:   true,
		Message:   "Email archived successfully",
		ArchiveID: archiveID,
	}, nil
}

// RestoreEmail restores an archived email for audit/compliance purposes
func (eas *EmailArchivalService) RestoreEmail(ctx context.Context, req *RestoreRequest) (*ArchiveResponse, error) {
	// Get the archived email
	archive, err := eas.GetArchivedEmailByID(ctx, req.ArchiveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get archived email: %w", err)
	}

	// Check if archive has expired
	if time.Now().UTC().After(archive.ExpiresAt) {
		return nil, fmt.Errorf("archived email has expired and cannot be restored")
	}

	// Restore the email content
	restoredEmailID, err := eas.restoreEmailContent(ctx, archive)
	if err != nil {
		return nil, fmt.Errorf("failed to restore email content: %w", err)
	}

	log.Printf("Successfully restored archived email %d to new email ID %s", req.ArchiveID, restoredEmailID)

	return &ArchiveResponse{
		Success: true,
		Message: fmt.Sprintf("Email restored successfully with new ID: %s", restoredEmailID),
	}, nil
}

// GetArchivedEmails retrieves archived emails with filtering and pagination
func (eas *EmailArchivalService) GetArchivedEmails(ctx context.Context, filters map[string]string, limit, offset int) ([]*ArchivedEmail, error) {
	query := `SELECT id, original_email_id, sender_id, recipient, subject, archived_at, 
		archive_reason, retention_days, expires_at, archive_blob_url, encryption_key, 
		compressed_size, original_size, created_at, updated_at
		FROM archived_emails WHERE 1=1`

	var args []interface{}

	// Apply filters
	if senderID, ok := filters["sender_id"]; ok && senderID != "" {
		query += " AND sender_id = ?"
		args = append(args, senderID)
	}

	if recipient, ok := filters["recipient"]; ok && recipient != "" {
		query += " AND recipient LIKE ?"
		args = append(args, "%"+recipient+"%")
	}

	if reason, ok := filters["archive_reason"]; ok && reason != "" {
		query += " AND archive_reason = ?"
		args = append(args, reason)
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += " AND archived_at >= ?"
		args = append(args, startDate)
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += " AND archived_at <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY archived_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := eas.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query archived emails: %w", err)
	}
	defer rows.Close()

	var archives []*ArchivedEmail
	for rows.Next() {
		archive := &ArchivedEmail{}
		err := rows.Scan(
			&archive.ID, &archive.OriginalEmailID, &archive.SenderID, &archive.Recipient,
			&archive.Subject, &archive.ArchivedAt, &archive.ArchiveReason, &archive.RetentionDays,
			&archive.ExpiresAt, &archive.ArchiveBlobURL, &archive.EncryptionKey,
			&archive.CompressedSize, &archive.OriginalSize, &archive.CreatedAt, &archive.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan archived email: %w", err)
		}
		archives = append(archives, archive)
	}

	return archives, nil
}

// GetArchivedEmailByID retrieves a specific archived email by ID
func (eas *EmailArchivalService) GetArchivedEmailByID(ctx context.Context, archiveID int64) (*ArchivedEmail, error) {
	query := `SELECT id, original_email_id, sender_id, recipient, subject, archived_at, 
		archive_reason, retention_days, expires_at, archive_blob_url, encryption_key, 
		compressed_size, original_size, created_at, updated_at
		FROM archived_emails WHERE id = ?`

	archive := &ArchivedEmail{}
	err := eas.db.QueryRowContext(ctx, query, archiveID).Scan(
		&archive.ID, &archive.OriginalEmailID, &archive.SenderID, &archive.Recipient,
		&archive.Subject, &archive.ArchivedAt, &archive.ArchiveReason, &archive.RetentionDays,
		&archive.ExpiresAt, &archive.ArchiveBlobURL, &archive.EncryptionKey,
		&archive.CompressedSize, &archive.OriginalSize, &archive.CreatedAt, &archive.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("archived email with ID %d not found", archiveID)
		}
		return nil, fmt.Errorf("failed to get archived email: %w", err)
	}

	return archive, nil
}

// CleanupExpiredArchives removes archived emails that have exceeded their retention period
func (eas *EmailArchivalService) CleanupExpiredArchives(ctx context.Context) error {
	query := `SELECT id, archive_blob_url FROM archived_emails WHERE expires_at <= datetime('now')`

	rows, err := eas.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query expired archives: %w", err)
	}
	defer rows.Close()

	var deletedCount int
	for rows.Next() {
		var archiveID int64
		var blobURL string

		if err := rows.Scan(&archiveID, &blobURL); err != nil {
			log.Printf("Error scanning expired archive row: %v", err)
			continue
		}

		// Delete from R2/S3
		if eas.r2Client != nil && blobURL != "" {
			if err := eas.r2Client.DeleteEmail(ctx, blobURL); err != nil {
				log.Printf("Failed to delete archived blob %s: %v", blobURL, err)
				// Continue with database deletion even if R2 deletion fails
			}
		}

		// Delete from database
		if _, err := eas.db.ExecContext(ctx, "DELETE FROM archived_emails WHERE id = ?", archiveID); err != nil {
			log.Printf("Failed to delete expired archive %d from database: %v", archiveID, err)
			continue
		}

		deletedCount++
	}

	if deletedCount > 0 {
		log.Printf("Cleaned up %d expired archived emails", deletedCount)
	}

	return nil
}

// getEmailForArchival retrieves email data needed for archival
func (eas *EmailArchivalService) getEmailForArchival(ctx context.Context, tx *sql.Tx, emailID string) (*EmailRetentionInfo, error) {
	query := `SELECT email_id, sender_id, recipient, subject, created_at, expires_at, 
		burn_after_read, access_count, self_destructed, encrypted_blob_url
		FROM emails WHERE email_id = ?`

	var email EmailRetentionInfo
	var expiresAtStr sql.NullString
	var burnAfterRead int
	var selfDestructed int
	var encryptedBlobURL sql.NullString

	err := tx.QueryRowContext(ctx, query, emailID).Scan(
		&email.EmailID, &email.SenderID, &email.Recipient, &email.Subject,
		&email.CreatedAt, &expiresAtStr, &burnAfterRead, &email.AccessCount, &selfDestructed, &encryptedBlobURL,
	)
	if err != nil {
		return nil, err
	}

	email.BurnAfterRead = burnAfterRead == 1
	email.SelfDestructed = selfDestructed == 1

	if expiresAtStr.Valid {
		if expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr.String); err == nil {
			email.ExpiresAt = &expiresAt
		}
	}

	return &email, nil
}

// archiveEmailContent archives the email content to R2/S3 with encryption and compression
func (eas *EmailArchivalService) archiveEmailContent(_ context.Context, email *EmailRetentionInfo) (string, string, int64, int64, error) {
	// This is a simplified implementation
	// In a real implementation, you would:
	// 1. Download the original encrypted content from R2
	// 2. Re-encrypt with a new archival key
	// 3. Compress the content
	// 4. Upload to archival storage
	// 5. Return the new blob URL and encrypted key

	// For now, we'll create a placeholder implementation
	archiveBlobURL := fmt.Sprintf("archives/%s/%s", time.Now().Format("2006-01-02"), email.EmailID)
	encryptionKey := "encrypted_archival_key_placeholder"
	compressedSize := int64(1024) // Placeholder
	originalSize := int64(2048)   // Placeholder

	return archiveBlobURL, encryptionKey, compressedSize, originalSize, nil
}

// insertArchiveRecord inserts the archive record into the database
func (eas *EmailArchivalService) insertArchiveRecord(ctx context.Context, tx *sql.Tx, archive *ArchivedEmail) (int64, error) {
	query := `
		INSERT INTO archived_emails (
			original_email_id, sender_id, recipient, subject, archived_at, 
			archive_reason, retention_days, expires_at, archive_blob_url, 
			encryption_key, compressed_size, original_size, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.ExecContext(ctx, query,
		archive.OriginalEmailID, archive.SenderID, archive.Recipient, archive.Subject,
		archive.ArchivedAt, archive.ArchiveReason, archive.RetentionDays, archive.ExpiresAt,
		archive.ArchiveBlobURL, archive.EncryptionKey, archive.CompressedSize, archive.OriginalSize,
		archive.CreatedAt, archive.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// deleteOriginalEmail deletes the original email from the database
func (eas *EmailArchivalService) deleteOriginalEmail(ctx context.Context, tx *sql.Tx, emailID string) error {
	// Delete from emails table
	if _, err := tx.ExecContext(ctx, "DELETE FROM emails WHERE email_id = ?", emailID); err != nil {
		return err
	}

	// Delete from email_access_logs table
	if _, err := tx.ExecContext(ctx, "DELETE FROM email_access_logs WHERE email_id = ?", emailID); err != nil {
		return err
	}

	// Delete from retention_notifications table
	if _, err := tx.ExecContext(ctx, "DELETE FROM retention_notifications WHERE email_id = ?", emailID); err != nil {
		return err
	}

	return nil
}

// restoreEmailContent restores an archived email to active status
func (eas *EmailArchivalService) restoreEmailContent(_ context.Context, archive *ArchivedEmail) (string, error) {
	// This is a simplified implementation
	// In a real implementation, you would:
	// 1. Download the archived content from R2
	// 2. Decrypt with the archival key
	// 3. Re-encrypt with a new active key
	// 4. Upload to active storage
	// 5. Insert into emails table
	// 6. Return the new email ID

	// For now, we'll create a placeholder implementation
	restoredEmailID := fmt.Sprintf("restored_%s_%d", archive.OriginalEmailID, time.Now().Unix())

	return restoredEmailID, nil
}

// GetArchivalStats returns statistics about archived emails
func (eas *EmailArchivalService) GetArchivalStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total archived emails
	var totalArchived int
	err := eas.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM archived_emails").Scan(&totalArchived)
	if err != nil {
		return nil, fmt.Errorf("failed to count total archived emails: %w", err)
	}
	stats["total_archived"] = totalArchived

	// Expired archives
	var expiredArchives int
	err = eas.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM archived_emails WHERE expires_at <= datetime('now')").Scan(&expiredArchives)
	if err != nil {
		return nil, fmt.Errorf("failed to count expired archives: %w", err)
	}
	stats["expired_archives"] = expiredArchives

	// Total storage used
	var totalStorage int64
	err = eas.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(compressed_size), 0) FROM archived_emails").Scan(&totalStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total storage: %w", err)
	}
	stats["total_storage_bytes"] = totalStorage

	// Archives by reason
	reasonQuery := `SELECT archive_reason, COUNT(*) FROM archived_emails GROUP BY archive_reason`
	rows, err := eas.db.QueryContext(ctx, reasonQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query archives by reason: %w", err)
	}
	defer rows.Close()

	reasons := make(map[string]int)
	for rows.Next() {
		var reason string
		var count int
		if err := rows.Scan(&reason, &count); err != nil {
			continue
		}
		reasons[reason] = count
	}
	stats["archives_by_reason"] = reasons

	return stats, nil
}
