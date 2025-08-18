package email

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// RetentionAnalyticsService provides comprehensive retention analytics functionality
type RetentionAnalyticsService struct {
	db *sql.DB
}

// RetentionAnalytics represents comprehensive retention analytics data
type RetentionAnalytics struct {
	OverallStats    OverallRetentionStats  `json:"overall_stats"`
	UserStats       []UserRetentionStats   `json:"user_stats"`
	RetentionTrends RetentionTrends        `json:"retention_trends"`
	CleanupLogs     []CleanupLogEntry      `json:"cleanup_logs"`
	ExpirationStats ExpirationDistribution `json:"expiration_stats"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// OverallRetentionStats represents overall retention statistics
type OverallRetentionStats struct {
	TotalEmailsSent           int     `json:"total_emails_sent"`
	TotalEmailsExpired        int     `json:"total_emails_expired"`
	TotalEmailsSelfDestructed int     `json:"total_emails_self_destructed"`
	TotalEmailsDeleted        int     `json:"total_emails_deleted"`
	TotalEmailsActive         int     `json:"total_emails_active"`
	AverageRetentionDays      float64 `json:"average_retention_days"`
	ExpirationRate            float64 `json:"expiration_rate"`    // Percentage of emails that expired
	SelfDestructRate          float64 `json:"self_destruct_rate"` // Percentage of emails that self-destructed
	CleanupRate               float64 `json:"cleanup_rate"`       // Percentage of emails cleaned up
}

// UserRetentionStats represents retention statistics for a specific user
type UserRetentionStats struct {
	UserID               string  `json:"user_id"`
	EmailsSent           int     `json:"emails_sent"`
	EmailsExpired        int     `json:"emails_expired"`
	EmailsSelfDestructed int     `json:"emails_self_destructed"`
	EmailsDeleted        int     `json:"emails_deleted"`
	EmailsActive         int     `json:"emails_active"`
	AverageRetentionDays float64 `json:"average_retention_days"`
	ExpirationRate       float64 `json:"expiration_rate"`
	SelfDestructRate     float64 `json:"self_destruct_rate"`
	CleanupRate          float64 `json:"cleanup_rate"`
}

// RetentionTrends represents retention trends over time
type RetentionTrends struct {
	DailyStats   []DailyRetentionStats   `json:"daily_stats"`
	WeeklyStats  []WeeklyRetentionStats  `json:"weekly_stats"`
	MonthlyStats []MonthlyRetentionStats `json:"monthly_stats"`
}

// DailyRetentionStats represents daily retention statistics
type DailyRetentionStats struct {
	Date                 time.Time `json:"date"`
	EmailsSent           int       `json:"emails_sent"`
	EmailsExpired        int       `json:"emails_expired"`
	EmailsDeleted        int       `json:"emails_deleted"`
	AverageRetentionDays float64   `json:"average_retention_days"`
}

// WeeklyRetentionStats represents weekly retention statistics
type WeeklyRetentionStats struct {
	WeekStart            time.Time `json:"week_start"`
	EmailsSent           int       `json:"emails_sent"`
	EmailsExpired        int       `json:"emails_expired"`
	EmailsDeleted        int       `json:"emails_deleted"`
	AverageRetentionDays float64   `json:"average_retention_days"`
}

// MonthlyRetentionStats represents monthly retention statistics
type MonthlyRetentionStats struct {
	Month                time.Time `json:"month"`
	EmailsSent           int       `json:"emails_sent"`
	EmailsExpired        int       `json:"emails_expired"`
	EmailsDeleted        int       `json:"emails_deleted"`
	AverageRetentionDays float64   `json:"average_retention_days"`
}

// CleanupLogEntry represents a cleanup log entry
type CleanupLogEntry struct {
	LogID            string    `json:"log_id"`
	EmailID          string    `json:"email_id"`
	SenderID         string    `json:"sender_id"`
	CleanupReason    string    `json:"cleanup_reason"`
	CleanupTime      time.Time `json:"cleanup_time"`
	Initiator        string    `json:"initiator"` // "worker", "manual"
	EmailsProcessed  int       `json:"emails_processed"`
	EmailsDeleted    int       `json:"emails_deleted"`
	EmailsSkipped    int       `json:"emails_skipped"`
	AuditLogsDeleted int       `json:"audit_logs_deleted"`
	Duration         string    `json:"duration"`
}

// ExpirationDistribution represents the distribution of email expiration periods
type ExpirationDistribution struct {
	LessThan1Day        int `json:"less_than_1_day"`
	OneToSevenDays      int `json:"one_to_seven_days"`
	OneToFourWeeks      int `json:"one_to_four_weeks"`
	OneToThreeMonths    int `json:"one_to_three_months"`
	MoreThanThreeMonths int `json:"more_than_three_months"`
	NoExpiration        int `json:"no_expiration"`
}

// AnalyticsFilters represents filters for analytics queries
type AnalyticsFilters struct {
	UserID    string    `json:"user_id,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Status    string    `json:"status,omitempty"` // "active", "expired", "deleted", "self_destructed"
}

// NewRetentionAnalyticsService creates a new retention analytics service
func NewRetentionAnalyticsService(db *sql.DB) *RetentionAnalyticsService {
	return &RetentionAnalyticsService{
		db: db,
	}
}

// GetRetentionAnalytics retrieves comprehensive retention analytics
func (ras *RetentionAnalyticsService) GetRetentionAnalytics(ctx context.Context, filters AnalyticsFilters) (*RetentionAnalytics, error) {
	analytics := &RetentionAnalytics{
		GeneratedAt: time.Now(),
	}

	// Get overall statistics
	overallStats, err := ras.getOverallRetentionStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}
	analytics.OverallStats = *overallStats

	// Get user statistics
	userStats, err := ras.getUserRetentionStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	analytics.UserStats = userStats

	// Get retention trends
	trends, err := ras.getRetentionTrends(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get retention trends: %w", err)
	}
	analytics.RetentionTrends = *trends

	// Get cleanup logs
	cleanupLogs, err := ras.getCleanupLogs(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleanup logs: %w", err)
	}
	analytics.CleanupLogs = cleanupLogs

	// Get expiration distribution
	expirationStats, err := ras.getExpirationDistribution(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiration distribution: %w", err)
	}
	analytics.ExpirationStats = *expirationStats

	return analytics, nil
}

// getOverallRetentionStats retrieves overall retention statistics
func (ras *RetentionAnalyticsService) getOverallRetentionStats(ctx context.Context, filters AnalyticsFilters) (*OverallRetentionStats, error) {
	stats := &OverallRetentionStats{}

	// Build base query with filters
	baseQuery := "FROM emails WHERE 1=1"
	args := []interface{}{}

	if filters.UserID != "" {
		baseQuery += " AND sender_id = ?"
		args = append(args, filters.UserID)
	}

	if !filters.StartDate.IsZero() {
		baseQuery += " AND created_at >= ?"
		args = append(args, filters.StartDate.Format("2006-01-02 15:04:05"))
	}

	if !filters.EndDate.IsZero() {
		baseQuery += " AND created_at <= ?"
		args = append(args, filters.EndDate.Format("2006-01-02 15:04:05"))
	}

	// Total emails sent
	query := "SELECT COUNT(*) " + baseQuery
	err := ras.db.QueryRowContext(ctx, query, args...).Scan(&stats.TotalEmailsSent)
	if err != nil {
		return nil, fmt.Errorf("failed to count total emails: %w", err)
	}

	// Total emails expired
	expiredQuery := "SELECT COUNT(*) " + baseQuery + " AND expires_at IS NOT NULL AND expires_at <= datetime('now')"
	err = ras.db.QueryRowContext(ctx, expiredQuery, args...).Scan(&stats.TotalEmailsExpired)
	if err != nil {
		return nil, fmt.Errorf("failed to count expired emails: %w", err)
	}

	// Total emails self-destructed
	selfDestructQuery := "SELECT COUNT(*) " + baseQuery + " AND self_destructed = 1"
	err = ras.db.QueryRowContext(ctx, selfDestructQuery, args...).Scan(&stats.TotalEmailsSelfDestructed)
	if err != nil {
		return nil, fmt.Errorf("failed to count self-destructed emails: %w", err)
	}

	// Total emails deleted (no content)
	deletedQuery := "SELECT COUNT(*) " + baseQuery + " AND encrypted_blob_url IS NULL"
	err = ras.db.QueryRowContext(ctx, deletedQuery, args...).Scan(&stats.TotalEmailsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to count deleted emails: %w", err)
	}

	// Total emails active
	activeQuery := "SELECT COUNT(*) " + baseQuery + " AND encrypted_blob_url IS NOT NULL AND (expires_at IS NULL OR expires_at > datetime('now')) AND self_destructed = 0"
	err = ras.db.QueryRowContext(ctx, activeQuery, args...).Scan(&stats.TotalEmailsActive)
	if err != nil {
		return nil, fmt.Errorf("failed to count active emails: %w", err)
	}

	// Calculate average retention days
	if stats.TotalEmailsSent > 0 {
		avgQuery := `
			SELECT AVG(
				CASE 
					WHEN expires_at IS NOT NULL THEN 
						CAST((julianday(expires_at) - julianday(created_at)) AS REAL)
					ELSE 
						CAST((julianday('now') - julianday(created_at)) AS REAL)
				END
			) ` + baseQuery
		err = ras.db.QueryRowContext(ctx, avgQuery, args...).Scan(&stats.AverageRetentionDays)
		if err != nil {
			log.Printf("Failed to calculate average retention days: %v", err)
			stats.AverageRetentionDays = 0
		}
	}

	// Calculate rates
	if stats.TotalEmailsSent > 0 {
		stats.ExpirationRate = float64(stats.TotalEmailsExpired) / float64(stats.TotalEmailsSent) * 100
		stats.SelfDestructRate = float64(stats.TotalEmailsSelfDestructed) / float64(stats.TotalEmailsSent) * 100
		stats.CleanupRate = float64(stats.TotalEmailsDeleted) / float64(stats.TotalEmailsSent) * 100
	}

	return stats, nil
}

// getUserRetentionStats retrieves retention statistics per user
func (ras *RetentionAnalyticsService) getUserRetentionStats(ctx context.Context, filters AnalyticsFilters) ([]UserRetentionStats, error) {
	query := `
		SELECT 
			sender_id,
			COUNT(*) as total_sent,
			COUNT(CASE WHEN expires_at IS NOT NULL AND expires_at <= datetime('now') THEN 1 END) as expired,
			COUNT(CASE WHEN self_destructed = 1 THEN 1 END) as self_destructed,
			COUNT(CASE WHEN encrypted_blob_url IS NULL THEN 1 END) as deleted,
			COUNT(CASE WHEN encrypted_blob_url IS NOT NULL AND (expires_at IS NULL OR expires_at > datetime('now')) AND self_destructed = 0 THEN 1 END) as active
		FROM emails
		WHERE 1=1
	`

	args := []interface{}{}

	if filters.UserID != "" {
		query += " AND sender_id = ?"
		args = append(args, filters.UserID)
	}

	if !filters.StartDate.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filters.StartDate.Format("2006-01-02 15:04:05"))
	}

	if !filters.EndDate.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filters.EndDate.Format("2006-01-02 15:04:05"))
	}

	query += " GROUP BY sender_id ORDER BY total_sent DESC"

	rows, err := ras.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user stats: %w", err)
	}
	defer rows.Close()

	var userStats []UserRetentionStats
	for rows.Next() {
		var stat UserRetentionStats
		err := rows.Scan(
			&stat.UserID, &stat.EmailsSent, &stat.EmailsExpired,
			&stat.EmailsSelfDestructed, &stat.EmailsDeleted, &stat.EmailsActive,
		)
		if err != nil {
			log.Printf("Error scanning user stat row: %v", err)
			continue
		}

		// Calculate rates
		if stat.EmailsSent > 0 {
			stat.ExpirationRate = float64(stat.EmailsExpired) / float64(stat.EmailsSent) * 100
			stat.SelfDestructRate = float64(stat.EmailsSelfDestructed) / float64(stat.EmailsSent) * 100
			stat.CleanupRate = float64(stat.EmailsDeleted) / float64(stat.EmailsSent) * 100
		}

		// Calculate average retention days for this user
		avgQuery := `
			SELECT AVG(
				CASE 
					WHEN expires_at IS NOT NULL THEN 
						CAST((julianday(expires_at) - julianday(created_at)) AS REAL)
					ELSE 
						CAST((julianday('now') - julianday(created_at)) AS REAL)
				END
			) FROM emails WHERE sender_id = ?
		`
		err = ras.db.QueryRowContext(ctx, avgQuery, stat.UserID).Scan(&stat.AverageRetentionDays)
		if err != nil {
			log.Printf("Failed to calculate average retention days for user %s: %v", stat.UserID, err)
			stat.AverageRetentionDays = 0
		}

		userStats = append(userStats, stat)
	}

	return userStats, nil
}

// getRetentionTrends retrieves retention trends over time
func (ras *RetentionAnalyticsService) getRetentionTrends(ctx context.Context, filters AnalyticsFilters) (*RetentionTrends, error) {
	trends := &RetentionTrends{}

	// Get daily stats for the last 30 days
	dailyStats, err := ras.getDailyRetentionStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	trends.DailyStats = dailyStats

	// Get weekly stats for the last 12 weeks
	weeklyStats, err := ras.getWeeklyRetentionStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly stats: %w", err)
	}
	trends.WeeklyStats = weeklyStats

	// Get monthly stats for the last 12 months
	monthlyStats, err := ras.getMonthlyRetentionStats(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly stats: %w", err)
	}
	trends.MonthlyStats = monthlyStats

	return trends, nil
}

// getDailyRetentionStats retrieves daily retention statistics
func (ras *RetentionAnalyticsService) getDailyRetentionStats(ctx context.Context, filters AnalyticsFilters) ([]DailyRetentionStats, error) {
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as emails_sent,
			COUNT(CASE WHEN expires_at IS NOT NULL AND expires_at <= datetime('now') THEN 1 END) as emails_expired,
			COUNT(CASE WHEN encrypted_blob_url IS NULL THEN 1 END) as emails_deleted,
			AVG(
				CASE 
					WHEN expires_at IS NOT NULL THEN 
						CAST((julianday(expires_at) - julianday(created_at)) AS REAL)
					ELSE 
						CAST((julianday('now') - julianday(created_at)) AS REAL)
				END
			) as avg_retention_days
		FROM emails
		WHERE created_at >= datetime('now', '-30 days')
	`

	args := []interface{}{}

	if filters.UserID != "" {
		query += " AND sender_id = ?"
		args = append(args, filters.UserID)
	}

	query += " GROUP BY DATE(created_at) ORDER BY date DESC"

	rows, err := ras.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily stats: %w", err)
	}
	defer rows.Close()

	var dailyStats []DailyRetentionStats
	for rows.Next() {
		var stat DailyRetentionStats
		var dateStr string
		var avgRetention sql.NullFloat64

		err := rows.Scan(&dateStr, &stat.EmailsSent, &stat.EmailsExpired, &stat.EmailsDeleted, &avgRetention)
		if err != nil {
			log.Printf("Error scanning daily stat row: %v", err)
			continue
		}

		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			stat.Date = date
		}

		if avgRetention.Valid {
			stat.AverageRetentionDays = avgRetention.Float64
		}

		dailyStats = append(dailyStats, stat)
	}

	return dailyStats, nil
}

// getWeeklyRetentionStats retrieves weekly retention statistics
func (ras *RetentionAnalyticsService) getWeeklyRetentionStats(ctx context.Context, filters AnalyticsFilters) ([]WeeklyRetentionStats, error) {
	query := `
		SELECT 
			DATE(created_at, 'weekday 0', '-6 days') as week_start,
			COUNT(*) as emails_sent,
			COUNT(CASE WHEN expires_at IS NOT NULL AND expires_at <= datetime('now') THEN 1 END) as emails_expired,
			COUNT(CASE WHEN encrypted_blob_url IS NULL THEN 1 END) as emails_deleted,
			AVG(
				CASE 
					WHEN expires_at IS NOT NULL THEN 
						CAST((julianday(expires_at) - julianday(created_at)) AS REAL)
					ELSE 
						CAST((julianday('now') - julianday(created_at)) AS REAL)
				END
			) as avg_retention_days
		FROM emails
		WHERE created_at >= datetime('now', '-84 days')
	`

	args := []interface{}{}

	if filters.UserID != "" {
		query += " AND sender_id = ?"
		args = append(args, filters.UserID)
	}

	query += " GROUP BY DATE(created_at, 'weekday 0', '-6 days') ORDER BY week_start DESC"

	rows, err := ras.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query weekly stats: %w", err)
	}
	defer rows.Close()

	var weeklyStats []WeeklyRetentionStats
	for rows.Next() {
		var stat WeeklyRetentionStats
		var dateStr string
		var avgRetention sql.NullFloat64

		err := rows.Scan(&dateStr, &stat.EmailsSent, &stat.EmailsExpired, &stat.EmailsDeleted, &avgRetention)
		if err != nil {
			log.Printf("Error scanning weekly stat row: %v", err)
			continue
		}

		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			stat.WeekStart = date
		}

		if avgRetention.Valid {
			stat.AverageRetentionDays = avgRetention.Float64
		}

		weeklyStats = append(weeklyStats, stat)
	}

	return weeklyStats, nil
}

// getMonthlyRetentionStats retrieves monthly retention statistics
func (ras *RetentionAnalyticsService) getMonthlyRetentionStats(ctx context.Context, filters AnalyticsFilters) ([]MonthlyRetentionStats, error) {
	query := `
		SELECT 
			DATE(created_at, 'start of month') as month,
			COUNT(*) as emails_sent,
			COUNT(CASE WHEN expires_at IS NOT NULL AND expires_at <= datetime('now') THEN 1 END) as emails_expired,
			COUNT(CASE WHEN encrypted_blob_url IS NULL THEN 1 END) as emails_deleted,
			AVG(
				CASE 
					WHEN expires_at IS NOT NULL THEN 
						CAST((julianday(expires_at) - julianday(created_at)) AS REAL)
					ELSE 
						CAST((julianday('now') - julianday(created_at)) AS REAL)
				END
			) as avg_retention_days
		FROM emails
		WHERE created_at >= datetime('now', '-365 days')
	`

	args := []interface{}{}

	if filters.UserID != "" {
		query += " AND sender_id = ?"
		args = append(args, filters.UserID)
	}

	query += " GROUP BY DATE(created_at, 'start of month') ORDER BY month DESC"

	rows, err := ras.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly stats: %w", err)
	}
	defer rows.Close()

	var monthlyStats []MonthlyRetentionStats
	for rows.Next() {
		var stat MonthlyRetentionStats
		var dateStr string
		var avgRetention sql.NullFloat64

		err := rows.Scan(&dateStr, &stat.EmailsSent, &stat.EmailsExpired, &stat.EmailsDeleted, &avgRetention)
		if err != nil {
			log.Printf("Error scanning monthly stat row: %v", err)
			continue
		}

		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			stat.Month = date
		}

		if avgRetention.Valid {
			stat.AverageRetentionDays = avgRetention.Float64
		}

		monthlyStats = append(monthlyStats, stat)
	}

	return monthlyStats, nil
}

// getCleanupLogs retrieves cleanup log entries
func (ras *RetentionAnalyticsService) getCleanupLogs(ctx context.Context, filters AnalyticsFilters) ([]CleanupLogEntry, error) {
	query := `
		SELECT 
			log_id, email_id, sender_id, cleanup_reason, cleanup_time, initiator,
			emails_processed, emails_deleted, emails_skipped, audit_logs_deleted, duration
		FROM cleanup_logs
		WHERE 1=1
	`

	args := []interface{}{}

	if filters.UserID != "" {
		query += " AND sender_id = ?"
		args = append(args, filters.UserID)
	}

	if !filters.StartDate.IsZero() {
		query += " AND cleanup_time >= ?"
		args = append(args, filters.StartDate.Format("2006-01-02 15:04:05"))
	}

	if !filters.EndDate.IsZero() {
		query += " AND cleanup_time <= ?"
		args = append(args, filters.EndDate.Format("2006-01-02 15:04:05"))
	}

	query += " ORDER BY cleanup_time DESC LIMIT 100"

	rows, err := ras.db.QueryContext(ctx, query, args...)
	if err != nil {
		// If table doesn't exist, return empty slice
		log.Printf("Cleanup logs table may not exist: %v", err)
		return []CleanupLogEntry{}, nil
	}
	defer rows.Close()

	var cleanupLogs []CleanupLogEntry
	for rows.Next() {
		var logEntry CleanupLogEntry
		var cleanupTimeStr string

		err := rows.Scan(
			&logEntry.LogID, &logEntry.EmailID, &logEntry.SenderID, &logEntry.CleanupReason,
			&cleanupTimeStr, &logEntry.Initiator, &logEntry.EmailsProcessed, &logEntry.EmailsDeleted,
			&logEntry.EmailsSkipped, &logEntry.AuditLogsDeleted, &logEntry.Duration,
		)
		if err != nil {
			log.Printf("Error scanning cleanup log row: %v", err)
			continue
		}

		if cleanupTime, err := time.Parse("2006-01-02 15:04:05", cleanupTimeStr); err == nil {
			logEntry.CleanupTime = cleanupTime
		}

		cleanupLogs = append(cleanupLogs, logEntry)
	}

	return cleanupLogs, nil
}

// getExpirationDistribution retrieves the distribution of email expiration periods
func (ras *RetentionAnalyticsService) getExpirationDistribution(ctx context.Context, filters AnalyticsFilters) (*ExpirationDistribution, error) {
	distribution := &ExpirationDistribution{}

	baseQuery := "FROM emails WHERE 1=1"
	args := []interface{}{}

	if filters.UserID != "" {
		baseQuery += " AND sender_id = ?"
		args = append(args, filters.UserID)
	}

	if !filters.StartDate.IsZero() {
		baseQuery += " AND created_at >= ?"
		args = append(args, filters.StartDate.Format("2006-01-02 15:04:05"))
	}

	if !filters.EndDate.IsZero() {
		baseQuery += " AND created_at <= ?"
		args = append(args, filters.EndDate.Format("2006-01-02 15:04:05"))
	}

	// Less than 1 day
	query := "SELECT COUNT(*) " + baseQuery + " AND expires_at IS NOT NULL AND julianday(expires_at) - julianday(created_at) < 1"
	err := ras.db.QueryRowContext(ctx, query, args...).Scan(&distribution.LessThan1Day)
	if err != nil {
		log.Printf("Failed to count emails expiring in less than 1 day: %v", err)
	}

	// One to seven days
	query = "SELECT COUNT(*) " + baseQuery + " AND expires_at IS NOT NULL AND julianday(expires_at) - julianday(created_at) >= 1 AND julianday(expires_at) - julianday(created_at) <= 7"
	err = ras.db.QueryRowContext(ctx, query, args...).Scan(&distribution.OneToSevenDays)
	if err != nil {
		log.Printf("Failed to count emails expiring in 1-7 days: %v", err)
	}

	// One to four weeks
	query = "SELECT COUNT(*) " + baseQuery + " AND expires_at IS NOT NULL AND julianday(expires_at) - julianday(created_at) > 7 AND julianday(expires_at) - julianday(created_at) <= 28"
	err = ras.db.QueryRowContext(ctx, query, args...).Scan(&distribution.OneToFourWeeks)
	if err != nil {
		log.Printf("Failed to count emails expiring in 1-4 weeks: %v", err)
	}

	// One to three months
	query = "SELECT COUNT(*) " + baseQuery + " AND expires_at IS NOT NULL AND julianday(expires_at) - julianday(created_at) > 28 AND julianday(expires_at) - julianday(created_at) <= 90"
	err = ras.db.QueryRowContext(ctx, query, args...).Scan(&distribution.OneToThreeMonths)
	if err != nil {
		log.Printf("Failed to count emails expiring in 1-3 months: %v", err)
	}

	// More than three months
	query = "SELECT COUNT(*) " + baseQuery + " AND expires_at IS NOT NULL AND julianday(expires_at) - julianday(created_at) > 90"
	err = ras.db.QueryRowContext(ctx, query, args...).Scan(&distribution.MoreThanThreeMonths)
	if err != nil {
		log.Printf("Failed to count emails expiring in more than 3 months: %v", err)
	}

	// No expiration
	query = "SELECT COUNT(*) " + baseQuery + " AND expires_at IS NULL"
	err = ras.db.QueryRowContext(ctx, query, args...).Scan(&distribution.NoExpiration)
	if err != nil {
		log.Printf("Failed to count emails with no expiration: %v", err)
	}

	return distribution, nil
}
