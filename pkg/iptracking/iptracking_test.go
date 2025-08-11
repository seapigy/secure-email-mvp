package iptracking

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create ip_access_attempts table
	_, err = db.Exec(`
		CREATE TABLE ip_access_attempts (
			ip_address TEXT PRIMARY KEY,
			failed_attempts INTEGER DEFAULT 0,
			last_attempt_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			lockout_until TIMESTAMP NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX idx_ip_access_attempts_last_attempt ON ip_access_attempts(last_attempt_at)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX idx_ip_access_attempts_lockout ON ip_access_attempts(lockout_until)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	return db
}

func TestCheckIPLockout_NoLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IP with no lockout
	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts)
		VALUES (?, ?)
	`, "192.168.1.1", 2)
	if err != nil {
		t.Fatalf("Failed to insert test IP: %v", err)
	}

	service := NewIPTrackingService(db)
	locked, err := service.CheckIPLockout("192.168.1.1")
	if err != nil {
		t.Fatalf("CheckIPLockout failed: %v", err)
	}

	if locked {
		t.Error("Expected no lockout, but IP was locked out")
	}
}

func TestCheckIPLockout_ActiveLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IP with active lockout
	lockoutTime := time.Now().Add(30 * time.Minute)
	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, lockout_until)
		VALUES (?, ?, ?)
	`, "192.168.1.2", 5, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert test IP: %v", err)
	}

	service := NewIPTrackingService(db)
	locked, err := service.CheckIPLockout("192.168.1.2")
	if err != nil {
		t.Fatalf("CheckIPLockout failed: %v", err)
	}

	if !locked {
		t.Error("Expected lockout, but IP was not locked out")
	}
}

func TestCheckIPLockout_ExpiredLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IP with expired lockout
	lockoutTime := time.Now().Add(-30 * time.Minute) // Past time
	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, lockout_until)
		VALUES (?, ?, ?)
	`, "192.168.1.3", 5, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert test IP: %v", err)
	}

	service := NewIPTrackingService(db)
	locked, err := service.CheckIPLockout("192.168.1.3")
	if err != nil {
		t.Fatalf("CheckIPLockout failed: %v", err)
	}

	if locked {
		t.Error("Expected no lockout after expiration, but IP was still locked out")
	}

	// Verify lockout was cleared
	var lockoutUntil *time.Time
	err = db.QueryRow("SELECT lockout_until FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.3").Scan(&lockoutUntil)
	if err != nil {
		t.Fatalf("Failed to check lockout status: %v", err)
	}

	if lockoutUntil != nil {
		t.Error("Expected lockout to be cleared, but it was still set")
	}
}

func TestIncrementFailedAttempt_NewIP(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewIPTrackingService(db)

	// Increment failed attempts for new IP
	err := service.IncrementFailedAttempt("192.168.1.4")
	if err != nil {
		t.Fatalf("IncrementFailedAttempt failed: %v", err)
	}

	// Check that record was created with 1 failed attempt
	var failedAttempts int
	err = db.QueryRow("SELECT failed_attempts FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.4").Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to check attempt count: %v", err)
	}

	if failedAttempts != 1 {
		t.Errorf("Expected failed attempts to be 1, got %d", failedAttempts)
	}
}

func TestIncrementFailedAttempt_ExistingIP(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IP with 2 failed attempts
	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, last_attempt_at)
		VALUES (?, ?, ?)
	`, "192.168.1.5", 2, time.Now().Add(-5*time.Minute))
	if err != nil {
		t.Fatalf("Failed to insert test IP: %v", err)
	}

	service := NewIPTrackingService(db)

	// Increment failed attempts
	err = service.IncrementFailedAttempt("192.168.1.5")
	if err != nil {
		t.Fatalf("IncrementFailedAttempt failed: %v", err)
	}

	// Check that attempt count was incremented
	var failedAttempts int
	err = db.QueryRow("SELECT failed_attempts FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.5").Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to check attempt count: %v", err)
	}

	if failedAttempts != 3 {
		t.Errorf("Expected failed attempts to be 3, got %d", failedAttempts)
	}
}

func TestIncrementFailedAttempt_ApplyLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IP with 4 failed attempts (1 away from lockout)
	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, last_attempt_at)
		VALUES (?, ?, ?)
	`, "192.168.1.6", 4, time.Now().Add(-5*time.Minute))
	if err != nil {
		t.Fatalf("Failed to insert test IP: %v", err)
	}

	service := NewIPTrackingService(db)

	// Increment failed attempts (should trigger lockout)
	err = service.IncrementFailedAttempt("192.168.1.6")
	if err != nil {
		t.Fatalf("IncrementFailedAttempt failed: %v", err)
	}

	// Check that lockout was applied
	var lockoutUntil *time.Time
	err = db.QueryRow("SELECT lockout_until FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.6").Scan(&lockoutUntil)
	if err != nil {
		t.Fatalf("Failed to check lockout: %v", err)
	}

	if lockoutUntil == nil {
		t.Error("Expected lockout to be applied, but it was not")
	}

	// Check that lockout time is in the future
	if time.Now().After(*lockoutUntil) {
		t.Error("Expected lockout time to be in the future")
	}

	// Check that attempt count was incremented
	var failedAttempts int
	err = db.QueryRow("SELECT failed_attempts FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.6").Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to check attempt count: %v", err)
	}

	if failedAttempts != 5 {
		t.Errorf("Expected failed attempts to be 5, got %d", failedAttempts)
	}
}

func TestIncrementFailedAttempt_OutsideWindow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IP with 3 failed attempts, but last attempt was 20 minutes ago (outside 15-minute window)
	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, last_attempt_at)
		VALUES (?, ?, ?)
	`, "192.168.1.7", 3, time.Now().Add(-20*time.Minute))
	if err != nil {
		t.Fatalf("Failed to insert test IP: %v", err)
	}

	service := NewIPTrackingService(db)

	// Increment failed attempts (should reset to 1 since outside window)
	err = service.IncrementFailedAttempt("192.168.1.7")
	if err != nil {
		t.Fatalf("IncrementFailedAttempt failed: %v", err)
	}

	// Check that attempt count was reset to 1
	var failedAttempts int
	err = db.QueryRow("SELECT failed_attempts FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.7").Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to check attempt count: %v", err)
	}

	if failedAttempts != 1 {
		t.Errorf("Expected failed attempts to be 1 (reset), got %d", failedAttempts)
	}
}

func TestResetFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IP with failed attempts and lockout
	lockoutTime := time.Now().Add(30 * time.Minute)
	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, lockout_until)
		VALUES (?, ?, ?)
	`, "192.168.1.8", 3, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert test IP: %v", err)
	}

	service := NewIPTrackingService(db)

	// Reset failed attempts
	err = service.ResetFailedAttempts("192.168.1.8")
	if err != nil {
		t.Fatalf("ResetFailedAttempts failed: %v", err)
	}

	// Check that attempt count was reset
	var failedAttempts int
	err = db.QueryRow("SELECT failed_attempts FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.8").Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to check attempt count: %v", err)
	}

	if failedAttempts != 0 {
		t.Errorf("Expected failed attempts to be 0, got %d", failedAttempts)
	}

	// Check that lockout was cleared
	var lockoutUntil *time.Time
	err = db.QueryRow("SELECT lockout_until FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.8").Scan(&lockoutUntil)
	if err != nil {
		t.Fatalf("Failed to check lockout: %v", err)
	}

	if lockoutUntil != nil {
		t.Error("Expected lockout to be cleared, but it was still set")
	}
}

func TestGetIPStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IP with specific settings
	lockoutTime := time.Now().Add(30 * time.Minute)
	lastAttemptTime := time.Now().Add(-5 * time.Minute)
	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, last_attempt_at, lockout_until)
		VALUES (?, ?, ?, ?)
	`, "192.168.1.9", 2, lastAttemptTime, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert test IP: %v", err)
	}

	service := NewIPTrackingService(db)

	// Get status
	status, err := service.GetIPStatus("192.168.1.9")
	if err != nil {
		t.Fatalf("GetIPStatus failed: %v", err)
	}

	// Verify status fields
	if status.IPAddress != "192.168.1.9" {
		t.Errorf("Expected IP address to be 192.168.1.9, got %s", status.IPAddress)
	}

	if status.FailedAttempts != 2 {
		t.Errorf("Expected failed attempts to be 2, got %d", status.FailedAttempts)
	}

	if !status.IsLockedOut() {
		t.Error("Expected IP to be locked out")
	}

	remainingAttempts := status.GetRemainingAttempts(5)
	if remainingAttempts != 3 {
		t.Errorf("Expected remaining attempts to be 3, got %d", remainingAttempts)
	}

	remainingTime := status.GetLockoutRemainingTime()
	if remainingTime <= 0 {
		t.Error("Expected remaining lockout time to be positive")
	}
}

func TestGetIPStatus_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewIPTrackingService(db)

	// Get status for non-existent IP
	status, err := service.GetIPStatus("192.168.1.10")
	if err != nil {
		t.Fatalf("GetIPStatus failed: %v", err)
	}

	// Verify default status
	if status.IPAddress != "192.168.1.10" {
		t.Errorf("Expected IP address to be 192.168.1.10, got %s", status.IPAddress)
	}

	if status.FailedAttempts != 0 {
		t.Errorf("Expected failed attempts to be 0, got %d", status.FailedAttempts)
	}

	if status.IsLockedOut() {
		t.Error("Expected IP to not be locked out")
	}
}

func TestCleanupOldRecords(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test IPs with different timestamps
	oldTime := time.Now().Add(-25 * time.Hour) // Older than 24 hours
	recentTime := time.Now().Add(-1 * time.Hour) // Within 24 hours

	_, err := db.Exec(`
		INSERT INTO ip_access_attempts (ip_address, failed_attempts, last_attempt_at)
		VALUES (?, ?, ?), (?, ?, ?)
	`, "192.168.1.11", 3, oldTime, "192.168.1.12", 2, recentTime)
	if err != nil {
		t.Fatalf("Failed to insert test IPs: %v", err)
	}

	service := NewIPTrackingService(db)

	// Run cleanup
	err = service.CleanupOldRecords()
	if err != nil {
		t.Fatalf("CleanupOldRecords failed: %v", err)
	}

	// Check that old record was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.11").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check old record: %v", err)
	}

	if count != 0 {
		t.Error("Expected old record to be deleted, but it still exists")
	}

	// Check that recent record was kept
	err = db.QueryRow("SELECT COUNT(*) FROM ip_access_attempts WHERE ip_address = ?", "192.168.1.12").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check recent record: %v", err)
	}

	if count != 1 {
		t.Error("Expected recent record to be kept, but it was deleted")
	}
}

func TestIPStatus_NotLockedOut(t *testing.T) {
	status := &IPStatus{
		IPAddress:      "192.168.1.13",
		FailedAttempts: 1,
		LastAttemptAt:  time.Now().Add(-5 * time.Minute),
		LockoutUntil:   nil, // No lockout
	}

	if status.IsLockedOut() {
		t.Error("Expected IP to not be locked out")
	}

	remainingAttempts := status.GetRemainingAttempts(5)
	if remainingAttempts != 4 {
		t.Errorf("Expected remaining attempts to be 4, got %d", remainingAttempts)
	}

	remainingTime := status.GetLockoutRemainingTime()
	if remainingTime != 0 {
		t.Errorf("Expected remaining lockout time to be 0, got %v", remainingTime)
	}
}

func TestIPStatus_AtMaxAttempts(t *testing.T) {
	status := &IPStatus{
		IPAddress:      "192.168.1.14",
		FailedAttempts: 5,
		LastAttemptAt:  time.Now().Add(-5 * time.Minute),
		LockoutUntil:   nil, // No lockout yet
	}

	remainingAttempts := status.GetRemainingAttempts(5)
	if remainingAttempts != 0 {
		t.Errorf("Expected remaining attempts to be 0, got %d", remainingAttempts)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts to be 5, got %d", config.MaxAttempts)
	}

	if config.LockoutDuration != 30*time.Minute {
		t.Errorf("Expected LockoutDuration to be 30 minutes, got %v", config.LockoutDuration)
	}

	if config.CleanupOlderThan != 24*time.Hour {
		t.Errorf("Expected CleanupOlderThan to be 24 hours, got %v", config.CleanupOlderThan)
	}

	if config.AttemptWindowDuration != 15*time.Minute {
		t.Errorf("Expected AttemptWindowDuration to be 15 minutes, got %v", config.AttemptWindowDuration)
	}
}
