package sessiontokens

import (
	"testing"
	"time"
)

func TestMockSessionTokenService_GenerateSessionToken(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	tests := []struct {
		name      string
		emailID   string
		userAgent string
		clientIP  string
		wantErr   bool
	}{
		{
			name:      "Valid session token generation",
			emailID:   "test-email-123",
			userAgent: "Mozilla/5.0",
			clientIP:  "192.168.1.1",
			wantErr:   false,
		},
		{
			name:      "Empty email ID",
			emailID:   "",
			userAgent: "Mozilla/5.0",
			clientIP:  "192.168.1.1",
			wantErr:   false, // Should still work
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := mock.GenerateSessionToken(tt.emailID, tt.userAgent, tt.clientIP)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSessionToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && token == "" {
				t.Error("GenerateSessionToken() returned empty token")
			}
			
			// Verify token is stored
			valid, err := mock.ValidateSessionToken(tt.emailID, token)
			if err != nil {
				t.Errorf("ValidateSessionToken() error = %v", err)
			}
			if !valid {
				t.Error("Generated token is not valid")
			}
		})
	}
}

func TestMockSessionTokenService_ValidateSessionToken(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	// Generate a valid token
	emailID := "test-email-123"
	token, _ := mock.GenerateSessionToken(emailID, "Mozilla/5.0", "192.168.1.1")
	
	tests := []struct {
		name    string
		emailID string
		token   string
		want    bool
	}{
		{
			name:    "Valid token",
			emailID: emailID,
			token:   token,
			want:    true,
		},
		{
			name:    "Invalid token",
			emailID: emailID,
			token:   "invalid-token",
			want:    false,
		},
		{
			name:    "Wrong email ID",
			emailID: "wrong-email",
			token:   token,
			want:    false,
		},
		{
			name:    "Empty token",
			emailID: emailID,
			token:   "",
			want:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mock.ValidateSessionToken(tt.emailID, tt.token)
			if err != nil {
				t.Errorf("ValidateSessionToken() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateSessionToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockSessionTokenService_MarkSessionTokenUsed(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	// Generate a valid token
	emailID := "test-email-123"
	token, _ := mock.GenerateSessionToken(emailID, "Mozilla/5.0", "192.168.1.1")
	
	tests := []struct {
		name    string
		emailID string
		token   string
		wantErr bool
	}{
		{
			name:    "Mark valid token as used",
			emailID: emailID,
			token:   token,
			wantErr: false,
		},
		{
			name:    "Mark invalid token as used",
			emailID: emailID,
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "Mark already used token",
			emailID: emailID,
			token:   token, // This will be used after first call
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mock.MarkSessionTokenUsed(tt.emailID, tt.token)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkSessionTokenUsed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// If successfully marked as used, verify it's no longer valid
			if !tt.wantErr {
				valid, _ := mock.ValidateSessionToken(tt.emailID, tt.token)
				if valid {
					t.Error("Token should not be valid after being marked as used")
				}
			}
		})
	}
}

func TestMockSessionTokenService_ExpiredToken(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	// Create an expired session manually
	emailID := "test-email-123"
	token := "expired-token"
	tokenHash := mock.GetTokenHash(token, emailID)
	
	expiredSession := &SessionInfo{
		SessionID:   "expired-session",
		EmailID:     emailID,
		TokenHash:   tokenHash,
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		Used:        false,
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		UserAgent:   "Mozilla/5.0",
		IPAddress:   "192.168.1.1",
	}
	
	mock.SetSession(emailID, token, expiredSession)
	
	// Verify expired token is not valid
	valid, err := mock.ValidateSessionToken(emailID, token)
	if err != nil {
		t.Errorf("ValidateSessionToken() error = %v", err)
	}
	if valid {
		t.Error("Expired token should not be valid")
	}
}

func TestMockSessionTokenService_CleanupExpiredSessions(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	// Create both valid and expired sessions
	emailID := "test-email-123"
	
	// Valid session
	validToken, _ := mock.GenerateSessionToken(emailID, "Mozilla/5.0", "192.168.1.1")
	
	// Expired session
	expiredToken := "expired-token"
	tokenHash := mock.GetTokenHash(expiredToken, emailID)
	expiredSession := &SessionInfo{
		SessionID:   "expired-session",
		EmailID:     emailID,
		TokenHash:   tokenHash,
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		Used:        false,
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		UserAgent:   "Mozilla/5.0",
		IPAddress:   "192.168.1.1",
	}
	mock.SetSession(emailID, expiredToken, expiredSession)
	
	// Verify both sessions exist before cleanup
	validBefore, _ := mock.ValidateSessionToken(emailID, validToken)
	expiredBefore, _ := mock.ValidateSessionToken(emailID, expiredToken)
	
	if !validBefore {
		t.Error("Valid token should be valid before cleanup")
	}
	if expiredBefore {
		t.Error("Expired token should not be valid before cleanup")
	}
	
	// Run cleanup
	err := mock.CleanupExpiredSessions()
	if err != nil {
		t.Errorf("CleanupExpiredSessions() error = %v", err)
	}
	
	// Verify valid session still exists, expired session is removed
	validAfter, _ := mock.ValidateSessionToken(emailID, validToken)
	expiredAfter, _ := mock.ValidateSessionToken(emailID, expiredToken)
	
	if !validAfter {
		t.Error("Valid token should still be valid after cleanup")
	}
	if expiredAfter {
		t.Error("Expired token should not be valid after cleanup")
	}
}

func TestMockSessionTokenService_GetSessionInfo(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	// Generate a valid token
	emailID := "test-email-123"
	token, _ := mock.GenerateSessionToken(emailID, "Mozilla/5.0", "192.168.1.1")
	
	tests := []struct {
		name    string
		emailID string
		token   string
		wantErr bool
	}{
		{
			name:    "Get valid session info",
			emailID: emailID,
			token:   token,
			wantErr: false,
		},
		{
			name:    "Get invalid session info",
			emailID: emailID,
			token:   "invalid-token",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionInfo, err := mock.GetSessionInfo(tt.emailID, tt.token)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSessionInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && sessionInfo == nil {
				t.Error("GetSessionInfo() returned nil session info")
			}
			
			if !tt.wantErr && sessionInfo.EmailID != tt.emailID {
				t.Errorf("GetSessionInfo() returned wrong email ID: %s, want %s", sessionInfo.EmailID, tt.emailID)
			}
		})
	}
}

func TestMockSessionTokenService_GetActiveSessions(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	// Generate multiple valid tokens for the same email
	emailID := "test-email-123"
	token1, _ := mock.GenerateSessionToken(emailID, "Mozilla/5.0", "192.168.1.1")
	// Add a small delay to ensure different timestamps
	time.Sleep(1 * time.Millisecond)
	token2, _ := mock.GenerateSessionToken(emailID, "Chrome/91.0", "192.168.1.2")
	
	// Verify tokens were generated
	if token1 == "" || token2 == "" {
		t.Error("Generated tokens should not be empty")
	}
	
	// Create an expired session
	expiredToken := "expired-token"
	tokenHash := mock.GetTokenHash(expiredToken, emailID)
	expiredSession := &SessionInfo{
		SessionID:   "expired-session",
		EmailID:     emailID,
		TokenHash:   tokenHash,
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		Used:        false,
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		UserAgent:   "Safari/14.0",
		IPAddress:   "192.168.1.3",
	}
	mock.SetSession(emailID, expiredToken, expiredSession)
	
	// Get active sessions
	activeSessions, err := mock.GetActiveSessions(emailID)
	if err != nil {
		t.Errorf("GetActiveSessions() error = %v", err)
	}
	
	// Should have 2 active sessions (token1 and token2)
	if len(activeSessions) != 2 {
		t.Errorf("GetActiveSessions() returned %d sessions, want 2", len(activeSessions))
	}
	
	// Verify the expired session is not included
	for _, session := range activeSessions {
		if session.TokenHash == tokenHash {
			t.Error("Expired session should not be in active sessions")
		}
	}
}

func TestMockSessionTokenService_RevokeSession(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	// Generate a valid token
	emailID := "test-email-123"
	token, _ := mock.GenerateSessionToken(emailID, "Mozilla/5.0", "192.168.1.1")
	
	tests := []struct {
		name    string
		emailID string
		token   string
		wantErr bool
	}{
		{
			name:    "Revoke valid session",
			emailID: emailID,
			token:   token,
			wantErr: false,
		},
		{
			name:    "Revoke invalid session",
			emailID: emailID,
			token:   "invalid-token",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mock.RevokeSession(tt.emailID, tt.token)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("RevokeSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// If successfully revoked, verify it's no longer valid
			if !tt.wantErr {
				valid, _ := mock.ValidateSessionToken(tt.emailID, tt.token)
				if valid {
					t.Error("Token should not be valid after being revoked")
				}
			}
		})
	}
}

func TestMockSessionTokenService_ClearSessions(t *testing.T) {
	mock := NewMockSessionTokenService()
	
	// Generate multiple tokens
	emailID1 := "test-email-1"
	emailID2 := "test-email-2"
	token1, _ := mock.GenerateSessionToken(emailID1, "Mozilla/5.0", "192.168.1.1")
	token2, _ := mock.GenerateSessionToken(emailID2, "Chrome/91.0", "192.168.1.2")
	
	// Verify tokens were generated
	if token1 == "" || token2 == "" {
		t.Error("Generated tokens should not be empty")
	}
	
	// Verify tokens exist
	valid1, _ := mock.ValidateSessionToken(emailID1, token1)
	valid2, _ := mock.ValidateSessionToken(emailID2, token2)
	
	if !valid1 || !valid2 {
		t.Error("Tokens should be valid before clearing")
	}
	
	// Clear all sessions
	mock.ClearSessions()
	
	// Verify tokens no longer exist
	valid1After, _ := mock.ValidateSessionToken(emailID1, token1)
	valid2After, _ := mock.ValidateSessionToken(emailID2, token2)
	
	if valid1After || valid2After {
		t.Error("Tokens should not be valid after clearing")
	}
}
