package dlp

import (
	"database/sql"
	"secure-email-mvp/pkg/models"
)

// SQLiteRepository implements the Database interface for DLP operations
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository for DLP operations
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// GetActiveDLPRules retrieves all active DLP rules from the database
func (r *SQLiteRepository) GetActiveDLPRules() ([]models.DLPRule, error) {
	query := `
		SELECT rule_id, rule_name, rule_type, pattern, description, 
		       severity, action, is_active, created_at, updated_at, 
		       created_by, priority
		FROM dlp_rules 
		WHERE is_active = TRUE 
		ORDER BY priority DESC, created_at ASC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.DLPRule
	for rows.Next() {
		var rule models.DLPRule
		err := rows.Scan(
			&rule.RuleID,
			&rule.RuleName,
			&rule.RuleType,
			&rule.Pattern,
			&rule.Description,
			&rule.Severity,
			&rule.Action,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
			&rule.CreatedBy,
			&rule.Priority,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

// CreateDLPScanResult stores a DLP scan result in the database
func (r *SQLiteRepository) CreateDLPScanResult(result *models.DLPScanResult) error {
	query := `
		INSERT INTO dlp_scan_results (
			scan_id, link_id, reply_id, attachment_id, rule_id, 
			content_type, matched_content, confidence_score, 
			action_taken, scan_timestamp, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query,
		result.ScanID,
		result.LinkID,
		result.ReplyID,
		result.AttachmentID,
		result.RuleID,
		result.ContentType,
		result.MatchedContent,
		result.ConfidenceScore,
		result.ActionTaken,
		result.ScanTimestamp,
		result.CreatedBy,
	)
	
	return err
}

// CreateComplianceAuditLog stores a compliance audit log entry
func (r *SQLiteRepository) CreateComplianceAuditLog(audit *models.ComplianceAuditLog) error {
	query := `
		INSERT INTO compliance_audit_log (
			audit_id, event_type, link_id, reply_id, attachment_id,
			policy_id, rule_id, user_id, ip_address, user_agent,
			event_details, severity, compliance_category, 
			retention_required, created_at, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query,
		audit.AuditID,
		audit.EventType,
		audit.LinkID,
		audit.ReplyID,
		audit.AttachmentID,
		audit.PolicyID,
		audit.RuleID,
		audit.UserID,
		audit.IPAddress,
		audit.UserAgent,
		audit.EventDetails,
		audit.Severity,
		audit.ComplianceCategory,
		audit.RetentionRequired,
		audit.CreatedAt,
		audit.CreatedBy,
	)
	
	return err
}







