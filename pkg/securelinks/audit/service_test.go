package audit

import (
	"testing"
	"secure-email-mvp/pkg/models"
)

// MockRepository implements the Repository interface for testing
type MockRepository struct {
	logs   []models.AuditLog
	nextID int
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		logs:   make([]models.AuditLog, 0),
		nextID: 1,
	}
}

func (m *MockRepository) CreateAuditLog(log *models.AuditLog) error {
	log.ID = m.nextID
	m.nextID++
	m.logs = append(m.logs, *log)
	return nil
}

func (m *MockRepository) GetAuditLogs(filters models.AuditLogFilters) ([]models.AuditLog, int, error) {
	var filteredLogs []models.AuditLog
	
	for _, log := range m.logs {
		// Apply filters
		if filters.UserID != "" && log.UserID != filters.UserID {
			continue
		}
		if filters.Action != "" && log.Action != filters.Action {
			continue
		}
		if filters.Entity != "" && log.Entity != filters.Entity {
			continue
		}
		if filters.Severity != "" && log.Severity != filters.Severity {
			continue
		}
		filteredLogs = append(filteredLogs, log)
	}
	
	total := len(filteredLogs)
	
	// Apply pagination
	if filters.Limit > 0 {
		start := filters.Offset
		end := start + filters.Limit
		if end > len(filteredLogs) {
			end = len(filteredLogs)
		}
		if start < len(filteredLogs) {
			filteredLogs = filteredLogs[start:end]
		} else {
			filteredLogs = []models.AuditLog{}
		}
	}
	
	return filteredLogs, total, nil
}

func TestService_LogAction(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	
	// Test logging an action
	err := service.LogAction("test-user", "test-action", "test-entity", "test details", "medium")
	if err != nil {
		t.Errorf("LogAction failed: %v", err)
	}
	
	// Verify log was created
	logs, _, err := service.GetAuditLogs(models.AuditLogFilters{})
	if err != nil {
		t.Errorf("GetAuditLogs failed: %v", err)
	}
	
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
	
	log := logs[0]
	if log.UserID != "test-user" {
		t.Errorf("Expected user ID 'test-user', got '%s'", log.UserID)
	}
	if log.Action != "test-action" {
		t.Errorf("Expected action 'test-action', got '%s'", log.Action)
	}
	if log.Entity != "test-entity" {
		t.Errorf("Expected entity 'test-entity', got '%s'", log.Entity)
	}
	if log.Details != "test details" {
		t.Errorf("Expected details 'test details', got '%s'", log.Details)
	}
	if log.Severity != "medium" {
		t.Errorf("Expected severity 'medium', got '%s'", log.Severity)
	}
}

func TestService_GetAuditLogs(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	
	// Create some test logs
	testLogs := []struct {
		userID   string
		action   string
		entity   string
		details  string
		severity string
	}{
		{"user1", "login", "admin", "Admin login", "low"},
		{"user2", "dlp_scan", "secure_link", "Detected sensitive data", "high"},
		{"user1", "update_policy", "system_security_policy", "Policy updated", "high"},
		{"user3", "create_secure_link", "secure_link", "Created new link", "medium"},
	}
	
	for _, testLog := range testLogs {
		err := service.LogAction(testLog.userID, testLog.action, testLog.entity, testLog.details, testLog.severity)
		if err != nil {
			t.Errorf("Failed to create test log: %v", err)
		}
	}
	
	// Test getting all logs
	logs, total, err := service.GetAuditLogs(models.AuditLogFilters{})
	if err != nil {
		t.Errorf("GetAuditLogs failed: %v", err)
	}
	
	if total != 4 {
		t.Errorf("Expected total 4 logs, got %d", total)
	}
	
	if len(logs) != 4 {
		t.Errorf("Expected 4 logs, got %d", len(logs))
	}
}

func TestService_GetAuditLogsWithFilters(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	
	// Create test logs
	testLogs := []struct {
		userID   string
		action   string
		entity   string
		details  string
		severity string
	}{
		{"user1", "login", "admin", "Admin login", "low"},
		{"user2", "dlp_scan", "secure_link", "Detected sensitive data", "high"},
		{"user1", "update_policy", "system_security_policy", "Policy updated", "high"},
		{"user3", "create_secure_link", "secure_link", "Created new link", "medium"},
	}
	
	for _, testLog := range testLogs {
		err := service.LogAction(testLog.userID, testLog.action, testLog.entity, testLog.details, testLog.severity)
		if err != nil {
			t.Errorf("Failed to create test log: %v", err)
		}
	}
	
	// Test filtering by user ID
	filters := models.AuditLogFilters{UserID: "user1"}
	logs, total, err := service.GetAuditLogs(filters)
	if err != nil {
		t.Errorf("GetAuditLogs with user filter failed: %v", err)
	}
	
	if total != 2 {
		t.Errorf("Expected 2 logs for user1, got %d", total)
	}
	
	for _, log := range logs {
		if log.UserID != "user1" {
			t.Errorf("Expected user ID 'user1', got '%s'", log.UserID)
		}
	}
	
	// Test filtering by severity
	filters = models.AuditLogFilters{Severity: "high"}
	logs, total, err = service.GetAuditLogs(filters)
	if err != nil {
		t.Errorf("GetAuditLogs with severity filter failed: %v", err)
	}
	
	if total != 2 {
		t.Errorf("Expected 2 high severity logs, got %d", total)
	}
	
	for _, log := range logs {
		if log.Severity != "high" {
			t.Errorf("Expected severity 'high', got '%s'", log.Severity)
		}
	}
}

func TestService_GetAuditLogsWithPagination(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	
	// Create 5 test logs
	for i := 1; i <= 5; i++ {
		err := service.LogAction(
			"user1",
			"test-action",
			"test-entity",
			"test details",
			"medium",
		)
		if err != nil {
			t.Errorf("Failed to create test log: %v", err)
		}
	}
	
	// Test pagination with limit 2
	filters := models.AuditLogFilters{Limit: 2, Offset: 0}
	logs, total, err := service.GetAuditLogs(filters)
	if err != nil {
		t.Errorf("GetAuditLogs with pagination failed: %v", err)
	}
	
	if total != 5 {
		t.Errorf("Expected total 5 logs, got %d", total)
	}
	
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs with limit 2, got %d", len(logs))
	}
	
	// Test pagination with offset
	filters = models.AuditLogFilters{Limit: 2, Offset: 2}
	logs, total, err = service.GetAuditLogs(filters)
	if err != nil {
		t.Errorf("GetAuditLogs with offset failed: %v", err)
	}
	
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs with offset 2, got %d", len(logs))
	}
}

func TestService_LogAdminLogin(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	
	err := service.LogAdminLogin("admin@test.com")
	if err != nil {
		t.Errorf("LogAdminLogin failed: %v", err)
	}
	
	logs, _, err := service.GetAuditLogs(models.AuditLogFilters{})
	if err != nil {
		t.Errorf("GetAuditLogs failed: %v", err)
	}
	
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
	
	log := logs[0]
	if log.UserID != "admin@test.com" {
		t.Errorf("Expected user ID 'admin@test.com', got '%s'", log.UserID)
	}
	if log.Action != "login" {
		t.Errorf("Expected action 'login', got '%s'", log.Action)
	}
	if log.Entity != "admin" {
		t.Errorf("Expected entity 'admin', got '%s'", log.Entity)
	}
	if log.Severity != "low" {
		t.Errorf("Expected severity 'low', got '%s'", log.Severity)
	}
}

func TestService_LogPolicyUpdate(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)
	
	err := service.LogPolicyUpdate("admin@test.com", "password_policy", "Updated password requirements")
	if err != nil {
		t.Errorf("LogPolicyUpdate failed: %v", err)
	}
	
	logs, _, err := service.GetAuditLogs(models.AuditLogFilters{})
	if err != nil {
		t.Errorf("GetAuditLogs failed: %v", err)
	}
	
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
	
	log := logs[0]
	if log.UserID != "admin@test.com" {
		t.Errorf("Expected user ID 'admin@test.com', got '%s'", log.UserID)
	}
	if log.Action != "update_policy" {
		t.Errorf("Expected action 'update_policy', got '%s'", log.Action)
	}
	if log.Entity != "system_security_policy" {
		t.Errorf("Expected entity 'system_security_policy', got '%s'", log.Entity)
	}
	if log.Details != "Updated password requirements" {
		t.Errorf("Expected details 'Updated password requirements', got '%s'", log.Details)
	}
	if log.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", log.Severity)
	}
}
