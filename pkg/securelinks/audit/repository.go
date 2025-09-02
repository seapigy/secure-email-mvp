package audit

import (
	"database/sql"
	"fmt"
	"strings"
	"secure-email-mvp/pkg/models"
)

// Repository defines the interface for audit log operations
type Repository interface {
	CreateAuditLog(log *models.AuditLog) error
	GetAuditLogs(filters models.AuditLogFilters) ([]models.AuditLog, int, error)
}

// SQLiteAuditRepository implements the Repository interface using SQLite
type SQLiteAuditRepository struct {
	db *sql.DB
}

// NewSQLiteAuditRepository creates a new SQLite audit repository
func NewSQLiteAuditRepository(db *sql.DB) *SQLiteAuditRepository {
	return &SQLiteAuditRepository{db: db}
}

// CreateAuditLog inserts a new audit log entry
func (r *SQLiteAuditRepository) CreateAuditLog(log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (timestamp, user_id, action, entity, details, severity)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, 
		log.Timestamp, 
		log.UserID, 
		log.Action, 
		log.Entity, 
		log.Details, 
		log.Severity,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	log.ID = int(id)
	return nil
}

// GetAuditLogs retrieves audit logs with optional filtering and pagination
func (r *SQLiteAuditRepository) GetAuditLogs(filters models.AuditLogFilters) ([]models.AuditLog, int, error) {
	// Build the base query
	baseQuery := "SELECT id, timestamp, user_id, action, entity, details, severity FROM audit_logs"
	countQuery := "SELECT COUNT(*) FROM audit_logs"
	
	// Build WHERE clause for filters
	var conditions []string
	var args []interface{}
	
	if filters.UserID != "" {
		conditions = append(conditions, "user_id LIKE ?")
		args = append(args, "%"+filters.UserID+"%")
	}
	
	if filters.Action != "" {
		conditions = append(conditions, "action LIKE ?")
		args = append(args, "%"+filters.Action+"%")
	}
	
	if filters.Entity != "" {
		conditions = append(conditions, "entity LIKE ?")
		args = append(args, "%"+filters.Entity+"%")
	}
	
	if filters.Severity != "" {
		conditions = append(conditions, "severity = ?")
		args = append(args, filters.Severity)
	}
	
	// Add WHERE clause if there are conditions
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}
	
	// Add ORDER BY
	baseQuery += " ORDER BY timestamp DESC"
	
	// Add LIMIT and OFFSET for pagination
	if filters.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, filters.Limit)
		
		if filters.Offset > 0 {
			baseQuery += " OFFSET ?"
			args = append(args, filters.Offset)
		}
	}
	
	// Get total count
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}
	
	// Get audit logs
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()
	
	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.UserID,
			&log.Action,
			&log.Entity,
			&log.Details,
			&log.Severity,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit logs: %w", err)
	}
	
	return logs, total, nil
}







