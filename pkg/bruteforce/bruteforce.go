package bruteforce

import (
	"database/sql"
	"fmt"
	"time"
)

// BruteForceProtection provides methods for managing brute-force protection
type BruteForceProtection struct {
	db *sql.DB
}

// NewBruteForceProtection creates a new brute-force protection instance
func NewBruteForceProtection(db *sql.DB) *BruteForceProtection {
	return &BruteForceProtection{
		db: db,
	}
}

// CheckLockout checks if an email is currently locked out due to brute-force protection
// Returns true if locked out, false if access is allowed
func (bf *BruteForceProtection) CheckLockout(emailID string) (bool, error) {
	var lockoutUntil *time.Time
	var maxAttempts, failedAttempts int

	query := `
		SELECT brute_force_lockout_until, brute_force_max_attempts, brute_force_failed_attempts
		FROM emails 
		WHERE email_id = ?
	`

	err := bf.db.QueryRow(query, emailID).Scan(&lockoutUntil, &maxAttempts, &failedAttempts)
	if err != nil {
		if err == sql.ErrNoRows {
			// Email not found, allow access (will be handled by caller)
			return false, nil
		}
		return false, fmt.Errorf("failed to check brute-force lockout: %w", err)
	}

	// If lockout is set and not expired, deny access
	if lockoutUntil != nil && time.Now().Before(*lockoutUntil) {
		return true, nil
	}

	// If lockout has expired, clear it
	if lockoutUntil != nil && time.Now().After(*lockoutUntil) {
		err = bf.clearLockout(emailID)
		if err != nil {
			return false, fmt.Errorf("failed to clear expired lockout: %w", err)
		}
	}

	return false, nil
}

// IncrementFailedAttempt increments the failed attempt count and applies lockout if needed
func (bf *BruteForceProtection) IncrementFailedAttempt(emailID string) error {
	// Get current brute-force settings
	var maxAttempts, lockoutDuration int
	var failedAttempts int

	query := `
		SELECT brute_force_max_attempts, brute_force_lockout_duration_minutes, brute_force_failed_attempts
		FROM emails 
		WHERE email_id = ?
	`

	err := bf.db.QueryRow(query, emailID).Scan(&maxAttempts, &lockoutDuration, &failedAttempts)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("email not found: %s", emailID)
		}
		return fmt.Errorf("failed to get brute-force settings: %w", err)
	}

	// Increment failed attempts
	failedAttempts++

	// Check if we need to apply lockout
	var lockoutUntil *time.Time
	if failedAttempts >= maxAttempts {
		lockoutTime := time.Now().Add(time.Duration(lockoutDuration) * time.Minute)
		lockoutUntil = &lockoutTime
	}

	// Update the database
	updateQuery := `
		UPDATE emails 
		SET brute_force_failed_attempts = ?, 
		    brute_force_last_failed_attempt = ?, 
		    brute_force_lockout_until = ?
		WHERE email_id = ?
	`

	_, err = bf.db.Exec(updateQuery, failedAttempts, time.Now(), lockoutUntil, emailID)
	if err != nil {
		return fmt.Errorf("failed to update brute-force attempt count: %w", err)
	}

	return nil
}

// ResetFailedAttempts resets the failed attempt count after successful access
func (bf *BruteForceProtection) ResetFailedAttempts(emailID string) error {
	query := `
		UPDATE emails 
		SET brute_force_failed_attempts = 0, 
		    brute_force_lockout_until = NULL
		WHERE email_id = ?
	`

	_, err := bf.db.Exec(query, emailID)
	if err != nil {
		return fmt.Errorf("failed to reset brute-force attempt count: %w", err)
	}

	return nil
}

// clearLockout clears an expired lockout
func (bf *BruteForceProtection) clearLockout(emailID string) error {
	query := `
		UPDATE emails 
		SET brute_force_lockout_until = NULL
		WHERE email_id = ?
	`

	_, err := bf.db.Exec(query, emailID)
	if err != nil {
		return fmt.Errorf("failed to clear lockout: %w", err)
	}

	return nil
}

// GetBruteForceStatus returns the current brute-force protection status for an email
func (bf *BruteForceProtection) GetBruteForceStatus(emailID string) (*BruteForceStatus, error) {
	var status BruteForceStatus

	query := `
		SELECT brute_force_failed_attempts, 
		       brute_force_last_failed_attempt, 
		       brute_force_lockout_until,
		       brute_force_max_attempts,
		       brute_force_lockout_duration_minutes
		FROM emails 
		WHERE email_id = ?
	`

	err := bf.db.QueryRow(query, emailID).Scan(
		&status.FailedAttempts,
		&status.LastFailedAttempt,
		&status.LockoutUntil,
		&status.MaxAttempts,
		&status.LockoutDurationMinutes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email not found: %s", emailID)
		}
		return nil, fmt.Errorf("failed to get brute-force status: %w", err)
	}

	return &status, nil
}

// BruteForceStatus represents the current brute-force protection status
type BruteForceStatus struct {
	FailedAttempts         int        `json:"failed_attempts"`
	LastFailedAttempt      *time.Time `json:"last_failed_attempt"`
	LockoutUntil           *time.Time `json:"lockout_until"`
	MaxAttempts            int        `json:"max_attempts"`
	LockoutDurationMinutes int        `json:"lockout_duration_minutes"`
}

// IsLockedOut returns true if the email is currently locked out
func (s *BruteForceStatus) IsLockedOut() bool {
	if s.LockoutUntil == nil {
		return false
	}
	return time.Now().Before(*s.LockoutUntil)
}

// GetRemainingAttempts returns the number of attempts remaining before lockout
func (s *BruteForceStatus) GetRemainingAttempts() int {
	if s.FailedAttempts >= s.MaxAttempts {
		return 0
	}
	return s.MaxAttempts - s.FailedAttempts
}

// GetLockoutRemainingTime returns the remaining lockout time, or 0 if not locked out
func (s *BruteForceStatus) GetLockoutRemainingTime() time.Duration {
	if s.LockoutUntil == nil || time.Now().After(*s.LockoutUntil) {
		return 0
	}
	return s.LockoutUntil.Sub(time.Now())
}
