// =============================================================================
// SECURE EMAIL MVP - EXPIRATION ALERTS WORKER
// =============================================================================
// Standalone worker for processing expiration alerts and reminders.
// Micro-Iteration 4.19: Email Read Receipt & Expiration Alerts
// =============================================================================

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/readreceipts"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	log.Printf("Starting Expiration Alerts Worker...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize database connection
	dbPath := os.Getenv("SQLITE_DB")
	if dbPath == "" {
		dbPath = "/var/db/secure-email.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Test database connectivity
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Create expiration worker
	worker := readreceipts.NewExpirationWorker(db)

	// Get processing interval from environment
	intervalStr := os.Getenv("EXPIRATION_WORKER_INTERVAL_MINUTES")
	if intervalStr == "" {
		intervalStr = "15" // Default to 15 minutes
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("Invalid EXPIRATION_WORKER_INTERVAL_MINUTES, using default 15 minutes")
		interval = 15
	}

	log.Printf("Expiration worker started with interval: %d minutes", interval)

	// Process expiration alerts periodically
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	defer ticker.Stop()

	// Process once immediately on startup
	ctx := context.Background()
	if err := worker.ProcessExpirationAlerts(ctx); err != nil {
		log.Printf("Error processing expiration alerts: %v", err)
	}

	// Process on ticker
	for range ticker.C {
		ctx := context.Background()
		if err := worker.ProcessExpirationAlerts(ctx); err != nil {
			log.Printf("Error processing expiration alerts: %v", err)
		}
	}
}












