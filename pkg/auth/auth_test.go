package auth

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@securesystem.email", true},
		{"test@securesystem.email", true},
		{"user@example.com", false},
		{"invalid-email", false},
		{"@securesystem.email", false},
		{"user@", false},
	}

	for _, tt := range tests {
		result := ValidateEmail(tt.email)
		if result != tt.valid {
			t.Errorf("ValidateEmail(%s) = %v, want %v", tt.email, result, tt.valid)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"short", false},
		{"validpass", true},
		{"verylongpassword" + string(make([]byte, 100)), false},
		{"12345678", true},
		{"", false},
	}

	for _, tt := range tests {
		result := ValidatePassword(tt.password)
		if result != tt.valid {
			t.Errorf("ValidatePassword(%s) = %v, want %v", tt.password, result, tt.valid)
		}
	}
}

func TestValidateTOTP(t *testing.T) {
	tests := []struct {
		code  string
		valid bool
	}{
		{"123456", true},
		{"000000", true},
		{"12345", false},
		{"1234567", false},
		{"abcdef", false},
		{"", false},
	}

	for _, tt := range tests {
		result := ValidateTOTP(tt.code)
		if result != tt.valid {
			t.Errorf("ValidateTOTP(%s) = %v, want %v", tt.code, result, tt.valid)
		}
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	email := "test@securesystem.email"

	hash, err := HashPassword(password, email)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty string")
	}

	// Verify the hash can be used for verification
	expectedHash := argon2.IDKey([]byte(password), []byte(email), 1, 64*1024, 4, 32)
	if hash != string(expectedHash) {
		t.Error("HashPassword result doesn't match expected Argon2 hash")
	}
}

func TestGenerateTOTPSecret(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}

	if secret == "" {
		t.Error("GenerateTOTPSecret returned empty string")
	}

	// Test that the secret is valid base32
	if len(secret) != 32 { // 20 bytes = 32 base32 characters
		t.Errorf("Expected 32 base32 characters, got %d", len(secret))
	}
}

func TestAuthenticate(t *testing.T) {
	// Set JWT_SECRET for tests
	os.Setenv("JWT_SECRET", "test-secret-32-bytes-1234567890ab")

	// Setup in-memory SQLite
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT UNIQUE, password_hash TEXT, totp_secret TEXT)`)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	// Insert test user
	email := "test@securesystem.email"
	password := "securepass123"
	totpSecret, _ := totp.Generate(totp.GenerateOpts{Issuer: "SecureEmail", AccountName: email})
	hash := argon2.IDKey([]byte(password), []byte(email), 1, 64*1024, 4, 32)
	id := "test-user-id"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		id, email, string(hash), totpSecret.Secret())
	if err != nil {
		t.Fatal("Failed to insert user:", err)
	}

	// Generate valid TOTP code
	totpCode, _ := totp.GenerateCode(totpSecret.Secret(), time.Now())

	// Test successful authentication
	token, userID, err := Authenticate(db, email, password, totpCode)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if userID != id {
		t.Errorf("Expected userID %s, got %s", id, userID)
	}
	if token == "" {
		t.Error("Expected non-empty JWT token")
	}

	// Test invalid password
	_, _, err = Authenticate(db, email, "wrongpass", totpCode)
	if err == nil {
		t.Error("Expected error for invalid password")
	}

	// Test invalid TOTP
	_, _, err = Authenticate(db, email, password, "000000")
	if err == nil {
		t.Error("Expected error for invalid TOTP")
	}

	// Test invalid email
	_, _, err = Authenticate(db, "invalid@example.com", password, totpCode)
	if err == nil {
		t.Error("Expected error for invalid email")
	}

	// Test non-existent user
	_, _, err = Authenticate(db, "nonexistent@securesystem.email", password, totpCode)
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestCreateUser(t *testing.T) {
	// Setup in-memory SQLite
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT UNIQUE, password_hash TEXT, totp_secret TEXT)`)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	// Test successful user creation
	email := "newuser@securesystem.email"
	password := "newpass123"

	userID, totpSecret, err := CreateUser(db, email, password)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if userID == "" {
		t.Error("Expected non-empty user ID")
	}
	if totpSecret == "" {
		t.Error("Expected non-empty TOTP secret")
	}

	// Verify user was created in database
	var storedEmail, storedHash, storedTOTP string
	err = db.QueryRow("SELECT email, password_hash, totp_secret FROM users WHERE id = ?", userID).Scan(&storedEmail, &storedHash, &storedTOTP)
	if err != nil {
		t.Fatalf("Failed to query created user: %v", err)
	}

	if storedEmail != email {
		t.Errorf("Expected email %s, got %s", email, storedEmail)
	}
	if storedHash == "" {
		t.Error("Expected non-empty password hash")
	}
	if storedTOTP != totpSecret {
		t.Errorf("Expected TOTP secret %s, got %s", totpSecret, storedTOTP)
	}

	// Test duplicate user creation
	_, _, err = CreateUser(db, email, password)
	if err == nil {
		t.Error("Expected error for duplicate user")
	}

	// Test invalid email
	_, _, err = CreateUser(db, "invalid@example.com", password)
	if err == nil {
		t.Error("Expected error for invalid email")
	}

	// Test invalid password
	_, _, err = CreateUser(db, "valid@securesystem.email", "short")
	if err == nil {
		t.Error("Expected error for invalid password")
	}
}

func TestValidateJWT(t *testing.T) {
	// Set JWT_SECRET for tests
	os.Setenv("JWT_SECRET", "test-secret-32-bytes-1234567890ab")

	// Test with valid JWT (we'll create one using the auth package)
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create a user and authenticate to get a valid JWT
	_, err = db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT UNIQUE, password_hash TEXT, totp_secret TEXT)`)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	email := "test@securesystem.email"
	password := "securepass123"
	totpSecret, _ := totp.Generate(totp.GenerateOpts{Issuer: "SecureEmail", AccountName: email})
	hash := argon2.IDKey([]byte(password), []byte(email), 1, 64*1024, 4, 32)
	id := "test-user-id"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		id, email, string(hash), totpSecret.Secret())
	if err != nil {
		t.Fatal("Failed to insert user:", err)
	}

	totpCode, _ := totp.GenerateCode(totpSecret.Secret(), time.Now())
	token, _, err := Authenticate(db, email, password, totpCode)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Test valid JWT
	userID, userEmail, err := ValidateJWT(token)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if userID != id {
		t.Errorf("Expected userID %s, got %s", id, userID)
	}
	if userEmail != email {
		t.Errorf("Expected email %s, got %s", email, userEmail)
	}

	// Test invalid JWT
	_, _, err = ValidateJWT("invalid.jwt.token")
	if err == nil {
		t.Error("Expected error for invalid JWT")
	}

	// Test empty JWT
	_, _, err = ValidateJWT("")
	if err == nil {
		t.Error("Expected error for empty JWT")
	}
}
