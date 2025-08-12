package lockout

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// LockoutConfig holds configuration for the lockout service
type LockoutConfig struct {
	MaxAttempts     int           // Maximum failed attempts before lockout
	LockoutDuration time.Duration // Duration of lockout
	AttemptWindow   time.Duration // Time window for counting attempts
	Enabled         bool          // Whether lockout is enabled
}

// DefaultConfig returns default lockout configuration from environment variables
func DefaultConfig() *LockoutConfig {
	config := &LockoutConfig{
		MaxAttempts:     5,                // Default: 5 attempts
		LockoutDuration: 30 * time.Minute, // Default: 30 minutes
		AttemptWindow:   15 * time.Minute, // Default: 15 minutes
		Enabled:         true,             // Default: enabled
	}

	// Load from environment variables
	if maxAttempts := os.Getenv("LOGIN_MAX_ATTEMPTS"); maxAttempts != "" {
		if val, err := strconv.Atoi(maxAttempts); err == nil && val > 0 {
			config.MaxAttempts = val
		}
	}

	if lockoutMinutes := os.Getenv("LOGIN_LOCKOUT_MINUTES"); lockoutMinutes != "" {
		if val, err := strconv.Atoi(lockoutMinutes); err == nil && val > 0 {
			config.LockoutDuration = time.Duration(val) * time.Minute
		}
	}

	if attemptWindow := os.Getenv("LOGIN_ATTEMPT_WINDOW_MINUTES"); attemptWindow != "" {
		if val, err := strconv.Atoi(attemptWindow); err == nil && val > 0 {
			config.AttemptWindow = time.Duration(val) * time.Minute
		}
	}

	if enabled := os.Getenv("LOGIN_RATE_LIMIT_ENABLED"); enabled != "" {
		config.Enabled = enabled == "1" || enabled == "true"
	}

	return config
}

// UserLockoutService provides methods for managing user account lockouts
type UserLockoutService struct {
	db     *sql.DB
	config *LockoutConfig
}

// NewUserLockoutService creates a new user lockout service
func NewUserLockoutService(db *sql.DB) *UserLockoutService {
	return &UserLockoutService{
		db:     db,
		config: DefaultConfig(),
	}
}

// NewUserLockoutServiceWithConfig creates a new user lockout service with custom config
func NewUserLockoutServiceWithConfig(db *sql.DB, config *LockoutConfig) *UserLockoutService {
	return &UserLockoutService{
		db:     db,
		config: config,
	}
}

// CheckUserLockout checks if a user account is currently locked out
// Returns true if locked out, false if access is allowed
func (s *UserLockoutService) CheckUserLockout(email string) (bool, *time.Time, error) {
	if !s.config.Enabled {
		return false, nil, nil
	}

	var lockedUntil *time.Time
	var failedAttempts int
	var lastFailedLogin *time.Time

	query := `
		SELECT account_locked_until, failed_login_attempts, last_failed_login
		FROM users 
		WHERE email = ?
	`

	err := s.db.QueryRow(query, email).Scan(&lockedUntil, &failedAttempts, &lastFailedLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found, allow access (will be handled by caller)
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check user lockout: %w", err)
	}

	// If lockout is set and not expired, deny access
	if lockedUntil != nil && time.Now().Before(*lockedUntil) {
		return true, lockedUntil, nil
	}

	// If lockout has expired, clear it
	if lockedUntil != nil && time.Now().After(*lockedUntil) {
		err = s.clearUserLockout(email)
		if err != nil {
			return false, nil, fmt.Errorf("failed to clear expired user lockout: %w", err)
		}
	}

	// Check if we need to reset attempts outside the window
	if lastFailedLogin != nil && time.Since(*lastFailedLogin) > s.config.AttemptWindow {
		err = s.ResetUserFailedAttempts(email)
		if err != nil {
			log.Printf("Failed to reset user failed attempts: %v", err)
		}
	}

	return false, nil, nil
}

// IncrementUserFailedAttempt increments the failed attempt count for a user
// and applies lockout if the threshold is reached
func (s *UserLockoutService) IncrementUserFailedAttempt(email string) error {
	if !s.config.Enabled {
		return nil
	}

	// Get current failed attempts
	var failedAttempts int
	var lastFailedLogin *time.Time

	query := `
		SELECT failed_login_attempts, last_failed_login
		FROM users 
		WHERE email = ?
	`

	err := s.db.QueryRow(query, email).Scan(&failedAttempts, &lastFailedLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found: %s", email)
		}
		return fmt.Errorf("failed to get user attempt count: %w", err)
	}

	// Check if we're within the attempt window
	if lastFailedLogin != nil && time.Since(*lastFailedLogin) > s.config.AttemptWindow {
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
		log.Printf("User account locked: %s (failed attempts: %d)", email, failedAttempts)
	}

	// Update the database
	updateQuery := `
		UPDATE users 
		SET failed_login_attempts = ?, 
		    last_failed_login = ?, 
		    account_locked_until = ?
		WHERE email = ?
	`

	_, err = s.db.Exec(updateQuery, failedAttempts, time.Now(), lockoutUntil, email)
	if err != nil {
		return fmt.Errorf("failed to update user attempt count: %w", err)
	}

	return nil
}

// ResetUserFailedAttempts resets the failed attempt count for a user after successful login
func (s *UserLockoutService) ResetUserFailedAttempts(email string) error {
	if !s.config.Enabled {
		return nil
	}

	query := `
		UPDATE users 
		SET failed_login_attempts = 0, 
		    account_locked_until = NULL
		WHERE email = ?
	`

	_, err := s.db.Exec(query, email)
	if err != nil {
		return fmt.Errorf("failed to reset user attempt count: %w", err)
	}

	log.Printf("User failed attempts reset: %s", email)
	return nil
}

// clearUserLockout clears an expired lockout for a user
func (s *UserLockoutService) clearUserLockout(email string) error {
	query := `
		UPDATE users 
		SET account_locked_until = NULL
		WHERE email = ?
	`

	_, err := s.db.Exec(query, email)
	if err != nil {
		return fmt.Errorf("failed to clear user lockout: %w", err)
	}

	log.Printf("User lockout cleared: %s", email)
	return nil
}

// GetUserLockoutStatus returns the current lockout status for a user
func (s *UserLockoutService) GetUserLockoutStatus(email string) (*UserLockoutStatus, error) {
	var status UserLockoutStatus

	query := `
		SELECT failed_login_attempts, 
		       last_failed_login, 
		       account_locked_until
		FROM users 
		WHERE email = ?
	`

	err := s.db.QueryRow(query, email).Scan(
		&status.FailedAttempts,
		&status.LastFailedLogin,
		&status.LockoutUntil,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get user lockout status: %w", err)
	}

	status.Email = email
	status.MaxAttempts = s.config.MaxAttempts
	status.LockoutDuration = s.config.LockoutDuration
	status.AttemptWindow = s.config.AttemptWindow

	return &status, nil
}

// UserLockoutStatus represents the current lockout status for a user
type UserLockoutStatus struct {
	Email           string        `json:"email"`
	FailedAttempts  int           `json:"failed_attempts"`
	LastFailedLogin *time.Time    `json:"last_failed_login"`
	LockoutUntil    *time.Time    `json:"lockout_until"`
	MaxAttempts     int           `json:"max_attempts"`
	LockoutDuration time.Duration `json:"lockout_duration"`
	AttemptWindow   time.Duration `json:"attempt_window"`
}

// IsLockedOut returns true if the user is currently locked out
func (s *UserLockoutStatus) IsLockedOut() bool {
	if s.LockoutUntil == nil {
		return false
	}
	return time.Now().Before(*s.LockoutUntil)
}

// GetRemainingAttempts returns the number of attempts remaining before lockout
func (s *UserLockoutStatus) GetRemainingAttempts() int {
	if s.FailedAttempts >= s.MaxAttempts {
		return 0
	}
	return s.MaxAttempts - s.FailedAttempts
}

// GetLockoutRemainingTime returns the remaining lockout time, or 0 if not locked out
func (s *UserLockoutStatus) GetLockoutRemainingTime() time.Duration {
	if s.LockoutUntil == nil || time.Now().After(*s.LockoutUntil) {
		return 0
	}
	return s.LockoutUntil.Sub(time.Now())
}

// IsWithinAttemptWindow returns true if the last failed attempt is within the attempt window
func (s *UserLockoutStatus) IsWithinAttemptWindow() bool {
	if s.LastFailedLogin == nil {
		return false
	}
	return time.Since(*s.LastFailedLogin) <= s.AttemptWindow
}
