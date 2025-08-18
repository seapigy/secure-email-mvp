package email

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
)

// RetentionForecast represents a retention forecast record
type RetentionForecast struct {
	ID                           int64     `json:"id"`
	ForecastType                 string    `json:"forecast_type"`
	ForecastKey                  string    `json:"forecast_key"`
	GeneratedAt                  time.Time `json:"generated_at"`
	TargetPeriodEnd              time.Time `json:"target_period_end"`
	PredictedUsageBytes          int64     `json:"predicted_usage_bytes"`
	PredictedArchivalCount       int       `json:"predicted_archival_count"`
	PredictedDeletionCount       int       `json:"predicted_deletion_count"`
	PredictedPolicyImpact        float64   `json:"predicted_policy_impact"`
	PredictedStorageSavingsBytes int64     `json:"predicted_storage_savings_bytes"`
	PredictedCostSavingsUSD      float64   `json:"predicted_cost_savings_usd"`
	ConfidenceScore              float64   `json:"confidence_score"`
	AccuracyScore                float64   `json:"accuracy_score"`
	ForecastModelVersion         string    `json:"forecast_model_version"`
	HistoricalDataPoints         int       `json:"historical_data_points"`
	DataFreshnessHours           int       `json:"data_freshness_hours"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

// ForecastMetrics represents the metrics used for forecasting
type ForecastMetrics struct {
	UsageBytes     int64   `json:"usage_bytes"`
	ArchivalCount  int     `json:"archival_count"`
	DeletionCount  int     `json:"deletion_count"`
	PolicyImpact   float64 `json:"policy_impact"`
	StorageSavings int64   `json:"storage_savings"`
	CostSavings    float64 `json:"cost_savings"`
	DataPoints     int     `json:"data_points"`
	TimeWindow     int     `json:"time_window_hours"`
}

// ForecastConfig represents configuration for forecasting
type ForecastConfig struct {
	PeriodsDays         []int   `json:"periods_days"`
	ConfidenceThreshold float64 `json:"confidence_threshold"`
	MinDataPoints       int     `json:"min_data_points"`
	MaxDataAgeHours     int     `json:"max_data_age_hours"`
	ModelVersion        string  `json:"model_version"`
	CostPerGBPerMonth   float64 `json:"cost_per_gb_per_month"`
}

// RetentionForecastService provides predictive forecasting for retention operations
type RetentionForecastService struct {
	db     *sql.DB
	config *ForecastConfig
}

// NewRetentionForecastService creates a new retention forecast service
func NewRetentionForecastService(db *sql.DB, config *ForecastConfig) *RetentionForecastService {
	if config == nil {
		config = &ForecastConfig{
			PeriodsDays:         []int{7, 30, 90},
			ConfidenceThreshold: 0.8,
			MinDataPoints:       10,
			MaxDataAgeHours:     168, // 7 days
			ModelVersion:        "v1.0",
			CostPerGBPerMonth:   0.02, // $0.02 per GB per month
		}
	}

	return &RetentionForecastService{
		db:     db,
		config: config,
	}
}

// GenerateForecasts generates forecasts for all configured periods and scopes
func (rfs *RetentionForecastService) GenerateForecasts(ctx context.Context) error {
	log.Println("Starting retention forecast generation...")

	// Generate global forecasts
	if err := rfs.generateGlobalForecasts(ctx); err != nil {
		log.Printf("Failed to generate global forecasts: %v", err)
	}

	// Generate user-specific forecasts
	if err := rfs.generateUserForecasts(ctx); err != nil {
		log.Printf("Failed to generate user forecasts: %v", err)
	}

	// Generate domain-specific forecasts
	if err := rfs.generateDomainForecasts(ctx); err != nil {
		log.Printf("Failed to generate domain forecasts: %v", err)
	}

	log.Println("Retention forecast generation completed")
	return nil
}

// generateGlobalForecasts generates forecasts for the global scope
func (rfs *RetentionForecastService) generateGlobalForecasts(ctx context.Context) error {
	for _, periodDays := range rfs.config.PeriodsDays {
		forecast, err := rfs.generateForecast(ctx, "global", "global", periodDays)
		if err != nil {
			log.Printf("Failed to generate global forecast for %d days: %v", periodDays, err)
			continue
		}

		if err := rfs.storeForecast(ctx, forecast); err != nil {
			log.Printf("Failed to store global forecast: %v", err)
		}
	}
	return nil
}

// generateUserForecasts generates forecasts for individual users
func (rfs *RetentionForecastService) generateUserForecasts(ctx context.Context) error {
	users, err := rfs.getActiveUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active users: %w", err)
	}

	for _, userID := range users {
		for _, periodDays := range rfs.config.PeriodsDays {
			forecast, err := rfs.generateForecast(ctx, "user", userID, periodDays)
			if err != nil {
				log.Printf("Failed to generate forecast for user %s (%d days): %v", userID, periodDays, err)
				continue
			}

			if err := rfs.storeForecast(ctx, forecast); err != nil {
				log.Printf("Failed to store forecast for user %s: %v", userID, err)
			}
		}
	}
	return nil
}

// generateDomainForecasts generates forecasts for domains
func (rfs *RetentionForecastService) generateDomainForecasts(ctx context.Context) error {
	domains, err := rfs.getActiveDomains(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active domains: %w", err)
	}

	for _, domain := range domains {
		for _, periodDays := range rfs.config.PeriodsDays {
			forecast, err := rfs.generateForecast(ctx, "domain", domain, periodDays)
			if err != nil {
				log.Printf("Failed to generate forecast for domain %s (%d days): %v", domain, periodDays, err)
				continue
			}

			if err := rfs.storeForecast(ctx, forecast); err != nil {
				log.Printf("Failed to store forecast for domain %s: %v", domain, err)
			}
		}
	}
	return nil
}

// generateForecast generates a single forecast for a specific scope and period
func (rfs *RetentionForecastService) generateForecast(ctx context.Context, forecastType, forecastKey string, periodDays int) (*RetentionForecast, error) {
	// Get historical metrics for the scope
	metrics, err := rfs.getHistoricalMetrics(ctx, forecastType, forecastKey, periodDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical metrics: %w", err)
	}

	// Check if we have enough data points
	if metrics.DataPoints < rfs.config.MinDataPoints {
		return nil, fmt.Errorf("insufficient data points: %d < %d", metrics.DataPoints, rfs.config.MinDataPoints)
	}

	// Calculate predictions using simple linear regression
	predictions := rfs.calculatePredictions(metrics, periodDays)

	// Calculate confidence score based on data quality and consistency
	confidenceScore := rfs.calculateConfidenceScore(metrics)

	// Calculate accuracy score based on historical accuracy
	accuracyScore := rfs.calculateAccuracyScore(ctx, forecastType, forecastKey)

	// Create forecast
	forecast := &RetentionForecast{
		ForecastType:                 forecastType,
		ForecastKey:                  forecastKey,
		GeneratedAt:                  time.Now(),
		TargetPeriodEnd:              time.Now().AddDate(0, 0, periodDays),
		PredictedUsageBytes:          predictions.UsageBytes,
		PredictedArchivalCount:       predictions.ArchivalCount,
		PredictedDeletionCount:       predictions.DeletionCount,
		PredictedPolicyImpact:        predictions.PolicyImpact,
		PredictedStorageSavingsBytes: predictions.StorageSavings,
		PredictedCostSavingsUSD:      predictions.CostSavings,
		ConfidenceScore:              confidenceScore,
		AccuracyScore:                accuracyScore,
		ForecastModelVersion:         rfs.config.ModelVersion,
		HistoricalDataPoints:         metrics.DataPoints,
		DataFreshnessHours:           metrics.TimeWindow,
	}

	return forecast, nil
}

// getHistoricalMetrics retrieves historical metrics for forecasting
func (rfs *RetentionForecastService) getHistoricalMetrics(ctx context.Context, forecastType, forecastKey string, periodDays int) (*ForecastMetrics, error) {
	// Calculate the time window for historical data (3x the forecast period)
	historicalDays := periodDays * 3
	startDate := time.Now().AddDate(0, 0, -historicalDays)

	var query string
	var args []interface{}

	switch forecastType {
	case "global":
		query = `
			SELECT 
				SUM(CASE WHEN e.encrypted_blob_url IS NOT NULL THEN e.encrypted_blob_url ELSE 0 END) as usage_bytes,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				AVG(CASE WHEN pel.impact_score IS NOT NULL THEN pel.impact_score ELSE 0 END) as policy_impact,
				SUM(CASE WHEN ae.compressed_size IS NOT NULL THEN ae.original_size - ae.compressed_size ELSE 0 END) as storage_savings,
				COUNT(*) as data_points,
				? as time_window_hours
			FROM emails e
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id
			WHERE e.created_at >= ?
		`
		args = []interface{}{historicalDays * 24, startDate}

	case "user":
		query = `
			SELECT 
				SUM(CASE WHEN e.encrypted_blob_url IS NOT NULL THEN e.encrypted_blob_url ELSE 0 END) as usage_bytes,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				AVG(CASE WHEN pel.impact_score IS NOT NULL THEN pel.impact_score ELSE 0 END) as policy_impact,
				SUM(CASE WHEN ae.compressed_size IS NOT NULL THEN ae.original_size - ae.compressed_size ELSE 0 END) as storage_savings,
				COUNT(*) as data_points,
				? as time_window_hours
			FROM emails e
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id
			WHERE e.sender_id = ? AND e.created_at >= ?
		`
		args = []interface{}{historicalDays * 24, forecastKey, startDate}

	case "domain":
		query = `
			SELECT 
				SUM(CASE WHEN e.encrypted_blob_url IS NOT NULL THEN e.encrypted_blob_url ELSE 0 END) as usage_bytes,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				AVG(CASE WHEN pel.impact_score IS NOT NULL THEN pel.impact_score ELSE 0 END) as policy_impact,
				SUM(CASE WHEN ae.compressed_size IS NOT NULL THEN ae.original_size - ae.compressed_size ELSE 0 END) as storage_savings,
				COUNT(*) as data_points,
				? as time_window_hours
			FROM emails e
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id
			WHERE (e.recipient LIKE ? OR e.recipient LIKE ?) AND e.created_at >= ?
		`
		domainPattern := "%@" + forecastKey
		args = []interface{}{historicalDays * 24, domainPattern, domainPattern, startDate}

	default:
		return nil, fmt.Errorf("unsupported forecast type: %s", forecastType)
	}

	var metrics ForecastMetrics
	err := rfs.db.QueryRowContext(ctx, query, args...).Scan(
		&metrics.UsageBytes,
		&metrics.ArchivalCount,
		&metrics.DeletionCount,
		&metrics.PolicyImpact,
		&metrics.StorageSavings,
		&metrics.DataPoints,
		&metrics.TimeWindow,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query historical metrics: %w", err)
	}

	// Calculate cost savings
	metrics.CostSavings = float64(metrics.StorageSavings) * rfs.config.CostPerGBPerMonth / (1024 * 1024 * 1024)

	return &metrics, nil
}

// calculatePredictions calculates predictions using simple linear regression
func (rfs *RetentionForecastService) calculatePredictions(metrics *ForecastMetrics, periodDays int) *ForecastMetrics {
	// Simple linear projection based on historical averages
	dailyUsage := float64(metrics.UsageBytes) / float64(metrics.TimeWindow/24)
	dailyArchival := float64(metrics.ArchivalCount) / float64(metrics.TimeWindow/24)
	dailyDeletion := float64(metrics.DeletionCount) / float64(metrics.TimeWindow/24)
	dailyStorageSavings := float64(metrics.StorageSavings) / float64(metrics.TimeWindow/24)

	predictions := &ForecastMetrics{
		UsageBytes:     int64(dailyUsage * float64(periodDays)),
		ArchivalCount:  int(dailyArchival * float64(periodDays)),
		DeletionCount:  int(dailyDeletion * float64(periodDays)),
		PolicyImpact:   metrics.PolicyImpact, // Keep historical average
		StorageSavings: int64(dailyStorageSavings * float64(periodDays)),
		CostSavings:    float64(dailyStorageSavings*float64(periodDays)) * rfs.config.CostPerGBPerMonth / (1024 * 1024 * 1024),
	}

	return predictions
}

// calculateConfidenceScore calculates confidence score based on data quality
func (rfs *RetentionForecastService) calculateConfidenceScore(metrics *ForecastMetrics) float64 {
	// Base confidence on data points and freshness
	dataQualityScore := math.Min(float64(metrics.DataPoints)/100.0, 1.0)
	freshnessScore := math.Min(float64(metrics.TimeWindow)/168.0, 1.0) // 168 hours = 7 days

	// Weighted average
	confidence := (dataQualityScore*0.7 + freshnessScore*0.3)
	return math.Min(confidence, 1.0)
}

// calculateAccuracyScore calculates accuracy score based on historical accuracy
func (rfs *RetentionForecastService) calculateAccuracyScore(ctx context.Context, forecastType, forecastKey string) float64 {
	// Query historical forecast accuracy
	query := `
		SELECT AVG(fal.overall_accuracy_score)
		FROM forecast_accuracy_logs fal
		JOIN retention_forecasts rf ON fal.forecast_id = rf.id
		WHERE rf.forecast_type = ? AND rf.forecast_key = ?
		AND fal.evaluated_at >= datetime('now', '-30 days')
	`

	var accuracy float64
	err := rfs.db.QueryRowContext(ctx, query, forecastType, forecastKey).Scan(&accuracy)
	if err != nil {
		// Default accuracy score for new forecasts
		return 0.5
	}

	return accuracy
}

// storeForecast stores a forecast in the database
func (rfs *RetentionForecastService) storeForecast(ctx context.Context, forecast *RetentionForecast) error {
	query := `
		INSERT OR REPLACE INTO retention_forecasts (
			forecast_type, forecast_key, generated_at, target_period_end,
			predicted_usage_bytes, predicted_archival_count, predicted_deletion_count,
			predicted_policy_impact, predicted_storage_savings_bytes, predicted_cost_savings_usd,
			confidence_score, accuracy_score, forecast_model_version,
			historical_data_points, data_freshness_hours, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := rfs.db.ExecContext(ctx, query,
		forecast.ForecastType,
		forecast.ForecastKey,
		forecast.GeneratedAt,
		forecast.TargetPeriodEnd,
		forecast.PredictedUsageBytes,
		forecast.PredictedArchivalCount,
		forecast.PredictedDeletionCount,
		forecast.PredictedPolicyImpact,
		forecast.PredictedStorageSavingsBytes,
		forecast.PredictedCostSavingsUSD,
		forecast.ConfidenceScore,
		forecast.AccuracyScore,
		forecast.ForecastModelVersion,
		forecast.HistoricalDataPoints,
		forecast.DataFreshnessHours,
		forecast.CreatedAt,
		forecast.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store forecast: %w", err)
	}

	return nil
}

// GetForecasts retrieves forecasts for a specific scope
func (rfs *RetentionForecastService) GetForecasts(ctx context.Context, forecastType, forecastKey string, limit int) ([]*RetentionForecast, error) {
	query := `
		SELECT id, forecast_type, forecast_key, generated_at, target_period_end,
			   predicted_usage_bytes, predicted_archival_count, predicted_deletion_count,
			   predicted_policy_impact, predicted_storage_savings_bytes, predicted_cost_savings_usd,
			   confidence_score, accuracy_score, forecast_model_version,
			   historical_data_points, data_freshness_hours, created_at, updated_at
		FROM retention_forecasts
		WHERE forecast_type = ? AND forecast_key = ?
		ORDER BY generated_at DESC
		LIMIT ?
	`

	rows, err := rfs.db.QueryContext(ctx, query, forecastType, forecastKey, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query forecasts: %w", err)
	}
	defer rows.Close()

	var forecasts []*RetentionForecast
	for rows.Next() {
		var forecast RetentionForecast
		err := rows.Scan(
			&forecast.ID,
			&forecast.ForecastType,
			&forecast.ForecastKey,
			&forecast.GeneratedAt,
			&forecast.TargetPeriodEnd,
			&forecast.PredictedUsageBytes,
			&forecast.PredictedArchivalCount,
			&forecast.PredictedDeletionCount,
			&forecast.PredictedPolicyImpact,
			&forecast.PredictedStorageSavingsBytes,
			&forecast.PredictedCostSavingsUSD,
			&forecast.ConfidenceScore,
			&forecast.AccuracyScore,
			&forecast.ForecastModelVersion,
			&forecast.HistoricalDataPoints,
			&forecast.DataFreshnessHours,
			&forecast.CreatedAt,
			&forecast.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan forecast: %w", err)
		}
		forecasts = append(forecasts, &forecast)
	}

	return forecasts, nil
}

// getActiveUsers retrieves list of active users for forecasting
func (rfs *RetentionForecastService) getActiveUsers(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT sender_id
		FROM emails
		WHERE created_at >= datetime('now', '-30 days')
		GROUP BY sender_id
		HAVING COUNT(*) >= 5
	`

	rows, err := rfs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		users = append(users, userID)
	}

	return users, nil
}

// getActiveDomains retrieves list of active domains for forecasting
func (rfs *RetentionForecastService) getActiveDomains(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT 
			CASE 
				WHEN recipient LIKE '%@%' THEN SUBSTR(recipient, INSTR(recipient, '@') + 1)
				ELSE NULL
			END as domain
		FROM emails
		WHERE created_at >= datetime('now', '-30 days')
		  AND recipient LIKE '%@%'
		GROUP BY domain
		HAVING COUNT(*) >= 10
	`

	rows, err := rfs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active domains: %w", err)
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domain sql.NullString
		if err := rows.Scan(&domain); err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		if domain.Valid {
			domains = append(domains, domain.String)
		}
	}

	return domains, nil
}

// EvaluateForecastAccuracy evaluates the accuracy of a forecast against actual data
func (rfs *RetentionForecastService) EvaluateForecastAccuracy(ctx context.Context, forecastID int64) error {
	// Get the forecast
	forecast, err := rfs.getForecastByID(ctx, forecastID)
	if err != nil {
		return fmt.Errorf("failed to get forecast: %w", err)
	}

	// Get actual metrics for the forecast period
	actualMetrics, err := rfs.getActualMetrics(ctx, forecast)
	if err != nil {
		return fmt.Errorf("failed to get actual metrics: %w", err)
	}

	// Calculate accuracy percentages
	usageAccuracy := rfs.calculateAccuracyPercentage(float64(forecast.PredictedUsageBytes), float64(actualMetrics.UsageBytes))
	archivalAccuracy := rfs.calculateAccuracyPercentage(float64(forecast.PredictedArchivalCount), float64(actualMetrics.ArchivalCount))
	deletionAccuracy := rfs.calculateAccuracyPercentage(float64(forecast.PredictedDeletionCount), float64(actualMetrics.DeletionCount))
	policyImpactAccuracy := rfs.calculateAccuracyPercentage(forecast.PredictedPolicyImpact, actualMetrics.PolicyImpact)

	// Calculate overall accuracy score
	overallAccuracy := (usageAccuracy + archivalAccuracy + deletionAccuracy + policyImpactAccuracy) / 4.0

	// Store accuracy log
	accuracyLog := &ForecastAccuracyLog{
		ForecastID:                     forecastID,
		ActualUsageBytes:               actualMetrics.UsageBytes,
		ActualArchivalCount:            actualMetrics.ArchivalCount,
		ActualDeletionCount:            actualMetrics.DeletionCount,
		ActualPolicyImpact:             actualMetrics.PolicyImpact,
		ActualStorageSavingsBytes:      actualMetrics.StorageSavings,
		ActualCostSavingsUSD:           actualMetrics.CostSavings,
		UsageAccuracyPercentage:        usageAccuracy,
		ArchivalAccuracyPercentage:     archivalAccuracy,
		DeletionAccuracyPercentage:     deletionAccuracy,
		PolicyImpactAccuracyPercentage: policyImpactAccuracy,
		OverallAccuracyScore:           overallAccuracy,
		EvaluatedAt:                    time.Now(),
		EvaluationWindowHours:          24,
	}

	return rfs.storeForecastAccuracyLog(ctx, accuracyLog)
}

// getForecastByID retrieves a forecast by ID
func (rfs *RetentionForecastService) getForecastByID(ctx context.Context, forecastID int64) (*RetentionForecast, error) {
	query := `
		SELECT id, forecast_type, forecast_key, generated_at, target_period_end,
			   predicted_usage_bytes, predicted_archival_count, predicted_deletion_count,
			   predicted_policy_impact, predicted_storage_savings_bytes, predicted_cost_savings_usd,
			   confidence_score, accuracy_score, forecast_model_version,
			   historical_data_points, data_freshness_hours, created_at, updated_at
		FROM retention_forecasts
		WHERE id = ?
	`

	var forecast RetentionForecast
	err := rfs.db.QueryRowContext(ctx, query, forecastID).Scan(
		&forecast.ID,
		&forecast.ForecastType,
		&forecast.ForecastKey,
		&forecast.GeneratedAt,
		&forecast.TargetPeriodEnd,
		&forecast.PredictedUsageBytes,
		&forecast.PredictedArchivalCount,
		&forecast.PredictedDeletionCount,
		&forecast.PredictedPolicyImpact,
		&forecast.PredictedStorageSavingsBytes,
		&forecast.PredictedCostSavingsUSD,
		&forecast.ConfidenceScore,
		&forecast.AccuracyScore,
		&forecast.ForecastModelVersion,
		&forecast.HistoricalDataPoints,
		&forecast.DataFreshnessHours,
		&forecast.CreatedAt,
		&forecast.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get forecast: %w", err)
	}

	return &forecast, nil
}

// getActualMetrics retrieves actual metrics for a forecast period
func (rfs *RetentionForecastService) getActualMetrics(ctx context.Context, forecast *RetentionForecast) (*ForecastMetrics, error) {
	// Calculate the actual period based on forecast generation time
	periodStart := forecast.GeneratedAt
	periodEnd := forecast.TargetPeriodEnd

	var query string
	var args []interface{}

	switch forecast.ForecastType {
	case "global":
		query = `
			SELECT 
				SUM(CASE WHEN e.encrypted_blob_url IS NOT NULL THEN e.encrypted_blob_url ELSE 0 END) as usage_bytes,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				AVG(CASE WHEN pel.impact_score IS NOT NULL THEN pel.impact_score ELSE 0 END) as policy_impact,
				SUM(CASE WHEN ae.compressed_size IS NOT NULL THEN ae.original_size - ae.compressed_size ELSE 0 END) as storage_savings
			FROM emails e
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id
			WHERE e.created_at >= ? AND e.created_at <= ?
		`
		args = []interface{}{periodStart, periodEnd}

	case "user":
		query = `
			SELECT 
				SUM(CASE WHEN e.encrypted_blob_url IS NOT NULL THEN e.encrypted_blob_url ELSE 0 END) as usage_bytes,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				AVG(CASE WHEN pel.impact_score IS NOT NULL THEN pel.impact_score ELSE 0 END) as policy_impact,
				SUM(CASE WHEN ae.compressed_size IS NOT NULL THEN ae.original_size - ae.compressed_size ELSE 0 END) as storage_savings
			FROM emails e
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id
			WHERE e.sender_id = ? AND e.created_at >= ? AND e.created_at <= ?
		`
		args = []interface{}{forecast.ForecastKey, periodStart, periodEnd}

	case "domain":
		query = `
			SELECT 
				SUM(CASE WHEN e.encrypted_blob_url IS NOT NULL THEN e.encrypted_blob_url ELSE 0 END) as usage_bytes,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				AVG(CASE WHEN pel.impact_score IS NOT NULL THEN pel.impact_score ELSE 0 END) as policy_impact,
				SUM(CASE WHEN ae.compressed_size IS NOT NULL THEN ae.original_size - ae.compressed_size ELSE 0 END) as storage_savings
			FROM emails e
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id
			WHERE (e.recipient LIKE ? OR e.recipient LIKE ?) AND e.created_at >= ? AND e.created_at <= ?
		`
		domainPattern := "%@" + forecast.ForecastKey
		args = []interface{}{domainPattern, domainPattern, periodStart, periodEnd}

	default:
		return nil, fmt.Errorf("unsupported forecast type: %s", forecast.ForecastType)
	}

	var metrics ForecastMetrics
	err := rfs.db.QueryRowContext(ctx, query, args...).Scan(
		&metrics.UsageBytes,
		&metrics.ArchivalCount,
		&metrics.DeletionCount,
		&metrics.PolicyImpact,
		&metrics.StorageSavings,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query actual metrics: %w", err)
	}

	// Calculate cost savings
	metrics.CostSavings = float64(metrics.StorageSavings) * rfs.config.CostPerGBPerMonth / (1024 * 1024 * 1024)

	return &metrics, nil
}

// calculateAccuracyPercentage calculates accuracy percentage between predicted and actual values
func (rfs *RetentionForecastService) calculateAccuracyPercentage(predicted, actual float64) float64 {
	if actual == 0 {
		if predicted == 0 {
			return 100.0 // Both zero, perfect accuracy
		}
		return 0.0 // Actual is zero but predicted is not
	}

	deviation := math.Abs(predicted-actual) / actual
	accuracy := (1.0 - deviation) * 100.0
	return math.Max(0.0, accuracy)
}

// ForecastAccuracyLog represents a forecast accuracy log record
type ForecastAccuracyLog struct {
	ID                             int64     `json:"id"`
	ForecastID                     int64     `json:"forecast_id"`
	ActualUsageBytes               int64     `json:"actual_usage_bytes"`
	ActualArchivalCount            int       `json:"actual_archival_count"`
	ActualDeletionCount            int       `json:"actual_deletion_count"`
	ActualPolicyImpact             float64   `json:"actual_policy_impact"`
	ActualStorageSavingsBytes      int64     `json:"actual_storage_savings_bytes"`
	ActualCostSavingsUSD           float64   `json:"actual_cost_savings_usd"`
	UsageAccuracyPercentage        float64   `json:"usage_accuracy_percentage"`
	ArchivalAccuracyPercentage     float64   `json:"archival_accuracy_percentage"`
	DeletionAccuracyPercentage     float64   `json:"deletion_accuracy_percentage"`
	PolicyImpactAccuracyPercentage float64   `json:"policy_impact_accuracy_percentage"`
	OverallAccuracyScore           float64   `json:"overall_accuracy_score"`
	EvaluatedAt                    time.Time `json:"evaluated_at"`
	EvaluationWindowHours          int       `json:"evaluation_window_hours"`
	CreatedAt                      time.Time `json:"created_at"`
}

// storeForecastAccuracyLog stores a forecast accuracy log
func (rfs *RetentionForecastService) storeForecastAccuracyLog(ctx context.Context, log *ForecastAccuracyLog) error {
	query := `
		INSERT INTO forecast_accuracy_logs (
			forecast_id, actual_usage_bytes, actual_archival_count, actual_deletion_count,
			actual_policy_impact, actual_storage_savings_bytes, actual_cost_savings_usd,
			usage_accuracy_percentage, archival_accuracy_percentage, deletion_accuracy_percentage,
			policy_impact_accuracy_percentage, overall_accuracy_score, evaluated_at, evaluation_window_hours
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := rfs.db.ExecContext(ctx, query,
		log.ForecastID,
		log.ActualUsageBytes,
		log.ActualArchivalCount,
		log.ActualDeletionCount,
		log.ActualPolicyImpact,
		log.ActualStorageSavingsBytes,
		log.ActualCostSavingsUSD,
		log.UsageAccuracyPercentage,
		log.ArchivalAccuracyPercentage,
		log.DeletionAccuracyPercentage,
		log.PolicyImpactAccuracyPercentage,
		log.OverallAccuracyScore,
		log.EvaluatedAt,
		log.EvaluationWindowHours,
	)

	if err != nil {
		return fmt.Errorf("failed to store forecast accuracy log: %w", err)
	}

	return nil
}
