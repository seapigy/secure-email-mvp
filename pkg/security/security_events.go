package security

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	// Authentication Events
	EventFailedLogin           SecurityEventType = "failed_login"
	EventInvalidJWT            SecurityEventType = "invalid_jwt"
	EventExpiredJWT            SecurityEventType = "expired_jwt"
	EventJWTTampering          SecurityEventType = "jwt_tampering"
	EventInvalidTOTP           SecurityEventType = "invalid_totp"
	EventTOTPBruteForce        SecurityEventType = "totp_brute_force"
	EventPasswordBruteForce    SecurityEventType = "password_brute_force"
	
	// Authorization Events
	EventPrivilegeEscalation   SecurityEventType = "privilege_escalation"
	EventUnauthorizedAccess    SecurityEventType = "unauthorized_access"
	EventCrossTenantAccess     SecurityEventType = "cross_tenant_access"
	EventRoleManipulation      SecurityEventType = "role_manipulation"
	
	// Input Validation Events
	EventSQLInjection          SecurityEventType = "sql_injection"
	EventXSSAttempt            SecurityEventType = "xss_attempt"
	EventCSRFAttempt           SecurityEventType = "csrf_attempt"
	EventInvalidInput          SecurityEventType = "invalid_input"
	EventLargePayload          SecurityEventType = "large_payload"
	
	// Rate Limiting Events
	EventRateLimitExceeded     SecurityEventType = "rate_limit_exceeded"
	EventBruteForceAttempt     SecurityEventType = "brute_force_attempt"
	EventDDoSAttempt           SecurityEventType = "ddos_attempt"
	
	// Compliance & Export Events
	EventUnauthorizedExport    SecurityEventType = "unauthorized_export"
	EventCSVInjection          SecurityEventType = "csv_injection"
	EventExportRateLimit       SecurityEventType = "export_rate_limit"
	
	// General Security Events
	EventSuspiciousActivity    SecurityEventType = "suspicious_activity"
	EventSecurityViolation     SecurityEventType = "security_violation"
	EventSystemCompromise      SecurityEventType = "system_compromise"
)

// SecurityEvent represents a security event log entry
type SecurityEvent struct {
	ID             string            `json:"id"`
	EventType      SecurityEventType `json:"event_type"`
	Severity       string            `json:"severity"` // low, medium, high, critical
	UserID         *string           `json:"user_id,omitempty"`
	OrganizationID *string           `json:"organization_id,omitempty"`
	IPAddress      string            `json:"ip_address"`
	UserAgent      string            `json:"user_agent"`
	Endpoint       string            `json:"endpoint"`
	Method         string            `json:"method"`
	Details        json.RawMessage   `json:"details"`
	Timestamp      time.Time         `json:"timestamp"`
	CreatedAt      time.Time         `json:"created_at"`
}

// SecurityEventDetails contains structured details for different event types
type SecurityEventDetails struct {
	// Authentication details
	Email           string `json:"email,omitempty"`
	FailedAttempts  int    `json:"failed_attempts,omitempty"`
	LockoutDuration int    `json:"lockout_duration,omitempty"`
	
	// JWT details
	JWTToken        string `json:"jwt_token,omitempty"`
	JWTAlgorithm    string `json:"jwt_algorithm,omitempty"`
	JWTExpiration   string `json:"jwt_expiration,omitempty"`
	
	// TOTP details
	TOTPCode        string `json:"totp_code,omitempty"`
	TOTPSkew        int    `json:"totp_skew,omitempty"`
	
	// Authorization details
	RequestedRole   string `json:"requested_role,omitempty"`
	CurrentRole     string `json:"current_role,omitempty"`
	TargetOrgID     string `json:"target_org_id,omitempty"`
	UserOrgID       string `json:"user_org_id,omitempty"`
	
	// Input validation details
	InputValue      string `json:"input_value,omitempty"`
	InputType       string `json:"input_type,omitempty"`
	PayloadSize     int    `json:"payload_size,omitempty"`
	
	// Rate limiting details
	RequestCount    int    `json:"request_count,omitempty"`
	TimeWindow      int    `json:"time_window,omitempty"`
	LimitThreshold  int    `json:"limit_threshold,omitempty"`
	
	// Export details
	ExportFormat    string `json:"export_format,omitempty"`
	ExportFilters   string `json:"export_filters,omitempty"`
	
	// General details
	ErrorCode       string `json:"error_code,omitempty"`
	ErrorMessage    string `json:"error_message,omitempty"`
	AdditionalInfo  string `json:"additional_info,omitempty"`
}

// LogSecurityEvent logs a security event with structured details
func LogSecurityEvent(db *sql.DB, eventType SecurityEventType, severity string, details *SecurityEventDetails, userID, orgID *string, ipAddress, userAgent, endpoint, method string) error {
	id := uuid.New().String()
	
	// Convert details to JSON
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal security event details: %v", err)
	}
	
	query := `
		INSERT INTO security_events (
			id, event_type, severity, user_id, organization_id, ip_address, 
			user_agent, endpoint, method, details, timestamp, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err = db.Exec(query, id, eventType, severity, userID, orgID, ipAddress, 
		userAgent, endpoint, method, string(detailsJSON), time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to log security event: %v", err)
	}
	
	log.Printf("[SECURITY] Logged %s event (severity: %s) for IP %s", eventType, severity, ipAddress)
	return nil
}

// LogFailedLogin logs a failed login attempt
func LogFailedLogin(db *sql.DB, email, ipAddress, userAgent string, failedAttempts int) error {
	details := &SecurityEventDetails{
		Email:          email,
		FailedAttempts: failedAttempts,
		ErrorCode:      "AUTH_FAILED",
		ErrorMessage:   "Invalid credentials",
	}
	
	severity := "low"
	if failedAttempts > 5 {
		severity = "medium"
	}
	if failedAttempts > 10 {
		severity = "high"
	}
	
	return LogSecurityEvent(db, EventFailedLogin, severity, details, nil, nil, ipAddress, userAgent, "/api/auth/login", "POST")
}

// LogInvalidJWT logs an invalid JWT attempt
func LogInvalidJWT(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, token, algorithm, expiration, errorCode string) error {
	details := &SecurityEventDetails{
		JWTToken:      token,
		JWTAlgorithm:  algorithm,
		JWTExpiration: expiration,
		ErrorCode:     errorCode,
		ErrorMessage:  "Invalid JWT token",
	}
	
	severity := "medium"
	if errorCode == "JWT_TAMPERING" {
		severity = "high"
	}
	
	return LogSecurityEvent(db, EventInvalidJWT, severity, details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogExpiredJWT logs an expired JWT attempt
func LogExpiredJWT(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, token, expiration string) error {
	details := &SecurityEventDetails{
		JWTToken:      token,
		JWTExpiration: expiration,
		ErrorCode:     "JWT_EXPIRED",
		ErrorMessage:  "JWT token has expired",
	}
	
	return LogSecurityEvent(db, EventExpiredJWT, "low", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogJWTTampering logs JWT tampering attempts
func LogJWTTampering(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, token, tamperingType string) error {
	details := &SecurityEventDetails{
		JWTToken:      token,
		ErrorCode:     "JWT_TAMPERING",
		ErrorMessage:  "JWT token tampering detected",
		AdditionalInfo: tamperingType,
	}
	
	return LogSecurityEvent(db, EventJWTTampering, "high", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogInvalidTOTP logs an invalid TOTP attempt
func LogInvalidTOTP(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, code string, skew int, failedAttempts int) error {
	details := &SecurityEventDetails{
		TOTPCode:       code,
		TOTPSkew:       skew,
		FailedAttempts: failedAttempts,
		ErrorCode:      "TOTP_INVALID",
		ErrorMessage:   "Invalid TOTP code",
	}
	
	severity := "medium"
	if failedAttempts > 3 {
		severity = "high"
	}
	
	return LogSecurityEvent(db, EventInvalidTOTP, severity, details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogTOTPBruteForce logs TOTP brute force attempts
func LogTOTPBruteForce(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method string, attemptCount int) error {
	details := &SecurityEventDetails{
		FailedAttempts: attemptCount,
		ErrorCode:      "TOTP_BRUTE_FORCE",
		ErrorMessage:   "TOTP brute force attempt detected",
	}
	
	return LogSecurityEvent(db, EventTOTPBruteForce, "high", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogPrivilegeEscalation logs privilege escalation attempts
func LogPrivilegeEscalation(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, currentRole, requestedRole string) error {
	details := &SecurityEventDetails{
		CurrentRole:   currentRole,
		RequestedRole: requestedRole,
		ErrorCode:     "PRIVILEGE_ESCALATION",
		ErrorMessage:  "Privilege escalation attempt detected",
	}
	
	return LogSecurityEvent(db, EventPrivilegeEscalation, "high", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogUnauthorizedAccess logs unauthorized access attempts
func LogUnauthorizedAccess(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, reason string) error {
	details := &SecurityEventDetails{
		ErrorCode:    "UNAUTHORIZED_ACCESS",
		ErrorMessage: reason,
	}
	
	return LogSecurityEvent(db, EventUnauthorizedAccess, "medium", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogCrossTenantAccess logs cross-tenant access attempts
func LogCrossTenantAccess(db *sql.DB, userID, userOrgID *string, ipAddress, userAgent, endpoint, method, targetOrgID string) error {
	details := &SecurityEventDetails{
		UserOrgID:    *userOrgID,
		TargetOrgID:  targetOrgID,
		ErrorCode:    "CROSS_TENANT_ACCESS",
		ErrorMessage: "Cross-tenant access attempt detected",
	}
	
	return LogSecurityEvent(db, EventCrossTenantAccess, "high", details, userID, userOrgID, ipAddress, userAgent, endpoint, method)
}

// LogSQLInjection logs SQL injection attempts
func LogSQLInjection(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, inputValue, inputType string) error {
	details := &SecurityEventDetails{
		InputValue:   inputValue,
		InputType:    inputType,
		ErrorCode:    "SQL_INJECTION",
		ErrorMessage: "SQL injection attempt detected",
	}
	
	return LogSecurityEvent(db, EventSQLInjection, "critical", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogXSSAttempt logs XSS attempts
func LogXSSAttempt(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, inputValue, inputType string) error {
	details := &SecurityEventDetails{
		InputValue:   inputValue,
		InputType:    inputType,
		ErrorCode:    "XSS_ATTEMPT",
		ErrorMessage: "XSS attempt detected",
	}
	
	return LogSecurityEvent(db, EventXSSAttempt, "high", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogRateLimitExceeded logs rate limit violations
func LogRateLimitExceeded(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method string, requestCount, timeWindow, limitThreshold int) error {
	details := &SecurityEventDetails{
		RequestCount:   requestCount,
		TimeWindow:     timeWindow,
		LimitThreshold: limitThreshold,
		ErrorCode:      "RATE_LIMIT_EXCEEDED",
		ErrorMessage:   "Rate limit exceeded",
	}
	
	severity := "medium"
	if requestCount > limitThreshold*2 {
		severity = "high"
	}
	
	return LogSecurityEvent(db, EventRateLimitExceeded, severity, details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogUnauthorizedExport logs unauthorized export attempts
func LogUnauthorizedExport(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, exportFormat, exportFilters string) error {
	details := &SecurityEventDetails{
		ExportFormat:  exportFormat,
		ExportFilters: exportFilters,
		ErrorCode:     "UNAUTHORIZED_EXPORT",
		ErrorMessage:  "Unauthorized export attempt",
	}
	
	return LogSecurityEvent(db, EventUnauthorizedExport, "high", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogCSVInjection logs CSV injection attempts
func LogCSVInjection(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method, inputValue string) error {
	details := &SecurityEventDetails{
		InputValue:   inputValue,
		InputType:    "csv_export",
		ErrorCode:    "CSV_INJECTION",
		ErrorMessage: "CSV injection attempt detected",
	}
	
	return LogSecurityEvent(db, EventCSVInjection, "high", details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// LogLargePayload logs large payload attempts
func LogLargePayload(db *sql.DB, userID, orgID *string, ipAddress, userAgent, endpoint, method string, payloadSize int) error {
	details := &SecurityEventDetails{
		PayloadSize:  payloadSize,
		ErrorCode:    "LARGE_PAYLOAD",
		ErrorMessage: "Large payload attempt detected",
	}
	
	severity := "medium"
	if payloadSize > 1024*1024 { // 1MB
		severity = "high"
	}
	
	return LogSecurityEvent(db, EventLargePayload, severity, details, userID, orgID, ipAddress, userAgent, endpoint, method)
}

// GetSecurityEvents retrieves security events with optional filters
func GetSecurityEvents(db *sql.DB, userID, orgID *string, eventType SecurityEventType, severity string, limit, offset int) ([]*SecurityEvent, error) {
	query := `SELECT id, event_type, severity, user_id, organization_id, ip_address, user_agent, endpoint, method, details, timestamp, created_at FROM security_events WHERE 1=1`
	args := []interface{}{}
	
	if userID != nil {
		query += ` AND user_id = ?`
		args = append(args, *userID)
	}
	
	if orgID != nil {
		query += ` AND organization_id = ?`
		args = append(args, *orgID)
	}
	
	if eventType != "" {
		query += ` AND event_type = ?`
		args = append(args, eventType)
	}
	
	if severity != "" {
		query += ` AND severity = ?`
		args = append(args, severity)
	}
	
	query += ` ORDER BY timestamp DESC`
	
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	} else {
		query += ` LIMIT 100` // Default limit
	}
	
	if offset > 0 {
		query += ` OFFSET ?`
		args = append(args, offset)
	}
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query security events: %v", err)
	}
	defer rows.Close()
	
	var events []*SecurityEvent
	for rows.Next() {
		var event SecurityEvent
		var detailsStr sql.NullString
		var userIDStr, orgIDStr sql.NullString
		
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Severity,
			&userIDStr,
			&orgIDStr,
			&event.IPAddress,
			&event.UserAgent,
			&event.Endpoint,
			&event.Method,
			&detailsStr,
			&event.Timestamp,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan security event: %v", err)
		}
		
		if userIDStr.Valid {
			event.UserID = &userIDStr.String
		}
		if orgIDStr.Valid {
			event.OrganizationID = &orgIDStr.String
		}
		if detailsStr.Valid {
			event.Details = json.RawMessage(detailsStr.String)
		}
		
		events = append(events, &event)
	}
	
	return events, nil
}

// GetSecurityEventStats retrieves security event statistics
func GetSecurityEventStats(db *sql.DB, orgID *string, days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 30 // Default to 30 days
	}
	
	query := `
		SELECT 
			event_type,
			severity,
			COUNT(*) as count
		FROM security_events
		WHERE timestamp >= datetime('now', '-? days')
	`
	args := []interface{}{days}
	
	if orgID != nil {
		query += ` AND organization_id = ?`
		args = append(args, *orgID)
	}
	
	query += ` GROUP BY event_type, severity ORDER BY count DESC`
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query security event stats: %v", err)
	}
	defer rows.Close()
	
	stats := map[string]interface{}{
		"total_events": 0,
		"by_type":      map[string]int{},
		"by_severity":  map[string]int{},
		"critical":     0,
		"high":         0,
		"medium":       0,
		"low":          0,
	}
	
	for rows.Next() {
		var eventType, severity string
		var count int
		
		err := rows.Scan(&eventType, &severity, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan security event stat: %v", err)
		}
		
		stats["total_events"] = stats["total_events"].(int) + count
		
		// Count by type
		typeCount := stats["by_type"].(map[string]int)
		typeCount[eventType] = typeCount[eventType] + count
		
		// Count by severity
		severityCount := stats["by_severity"].(map[string]int)
		severityCount[severity] = severityCount[severity] + count
		
		// Count individual severities
		switch severity {
		case "critical":
			stats["critical"] = stats["critical"].(int) + count
		case "high":
			stats["high"] = stats["high"].(int) + count
		case "medium":
			stats["medium"] = stats["medium"].(int) + count
		case "low":
			stats["low"] = stats["low"].(int) + count
		}
	}
	
	return stats, nil
}
