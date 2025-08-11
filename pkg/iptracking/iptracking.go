package iptracking

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// IPTrackingConfig holds configuration for IP-based tracking
type IPTrackingConfig struct {
	MaxAttempts           int           // Maximum failed attempts before lockout (default: 5)
	LockoutDuration       time.Duration // Lockout duration (default: 30 minutes)
	CleanupOlderThan      time.Duration // Cleanup records older than this (default: 24 hours)
	AttemptWindowDuration time.Duration // Time window for counting attempts (default: 15 minutes)
}

// DefaultConfig returns the default IP tracking configuration
func DefaultConfig() *IPTrackingConfig {
	return &IPTrackingConfig{
		MaxAttempts:           5,
		LockoutDuration:       30 * time.Minute,
		CleanupOlderThan:      24 * time.Hour,
		AttemptWindowDuration: 15 * time.Minute,
	}
}

// IPTrackingService provides methods for managing IP-based access tracking
type IPTrackingService struct {
	db     *sql.DB
	config *IPTrackingConfig
}

// NewIPTrackingService creates a new IP tracking service
func NewIPTrackingService(db *sql.DB) *IPTrackingService {
	return &IPTrackingService{
		db:     db,
		config: DefaultConfig(),
	}
}

// NewIPTrackingServiceWithConfig creates a new IP tracking service with custom configuration
func NewIPTrackingServiceWithConfig(db *sql.DB, config *IPTrackingConfig) *IPTrackingService {
	return &IPTrackingService{
		db:     db,
		config: config,
	}
}

// CheckIPLockout checks if an IP address is currently locked out
// Returns true if locked out, false if access is allowed
func (s *IPTrackingService) CheckIPLockout(ipAddress string) (bool, error) {
	var lockoutUntil *time.Time
	var failedAttempts int

	query := `
		SELECT lockout_until, failed_attempts
		FROM ip_access_attempts 
		WHERE ip_address = ?
	`

	err := s.db.QueryRow(query, ipAddress).Scan(&lockoutUntil, &failedAttempts)
	if err != nil {
		if err == sql.ErrNoRows {
			// IP not found, allow access
			return false, nil
		}
		return false, fmt.Errorf("failed to check IP lockout: %w", err)
	}

	// If lockout is set and not expired, deny access
	if lockoutUntil != nil && time.Now().Before(*lockoutUntil) {
		return true, nil
	}

	// If lockout has expired, clear it
	if lockoutUntil != nil && time.Now().After(*lockoutUntil) {
		err = s.clearIPLockout(ipAddress)
		if err != nil {
			return false, fmt.Errorf("failed to clear expired IP lockout: %w", err)
		}
	}

	return false, nil
}

// IncrementFailedAttempt increments the failed attempt count for an IP address
// and applies lockout if the threshold is reached
func (s *IPTrackingService) IncrementFailedAttempt(ipAddress string) error {
	// Check if IP already exists
	var failedAttempts int
	var lastAttemptAt time.Time

	query := `
		SELECT failed_attempts, last_attempt_at
		FROM ip_access_attempts 
		WHERE ip_address = ?
	`

	err := s.db.QueryRow(query, ipAddress).Scan(&failedAttempts, &lastAttemptAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// IP doesn't exist, create new record
			return s.createIPRecord(ipAddress)
		}
		return fmt.Errorf("failed to get IP attempt count: %w", err)
	}

	// Check if we're within the attempt window
	if time.Since(lastAttemptAt) > s.config.AttemptWindowDuration {
		// Reset attempts if outside the window
		failedAttempts = 0
	}

	// Increment failed attempts
	failedAttempts++

	// Check if we need to apply lockout
	var lockoutUntil *time.Time
	if failedAttempts >= s.config.MaxAttempts {
		lockoutTime := time.Now().Add(s.config.LockoutDuration)
		lockoutUntil = &lockoutTime
	}

	// Update the database
	updateQuery := `
		UPDATE ip_access_attempts 
		SET failed_attempts = ?, 
		    last_attempt_at = ?, 
		    lockout_until = ?
		WHERE ip_address = ?
	`

	_, err = s.db.Exec(updateQuery, failedAttempts, time.Now(), lockoutUntil, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to update IP attempt count: %w", err)
	}

	return nil
}

// ResetFailedAttempts resets the failed attempt count for an IP address after successful access
func (s *IPTrackingService) ResetFailedAttempts(ipAddress string) error {
	query := `
		UPDATE ip_access_attempts 
		SET failed_attempts = 0, 
		    lockout_until = NULL
		WHERE ip_address = ?
	`

	_, err := s.db.Exec(query, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to reset IP attempt count: %w", err)
	}

	return nil
}

// GetIPStatus returns the current status for an IP address
func (s *IPTrackingService) GetIPStatus(ipAddress string) (*IPStatus, error) {
	var status IPStatus

	query := `
		SELECT failed_attempts, last_attempt_at, lockout_until
		FROM ip_access_attempts 
		WHERE ip_address = ?
	`

	err := s.db.QueryRow(query, ipAddress).Scan(
		&status.FailedAttempts,
		&status.LastAttemptAt,
		&status.LockoutUntil,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return &IPStatus{
				IPAddress:      ipAddress,
				FailedAttempts: 0,
				LastAttemptAt:  time.Time{},
				LockoutUntil:   nil,
			}, nil
		}
		return nil, fmt.Errorf("failed to get IP status: %w", err)
	}

	status.IPAddress = ipAddress
	return &status, nil
}

// CleanupOldRecords removes IP records older than the configured cleanup duration
func (s *IPTrackingService) CleanupOldRecords() error {
	cutoffTime := time.Now().Add(-s.config.CleanupOlderThan)

	query := `
		DELETE FROM ip_access_attempts 
		WHERE last_attempt_at < ?
	`

	result, err := s.db.Exec(query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup old IP records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get cleanup result: %w", err)
	}

	if rowsAffected > 0 {
		log.Printf("Cleaned up %d old IP access records", rowsAffected)
	}

	return nil
}

// createIPRecord creates a new IP record with 1 failed attempt
func (s *IPTrackingService) createIPRecord(ipAddress string) error {
	query := `
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, last_attempt_at)
		VALUES (?, 1, ?)
	`

	_, err := s.db.Exec(query, ipAddress, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create IP record: %w", err)
	}

	return nil
}

// clearIPLockout clears an expired lockout for an IP address
func (s *IPTrackingService) clearIPLockout(ipAddress string) error {
	query := `
		UPDATE ip_access_attempts 
		SET lockout_until = NULL
		WHERE ip_address = ?
	`

	_, err := s.db.Exec(query, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to clear IP lockout: %w", err)
	}

	return nil
}

// IPStatus represents the current status of an IP address
type IPStatus struct {
	IPAddress      string     `json:"ip_address"`
	FailedAttempts int        `json:"failed_attempts"`
	LastAttemptAt  time.Time  `json:"last_attempt_at"`
	LockoutUntil   *time.Time `json:"lockout_until"`
}

// IsLockedOut returns true if the IP is currently locked out
func (s *IPStatus) IsLockedOut() bool {
	if s.LockoutUntil == nil {
		return false
	}
	return time.Now().Before(*s.LockoutUntil)
}

// GetRemainingAttempts returns the number of attempts remaining before lockout
func (s *IPStatus) GetRemainingAttempts(maxAttempts int) int {
	if s.FailedAttempts >= maxAttempts {
		return 0
	}
	return maxAttempts - s.FailedAttempts
}

// GetLockoutRemainingTime returns the remaining lockout time, or 0 if not locked out
func (s *IPStatus) GetLockoutRemainingTime() time.Duration {
	if s.LockoutUntil == nil || time.Now().After(*s.LockoutUntil) {
		return 0
	}
	return s.LockoutUntil.Sub(time.Now())
}
