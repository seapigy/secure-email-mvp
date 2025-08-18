package email

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
)

// RetentionAnomaly represents a retention anomaly record
type RetentionAnomaly struct {
	ID                    int64      `json:"id"`
	AnomalyType           string     `json:"anomaly_type"`
	Severity              string     `json:"severity"`
	Title                 string     `json:"title"`
	Description           string     `json:"description"`
	DetectedAt            time.Time  `json:"detected_at"`
	ScopeType             string     `json:"scope_type"`
	ScopeKey              *string    `json:"scope_key,omitempty"`
	BaselineValue         float64    `json:"baseline_value"`
	CurrentValue          float64    `json:"current_value"`
	DeviationPercentage   float64    `json:"deviation_percentage"`
	ThresholdPercentage   float64    `json:"threshold_percentage"`
	AffectedPolicies      *string    `json:"affected_policies,omitempty"`
	AffectedEmailsCount   int        `json:"affected_emails_count"`
	TimeWindowHours       int        `json:"time_window_hours"`
	Status                string     `json:"status"`
	AcknowledgedAt        *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy        *string    `json:"acknowledged_by,omitempty"`
	ResolutionNotes       *string    `json:"resolution_notes,omitempty"`
	ResolvedAt            *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy            *string    `json:"resolved_by,omitempty"`
	RecommendedAction     *string    `json:"recommended_action,omitempty"`
	AutoResolutionEnabled bool       `json:"auto_resolution_enabled"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// AnomalyConfig represents configuration for anomaly detection
type AnomalyConfig struct {
	SpikeDeletionThreshold     float64 `json:"spike_deletion_threshold"`
	DropPolicyMatchesThreshold float64 `json:"drop_policy_matches_threshold"`
	ForecastDeviationThreshold float64 `json:"forecast_deviation_threshold"`
	UnusualArchivalThreshold   float64 `json:"unusual_archival_threshold"`
	DetectionWindowHours       int     `json:"detection_window_hours"`
	AutoResolutionEnabled      bool    `json:"auto_resolution_enabled"`
	MinConfidenceThreshold     float64 `json:"min_confidence_threshold"`
}

// AnomalyMetrics represents metrics for anomaly detection
type AnomalyMetrics struct {
	DeletionCount     int     `json:"deletion_count"`
	PolicyMatches     int     `json:"policy_matches"`
	ArchivalCount     int     `json:"archival_count"`
	ForecastDeviation float64 `json:"forecast_deviation"`
	TimeWindow        int     `json:"time_window_hours"`
	BaselinePeriod    int     `json:"baseline_period_hours"`
}

// RetentionAnomalyDetector provides anomaly detection for retention operations
type RetentionAnomalyDetector struct {
	db     *sql.DB
	config *AnomalyConfig
}

// NewRetentionAnomalyDetector creates a new retention anomaly detector
func NewRetentionAnomalyDetector(db *sql.DB, config *AnomalyConfig) *RetentionAnomalyDetector {
	if config == nil {
		config = &AnomalyConfig{
			SpikeDeletionThreshold:     200.0, // 200% increase
			DropPolicyMatchesThreshold: 50.0,  // 50% decrease
			ForecastDeviationThreshold: 25.0,  // 25% deviation
			UnusualArchivalThreshold:   150.0, // 150% increase
			DetectionWindowHours:       24,
			AutoResolutionEnabled:      false,
			MinConfidenceThreshold:     0.8,
		}
	}

	return &RetentionAnomalyDetector{
		db:     db,
		config: config,
	}
}

// DetectAnomalies runs anomaly detection for all scopes
func (rad *RetentionAnomalyDetector) DetectAnomalies(ctx context.Context) error {
	log.Println("Starting retention anomaly detection...")

	// Detect global anomalies
	if err := rad.detectGlobalAnomalies(ctx); err != nil {
		log.Printf("Failed to detect global anomalies: %v", err)
	}

	// Detect user-specific anomalies
	if err := rad.detectUserAnomalies(ctx); err != nil {
		log.Printf("Failed to detect user anomalies: %v", err)
	}

	// Detect domain-specific anomalies
	if err := rad.detectDomainAnomalies(ctx); err != nil {
		log.Printf("Failed to detect domain anomalies: %v", err)
	}

	log.Println("Retention anomaly detection completed")
	return nil
}

// detectGlobalAnomalies detects anomalies for the global scope
func (rad *RetentionAnomalyDetector) detectGlobalAnomalies(ctx context.Context) error {
	// Detect spike in deletions
	if err := rad.detectSpikeDeletions(ctx, "global", nil); err != nil {
		log.Printf("Failed to detect global deletion spike: %v", err)
	}

	// Detect drop in policy matches
	if err := rad.detectDropPolicyMatches(ctx, "global", nil); err != nil {
		log.Printf("Failed to detect global policy match drop: %v", err)
	}

	// Detect forecast deviation
	if err := rad.detectForecastDeviation(ctx, "global", nil); err != nil {
		log.Printf("Failed to detect global forecast deviation: %v", err)
	}

	// Detect unusual archival activity
	if err := rad.detectUnusualArchival(ctx, "global", nil); err != nil {
		log.Printf("Failed to detect global unusual archival: %v", err)
	}

	return nil
}

// detectUserAnomalies detects anomalies for individual users
func (rad *RetentionAnomalyDetector) detectUserAnomalies(ctx context.Context) error {
	users, err := rad.getActiveUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active users: %w", err)
	}

	for _, userID := range users {
		// Detect spike in deletions
		if err := rad.detectSpikeDeletions(ctx, "user", &userID); err != nil {
			log.Printf("Failed to detect deletion spike for user %s: %v", userID, err)
		}

		// Detect drop in policy matches
		if err := rad.detectDropPolicyMatches(ctx, "user", &userID); err != nil {
			log.Printf("Failed to detect policy match drop for user %s: %v", userID, err)
		}

		// Detect forecast deviation
		if err := rad.detectForecastDeviation(ctx, "user", &userID); err != nil {
			log.Printf("Failed to detect forecast deviation for user %s: %v", userID, err)
		}

		// Detect unusual archival activity
		if err := rad.detectUnusualArchival(ctx, "user", &userID); err != nil {
			log.Printf("Failed to detect unusual archival for user %s: %v", userID, err)
		}
	}

	return nil
}

// detectDomainAnomalies detects anomalies for domains
func (rad *RetentionAnomalyDetector) detectDomainAnomalies(ctx context.Context) error {
	domains, err := rad.getActiveDomains(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active domains: %w", err)
	}

	for _, domain := range domains {
		// Detect spike in deletions
		if err := rad.detectSpikeDeletions(ctx, "domain", &domain); err != nil {
			log.Printf("Failed to detect deletion spike for domain %s: %v", domain, err)
		}

		// Detect drop in policy matches
		if err := rad.detectDropPolicyMatches(ctx, "domain", &domain); err != nil {
			log.Printf("Failed to detect policy match drop for domain %s: %v", domain, err)
		}

		// Detect forecast deviation
		if err := rad.detectForecastDeviation(ctx, "domain", &domain); err != nil {
			log.Printf("Failed to detect forecast deviation for domain %s: %v", domain, err)
		}

		// Detect unusual archival activity
		if err := rad.detectUnusualArchival(ctx, "domain", &domain); err != nil {
			log.Printf("Failed to detect unusual archival for domain %s: %v", domain, err)
		}
	}

	return nil
}

// detectSpikeDeletions detects sudden spikes in deletion operations
func (rad *RetentionAnomalyDetector) detectSpikeDeletions(ctx context.Context, scopeType string, scopeKey *string) error {
	// Get current deletion count
	currentMetrics, err := rad.getCurrentMetrics(ctx, scopeType, scopeKey, rad.config.DetectionWindowHours)
	if err != nil {
		return fmt.Errorf("failed to get current metrics: %w", err)
	}

	// Get baseline deletion count
	baselineMetrics, err := rad.getBaselineMetrics(ctx, scopeType, scopeKey, rad.config.DetectionWindowHours*7) // 7x longer baseline
	if err != nil {
		return fmt.Errorf("failed to get baseline metrics: %w", err)
	}

	// Calculate deviation
	if baselineMetrics.DeletionCount > 0 {
		deviation := float64(currentMetrics.DeletionCount-baselineMetrics.DeletionCount) / float64(baselineMetrics.DeletionCount) * 100.0

		if deviation > rad.config.SpikeDeletionThreshold {
			// Create anomaly
			anomaly := &RetentionAnomaly{
				AnomalyType:           "spike_deletion",
				Severity:              rad.calculateSeverity(deviation),
				Title:                 fmt.Sprintf("Unusual spike in email deletions detected"),
				Description:           fmt.Sprintf("Deletion count increased by %.1f%% compared to baseline", deviation),
				DetectedAt:            time.Now(),
				ScopeType:             scopeType,
				ScopeKey:              scopeKey,
				BaselineValue:         float64(baselineMetrics.DeletionCount),
				CurrentValue:          float64(currentMetrics.DeletionCount),
				DeviationPercentage:   deviation,
				ThresholdPercentage:   rad.config.SpikeDeletionThreshold,
				TimeWindowHours:       rad.config.DetectionWindowHours,
				Status:                "active",
				AutoResolutionEnabled: rad.config.AutoResolutionEnabled,
			}

			if err := rad.storeAnomaly(ctx, anomaly); err != nil {
				return fmt.Errorf("failed to store deletion spike anomaly: %w", err)
			}
		}
	}

	return nil
}

// detectDropPolicyMatches detects unusual drops in policy matches
func (rad *RetentionAnomalyDetector) detectDropPolicyMatches(ctx context.Context, scopeType string, scopeKey *string) error {
	// Get current policy match count
	currentMetrics, err := rad.getCurrentMetrics(ctx, scopeType, scopeKey, rad.config.DetectionWindowHours)
	if err != nil {
		return fmt.Errorf("failed to get current metrics: %w", err)
	}

	// Get baseline policy match count
	baselineMetrics, err := rad.getBaselineMetrics(ctx, scopeType, scopeKey, rad.config.DetectionWindowHours*7)
	if err != nil {
		return fmt.Errorf("failed to get baseline metrics: %w", err)
	}

	// Calculate deviation (negative for drops)
	if baselineMetrics.PolicyMatches > 0 {
		deviation := float64(baselineMetrics.PolicyMatches-currentMetrics.PolicyMatches) / float64(baselineMetrics.PolicyMatches) * 100.0

		if deviation > rad.config.DropPolicyMatchesThreshold {
			// Create anomaly
			anomaly := &RetentionAnomaly{
				AnomalyType:           "drop_policy_matches",
				Severity:              rad.calculateSeverity(deviation),
				Title:                 fmt.Sprintf("Unusual drop in policy matches detected"),
				Description:           fmt.Sprintf("Policy matches decreased by %.1f%% compared to baseline", deviation),
				DetectedAt:            time.Now(),
				ScopeType:             scopeType,
				ScopeKey:              scopeKey,
				BaselineValue:         float64(baselineMetrics.PolicyMatches),
				CurrentValue:          float64(currentMetrics.PolicyMatches),
				DeviationPercentage:   deviation,
				ThresholdPercentage:   rad.config.DropPolicyMatchesThreshold,
				TimeWindowHours:       rad.config.DetectionWindowHours,
				Status:                "active",
				AutoResolutionEnabled: rad.config.AutoResolutionEnabled,
			}

			if err := rad.storeAnomaly(ctx, anomaly); err != nil {
				return fmt.Errorf("failed to store policy match drop anomaly: %w", err)
			}
		}
	}

	return nil
}

// detectForecastDeviation detects deviations from forecasts
func (rad *RetentionAnomalyDetector) detectForecastDeviation(ctx context.Context, scopeType string, scopeKey *string) error {
	// Get latest forecast
	forecast, err := rad.getLatestForecast(ctx, scopeType, scopeKey)
	if err != nil {
		// No forecast available, skip this check
		return nil
	}

	// Get actual metrics for the forecast period
	actualMetrics, err := rad.getActualMetricsForForecast(ctx, forecast)
	if err != nil {
		return fmt.Errorf("failed to get actual metrics: %w", err)
	}

	// Calculate deviation for usage
	if forecast.PredictedUsageBytes > 0 {
		deviation := math.Abs(float64(actualMetrics.UsageBytes-forecast.PredictedUsageBytes)) / float64(forecast.PredictedUsageBytes) * 100.0

		if deviation > rad.config.ForecastDeviationThreshold {
			// Create anomaly
			anomaly := &RetentionAnomaly{
				AnomalyType:           "forecast_deviation",
				Severity:              rad.calculateSeverity(deviation),
				Title:                 fmt.Sprintf("Forecast deviation detected"),
				Description:           fmt.Sprintf("Actual usage deviated by %.1f%% from forecast", deviation),
				DetectedAt:            time.Now(),
				ScopeType:             scopeType,
				ScopeKey:              scopeKey,
				BaselineValue:         float64(forecast.PredictedUsageBytes),
				CurrentValue:          float64(actualMetrics.UsageBytes),
				DeviationPercentage:   deviation,
				ThresholdPercentage:   rad.config.ForecastDeviationThreshold,
				TimeWindowHours:       rad.config.DetectionWindowHours,
				Status:                "active",
				AutoResolutionEnabled: rad.config.AutoResolutionEnabled,
			}

			if err := rad.storeAnomaly(ctx, anomaly); err != nil {
				return fmt.Errorf("failed to store forecast deviation anomaly: %w", err)
			}
		}
	}

	return nil
}

// detectUnusualArchival detects unusual archival activity
func (rad *RetentionAnomalyDetector) detectUnusualArchival(ctx context.Context, scopeType string, scopeKey *string) error {
	// Get current archival count
	currentMetrics, err := rad.getCurrentMetrics(ctx, scopeType, scopeKey, rad.config.DetectionWindowHours)
	if err != nil {
		return fmt.Errorf("failed to get current metrics: %w", err)
	}

	// Get baseline archival count
	baselineMetrics, err := rad.getBaselineMetrics(ctx, scopeType, scopeKey, rad.config.DetectionWindowHours*7)
	if err != nil {
		return fmt.Errorf("failed to get baseline metrics: %w", err)
	}

	// Calculate deviation
	if baselineMetrics.ArchivalCount > 0 {
		deviation := float64(currentMetrics.ArchivalCount-baselineMetrics.ArchivalCount) / float64(baselineMetrics.ArchivalCount) * 100.0

		if deviation > rad.config.UnusualArchivalThreshold {
			// Create anomaly
			anomaly := &RetentionAnomaly{
				AnomalyType:           "unusual_archival",
				Severity:              rad.calculateSeverity(deviation),
				Title:                 fmt.Sprintf("Unusual archival activity detected"),
				Description:           fmt.Sprintf("Archival count increased by %.1f%% compared to baseline", deviation),
				DetectedAt:            time.Now(),
				ScopeType:             scopeType,
				ScopeKey:              scopeKey,
				BaselineValue:         float64(baselineMetrics.ArchivalCount),
				CurrentValue:          float64(currentMetrics.ArchivalCount),
				DeviationPercentage:   deviation,
				ThresholdPercentage:   rad.config.UnusualArchivalThreshold,
				TimeWindowHours:       rad.config.DetectionWindowHours,
				Status:                "active",
				AutoResolutionEnabled: rad.config.AutoResolutionEnabled,
			}

			if err := rad.storeAnomaly(ctx, anomaly); err != nil {
				return fmt.Errorf("failed to store unusual archival anomaly: %w", err)
			}
		}
	}

	return nil
}

// getCurrentMetrics retrieves current metrics for anomaly detection
func (rad *RetentionAnomalyDetector) getCurrentMetrics(ctx context.Context, scopeType string, scopeKey *string, hours int) (*AnomalyMetrics, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var query string
	var args []interface{}

	switch scopeType {
	case "global":
		query = `
			SELECT 
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				COUNT(CASE WHEN pel.evaluation_result = 'matched' THEN 1 END) as policy_matches,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				? as time_window_hours
			FROM emails e
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id AND cl.cleanup_time >= ?
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id AND pel.evaluated_at >= ?
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id AND ae.archived_at >= ?
			WHERE e.created_at >= ?
		`
		args = []interface{}{hours, startTime, startTime, startTime, startTime}

	case "user":
		query = `
			SELECT 
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				COUNT(CASE WHEN pel.evaluation_result = 'matched' THEN 1 END) as policy_matches,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				? as time_window_hours
			FROM emails e
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id AND cl.cleanup_time >= ?
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id AND pel.evaluated_at >= ?
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id AND ae.archived_at >= ?
			WHERE e.sender_id = ? AND e.created_at >= ?
		`
		args = []interface{}{hours, startTime, startTime, startTime, *scopeKey, startTime}

	case "domain":
		query = `
			SELECT 
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				COUNT(CASE WHEN pel.evaluation_result = 'matched' THEN 1 END) as policy_matches,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				? as time_window_hours
			FROM emails e
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id AND cl.cleanup_time >= ?
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id AND pel.evaluated_at >= ?
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id AND ae.archived_at >= ?
			WHERE (e.recipient LIKE ? OR e.recipient LIKE ?) AND e.created_at >= ?
		`
		domainPattern := "%@" + *scopeKey
		args = []interface{}{hours, startTime, startTime, startTime, domainPattern, domainPattern, startTime}

	default:
		return nil, fmt.Errorf("unsupported scope type: %s", scopeType)
	}

	var metrics AnomalyMetrics
	err := rad.db.QueryRowContext(ctx, query, args...).Scan(
		&metrics.DeletionCount,
		&metrics.PolicyMatches,
		&metrics.ArchivalCount,
		&metrics.TimeWindow,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query current metrics: %w", err)
	}

	return &metrics, nil
}

// getBaselineMetrics retrieves baseline metrics for comparison
func (rad *RetentionAnomalyDetector) getBaselineMetrics(ctx context.Context, scopeType string, scopeKey *string, hours int) (*AnomalyMetrics, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var query string
	var args []interface{}

	switch scopeType {
	case "global":
		query = `
			SELECT 
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				COUNT(CASE WHEN pel.evaluation_result = 'matched' THEN 1 END) as policy_matches,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				? as time_window_hours
			FROM emails e
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id AND cl.cleanup_time >= ?
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id AND pel.evaluated_at >= ?
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id AND ae.archived_at >= ?
			WHERE e.created_at >= ?
		`
		args = []interface{}{hours, startTime, startTime, startTime, startTime}

	case "user":
		query = `
			SELECT 
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				COUNT(CASE WHEN pel.evaluation_result = 'matched' THEN 1 END) as policy_matches,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				? as time_window_hours
			FROM emails e
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id AND cl.cleanup_time >= ?
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id AND pel.evaluated_at >= ?
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id AND ae.archived_at >= ?
			WHERE e.sender_id = ? AND e.created_at >= ?
		`
		args = []interface{}{hours, startTime, startTime, startTime, *scopeKey, startTime}

	case "domain":
		query = `
			SELECT 
				COUNT(CASE WHEN cl.cleanup_reason IS NOT NULL THEN 1 END) as deletion_count,
				COUNT(CASE WHEN pel.evaluation_result = 'matched' THEN 1 END) as policy_matches,
				COUNT(CASE WHEN ae.id IS NOT NULL THEN 1 END) as archival_count,
				? as time_window_hours
			FROM emails e
			LEFT JOIN cleanup_logs cl ON e.email_id = cl.email_id AND cl.cleanup_time >= ?
			LEFT JOIN policy_evaluation_logs pel ON e.email_id = pel.email_id AND pel.evaluated_at >= ?
			LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id AND ae.archived_at >= ?
			WHERE (e.recipient LIKE ? OR e.recipient LIKE ?) AND e.created_at >= ?
		`
		domainPattern := "%@" + *scopeKey
		args = []interface{}{hours, startTime, startTime, startTime, domainPattern, domainPattern, startTime}

	default:
		return nil, fmt.Errorf("unsupported scope type: %s", scopeType)
	}

	var metrics AnomalyMetrics
	err := rad.db.QueryRowContext(ctx, query, args...).Scan(
		&metrics.DeletionCount,
		&metrics.PolicyMatches,
		&metrics.ArchivalCount,
		&metrics.TimeWindow,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query baseline metrics: %w", err)
	}

	return &metrics, nil
}

// getLatestForecast retrieves the latest forecast for a scope
func (rad *RetentionAnomalyDetector) getLatestForecast(ctx context.Context, scopeType string, scopeKey *string) (*RetentionForecast, error) {
	var query string
	var args []interface{}

	if scopeKey != nil {
		query = `
			SELECT id, forecast_type, forecast_key, generated_at, target_period_end,
				   predicted_usage_bytes, predicted_archival_count, predicted_deletion_count,
				   predicted_policy_impact, predicted_storage_savings_bytes, predicted_cost_savings_usd,
				   confidence_score, accuracy_score, forecast_model_version,
				   historical_data_points, data_freshness_hours, created_at, updated_at
			FROM retention_forecasts
			WHERE forecast_type = ? AND forecast_key = ?
			ORDER BY generated_at DESC
			LIMIT 1
		`
		args = []interface{}{scopeType, *scopeKey}
	} else {
		query = `
			SELECT id, forecast_type, forecast_key, generated_at, target_period_end,
				   predicted_usage_bytes, predicted_archival_count, predicted_deletion_count,
				   predicted_policy_impact, predicted_storage_savings_bytes, predicted_cost_savings_usd,
				   confidence_score, accuracy_score, forecast_model_version,
				   historical_data_points, data_freshness_hours, created_at, updated_at
			FROM retention_forecasts
			WHERE forecast_type = ?
			ORDER BY generated_at DESC
			LIMIT 1
		`
		args = []interface{}{scopeType}
	}

	var forecast RetentionForecast
	err := rad.db.QueryRowContext(ctx, query, args...).Scan(
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
		return nil, fmt.Errorf("failed to get latest forecast: %w", err)
	}

	return &forecast, nil
}

// getActualMetricsForForecast retrieves actual metrics for a forecast period
func (rad *RetentionAnomalyDetector) getActualMetricsForForecast(ctx context.Context, forecast *RetentionForecast) (*ForecastMetrics, error) {
	// This is a simplified version - in a real implementation, you'd calculate actual metrics
	// for the specific forecast period
	return &ForecastMetrics{
		UsageBytes:     forecast.PredictedUsageBytes, // Placeholder
		ArchivalCount:  forecast.PredictedArchivalCount,
		DeletionCount:  forecast.PredictedDeletionCount,
		PolicyImpact:   forecast.PredictedPolicyImpact,
		StorageSavings: forecast.PredictedStorageSavingsBytes,
		CostSavings:    forecast.PredictedCostSavingsUSD,
	}, nil
}

// calculateSeverity calculates severity based on deviation percentage
func (rad *RetentionAnomalyDetector) calculateSeverity(deviation float64) string {
	if deviation > 500 {
		return "critical"
	} else if deviation > 200 {
		return "high"
	} else if deviation > 100 {
		return "medium"
	} else {
		return "low"
	}
}

// storeAnomaly stores an anomaly in the database
func (rad *RetentionAnomalyDetector) storeAnomaly(ctx context.Context, anomaly *RetentionAnomaly) error {
	query := `
		INSERT INTO retention_anomalies (
			anomaly_type, severity, title, description, detected_at,
			scope_type, scope_key, baseline_value, current_value,
			deviation_percentage, threshold_percentage, affected_policies,
			affected_emails_count, time_window_hours, status,
			recommended_action, auto_resolution_enabled, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := rad.db.ExecContext(ctx, query,
		anomaly.AnomalyType,
		anomaly.Severity,
		anomaly.Title,
		anomaly.Description,
		anomaly.DetectedAt,
		anomaly.ScopeType,
		anomaly.ScopeKey,
		anomaly.BaselineValue,
		anomaly.CurrentValue,
		anomaly.DeviationPercentage,
		anomaly.ThresholdPercentage,
		anomaly.AffectedPolicies,
		anomaly.AffectedEmailsCount,
		anomaly.TimeWindowHours,
		anomaly.Status,
		anomaly.RecommendedAction,
		anomaly.AutoResolutionEnabled,
		anomaly.CreatedAt,
		anomaly.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store anomaly: %w", err)
	}

	return nil
}

// GetAnomalies retrieves anomalies with filtering
func (rad *RetentionAnomalyDetector) GetAnomalies(ctx context.Context, scopeType, scopeKey, severity, status string, limit int) ([]*RetentionAnomaly, error) {
	query := `
		SELECT id, anomaly_type, severity, title, description, detected_at,
			   scope_type, scope_key, baseline_value, current_value,
			   deviation_percentage, threshold_percentage, affected_policies,
			   affected_emails_count, time_window_hours, status,
			   acknowledged_at, acknowledged_by, resolution_notes,
			   resolved_at, resolved_by, recommended_action,
			   auto_resolution_enabled, created_at, updated_at
		FROM retention_anomalies
		WHERE 1=1
	`
	var args []interface{}

	if scopeType != "" {
		query += " AND scope_type = ?"
		args = append(args, scopeType)
	}

	if scopeKey != "" {
		query += " AND scope_key = ?"
		args = append(args, scopeKey)
	}

	if severity != "" {
		query += " AND severity = ?"
		args = append(args, severity)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY detected_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := rad.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query anomalies: %w", err)
	}
	defer rows.Close()

	var anomalies []*RetentionAnomaly
	for rows.Next() {
		var anomaly RetentionAnomaly
		err := rows.Scan(
			&anomaly.ID,
			&anomaly.AnomalyType,
			&anomaly.Severity,
			&anomaly.Title,
			&anomaly.Description,
			&anomaly.DetectedAt,
			&anomaly.ScopeType,
			&anomaly.ScopeKey,
			&anomaly.BaselineValue,
			&anomaly.CurrentValue,
			&anomaly.DeviationPercentage,
			&anomaly.ThresholdPercentage,
			&anomaly.AffectedPolicies,
			&anomaly.AffectedEmailsCount,
			&anomaly.TimeWindowHours,
			&anomaly.Status,
			&anomaly.AcknowledgedAt,
			&anomaly.AcknowledgedBy,
			&anomaly.ResolutionNotes,
			&anomaly.ResolvedAt,
			&anomaly.ResolvedBy,
			&anomaly.RecommendedAction,
			&anomaly.AutoResolutionEnabled,
			&anomaly.CreatedAt,
			&anomaly.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan anomaly: %w", err)
		}
		anomalies = append(anomalies, &anomaly)
	}

	return anomalies, nil
}

// AcknowledgeAnomaly acknowledges an anomaly
func (rad *RetentionAnomalyDetector) AcknowledgeAnomaly(ctx context.Context, anomalyID int64, acknowledgedBy, resolutionNotes string) error {
	query := `
		UPDATE retention_anomalies
		SET status = 'acknowledged',
			acknowledged_at = ?,
			acknowledged_by = ?,
			resolution_notes = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := rad.db.ExecContext(ctx, query,
		time.Now(),
		acknowledgedBy,
		resolutionNotes,
		time.Now(),
		anomalyID,
	)

	if err != nil {
		return fmt.Errorf("failed to acknowledge anomaly: %w", err)
	}

	return nil
}

// getActiveUsers retrieves list of active users for anomaly detection
func (rad *RetentionAnomalyDetector) getActiveUsers(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT sender_id
		FROM emails
		WHERE created_at >= datetime('now', '-30 days')
		GROUP BY sender_id
		HAVING COUNT(*) >= 5
	`

	rows, err := rad.db.QueryContext(ctx, query)
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

// getActiveDomains retrieves list of active domains for anomaly detection
func (rad *RetentionAnomalyDetector) getActiveDomains(ctx context.Context) ([]string, error) {
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

	rows, err := rad.db.QueryContext(ctx, query)
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
