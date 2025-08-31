package email

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"
)

// RetentionRecommendation represents a policy recommendation
type RetentionRecommendation struct {
	ID                int64     `json:"id"`
	RecommendationType string   `json:"recommendation_type"`
	Priority          string    `json:"priority"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	CurrentState      string    `json:"current_state,omitempty"`
	RecommendedAction string    `json:"recommended_action"`
	ExpectedImpact    string    `json:"expected_impact"`
	ImpactScore       float64   `json:"impact_score"`
	ConfidenceScore   float64   `json:"confidence_score"`
	RiskLevel         string    `json:"risk_level"`
	UserID            *string   `json:"user_id,omitempty"`
	Domain            *string   `json:"domain,omitempty"`
	PolicyID          *int64    `json:"policy_id,omitempty"`
	Status            string    `json:"status"`
	AppliedAt         *time.Time `json:"applied_at,omitempty"`
	AppliedBy         *string   `json:"applied_by,omitempty"`
	AppliedResult     *string   `json:"applied_result,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// RecommendationAction represents a recommended action
type RecommendationAction struct {
	ActionType    string                 `json:"action_type"` // "create_policy", "update_policy", "delete_policy", "adjust_retention"
	PolicyID      *int64                 `json:"policy_id,omitempty"`
	NewSettings   map[string]interface{} `json:"new_settings"`
	Reasoning     string                 `json:"reasoning"`
	ExpectedSavings int64                `json:"expected_savings"`
}

// RecommendationImpact represents expected impact of a recommendation
type RecommendationImpact struct {
	StorageSavingsBytes int64   `json:"storage_savings_bytes"`
	CostSavingsUSD      float64 `json:"cost_savings_usd"`
	EmailsAffected      int     `json:"emails_affected"`
	PoliciesAffected    int     `json:"policies_affected"`
	RiskLevel           string  `json:"risk_level"`
	ImplementationTime  string  `json:"implementation_time"`
}

// UsagePattern represents email usage patterns for analysis
type UsagePattern struct {
	UserID           string  `json:"user_id"`
	Domain           string  `json:"domain"`
	AvgEmailSize     int64   `json:"avg_email_size"`
	EmailFrequency   float64 `json:"email_frequency"` // emails per day
	RetentionDays    int     `json:"retention_days"`
	ArchiveRate      float64 `json:"archive_rate"` // percentage archived vs deleted
	StorageUsage     int64   `json:"storage_usage"`
	RestoreFrequency int     `json:"restore_frequency"`
}

// RetentionAdvisorService provides methods for generating policy recommendations
type RetentionAdvisorService struct {
	db *sql.DB
}

// NewRetentionAdvisorService creates a new retention advisor service
func NewRetentionAdvisorService(db *sql.DB) *RetentionAdvisorService {
	return &RetentionAdvisorService{
		db: db,
	}
}

// GenerateRecommendations generates policy recommendations based on usage patterns
func (ras *RetentionAdvisorService) GenerateRecommendations(ctx context.Context) error {
	log.Println("Generating retention policy recommendations...")

	// Analyze usage patterns
	usagePatterns, err := ras.analyzeUsagePatterns(ctx)
	if err != nil {
		return fmt.Errorf("failed to analyze usage patterns: %w", err)
	}

	// Generate recommendations for each pattern
	for _, pattern := range usagePatterns {
		recommendations := ras.generateRecommendationsForPattern(ctx, pattern)
		
		for _, rec := range recommendations {
			if err := ras.saveRecommendation(ctx, rec); err != nil {
				log.Printf("Failed to save recommendation: %v", err)
				continue
			}
		}
	}

	// Generate system-wide recommendations
	systemRecommendations := ras.generateSystemRecommendations(ctx)
	for _, rec := range systemRecommendations {
		if err := ras.saveRecommendation(ctx, rec); err != nil {
			log.Printf("Failed to save system recommendation: %v", err)
			continue
		}
	}

	log.Printf("Generated %d recommendations", len(usagePatterns)*3+len(systemRecommendations))
	return nil
}

// GetRecommendations retrieves recommendations with filtering
func (ras *RetentionAdvisorService) GetRecommendations(ctx context.Context, filters map[string]string, limit, offset int) ([]*RetentionRecommendation, error) {
	query := `
		SELECT id, recommendation_type, priority, title, description, current_state,
		       recommended_action, expected_impact, impact_score, confidence_score,
		       risk_level, user_id, domain, policy_id, status, applied_at, applied_by,
		       applied_result, created_at, updated_at
		FROM retention_recommendations WHERE 1=1
	`

	var args []interface{}

	// Apply filters
	if recType, ok := filters["recommendation_type"]; ok {
		query += " AND recommendation_type = ?"
		args = append(args, recType)
	}

	if priority, ok := filters["priority"]; ok {
		query += " AND priority = ?"
		args = append(args, priority)
	}

	if status, ok := filters["status"]; ok {
		query += " AND status = ?"
		args = append(args, status)
	}

	if userID, ok := filters["user_id"]; ok {
		query += " AND user_id = ?"
		args = append(args, userID)
	}

	if domain, ok := filters["domain"]; ok {
		query += " AND domain = ?"
		args = append(args, domain)
	}

	query += " ORDER BY impact_score DESC, confidence_score DESC, created_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := ras.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query recommendations: %w", err)
	}
	defer rows.Close()

	var recommendations []*RetentionRecommendation
	for rows.Next() {
		rec := &RetentionRecommendation{}
		err := rows.Scan(
			&rec.ID, &rec.RecommendationType, &rec.Priority, &rec.Title, &rec.Description,
			&rec.CurrentState, &rec.RecommendedAction, &rec.ExpectedImpact,
			&rec.ImpactScore, &rec.ConfidenceScore, &rec.RiskLevel, &rec.UserID,
			&rec.Domain, &rec.PolicyID, &rec.Status, &rec.AppliedAt, &rec.AppliedBy,
			&rec.AppliedResult, &rec.CreatedAt, &rec.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation: %w", err)
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations, nil
}

// ApplyRecommendation applies a recommendation with preview mode support
func (ras *RetentionAdvisorService) ApplyRecommendation(ctx context.Context, recommendationID int64, appliedBy string, preview bool) (map[string]interface{}, error) {
	// Get recommendation
	rec, err := ras.getRecommendationByID(ctx, recommendationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation: %w", err)
	}

	if rec.Status != "pending" {
		return nil, fmt.Errorf("recommendation is not in pending status")
	}

	// Parse recommended action
	var action RecommendationAction
	if err := json.Unmarshal([]byte(rec.RecommendedAction), &action); err != nil {
		return nil, fmt.Errorf("failed to parse recommended action: %w", err)
	}

	// Apply the action
	result := map[string]interface{}{
		"recommendation_id": recommendationID,
		"preview_mode":      preview,
		"applied_by":        appliedBy,
		"applied_at":        time.Now(),
	}

	switch action.ActionType {
	case "create_policy":
		result["action"] = "create_policy"
		if !preview {
			if err := ras.createPolicyFromRecommendation(ctx, action, appliedBy); err != nil {
				result["status"] = "failed"
				result["error"] = err.Error()
				return result, err
			}
		}
		result["status"] = "success"

	case "update_policy":
		result["action"] = "update_policy"
		if !preview {
			if err := ras.updatePolicyFromRecommendation(ctx, action, appliedBy); err != nil {
				result["status"] = "failed"
				result["error"] = err.Error()
				return result, err
			}
		}
		result["status"] = "success"

	case "delete_policy":
		result["action"] = "delete_policy"
		if !preview {
			if err := ras.deletePolicyFromRecommendation(ctx, action, appliedBy); err != nil {
				result["status"] = "failed"
				result["error"] = err.Error()
				return result, err
			}
		}
		result["status"] = "success"

	default:
		return nil, fmt.Errorf("unknown action type: %s", action.ActionType)
	}

	// Update recommendation status if not preview
	if !preview {
		status := "applied"
		if result["status"] == "failed" {
			status = "failed"
		}
		
		if err := ras.updateRecommendationStatus(ctx, recommendationID, status, appliedBy, result["status"].(string)); err != nil {
			log.Printf("Failed to update recommendation status: %v", err)
		}
	}

	return result, nil
}

// analyzeUsagePatterns analyzes email usage patterns
func (ras *RetentionAdvisorService) analyzeUsagePatterns(ctx context.Context) ([]UsagePattern, error) {
	query := `
		SELECT 
			e.sender_id,
			SUBSTR(e.sender_id, INSTR(e.sender_id, '@') + 1) as domain,
			AVG(LENGTH(e.content)) as avg_email_size,
			COUNT(*) / 30.0 as email_frequency,
			COALESCE(p.retention_days, 30) as retention_days,
			SUM(CASE WHEN ae.id IS NOT NULL THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as archive_rate,
			SUM(LENGTH(e.content)) as storage_usage,
			COUNT(DISTINCT CASE WHEN ae.id IS NOT NULL THEN ae.id END) as restore_frequency
		FROM emails e
		LEFT JOIN retention_policies p ON (
			(p.user_id = e.sender_id OR p.user_id IS NULL) AND
			(p.sender_domain = SUBSTR(e.sender_id, INSTR(e.sender_id, '@') + 1) OR p.sender_domain IS NULL) AND
			p.active = 1
		)
		LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id
		WHERE e.created_at >= datetime('now', '-30 days')
		GROUP BY e.sender_id, domain
		HAVING COUNT(*) >= 5
	`

	rows, err := ras.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze usage patterns: %w", err)
	}
	defer rows.Close()

	var patterns []UsagePattern
	for rows.Next() {
		var pattern UsagePattern
		err := rows.Scan(
			&pattern.UserID, &pattern.Domain, &pattern.AvgEmailSize,
			&pattern.EmailFrequency, &pattern.RetentionDays, &pattern.ArchiveRate,
			&pattern.StorageUsage, &pattern.RestoreFrequency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage pattern: %w", err)
		}
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// generateRecommendationsForPattern generates recommendations for a specific usage pattern
func (ras *RetentionAdvisorService) generateRecommendationsForPattern(ctx context.Context, pattern UsagePattern) []*RetentionRecommendation {
	var recommendations []*RetentionRecommendation

	// Recommendation 1: Optimize retention days based on usage frequency
	if pattern.EmailFrequency > 10 && pattern.RetentionDays > 60 {
		// High frequency user with long retention - suggest shorter retention
		optimalRetention := int(math.Min(float64(pattern.RetentionDays), 30))
		rec := ras.createRetentionOptimizationRecommendation(pattern, optimalRetention)
		recommendations = append(recommendations, rec)
	} else if pattern.EmailFrequency < 2 && pattern.RetentionDays < 30 {
		// Low frequency user with short retention - suggest longer retention
		optimalRetention := int(math.Max(float64(pattern.RetentionDays), 60))
		rec := ras.createRetentionOptimizationRecommendation(pattern, optimalRetention)
		recommendations = append(recommendations, rec)
	}

	// Recommendation 2: Optimize archival strategy based on restore frequency
	if pattern.RestoreFrequency > 5 && pattern.ArchiveRate < 50 {
		// High restore frequency but low archive rate - suggest more archival
		rec := ras.createArchivalOptimizationRecommendation(pattern, true)
		recommendations = append(recommendations, rec)
	} else if pattern.RestoreFrequency == 0 && pattern.ArchiveRate > 80 {
		// No restores but high archive rate - suggest less archival
		rec := ras.createArchivalOptimizationRecommendation(pattern, false)
		recommendations = append(recommendations, rec)
	}

	// Recommendation 3: Storage optimization for large emails
	if pattern.AvgEmailSize > 1024*1024 { // > 1MB
		rec := ras.createStorageOptimizationRecommendation(pattern)
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// generateSystemRecommendations generates system-wide recommendations
func (ras *RetentionAdvisorService) generateSystemRecommendations(ctx context.Context) []*RetentionRecommendation {
	var recommendations []*RetentionRecommendation

	// Analyze under-utilized policies
	underutilizedPolicies, err := ras.findUnderutilizedPolicies(ctx)
	if err == nil && len(underutilizedPolicies) > 0 {
		rec := ras.createPolicyCleanupRecommendation(underutilizedPolicies)
		recommendations = append(recommendations, rec)
	}

	// Analyze high-risk patterns
	highRiskPatterns, err := ras.findHighRiskPatterns(ctx)
	if err == nil && len(highRiskPatterns) > 0 {
		rec := ras.createRiskMitigationRecommendation(highRiskPatterns)
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// createRetentionOptimizationRecommendation creates a retention optimization recommendation
func (ras *RetentionAdvisorService) createRetentionOptimizationRecommendation(pattern UsagePattern, optimalRetention int) *RetentionRecommendation {
	action := RecommendationAction{
		ActionType: "update_policy",
		NewSettings: map[string]interface{}{
			"retention_days": optimalRetention,
		},
		Reasoning:        fmt.Sprintf("User has %f emails/day, optimal retention is %d days", pattern.EmailFrequency, optimalRetention),
		ExpectedSavings:  int64(float64(pattern.StorageUsage) * 0.3), // Estimate 30% savings
	}

	actionJSON, _ := json.Marshal(action)
	impact := RecommendationImpact{
		StorageSavingsBytes: action.ExpectedSavings,
		CostSavingsUSD:      float64(action.ExpectedSavings) / (1024 * 1024 * 1024) * 0.023, // $0.023/GB/month
		EmailsAffected:      1,
		PoliciesAffected:    1,
		RiskLevel:           "low",
		ImplementationTime:  "5 minutes",
	}
	impactJSON, _ := json.Marshal(impact)

	return &RetentionRecommendation{
		RecommendationType: "policy_optimization",
		Priority:          "medium",
		Title:             fmt.Sprintf("Optimize Retention for %s", pattern.UserID),
		Description:       fmt.Sprintf("Adjust retention from %d to %d days based on usage patterns", pattern.RetentionDays, optimalRetention),
		RecommendedAction: string(actionJSON),
		ExpectedImpact:    string(impactJSON),
		ImpactScore:       0.7,
		ConfidenceScore:   0.8,
		RiskLevel:         "low",
		UserID:            &pattern.UserID,
		Status:            "pending",
	}
}

// createArchivalOptimizationRecommendation creates an archival optimization recommendation
func (ras *RetentionAdvisorService) createArchivalOptimizationRecommendation(pattern UsagePattern, increaseArchival bool) *RetentionRecommendation {
	action := RecommendationAction{
		ActionType: "update_policy",
		NewSettings: map[string]interface{}{
			"archive_instead": increaseArchival,
		},
		Reasoning:        fmt.Sprintf("Restore frequency: %d, current archive rate: %.1f%%", pattern.RestoreFrequency, pattern.ArchiveRate),
		ExpectedSavings:  int64(float64(pattern.StorageUsage) * 0.2), // Estimate 20% savings
	}

	actionJSON, _ := json.Marshal(action)
	impact := RecommendationImpact{
		StorageSavingsBytes: action.ExpectedSavings,
		CostSavingsUSD:      float64(action.ExpectedSavings) / (1024 * 1024 * 1024) * 0.023,
		EmailsAffected:      1,
		PoliciesAffected:    1,
		RiskLevel:           "low",
		ImplementationTime:  "5 minutes",
	}
	impactJSON, _ := json.Marshal(impact)

	title := "Increase Archival"
	description := "Increase archival rate to reduce storage costs"
	if !increaseArchival {
		title = "Decrease Archival"
		description = "Decrease archival rate to reduce restore overhead"
	}

	return &RetentionRecommendation{
		RecommendationType: "storage_optimization",
		Priority:          "medium",
		Title:             title,
		Description:       description,
		RecommendedAction: string(actionJSON),
		ExpectedImpact:    string(impactJSON),
		ImpactScore:       0.6,
		ConfidenceScore:   0.7,
		RiskLevel:         "low",
		UserID:            &pattern.UserID,
		Status:            "pending",
	}
}

// createStorageOptimizationRecommendation creates a storage optimization recommendation
func (ras *RetentionAdvisorService) createStorageOptimizationRecommendation(pattern UsagePattern) *RetentionRecommendation {
	action := RecommendationAction{
		ActionType: "create_policy",
		NewSettings: map[string]interface{}{
			"user_id":        pattern.UserID,
			"retention_days": 15, // Shorter retention for large emails
			"archive_instead": true,
		},
		Reasoning:        fmt.Sprintf("Large average email size: %d bytes", pattern.AvgEmailSize),
		ExpectedSavings:  int64(float64(pattern.StorageUsage) * 0.4), // Estimate 40% savings
	}

	actionJSON, _ := json.Marshal(action)
	impact := RecommendationImpact{
		StorageSavingsBytes: action.ExpectedSavings,
		CostSavingsUSD:      float64(action.ExpectedSavings) / (1024 * 1024 * 1024) * 0.023,
		EmailsAffected:      1,
		PoliciesAffected:    1,
		RiskLevel:           "medium",
		ImplementationTime:  "10 minutes",
	}
	impactJSON, _ := json.Marshal(impact)

	return &RetentionRecommendation{
		RecommendationType: "storage_optimization",
		Priority:          "high",
		Title:             "Optimize Storage for Large Emails",
		Description:       fmt.Sprintf("Create policy for user with large emails (avg: %d bytes)", pattern.AvgEmailSize),
		RecommendedAction: string(actionJSON),
		ExpectedImpact:    string(impactJSON),
		ImpactScore:       0.8,
		ConfidenceScore:   0.9,
		RiskLevel:         "medium",
		UserID:            &pattern.UserID,
		Status:            "pending",
	}
}

// findUnderutilizedPolicies finds policies that are rarely used
func (ras *RetentionAdvisorService) findUnderutilizedPolicies(ctx context.Context) ([]int64, error) {
	query := `
		SELECT p.id FROM retention_policies p
		LEFT JOIN policy_evaluation_logs pel ON p.id = pel.policy_id
		WHERE p.active = 1
		GROUP BY p.id
		HAVING COUNT(pel.id) < 10
		ORDER BY COUNT(pel.id) ASC
		LIMIT 5
	`

	rows, err := ras.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find underutilized policies: %w", err)
	}
	defer rows.Close()

	var policyIDs []int64
	for rows.Next() {
		var policyID int64
		if err := rows.Scan(&policyID); err != nil {
			return nil, fmt.Errorf("failed to scan policy ID: %w", err)
		}
		policyIDs = append(policyIDs, policyID)
	}

	return policyIDs, nil
}

// findHighRiskPatterns finds high-risk usage patterns
func (ras *RetentionAdvisorService) findHighRiskPatterns(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT e.sender_id FROM emails e
		JOIN archived_emails ae ON e.email_id = ae.original_email_id
		WHERE ae.archived_at >= datetime('now', '-7 days')
		GROUP BY e.sender_id
		HAVING COUNT(ae.id) > 50
	`

	rows, err := ras.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find high-risk patterns: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// createPolicyCleanupRecommendation creates a policy cleanup recommendation
func (ras *RetentionAdvisorService) createPolicyCleanupRecommendation(policyIDs []int64) *RetentionRecommendation {
	action := RecommendationAction{
		ActionType: "delete_policy",
		NewSettings: map[string]interface{}{
			"policy_ids": policyIDs,
		},
		Reasoning:        fmt.Sprintf("Found %d underutilized policies", len(policyIDs)),
		ExpectedSavings:  0, // No direct storage savings, but reduces complexity
	}

	actionJSON, _ := json.Marshal(action)
	impact := RecommendationImpact{
		StorageSavingsBytes: 0,
		CostSavingsUSD:      0,
		EmailsAffected:      0,
		PoliciesAffected:    len(policyIDs),
		RiskLevel:           "low",
		ImplementationTime:  "15 minutes",
	}
	impactJSON, _ := json.Marshal(impact)

	return &RetentionRecommendation{
		RecommendationType: "policy_optimization",
		Priority:          "low",
		Title:             "Clean Up Underutilized Policies",
		Description:       fmt.Sprintf("Remove %d policies that are rarely used", len(policyIDs)),
		RecommendedAction: string(actionJSON),
		ExpectedImpact:    string(impactJSON),
		ImpactScore:       0.3,
		ConfidenceScore:   0.9,
		RiskLevel:         "low",
		Status:            "pending",
	}
}

// createRiskMitigationRecommendation creates a risk mitigation recommendation
func (ras *RetentionAdvisorService) createRiskMitigationRecommendation(userIDs []string) *RetentionRecommendation {
	action := RecommendationAction{
		ActionType: "create_policy",
		NewSettings: map[string]interface{}{
			"user_ids":       userIDs,
			"retention_days": 7, // Shorter retention for high-risk users
			"archive_instead": true,
		},
		Reasoning:        fmt.Sprintf("Found %d users with excessive archival", len(userIDs)),
		ExpectedSavings:  int64(len(userIDs) * 1024 * 1024 * 100), // Estimate 100MB per user
	}

	actionJSON, _ := json.Marshal(action)
	impact := RecommendationImpact{
		StorageSavingsBytes: action.ExpectedSavings,
		CostSavingsUSD:      float64(action.ExpectedSavings) / (1024 * 1024 * 1024) * 0.023,
		EmailsAffected:      len(userIDs),
		PoliciesAffected:    1,
		RiskLevel:           "high",
		ImplementationTime:  "20 minutes",
	}
	impactJSON, _ := json.Marshal(impact)

	return &RetentionRecommendation{
		RecommendationType: "risk_mitigation",
		Priority:          "critical",
		Title:             "Mitigate High-Risk Usage Patterns",
		Description:       fmt.Sprintf("Create restrictive policies for %d high-risk users", len(userIDs)),
		RecommendedAction: string(actionJSON),
		ExpectedImpact:    string(impactJSON),
		ImpactScore:       0.9,
		ConfidenceScore:   0.8,
		RiskLevel:         "high",
		Status:            "pending",
	}
}

// Helper methods for applying recommendations
func (ras *RetentionAdvisorService) createPolicyFromRecommendation(ctx context.Context, action RecommendationAction, appliedBy string) error {
	// Implementation would create a new policy based on the action
	log.Printf("Creating policy from recommendation: %+v", action)
	return nil
}

func (ras *RetentionAdvisorService) updatePolicyFromRecommendation(ctx context.Context, action RecommendationAction, appliedBy string) error {
	// Implementation would update an existing policy based on the action
	log.Printf("Updating policy from recommendation: %+v", action)
	return nil
}

func (ras *RetentionAdvisorService) deletePolicyFromRecommendation(ctx context.Context, action RecommendationAction, appliedBy string) error {
	// Implementation would delete a policy based on the action
	log.Printf("Deleting policy from recommendation: %+v", action)
	return nil
}

// Database helper methods
func (ras *RetentionAdvisorService) saveRecommendation(ctx context.Context, rec *RetentionRecommendation) error {
	query := `
		INSERT INTO retention_recommendations (
			recommendation_type, priority, title, description, current_state,
			recommended_action, expected_impact, impact_score, confidence_score,
			risk_level, user_id, domain, policy_id, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := ras.db.ExecContext(ctx, query,
		rec.RecommendationType, rec.Priority, rec.Title, rec.Description,
		rec.CurrentState, rec.RecommendedAction, rec.ExpectedImpact,
		rec.ImpactScore, rec.ConfidenceScore, rec.RiskLevel, rec.UserID,
		rec.Domain, rec.PolicyID, rec.Status, now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to save recommendation: %w", err)
	}

	return nil
}

func (ras *RetentionAdvisorService) getRecommendationByID(ctx context.Context, id int64) (*RetentionRecommendation, error) {
	query := `
		SELECT id, recommendation_type, priority, title, description, current_state,
		       recommended_action, expected_impact, impact_score, confidence_score,
		       risk_level, user_id, domain, policy_id, status, applied_at, applied_by,
		       applied_result, created_at, updated_at
		FROM retention_recommendations WHERE id = ?
	`

	rec := &RetentionRecommendation{}
	err := ras.db.QueryRowContext(ctx, query, id).Scan(
		&rec.ID, &rec.RecommendationType, &rec.Priority, &rec.Title, &rec.Description,
		&rec.CurrentState, &rec.RecommendedAction, &rec.ExpectedImpact,
		&rec.ImpactScore, &rec.ConfidenceScore, &rec.RiskLevel, &rec.UserID,
		&rec.Domain, &rec.PolicyID, &rec.Status, &rec.AppliedAt, &rec.AppliedBy,
		&rec.AppliedResult, &rec.CreatedAt, &rec.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("recommendation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation: %w", err)
	}

	return rec, nil
}

func (ras *RetentionAdvisorService) updateRecommendationStatus(ctx context.Context, id int64, status, appliedBy, result string) error {
	query := `
		UPDATE retention_recommendations SET
			status = ?, applied_at = ?, applied_by = ?, applied_result = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := ras.db.ExecContext(ctx, query, status, now, appliedBy, result, now, id)
	if err != nil {
		return fmt.Errorf("failed to update recommendation status: %w", err)
	}

	return nil
}

















