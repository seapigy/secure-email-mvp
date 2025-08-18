package audit

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// EmailAccessLog represents a single access attempt log entry
type EmailAccessLog struct {
	ID           int64     `json:"id"`
	EmailID      string    `json:"email_id"`
	UserID       *string   `json:"user_id,omitempty"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent,omitempty"`
	Status       string    `json:"status"`
	AttemptCount int       `json:"attempt_count"`
	Result       string    `json:"result"` // success, failed_password, expired, burn_after_read, rate_limited, etc.
	CreatedAt    time.Time `json:"created_at"`
}

// RateLimitConfig holds configuration for rate limiting decryption attempts
type RateLimitConfig struct {
	MaxAttempts int           `json:"max_attempts"` // Maximum attempts allowed (default: 3)
	TimeWindow  time.Duration `json:"time_window"`  // Time window for rate limiting (default: 5 minutes)
}

// DefaultRateLimitConfig provides sensible defaults for Micro-Iteration 4.22
var DefaultRateLimitConfig = RateLimitConfig{
	MaxAttempts: 3,
	TimeWindow:  5 * time.Minute,
}

// EmailAccessAuditor handles audit logging and rate limiting for email access
// Enhanced for Micro-Iteration 4.22: security hardening with audit logging,
// rate-limiting decryption attempts, and concurrent access protection
type EmailAccessAuditor struct {
	db              *sql.DB
	rateLimitConfig RateLimitConfig
}

// NewEmailAccessAuditor creates a new email access auditor with enhanced security
func NewEmailAccessAuditor(db *sql.DB, config RateLimitConfig) *EmailAccessAuditor {
	return &EmailAccessAuditor{
		db:              db,
		rateLimitConfig: config,
	}
}

// LogAccess logs an email access attempt with enhanced details for Micro-Iteration 4.22
// Logs every email retrieval attempt with: timestamp, requesting IP, user agent,
// email_id, and result (success, failed password, expired, burn_after_read triggered, etc.)
func (a *EmailAccessAuditor) LogAccess(ctx context.Context, emailID, ipAddress string, userID *string, result string, userAgent string) error {
	// Get current attempt count for this IP/email combination
	attemptCount, err := a.getCurrentAttemptCount(ctx, emailID, ipAddress)
	if err != nil {
		log.Printf("Failed to get attempt count: %v", err)
		attemptCount = 1 // Default to 1 if we can't get the count
	} else {
		attemptCount++
	}

	// Insert the log entry with enhanced details
	query := `
		INSERT INTO email_access_logs (email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Determine status based on result
	status := "success"
	if result != "success" {
		status = "fail"
	}

	_, err = a.db.ExecContext(ctx, query, emailID, userID, ipAddress, userAgent, status, attemptCount, result, time.Now())
	if err != nil {
		return fmt.Errorf("failed to log email access: %w", err)
	}

	log.Printf("Email access logged: email_id=%s, user_id=%v, ip=%s, result=%s, attempt=%d, user_agent=%s",
		emailID, userID, ipAddress, result, attemptCount, userAgent)

	return nil
}

// CheckRateLimit checks if the current request should be rate limited
// Enhanced for Micro-Iteration 4.22: Track failed decryption attempts by IP and/or email_id
// Limit to 3 failed attempts per 5 minutes per IP (configurable)
func (a *EmailAccessAuditor) CheckRateLimit(ctx context.Context, emailID, ipAddress string) (bool, error) {
	// Count failed attempts within the time window
	query := `
		SELECT COUNT(*) 
		FROM email_access_logs 
		WHERE email_id = ? AND ip_address = ? AND status = 'fail'
		AND created_at > datetime('now', '-? seconds')
	`

	timeWindowSeconds := int(a.rateLimitConfig.TimeWindow.Seconds())

	var count int
	err := a.db.QueryRowContext(ctx, query, emailID, ipAddress, timeWindowSeconds).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	isLimited := count >= a.rateLimitConfig.MaxAttempts

	if isLimited {
		log.Printf("Rate limit exceeded: email_id=%s, ip=%s, failed_attempts=%d, limit=%d, window=%v",
			emailID, ipAddress, count, a.rateLimitConfig.MaxAttempts, a.rateLimitConfig.TimeWindow)
	}

	return isLimited, nil
}

// GetAccessLogs retrieves access logs for an email with enhanced details
// Enhanced for Micro-Iteration 4.22: Make retrieval logs queryable for admin debugging
func (a *EmailAccessAuditor) GetAccessLogs(ctx context.Context, emailID string, limit int) ([]EmailAccessLog, error) {
	query := `
		SELECT id, email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at
		FROM email_access_logs 
		WHERE email_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := a.db.QueryContext(ctx, query, emailID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get access logs: %w", err)
	}
	defer rows.Close()

	var logs []EmailAccessLog
	for rows.Next() {
		var log EmailAccessLog
		var userID sql.NullString
		var userAgent sql.NullString

		err := rows.Scan(&log.ID, &log.EmailID, &userID, &log.IPAddress, &userAgent, &log.Status, &log.AttemptCount, &log.Result, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access log: %w", err)
		}

		if userID.Valid {
			log.UserID = &userID.String
		}
		if userAgent.Valid {
			log.UserAgent = userAgent.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetAccessLogsByIP retrieves access logs for a specific IP address
// Enhanced for Micro-Iteration 4.22: Admin debugging capabilities
func (a *EmailAccessAuditor) GetAccessLogsByIP(ctx context.Context, ipAddress string, limit int) ([]EmailAccessLog, error) {
	query := `
		SELECT id, email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at
		FROM email_access_logs 
		WHERE ip_address = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := a.db.QueryContext(ctx, query, ipAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get access logs by IP: %w", err)
	}
	defer rows.Close()

	var logs []EmailAccessLog
	for rows.Next() {
		var log EmailAccessLog
		var userID sql.NullString
		var userAgent sql.NullString

		err := rows.Scan(&log.ID, &log.EmailID, &userID, &log.IPAddress, &userAgent, &log.Status, &log.AttemptCount, &log.Result, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access log: %w", err)
		}

		if userID.Valid {
			log.UserID = &userID.String
		}
		if userAgent.Valid {
			log.UserAgent = userAgent.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetFailedAttemptsSummary returns a summary of failed attempts for admin debugging
// Enhanced for Micro-Iteration 4.22: Admin debugging capabilities
func (a *EmailAccessAuditor) GetFailedAttemptsSummary(ctx context.Context, hours int) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_failed_attempts,
			COUNT(DISTINCT email_id) as unique_emails,
			COUNT(DISTINCT ip_address) as unique_ips,
			COUNT(DISTINCT user_id) as unique_users
		FROM email_access_logs 
		WHERE status = 'fail' 
		AND created_at > datetime('now', '-? hours')
	`

	var totalFailed, uniqueEmails, uniqueIPs, uniqueUsers int
	err := a.db.QueryRowContext(ctx, query, hours).Scan(&totalFailed, &uniqueEmails, &uniqueIPs, &uniqueUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed attempts summary: %w", err)
	}

	// Get top IPs with failed attempts
	topIPsQuery := `
		SELECT ip_address, COUNT(*) as attempt_count
		FROM email_access_logs 
		WHERE status = 'fail' 
		AND created_at > datetime('now', '-? hours')
		GROUP BY ip_address 
		ORDER BY attempt_count DESC 
		LIMIT 10
	`

	rows, err := a.db.QueryContext(ctx, topIPsQuery, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to get top IPs: %w", err)
	}
	defer rows.Close()

	var topIPs []map[string]interface{}
	for rows.Next() {
		var ip string
		var count int
		if err := rows.Scan(&ip, &count); err != nil {
			return nil, fmt.Errorf("failed to scan top IP: %w", err)
		}
		topIPs = append(topIPs, map[string]interface{}{
			"ip_address":    ip,
			"attempt_count": count,
		})
	}

	return map[string]interface{}{
		"total_failed_attempts": totalFailed,
		"unique_emails":         uniqueEmails,
		"unique_ips":            uniqueIPs,
		"unique_users":          uniqueUsers,
		"top_ips":               topIPs,
		"time_window_hours":     hours,
	}, nil
}

// CleanupOldLogs removes access logs older than the specified duration
// Enhanced for Micro-Iteration 4.22: Automatic cleanup for performance
func (a *EmailAccessAuditor) CleanupOldLogs(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM email_access_logs 
		WHERE created_at < datetime('now', '-? seconds')
	`

	olderThanSeconds := int(olderThan.Seconds())
	result, err := a.db.ExecContext(ctx, query, olderThanSeconds)
	if err != nil {
		return fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Cleaned up %d old email access logs (older than %v)", rowsAffected, olderThan)

	return nil
}

// getCurrentAttemptCount gets the current attempt count for an IP/email combination
func (a *EmailAccessAuditor) getCurrentAttemptCount(ctx context.Context, emailID, ipAddress string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM email_access_logs 
		WHERE email_id = ? AND ip_address = ?
	`

	var count int
	err := a.db.QueryRowContext(ctx, query, emailID, ipAddress).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get attempt count: %w", err)
	}

	return count, nil
}

// GetRateLimitConfig returns the current rate limit configuration
func (a *EmailAccessAuditor) GetRateLimitConfig() RateLimitConfig {
	return a.rateLimitConfig
}

// UpdateRateLimitConfig updates the rate limit configuration
func (a *EmailAccessAuditor) UpdateRateLimitConfig(config RateLimitConfig) {
	a.rateLimitConfig = config
	log.Printf("Updated rate limit config: max_attempts=%d, time_window=%v",
		config.MaxAttempts, config.TimeWindow)
}

// GetSenderAccessInsights returns anonymized access insights for a specific email
// This is used for sender-side insights (Micro-Iteration 4.23)
func (a *EmailAccessAuditor) GetSenderAccessInsights(ctx context.Context, emailID string) (map[string]interface{}, error) {
	// Get the most recent access log entry
	var lastAccess *EmailAccessLog
	recentLogs, err := a.GetAccessLogs(ctx, emailID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent access logs: %w", err)
	}

	if len(recentLogs) > 0 {
		lastAccess = &recentLogs[0]
	}

	// Get total access count
	var totalAccessCount int
	err = a.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM email_access_logs 
		WHERE email_id = ? AND status = 'success'
	`, emailID).Scan(&totalAccessCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total access count: %w", err)
	}

	// Prepare response
	insights := map[string]interface{}{
		"email_id":           emailID,
		"total_access_count": totalAccessCount,
	}

	// Add last access information if available
	if lastAccess != nil {
		// Anonymize the IP address for privacy compliance
		anonymizedIP, err := AnonymizeIP(lastAccess.IPAddress)
		if err != nil {
			// If anonymization fails, use a generic indicator
			anonymizedIP = "Unknown"
		}

		insights["last_accessed_at"] = lastAccess.CreatedAt
		insights["last_access_ip"] = anonymizedIP
		insights["last_access_result"] = lastAccess.Result
	} else {
		insights["last_accessed_at"] = nil
		insights["last_access_ip"] = nil
		insights["last_access_result"] = nil
	}

	return insights, nil
}

// GetAccessLogsForAdmin retrieves access logs with filtering and pagination for admin use
// This is used for admin access log query endpoint (Micro-Iteration 4.23)
func (a *EmailAccessAuditor) GetAccessLogsForAdmin(ctx context.Context, filters map[string]string, limit, offset int) ([]EmailAccessLog, error) {
	// Build the base query
	query := `
		SELECT id, email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at
		FROM email_access_logs 
		WHERE 1=1
	`

	var args []interface{}
	argIndex := 1

	// Add filters
	if emailID, ok := filters["email_id"]; ok && emailID != "" {
		query += " AND email_id = ?"
		args = append(args, emailID)
		argIndex++
	}

	if userID, ok := filters["user_id"]; ok && userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
		argIndex++
	}

	if result, ok := filters["result"]; ok && result != "" {
		query += " AND result = ?"
		args = append(args, result)
		argIndex++
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += " AND created_at >= ?"
		args = append(args, startDate)
		argIndex++
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += " AND created_at <= ?"
		args = append(args, endDate)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Execute query
	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get access logs for admin: %w", err)
	}
	defer rows.Close()

	var logs []EmailAccessLog
	for rows.Next() {
		var log EmailAccessLog
		var userID sql.NullString
		var userAgent sql.NullString

		err := rows.Scan(&log.ID, &log.EmailID, &userID, &log.IPAddress, &userAgent, &log.Status, &log.AttemptCount, &log.Result, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access log: %w", err)
		}

		if userID.Valid {
			log.UserID = &userID.String
		}
		if userAgent.Valid {
			log.UserAgent = userAgent.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetAccessLogsCountForAdmin returns the total count of access logs matching the filters
// This is used for pagination in admin access log query endpoint (Micro-Iteration 4.23)
func (a *EmailAccessAuditor) GetAccessLogsCountForAdmin(ctx context.Context, filters map[string]string) (int, error) {
	// Build the base query
	query := `
		SELECT COUNT(*)
		FROM email_access_logs 
		WHERE 1=1
	`

	var args []interface{}

	// Add filters
	if emailID, ok := filters["email_id"]; ok && emailID != "" {
		query += " AND email_id = ?"
		args = append(args, emailID)
	}

	if userID, ok := filters["user_id"]; ok && userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
	}

	if result, ok := filters["result"]; ok && result != "" {
		query += " AND result = ?"
		args = append(args, result)
	}

	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += " AND created_at >= ?"
		args = append(args, startDate)
	}

	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += " AND created_at <= ?"
		args = append(args, endDate)
	}

	// Execute query
	var count int
	err := a.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get access logs count for admin: %w", err)
	}

	return count, nil
}
