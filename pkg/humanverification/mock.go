package humanverification

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockHumanVerificationService is a mock implementation for testing
type MockHumanVerificationService struct {
	mu              sync.RWMutex
	verifications   map[string]bool
	challenges      map[string]*Challenge
	logs            []*VerificationLog
	shouldFail      bool
	verificationType VerificationType
}

// NewMockHumanVerificationService creates a new mock human verification service
func NewMockHumanVerificationService() *MockHumanVerificationService {
	return &MockHumanVerificationService{
		verifications:   make(map[string]bool),
		challenges:      make(map[string]*Challenge),
		logs:            make([]*VerificationLog, 0),
		shouldFail:      false,
		verificationType: VerificationTypeProofOfWork,
	}
}

// SetVerificationResult sets whether verification should succeed or fail
func (m *MockHumanVerificationService) SetVerificationResult(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

// SetVerificationType sets the verification type for the mock
func (m *MockHumanVerificationService) SetVerificationType(verificationType VerificationType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verificationType = verificationType
}

// SetSpecificVerification sets a specific verification result for a token
func (m *MockHumanVerificationService) SetSpecificVerification(token string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verifications[token] = success
}

// SetChallenge sets a specific challenge for testing
func (m *MockHumanVerificationService) SetChallenge(challengeID string, challenge *Challenge) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.challenges[challengeID] = challenge
}

// GetLogs returns all verification logs
func (m *MockHumanVerificationService) GetLogs() []*VerificationLog {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	logs := make([]*VerificationLog, len(m.logs))
	copy(logs, m.logs)
	return logs
}

// ClearLogs clears all verification logs
func (m *MockHumanVerificationService) ClearLogs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = make([]*VerificationLog, 0)
}

// VerifyResponse implements HumanVerificationService.VerifyResponse
func (m *MockHumanVerificationService) VerifyResponse(ctx context.Context, emailID, token string, verificationType VerificationType) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Check if there's a specific result for this token
	if success, exists := m.verifications[token]; exists {
		return success, nil
	}
	
	// Return the global setting
	return !m.shouldFail, nil
}

// GenerateChallenge implements HumanVerificationService.GenerateChallenge
func (m *MockHumanVerificationService) GenerateChallenge(ctx context.Context) (*Challenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	challengeID := uuid.New().String()
	challenge := &Challenge{
		ID:       challengeID,
		Prefix:   "mock_prefix_" + challengeID[:8],
		Target:   "0000",
		MaxNonce: 1000000,
	}
	
	m.challenges[challengeID] = challenge
	return challenge, nil
}

// LogVerification implements HumanVerificationService.LogVerification
func (m *MockHumanVerificationService) LogVerification(ctx context.Context, logEntry *VerificationLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if logEntry.ID == "" {
		logEntry.ID = uuid.New().String()
	}
	
	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now()
	}
	
	m.logs = append(m.logs, logEntry)
	return nil
}

// GetVerificationStats implements HumanVerificationService.GetVerificationStats
func (m *MockHumanVerificationService) GetVerificationStats(ctx context.Context, emailID, ipAddress string, duration time.Duration) (*VerificationStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	since := time.Now().Add(-duration)
	
	var stats VerificationStats
	for _, log := range m.logs {
		if log.Timestamp.Before(since) {
			continue
		}
		
		if log.EmailID != emailID && log.IPAddress != ipAddress {
			continue
		}
		
		stats.TotalAttempts++
		if log.Result == "success" {
			stats.SuccessAttempts++
		} else if log.Result == "failure" {
			stats.FailedAttempts++
		}
	}
	
	if stats.TotalAttempts > 0 {
		stats.FailureRate = float64(stats.FailedAttempts) / float64(stats.TotalAttempts)
	}
	
	return &stats, nil
}
