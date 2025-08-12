package lockout

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupTestDB creates a test database with the required schema
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create users table with lockout fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			totp_secret TEXT,
			failed_login_attempts INTEGER DEFAULT 0,
			last_failed_login TIMESTAMP,
			account_locked_until TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	return db
}

// insertTestUser inserts a test user into the database
func insertTestUser(t *testing.T, db *sql.DB, email string) {
	_, err := db.Exec(`
		INSERT INTO users (id, email, password_hash, totp_secret) 
		VALUES (?, ?, ?, ?)
	`, "test-user-123", email, "hashed_password", "test_totp_secret")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
}

func TestNewUserLockoutService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewUserLockoutService(db)
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.db != db {
		t.Error("Expected database to be set")
	}

	if service.config == nil {
		t.Fatal("Expected config to be set")
	}
}

func TestNewUserLockoutServiceWithConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	config := &LockoutConfig{
		MaxAttempts:     3,
		LockoutDuration: 10 * time.Minute,
		AttemptWindow:   5 * time.Minute,
		Enabled:         true,
	}

	service := NewUserLockoutServiceWithConfig(db, config)
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.config != config {
		t.Error("Expected custom config to be set")
	}
}

func TestDefaultConfig(t *testing.T) {
	// Clear environment variables for this test
	os.Unsetenv("LOGIN_MAX_ATTEMPTS")
	os.Unsetenv("LOGIN_LOCKOUT_MINUTES")
	os.Unsetenv("LOGIN_ATTEMPT_WINDOW_MINUTES")
	os.Unsetenv("LOGIN_RATE_LIMIT_ENABLED")

	config := DefaultConfig()

	if config.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts to be 5, got %d", config.MaxAttempts)
	}

	if config.LockoutDuration != 30*time.Minute {
		t.Errorf("Expected LockoutDuration to be 30 minutes, got %v", config.LockoutDuration)
	}

	if config.AttemptWindow != 15*time.Minute {
		t.Errorf("Expected AttemptWindow to be 15 minutes, got %v", config.AttemptWindow)
	}

	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
}

func TestDefaultConfigWithEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("LOGIN_MAX_ATTEMPTS", "3")
	os.Setenv("LOGIN_LOCKOUT_MINUTES", "10")
	os.Setenv("LOGIN_ATTEMPT_WINDOW_MINUTES", "5")
	os.Setenv("LOGIN_RATE_LIMIT_ENABLED", "1")

	config := DefaultConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts to be 3, got %d", config.MaxAttempts)
	}

	if config.LockoutDuration != 10*time.Minute {
		t.Errorf("Expected LockoutDuration to be 10 minutes, got %v", config.LockoutDuration)
	}

	if config.AttemptWindow != 5*time.Minute {
		t.Errorf("Expected AttemptWindow to be 5 minutes, got %v", config.AttemptWindow)
	}

	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}

	// Clean up
	os.Unsetenv("LOGIN_MAX_ATTEMPTS")
	os.Unsetenv("LOGIN_LOCKOUT_MINUTES")
	os.Unsetenv("LOGIN_ATTEMPT_WINDOW_MINUTES")
	os.Unsetenv("LOGIN_RATE_LIMIT_ENABLED")
}

func TestCheckUserLockout_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewUserLockoutService(db)
	locked, lockedUntil, err := service.CheckUserLockout("nonexistent@example.com")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if locked {
		t.Error("Expected user not to be locked")
	}

	if lockedUntil != nil {
		t.Error("Expected no lockout time")
	}
}

func TestCheckUserLockout_NotLocked(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	service := NewUserLockoutService(db)
	locked, lockedUntil, err := service.CheckUserLockout(email)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if locked {
		t.Error("Expected user not to be locked")
	}

	if lockedUntil != nil {
		t.Error("Expected no lockout time")
	}
}

func TestCheckUserLockout_Locked(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	// Manually set lockout
	lockoutTime := time.Now().Add(30 * time.Minute)
	_, err := db.Exec(`
		UPDATE users 
		SET account_locked_until = ?, failed_login_attempts = 5
		WHERE email = ?
	`, lockoutTime, email)
	if err != nil {
		t.Fatalf("Failed to set lockout: %v", err)
	}

	service := NewUserLockoutService(db)
	locked, lockedUntil, err := service.CheckUserLockout(email)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !locked {
		t.Error("Expected user to be locked")
	}

	if lockedUntil == nil {
		t.Fatal("Expected lockout time to be set")
	}

	if !lockedUntil.Equal(lockoutTime) {
		t.Errorf("Expected lockout time to match, got %v", lockedUntil)
	}
}

func TestCheckUserLockout_ExpiredLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	// Manually set expired lockout
	lockoutTime := time.Now().Add(-30 * time.Minute) // Past time
	_, err := db.Exec(`
		UPDATE users 
		SET account_locked_until = ?, failed_login_attempts = 5
		WHERE email = ?
	`, lockoutTime, email)
	if err != nil {
		t.Fatalf("Failed to set lockout: %v", err)
	}

	service := NewUserLockoutService(db)
	locked, lockedUntil, err := service.CheckUserLockout(email)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if locked {
		t.Error("Expected user not to be locked (expired)")
	}

	if lockedUntil != nil {
		t.Error("Expected no lockout time (expired)")
	}

	// Verify lockout was cleared
	var currentLockout *time.Time
	err = db.QueryRow("SELECT account_locked_until FROM users WHERE email = ?", email).Scan(&currentLockout)
	if err != nil {
		t.Fatalf("Failed to check lockout status: %v", err)
	}

	if currentLockout != nil {
		t.Error("Expected lockout to be cleared")
	}
}

func TestIncrementUserFailedAttempt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	service := NewUserLockoutService(db)

	// Increment failed attempts
	err := service.IncrementUserFailedAttempt(email)
	if err != nil {
		t.Fatalf("Failed to increment failed attempts: %v", err)
	}

	// Check that attempts were incremented
	var failedAttempts int
	err = db.QueryRow("SELECT failed_login_attempts FROM users WHERE email = ?", email).Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to get failed attempts: %v", err)
	}

	if failedAttempts != 1 {
		t.Errorf("Expected 1 failed attempt, got %d", failedAttempts)
	}
}

func TestIncrementUserFailedAttempt_TriggerLockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	service := NewUserLockoutService(db)

	// Increment failed attempts to trigger lockout
	for i := 0; i < 5; i++ {
		err := service.IncrementUserFailedAttempt(email)
		if err != nil {
			t.Fatalf("Failed to increment failed attempts: %v", err)
		}
	}

	// Check that lockout was triggered
	var lockedUntil *time.Time
	err := db.QueryRow("SELECT account_locked_until FROM users WHERE email = ?", email).Scan(&lockedUntil)
	if err != nil {
		t.Fatalf("Failed to get locked until: %v", err)
	}

	if lockedUntil == nil {
		t.Error("Expected lockout to be triggered")
	}

	// Check that it's locked
	locked, _, err := service.CheckUserLockout(email)
	if err != nil {
		t.Fatalf("Failed to check user lockout: %v", err)
	}

	if !locked {
		t.Error("Expected user to be locked after 5 failed attempts")
	}
}

func TestIncrementUserFailedAttempt_OutsideWindow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	// Set old failed login time (outside window)
	oldTime := time.Now().Add(-20 * time.Minute) // Outside 15-minute window
	_, err := db.Exec(`
		UPDATE users 
		SET failed_login_attempts = 3, last_failed_login = ?
		WHERE email = ?
	`, oldTime, email)
	if err != nil {
		t.Fatalf("Failed to set old failed login: %v", err)
	}

	service := NewUserLockoutService(db)

	// Increment failed attempts
	err = service.IncrementUserFailedAttempt(email)
	if err != nil {
		t.Fatalf("Failed to increment failed attempts: %v", err)
	}

	// Check that attempts were reset to 1 (not 4)
	var failedAttempts int
	err = db.QueryRow("SELECT failed_login_attempts FROM users WHERE email = ?", email).Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to get failed attempts: %v", err)
	}

	if failedAttempts != 1 {
		t.Errorf("Expected 1 failed attempt (reset), got %d", failedAttempts)
	}
}

func TestResetUserFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	// Set failed attempts and lockout
	lockoutTime := time.Now().Add(30 * time.Minute)
	_, err := db.Exec(`
		UPDATE users 
		SET failed_login_attempts = 3, account_locked_until = ?
		WHERE email = ?
	`, lockoutTime, email)
	if err != nil {
		t.Fatalf("Failed to set failed attempts: %v", err)
	}

	service := NewUserLockoutService(db)

	// Reset failed attempts
	err = service.ResetUserFailedAttempts(email)
	if err != nil {
		t.Fatalf("Failed to reset failed attempts: %v", err)
	}

	// Check that attempts were reset
	var failedAttempts int
	var lockedUntil *time.Time
	err = db.QueryRow("SELECT failed_login_attempts, account_locked_until FROM users WHERE email = ?", email).Scan(&failedAttempts, &lockedUntil)
	if err != nil {
		t.Fatalf("Failed to get user status: %v", err)
	}

	if failedAttempts != 0 {
		t.Errorf("Expected 0 failed attempts, got %d", failedAttempts)
	}

	if lockedUntil != nil {
		t.Error("Expected lockout to be cleared")
	}
}

func TestGetUserLockoutStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	// Set some failed attempts
	lastFailedLogin := time.Now().Add(-5 * time.Minute)
	_, err := db.Exec(`
		UPDATE users 
		SET failed_login_attempts = 2, last_failed_login = ?
		WHERE email = ?
	`, lastFailedLogin, email)
	if err != nil {
		t.Fatalf("Failed to set failed attempts: %v", err)
	}

	service := NewUserLockoutService(db)
	status, err := service.GetUserLockoutStatus(email)
	if err != nil {
		t.Fatalf("Failed to get lockout status: %v", err)
	}

	if status.Email != email {
		t.Errorf("Expected email %s, got %s", email, status.Email)
	}

	if status.FailedAttempts != 2 {
		t.Errorf("Expected 2 failed attempts, got %d", status.FailedAttempts)
	}

	if status.MaxAttempts != 5 {
		t.Errorf("Expected 5 max attempts, got %d", status.MaxAttempts)
	}

	if status.GetRemainingAttempts() != 3 {
		t.Errorf("Expected 3 remaining attempts, got %d", status.GetRemainingAttempts())
	}

	if status.IsLockedOut() {
		t.Error("Expected user not to be locked out")
	}

	if !status.IsWithinAttemptWindow() {
		t.Error("Expected to be within attempt window")
	}
}

func TestUserLockoutStatus_IsLockedOut(t *testing.T) {
	now := time.Now()
	future := now.Add(30 * time.Minute)
	past := now.Add(-30 * time.Minute)

	status := &UserLockoutStatus{
		LockoutUntil: &future,
	}

	if !status.IsLockedOut() {
		t.Error("Expected user to be locked out")
	}

	status.LockoutUntil = &past
	if status.IsLockedOut() {
		t.Error("Expected user not to be locked out (expired)")
	}

	status.LockoutUntil = nil
	if status.IsLockedOut() {
		t.Error("Expected user not to be locked out (no lockout)")
	}
}

func TestUserLockoutStatus_GetRemainingAttempts(t *testing.T) {
	status := &UserLockoutStatus{
		FailedAttempts: 2,
		MaxAttempts:    5,
	}

	if status.GetRemainingAttempts() != 3 {
		t.Errorf("Expected 3 remaining attempts, got %d", status.GetRemainingAttempts())
	}

	status.FailedAttempts = 5
	if status.GetRemainingAttempts() != 0 {
		t.Errorf("Expected 0 remaining attempts, got %d", status.GetRemainingAttempts())
	}
}

func TestUserLockoutStatus_GetLockoutRemainingTime(t *testing.T) {
	now := time.Now()
	future := now.Add(30 * time.Minute)

	status := &UserLockoutStatus{
		LockoutUntil: &future,
	}

	remaining := status.GetLockoutRemainingTime()
	if remaining <= 0 {
		t.Errorf("Expected positive remaining time, got %v", remaining)
	}

	status.LockoutUntil = &now
	remaining = status.GetLockoutRemainingTime()
	if remaining != 0 {
		t.Errorf("Expected 0 remaining time, got %v", remaining)
	}
}

func TestUserLockoutStatus_IsWithinAttemptWindow(t *testing.T) {
	now := time.Now()
	recent := now.Add(-5 * time.Minute)
	old := now.Add(-20 * time.Minute)

	status := &UserLockoutStatus{
		LastFailedLogin: &recent,
		AttemptWindow:   15 * time.Minute,
	}

	if !status.IsWithinAttemptWindow() {
		t.Error("Expected to be within attempt window")
	}

	status.LastFailedLogin = &old
	if status.IsWithinAttemptWindow() {
		t.Error("Expected to be outside attempt window")
	}

	status.LastFailedLogin = nil
	if status.IsWithinAttemptWindow() {
		t.Error("Expected to be outside attempt window (no last failed login)")
	}
}

func TestCheckUserLockout_Disabled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	config := &LockoutConfig{
		Enabled: false,
	}
	service := NewUserLockoutServiceWithConfig(db, config)

	locked, lockedUntil, err := service.CheckUserLockout(email)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if locked {
		t.Error("Expected user not to be locked (service disabled)")
	}

	if lockedUntil != nil {
		t.Error("Expected no lockout time (service disabled)")
	}
}

func TestIncrementUserFailedAttempt_Disabled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	email := "test@example.com"
	insertTestUser(t, db, email)

	config := &LockoutConfig{
		Enabled: false,
	}
	service := NewUserLockoutServiceWithConfig(db, config)

	// Should not error when disabled
	err := service.IncrementUserFailedAttempt(email)
	if err != nil {
		t.Fatalf("Expected no error when disabled, got %v", err)
	}

	// Check that attempts were not incremented
	var failedAttempts int
	err = db.QueryRow("SELECT failed_login_attempts FROM users WHERE email = ?", email).Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to get failed attempts: %v", err)
	}

	if failedAttempts != 0 {
		t.Errorf("Expected 0 failed attempts (service disabled), got %d", failedAttempts)
	}
}
