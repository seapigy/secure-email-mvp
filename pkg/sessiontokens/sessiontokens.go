package sessiontokens

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/argon2"
)

// SessionTokenService provides session token functionality
type SessionTokenService interface {
	GenerateSessionToken(emailID string, userAgent, clientIP string) (string, error)
	ValidateSessionToken(emailID, sessionToken string) (bool, error)
	MarkSessionTokenUsed(emailID, sessionToken string) error
	CleanupExpiredSessions() error
	GetSessionInfo(emailID, sessionToken string) (*SessionInfo, error)
	GetActiveSessions(emailID string) ([]*SessionInfo, error)
	RevokeSession(emailID, sessionToken string) error
}

// SessionInfo represents information about a session token
type SessionInfo struct {
	SessionID   string    `json:"session_id"`
	EmailID     string    `json:"email_id"`
	TokenHash   string    `json:"token_hash"`
	ExpiresAt   time.Time `json:"expires_at"`
	Used        bool      `json:"used"`
	CreatedAt   time.Time `json:"created_at"`
	UserAgent   string    `json:"user_agent"`
	IPAddress   string    `json:"ip_address"`
}

// SessionTokenServiceImpl implements SessionTokenService
type SessionTokenServiceImpl struct {
	db *sql.DB
}

// NewSessionTokenService creates a new session token service
func NewSessionTokenService(db *sql.DB) SessionTokenService {
	return &SessionTokenServiceImpl{
		db: db,
	}
}

// GenerateSessionToken creates a new session token for email access
func (s *SessionTokenServiceImpl) GenerateSessionToken(emailID string, userAgent, clientIP string) (string, error) {
	// Generate a high-entropy random token (256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	
	// Convert to hex string for easy transmission
	sessionToken := hex.EncodeToString(tokenBytes)
	
	// Hash the token with Argon2id for secure storage
	tokenHash, err := s.hashToken(sessionToken, emailID)
	if err != nil {
		return "", fmt.Errorf("failed to hash session token: %w", err)
	}
	
	// Generate session ID
	sessionID := s.generateSessionID()
	
	// Set expiration time (5 minutes from now)
	expiresAt := time.Now().Add(5 * time.Minute)
	
	// Store the session in the database
	_, err = s.db.Exec(`
		INSERT INTO email_sessions (
			session_id, email_id, token_hash, expires_at, 
			user_agent, ip_address, created_at
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		sessionID, emailID, tokenHash, expiresAt, userAgent, clientIP)
	
	if err != nil {
		return "", fmt.Errorf("failed to store session token: %w", err)
	}
	
	log.Printf("Session token generated for email %s: %s", emailID, sessionToken[:16]+"...")
	return sessionToken, nil
}

// ValidateSessionToken validates a session token for email access
func (s *SessionTokenServiceImpl) ValidateSessionToken(emailID, sessionToken string) (bool, error) {
	// Hash the provided token
	tokenHash, err := s.hashToken(sessionToken, emailID)
	if err != nil {
		return false, fmt.Errorf("failed to hash session token: %w", err)
	}
	
	// Check if token exists, is not expired, and is not used
	var sessionID string
	var expiresAt time.Time
	var used bool
	
	err = s.db.QueryRow(`
		SELECT session_id, expires_at, used 
		FROM email_sessions 
		WHERE email_id = ? AND token_hash = ?`, emailID, tokenHash).Scan(&sessionID, &expiresAt, &used)
	
	if err == sql.ErrNoRows {
		return false, nil // Token not found
	}
	if err != nil {
		return false, fmt.Errorf("failed to validate session token: %w", err)
	}
	
	// Check if token is expired
	if time.Now().After(expiresAt) {
		return false, nil // Token expired
	}
	
	// Check if token is already used
	if used {
		return false, nil // Token already used
	}
	
	return true, nil
}

// MarkSessionTokenUsed marks a session token as used (for one-time links)
func (s *SessionTokenServiceImpl) MarkSessionTokenUsed(emailID, sessionToken string) error {
	// Hash the provided token
	tokenHash, err := s.hashToken(sessionToken, emailID)
	if err != nil {
		return fmt.Errorf("failed to hash session token: %w", err)
	}
	
	// Mark the token as used
	result, err := s.db.Exec(`
		UPDATE email_sessions 
		SET used = TRUE 
		WHERE email_id = ? AND token_hash = ?`, emailID, tokenHash)
	
	if err != nil {
		return fmt.Errorf("failed to mark session token as used: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("session token not found or already used")
	}
	
	log.Printf("Session token marked as used for email %s", emailID)
	return nil
}

// CleanupExpiredSessions removes expired session tokens from the database
func (s *SessionTokenServiceImpl) CleanupExpiredSessions() error {
	result, err := s.db.Exec(`
		DELETE FROM email_sessions 
		WHERE expires_at < CURRENT_TIMESTAMP`)
	
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d expired session tokens", rowsAffected)
	}
	
	return nil
}

// GetSessionInfo retrieves information about a specific session
func (s *SessionTokenServiceImpl) GetSessionInfo(emailID, sessionToken string) (*SessionInfo, error) {
	// Hash the provided token
	tokenHash, err := s.hashToken(sessionToken, emailID)
	if err != nil {
		return nil, fmt.Errorf("failed to hash session token: %w", err)
	}
	
	var session SessionInfo
	err = s.db.QueryRow(`
		SELECT session_id, email_id, token_hash, expires_at, used, 
		       created_at, user_agent, ip_address
		FROM email_sessions 
		WHERE email_id = ? AND token_hash = ?`, emailID, tokenHash).Scan(
		&session.SessionID, &session.EmailID, &session.TokenHash, &session.ExpiresAt,
		&session.Used, &session.CreatedAt, &session.UserAgent, &session.IPAddress)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session info: %w", err)
	}
	
	return &session, nil
}

// GetActiveSessions retrieves all active (non-expired) sessions for an email
func (s *SessionTokenServiceImpl) GetActiveSessions(emailID string) ([]*SessionInfo, error) {
	rows, err := s.db.Query(`
		SELECT session_id, email_id, token_hash, expires_at, used, 
		       created_at, user_agent, ip_address
		FROM email_sessions 
		WHERE email_id = ? AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC`, emailID)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()
	
	var sessions []*SessionInfo
	for rows.Next() {
		var session SessionInfo
		err := rows.Scan(
			&session.SessionID, &session.EmailID, &session.TokenHash, &session.ExpiresAt,
			&session.Used, &session.CreatedAt, &session.UserAgent, &session.IPAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session info: %w", err)
		}
		
		sessions = append(sessions, &session)
	}
	
	return sessions, nil
}

// RevokeSession revokes a specific session token
func (s *SessionTokenServiceImpl) RevokeSession(emailID, sessionToken string) error {
	// Hash the provided token
	tokenHash, err := s.hashToken(sessionToken, emailID)
	if err != nil {
		return fmt.Errorf("failed to hash session token: %w", err)
	}
	
	result, err := s.db.Exec(`
		DELETE FROM email_sessions 
		WHERE email_id = ? AND token_hash = ?`, emailID, tokenHash)
	
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	
	log.Printf("Session revoked for email %s", emailID)
	return nil
}

// Helper functions

// hashToken creates an Argon2id hash of the session token using emailID as salt
func (s *SessionTokenServiceImpl) hashToken(sessionToken, emailID string) (string, error) {
	// Use Argon2id with emailID as salt for additional security
	salt := []byte(emailID)
	hash := argon2.IDKey([]byte(sessionToken), salt, 1, 64*1024, 4, 32)
	return hex.EncodeToString(hash), nil
}

// generateSessionID generates a unique session ID
func (s *SessionTokenServiceImpl) generateSessionID() string {
	// Generate a simple session ID for now
	// In production, you might want to use a proper UUID library
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
