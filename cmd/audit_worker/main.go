// =============================================================================
// SECURE EMAIL MVP - AUDIT LOG WORKER
// =============================================================================
// This worker handles periodic audit log cleanup and maintenance tasks.
// =============================================================================

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/audit"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	log.Printf("Starting Audit Log Worker...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database path
	dbPath := os.Getenv("SQLITE_DB")
	if dbPath == "" {
		dbPath = "/var/db/secure-email.db"
	}

	// Connect to database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Initialize audit services
	auditService := audit.NewAuditService(db)
	exportService := audit.NewExportService(db, "")

	// Get cleanup interval from environment
	cleanupIntervalStr := os.Getenv("AUDIT_CLEANUP_INTERVAL_MINUTES")
	if cleanupIntervalStr == "" {
		cleanupIntervalStr = "60" // Default to 60 minutes
	}
	cleanupInterval, err := strconv.Atoi(cleanupIntervalStr)
	if err != nil {
		log.Printf("Invalid AUDIT_CLEANUP_INTERVAL_MINUTES, using default 60 minutes")
		cleanupInterval = 60
	}

	log.Printf("Audit worker started with cleanup interval: %d minutes", cleanupInterval)

	// Create ticker for periodic cleanup
	ticker := time.NewTicker(time.Duration(cleanupInterval) * time.Minute)
	defer ticker.Stop()

	// Run initial cleanup
	log.Printf("Running initial audit cleanup...")
	runAuditCleanup(context.Background(), auditService, exportService)

	// Run periodic cleanup
	for {
		select {
		case <-ticker.C:
			log.Printf("Running periodic audit cleanup...")
			runAuditCleanup(context.Background(), auditService, exportService)
		}
	}
}

// runAuditCleanup performs all audit cleanup tasks
func runAuditCleanup(ctx context.Context, auditService *audit.AuditService, exportService *audit.ExportService) {
	// Purge expired audit logs
	if err := auditService.PurgeExpiredLogs(ctx); err != nil {
		log.Printf("Failed to purge expired audit logs: %v", err)
	} else {
		log.Printf("Successfully purged expired audit logs")
	}

	// Cleanup expired exports
	if err := exportService.CleanupExpiredExports(ctx); err != nil {
		log.Printf("Failed to cleanup expired exports: %v", err)
	} else {
		log.Printf("Successfully cleaned up expired exports")
	}
}

