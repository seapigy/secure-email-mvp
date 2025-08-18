package models

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ComplianceLog represents a compliance event log entry
type ComplianceLog struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organization_id"`
	Timestamp      time.Time       `json:"timestamp"`
	Action         string          `json:"action"`
	Details        json.RawMessage `json:"details"`
	CreatedAt      time.Time       `json:"created_at"`
}

// ComplianceSummary represents aggregated compliance metrics for an organization
type ComplianceSummary struct {
	OrganizationID        string     `json:"organization_id"`
	OrganizationName      string     `json:"organization_name"`
	TotalUsers            int        `json:"total_users"`
	PolicyViolations      int        `json:"policy_violations"`
	DataRetentionEvents   int        `json:"data_retention_events"`
	ExportRequests        int        `json:"export_requests"`
	AccessDenials         int        `json:"access_denials"`
	DataDeletions         int        `json:"data_deletions"`
	Last30DaysActivity    int        `json:"last_30d_activity"`
	LastActivityTimestamp *time.Time `json:"last_activity_timestamp,omitempty"`
}

// ComplianceLogFilter represents filters for querying compliance logs
type ComplianceLogFilter struct {
	Action    string     `json:"action,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// ComplianceActivity represents recent compliance activity
type ComplianceActivity struct {
	OrganizationID string    `json:"organization_id"`
	Action         string    `json:"action"`
	Count          int       `json:"count"`
	ActivityDate   time.Time `json:"activity_date"`
}

// LogComplianceEvent logs a compliance event for an organization
func LogComplianceEvent(db *sql.DB, organizationID, action string, details map[string]interface{}) error {
	id := uuid.New().String()

	// Convert details to JSON
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %v", err)
	}

	query := `INSERT INTO organization_compliance_logs (id, organization_id, action, details) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(query, id, organizationID, action, string(detailsJSON))
	if err != nil {
		return fmt.Errorf("failed to log compliance event: %v", err)
	}

	log.Printf("[COMPLIANCE] Logged %s event for organization %s", action, organizationID)
	return nil
}

// GetComplianceSummary retrieves compliance summary for an organization
func GetComplianceSummary(db *sql.DB, organizationID string) (*ComplianceSummary, error) {
	query := `
		SELECT 
			organization_id,
			organization_name,
			total_users,
			policy_violations,
			data_retention_events,
			export_requests,
			access_denials,
			data_deletions,
			last_30d_activity,
			last_activity_timestamp
		FROM organization_compliance_summary
		WHERE organization_id = ?
	`

	var summary ComplianceSummary
	var lastActivityStr sql.NullString

	err := db.QueryRow(query, organizationID).Scan(
		&summary.OrganizationID,
		&summary.OrganizationName,
		&summary.TotalUsers,
		&summary.PolicyViolations,
		&summary.DataRetentionEvents,
		&summary.ExportRequests,
		&summary.AccessDenials,
		&summary.DataDeletions,
		&summary.Last30DaysActivity,
		&lastActivityStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found: %s", organizationID)
		}
		return nil, fmt.Errorf("failed to get compliance summary: %v", err)
	}

	// Parse last activity timestamp if available
	if lastActivityStr.Valid {
		lastActivity, err := time.Parse("2006-01-02 15:04:05", lastActivityStr.String)
		if err == nil {
			summary.LastActivityTimestamp = &lastActivity
		}
	}

	return &summary, nil
}

// GetComplianceLogs retrieves compliance logs for an organization with optional filters
func GetComplianceLogs(db *sql.DB, organizationID string, filter *ComplianceLogFilter) ([]*ComplianceLog, error) {
	// Build query with filters
	query := `SELECT id, organization_id, timestamp, action, details, created_at FROM organization_compliance_logs WHERE organization_id = ?`
	args := []interface{}{organizationID}

	// Add action filter
	if filter != nil && filter.Action != "" {
		query += ` AND action = ?`
		args = append(args, filter.Action)
	}

	// Add date range filters
	if filter != nil && filter.StartDate != nil {
		query += ` AND timestamp >= ?`
		args = append(args, filter.StartDate.Format("2006-01-02 15:04:05"))
	}

	if filter != nil && filter.EndDate != nil {
		query += ` AND timestamp <= ?`
		args = append(args, filter.EndDate.Format("2006-01-02 15:04:05"))
	}

	// Add ordering and pagination
	query += ` ORDER BY timestamp DESC`

	if filter != nil && filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	} else {
		query += ` LIMIT 100` // Default limit
	}

	if filter != nil && filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance logs: %v", err)
	}
	defer rows.Close()

	var logs []*ComplianceLog
	for rows.Next() {
		var log ComplianceLog
		var detailsStr sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.OrganizationID,
			&log.Timestamp,
			&log.Action,
			&detailsStr,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance log: %v", err)
		}

		// Parse details JSON if available
		if detailsStr.Valid {
			log.Details = json.RawMessage(detailsStr.String)
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

// GetComplianceActivity retrieves recent compliance activity for an organization
func GetComplianceActivity(db *sql.DB, organizationID string, days int) ([]*ComplianceActivity, error) {
	if days <= 0 {
		days = 30 // Default to 30 days
	}

	query := `
		SELECT 
			organization_id,
			action,
			COUNT(*) as count,
			DATE(timestamp) as activity_date
		FROM organization_compliance_logs
		WHERE organization_id = ? AND timestamp >= datetime('now', '-? days')
		GROUP BY organization_id, action, DATE(timestamp)
		ORDER BY activity_date DESC, count DESC
	`

	rows, err := db.Query(query, organizationID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance activity: %v", err)
	}
	defer rows.Close()

	var activities []*ComplianceActivity
	for rows.Next() {
		var activity ComplianceActivity
		var dateStr string

		err := rows.Scan(
			&activity.OrganizationID,
			&activity.Action,
			&activity.Count,
			&dateStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance activity: %v", err)
		}

		// Parse date
		activity.ActivityDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse activity date: %v", err)
		}

		activities = append(activities, &activity)
	}

	return activities, nil
}

// ExportComplianceLogsCSV exports compliance logs as CSV data
func ExportComplianceLogsCSV(db *sql.DB, organizationID string, filter *ComplianceLogFilter) ([]byte, error) {
	// Get compliance logs
	logs, err := GetComplianceLogs(db, organizationID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance logs for export: %v", err)
	}

	// Create CSV buffer
	var csvBuffer strings.Builder
	writer := csv.NewWriter(&csvBuffer)

	// Write header
	header := []string{
		"ID",
		"Organization ID",
		"Timestamp",
		"Action",
		"Details",
		"Created At",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data rows
	for _, log := range logs {
		row := []string{
			log.ID,
			log.OrganizationID,
			log.Timestamp.Format("2006-01-02 15:04:05"),
			log.Action,
			string(log.Details),
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV: %v", err)
	}

	return []byte(csvBuffer.String()), nil
}

// GetComplianceStats retrieves high-level compliance statistics
func GetComplianceStats(db *sql.DB, organizationID string) (map[string]interface{}, error) {
	// Get summary
	summary, err := GetComplianceSummary(db, organizationID)
	if err != nil {
		return nil, err
	}

	// Get recent activity
	activity, err := GetComplianceActivity(db, organizationID, 30)
	if err != nil {
		return nil, err
	}

	// Calculate additional metrics
	totalEvents := summary.PolicyViolations + summary.DataRetentionEvents +
		summary.ExportRequests + summary.AccessDenials + summary.DataDeletions

	stats := map[string]interface{}{
		"organization_id":         summary.OrganizationID,
		"organization_name":       summary.OrganizationName,
		"total_users":             summary.TotalUsers,
		"total_compliance_events": totalEvents,
		"policy_violations":       summary.PolicyViolations,
		"data_retention_events":   summary.DataRetentionEvents,
		"export_requests":         summary.ExportRequests,
		"access_denials":          summary.AccessDenials,
		"data_deletions":          summary.DataDeletions,
		"last_30d_activity":       summary.Last30DaysActivity,
		"last_activity":           summary.LastActivityTimestamp,
		"recent_activity":         activity,
	}

	// Add percentage calculations
	if totalEvents > 0 {
		stats["policy_violation_rate"] = float64(summary.PolicyViolations) / float64(totalEvents) * 100
		stats["data_retention_rate"] = float64(summary.DataRetentionEvents) / float64(totalEvents) * 100
		stats["export_request_rate"] = float64(summary.ExportRequests) / float64(totalEvents) * 100
		stats["access_denial_rate"] = float64(summary.AccessDenials) / float64(totalEvents) * 100
		stats["data_deletion_rate"] = float64(summary.DataDeletions) / float64(totalEvents) * 100
	} else {
		stats["policy_violation_rate"] = 0.0
		stats["data_retention_rate"] = 0.0
		stats["export_request_rate"] = 0.0
		stats["access_denial_rate"] = 0.0
		stats["data_deletion_rate"] = 0.0
	}

	return stats, nil
}

// ValidateComplianceAction validates if a compliance action is valid
func ValidateComplianceAction(action string) bool {
	validActions := []string{
		"policy_violation",
		"user_data_retained",
		"export_requested",
		"access_denied",
		"data_deleted",
		"user_login",
		"user_logout",
		"data_export",
		"data_import",
		"configuration_change",
	}

	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// GetComplianceActionTypes returns all valid compliance action types
func GetComplianceActionTypes() []string {
	return []string{
		"policy_violation",
		"user_data_retained",
		"export_requested",
		"access_denied",
		"data_deleted",
		"user_login",
		"user_logout",
		"data_export",
		"data_import",
		"configuration_change",
	}
}
