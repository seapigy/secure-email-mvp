// =============================================================================
// SECURE EMAIL MVP - AUDIT LOG SYSTEM
// =============================================================================
// Package audit provides comprehensive audit logging for the Secure Email MVP.
// Tracks all system events with detailed metadata and supports filtering/export.
// =============================================================================

package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of audit event
type EventType string

const (
	EventTypeEmailCreation           EventType = "email_creation"
	EventTypeEmailAccess             EventType = "email_access"
	EventTypeEmailDeletion           EventType = "email_deletion"
	EventTypeLoginAttempt            EventType = "login_attempt"
	EventTypeAPIKeyUse               EventType = "api_key_use"
	EventTypeReadReceipt             EventType = "read_receipt"
	EventTypeExpirationAlert         EventType = "expiration_alert"
	EventTypeSystemEvent             EventType = "system_event"
	EventTypeUserRegistration        EventType = "user_registration"
	EventTypePasswordChange          EventType = "password_change"
	EventTypeMFASetup                EventType = "mfa_setup"
	EventTypeGeolocationVerification EventType = "geolocation_verification"
)

// Outcome represents the outcome of an event
type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
	OutcomeBlocked Outcome = "blocked"
)

// Severity represents the severity level of an event
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	LogID          string                 `json:"log_id"`
	Timestamp      time.Time              `json:"timestamp"`
	EventType      EventType              `json:"event_type"`
	UserID         *string                `json:"user_id,omitempty"`
	IPAddress      *string                `json:"ip_address,omitempty"`
	UserAgent      *string                `json:"user_agent,omitempty"`
	RelatedEmailID *string                `json:"related_email_id,omitempty"`
	Outcome        Outcome                `json:"outcome"`
	Details        map[string]interface{} `json:"details,omitempty"`
	Severity       Severity               `json:"severity"`
	SessionID      *string                `json:"session_id,omitempty"`
	RequestID      *string                `json:"request_id,omitempty"`
	Country        *string                `json:"country,omitempty"`
	City           *string                `json:"city,omitempty"`
	DeviceType     *string                `json:"device_type,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// AuditLogFilter represents filter criteria for querying audit logs
type AuditLogFilter struct {
	DateFrom        *time.Time  `json:"date_from,omitempty"`
	DateTo          *time.Time  `json:"date_to,omitempty"`
	EventTypes      []EventType `json:"event_types,omitempty"`
	UserIDs         []string    `json:"user_ids,omitempty"`
	Outcomes        []Outcome   `json:"outcomes,omitempty"`
	Severities      []Severity  `json:"severities,omitempty"`
	IPAddresses     []string    `json:"ip_addresses,omitempty"`
	RelatedEmailIDs []string    `json:"related_email_ids,omitempty"`
	SessionIDs      []string    `json:"session_ids,omitempty"`
	Countries       []string    `json:"countries,omitempty"`
	DeviceTypes     []string    `json:"device_types,omitempty"`
	SearchTerm      *string     `json:"search_term,omitempty"`
}

// AuditLogQuery represents a paginated query result
type AuditLogQuery struct {
	Events   []AuditEvent `json:"events"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	HasMore  bool         `json:"has_more"`
}

// ExportRequest represents an audit log export request
type ExportRequest struct {
	ExportID     string         `json:"export_id"`
	UserID       string         `json:"user_id"`
	ExportType   string         `json:"export_type"` // "csv" or "json"
	DateFrom     *time.Time     `json:"date_from,omitempty"`
	DateTo       *time.Time     `json:"date_to,omitempty"`
	EventTypes   []EventType    `json:"event_types,omitempty"`
	Filters      AuditLogFilter `json:"filters,omitempty"`
	Status       string         `json:"status"`
	FilePath     *string        `json:"file_path,omitempty"`
	FileSize     *int64         `json:"file_size,omitempty"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
}

// SavedFilter represents a saved user filter
type SavedFilter struct {
	FilterID     string         `json:"filter_id"`
	UserID       string         `json:"user_id"`
	FilterName   string         `json:"filter_name"`
	FilterConfig AuditLogFilter `json:"filter_config"`
	IsDefault    bool           `json:"is_default"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// RetentionPolicy represents a retention policy for audit logs
type RetentionPolicy struct {
	RetentionID   string    `json:"retention_id"`
	EventType     string    `json:"event_type"`
	RetentionDays int       `json:"retention_days"`
	AutoPurge     bool      `json:"auto_purge"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AuditService provides audit logging functionality
type AuditService struct {
	db *sql.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

// RecordEvent records a new audit event
func (s *AuditService) RecordEvent(ctx context.Context, event *AuditEvent) error {
	if event.LogID == "" {
		event.LogID = uuid.New().String()
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	detailsJSON := "{}"
	if event.Details != nil {
		if details, err := json.Marshal(event.Details); err == nil {
			detailsJSON = string(details)
		}
	}

	query := `
		INSERT INTO audit_log (
			log_id, timestamp, event_type, user_id, ip_address, user_agent,
			related_email_id, outcome, details, severity, session_id, request_id,
			country, city, device_type, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		event.LogID, event.Timestamp, event.EventType, event.UserID, event.IPAddress,
		event.UserAgent, event.RelatedEmailID, event.Outcome, detailsJSON,
		event.Severity, event.SessionID, event.RequestID, event.Country,
		event.City, event.DeviceType, event.CreatedAt,
	)

	if err != nil {
		log.Printf("Failed to record audit event: %v", err)
		return fmt.Errorf("failed to record audit event: %w", err)
	}

	return nil
}

// QueryEvents queries audit events with filtering and pagination
func (s *AuditService) QueryEvents(ctx context.Context, filter AuditLogFilter, page, pageSize int) (*AuditLogQuery, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	if filter.DateFrom != nil {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, *filter.DateFrom)
	}

	if filter.DateTo != nil {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, *filter.DateTo)
	}

	if len(filter.EventTypes) > 0 {
		placeholders := make([]string, len(filter.EventTypes))
		for i := range filter.EventTypes {
			placeholders[i] = "?"
			args = append(args, filter.EventTypes[i])
		}
		whereClauses = append(whereClauses, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.UserIDs) > 0 {
		placeholders := make([]string, len(filter.UserIDs))
		for i := range filter.UserIDs {
			placeholders[i] = "?"
			args = append(args, filter.UserIDs[i])
		}
		whereClauses = append(whereClauses, fmt.Sprintf("user_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Outcomes) > 0 {
		placeholders := make([]string, len(filter.Outcomes))
		for i := range filter.Outcomes {
			placeholders[i] = "?"
			args = append(args, filter.Outcomes[i])
		}
		whereClauses = append(whereClauses, fmt.Sprintf("outcome IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Severities) > 0 {
		placeholders := make([]string, len(filter.Severities))
		for i := range filter.Severities {
			placeholders[i] = "?"
			args = append(args, filter.Severities[i])
		}
		whereClauses = append(whereClauses, fmt.Sprintf("severity IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.IPAddresses) > 0 {
		placeholders := make([]string, len(filter.IPAddresses))
		for i := range filter.IPAddresses {
			placeholders[i] = "?"
			args = append(args, filter.IPAddresses[i])
		}
		whereClauses = append(whereClauses, fmt.Sprintf("ip_address IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.RelatedEmailIDs) > 0 {
		placeholders := make([]string, len(filter.RelatedEmailIDs))
		for i := range filter.RelatedEmailIDs {
			placeholders[i] = "?"
			args = append(args, filter.RelatedEmailIDs[i])
		}
		whereClauses = append(whereClauses, fmt.Sprintf("related_email_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		whereClauses = append(whereClauses, "(details LIKE ? OR user_agent LIKE ?)")
		searchTerm := "%" + *filter.SearchTerm + "%"
		args = append(args, searchTerm, searchTerm)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_log %s", whereClause)
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit events: %w", err)
	}

	// Query events
	query := fmt.Sprintf(`
		SELECT log_id, timestamp, event_type, user_id, ip_address, user_agent,
		       related_email_id, outcome, details, severity, session_id, request_id,
		       country, city, device_type, created_at
		FROM audit_log %s
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pageSize, offset)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var detailsJSON string
		var userID, ipAddress, userAgent, relatedEmailID, sessionID, requestID, country, city, deviceType sql.NullString

		err := rows.Scan(
			&event.LogID, &event.Timestamp, &event.EventType, &userID, &ipAddress, &userAgent,
			&relatedEmailID, &event.Outcome, &detailsJSON, &event.Severity, &sessionID, &requestID,
			&country, &city, &deviceType, &event.CreatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan audit event: %v", err)
			continue
		}

		if userID.Valid {
			event.UserID = &userID.String
		}
		if ipAddress.Valid {
			event.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			event.UserAgent = &userAgent.String
		}
		if relatedEmailID.Valid {
			event.RelatedEmailID = &relatedEmailID.String
		}
		if sessionID.Valid {
			event.SessionID = &sessionID.String
		}
		if requestID.Valid {
			event.RequestID = &requestID.String
		}
		if country.Valid {
			event.Country = &country.String
		}
		if city.Valid {
			event.City = &city.String
		}
		if deviceType.Valid {
			event.DeviceType = &deviceType.String
		}

		if detailsJSON != "" && detailsJSON != "{}" {
			if err := json.Unmarshal([]byte(detailsJSON), &event.Details); err != nil {
				log.Printf("Failed to unmarshal event details: %v", err)
			}
		}

		events = append(events, event)
	}

	hasMore := (page * pageSize) < total

	return &AuditLogQuery{
		Events:   events,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}

// GetEventTypes returns all available event types
func (s *AuditService) GetEventTypes(ctx context.Context) ([]string, error) {
	query := "SELECT DISTINCT event_type FROM audit_log ORDER BY event_type"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get event types: %w", err)
	}
	defer rows.Close()

	var eventTypes []string
	for rows.Next() {
		var eventType string
		if err := rows.Scan(&eventType); err != nil {
			log.Printf("Failed to scan event type: %v", err)
			continue
		}
		eventTypes = append(eventTypes, eventType)
	}

	return eventTypes, nil
}

// GetRetentionPolicies returns all retention policies
func (s *AuditService) GetRetentionPolicies(ctx context.Context) ([]RetentionPolicy, error) {
	query := "SELECT retention_id, event_type, retention_days, auto_purge, created_at, updated_at FROM audit_log_retention ORDER BY event_type"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get retention policies: %w", err)
	}
	defer rows.Close()

	var policies []RetentionPolicy
	for rows.Next() {
		var policy RetentionPolicy
		err := rows.Scan(
			&policy.RetentionID, &policy.EventType, &policy.RetentionDays,
			&policy.AutoPurge, &policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan retention policy: %v", err)
			continue
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// UpdateRetentionPolicy updates a retention policy
func (s *AuditService) UpdateRetentionPolicy(ctx context.Context, policy RetentionPolicy) error {
	query := `
		UPDATE audit_log_retention 
		SET retention_days = ?, auto_purge = ?, updated_at = CURRENT_TIMESTAMP
		WHERE retention_id = ?
	`

	_, err := s.db.ExecContext(ctx, query, policy.RetentionDays, policy.AutoPurge, policy.RetentionID)
	if err != nil {
		return fmt.Errorf("failed to update retention policy: %w", err)
	}

	return nil
}

// PurgeExpiredLogs purges logs based on retention policies
func (s *AuditService) PurgeExpiredLogs(ctx context.Context) error {
	policies, err := s.GetRetentionPolicies(ctx)
	if err != nil {
		return fmt.Errorf("failed to get retention policies: %w", err)
	}

	for _, policy := range policies {
		if !policy.AutoPurge {
			continue
		}

		cutoffDate := time.Now().UTC().AddDate(0, 0, -policy.RetentionDays)
		query := "DELETE FROM audit_log WHERE event_type = ? AND created_at < ?"

		result, err := s.db.ExecContext(ctx, query, policy.EventType, cutoffDate)
		if err != nil {
			log.Printf("Failed to purge expired logs for event type %s: %v", policy.EventType, err)
			continue
		}

		deleted, _ := result.RowsAffected()
		if deleted > 0 {
			log.Printf("Purged %d expired audit logs for event type %s", deleted, policy.EventType)
		}
	}

	return nil
}

// GetUserEvents returns audit events for a specific user
func (s *AuditService) GetUserEvents(ctx context.Context, userID string, limit int) ([]AuditEvent, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
		SELECT log_id, timestamp, event_type, user_id, ip_address, user_agent,
		       related_email_id, outcome, details, severity, session_id, request_id,
		       country, city, device_type, created_at
		FROM audit_log
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var detailsJSON string
		var ipAddress, userAgent, relatedEmailID, sessionID, requestID, country, city, deviceType sql.NullString

		err := rows.Scan(
			&event.LogID, &event.Timestamp, &event.EventType, &event.UserID, &ipAddress, &userAgent,
			&relatedEmailID, &event.Outcome, &detailsJSON, &event.Severity, &sessionID, &requestID,
			&country, &city, &deviceType, &event.CreatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan user event: %v", err)
			continue
		}

		if ipAddress.Valid {
			event.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			event.UserAgent = &userAgent.String
		}
		if relatedEmailID.Valid {
			event.RelatedEmailID = &relatedEmailID.String
		}
		if sessionID.Valid {
			event.SessionID = &sessionID.String
		}
		if requestID.Valid {
			event.RequestID = &requestID.String
		}
		if country.Valid {
			event.Country = &country.String
		}
		if city.Valid {
			event.City = &city.String
		}
		if deviceType.Valid {
			event.DeviceType = &deviceType.String
		}

		if detailsJSON != "" && detailsJSON != "{}" {
			if err := json.Unmarshal([]byte(detailsJSON), &event.Details); err != nil {
				log.Printf("Failed to unmarshal event details: %v", err)
			}
		}

		events = append(events, event)
	}

	return events, nil
}











