package quota

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// QUOTA MANAGEMENT SERVICE
// =============================================================================

// QuotaService handles user and domain-based quota enforcement
type QuotaService struct {
	db *sql.DB
}

// Quota represents a quota configuration
type Quota struct {
	ID         string    `json:"id" db:"id"`
	EntityType string    `json:"entity_type" db:"entity_type"` // "user", "domain", "global"
	EntityID   string    `json:"entity_id" db:"entity_id"`
	QuotaType  string    `json:"quota_type" db:"quota_type"` // "email_send", "link_create", "storage", "bandwidth"
	Limit      int64     `json:"limit" db:"limit"`
	Period     string    `json:"period" db:"period"` // "minute", "hour", "day", "month"
	ResetAt    time.Time `json:"reset_at" db:"reset_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	IsActive   bool      `json:"is_active" db:"is_active"`
}

// QuotaUsage represents current quota usage
type QuotaUsage struct {
	ID          string    `json:"id" db:"id"`
	QuotaID     string    `json:"quota_id" db:"quota_id"`
	EntityType  string    `json:"entity_type" db:"entity_type"`
	EntityID    string    `json:"entity_id" db:"entity_id"`
	QuotaType   string    `json:"quota_type" db:"quota_type"`
	Usage       int64     `json:"usage" db:"usage"`
	Period      string    `json:"period" db:"period"`
	PeriodStart time.Time `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time `json:"period_end" db:"period_end"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// QuotaCheck represents the result of a quota check
type QuotaCheck struct {
	Allowed      bool      `json:"allowed"`
	QuotaType    string    `json:"quota_type"`
	Limit        int64     `json:"limit"`
	Usage        int64     `json:"usage"`
	Remaining    int64     `json:"remaining"`
	ResetAt      time.Time `json:"reset_at"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// QuotaRequest represents a request to consume quota
type QuotaRequest struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	QuotaType  string `json:"quota_type"`
	Amount     int64  `json:"amount"`
}

// NewQuotaService creates a new quota management service
func NewQuotaService(db *sql.DB) *QuotaService {
	return &QuotaService{
		db: db,
	}
}

// CheckQuota checks if a quota operation is allowed
func (q *QuotaService) CheckQuota(ctx context.Context, req *QuotaRequest) (*QuotaCheck, error) {
	// Get applicable quota for this entity and type
	quota, err := q.getApplicableQuota(ctx, req.EntityType, req.EntityID, req.QuotaType)
	if err != nil {
		return &QuotaCheck{
			Allowed:      true, // Allow if no quota is configured
			ErrorMessage: "No quota configured",
		}, nil
	}

	// Get current usage
	usage, err := q.getCurrentUsage(ctx, quota, req.EntityType, req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	// Check if the requested amount would exceed the quota
	newUsage := usage.Usage + req.Amount
	allowed := newUsage <= quota.Limit

	result := &QuotaCheck{
		Allowed:   allowed,
		QuotaType: quota.QuotaType,
		Limit:     quota.Limit,
		Usage:     usage.Usage,
		Remaining: quota.Limit - usage.Usage,
		ResetAt:   usage.PeriodEnd,
	}

	if !allowed {
		result.ErrorMessage = fmt.Sprintf("Quota exceeded. Limit: %d, Current usage: %d, Requested: %d",
			quota.Limit, usage.Usage, req.Amount)
	}

	return result, nil
}

// ConsumeQuota consumes quota for an operation
func (q *QuotaService) ConsumeQuota(ctx context.Context, req *QuotaRequest) (*QuotaCheck, error) {
	// First check if the operation is allowed
	check, err := q.CheckQuota(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check quota: %w", err)
	}

	if !check.Allowed {
		return check, fmt.Errorf("quota exceeded: %s", check.ErrorMessage)
	}

	// Get the quota and usage
	quota, err := q.getApplicableQuota(ctx, req.EntityType, req.EntityID, req.QuotaType)
	if err != nil {
		// If no quota is configured, allow the operation
		return &QuotaCheck{Allowed: true}, nil
	}

	usage, err := q.getCurrentUsage(ctx, quota, req.EntityType, req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	// Update usage
	newUsage := usage.Usage + req.Amount
	err = q.updateUsage(ctx, usage.ID, newUsage)
	if err != nil {
		return nil, fmt.Errorf("failed to update usage: %w", err)
	}

	// Return updated check result
	check.Usage = newUsage
	check.Remaining = quota.Limit - newUsage

	return check, nil
}

// GetQuotaUsage retrieves current quota usage for an entity
func (q *QuotaService) GetQuotaUsage(ctx context.Context, entityType, entityID string) ([]*QuotaUsage, error) {
	query := `
		SELECT qu.id, qu.quota_id, qu.entity_type, qu.entity_id, qu.quota_type,
		       qu.usage, qu.period, qu.period_start, qu.period_end, qu.updated_at
		FROM quota_usage qu
		JOIN quotas q ON qu.quota_id = q.id
		WHERE qu.entity_type = ? AND qu.entity_id = ? AND q.is_active = 1
		ORDER BY qu.quota_type
	`

	rows, err := q.db.QueryContext(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query quota usage: %w", err)
	}
	defer rows.Close()

	var usages []*QuotaUsage
	for rows.Next() {
		var usage QuotaUsage
		err := rows.Scan(
			&usage.ID, &usage.QuotaID, &usage.EntityType, &usage.EntityID,
			&usage.QuotaType, &usage.Usage, &usage.Period, &usage.PeriodStart,
			&usage.PeriodEnd, &usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota usage: %w", err)
		}
		usages = append(usages, &usage)
	}

	return usages, nil
}

// CreateQuota creates a new quota configuration
func (q *QuotaService) CreateQuota(ctx context.Context, quota *Quota) error {
	quota.ID = q.generateQuotaID()
	quota.CreatedAt = time.Now()
	quota.UpdatedAt = time.Now()
	quota.IsActive = true

	// Calculate reset time based on period
	quota.ResetAt = q.calculateNextReset(quota.Period, time.Now())

	query := `
		INSERT INTO quotas (
			id, entity_type, entity_id, quota_type, limit_value, period,
			reset_at, created_at, updated_at, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := q.db.ExecContext(ctx, query,
		quota.ID, quota.EntityType, quota.EntityID, quota.QuotaType,
		quota.Limit, quota.Period, quota.ResetAt, quota.CreatedAt,
		quota.UpdatedAt, quota.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create quota: %w", err)
	}

	return nil
}

// UpdateQuota updates an existing quota configuration
func (q *QuotaService) UpdateQuota(ctx context.Context, quotaID string, limit int64, period string) error {
	resetAt := q.calculateNextReset(period, time.Now())

	query := `
		UPDATE quotas 
		SET limit_value = ?, period = ?, reset_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := q.db.ExecContext(ctx, query, limit, period, resetAt, time.Now(), quotaID)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	return nil
}

// DeleteQuota deactivates a quota configuration
func (q *QuotaService) DeleteQuota(ctx context.Context, quotaID string) error {
	query := `UPDATE quotas SET is_active = 0, updated_at = ? WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, time.Now(), quotaID)
	return err
}

// ResetExpiredQuotas resets usage for all expired quota periods
func (q *QuotaService) ResetExpiredQuotas(ctx context.Context) error {
	now := time.Now()

	// Get all quotas that need to be reset
	query := `
		SELECT id, entity_type, entity_id, quota_type, period, reset_at
		FROM quotas
		WHERE is_active = 1 AND reset_at <= ?
	`

	rows, err := q.db.QueryContext(ctx, query, now)
	if err != nil {
		return fmt.Errorf("failed to query expired quotas: %w", err)
	}
	defer rows.Close()

	var expiredQuotas []Quota
	for rows.Next() {
		var quota Quota
		err := rows.Scan(
			&quota.ID, &quota.EntityType, &quota.EntityID,
			&quota.QuotaType, &quota.Period, &quota.ResetAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan expired quota: %w", err)
		}
		expiredQuotas = append(expiredQuotas, quota)
	}

	// Reset each expired quota
	for _, quota := range expiredQuotas {
		if err := q.resetQuotaUsage(ctx, &quota); err != nil {
			return fmt.Errorf("failed to reset quota %s: %w", quota.ID, err)
		}
	}

	return nil
}

// GetQuotaStats retrieves quota statistics for monitoring
func (q *QuotaService) GetQuotaStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total active quotas
	var totalQuotas int64
	err := q.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM quotas WHERE is_active = 1").Scan(&totalQuotas)
	if err != nil {
		return nil, fmt.Errorf("failed to get total quotas: %w", err)
	}
	stats["total_active_quotas"] = totalQuotas

	// Quotas by type
	quotasByType := make(map[string]int64)
	rows, err := q.db.QueryContext(ctx, "SELECT quota_type, COUNT(*) FROM quotas WHERE is_active = 1 GROUP BY quota_type")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var quotaType string
			var count int64
			if err := rows.Scan(&quotaType, &count); err == nil {
				quotasByType[quotaType] = count
			}
		}
	}
	stats["quotas_by_type"] = quotasByType

	// Usage exceeding 90% of quota
	var nearLimitCount int64
	nearLimitQuery := `
		SELECT COUNT(*)
		FROM quota_usage qu
		JOIN quotas q ON qu.quota_id = q.id
		WHERE q.is_active = 1 AND qu.usage > (q.limit_value * 0.9)
	`
	err = q.db.QueryRowContext(ctx, nearLimitQuery).Scan(&nearLimitCount)
	if err == nil {
		stats["near_limit_count"] = nearLimitCount
	}

	return stats, nil
}

// Helper methods

// getApplicableQuota finds the most specific quota that applies to an entity
func (q *QuotaService) getApplicableQuota(ctx context.Context, entityType, entityID, quotaType string) (*Quota, error) {
	// Try to find specific quota for this entity
	query := `
		SELECT id, entity_type, entity_id, quota_type, limit_value, period, reset_at,
		       created_at, updated_at, is_active
		FROM quotas
		WHERE entity_type = ? AND entity_id = ? AND quota_type = ? AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var quota Quota
	err := q.db.QueryRowContext(ctx, query, entityType, entityID, quotaType).Scan(
		&quota.ID, &quota.EntityType, &quota.EntityID, &quota.QuotaType,
		&quota.Limit, &quota.Period, &quota.ResetAt, &quota.CreatedAt,
		&quota.UpdatedAt, &quota.IsActive,
	)

	if err == nil {
		return &quota, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query specific quota: %w", err)
	}

	// If no specific quota found, try domain-level quota
	if entityType == "user" {
		// Extract domain from user email (assuming entityID is email)
		domain := extractDomain(entityID)
		if domain != "" {
			err = q.db.QueryRowContext(ctx, query, "domain", domain, quotaType).Scan(
				&quota.ID, &quota.EntityType, &quota.EntityID, &quota.QuotaType,
				&quota.Limit, &quota.Period, &quota.ResetAt, &quota.CreatedAt,
				&quota.UpdatedAt, &quota.IsActive,
			)
			if err == nil {
				return &quota, nil
			}
		}
	}

	// If no domain quota found, try global quota
	err = q.db.QueryRowContext(ctx, query, "global", "default", quotaType).Scan(
		&quota.ID, &quota.EntityType, &quota.EntityID, &quota.QuotaType,
		&quota.Limit, &quota.Period, &quota.ResetAt, &quota.CreatedAt,
		&quota.UpdatedAt, &quota.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no quota configured for %s %s %s", entityType, entityID, quotaType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query global quota: %w", err)
	}

	return &quota, nil
}

// getCurrentUsage gets or creates current usage record for a quota
func (q *QuotaService) getCurrentUsage(ctx context.Context, quota *Quota, entityType, entityID string) (*QuotaUsage, error) {
	now := time.Now()
	periodStart, periodEnd := q.calculatePeriodBounds(quota.Period, now)

	// Try to get existing usage record for current period
	query := `
		SELECT id, quota_id, entity_type, entity_id, quota_type, usage,
		       period, period_start, period_end, updated_at
		FROM quota_usage
		WHERE quota_id = ? AND entity_type = ? AND entity_id = ? 
		AND period_start = ? AND period_end = ?
	`

	var usage QuotaUsage
	err := q.db.QueryRowContext(ctx, query, quota.ID, entityType, entityID, periodStart, periodEnd).Scan(
		&usage.ID, &usage.QuotaID, &usage.EntityType, &usage.EntityID,
		&usage.QuotaType, &usage.Usage, &usage.Period, &usage.PeriodStart,
		&usage.PeriodEnd, &usage.UpdatedAt,
	)

	if err == nil {
		return &usage, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query usage: %w", err)
	}

	// Create new usage record for current period
	usage = QuotaUsage{
		ID:          q.generateUsageID(),
		QuotaID:     quota.ID,
		EntityType:  entityType,
		EntityID:    entityID,
		QuotaType:   quota.QuotaType,
		Usage:       0,
		Period:      quota.Period,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		UpdatedAt:   now,
	}

	insertQuery := `
		INSERT INTO quota_usage (
			id, quota_id, entity_type, entity_id, quota_type, usage,
			period, period_start, period_end, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = q.db.ExecContext(ctx, insertQuery,
		usage.ID, usage.QuotaID, usage.EntityType, usage.EntityID,
		usage.QuotaType, usage.Usage, usage.Period, usage.PeriodStart,
		usage.PeriodEnd, usage.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create usage record: %w", err)
	}

	return &usage, nil
}

// updateUsage updates the usage amount for a usage record
func (q *QuotaService) updateUsage(ctx context.Context, usageID string, newUsage int64) error {
	query := `UPDATE quota_usage SET usage = ?, updated_at = ? WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, newUsage, time.Now(), usageID)
	return err
}

// resetQuotaUsage resets usage for a quota and updates its reset time
func (q *QuotaService) resetQuotaUsage(ctx context.Context, quota *Quota) error {
	// Update quota reset time
	nextReset := q.calculateNextReset(quota.Period, time.Now())
	updateQuotaQuery := `UPDATE quotas SET reset_at = ?, updated_at = ? WHERE id = ?`
	_, err := q.db.ExecContext(ctx, updateQuotaQuery, nextReset, time.Now(), quota.ID)
	if err != nil {
		return fmt.Errorf("failed to update quota reset time: %w", err)
	}

	// Delete old usage records for this quota (they'll be recreated as needed)
	deleteUsageQuery := `DELETE FROM quota_usage WHERE quota_id = ?`
	_, err = q.db.ExecContext(ctx, deleteUsageQuery, quota.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old usage records: %w", err)
	}

	return nil
}

// calculatePeriodBounds calculates the start and end times for a quota period
func (q *QuotaService) calculatePeriodBounds(period string, now time.Time) (time.Time, time.Time) {
	switch period {
	case "minute":
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
		end := start.Add(time.Minute)
		return start, end
	case "hour":
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		end := start.Add(time.Hour)
		return start, end
	case "day":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 0, 1)
		return start, end
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0)
		return start, end
	default:
		// Default to daily
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 0, 1)
		return start, end
	}
}

// calculateNextReset calculates the next reset time for a quota period
func (q *QuotaService) calculateNextReset(period string, now time.Time) time.Time {
	_, end := q.calculatePeriodBounds(period, now)
	return end
}

// extractDomain extracts the domain part from an email address
func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func (q *QuotaService) generateQuotaID() string {
	return fmt.Sprintf("quota_%d", time.Now().UnixNano())
}

func (q *QuotaService) generateUsageID() string {
	return fmt.Sprintf("usage_%d", time.Now().UnixNano())
}

// GetDefaultQuotas returns default quota configurations for different entity types
func GetDefaultQuotas() []*Quota {
	now := time.Now()
	return []*Quota{
		{
			EntityType: "global",
			EntityID:   "default",
			QuotaType:  "email_send",
			Limit:      1000,
			Period:     "hour",
			ResetAt:    now.Add(time.Hour),
			IsActive:   true,
		},
		{
			EntityType: "global",
			EntityID:   "default",
			QuotaType:  "link_create",
			Limit:      5000,
			Period:     "hour",
			ResetAt:    now.Add(time.Hour),
			IsActive:   true,
		},
		{
			EntityType: "user",
			EntityID:   "default",
			QuotaType:  "email_send",
			Limit:      100,
			Period:     "hour",
			ResetAt:    now.Add(time.Hour),
			IsActive:   true,
		},
		{
			EntityType: "user",
			EntityID:   "default",
			QuotaType:  "link_create",
			Limit:      500,
			Period:     "hour",
			ResetAt:    now.Add(time.Hour),
			IsActive:   true,
		},
	}
}
