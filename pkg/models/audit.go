package models

import (
	"time"
)

// AuditLog represents an audit log entry for tracking system events
type AuditLog struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	Details   string    `json:"details"`
	Severity  string    `json:"severity"`
}

// AuditLogFilters represents filters for querying audit logs
type AuditLogFilters struct {
	UserID   string `json:"user_id,omitempty"`
	Action   string `json:"action,omitempty"`
	Entity   string `json:"entity,omitempty"`
	Severity string `json:"severity,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// AuditLogResponse represents the API response for audit logs
type AuditLogResponse struct {
	Success bool        `json:"success"`
	Logs    []AuditLog  `json:"logs"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	Filters AuditLogFilters `json:"filters,omitempty"`
}






