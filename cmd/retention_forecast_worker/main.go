package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"secure-email-mvp/pkg/email"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// RetentionForecastWorker runs predictive forecasting for retention operations
type RetentionForecastWorker struct {
	db                    *sql.DB
	forecastService       *email.RetentionForecastService
	anomalyDetector       *email.RetentionAnomalyDetector
	forecastIntervalHours int
	anomalyIntervalHours  int
	ctx                   context.Context
	cancel                context.CancelFunc
}

// NewRetentionForecastWorker creates a new retention forecast worker
func NewRetentionForecastWorker(db *sql.DB, forecastConfig *email.ForecastConfig, anomalyConfig *email.AnomalyConfig) *RetentionForecastWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &RetentionForecastWorker{
		db:                    db,
		forecastService:       email.NewRetentionForecastService(db, forecastConfig),
		anomalyDetector:       email.NewRetentionAnomalyDetector(db, anomalyConfig),
		forecastIntervalHours: 24, // Generate forecasts daily
		anomalyIntervalHours:  6,  // Check for anomalies every 6 hours
		ctx:                   ctx,
		cancel:                cancel,
	}
}

// Start begins the worker process
func (w *RetentionForecastWorker) Start() error {
	log.Println("Starting Retention Forecast Worker...")

	// Start forecast generation goroutine
	go w.runForecastGeneration()

	// Start anomaly detection goroutine
	go w.runAnomalyDetection()

	// Start forecast accuracy evaluation goroutine
	go w.runForecastAccuracyEvaluation()

	log.Println("Retention Forecast Worker started successfully")
	return nil
}

// Stop gracefully stops the worker
func (w *RetentionForecastWorker) Stop() {
	log.Println("Stopping Retention Forecast Worker...")
	w.cancel()
}

// runForecastGeneration runs forecast generation periodically
func (w *RetentionForecastWorker) runForecastGeneration() {
	ticker := time.NewTicker(time.Duration(w.forecastIntervalHours) * time.Hour)
	defer ticker.Stop()

	// Generate initial forecasts
	if err := w.generateForecasts(); err != nil {
		log.Printf("Failed to generate initial forecasts: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := w.generateForecasts(); err != nil {
				log.Printf("Failed to generate forecasts: %v", err)
			}
		case <-w.ctx.Done():
			log.Println("Forecast generation stopped")
			return
		}
	}
}

// runAnomalyDetection runs anomaly detection periodically
func (w *RetentionForecastWorker) runAnomalyDetection() {
	ticker := time.NewTicker(time.Duration(w.anomalyIntervalHours) * time.Hour)
	defer ticker.Stop()

	// Run initial anomaly detection
	if err := w.detectAnomalies(); err != nil {
		log.Printf("Failed to run initial anomaly detection: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := w.detectAnomalies(); err != nil {
				log.Printf("Failed to detect anomalies: %v", err)
			}
		case <-w.ctx.Done():
			log.Println("Anomaly detection stopped")
			return
		}
	}
}

// runForecastAccuracyEvaluation runs forecast accuracy evaluation periodically
func (w *RetentionForecastWorker) runForecastAccuracyEvaluation() {
	ticker := time.NewTicker(12 * time.Hour) // Evaluate accuracy every 12 hours
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.evaluateForecastAccuracy(); err != nil {
				log.Printf("Failed to evaluate forecast accuracy: %v", err)
			}
		case <-w.ctx.Done():
			log.Println("Forecast accuracy evaluation stopped")
			return
		}
	}
}

// generateForecasts generates retention forecasts
func (w *RetentionForecastWorker) generateForecasts() error {
	log.Println("Generating retention forecasts...")
	startTime := time.Now()

	if err := w.forecastService.GenerateForecasts(w.ctx); err != nil {
		return err
	}

	duration := time.Since(startTime)
	log.Printf("Retention forecasts generated successfully in %v", duration)
	return nil
}

// detectAnomalies runs anomaly detection
func (w *RetentionForecastWorker) detectAnomalies() error {
	log.Println("Running retention anomaly detection...")
	startTime := time.Now()

	if err := w.anomalyDetector.DetectAnomalies(w.ctx); err != nil {
		return err
	}

	duration := time.Since(startTime)
	log.Printf("Retention anomaly detection completed in %v", duration)
	return nil
}

// evaluateForecastAccuracy evaluates the accuracy of recent forecasts
func (w *RetentionForecastWorker) evaluateForecastAccuracy() error {
	log.Println("Evaluating forecast accuracy...")
	startTime := time.Now()

	// Get recent forecasts that have passed their target period
	query := `
		SELECT id FROM retention_forecasts
		WHERE target_period_end <= datetime('now')
		  AND generated_at >= datetime('now', '-7 days')
		ORDER BY target_period_end DESC
		LIMIT 50
	`

	rows, err := w.db.QueryContext(w.ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	evaluatedCount := 0
	for rows.Next() {
		var forecastID int64
		if err := rows.Scan(&forecastID); err != nil {
			log.Printf("Failed to scan forecast ID: %v", err)
			continue
		}

		if err := w.forecastService.EvaluateForecastAccuracy(w.ctx, forecastID); err != nil {
			log.Printf("Failed to evaluate forecast %d: %v", forecastID, err)
			continue
		}

		evaluatedCount++
	}

	duration := time.Since(startTime)
	log.Printf("Forecast accuracy evaluation completed: %d forecasts evaluated in %v", evaluatedCount, duration)
	return nil
}

func main() {
	// Parse command line flags
	var (
		dbPath              = flag.String("db", "/var/db/secure-email.db", "Database path")
		forecastInterval    = flag.Int("forecast-interval", 24, "Forecast generation interval in hours")
		anomalyInterval     = flag.Int("anomaly-interval", 6, "Anomaly detection interval in hours")
		spikeThreshold      = flag.Float64("spike-threshold", 200.0, "Spike deletion threshold percentage")
		dropThreshold       = flag.Float64("drop-threshold", 50.0, "Drop policy matches threshold percentage")
		forecastThreshold   = flag.Float64("forecast-threshold", 25.0, "Forecast deviation threshold percentage")
		archivalThreshold   = flag.Float64("archival-threshold", 150.0, "Unusual archival threshold percentage")
		detectionWindow     = flag.Int("detection-window", 24, "Anomaly detection window in hours")
		autoResolution      = flag.Bool("auto-resolution", false, "Enable automatic anomaly resolution")
		confidenceThreshold = flag.Float64("confidence-threshold", 0.8, "Minimum confidence threshold for forecasts")
		costPerGBPerMonth   = flag.Float64("cost-per-gb", 0.02, "Cost per GB per month for cost calculations")
	)
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Connect to database
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create forecast configuration
	forecastConfig := &email.ForecastConfig{
		PeriodsDays:         []int{7, 30, 90},
		ConfidenceThreshold: *confidenceThreshold,
		MinDataPoints:       10,
		MaxDataAgeHours:     168, // 7 days
		ModelVersion:        "v1.0",
		CostPerGBPerMonth:   *costPerGBPerMonth,
	}

	// Create anomaly detection configuration
	anomalyConfig := &email.AnomalyConfig{
		SpikeDeletionThreshold:     *spikeThreshold,
		DropPolicyMatchesThreshold: *dropThreshold,
		ForecastDeviationThreshold: *forecastThreshold,
		UnusualArchivalThreshold:   *archivalThreshold,
		DetectionWindowHours:       *detectionWindow,
		AutoResolutionEnabled:      *autoResolution,
		MinConfidenceThreshold:     *confidenceThreshold,
	}

	// Create and start worker
	worker := NewRetentionForecastWorker(db, forecastConfig, anomalyConfig)
	worker.forecastIntervalHours = *forecastInterval
	worker.anomalyIntervalHours = *anomalyInterval

	if err := worker.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Retention Forecast Worker running. Press Ctrl+C to stop.")
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down Retention Forecast Worker...")
	worker.Stop()

	// Wait a bit for goroutines to finish
	time.Sleep(2 * time.Second)
	log.Println("Retention Forecast Worker stopped")
}
