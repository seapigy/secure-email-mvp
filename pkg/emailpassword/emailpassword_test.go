package emailpassword

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create emails table with password protection fields
	_, err = db.Exec(`
		CREATE TABLE emails (
			email_id TEXT PRIMARY KEY,
			is_password_protected BOOLEAN DEFAULT FALSE,
			password_hash TEXT,
			password_salt TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create index
	_, err = db.Exec(`CREATE INDEX idx_emails_password_protection ON emails(is_password_protected)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	return db
}

func TestSetEmailPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-1")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailPasswordService(db)

	// Test setting password
	err = service.SetEmailPassword("test-email-1", "testpassword123")
	if err != nil {
		t.Fatalf("SetEmailPassword failed: %v", err)
	}

	// Verify password was set
	var isPasswordProtected bool
	var passwordHash, passwordSalt string
	err = db.QueryRow(`
		SELECT is_password_protected, password_hash, password_salt
		FROM emails WHERE email_id = ?
	`, "test-email-1").Scan(&isPasswordProtected, &passwordHash, &passwordSalt)
	if err != nil {
		t.Fatalf("Failed to verify password was set: %v", err)
	}

	if !isPasswordProtected {
		t.Error("Expected email to be password-protected")
	}
	if passwordHash == "" {
		t.Error("Expected password hash to be set")
	}
	if passwordSalt == "" {
		t.Error("Expected password salt to be set")
	}
}

func TestSetEmailPassword_EmptyEmailID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	err := service.SetEmailPassword("", "testpassword123")
	if err == nil {
		t.Error("Expected error for empty email ID")
	}
}

func TestSetEmailPassword_EmptyPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-2")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailPasswordService(db)

	err = service.SetEmailPassword("test-email-2", "")
	if err == nil {
		t.Error("Expected error for empty password")
	}
}

func TestSetEmailPassword_EmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	err := service.SetEmailPassword("nonexistent-email", "testpassword123")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestCheckEmailPassword_CorrectPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-3")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailPasswordService(db)

	// Set password
	err = service.SetEmailPassword("test-email-3", "testpassword123")
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	// Check correct password
	valid, err := service.CheckEmailPassword("test-email-3", "testpassword123")
	if err != nil {
		t.Fatalf("CheckEmailPassword failed: %v", err)
	}

	if !valid {
		t.Error("Expected correct password to be valid")
	}
}

func TestCheckEmailPassword_IncorrectPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-4")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailPasswordService(db)

	// Set password
	err = service.SetEmailPassword("test-email-4", "testpassword123")
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	// Check incorrect password
	valid, err := service.CheckEmailPassword("test-email-4", "wrongpassword")
	if err != nil {
		t.Fatalf("CheckEmailPassword failed: %v", err)
	}

	if valid {
		t.Error("Expected incorrect password to be invalid")
	}
}

func TestCheckEmailPassword_NotPasswordProtected(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email (not password-protected)
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-5")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailPasswordService(db)

	// Check password for non-protected email
	valid, err := service.CheckEmailPassword("test-email-5", "anypassword")
	if err != nil {
		t.Fatalf("CheckEmailPassword failed: %v", err)
	}

	if !valid {
		t.Error("Expected non-protected email to be accessible")
	}
}

func TestCheckEmailPassword_EmptyEmailID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	_, err := service.CheckEmailPassword("", "testpassword123")
	if err == nil {
		t.Error("Expected error for empty email ID")
	}
}

func TestCheckEmailPassword_EmptyPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-6")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailPasswordService(db)

	_, err = service.CheckEmailPassword("test-email-6", "")
	if err == nil {
		t.Error("Expected error for empty password")
	}
}

func TestCheckEmailPassword_EmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	_, err := service.CheckEmailPassword("nonexistent-email", "testpassword123")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestClearEmailPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-7")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailPasswordService(db)

	// Set password
	err = service.SetEmailPassword("test-email-7", "testpassword123")
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	// Clear password
	err = service.ClearEmailPassword("test-email-7")
	if err != nil {
		t.Fatalf("ClearEmailPassword failed: %v", err)
	}

	// Verify password was cleared
	var isPasswordProtected bool
	var passwordHash, passwordSalt *string
	err = db.QueryRow(`
		SELECT is_password_protected, password_hash, password_salt
		FROM emails WHERE email_id = ?
	`, "test-email-7").Scan(&isPasswordProtected, &passwordHash, &passwordSalt)
	if err != nil {
		t.Fatalf("Failed to verify password was cleared: %v", err)
	}

	if isPasswordProtected {
		t.Error("Expected email to not be password-protected")
	}
	if passwordHash != nil {
		t.Error("Expected password hash to be NULL")
	}
	if passwordSalt != nil {
		t.Error("Expected password salt to be NULL")
	}
}

func TestClearEmailPassword_EmptyEmailID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	err := service.ClearEmailPassword("")
	if err == nil {
		t.Error("Expected error for empty email ID")
	}
}

func TestClearEmailPassword_EmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	err := service.ClearEmailPassword("nonexistent-email")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestIsEmailPasswordProtected(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test emails
	_, err := db.Exec(`
		INSERT INTO emails (email_id, is_password_protected) 
		VALUES (?, ?), (?, ?)
	`, "test-email-8", true, "test-email-9", false)
	if err != nil {
		t.Fatalf("Failed to insert test emails: %v", err)
	}

	service := NewEmailPasswordService(db)

	// Test password-protected email
	protected, err := service.IsEmailPasswordProtected("test-email-8")
	if err != nil {
		t.Fatalf("IsEmailPasswordProtected failed: %v", err)
	}
	if !protected {
		t.Error("Expected email to be password-protected")
	}

	// Test non-password-protected email
	protected, err = service.IsEmailPasswordProtected("test-email-9")
	if err != nil {
		t.Fatalf("IsEmailPasswordProtected failed: %v", err)
	}
	if protected {
		t.Error("Expected email to not be password-protected")
	}
}

func TestIsEmailPasswordProtected_EmptyEmailID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	_, err := service.IsEmailPasswordProtected("")
	if err == nil {
		t.Error("Expected error for empty email ID")
	}
}

func TestIsEmailPasswordProtected_EmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	_, err := service.IsEmailPasswordProtected("nonexistent-email")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	// Test valid password
	err := service.ValidatePasswordStrength("strongpassword123")
	if err != nil {
		t.Errorf("Expected valid password to pass validation: %v", err)
	}

	// Test empty password
	err = service.ValidatePasswordStrength("")
	if err == nil {
		t.Error("Expected error for empty password")
	}

	// Test too short password
	err = service.ValidatePasswordStrength("short")
	if err == nil {
		t.Error("Expected error for too short password")
	}

	// Test too long password
	longPassword := string(make([]byte, 129))
	err = service.ValidatePasswordStrength(longPassword)
	if err == nil {
		t.Error("Expected error for too long password")
	}

	// Test weak passwords
	weakPasswords := []string{"password", "123456", "qwerty", "admin"}
	for _, weak := range weakPasswords {
		err = service.ValidatePasswordStrength(weak)
		if err == nil {
			t.Errorf("Expected error for weak password: %s", weak)
		}
	}
}

func TestGetPasswordProtectionStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test emails
	_, err := db.Exec(`
		INSERT INTO emails (email_id, is_password_protected, password_hash) 
		VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)
	`, "test-email-10", true, "hash1", "test-email-11", false, nil, "test-email-12", true, nil)
	if err != nil {
		t.Fatalf("Failed to insert test emails: %v", err)
	}

	service := NewEmailPasswordService(db)

	// Test password-protected email with password set
	status, err := service.GetPasswordProtectionStatus("test-email-10")
	if err != nil {
		t.Fatalf("GetPasswordProtectionStatus failed: %v", err)
	}
	if !status.IsPasswordProtected {
		t.Error("Expected email to be password-protected")
	}
	if !status.HasPasswordSet {
		t.Error("Expected email to have password set")
	}

	// Test non-password-protected email
	status, err = service.GetPasswordProtectionStatus("test-email-11")
	if err != nil {
		t.Fatalf("GetPasswordProtectionStatus failed: %v", err)
	}
	if status.IsPasswordProtected {
		t.Error("Expected email to not be password-protected")
	}
	if status.HasPasswordSet {
		t.Error("Expected email to not have password set")
	}

	// Test password-protected email without password set
	status, err = service.GetPasswordProtectionStatus("test-email-12")
	if err != nil {
		t.Fatalf("GetPasswordProtectionStatus failed: %v", err)
	}
	if !status.IsPasswordProtected {
		t.Error("Expected email to be password-protected")
	}
	if status.HasPasswordSet {
		t.Error("Expected email to not have password set")
	}
}

func TestGetPasswordProtectionStatus_EmptyEmailID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	_, err := service.GetPasswordProtectionStatus("")
	if err == nil {
		t.Error("Expected error for empty email ID")
	}
}

func TestGetPasswordProtectionStatus_EmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewEmailPasswordService(db)

	_, err := service.GetPasswordProtectionStatus("nonexistent-email")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestDefaultArgon2Config(t *testing.T) {
	config := DefaultArgon2Config()

	if config.Memory != 64*1024 {
		t.Errorf("Expected Memory to be 64*1024, got %d", config.Memory)
	}

	if config.Time != 3 {
		t.Errorf("Expected Time to be 3, got %d", config.Time)
	}

	if config.Parallelism != 2 {
		t.Errorf("Expected Parallelism to be 2, got %d", config.Parallelism)
	}

	if config.KeyLength != 32 {
		t.Errorf("Expected KeyLength to be 32, got %d", config.KeyLength)
	}

	if config.SaltLength != 16 {
		t.Errorf("Expected SaltLength to be 16, got %d", config.SaltLength)
	}
}

func TestPasswordHashingUniqueness(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-13")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailPasswordService(db)

	// Set password multiple times
	password := "testpassword123"
	err = service.SetEmailPassword("test-email-13", password)
	if err != nil {
		t.Fatalf("Failed to set password first time: %v", err)
	}

	// Get first hash and salt
	var hash1, salt1 string
	err = db.QueryRow(`
		SELECT password_hash, password_salt
		FROM emails WHERE email_id = ?
	`, "test-email-13").Scan(&hash1, &salt1)
	if err != nil {
		t.Fatalf("Failed to get first hash and salt: %v", err)
	}

	// Set password again
	err = service.SetEmailPassword("test-email-13", password)
	if err != nil {
		t.Fatalf("Failed to set password second time: %v", err)
	}

	// Get second hash and salt
	var hash2, salt2 string
	err = db.QueryRow(`
		SELECT password_hash, password_salt
		FROM emails WHERE email_id = ?
	`, "test-email-13").Scan(&hash2, &salt2)
	if err != nil {
		t.Fatalf("Failed to get second hash and salt: %v", err)
	}

	// Verify salt is different (due to random generation)
	if salt1 == salt2 {
		t.Error("Expected different salts for each password set")
	}

	// Verify hash is different (due to different salt)
	if hash1 == hash2 {
		t.Error("Expected different hashes for each password set")
	}

	// Verify both passwords still work
	valid1, err := service.CheckEmailPassword("test-email-13", password)
	if err != nil {
		t.Fatalf("CheckEmailPassword failed: %v", err)
	}
	if !valid1 {
		t.Error("Expected password to be valid after second set")
	}
}
