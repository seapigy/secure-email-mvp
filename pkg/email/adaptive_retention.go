package email

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

// AdaptivePolicyChange represents a proposed or applied policy change
type AdaptivePolicyChange struct {
	ID          int64     `json:"id"`
	PolicyID    int64     `json:"policy_id"`
	
	// Change details
	ChangeType      string  `json:"change_type"`       // "retention_days", "archive_instead", "archive_retention_days"
	OldValue        string  `json:"old_value"`         // Previous value
	NewValue        string  `json:"new_value"`         // New value
	ChangeReason    string  `json:"change_reason"`     // Reason for the change
	ChangePercentage float64 `json:"change_percentage"` // Percentage change
	
	// Impact analysis
	ExpectedStorageSavingsBytes int64   `json:"expected_storage_savings_bytes"`
	ExpectedArchivalLoadImpact  float64 `json:"expected_archival_load_impact"`
	RiskAssessment             string  `json:"risk_assessment"` // "low", "medium", "high"
	
	// Safety controls
	CooldownUntil        *time.Time `json:"cooldown_until,omitempty"`
	RequiresAdminApproval bool      `json:"requires_admin_approval"`
	
	// Status tracking
	Status        string     `json:"status"`         // "pending", "approved", "applied", "rejected"
	AppliedAt     *time.Time `json:"applied_at,omitempty"`
	AppliedBy     string     `json:"applied_by"`     // "system" or admin user
	AppliedResult string     `json:"applied_result"` // "success", "partial", "failed"
	
	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AdaptivePolicyConfig represents configuration for adaptive adjustments
type AdaptivePolicyConfig struct {
	ID          int64 `json:"id"`
	PolicyID    int64 `json:"policy_id"`
	
	// Configuration settings
	AdaptiveEnabled        bool    `json:"adaptive_enabled"`
	MaxChangePercentage    float64 `json:"max_change_percentage"`
	CooldownDays           int     `json:"cooldown_days"`
	RequiresAdminApproval  bool    `json:"requires_admin_approval"`
	
	// Thresholds
	MinRetentionDays        int `json:"min_retention_days"`
	MaxRetentionDays        int `json:"max_retention_days"`
	MinArchiveRetentionDays int `json:"min_archive_retention_days"`
	MaxArchiveRetentionDays int `json:"max_archive_retention_days"`
	
	// Safety limits
	MaxStorageImpactBytes   int64   `json:"max_storage_impact_bytes"`
	MaxArchivalLoadImpact   float64 `json:"max_archival_load_impact"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PolicyPerformanceMetrics represents performance metrics for a policy
type PolicyPerformanceMetrics struct {
	PolicyID                int64   `json:"policy_id"`
	TotalEvaluations        int     `json:"total_evaluations"`
	MatchRate               float64 `json:"match_rate"`
	ApplicationRate         float64 `json:"application_rate"`
	AvgImpactScore          float64 `json:"avg_impact_score"`
	AvgArchivalLoadImpact   float64 `json:"avg_archival_load_impact"`
	StorageSavingsBytes     int64   `json:"storage_savings_bytes"`
	ArchivalLoadPercentage  float64 `json:"archival_load_percentage"`
	LastEvaluationTime      time.Time `json:"last_evaluation_time"`
}

// AdaptiveRetentionEngine provides adaptive policy enforcement
type AdaptiveRetentionEngine struct {
	db                    *sql.DB
	monitorService        *RetentionMonitorService
	policyEngine          *RetentionPolicyEngine
	adaptiveEnabled       bool
	maxChangePercentage   float64
	cooldownDays          int
	requiresAdminApproval bool
}

// NewAdaptiveRetentionEngine creates a new adaptive retention engine
func NewAdaptiveRetentionEngine(db *sql.DB, monitorService *RetentionMonitorService, policyEngine *RetentionPolicyEngine) *AdaptiveRetentionEngine {
	return &AdaptiveRetentionEngine{
		db:                    db,
		monitorService:        monitorService,
		policyEngine:          policyEngine,
		adaptiveEnabled:       getEnvBool("ENABLE_ADAPTIVE_POLICY_ENFORCEMENT", false),
		maxChangePercentage:   getEnvFloat("ADAPTIVE_POLICY_MAX_CHANGE_PERCENT", 20.0),
		cooldownDays:          getEnvInt("ADAPTIVE_POLICY_COOLDOWN_DAYS", 7),
		requiresAdminApproval: true, // Default to requiring admin approval for safety
	}
}

// Start begins the adaptive retention engine
func (are *AdaptiveRetentionEngine) Start(ctx context.Context) {
	if !are.adaptiveEnabled {
		log.Println("Adaptive Retention Engine is disabled")
		return
	}
	
	log.Println("Starting Adaptive Retention Engine...")
	
	// Start periodic analysis goroutine
	go are.runPeriodicAnalysis(ctx)
	
	log.Println("Adaptive Retention Engine started successfully")
}

// AnalyzePolicyPerformance analyzes the performance of a specific policy
func (are *AdaptiveRetentionEngine) AnalyzePolicyPerformance(ctx context.Context, policyID int64) (*PolicyPerformanceMetrics, error) {
	query := `
		SELECT 
			pel.policy_id,
			COUNT(*) as total_evaluations,
			SUM(CASE WHEN pel.evaluation_result = 'matched' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as match_rate,
			SUM(CASE WHEN pel.evaluation_result = 'applied' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as application_rate,
			AVG(pel.impact_score) as avg_impact_score,
			AVG(pel.archival_load_impact) as avg_archival_load_impact,
			SUM(pel.storage_savings_bytes) as storage_savings_bytes,
			MAX(pel.evaluated_at) as last_evaluation_time
		FROM policy_evaluation_logs pel
		WHERE pel.policy_id = ?
		GROUP BY pel.policy_id
	`
	
	var metrics PolicyPerformanceMetrics
	err := are.db.QueryRowContext(ctx, query, policyID).Scan(
		&metrics.PolicyID, &metrics.TotalEvaluations, &metrics.MatchRate,
		&metrics.ApplicationRate, &metrics.AvgImpactScore, &metrics.AvgArchivalLoadImpact,
		&metrics.StorageSavingsBytes, &metrics.LastEvaluationTime,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no performance data found for policy %d", policyID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to analyze policy performance: %w", err)
	}
	
	// Calculate archival load percentage
	metrics.ArchivalLoadPercentage = are.calculateArchivalLoadPercentage(ctx, policyID)
	
	return &metrics, nil
}

// GenerateAdaptiveRecommendations generates adaptive policy recommendations
func (are *AdaptiveRetentionEngine) GenerateAdaptiveRecommendations(ctx context.Context) ([]*AdaptivePolicyChange, error) {
	if !are.adaptiveEnabled {
		return nil, fmt.Errorf("adaptive policy enforcement is disabled")
	}
	
	log.Println("Generating adaptive policy recommendations...")
	
	// Get all active policies with adaptive configuration
	policies, err := are.getAdaptivePolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get adaptive policies: %w", err)
	}
	
	var recommendations []*AdaptivePolicyChange
	
	for _, policy := range policies {
		// Check if policy is in cooldown
		if are.isPolicyInCooldown(ctx, policy.ID) {
			continue
		}
		
		// Analyze policy performance
		metrics, err := are.AnalyzePolicyPerformance(ctx, policy.ID)
		if err != nil {
			log.Printf("Failed to analyze policy %d performance: %v", policy.ID, err)
			continue
		}
		
		// Generate recommendations based on performance
		policyRecommendations := are.generatePolicyRecommendations(ctx, policy, metrics)
		recommendations = append(recommendations, policyRecommendations...)
	}
	
	log.Printf("Generated %d adaptive recommendations", len(recommendations))
	return recommendations, nil
}

// ApplyAdaptiveChange applies an adaptive policy change
func (are *AdaptiveRetentionEngine) ApplyAdaptiveChange(ctx context.Context, changeID int64, appliedBy string, preview bool) (map[string]interface{}, error) {
	// Get the change
	change, err := are.getAdaptiveChange(ctx, changeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adaptive change: %w", err)
	}
	
	// Check if change is already applied
	if change.Status == "applied" {
		return nil, fmt.Errorf("change is already applied")
	}
	
	// Check if admin approval is required
	if change.RequiresAdminApproval && appliedBy == "system" {
		return nil, fmt.Errorf("admin approval required for this change")
	}
	
	// Get the policy
	policy, err := are.policyEngine.GetPolicyByID(ctx, change.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	
	if preview {
		// Return preview of what would happen
		return are.generateChangePreview(ctx, change, policy)
	}
	
	// Apply the change
	result := are.applyPolicyChange(ctx, change, policy)
	
	// Update change status
	if err := are.updateChangeStatus(ctx, changeID, "applied", appliedBy, result); err != nil {
		return nil, fmt.Errorf("failed to update change status: %w", err)
	}
	
	// Set cooldown period
	if err := are.setPolicyCooldown(ctx, change.PolicyID, are.cooldownDays); err != nil {
		log.Printf("Failed to set policy cooldown: %v", err)
	}
	
	return map[string]interface{}{
		"change_id":      changeID,
		"policy_id":      change.PolicyID,
		"change_type":    change.ChangeType,
		"old_value":      change.OldValue,
		"new_value":      change.NewValue,
		"applied_by":     appliedBy,
		"applied_result": result,
		"applied_at":     time.Now(),
	}, nil
}

// GetAdaptiveChanges retrieves adaptive policy changes with filtering
func (are *AdaptiveRetentionEngine) GetAdaptiveChanges(ctx context.Context, filters map[string]string, limit, offset int) ([]*AdaptivePolicyChange, error) {
	query := `
		SELECT id, policy_id, change_type, old_value, new_value, change_reason, change_percentage,
		       expected_storage_savings_bytes, expected_archival_load_impact, risk_assessment,
		       cooldown_until, requires_admin_approval, status, applied_at, applied_by, applied_result,
		       created_at, updated_at
		FROM adaptive_policy_changes
		WHERE 1=1
	`
	
	var args []interface{}
	
	// Add filters
	if policyID, ok := filters["policy_id"]; ok {
		query += " AND policy_id = ?"
		args = append(args, policyID)
	}
	
	if status, ok := filters["status"]; ok {
		query += " AND status = ?"
		args = append(args, status)
	}
	
	if changeType, ok := filters["change_type"]; ok {
		query += " AND change_type = ?"
		args = append(args, changeType)
	}
	
	query += " ORDER BY created_at DESC"
	
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}
	
	rows, err := are.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query adaptive changes: %w", err)
	}
	defer rows.Close()
	
	var changes []*AdaptivePolicyChange
	for rows.Next() {
		var change AdaptivePolicyChange
		var cooldownUntil, appliedAt sql.NullTime
		
		err := rows.Scan(
			&change.ID, &change.PolicyID, &change.ChangeType, &change.OldValue, &change.NewValue,
			&change.ChangeReason, &change.ChangePercentage, &change.ExpectedStorageSavingsBytes,
			&change.ExpectedArchivalLoadImpact, &change.RiskAssessment, &cooldownUntil,
			&change.RequiresAdminApproval, &change.Status, &appliedAt, &change.AppliedBy,
			&change.AppliedResult, &change.CreatedAt, &change.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan adaptive change: %w", err)
		}
		
		if cooldownUntil.Valid {
			change.CooldownUntil = &cooldownUntil.Time
		}
		if appliedAt.Valid {
			change.AppliedAt = &appliedAt.Time
		}
		
		changes = append(changes, &change)
	}
	
	return changes, nil
}

// EnableAdaptivePolicy enables adaptive adjustments for a specific policy
func (are *AdaptiveRetentionEngine) EnableAdaptivePolicy(ctx context.Context, policyID int64, config *AdaptivePolicyConfig) error {
	query := `
		INSERT OR REPLACE INTO adaptive_policy_config (
			policy_id, adaptive_enabled, max_change_percentage, cooldown_days,
			requires_admin_approval, min_retention_days, max_retention_days,
			min_archive_retention_days, max_archive_retention_days,
			max_storage_impact_bytes, max_archival_load_impact, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	_, err := are.db.ExecContext(ctx, query,
		policyID, config.AdaptiveEnabled, config.MaxChangePercentage, config.CooldownDays,
		config.RequiresAdminApproval, config.MinRetentionDays, config.MaxRetentionDays,
		config.MinArchiveRetentionDays, config.MaxArchiveRetentionDays,
		config.MaxStorageImpactBytes, config.MaxArchivalLoadImpact, now, now,
	)
	
	if err != nil {
		return fmt.Errorf("failed to enable adaptive policy: %w", err)
	}
	
	log.Printf("Enabled adaptive policy for policy ID %d", policyID)
	return nil
}

// DisableAdaptivePolicy disables adaptive adjustments for a specific policy
func (are *AdaptiveRetentionEngine) DisableAdaptivePolicy(ctx context.Context, policyID int64) error {
	query := `UPDATE adaptive_policy_config SET adaptive_enabled = 0, updated_at = ? WHERE policy_id = ?`
	_, err := are.db.ExecContext(ctx, query, time.Now(), policyID)
	if err != nil {
		return fmt.Errorf("failed to disable adaptive policy: %w", err)
	}
	
	log.Printf("Disabled adaptive policy for policy ID %d", policyID)
	return nil
}

// runPeriodicAnalysis runs periodic analysis for adaptive recommendations
func (are *AdaptiveRetentionEngine) runPeriodicAnalysis(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if are.adaptiveEnabled {
				recommendations, err := are.GenerateAdaptiveRecommendations(ctx)
				if err != nil {
					log.Printf("Failed to generate adaptive recommendations: %v", err)
					continue
				}
				
				// Store recommendations
				for _, rec := range recommendations {
					if err := are.storeAdaptiveChange(ctx, rec); err != nil {
						log.Printf("Failed to store adaptive change: %v", err)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// Helper methods
func (are *AdaptiveRetentionEngine) getAdaptivePolicies(ctx context.Context) ([]*RetentionPolicy, error) {
	query := `
		SELECT p.id, p.name, p.description, p.priority, p.active, p.user_id, p.sender_domain,
		       p.recipient_domain, p.email_status, p.custom_tags, p.min_age_hours, p.max_age_hours,
		       p.retention_days, p.archive_instead, p.archive_retention_days, p.created_by,
		       p.created_at, p.updated_at
		FROM retention_policies p
		JOIN adaptive_policy_config apc ON p.id = apc.policy_id
		WHERE p.active = 1 AND apc.adaptive_enabled = 1
		ORDER BY p.priority DESC
	`
	
	rows, err := are.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query adaptive policies: %w", err)
	}
	defer rows.Close()
	
	var policies []*RetentionPolicy
	for rows.Next() {
		var policy RetentionPolicy
		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Active,
			&policy.UserID, &policy.SenderDomain, &policy.RecipientDomain, &policy.EmailStatus,
			&policy.CustomTags, &policy.MinAgeHours, &policy.MaxAgeHours, &policy.RetentionDays,
			&policy.ArchiveInstead, &policy.ArchiveRetentionDays, &policy.CreatedBy,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, &policy)
	}
	
	return policies, nil
}

func (are *AdaptiveRetentionEngine) isPolicyInCooldown(ctx context.Context, policyID int64) bool {
	query := `
		SELECT COUNT(*) FROM adaptive_policy_changes
		WHERE policy_id = ? AND cooldown_until > datetime('now')
	`
	
	var count int
	err := are.db.QueryRowContext(ctx, query, policyID).Scan(&count)
	if err != nil {
		log.Printf("Failed to check policy cooldown: %v", err)
		return false
	}
	
	return count > 0
}

func (are *AdaptiveRetentionEngine) generatePolicyRecommendations(ctx context.Context, policy *RetentionPolicy, metrics *PolicyPerformanceMetrics) []*AdaptivePolicyChange {
	var recommendations []*AdaptivePolicyChange
	
	// Get adaptive config for this policy
	config, err := are.getAdaptiveConfig(ctx, policy.ID)
	if err != nil {
		log.Printf("Failed to get adaptive config for policy %d: %v", policy.ID, err)
		return recommendations
	}
	
	// Check if archival load is too high
	if metrics.ArchivalLoadPercentage > 80.0 {
		// Recommend reducing retention days
		if policy.RetentionDays > config.MinRetentionDays {
			newRetentionDays := int(math.Max(float64(config.MinRetentionDays), float64(policy.RetentionDays)*0.9))
			changePercentage := float64(policy.RetentionDays-newRetentionDays) / float64(policy.RetentionDays) * 100
			
			if changePercentage <= config.MaxChangePercentage {
				recommendations = append(recommendations, &AdaptivePolicyChange{
					PolicyID:              policy.ID,
					ChangeType:            "retention_days",
					OldValue:              fmt.Sprintf("%d", policy.RetentionDays),
					NewValue:              fmt.Sprintf("%d", newRetentionDays),
					ChangeReason:          "High archival load detected",
					ChangePercentage:      changePercentage,
					RiskAssessment:        "medium",
					RequiresAdminApproval: config.RequiresAdminApproval,
					Status:                "pending",
					CreatedAt:             time.Now(),
					UpdatedAt:             time.Now(),
				})
			}
		}
	}
	
	// Check if archival load is very low and storage is high
	if metrics.ArchivalLoadPercentage < 20.0 && metrics.StorageSavingsBytes > config.MaxStorageImpactBytes/2 {
		// Recommend increasing retention days
		if policy.RetentionDays < config.MaxRetentionDays {
			newRetentionDays := int(math.Min(float64(config.MaxRetentionDays), float64(policy.RetentionDays)*1.1))
			changePercentage := float64(newRetentionDays-policy.RetentionDays) / float64(policy.RetentionDays) * 100
			
			if changePercentage <= config.MaxChangePercentage {
				recommendations = append(recommendations, &AdaptivePolicyChange{
					PolicyID:              policy.ID,
					ChangeType:            "retention_days",
					OldValue:              fmt.Sprintf("%d", policy.RetentionDays),
					NewValue:              fmt.Sprintf("%d", newRetentionDays),
					ChangeReason:          "Low archival load, high storage usage",
					ChangePercentage:      changePercentage,
					RiskAssessment:        "low",
					RequiresAdminApproval: config.RequiresAdminApproval,
					Status:                "pending",
					CreatedAt:             time.Now(),
					UpdatedAt:             time.Now(),
				})
			}
		}
	}
	
	// Check if policy is underutilized
	if metrics.MatchRate < 10.0 && metrics.TotalEvaluations > 100 {
		// Recommend switching to archival instead of deletion
		if !policy.ArchiveInstead {
			recommendations = append(recommendations, &AdaptivePolicyChange{
				PolicyID:              policy.ID,
				ChangeType:            "archive_instead",
				OldValue:              "false",
				NewValue:              "true",
				ChangeReason:          "Policy underutilized, switching to archival",
				ChangePercentage:      100.0,
				RiskAssessment:        "low",
				RequiresAdminApproval: config.RequiresAdminApproval,
				Status:                "pending",
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			})
		}
	}
	
	return recommendations
}

func (are *AdaptiveRetentionEngine) getAdaptiveConfig(ctx context.Context, policyID int64) (*AdaptivePolicyConfig, error) {
	query := `
		SELECT id, policy_id, adaptive_enabled, max_change_percentage, cooldown_days,
		       requires_admin_approval, min_retention_days, max_retention_days,
		       min_archive_retention_days, max_archive_retention_days,
		       max_storage_impact_bytes, max_archival_load_impact, created_at, updated_at
		FROM adaptive_policy_config
		WHERE policy_id = ?
	`
	
	var config AdaptivePolicyConfig
	err := are.db.QueryRowContext(ctx, query, policyID).Scan(
		&config.ID, &config.PolicyID, &config.AdaptiveEnabled, &config.MaxChangePercentage,
		&config.CooldownDays, &config.RequiresAdminApproval, &config.MinRetentionDays,
		&config.MaxRetentionDays, &config.MinArchiveRetentionDays, &config.MaxArchiveRetentionDays,
		&config.MaxStorageImpactBytes, &config.MaxArchivalLoadImpact, &config.CreatedAt, &config.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		// Return default config
		return &AdaptivePolicyConfig{
			PolicyID:                policyID,
			AdaptiveEnabled:         false,
			MaxChangePercentage:     are.maxChangePercentage,
			CooldownDays:            are.cooldownDays,
			RequiresAdminApproval:   are.requiresAdminApproval,
			MinRetentionDays:        1,
			MaxRetentionDays:        365,
			MinArchiveRetentionDays: 30,
			MaxArchiveRetentionDays: 2555,
			MaxStorageImpactBytes:   1073741824, // 1GB
			MaxArchivalLoadImpact:   0.5,
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get adaptive config: %w", err)
	}
	
	return &config, nil
}

func (are *AdaptiveRetentionEngine) calculateArchivalLoadPercentage(ctx context.Context, policyID int64) float64 {
	query := `
		SELECT AVG(archival_load_impact) * 100.0
		FROM policy_evaluation_logs
		WHERE policy_id = ? AND archival_load_impact > 0
		AND evaluated_at > datetime('now', '-7 days')
	`
	
	var percentage float64
	err := are.db.QueryRowContext(ctx, query, policyID).Scan(&percentage)
	if err != nil {
		return 0.0
	}
	
	return percentage
}

func (are *AdaptiveRetentionEngine) storeAdaptiveChange(ctx context.Context, change *AdaptivePolicyChange) error {
	query := `
		INSERT INTO adaptive_policy_changes (
			policy_id, change_type, old_value, new_value, change_reason, change_percentage,
			expected_storage_savings_bytes, expected_archival_load_impact, risk_assessment,
			cooldown_until, requires_admin_approval, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	cooldownUntil := time.Now().AddDate(0, 0, are.cooldownDays)
	
	_, err := are.db.ExecContext(ctx, query,
		change.PolicyID, change.ChangeType, change.OldValue, change.NewValue,
		change.ChangeReason, change.ChangePercentage, change.ExpectedStorageSavingsBytes,
		change.ExpectedArchivalLoadImpact, change.RiskAssessment, cooldownUntil,
		change.RequiresAdminApproval, change.Status, change.CreatedAt, change.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store adaptive change: %w", err)
	}
	
	return nil
}

func (are *AdaptiveRetentionEngine) getAdaptiveChange(ctx context.Context, changeID int64) (*AdaptivePolicyChange, error) {
	query := `
		SELECT id, policy_id, change_type, old_value, new_value, change_reason, change_percentage,
		       expected_storage_savings_bytes, expected_archival_load_impact, risk_assessment,
		       cooldown_until, requires_admin_approval, status, applied_at, applied_by, applied_result,
		       created_at, updated_at
		FROM adaptive_policy_changes
		WHERE id = ?
	`
	
	var change AdaptivePolicyChange
	var cooldownUntil, appliedAt sql.NullTime
	
	err := are.db.QueryRowContext(ctx, query, changeID).Scan(
		&change.ID, &change.PolicyID, &change.ChangeType, &change.OldValue, &change.NewValue,
		&change.ChangeReason, &change.ChangePercentage, &change.ExpectedStorageSavingsBytes,
		&change.ExpectedArchivalLoadImpact, &change.RiskAssessment, &cooldownUntil,
		&change.RequiresAdminApproval, &change.Status, &appliedAt, &change.AppliedBy,
		&change.AppliedResult, &change.CreatedAt, &change.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("adaptive change with ID %d not found", changeID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get adaptive change: %w", err)
	}
	
	if cooldownUntil.Valid {
		change.CooldownUntil = &cooldownUntil.Time
	}
	if appliedAt.Valid {
		change.AppliedAt = &appliedAt.Time
	}
	
	return &change, nil
}

func (are *AdaptiveRetentionEngine) generateChangePreview(ctx context.Context, change *AdaptivePolicyChange, policy *RetentionPolicy) (map[string]interface{}, error) {
	// Calculate expected impact
	expectedSavings := int64(0)
	expectedLoadImpact := 0.0
	
	switch change.ChangeType {
	case "retention_days":
		oldDays, _ := strconv.Atoi(change.OldValue)
		newDays, _ := strconv.Atoi(change.NewValue)
		expectedSavings = int64(oldDays-newDays) * 1024 * 1024 // Rough estimate: 1MB per day
		expectedLoadImpact = float64(oldDays-newDays) / float64(oldDays) * 100
	case "archive_instead":
		expectedSavings = 0 // No immediate savings, but better long-term retention
		expectedLoadImpact = 10.0 // Moderate increase in archival load
	}
	
	return map[string]interface{}{
		"change_id":                    change.ID,
		"policy_id":                    change.PolicyID,
		"policy_name":                  policy.Name,
		"change_type":                  change.ChangeType,
		"old_value":                    change.OldValue,
		"new_value":                    change.NewValue,
		"change_reason":                change.ChangeReason,
		"change_percentage":            change.ChangePercentage,
		"expected_storage_savings":     expectedSavings,
		"expected_archival_load_impact": expectedLoadImpact,
		"risk_assessment":              change.RiskAssessment,
		"requires_admin_approval":      change.RequiresAdminApproval,
		"preview":                      true,
	}, nil
}

func (are *AdaptiveRetentionEngine) applyPolicyChange(ctx context.Context, change *AdaptivePolicyChange, policy *RetentionPolicy) string {
	// Apply the change to the policy
	switch change.ChangeType {
	case "retention_days":
		newDays, err := strconv.Atoi(change.NewValue)
		if err != nil {
			return "failed"
		}
		policy.RetentionDays = newDays
		
	case "archive_instead":
		archiveInstead := change.NewValue == "true"
		policy.ArchiveInstead = archiveInstead
		
	case "archive_retention_days":
		newDays, err := strconv.Atoi(change.NewValue)
		if err != nil {
			return "failed"
		}
		policy.ArchiveRetentionDays = newDays
	}
	
	// Update the policy in the database
	if err := are.policyEngine.UpdatePolicy(ctx, policy); err != nil {
		log.Printf("Failed to update policy %d: %v", policy.ID, err)
		return "failed"
	}
	
	return "success"
}

func (are *AdaptiveRetentionEngine) updateChangeStatus(ctx context.Context, changeID int64, status, appliedBy, result string) error {
	query := `
		UPDATE adaptive_policy_changes
		SET status = ?, applied_by = ?, applied_result = ?, applied_at = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := are.db.ExecContext(ctx, query, status, appliedBy, result, now, now, changeID)
	if err != nil {
		return fmt.Errorf("failed to update change status: %w", err)
	}
	
	return nil
}

func (are *AdaptiveRetentionEngine) setPolicyCooldown(ctx context.Context, policyID int64, cooldownDays int) error {
	query := `
		UPDATE adaptive_policy_changes
		SET cooldown_until = datetime('now', '+' || ? || ' days')
		WHERE policy_id = ? AND status = 'applied'
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	_, err := are.db.ExecContext(ctx, query, cooldownDays, policyID)
	if err != nil {
		return fmt.Errorf("failed to set policy cooldown: %w", err)
	}
	
	return nil
}

// Environment variable helpers
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
