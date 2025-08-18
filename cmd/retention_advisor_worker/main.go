// =============================================================================
// SECURE EMAIL MVP - RETENTION ADVISOR WORKER
// =============================================================================
// Background worker for generating intelligent retention policy recommendations.
// Micro-Iteration 4.27: Intelligent Retention Insights & Proactive Policy Recommendations
// =============================================================================

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/email"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	log.Printf("Starting Retention Advisor Worker...")

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

	// Initialize services
	insightsService := email.NewRetentionInsightsService(db)
	advisorService := email.NewRetentionAdvisorService(db)

	// Get configuration from environment
	enableInsights := getEnvBool("ENABLE_RETENTION_INSIGHTS", true)
	enableRecommendations := getEnvBool("ENABLE_POLICY_RECOMMENDATIONS", true)
	insightsIntervalHours := getEnvInt("INSIGHTS_ROLLUP_INTERVAL_HOURS", 24)
	recommendationsIntervalHours := getEnvInt("RECOMMENDATION_GENERATION_INTERVAL_HOURS", 168) // Weekly

	log.Printf("Retention Advisor Worker Configuration:")
	log.Printf("- Enable Insights: %v", enableInsights)
	log.Printf("- Enable Recommendations: %v", enableRecommendations)
	log.Printf("- Insights Interval: %d hours", insightsIntervalHours)
	log.Printf("- Recommendations Interval: %d hours", recommendationsIntervalHours)

	// Create tickers for periodic tasks
	var insightsTicker, recommendationsTicker *time.Ticker
	if enableInsights {
		insightsTicker = time.NewTicker(time.Duration(insightsIntervalHours) * time.Hour)
		defer insightsTicker.Stop()
	}

	if enableRecommendations {
		recommendationsTicker = time.NewTicker(time.Duration(recommendationsIntervalHours) * time.Hour)
		defer recommendationsTicker.Stop()
	}

	// Run initial tasks
	log.Printf("Running initial retention analysis...")
	if enableInsights {
		if err := runInsightsGeneration(context.Background(), insightsService); err != nil {
			log.Printf("Failed to generate initial insights: %v", err)
		}
	}

	if enableRecommendations {
		if err := runRecommendationsGeneration(context.Background(), advisorService); err != nil {
			log.Printf("Failed to generate initial recommendations: %v", err)
		}
	}

	// Main loop for periodic tasks
	for {
		select {
		case <-insightsTicker.C:
			if enableInsights {
				log.Printf("Running periodic insights generation...")
				if err := runInsightsGeneration(context.Background(), insightsService); err != nil {
					log.Printf("Failed to generate insights: %v", err)
				}
			}

		case <-recommendationsTicker.C:
			if enableRecommendations {
				log.Printf("Running periodic recommendations generation...")
				if err := runRecommendationsGeneration(context.Background(), advisorService); err != nil {
					log.Printf("Failed to generate recommendations: %v", err)
				}
			}
		}
	}
}

// runInsightsGeneration runs the insights generation process
func runInsightsGeneration(ctx context.Context, insightsService *email.RetentionInsightsService) error {
	startTime := time.Now()
	log.Printf("Starting insights generation at %s", startTime.Format("2006-01-02 15:04:05"))

	// Generate insights for the last 7 days to catch up on any missed days
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -i)
		if err := insightsService.GenerateDailyInsights(ctx, date); err != nil {
			log.Printf("Failed to generate insights for %s: %v", date.Format("2006-01-02"), err)
			continue
		}
		log.Printf("Generated insights for %s", date.Format("2006-01-02"))
	}

	duration := time.Since(startTime)
	log.Printf("Insights generation completed in %v", duration)
	return nil
}

// runRecommendationsGeneration runs the recommendations generation process
func runRecommendationsGeneration(ctx context.Context, advisorService *email.RetentionAdvisorService) error {
	startTime := time.Now()
	log.Printf("Starting recommendations generation at %s", startTime.Format("2006-01-02 15:04:05"))

	// Generate recommendations
	if err := advisorService.GenerateRecommendations(ctx); err != nil {
		return err
	}

	duration := time.Since(startTime)
	log.Printf("Recommendations generation completed in %v", duration)
	return nil
}

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}



