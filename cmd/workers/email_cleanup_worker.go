package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/storage"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// R2ClientInterface defines the interface for R2 operations
type R2ClientInterface interface {
	DeleteEmail(ctx context.Context, blobID string) error
	GetEmail(ctx context.Context, blobID string) ([]byte, error)
	UploadEmail(ctx context.Context, blobID string, data []byte) error
	EmailExists(ctx context.Context, blobID string) (bool, error)
	GetEmailMetadata(ctx context.Context, blobID string) (map[string]string, error)
}

// EmailCleanupWorker handles automated deletion of expired and consumed emails
type EmailCleanupWorker struct {
	db              *sql.DB
	r2Client        R2ClientInterface
	cleanupInterval time.Duration
	stopChan        chan bool
}

// NewEmailCleanupWorker creates a new cleanup worker instance
func NewEmailCleanupWorker(dbPath string, cleanupIntervalMinutes int) (*EmailCleanupWorker, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test database connectivity
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize R2 client
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 client: %w", err)
	}

	interval := time.Duration(cleanupIntervalMinutes) * time.Minute
	if interval < time.Minute {
		interval = time.Minute // Minimum 1 minute interval
	}

	return &EmailCleanupWorker{
		db:              db,
		r2Client:        r2Client,
		cleanupInterval: interval,
		stopChan:        make(chan bool),
	}, nil
}

// NewEmailCleanupWorkerWithR2Client creates a new cleanup worker with a custom R2 client (for testing)
func NewEmailCleanupWorkerWithR2Client(db *sql.DB, r2Client R2ClientInterface, cleanupIntervalMinutes int) (*EmailCleanupWorker, error) {
	interval := time.Duration(cleanupIntervalMinutes) * time.Minute
	if interval < time.Minute {
		interval = time.Minute // Minimum 1 minute interval
	}

	return &EmailCleanupWorker{
		db:              db,
		r2Client:        r2Client,
		cleanupInterval: interval,
		stopChan:        make(chan bool),
	}, nil
}

// Start begins the cleanup worker in a goroutine
func (w *EmailCleanupWorker) Start() {
	log.Printf("Starting email cleanup worker with interval: %v", w.cleanupInterval)

	go func() {
		ticker := time.NewTicker(w.cleanupInterval)
		defer ticker.Stop()

		// Run initial cleanup immediately
		w.performCleanup()

		for {
			select {
			case <-ticker.C:
				w.performCleanup()
			case <-w.stopChan:
				log.Printf("Email cleanup worker stopped")
				return
			}
		}
	}()
}

// Stop gracefully stops the cleanup worker
func (w *EmailCleanupWorker) Stop() {
	log.Printf("Stopping email cleanup worker...")
	close(w.stopChan)
	w.db.Close()
}

// performCleanup executes the actual cleanup process
func (w *EmailCleanupWorker) performCleanup() {
	log.Printf("Starting email cleanup process...")
	startTime := time.Now()

	// Get emails to delete
	emailsToDelete, err := w.getEmailsToDelete()
	if err != nil {
		log.Printf("Error getting emails to delete: %v", err)
		return
	}

	if len(emailsToDelete) == 0 {
		log.Printf("No emails to delete in this cleanup cycle")
		return
	}

	log.Printf("Found %d emails to delete", len(emailsToDelete))

	// Process each email for deletion
	successCount := 0
	failureCount := 0

	for _, email := range emailsToDelete {
		if err := w.deleteEmail(email); err != nil {
			log.Printf("Failed to delete email %s: %v", email.ID, err)
			failureCount++
		} else {
			log.Printf("Successfully deleted email %s (reason: %s)", email.ID, email.DeleteReason)
			successCount++
		}
	}

	duration := time.Since(startTime)
	log.Printf("Cleanup completed in %v: %d successful, %d failed", duration, successCount, failureCount)
}

// EmailToDelete represents an email that needs to be deleted
type EmailToDelete struct {
	ID            string
	BlobID        string
	DeleteReason  string
	ExpiresAt     *time.Time
	BurnAfterRead bool
	AccessCount   int
}

// getEmailsToDelete retrieves emails that should be deleted
func (w *EmailCleanupWorker) getEmailsToDelete() ([]EmailToDelete, error) {
	// Use Go's time instead of SQLite's datetime function to avoid timezone issues
	currentTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	query := `
		SELECT email_id, encrypted_blob_url, expires_at, burn_after_read, access_count
		FROM emails 
		WHERE (
			(expires_at IS NOT NULL AND expires_at <= ?) OR
			(burn_after_read = 1 AND access_count > 0)
		) AND encrypted_blob_url IS NOT NULL
		ORDER BY created_at ASC
	`

	rows, err := w.db.Query(query, currentTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query emails to delete: %w", err)
	}
	defer rows.Close()

	var emails []EmailToDelete
	for rows.Next() {
		var email EmailToDelete
		var expiresAtStr sql.NullString
		var burnAfterRead int

		err := rows.Scan(&email.ID, &email.BlobID, &expiresAtStr, &burnAfterRead, &email.AccessCount)
		if err != nil {
			log.Printf("Error scanning email row: %v", err)
			continue
		}

		email.BurnAfterRead = burnAfterRead == 1

		// Parse expires_at if present
		if expiresAtStr.Valid {
			if expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr.String); err == nil {
				email.ExpiresAt = &expiresAt
			}
		}

		// Determine delete reason
		if email.ExpiresAt != nil && time.Now().UTC().After(*email.ExpiresAt) {
			email.DeleteReason = "expired"
		} else if email.BurnAfterRead && email.AccessCount > 0 {
			email.DeleteReason = "burn-after-read"
		}

		emails = append(emails, email)
	}

	return emails, nil
}

// deleteEmail deletes a single email from both database and R2 storage
func (w *EmailCleanupWorker) deleteEmail(email EmailToDelete) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Delete from R2 storage first
	if err := w.r2Client.DeleteEmail(ctx, email.BlobID); err != nil {
		log.Printf("Failed to delete email %s from R2: %v", email.ID, err)
		// Continue with database cleanup even if R2 deletion fails
	} else {
		log.Printf("Successfully deleted email blob %s from R2", email.BlobID)
	}

	// Soft delete from database (remove encryption data but keep metadata)
	_, err := w.db.Exec(`
		UPDATE emails SET 
			encrypted_blob_url = NULL,
			encrypted_key = NULL,
			encryption_nonce = NULL,
			encryption_auth_tag = NULL,
			sha256_hash = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE email_id = ?`,
		email.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark email as deleted in database: %w", err)
	}

	return nil
}

// RunCleanupOnce performs a single cleanup cycle (useful for manual execution)
func (w *EmailCleanupWorker) RunCleanupOnce() error {
	log.Printf("Running one-time email cleanup...")
	w.performCleanup()
	return nil
}

// GetCleanupStats returns statistics about the cleanup process
func (w *EmailCleanupWorker) GetCleanupStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count expired emails
	var expiredCount int
	err := w.db.QueryRow(`
		SELECT COUNT(*) FROM emails 
		WHERE expires_at IS NOT NULL AND expires_at <= datetime('now') AND encrypted_blob_url IS NOT NULL
	`).Scan(&expiredCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count expired emails: %w", err)
	}

	// Count burn-after-read emails that have been accessed
	var burnAfterReadCount int
	err = w.db.QueryRow(`
		SELECT COUNT(*) FROM emails 
		WHERE burn_after_read = 1 AND access_count > 0 AND encrypted_blob_url IS NOT NULL
	`).Scan(&burnAfterReadCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count burn-after-read emails: %w", err)
	}

	// Count total emails with content
	var totalWithContent int
	err = w.db.QueryRow(`
		SELECT COUNT(*) FROM emails WHERE encrypted_blob_url IS NOT NULL
	`).Scan(&totalWithContent)
	if err != nil {
		return nil, fmt.Errorf("failed to count total emails: %w", err)
	}

	stats["expired_emails"] = expiredCount
	stats["burn_after_read_emails"] = burnAfterReadCount
	stats["total_emails_with_content"] = totalWithContent
	stats["cleanup_interval_minutes"] = int(w.cleanupInterval.Minutes())

	return stats, nil
}

// main function for standalone worker execution
func main() {
	log.Printf("Starting Email Cleanup Worker...")

	// Get configuration from environment
	cleanupIntervalStr := os.Getenv("EMAIL_CLEANUP_INTERVAL_MINUTES")
	if cleanupIntervalStr == "" {
		cleanupIntervalStr = "15" // Default to 15 minutes
	}

	cleanupInterval, err := strconv.Atoi(cleanupIntervalStr)
	if err != nil {
		log.Fatalf("Invalid EMAIL_CLEANUP_INTERVAL_MINUTES: %v", err)
	}

	dbPath := os.Getenv("SQLITE_DB")
	if dbPath == "" {
		dbPath = "/var/db/secure-email.db"
	}

	// Create and start the worker
	worker, err := NewEmailCleanupWorker(dbPath, cleanupInterval)
	if err != nil {
		log.Fatalf("Failed to create cleanup worker: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the worker
	worker.Start()

	// Wait for shutdown signal
	<-sigChan
	log.Printf("Received shutdown signal, stopping worker...")
	worker.Stop()
}
