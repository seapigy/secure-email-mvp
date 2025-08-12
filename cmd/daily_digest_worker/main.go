// =============================================================================
// SECURE EMAIL MVP - DAILY DIGEST WORKER
// =============================================================================
// This worker handles daily digest delivery scheduling and execution.
// Runs as a scheduled job to compile and send daily digest notifications.
// =============================================================================

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/notification"
	_ "modernc.org/sqlite"
)

// DailyDigestWorker handles daily digest delivery
type DailyDigestWorker struct {
	db                    *sql.DB
	notificationService   *notification.NotificationService
	checkIntervalMinutes  int
	digestDeliveryHour    int
	digestDeliveryMinute  int
}

// NewDailyDigestWorker creates a new daily digest worker
func NewDailyDigestWorker(dbPath string, checkIntervalMinutes, digestDeliveryHour, digestDeliveryMinute int) (*DailyDigestWorker, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	notificationService := notification.NewNotificationService(db)

	return &DailyDigestWorker{
		db:                   db,
		notificationService:  notificationService,
		checkIntervalMinutes: checkIntervalMinutes,
		digestDeliveryHour:   digestDeliveryHour,
		digestDeliveryMinute: digestDeliveryMinute,
	}, nil
}

// Start starts the daily digest worker
func (w *DailyDigestWorker) Start(ctx context.Context) {
	log.Printf("Starting daily digest worker (check interval: %d minutes, delivery time: %02d:%02d UTC)",
		w.checkIntervalMinutes, w.digestDeliveryHour, w.digestDeliveryMinute)

	ticker := time.NewTicker(time.Duration(w.checkIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	// Run initial check
	w.checkAndSendDigests(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Daily digest worker stopped")
			return
		case <-ticker.C:
			w.checkAndSendDigests(ctx)
		}
	}
}

// checkAndSendDigests checks if it's time to send digests and sends them
func (w *DailyDigestWorker) checkAndSendDigests(ctx context.Context) {
	now := time.Now().UTC()
	
	// Check if it's time to send digests (within the check interval)
	if w.isDigestDeliveryTime(now) {
		log.Printf("It's digest delivery time (%02d:%02d UTC), processing daily digests", 
			w.digestDeliveryHour, w.digestDeliveryMinute)
		
		// Calculate digest date (yesterday's date)
		digestDate := now.AddDate(0, 0, -1).Truncate(24 * time.Hour)
		
		if err := w.processDailyDigests(ctx, digestDate); err != nil {
			log.Printf("Failed to process daily digests: %v", err)
		}
	}
}

// isDigestDeliveryTime checks if it's time to send digests
func (w *DailyDigestWorker) isDigestDeliveryTime(now time.Time) bool {
	// Check if current time is within the delivery window
	deliveryTime := time.Date(now.Year(), now.Month(), now.Day(), 
		w.digestDeliveryHour, w.digestDeliveryMinute, 0, 0, time.UTC)
	
	// Allow a window of checkIntervalMinutes around the delivery time
	windowStart := deliveryTime.Add(-time.Duration(w.checkIntervalMinutes) * time.Minute)
	windowEnd := deliveryTime.Add(time.Duration(w.checkIntervalMinutes) * time.Minute)
	
	return now.After(windowStart) && now.Before(windowEnd)
}

// processDailyDigests processes daily digests for all eligible users
func (w *DailyDigestWorker) processDailyDigests(ctx context.Context, digestDate time.Time) error {
	log.Printf("Processing daily digests for %s", digestDate.Format("2006-01-02"))

	// Get all users with daily digest enabled
	userIDs, err := w.notificationService.GetUsersWithDailyDigestEnabled(ctx)
	if err != nil {
		return err
	}

	log.Printf("Found %d users with daily digest enabled", len(userIDs))

	// Process each user
	for _, userID := range userIDs {
		// Check if digest was already sent for this date
		if w.isDigestAlreadySent(ctx, userID, digestDate) {
			log.Printf("Digest already sent for user %s on %s", userID, digestDate.Format("2006-01-02"))
			continue
		}

		// Send digest for this user
		if err := w.notificationService.SendDailyDigest(ctx, userID, digestDate); err != nil {
			log.Printf("Failed to send digest for user %s: %v", userID, err)
			continue
		}

		log.Printf("Successfully sent digest for user %s", userID)
	}

	return nil
}

// isDigestAlreadySent checks if a digest was already sent for the given date
func (w *DailyDigestWorker) isDigestAlreadySent(ctx context.Context, userID string, digestDate time.Time) bool {
	query := `
		SELECT COUNT(*) FROM daily_digest_deliveries
		WHERE user_id = ? AND digest_date = ?
	`
	
	var count int
	err := w.db.QueryRowContext(ctx, query, userID, digestDate.Format("2006-01-02")).Scan(&count)
	if err != nil {
		log.Printf("Failed to check if digest already sent: %v", err)
		return false
	}
	
	return count > 0
}

// Close closes the worker and database connection
func (w *DailyDigestWorker) Close() error {
	return w.db.Close()
}

func main() {
	// Get configuration from environment variables
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "secure_email.db"
	}

	checkIntervalStr := os.Getenv("DIGEST_CHECK_INTERVAL_MINUTES")
	if checkIntervalStr == "" {
		checkIntervalStr = "15" // Default to checking every 15 minutes
	}
	checkInterval, err := strconv.Atoi(checkIntervalStr)
	if err != nil {
		log.Fatalf("Invalid DIGEST_CHECK_INTERVAL_MINUTES: %v", err)
	}

	deliveryHourStr := os.Getenv("DIGEST_DELIVERY_HOUR")
	if deliveryHourStr == "" {
		deliveryHourStr = "8" // Default to 8 AM UTC
	}
	deliveryHour, err := strconv.Atoi(deliveryHourStr)
	if err != nil {
		log.Fatalf("Invalid DIGEST_DELIVERY_HOUR: %v", err)
	}

	deliveryMinuteStr := os.Getenv("DIGEST_DELIVERY_MINUTE")
	if deliveryMinuteStr == "" {
		deliveryMinuteStr = "0" // Default to 0 minutes
	}
	deliveryMinute, err := strconv.Atoi(deliveryMinuteStr)
	if err != nil {
		log.Fatalf("Invalid DIGEST_DELIVERY_MINUTE: %v", err)
	}

	// Validate time values
	if deliveryHour < 0 || deliveryHour > 23 {
		log.Fatalf("Invalid DIGEST_DELIVERY_HOUR: must be 0-23")
	}
	if deliveryMinute < 0 || deliveryMinute > 59 {
		log.Fatalf("Invalid DIGEST_DELIVERY_MINUTE: must be 0-59")
	}

	// Create and start worker
	worker, err := NewDailyDigestWorker(dbPath, checkInterval, deliveryHour, deliveryMinute)
	if err != nil {
		log.Fatalf("Failed to create daily digest worker: %v", err)
	}
	defer worker.Close()

	// Create context for graceful shutdown
	ctx := context.Background()

	// Start the worker
	worker.Start(ctx)
}
