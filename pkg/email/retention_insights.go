package email

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// RetentionInsight represents a retention insight record
type RetentionInsight struct {
	ID                          int64     `json:"id"`
	InsightDate                 time.Time `json:"insight_date"`
	InsightType                 string    `json:"insight_type"`
	MostCommonPolicyTriggers    string    `json:"most_common_policy_triggers,omitempty"`
	VolumeTrendsArchivedVsDeleted string  `json:"volume_trends_archived_vs_deleted,omitempty"`
	AvgStorageSavingsCompression float64   `json:"avg_storage_savings_compression"`
	PolicyEffectivenessScore    float64   `json:"policy_effectiveness_score"`
	TotalStorageSavingsBytes    int64     `json:"total_storage_savings_bytes"`
	EstimatedCostSavingsUSD     float64   `json:"estimated_cost_savings_usd"`
	CompressionRatioAvg         float64   `json:"compression_ratio_avg"`
	PoliciesMostEffective       string    `json:"policies_most_effective,omitempty"`
	PoliciesLeastEffective      string    `json:"policies_least_effective,omitempty"`
	OverrideFrequency           int       `json:"override_frequency"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

// PolicyTrigger represents a policy trigger with count
type PolicyTrigger struct {
	TriggerType string `json:"trigger_type"`
	Count       int    `json:"count"`
	Percentage  float64 `json:"percentage"`
}

// VolumeTrend represents volume trends for archived vs deleted
type VolumeTrend struct {
	Date      string `json:"date"`
	Archived  int    `json:"archived"`
	Deleted   int    `json:"deleted"`
	Total     int    `json:"total"`
	ArchivedPercentage float64 `json:"archived_percentage"`
}

// PolicyEffectiveness represents policy effectiveness metrics
type PolicyEffectiveness struct {
	PolicyID   int64   `json:"policy_id"`
	PolicyName string  `json:"policy_name"`
	EffectivenessScore float64 `json:"effectiveness_score"`
	StorageSavings     int64   `json:"storage_savings"`
	EmailCount         int     `json:"email_count"`
}

// RetentionInsightsService provides methods for generating and retrieving retention insights
type RetentionInsightsService struct {
	db *sql.DB
}

// NewRetentionInsightsService creates a new retention insights service
func NewRetentionInsightsService(db *sql.DB) *RetentionInsightsService {
	return &RetentionInsightsService{
		db: db,
	}
}

// GenerateDailyInsights generates daily retention insights
func (ris *RetentionInsightsService) GenerateDailyInsights(ctx context.Context, date time.Time) error {
	log.Printf("Generating daily retention insights for %s", date.Format("2006-01-02"))

	// Check if insights already exist for this date
	existing, err := ris.getInsightByDateAndType(ctx, date, "daily_rollup")
	if err == nil && existing != nil {
		log.Printf("Daily insights already exist for %s, skipping generation", date.Format("2006-01-02"))
		return nil
	}

	// Generate insights
	insight := &RetentionInsight{
		InsightDate: date,
		InsightType: "daily_rollup",
	}

	// Analyze most common policy triggers
	policyTriggers, err := ris.analyzePolicyTriggers(ctx, date)
	if err != nil {
		log.Printf("Failed to analyze policy triggers: %v", err)
	} else {
		insight.MostCommonPolicyTriggers = policyTriggers
	}

	// Analyze volume trends
	volumeTrends, err := ris.analyzeVolumeTrends(ctx, date)
	if err != nil {
		log.Printf("Failed to analyze volume trends: %v", err)
	} else {
		insight.VolumeTrendsArchivedVsDeleted = volumeTrends
	}

	// Calculate storage savings
	storageSavings, compressionRatio, err := ris.calculateStorageSavings(ctx, date)
	if err != nil {
		log.Printf("Failed to calculate storage savings: %v", err)
	} else {
		insight.TotalStorageSavingsBytes = storageSavings
		insight.CompressionRatioAvg = compressionRatio
		insight.AvgStorageSavingsCompression = compressionRatio
	}

	// Calculate policy effectiveness
	effectivenessScore, err := ris.calculatePolicyEffectiveness(ctx, date)
	if err != nil {
		log.Printf("Failed to calculate policy effectiveness: %v", err)
	} else {
		insight.PolicyEffectivenessScore = effectivenessScore
	}

	// Analyze policy performance
	mostEffective, leastEffective, err := ris.analyzePolicyPerformance(ctx, date)
	if err != nil {
		log.Printf("Failed to analyze policy performance: %v", err)
	} else {
		insight.PoliciesMostEffective = mostEffective
		insight.PoliciesLeastEffective = leastEffective
	}

	// Count policy overrides
	overrideCount, err := ris.countPolicyOverrides(ctx, date)
	if err != nil {
		log.Printf("Failed to count policy overrides: %v", err)
	} else {
		insight.OverrideFrequency = overrideCount
	}

	// Estimate cost savings
	costSavings := ris.estimateCostSavings(storageSavings, compressionRatio)
	insight.EstimatedCostSavingsUSD = costSavings

	// Save insight
	if err := ris.saveInsight(ctx, insight); err != nil {
		return fmt.Errorf("failed to save insight: %w", err)
	}

	log.Printf("Successfully generated daily insights for %s", date.Format("2006-01-02"))
	return nil
}

// GetInsights retrieves retention insights with filtering
func (ris *RetentionInsightsService) GetInsights(ctx context.Context, filters map[string]string, limit, offset int) ([]*RetentionInsight, error) {
	query := `
		SELECT id, insight_date, insight_type, most_common_policy_triggers,
		       volume_trends_archived_vs_deleted, avg_storage_savings_compression,
		       policy_effectiveness_score, total_storage_savings_bytes,
		       estimated_cost_savings_usd, compression_ratio_avg,
		       policies_most_effective, policies_least_effective,
		       override_frequency, created_at, updated_at
		FROM retention_insights WHERE 1=1
	`

	var args []interface{}

	// Apply filters
	if insightType, ok := filters["insight_type"]; ok {
		query += " AND insight_type = ?"
		args = append(args, insightType)
	}

	if startDate, ok := filters["start_date"]; ok {
		query += " AND insight_date >= ?"
		args = append(args, startDate)
	}

	if endDate, ok := filters["end_date"]; ok {
		query += " AND insight_date <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY insight_date DESC, insight_type"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := ris.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query insights: %w", err)
	}
	defer rows.Close()

	var insights []*RetentionInsight
	for rows.Next() {
		insight := &RetentionInsight{}
		err := rows.Scan(
			&insight.ID, &insight.InsightDate, &insight.InsightType,
			&insight.MostCommonPolicyTriggers, &insight.VolumeTrendsArchivedVsDeleted,
			&insight.AvgStorageSavingsCompression, &insight.PolicyEffectivenessScore,
			&insight.TotalStorageSavingsBytes, &insight.EstimatedCostSavingsUSD,
			&insight.CompressionRatioAvg, &insight.PoliciesMostEffective,
			&insight.PoliciesLeastEffective, &insight.OverrideFrequency,
			&insight.CreatedAt, &insight.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan insight: %w", err)
		}
		insights = append(insights, insight)
	}

	return insights, nil
}

// GetTrendAnalysis generates trend analysis for a date range
func (ris *RetentionInsightsService) GetTrendAnalysis(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	query := `
		SELECT 
			insight_date,
			policy_effectiveness_score,
			avg_storage_savings_compression,
			total_storage_savings_bytes,
			estimated_cost_savings_usd,
			override_frequency
		FROM retention_insights
		WHERE insight_type = 'daily_rollup'
		  AND insight_date >= ?
		  AND insight_date <= ?
		ORDER BY insight_date
	`

	rows, err := ris.db.QueryContext(ctx, query, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query trend data: %w", err)
	}
	defer rows.Close()

	var trends []map[string]interface{}
	var totalStorageSavings int64
	var totalCostSavings float64
	var avgEffectiveness float64
	var totalOverrides int
	count := 0

	for rows.Next() {
		var date time.Time
		var effectiveness, compression, costSavings float64
		var storageSavings int64
		var overrides int

		err := rows.Scan(&date, &effectiveness, &compression, &storageSavings, &costSavings, &overrides)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trend data: %w", err)
		}

		trends = append(trends, map[string]interface{}{
			"date":                    date.Format("2006-01-02"),
			"policy_effectiveness":    effectiveness,
			"compression_ratio":       compression,
			"storage_savings_bytes":   storageSavings,
			"cost_savings_usd":        costSavings,
			"override_frequency":      overrides,
		})

		totalStorageSavings += storageSavings
		totalCostSavings += costSavings
		avgEffectiveness += effectiveness
		totalOverrides += overrides
		count++
	}

	if count > 0 {
		avgEffectiveness /= float64(count)
	}

	// Calculate trends
	trendAnalysis := map[string]interface{}{
		"date_range": map[string]string{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"daily_trends": trends,
		"summary": map[string]interface{}{
			"total_storage_savings_bytes": totalStorageSavings,
			"total_cost_savings_usd":      totalCostSavings,
			"avg_policy_effectiveness":    avgEffectiveness,
			"total_overrides":             totalOverrides,
			"data_points":                 count,
		},
	}

	return trendAnalysis, nil
}

// analyzePolicyTriggers analyzes the most common policy triggers
func (ris *RetentionInsightsService) analyzePolicyTriggers(ctx context.Context, date time.Time) (string, error) {
	query := `
		SELECT 
			CASE 
				WHEN p.user_id IS NOT NULL THEN 'user_specific'
				WHEN p.sender_domain IS NOT NULL THEN 'sender_domain'
				WHEN p.recipient_domain IS NOT NULL THEN 'recipient_domain'
				WHEN p.email_status IS NOT NULL THEN 'email_status'
				WHEN p.min_age_hours IS NOT NULL OR p.max_age_hours IS NOT NULL THEN 'age_based'
				ELSE 'default'
			END as trigger_type,
			COUNT(*) as count
		FROM policy_evaluation_logs pel
		JOIN retention_policies p ON pel.policy_id = p.id
		WHERE DATE(pel.evaluated_at) = ?
		  AND pel.evaluation_result = 'applied'
		GROUP BY trigger_type
		ORDER BY count DESC
	`

	rows, err := ris.db.QueryContext(ctx, query, date.Format("2006-01-02"))
	if err != nil {
		return "", fmt.Errorf("failed to query policy triggers: %w", err)
	}
	defer rows.Close()

	var triggers []PolicyTrigger
	totalCount := 0

	for rows.Next() {
		var triggerType string
		var count int
		err := rows.Scan(&triggerType, &count)
		if err != nil {
			return "", fmt.Errorf("failed to scan policy trigger: %w", err)
		}
		totalCount += count
		triggers = append(triggers, PolicyTrigger{TriggerType: triggerType, Count: count})
	}

	// Calculate percentages
	for i := range triggers {
		if totalCount > 0 {
			triggers[i].Percentage = float64(triggers[i].Count) / float64(totalCount) * 100
		}
	}

	jsonData, err := json.Marshal(triggers)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy triggers: %w", err)
	}

	return string(jsonData), nil
}

// analyzeVolumeTrends analyzes volume trends for archived vs deleted emails
func (ris *RetentionInsightsService) analyzeVolumeTrends(ctx context.Context, date time.Time) (string, error) {
	// Get archived emails count
	archivedQuery := `
		SELECT COUNT(*) FROM archived_emails 
		WHERE DATE(archived_at) = ?
	`
	var archivedCount int
	err := ris.db.QueryRowContext(ctx, archivedQuery, date.Format("2006-01-02")).Scan(&archivedCount)
	if err != nil {
		return "", fmt.Errorf("failed to get archived count: %w", err)
	}

	// Get deleted emails count (from policy evaluation logs)
	deletedQuery := `
		SELECT COUNT(*) FROM policy_evaluation_logs pel
		JOIN retention_policies p ON pel.policy_id = p.id
		WHERE DATE(pel.evaluated_at) = ?
		  AND pel.evaluation_result = 'applied'
		  AND p.archive_instead = 0
	`
	var deletedCount int
	err = ris.db.QueryRowContext(ctx, deletedQuery, date.Format("2006-01-02")).Scan(&deletedCount)
	if err != nil {
		return "", fmt.Errorf("failed to get deleted count: %w", err)
	}

	total := archivedCount + deletedCount
	archivedPercentage := 0.0
	if total > 0 {
		archivedPercentage = float64(archivedCount) / float64(total) * 100
	}

	trend := VolumeTrend{
		Date:              date.Format("2006-01-02"),
		Archived:          archivedCount,
		Deleted:           deletedCount,
		Total:             total,
		ArchivedPercentage: archivedPercentage,
	}

	jsonData, err := json.Marshal(trend)
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume trend: %w", err)
	}

	return string(jsonData), nil
}

// calculateStorageSavings calculates storage savings from compression
func (ris *RetentionInsightsService) calculateStorageSavings(ctx context.Context, date time.Time) (int64, float64, error) {
	query := `
		SELECT 
			SUM(original_size) as total_original,
			SUM(compressed_size) as total_compressed
		FROM archived_emails
		WHERE DATE(archived_at) = ?
	`

	var totalOriginal, totalCompressed int64
	err := ris.db.QueryRowContext(ctx, query, date.Format("2006-01-02")).Scan(&totalOriginal, &totalCompressed)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to calculate storage savings: %w", err)
	}

	savings := totalOriginal - totalCompressed
	compressionRatio := 0.0
	if totalOriginal > 0 {
		compressionRatio = float64(totalCompressed) / float64(totalOriginal)
	}

	return savings, compressionRatio, nil
}

// calculatePolicyEffectiveness calculates overall policy effectiveness score
func (ris *RetentionInsightsService) calculatePolicyEffectiveness(ctx context.Context, date time.Time) (float64, error) {
	query := `
		SELECT 
			AVG(pel.impact_score) as avg_impact,
			COUNT(*) as total_evaluations,
			SUM(CASE WHEN pel.evaluation_result = 'applied' THEN 1 ELSE 0 END) as applied_count
		FROM policy_evaluation_logs pel
		WHERE DATE(pel.evaluated_at) = ?
	`

	var avgImpact float64
	var totalEvaluations, appliedCount int
	err := ris.db.QueryRowContext(ctx, query, date.Format("2006-01-02")).Scan(&avgImpact, &totalEvaluations, &appliedCount)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate policy effectiveness: %w", err)
	}

	// Calculate effectiveness score (0.0 to 1.0)
	effectiveness := 0.0
	if totalEvaluations > 0 {
		applicationRate := float64(appliedCount) / float64(totalEvaluations)
		effectiveness = (avgImpact + applicationRate) / 2.0
	}

	return effectiveness, nil
}

// analyzePolicyPerformance analyzes most and least effective policies
func (ris *RetentionInsightsService) analyzePolicyPerformance(ctx context.Context, date time.Time) (string, string, error) {
	query := `
		SELECT 
			p.id,
			p.name,
			AVG(pel.impact_score) as avg_impact,
			SUM(pel.storage_savings_bytes) as total_savings,
			COUNT(*) as email_count
		FROM policy_evaluation_logs pel
		JOIN retention_policies p ON pel.policy_id = p.id
		WHERE DATE(pel.evaluated_at) = ?
		  AND p.active = 1
		GROUP BY p.id, p.name
		HAVING email_count >= 5
		ORDER BY avg_impact DESC
	`

	rows, err := ris.db.QueryContext(ctx, query, date.Format("2006-01-02"))
	if err != nil {
		return "", "", fmt.Errorf("failed to analyze policy performance: %w", err)
	}
	defer rows.Close()

	var policies []PolicyEffectiveness
	for rows.Next() {
		var policy PolicyEffectiveness
		err := rows.Scan(&policy.PolicyID, &policy.PolicyName, &policy.EffectivenessScore, &policy.StorageSavings, &policy.EmailCount)
		if err != nil {
			return "", "", fmt.Errorf("failed to scan policy performance: %w", err)
		}
		policies = append(policies, policy)
	}

	// Get top 5 most effective
	mostEffective := policies
	if len(policies) > 5 {
		mostEffective = policies[:5]
	}

	// Get bottom 5 least effective
	leastEffective := []PolicyEffectiveness{}
	if len(policies) > 5 {
		leastEffective = policies[len(policies)-5:]
	}

	mostEffectiveJSON, err := json.Marshal(mostEffective)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal most effective policies: %w", err)
	}

	leastEffectiveJSON, err := json.Marshal(leastEffective)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal least effective policies: %w", err)
	}

	return string(mostEffectiveJSON), string(leastEffectiveJSON), nil
}

// countPolicyOverrides counts policy overrides
func (ris *RetentionInsightsService) countPolicyOverrides(ctx context.Context, date time.Time) (int, error) {
	query := `
		SELECT COUNT(*) FROM policy_evaluation_logs pel
		WHERE DATE(pel.evaluated_at) = ?
		  AND pel.evaluation_result = 'not_matched'
		  AND pel.match_score > 0
	`

	var count int
	err := ris.db.QueryRowContext(ctx, query, date.Format("2006-01-02")).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count policy overrides: %w", err)
	}

	return count, nil
}

// estimateCostSavings estimates cost savings based on storage reduction
func (ris *RetentionInsightsService) estimateCostSavings(storageSavings int64, compressionRatio float64) float64 {
	// Estimate cost savings based on typical cloud storage costs
	// Assuming $0.023 per GB per month for standard storage
	storageCostPerGBPerMonth := 0.023
	storageSavingsGB := float64(storageSavings) / (1024 * 1024 * 1024)
	
	// Calculate monthly savings
	monthlySavings := storageSavingsGB * storageCostPerGBPerMonth
	
	// Apply compression bonus
	compressionBonus := 1.0 + (1.0 - compressionRatio) * 0.5
	
	return monthlySavings * compressionBonus
}

// getInsightByDateAndType gets an insight by date and type
func (ris *RetentionInsightsService) getInsightByDateAndType(ctx context.Context, date time.Time, insightType string) (*RetentionInsight, error) {
	query := `
		SELECT id, insight_date, insight_type, most_common_policy_triggers,
		       volume_trends_archived_vs_deleted, avg_storage_savings_compression,
		       policy_effectiveness_score, total_storage_savings_bytes,
		       estimated_cost_savings_usd, compression_ratio_avg,
		       policies_most_effective, policies_least_effective,
		       override_frequency, created_at, updated_at
		FROM retention_insights
		WHERE insight_date = ? AND insight_type = ?
	`

	insight := &RetentionInsight{}
	err := ris.db.QueryRowContext(ctx, query, date.Format("2006-01-02"), insightType).Scan(
		&insight.ID, &insight.InsightDate, &insight.InsightType,
		&insight.MostCommonPolicyTriggers, &insight.VolumeTrendsArchivedVsDeleted,
		&insight.AvgStorageSavingsCompression, &insight.PolicyEffectivenessScore,
		&insight.TotalStorageSavingsBytes, &insight.EstimatedCostSavingsUSD,
		&insight.CompressionRatioAvg, &insight.PoliciesMostEffective,
		&insight.PoliciesLeastEffective, &insight.OverrideFrequency,
		&insight.CreatedAt, &insight.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get insight: %w", err)
	}

	return insight, nil
}

// saveInsight saves a retention insight
func (ris *RetentionInsightsService) saveInsight(ctx context.Context, insight *RetentionInsight) error {
	query := `
		INSERT OR REPLACE INTO retention_insights (
			insight_date, insight_type, most_common_policy_triggers,
			volume_trends_archived_vs_deleted, avg_storage_savings_compression,
			policy_effectiveness_score, total_storage_savings_bytes,
			estimated_cost_savings_usd, compression_ratio_avg,
			policies_most_effective, policies_least_effective,
			override_frequency, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := ris.db.ExecContext(ctx, query,
		insight.InsightDate.Format("2006-01-02"), insight.InsightType,
		insight.MostCommonPolicyTriggers, insight.VolumeTrendsArchivedVsDeleted,
		insight.AvgStorageSavingsCompression, insight.PolicyEffectivenessScore,
		insight.TotalStorageSavingsBytes, insight.EstimatedCostSavingsUSD,
		insight.CompressionRatioAvg, insight.PoliciesMostEffective,
		insight.PoliciesLeastEffective, insight.OverrideFrequency,
		now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to save insight: %w", err)
	}

	return nil
}
