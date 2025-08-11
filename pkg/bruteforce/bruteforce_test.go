package bruteforce

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

	// Create emails table with brute-force fields
	_, err = db.Exec(`
		CREATE TABLE emails (
			email_id TEXT PRIMARY KEY,
			brute_force_failed_attempts INTEGER DEFAULT 0,
			brute_force_last_failed_attempt DATETIME,
			brute_force_lockout_until DATETIME,
			brute_force_max_attempts INTEGER DEFAULT 3,
			brute_force_lockout_duration_minutes INTEGER DEFAULT 15
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

func TestCheckLockout_NoLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email with no lockout
	_, err := db.Exec(`
		INSERT INTO emails (email_id, brute_force_failed_attempts, brute_force_max_attempts)
		VALUES (?, ?, ?)
	`, "test-email-1", 0, 3)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	bf := NewBruteForceProtection(db)
	locked, err := bf.CheckLockout("test-email-1")
	if err != nil {
		t.Fatalf("CheckLockout failed: %v", err)
	}

	if locked {
		t.Error("Expected no lockout, but email was locked out")
	}
}

func TestCheckLockout_ActiveLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email with active lockout
	lockoutTime := time.Now().Add(30 * time.Minute)
	_, err := db.Exec(`
		INSERT INTO emails (email_id, brute_force_failed_attempts, brute_force_max_attempts, brute_force_lockout_until)
		VALUES (?, ?, ?, ?)
	`, "test-email-2", 3, 3, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	bf := NewBruteForceProtection(db)
	locked, err := bf.CheckLockout("test-email-2")
	if err != nil {
		t.Fatalf("CheckLockout failed: %v", err)
	}

	if !locked {
		t.Error("Expected lockout, but email was not locked out")
	}
}

func TestCheckLockout_ExpiredLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email with expired lockout
	lockoutTime := time.Now().Add(-30 * time.Minute) // Past time
	_, err := db.Exec(`
		INSERT INTO emails (email_id, brute_force_failed_attempts, brute_force_max_attempts, brute_force_lockout_until)
		VALUES (?, ?, ?, ?)
	`, "test-email-3", 3, 3, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	bf := NewBruteForceProtection(db)
	locked, err := bf.CheckLockout("test-email-3")
	if err != nil {
		t.Fatalf("CheckLockout failed: %v", err)
	}

	if locked {
		t.Error("Expected no lockout after expiration, but email was still locked out")
	}

	// Verify lockout was cleared
	var lockoutUntil *time.Time
	err = db.QueryRow("SELECT brute_force_lockout_until FROM emails WHERE email_id = ?", "test-email-3").Scan(&lockoutUntil)
	if err != nil {
		t.Fatalf("Failed to check lockout status: %v", err)
	}

	if lockoutUntil != nil {
		t.Error("Expected lockout to be cleared, but it was still set")
	}
}

func TestIncrementFailedAttempt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`
		INSERT INTO emails (email_id, brute_force_failed_attempts, brute_force_max_attempts, brute_force_lockout_duration_minutes)
		VALUES (?, ?, ?, ?)
	`, "test-email-4", 0, 3, 15)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	bf := NewBruteForceProtection(db)

	// Increment failed attempts
	err = bf.IncrementFailedAttempt("test-email-4")
	if err != nil {
		t.Fatalf("IncrementFailedAttempt failed: %v", err)
	}

	// Check that attempt count was incremented
	var failedAttempts int
	err = db.QueryRow("SELECT brute_force_failed_attempts FROM emails WHERE email_id = ?", "test-email-4").Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to check attempt count: %v", err)
	}

	if failedAttempts != 1 {
		t.Errorf("Expected failed attempts to be 1, got %d", failedAttempts)
	}
}

func TestIncrementFailedAttempt_ApplyLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email with 2 failed attempts (1 away from lockout)
	_, err := db.Exec(`
		INSERT INTO emails (email_id, brute_force_failed_attempts, brute_force_max_attempts, brute_force_lockout_duration_minutes)
		VALUES (?, ?, ?, ?)
	`, "test-email-5", 2, 3, 15)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	bf := NewBruteForceProtection(db)

	// Increment failed attempts (should trigger lockout)
	err = bf.IncrementFailedAttempt("test-email-5")
	if err != nil {
		t.Fatalf("IncrementFailedAttempt failed: %v", err)
	}

	// Check that lockout was applied
	var lockoutUntil *time.Time
	err = db.QueryRow("SELECT brute_force_lockout_until FROM emails WHERE email_id = ?", "test-email-5").Scan(&lockoutUntil)
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
	err = db.QueryRow("SELECT brute_force_failed_attempts FROM emails WHERE email_id = ?", "test-email-5").Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to check attempt count: %v", err)
	}

	if failedAttempts != 3 {
		t.Errorf("Expected failed attempts to be 3, got %d", failedAttempts)
	}
}

func TestResetFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email with failed attempts and lockout
	lockoutTime := time.Now().Add(30 * time.Minute)
	_, err := db.Exec(`
		INSERT INTO emails (email_id, brute_force_failed_attempts, brute_force_max_attempts, brute_force_lockout_until)
		VALUES (?, ?, ?, ?)
	`, "test-email-6", 3, 3, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	bf := NewBruteForceProtection(db)

	// Reset failed attempts
	err = bf.ResetFailedAttempts("test-email-6")
	if err != nil {
		t.Fatalf("ResetFailedAttempts failed: %v", err)
	}

	// Check that attempt count was reset
	var failedAttempts int
	err = db.QueryRow("SELECT brute_force_failed_attempts FROM emails WHERE email_id = ?", "test-email-6").Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to check attempt count: %v", err)
	}

	if failedAttempts != 0 {
		t.Errorf("Expected failed attempts to be 0, got %d", failedAttempts)
	}

	// Check that lockout was cleared
	var lockoutUntil *time.Time
	err = db.QueryRow("SELECT brute_force_lockout_until FROM emails WHERE email_id = ?", "test-email-6").Scan(&lockoutUntil)
	if err != nil {
		t.Fatalf("Failed to check lockout: %v", err)
	}

	if lockoutUntil != nil {
		t.Error("Expected lockout to be cleared, but it was still set")
	}
}

func TestGetBruteForceStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email with specific settings
	lockoutTime := time.Now().Add(30 * time.Minute)
	lastFailedTime := time.Now().Add(-5 * time.Minute)
	_, err := db.Exec(`
		INSERT INTO emails (email_id, brute_force_failed_attempts, brute_force_last_failed_attempt, 
		                   brute_force_lockout_until, brute_force_max_attempts, brute_force_lockout_duration_minutes)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "test-email-7", 2, lastFailedTime, lockoutTime, 5, 20)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	bf := NewBruteForceProtection(db)

	// Get status
	status, err := bf.GetBruteForceStatus("test-email-7")
	if err != nil {
		t.Fatalf("GetBruteForceStatus failed: %v", err)
	}

	// Verify status fields
	if status.FailedAttempts != 2 {
		t.Errorf("Expected failed attempts to be 2, got %d", status.FailedAttempts)
	}

	if status.MaxAttempts != 5 {
		t.Errorf("Expected max attempts to be 5, got %d", status.MaxAttempts)
	}

	if status.LockoutDurationMinutes != 20 {
		t.Errorf("Expected lockout duration to be 20, got %d", status.LockoutDurationMinutes)
	}

	if !status.IsLockedOut() {
		t.Error("Expected email to be locked out")
	}

	remainingAttempts := status.GetRemainingAttempts()
	if remainingAttempts != 3 {
		t.Errorf("Expected remaining attempts to be 3, got %d", remainingAttempts)
	}

	remainingTime := status.GetLockoutRemainingTime()
	if remainingTime <= 0 {
		t.Error("Expected remaining lockout time to be positive")
	}
}

func TestBruteForceStatus_NotLockedOut(t *testing.T) {
	status := &BruteForceStatus{
		FailedAttempts:         1,
		MaxAttempts:            3,
		LockoutDurationMinutes: 15,
		LockoutUntil:           nil, // No lockout
	}

	if status.IsLockedOut() {
		t.Error("Expected email to not be locked out")
	}

	remainingAttempts := status.GetRemainingAttempts()
	if remainingAttempts != 2 {
		t.Errorf("Expected remaining attempts to be 2, got %d", remainingAttempts)
	}

	remainingTime := status.GetLockoutRemainingTime()
	if remainingTime != 0 {
		t.Errorf("Expected remaining lockout time to be 0, got %v", remainingTime)
	}
}

func TestBruteForceStatus_AtMaxAttempts(t *testing.T) {
	status := &BruteForceStatus{
		FailedAttempts:         3,
		MaxAttempts:            3,
		LockoutDurationMinutes: 15,
		LockoutUntil:           nil, // No lockout yet
	}

	remainingAttempts := status.GetRemainingAttempts()
	if remainingAttempts != 0 {
		t.Errorf("Expected remaining attempts to be 0, got %d", remainingAttempts)
	}
}

func TestIncrementFailedAttempt_EmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	bf := NewBruteForceProtection(db)

	// Try to increment attempts for non-existent email
	err := bf.IncrementFailedAttempt("non-existent-email")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestGetBruteForceStatus_EmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	bf := NewBruteForceProtection(db)

	// Try to get status for non-existent email
	_, err := bf.GetBruteForceStatus("non-existent-email")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}
