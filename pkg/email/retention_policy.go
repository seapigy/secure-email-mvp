package email

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// RetentionPolicy represents a configurable retention rule
type RetentionPolicy struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"` // Higher number = higher priority
	Active      bool   `json:"active"`

	// Rule conditions
	UserID          *string `json:"user_id,omitempty"`          // Specific user
	SenderDomain    *string `json:"sender_domain,omitempty"`    // Sender email domain
	RecipientDomain *string `json:"recipient_domain,omitempty"` // Recipient email domain
	EmailStatus     *string `json:"email_status,omitempty"`     // read, unread, expired
	CustomTags      *string `json:"custom_tags,omitempty"`      // JSON array of tags
	MinAgeHours     *int    `json:"min_age_hours,omitempty"`    // Minimum email age
	MaxAgeHours     *int    `json:"max_age_hours,omitempty"`    // Maximum email age

	// Actions
	RetentionDays        int  `json:"retention_days"`         // How long to keep emails
	ArchiveInstead       bool `json:"archive_instead"`        // Archive instead of delete
	ArchiveRetentionDays int  `json:"archive_retention_days"` // How long to keep archived

	// Metadata
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PolicyMatch represents a policy that matches an email
type PolicyMatch struct {
	Policy  *RetentionPolicy
	Score   int // Higher score = better match
	Reasons []string
}

// RetentionPolicyEngine manages retention policies and evaluation
type RetentionPolicyEngine struct {
	db *sql.DB
}

// NewRetentionPolicyEngine creates a new policy engine
func NewRetentionPolicyEngine(db *sql.DB) *RetentionPolicyEngine {
	return &RetentionPolicyEngine{
		db: db,
	}
}

// CreatePolicy creates a new retention policy
func (rpe *RetentionPolicyEngine) CreatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	query := `
		INSERT INTO retention_policies (
			name, description, priority, active, user_id, sender_domain, 
			recipient_domain, email_status, custom_tags, min_age_hours, 
			max_age_hours, retention_days, archive_instead, archive_retention_days, 
			created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	result, err := rpe.db.ExecContext(ctx, query,
		policy.Name, policy.Description, policy.Priority, policy.Active,
		policy.UserID, policy.SenderDomain, policy.RecipientDomain,
		policy.EmailStatus, policy.CustomTags, policy.MinAgeHours,
		policy.MaxAgeHours, policy.RetentionDays, policy.ArchiveInstead,
		policy.ArchiveRetentionDays, policy.CreatedBy, policy.CreatedAt, policy.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get policy ID: %w", err)
	}

	policy.ID = id
	return nil
}

// UpdatePolicy updates an existing retention policy
func (rpe *RetentionPolicyEngine) UpdatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	query := `
		UPDATE retention_policies SET
			name = ?, description = ?, priority = ?, active = ?, user_id = ?,
			sender_domain = ?, recipient_domain = ?, email_status = ?, custom_tags = ?,
			min_age_hours = ?, max_age_hours = ?, retention_days = ?, archive_instead = ?,
			archive_retention_days = ?, updated_at = ?
		WHERE id = ?
	`

	policy.UpdatedAt = time.Now()

	result, err := rpe.db.ExecContext(ctx, query,
		policy.Name, policy.Description, policy.Priority, policy.Active,
		policy.UserID, policy.SenderDomain, policy.RecipientDomain,
		policy.EmailStatus, policy.CustomTags, policy.MinAgeHours,
		policy.MaxAgeHours, policy.RetentionDays, policy.ArchiveInstead,
		policy.ArchiveRetentionDays, policy.UpdatedAt, policy.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("policy with ID %d not found", policy.ID)
	}

	return nil
}

// DeletePolicy removes a retention policy
func (rpe *RetentionPolicyEngine) DeletePolicy(ctx context.Context, policyID int64) error {
	query := `DELETE FROM retention_policies WHERE id = ?`

	result, err := rpe.db.ExecContext(ctx, query, policyID)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("policy with ID %d not found", policyID)
	}

	return nil
}

// GetPolicies retrieves retention policies with optional filtering
func (rpe *RetentionPolicyEngine) GetPolicies(ctx context.Context, filters map[string]string, limit, offset int) ([]*RetentionPolicy, error) {
	query := `SELECT id, name, description, priority, active, user_id, sender_domain, 
		recipient_domain, email_status, custom_tags, min_age_hours, max_age_hours, 
		retention_days, archive_instead, archive_retention_days, created_by, created_at, updated_at
		FROM retention_policies WHERE 1=1`

	var args []interface{}

	// Apply filters
	if active, ok := filters["active"]; ok {
		query += " AND active = ?"
		args = append(args, active == "true")
	}

	if userID, ok := filters["user_id"]; ok {
		query += " AND (user_id = ? OR user_id IS NULL)"
		args = append(args, userID)
	}

	if domain, ok := filters["domain"]; ok {
		query += " AND (sender_domain = ? OR recipient_domain = ?)"
		args = append(args, domain, domain)
	}

	query += " ORDER BY priority DESC, created_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := rpe.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	var policies []*RetentionPolicy
	for rows.Next() {
		policy := &RetentionPolicy{}
		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Active,
			&policy.UserID, &policy.SenderDomain, &policy.RecipientDomain,
			&policy.EmailStatus, &policy.CustomTags, &policy.MinAgeHours,
			&policy.MaxAgeHours, &policy.RetentionDays, &policy.ArchiveInstead,
			&policy.ArchiveRetentionDays, &policy.CreatedBy, &policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// GetPolicyByID retrieves a specific policy by ID
func (rpe *RetentionPolicyEngine) GetPolicyByID(ctx context.Context, policyID int64) (*RetentionPolicy, error) {
	query := `SELECT id, name, description, priority, active, user_id, sender_domain, 
		recipient_domain, email_status, custom_tags, min_age_hours, max_age_hours, 
		retention_days, archive_instead, archive_retention_days, created_by, created_at, updated_at
		FROM retention_policies WHERE id = ?`

	policy := &RetentionPolicy{}
	err := rpe.db.QueryRowContext(ctx, query, policyID).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Active,
		&policy.UserID, &policy.SenderDomain, &policy.RecipientDomain,
		&policy.EmailStatus, &policy.CustomTags, &policy.MinAgeHours,
		&policy.MaxAgeHours, &policy.RetentionDays, &policy.ArchiveInstead,
		&policy.ArchiveRetentionDays, &policy.CreatedBy, &policy.CreatedAt, &policy.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("policy with ID %d not found", policyID)
		}
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	return policy, nil
}

// EvaluatePoliciesForEmail determines which policies match a given email
func (rpe *RetentionPolicyEngine) EvaluatePoliciesForEmail(ctx context.Context, email *EmailRetentionInfo) ([]*PolicyMatch, error) {
	// Get all active policies
	policies, err := rpe.GetPolicies(ctx, map[string]string{"active": "true"}, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	var matches []*PolicyMatch

	for _, policy := range policies {
		match := rpe.evaluatePolicyMatch(policy, email)
		if match != nil {
			matches = append(matches, match)
		}
	}

	return matches, nil
}

// GetBestMatchingPolicy returns the policy with the highest priority that matches the email
func (rpe *RetentionPolicyEngine) GetBestMatchingPolicy(ctx context.Context, email *EmailRetentionInfo) (*RetentionPolicy, error) {
	matches, err := rpe.EvaluatePoliciesForEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil // No matching policies
	}

	// Find the policy with the highest priority
	bestMatch := matches[0]
	for _, match := range matches[1:] {
		if match.Policy.Priority > bestMatch.Policy.Priority {
			bestMatch = match
		}
	}

	return bestMatch.Policy, nil
}

// evaluatePolicyMatch determines if a policy matches an email and calculates a match score
func (rpe *RetentionPolicyEngine) evaluatePolicyMatch(policy *RetentionPolicy, email *EmailRetentionInfo) *PolicyMatch {
	var reasons []string
	score := 0

	// Check user ID match
	if policy.UserID != nil && *policy.UserID == email.SenderID {
		score += 100
		reasons = append(reasons, "user_id_match")
	}

	// Check sender domain match
	if policy.SenderDomain != nil {
		senderDomain := extractDomain(email.SenderID)
		if senderDomain == *policy.SenderDomain {
			score += 50
			reasons = append(reasons, "sender_domain_match")
		}
	}

	// Check recipient domain match
	if policy.RecipientDomain != nil {
		recipientDomain := extractDomain(email.Recipient)
		if recipientDomain == *policy.RecipientDomain {
			score += 50
			reasons = append(reasons, "recipient_domain_match")
		}
	}

	// Check email status match
	if policy.EmailStatus != nil {
		emailStatus := determineEmailStatus(email)
		if emailStatus == *policy.EmailStatus {
			score += 30
			reasons = append(reasons, "status_match")
		}
	}

	// Check age constraints
	emailAge := time.Since(email.CreatedAt)
	if policy.MinAgeHours != nil {
		minAge := time.Duration(*policy.MinAgeHours) * time.Hour
		if emailAge < minAge {
			return nil // Email too young
		}
		score += 20
		reasons = append(reasons, "min_age_met")
	}

	if policy.MaxAgeHours != nil {
		maxAge := time.Duration(*policy.MaxAgeHours) * time.Hour
		if emailAge > maxAge {
			return nil // Email too old
		}
		score += 20
		reasons = append(reasons, "max_age_met")
	}

	// Check custom tags (if implemented)
	if policy.CustomTags != nil {
		// This would require email tags to be implemented
		// For now, we'll skip this check
	}

	// If we have any matches, return the policy match
	if score > 0 {
		return &PolicyMatch{
			Policy:  policy,
			Score:   score,
			Reasons: reasons,
		}
	}

	return nil
}

// extractDomain extracts the domain from an email address
func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// determineEmailStatus determines the current status of an email
func determineEmailStatus(email *EmailRetentionInfo) string {
	if email.Status == "expired" {
		return "expired"
	}
	if email.Status == "self_destructed" {
		return "expired"
	}
	if email.AccessCount > 0 {
		return "read"
	}
	return "unread"
}

// GetDefaultRetentionPolicy returns the default policy when no specific policy matches
func (rpe *RetentionPolicyEngine) GetDefaultRetentionPolicy() *RetentionPolicy {
	defaultRetentionDays := getDefaultRetentionDays()
	defaultArchiveRetentionDays := getDefaultArchiveRetentionDays()

	return &RetentionPolicy{
		Name:                 "Default Policy",
		Description:          "Default retention policy when no specific policy matches",
		Priority:             0,
		Active:               true,
		RetentionDays:        defaultRetentionDays,
		ArchiveInstead:       getDefaultArchiveInstead(),
		ArchiveRetentionDays: defaultArchiveRetentionDays,
		CreatedBy:            "system",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

// Helper functions to get configuration from environment variables
func getDefaultRetentionDays() int {
	if days := os.Getenv("DEFAULT_RETENTION_DAYS"); days != "" {
		if d, err := strconv.Atoi(days); err == nil && d > 0 {
			return d
		}
	}
	return 30 // Default to 30 days
}

func getDefaultArchiveRetentionDays() int {
	if days := os.Getenv("DEFAULT_ARCHIVE_RETENTION_DAYS"); days != "" {
		if d, err := strconv.Atoi(days); err == nil && d > 0 {
			return d
		}
	}
	return 365 // Default to 1 year
}

func getDefaultArchiveInstead() bool {
	if archive := os.Getenv("DEFAULT_ARCHIVE_INSTEAD"); archive != "" {
		return archive == "true"
	}
	return false // Default to deletion
}
