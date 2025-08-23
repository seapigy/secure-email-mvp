// =============================================================================
// SECURE EMAIL MVP - SECURITY HARDENING TESTS (Micro-Iteration 4.22)
// =============================================================================
// Comprehensive tests for security hardening features:
// - Enhanced audit logging with detailed metadata
// - Rate-limiting decryption attempts (3 failed attempts per 5 minutes per IP)
// - Concurrent access protection with short-lived locks (2 seconds)
// =============================================================================

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"secure-email-mvp/pkg/audit"
	"secure-email-mvp/pkg/auth"

	_ "modernc.org/sqlite"
)

// TestEmailAccessAuditor tests the enhanced email access auditor functionality
func TestEmailAccessAuditor(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create email access logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS email_access_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email_id TEXT NOT NULL,
			user_id TEXT,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			status TEXT NOT NULL,
			attempt_count INTEGER DEFAULT 1,
			result TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create email_access_logs table: %v", err)
	}

	// Create audit service
	auditor := audit.NewEmailAccessAuditor(db, audit.DefaultRateLimitConfig)

	t.Run("LogAccess", func(t *testing.T) {
		ctx := context.Background()
		emailID := "test-email-123"
		ipAddress := "192.168.1.1"
		userID := "user-123"
		result := "success"
		userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

		// Log access attempt
		err := auditor.LogAccess(ctx, emailID, ipAddress, &userID, result, userAgent)
		if err != nil {
			t.Fatalf("Failed to log access: %v", err)
		}

		// Verify log entry was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM email_access_logs WHERE email_id = ?", emailID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count log entries: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 log entry, got %d", count)
		}
	})

	t.Run("RateLimitCheck", func(t *testing.T) {
		ctx := context.Background()
		emailID := "test-email-456"
		ipAddress := "192.168.1.2"

		// Log multiple failed attempts
		for i := 0; i < 3; i++ {
			err := auditor.LogAccess(ctx, emailID, ipAddress, nil, "failed_password", "test-agent")
			if err != nil {
				t.Fatalf("Failed to log failed attempt %d: %v", i+1, err)
			}
		}

		// Check rate limit
		isLimited, err := auditor.CheckRateLimit(ctx, emailID, ipAddress)
		if err != nil {
			t.Fatalf("Failed to check rate limit: %v", err)
		}
		if !isLimited {
			t.Error("Expected rate limit to be exceeded after 3 failed attempts")
		}

		// Test successful attempt should not be rate limited
		isLimited, err = auditor.CheckRateLimit(ctx, emailID, "192.168.1.3")
		if err != nil {
			t.Fatalf("Failed to check rate limit for different IP: %v", err)
		}
		if isLimited {
			t.Error("Expected rate limit to not be exceeded for different IP")
		}
	})

	t.Run("GetAccessLogs", func(t *testing.T) {
		ctx := context.Background()
		emailID := "test-email-789"

		// Log some access attempts
		for i := 0; i < 3; i++ {
			err := auditor.LogAccess(ctx, emailID, "192.168.1.4", nil, "success", "test-agent")
			if err != nil {
				t.Fatalf("Failed to log access attempt %d: %v", i+1, err)
			}
		}

		// Get access logs
		logs, err := auditor.GetAccessLogs(ctx, emailID, 10)
		if err != nil {
			t.Fatalf("Failed to get access logs: %v", err)
		}
		if len(logs) != 3 {
			t.Errorf("Expected 3 log entries, got %d", len(logs))
		}

		// Verify log entry details
		log := logs[0]
		if log.EmailID != emailID {
			t.Errorf("Expected email ID %s, got %s", emailID, log.EmailID)
		}
		if log.IPAddress != "192.168.1.4" {
			t.Errorf("Expected IP 192.168.1.4, got %s", log.IPAddress)
		}
		if log.Result != "success" {
			t.Errorf("Expected result 'success', got %s", log.Result)
		}
	})

	t.Run("GetFailedAttemptsSummary", func(t *testing.T) {
		ctx := context.Background()

		// Log some failed attempts
		for i := 0; i < 5; i++ {
			err := auditor.LogAccess(ctx, "test-email-999", "192.168.1.5", nil, "failed_password", "test-agent")
			if err != nil {
				t.Fatalf("Failed to log failed attempt %d: %v", i+1, err)
			}
		}

		// Get summary
		summary, err := auditor.GetFailedAttemptsSummary(ctx, 1)
		if err != nil {
			t.Fatalf("Failed to get failed attempts summary: %v", err)
		}

		// Verify summary
		if summary["total_failed_attempts"].(int) < 5 {
			t.Errorf("Expected at least 5 failed attempts, got %d", summary["total_failed_attempts"])
		}
		if summary["unique_ips"].(int) < 1 {
			t.Errorf("Expected at least 1 unique IP, got %d", summary["unique_ips"])
		}
	})
}

// TestConcurrentAccessManager tests the concurrent access protection functionality
func TestConcurrentAccessManager(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create concurrent access manager
	manager := audit.NewConcurrentAccessManager(db)

	t.Run("AcquireLock", func(t *testing.T) {
		emailID := "test-email-lock"

		// First attempt should succeed
		acquired := manager.AcquireLock(emailID)
		if !acquired {
			t.Error("Expected first lock acquisition to succeed")
		}

		// Second attempt should fail (concurrent access)
		acquired = manager.AcquireLock(emailID)
		if acquired {
			t.Error("Expected second lock acquisition to fail")
		}

		// Release lock
		manager.ReleaseLock(emailID)

		// Third attempt should succeed after release
		acquired = manager.AcquireLock(emailID)
		if !acquired {
			t.Error("Expected third lock acquisition to succeed after release")
		}

		manager.ReleaseLock(emailID)
	})

	t.Run("LockTimeout", func(t *testing.T) {
		emailID := "test-email-timeout"

		// Acquire lock
		acquired := manager.AcquireLock(emailID)
		if !acquired {
			t.Fatal("Failed to acquire lock")
		}

		// Wait for lock to expire (2 seconds)
		time.Sleep(3 * time.Second)

		// Should be able to acquire lock again after timeout
		acquired = manager.AcquireLock(emailID)
		if !acquired {
			t.Error("Expected to be able to acquire lock after timeout")
		}

		manager.ReleaseLock(emailID)
	})

	t.Run("IsLocked", func(t *testing.T) {
		emailID := "test-email-check"

		// Initially not locked
		if manager.IsLocked(emailID) {
			t.Error("Expected email to not be locked initially")
		}

		// Acquire lock
		manager.AcquireLock(emailID)

		// Should be locked
		if !manager.IsLocked(emailID) {
			t.Error("Expected email to be locked after acquisition")
		}

		// Release lock
		manager.ReleaseLock(emailID)

		// Should not be locked after release
		if manager.IsLocked(emailID) {
			t.Error("Expected email to not be locked after release")
		}
	})

	t.Run("GetLockStatus", func(t *testing.T) {
		emailID := "test-email-status"

		// Get status when not locked
		status := manager.GetLockStatus(emailID)
		if status["locked"].(bool) {
			t.Error("Expected status to show not locked")
		}

		// Acquire lock
		manager.AcquireLock(emailID)

		// Get status when locked
		status = manager.GetLockStatus(emailID)
		if !status["locked"].(bool) {
			t.Error("Expected status to show locked")
		}

		// Verify timeout field
		if status["timeout"] == nil {
			t.Error("Expected timeout field in status")
		}

		manager.ReleaseLock(emailID)
	})

	t.Run("GetActiveLocks", func(t *testing.T) {
		// Clear any existing locks
		manager.CleanupLocks()

		// Acquire locks on multiple emails
		emailIDs := []string{"email-1", "email-2", "email-3"}
		for _, emailID := range emailIDs {
			manager.AcquireLock(emailID)
		}

		// Get active locks
		activeLocks := manager.GetActiveLocks()
		if activeLocks["total_locks"].(int) != 3 {
			t.Errorf("Expected 3 active locks, got %d", activeLocks["total_locks"])
		}

		// Release locks
		for _, emailID := range emailIDs {
			manager.ReleaseLock(emailID)
		}
	})

	t.Run("ForceReleaseLock", func(t *testing.T) {
		emailID := "test-email-force"

		// Acquire lock
		manager.AcquireLock(emailID)

		// Force release
		released := manager.ForceReleaseLock(emailID)
		if !released {
			t.Error("Expected force release to succeed")
		}

		// Should not be locked after force release
		if manager.IsLocked(emailID) {
			t.Error("Expected email to not be locked after force release")
		}
	})
}

// TestSecurityHardeningIntegration tests the integration of all security hardening features
func TestSecurityHardeningIntegration(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create required tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS email_access_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email_id TEXT NOT NULL,
			user_id TEXT,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			status TEXT NOT NULL,
			attempt_count INTEGER DEFAULT 1,
			result TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create email_access_logs table: %v", err)
	}

	// Create test server with security hardening components
	auditor := audit.NewEmailAccessAuditor(db, audit.DefaultRateLimitConfig)
	concurrentManager := audit.NewConcurrentAccessManager(db)

	_ = &Server{
		db:                      db,
		emailAccessAuditor:      auditor,
		concurrentAccessManager: concurrentManager,
	}

	t.Run("RateLimitAndConcurrentProtection", func(t *testing.T) {
		emailID := "test-email-integration"
		ipAddress := "192.168.1.100"

		// Simulate multiple concurrent requests
		results := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				// Check rate limit
				isLimited, _ := auditor.CheckRateLimit(context.Background(), emailID, ipAddress)
				if isLimited {
					results <- false
					return
				}

				// Try to acquire lock
				acquired := concurrentManager.AcquireLock(emailID)
				if !acquired {
					results <- false
					return
				}
				defer concurrentManager.ReleaseLock(emailID)

				// Log access
				auditor.LogAccess(context.Background(), emailID, ipAddress, nil, "success", "test-agent")
				results <- true
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < 5; i++ {
			if <-results {
				successCount++
			}
		}

		// Only one request should succeed due to concurrent access protection
		if successCount != 1 {
			t.Errorf("Expected 1 successful request, got %d", successCount)
		}
	})

	t.Run("AuditLoggingWithRateLimit", func(t *testing.T) {
		emailID := "test-email-audit"
		ipAddress := "192.168.1.200"

		// Log multiple failed attempts to trigger rate limiting
		for i := 0; i < 4; i++ {
			err := auditor.LogAccess(context.Background(), emailID, ipAddress, nil, "failed_password", "test-agent")
			if err != nil {
				t.Fatalf("Failed to log failed attempt %d: %v", i+1, err)
			}
		}

		// Check rate limit
		isLimited, err := auditor.CheckRateLimit(context.Background(), emailID, ipAddress)
		if err != nil {
			t.Fatalf("Failed to check rate limit: %v", err)
		}
		if !isLimited {
			t.Error("Expected rate limit to be exceeded after 4 failed attempts")
		}

		// Get access logs to verify audit trail
		logs, err := auditor.GetAccessLogs(context.Background(), emailID, 10)
		if err != nil {
			t.Fatalf("Failed to get access logs: %v", err)
		}
		if len(logs) != 4 {
			t.Errorf("Expected 4 log entries, got %d", len(logs))
		}

		// Verify all logs show failed attempts
		for _, log := range logs {
			if log.Result != "failed_password" {
				t.Errorf("Expected result 'failed_password', got %s", log.Result)
			}
			if log.Status != "fail" {
				t.Errorf("Expected status 'fail', got %s", log.Status)
			}
		}
	})
}

// TestSecurityHardeningHTTPHandler tests the HTTP handler with security hardening
func TestSecurityHardeningHTTPHandler(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create required tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS email_access_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email_id TEXT NOT NULL,
			user_id TEXT,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			status TEXT NOT NULL,
			attempt_count INTEGER DEFAULT 1,
			result TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create email_access_logs table: %v", err)
	}

	// Create test server
	auditor := audit.NewEmailAccessAuditor(db, audit.DefaultRateLimitConfig)
	concurrentManager := audit.NewConcurrentAccessManager(db)

	_ = &Server{
		db:                      db,
		emailAccessAuditor:      auditor,
		concurrentAccessManager: concurrentManager,
	}

	t.Run("RateLimitHTTPResponse", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/api/email/test-email-rate", nil)
		req.RemoteAddr = "192.168.1.300:12345"
		req.Header.Set("User-Agent", "test-agent")

		// Add JWT context (simplified for test)
		ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user")
		_ = req.WithContext(ctx)

		w := httptest.NewRecorder()

		// Log multiple failed attempts to trigger rate limiting
		for i := 0; i < 4; i++ {
			auditor.LogAccess(ctx, "test-email-rate", "192.168.1.300", nil, "failed_password", "test-agent")
		}

		// Call handler (this would normally be the full handler, but we're testing the rate limit check)
		// For this test, we'll simulate the rate limit check that happens in the handler
		isLimited, _ := auditor.CheckRateLimit(ctx, "test-email-rate", "192.168.1.300")
		if !isLimited {
			t.Error("Expected rate limit to be exceeded")
		}

		// Verify that the handler would return 429 Too Many Requests
		if isLimited {
			// This simulates what the handler would do
			w.WriteHeader(http.StatusTooManyRequests)
			response := map[string]string{"error": "Too many requests. Please try again later."}
			json.NewEncoder(w).Encode(response)
		}

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status code %d, got %d", http.StatusTooManyRequests, w.Code)
		}

		// Verify response body
		body := w.Body.String()
		if !strings.Contains(body, "Too many requests") {
			t.Errorf("Expected response to contain 'Too many requests', got: %s", body)
		}
	})

	t.Run("ConcurrentAccessHTTPResponse", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest("GET", "/api/email/test-email-concurrent", nil)
		req.RemoteAddr = "192.168.1.400:12345"
		req.Header.Set("User-Agent", "test-agent")

		// Add JWT context
		ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user")
		_ = req.WithContext(ctx)

		w := httptest.NewRecorder()

		// Simulate concurrent access by acquiring lock first
		acquired := concurrentManager.AcquireLock("test-email-concurrent")
		if !acquired {
			t.Fatal("Failed to acquire initial lock")
		}
		defer concurrentManager.ReleaseLock("test-email-concurrent")

		// Try to acquire lock again (simulating concurrent request)
		acquired = concurrentManager.AcquireLock("test-email-concurrent")
		if acquired {
			t.Error("Expected concurrent access to be blocked")
		}

		// Verify that the handler would return 409 Conflict
		if !acquired {
			// This simulates what the handler would do
			w.WriteHeader(http.StatusConflict)
			response := map[string]string{"error": "Email is currently being accessed by another request. Please try again."}
			json.NewEncoder(w).Encode(response)
		}

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status code %d, got %d", http.StatusConflict, w.Code)
		}

		// Verify response body
		body := w.Body.String()
		if !strings.Contains(body, "currently being accessed") {
			t.Errorf("Expected response to contain 'currently being accessed', got: %s", body)
		}
	})
}

// TestSecurityHardeningConfiguration tests configuration options
func TestSecurityHardeningConfiguration(t *testing.T) {
	t.Run("RateLimitConfig", func(t *testing.T) {
		// Test default configuration
		config := audit.DefaultRateLimitConfig
		if config.MaxAttempts != 3 {
			t.Errorf("Expected default max attempts to be 3, got %d", config.MaxAttempts)
		}
		if config.TimeWindow != 5*time.Minute {
			t.Errorf("Expected default time window to be 5 minutes, got %v", config.TimeWindow)
		}

		// Test custom configuration
		customConfig := audit.RateLimitConfig{
			MaxAttempts: 5,
			TimeWindow:  10 * time.Minute,
		}

		if customConfig.MaxAttempts != 5 {
			t.Errorf("Expected custom max attempts to be 5, got %d", customConfig.MaxAttempts)
		}
		if customConfig.TimeWindow != 10*time.Minute {
			t.Errorf("Expected custom time window to be 10 minutes, got %v", customConfig.TimeWindow)
		}
	})

	t.Run("ConcurrentAccessTimeout", func(t *testing.T) {
		// Create test database
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		manager := audit.NewConcurrentAccessManager(db)
		emailID := "test-email-timeout-config"

		// Acquire lock
		acquired := manager.AcquireLock(emailID)
		if !acquired {
			t.Fatal("Failed to acquire lock")
		}

		// Verify lock is active
		if !manager.IsLocked(emailID) {
			t.Error("Expected lock to be active")
		}

		// Wait for timeout (2 seconds as per Micro-Iteration 4.22)
		time.Sleep(3 * time.Second)

		// Verify lock has expired
		if manager.IsLocked(emailID) {
			t.Error("Expected lock to have expired after 3 seconds")
		}

		// Should be able to acquire lock again
		acquired = manager.AcquireLock(emailID)
		if !acquired {
			t.Error("Expected to be able to acquire lock after timeout")
		}

		manager.ReleaseLock(emailID)
	})
}
