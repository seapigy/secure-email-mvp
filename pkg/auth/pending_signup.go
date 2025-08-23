package auth

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PendingSignup represents a signup request that is waiting for fallback email confirmation
type PendingSignup struct {
	ID                      string    `json:"id"`
	Email                   string    `json:"email"`
	PasswordHash            string    `json:"-"` // Never expose password hash
	TOTPSecret              string    `json:"-"` // Never expose TOTP secret
	FallbackEmail           string    `json:"fallback_email"`
	FallbackToken           string    `json:"-"`
	FallbackTokenExpiration time.Time `json:"-"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// PendingSignupService handles operations for pending signup requests
type PendingSignupService struct {
	db *sql.DB
}

// NewPendingSignupService creates a new pending signup service
func NewPendingSignupService(db *sql.DB) *PendingSignupService {
	return &PendingSignupService{db: db}
}

// CreatePendingSignup stores a new pending signup request
func (s *PendingSignupService) CreatePendingSignup(
	email, passwordHash, totpSecret, fallbackEmail, fallbackToken string,
	fallbackTokenExpiration time.Time,
) (*PendingSignup, error) {
	id := uuid.New().String()

	query := `
		INSERT INTO pending_signups (
			id, email, password_hash, totp_secret, fallback_email, 
			fallback_token, fallback_token_expiration, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := s.db.Exec(query,
		id, email, passwordHash, totpSecret, fallbackEmail,
		fallbackToken, fallbackTokenExpiration, now, now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create pending signup: %w", err)
	}

	return &PendingSignup{
		ID:                      id,
		Email:                   email,
		PasswordHash:            passwordHash,
		TOTPSecret:              totpSecret,
		FallbackEmail:           fallbackEmail,
		FallbackToken:           fallbackToken,
		FallbackTokenExpiration: fallbackTokenExpiration,
		CreatedAt:               now,
		UpdatedAt:               now,
	}, nil
}

// GetPendingSignupByToken retrieves a pending signup by fallback token
func (s *PendingSignupService) GetPendingSignupByToken(token string) (*PendingSignup, error) {
	query := `
		SELECT id, email, password_hash, totp_secret, fallback_email,
		       fallback_token, fallback_token_expiration, created_at, updated_at
		FROM pending_signups 
		WHERE fallback_token = ?
	`

	var ps PendingSignup
	err := s.db.QueryRow(query, token).Scan(
		&ps.ID, &ps.Email, &ps.PasswordHash, &ps.TOTPSecret, &ps.FallbackEmail,
		&ps.FallbackToken, &ps.FallbackTokenExpiration, &ps.CreatedAt, &ps.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pending signup not found")
		}
		return nil, fmt.Errorf("failed to get pending signup: %w", err)
	}

	return &ps, nil
}

// GetPendingSignupByEmail retrieves a pending signup by email address
func (s *PendingSignupService) GetPendingSignupByEmail(email string) (*PendingSignup, error) {
	query := `
		SELECT id, email, password_hash, totp_secret, fallback_email,
		       fallback_token, fallback_token_expiration, created_at, updated_at
		FROM pending_signups 
		WHERE email = ?
	`

	var ps PendingSignup
	err := s.db.QueryRow(query, email).Scan(
		&ps.ID, &ps.Email, &ps.PasswordHash, &ps.TOTPSecret, &ps.FallbackEmail,
		&ps.FallbackToken, &ps.FallbackTokenExpiration, &ps.CreatedAt, &ps.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pending signup not found")
		}
		return nil, fmt.Errorf("failed to get pending signup: %w", err)
	}

	return &ps, nil
}

// DeletePendingSignup removes a pending signup after successful confirmation
func (s *PendingSignupService) DeletePendingSignup(id string) error {
	query := `DELETE FROM pending_signups WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete pending signup: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pending signup not found")
	}

	return nil
}

// CleanupExpiredSignups removes expired pending signups
func (s *PendingSignupService) CleanupExpiredSignups() (int64, error) {
	query := `DELETE FROM pending_signups WHERE fallback_token_expiration < ?`

	result, err := s.db.Exec(query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired signups: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// IsEmailPending checks if an email address has a pending signup
func (s *PendingSignupService) IsEmailPending(email string) (bool, error) {
	query := `SELECT COUNT(*) FROM pending_signups WHERE email = ?`

	var count int
	err := s.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check pending email: %w", err)
	}

	return count > 0, nil
}
