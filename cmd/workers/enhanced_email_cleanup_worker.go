package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/email"
	"secure-email-mvp/pkg/storage"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// EnhancedEmailCleanupWorker provides advanced email cleanup with retention policies
type EnhancedEmailCleanupWorker struct {
	db                  *sql.DB
	r2Client            *storage.R2Client
	retentionService    *email.EmailRetentionService
	cleanupInterval     time.Duration
	stopChan            chan bool
	lastCleanupTime     time.Time
	cleanupStats        email.CleanupStats
	enableNotifications bool
	cleanupAuditLogs    bool
}

// NewEnhancedEmailCleanupWorker creates a new enhanced cleanup worker
func NewEnhancedEmailCleanupWorker(dbPath string, cleanupIntervalMinutes int) (*EnhancedEmailCleanupWorker, error) {
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
		log.Printf("Warning: Failed to initialize R2 client: %v", err)
		r2Client = nil
	}

	// Initialize retention service
	retentionService := email.NewEmailRetentionService(db, r2Client)

	// Get configuration from environment
	enableNotifications := getEnableNotifications()
	cleanupAuditLogs := getCleanupAuditLogs()

	interval := time.Duration(cleanupIntervalMinutes) * time.Minute
	if interval < time.Minute {
		interval = time.Minute // Minimum 1 minute interval
	}

	return &EnhancedEmailCleanupWorker{
		db:                  db,
		r2Client:            r2Client,
		retentionService:    retentionService,
		cleanupInterval:     interval,
		stopChan:            make(chan bool),
		enableNotifications: enableNotifications,
		cleanupAuditLogs:    cleanupAuditLogs,
	}, nil
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

// Start begins the enhanced cleanup worker in a goroutine
func (w *EnhancedEmailCleanupWorker) Start() {
	log.Printf("Starting enhanced email cleanup worker with interval: %v", w.cleanupInterval)
	log.Printf("Configuration: notifications=%v, cleanup_audit_logs=%v", w.enableNotifications, w.cleanupAuditLogs)

	go func() {
		ticker := time.NewTicker(w.cleanupInterval)
		defer ticker.Stop()

		// Run initial cleanup immediately
		w.performEnhancedCleanup()

		for {
			select {
			case <-ticker.C:
				w.performEnhancedCleanup()
			case <-w.stopChan:
				log.Printf("Enhanced email cleanup worker stopped")
				return
			}
		}
	}()
}

// Stop gracefully stops the enhanced cleanup worker
func (w *EnhancedEmailCleanupWorker) Stop() {
	log.Printf("Stopping enhanced email cleanup worker...")
	close(w.stopChan)
	w.db.Close()
}

// performEnhancedCleanup executes the enhanced cleanup process using the retention service
func (w *EnhancedEmailCleanupWorker) performEnhancedCleanup() {
	log.Printf("Starting enhanced email cleanup process...")
	startTime := time.Now()

	// Use the retention service to perform cleanup
	ctx := context.Background()
	err := w.retentionService.PerformCleanup(ctx)
	if err != nil {
		log.Printf("Error during enhanced cleanup: %v", err)
		return
	}

	// Update worker statistics
	w.lastCleanupTime = time.Now()
	w.cleanupStats = w.retentionService.GetCleanupStats()

	duration := time.Since(startTime)
	log.Printf("Enhanced cleanup completed in %v", duration)

	// Log cleanup statistics if notifications are enabled
	if w.enableNotifications {
		log.Printf("Cleanup Statistics: Total=%d, Expired=%d, Burned=%d, SelfDestructed=%d, Failed=%d, AuditLogsDeleted=%d",
			w.cleanupStats.TotalProcessed,
			w.cleanupStats.ExpiredDeleted,
			w.cleanupStats.BurnAfterReadDeleted,
			w.cleanupStats.SelfDestructedDeleted,
			w.cleanupStats.FailedDeletions,
			w.cleanupStats.AuditLogsDeleted,
		)
	}
}

// RunEnhancedCleanupOnce performs a single enhanced cleanup cycle
func (w *EnhancedEmailCleanupWorker) RunEnhancedCleanupOnce() error {
	log.Printf("Running one-time enhanced email cleanup...")
	w.performEnhancedCleanup()
	return nil
}

// GetEnhancedCleanupStats returns comprehensive statistics about the cleanup process
func (w *EnhancedEmailCleanupWorker) GetEnhancedCleanupStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get retention statistics
	ctx := context.Background()
	retentionStats, err := w.retentionService.GetRetentionStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get retention statistics: %w", err)
	}

	// Combine retention stats with worker stats
	stats["retention_statistics"] = retentionStats
	stats["worker_statistics"] = map[string]interface{}{
		"last_cleanup_time":    w.lastCleanupTime,
		"cleanup_interval":     w.cleanupInterval,
		"enable_notifications": w.enableNotifications,
		"cleanup_audit_logs":   w.cleanupAuditLogs,
		"cleanup_stats":        w.cleanupStats,
	}

	return stats, nil
}

// GetEmailsPendingCleanup returns emails pending cleanup with filtering
func (w *EnhancedEmailCleanupWorker) GetEmailsPendingCleanup(filters map[string]string, limit, offset int) ([]email.EmailRetentionInfo, error) {
	ctx := context.Background()
	return w.retentionService.GetEmailsPendingCleanup(ctx, filters, limit, offset)
}

// GetEmailsPendingCleanupCount returns the count of emails pending cleanup
func (w *EnhancedEmailCleanupWorker) GetEmailsPendingCleanupCount(filters map[string]string) (int, error) {
	ctx := context.Background()
	return w.retentionService.GetEmailsPendingCleanupCount(ctx, filters)
}

// SetEmailExpiration sets the expiration time for a specific email
func (w *EnhancedEmailCleanupWorker) SetEmailExpiration(emailID, senderID string, expiresAt *time.Time) error {
	return w.retentionService.SetEmailExpiration(emailID, senderID, expiresAt)
}

// GetDefaultExpirationTime returns the default expiration time for new emails
func (w *EnhancedEmailCleanupWorker) GetDefaultExpirationTime() time.Time {
	return w.retentionService.GetDefaultExpirationTime()
}

// RunEnhancedWorkerMain is the main function for standalone enhanced worker execution
func RunEnhancedWorkerMain() {
	log.Printf("Starting Enhanced Email Cleanup Worker...")

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

	// Create and start the enhanced worker
	worker, err := NewEnhancedEmailCleanupWorker(dbPath, cleanupInterval)
	if err != nil {
		log.Fatalf("Failed to create enhanced cleanup worker: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the worker
	worker.Start()

	// Wait for shutdown signal
	<-sigChan
	log.Printf("Received shutdown signal, stopping enhanced worker...")
	worker.Stop()
}
