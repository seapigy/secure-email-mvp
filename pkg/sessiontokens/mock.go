package sessiontokens

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// MockSessionTokenService is a mock implementation for testing
type MockSessionTokenService struct {
	sessions map[string]map[string]*SessionInfo // emailID -> tokenHash -> SessionInfo
	tokens   map[string]string                  // tokenHash -> sessionToken
}

// NewMockSessionTokenService creates a new mock session token service
func NewMockSessionTokenService() *MockSessionTokenService {
	return &MockSessionTokenService{
		sessions: make(map[string]map[string]*SessionInfo),
		tokens:   make(map[string]string),
	}
}

// GenerateSessionToken creates a mock session token
func (m *MockSessionTokenService) GenerateSessionToken(emailID string, userAgent, clientIP string) (string, error) {
	// Generate a simple token for testing
	sessionToken := fmt.Sprintf("mock-token-%s-%d", emailID, time.Now().UnixNano())
	
	// Hash the token
	tokenHash, err := m.hashToken(sessionToken, emailID)
	if err != nil {
		return "", err
	}
	
	// Store the token mapping
	m.tokens[tokenHash] = sessionToken
	
	// Create session info
	sessionInfo := &SessionInfo{
		SessionID:   fmt.Sprintf("mock-session-%s", tokenHash[:8]),
		EmailID:     emailID,
		TokenHash:   tokenHash,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		Used:        false,
		CreatedAt:   time.Now(),
		UserAgent:   userAgent,
		IPAddress:   clientIP,
	}
	
	// Store the session
	if m.sessions[emailID] == nil {
		m.sessions[emailID] = make(map[string]*SessionInfo)
	}
	m.sessions[emailID][tokenHash] = sessionInfo
	
	return sessionToken, nil
}

// ValidateSessionToken validates a mock session token
func (m *MockSessionTokenService) ValidateSessionToken(emailID, sessionToken string) (bool, error) {
	// Hash the provided token
	tokenHash, err := m.hashToken(sessionToken, emailID)
	if err != nil {
		return false, err
	}
	
	// Check if session exists
	if emailSessions, exists := m.sessions[emailID]; exists {
		if session, found := emailSessions[tokenHash]; found {
			// Check if token is expired
			if time.Now().After(session.ExpiresAt) {
				return false, nil // Token expired
			}
			
			// Check if token is already used
			if session.Used {
				return false, nil // Token already used
			}
			
			return true, nil // Token is valid
		}
	}
	
	return false, nil // Token not found
}

// MarkSessionTokenUsed marks a mock session token as used
func (m *MockSessionTokenService) MarkSessionTokenUsed(emailID, sessionToken string) error {
	// Hash the provided token
	tokenHash, err := m.hashToken(sessionToken, emailID)
	if err != nil {
		return err
	}
	
	// Mark the token as used
	if emailSessions, exists := m.sessions[emailID]; exists {
		if session, found := emailSessions[tokenHash]; found {
			if session.Used {
				return fmt.Errorf("session token not found or already used")
			}
			session.Used = true
			return nil
		}
	}
	
	return fmt.Errorf("session token not found or already used")
}

// CleanupExpiredSessions removes expired mock session tokens
func (m *MockSessionTokenService) CleanupExpiredSessions() error {
	now := time.Now()
	
	for emailID, emailSessions := range m.sessions {
		for tokenHash, session := range emailSessions {
			if now.After(session.ExpiresAt) {
				delete(emailSessions, tokenHash)
				delete(m.tokens, tokenHash)
			}
		}
		
		// Remove empty email sessions
		if len(emailSessions) == 0 {
			delete(m.sessions, emailID)
		}
	}
	
	return nil
}

// GetSessionInfo retrieves information about a specific mock session
func (m *MockSessionTokenService) GetSessionInfo(emailID, sessionToken string) (*SessionInfo, error) {
	// Hash the provided token
	tokenHash, err := m.hashToken(sessionToken, emailID)
	if err != nil {
		return nil, err
	}
	
	// Get session info
	if emailSessions, exists := m.sessions[emailID]; exists {
		if session, found := emailSessions[tokenHash]; found {
			return session, nil
		}
	}
	
	return nil, fmt.Errorf("session not found")
}

// GetActiveSessions retrieves all active mock sessions for an email
func (m *MockSessionTokenService) GetActiveSessions(emailID string) ([]*SessionInfo, error) {
	now := time.Now()
	var activeSessions []*SessionInfo
	
	if emailSessions, exists := m.sessions[emailID]; exists {
		for _, session := range emailSessions {
			if now.Before(session.ExpiresAt) {
				activeSessions = append(activeSessions, session)
			}
		}
	}
	
	return activeSessions, nil
}

// RevokeSession revokes a specific mock session token
func (m *MockSessionTokenService) RevokeSession(emailID, sessionToken string) error {
	// Hash the provided token
	tokenHash, err := m.hashToken(sessionToken, emailID)
	if err != nil {
		return err
	}
	
	// Remove the session
	if emailSessions, exists := m.sessions[emailID]; exists {
		if _, found := emailSessions[tokenHash]; found {
			delete(emailSessions, tokenHash)
			delete(m.tokens, tokenHash)
			
			// Remove empty email sessions
			if len(emailSessions) == 0 {
				delete(m.sessions, emailID)
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("session not found")
}

// Helper functions

// hashToken creates a mock hash of the session token
func (m *MockSessionTokenService) hashToken(sessionToken, emailID string) (string, error) {
	// Simple hash for testing - combine token and emailID
	combined := sessionToken + "|" + emailID
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:]), nil
}

// SetSession allows setting a session for testing
func (m *MockSessionTokenService) SetSession(emailID, sessionToken string, sessionInfo *SessionInfo) {
	tokenHash, _ := m.hashToken(sessionToken, emailID)
	
	if m.sessions[emailID] == nil {
		m.sessions[emailID] = make(map[string]*SessionInfo)
	}
	m.sessions[emailID][tokenHash] = sessionInfo
	m.tokens[tokenHash] = sessionToken
}

// ClearSessions clears all sessions for testing
func (m *MockSessionTokenService) ClearSessions() {
	m.sessions = make(map[string]map[string]*SessionInfo)
	m.tokens = make(map[string]string)
}

// GetTokenHash returns the hash for a given token (for testing)
func (m *MockSessionTokenService) GetTokenHash(sessionToken, emailID string) string {
	hash, _ := m.hashToken(sessionToken, emailID)
	return hash
}
