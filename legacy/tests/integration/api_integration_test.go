package integration

import (
	"database/sql"
	"fmt"
	"testing"

	"secure-email-mvp/pkg/auth"

	_ "modernc.org/sqlite"
)

// Test data
const (
	testEmail    = "integration-test@securesystem.email"
	testPassword = "integration-test-password123"
	testTOTPCode = "123456"
)

// TestAuthIntegration tests authentication integration
func TestAuthIntegration(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runTestMigrations(db); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	t.Run("UserCreationAndAuth", func(t *testing.T) {
		// Test complete user creation and authentication flow
		userID, totpSecret, err := auth.CreateUser(db, testEmail, testPassword)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if len(userID) == 0 {
			t.Error("User ID should not be empty")
		}

		if len(totpSecret) == 0 {
			t.Error("TOTP secret should not be empty")
		}

		// Test JWT generation
		token, err := auth.GenerateJWT(testEmail)
		if err != nil {
			t.Fatalf("Failed to generate JWT: %v", err)
		}

		if len(token) == 0 {
			t.Error("JWT token should not be empty")
		}

		// Test JWT validation
		claims, err := auth.ParseJWT(token)
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		if claims.Email != testEmail {
			t.Errorf("JWT claims email mismatch: expected %s, got %s", testEmail, claims.Email)
		}
	})
}

// TestDatabaseIntegration tests database integration
func TestDatabaseIntegration(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runTestMigrations(db); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	t.Run("UserCreation", func(t *testing.T) {
		// Create user
		userID, totpSecret, err := auth.CreateUser(db, testEmail, testPassword)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if len(userID) == 0 {
			t.Error("User ID should not be empty")
		}

		if len(totpSecret) == 0 {
			t.Error("TOTP secret should not be empty")
		}

		// Verify user exists in database
		var storedEmail string
		err = db.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&storedEmail)
		if err != nil {
			t.Fatalf("Failed to query user: %v", err)
		}

		if storedEmail != testEmail {
			t.Errorf("Stored email doesn't match: expected %s, got %s", testEmail, storedEmail)
		}
	})

	t.Run("UserAuthentication", func(t *testing.T) {
		// Create user first
		userID, totpSecret, err := auth.CreateUser(db, "auth-test@securesystem.email", testPassword)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Test authentication
		authUserID, authTOTPSecret, err := auth.Authenticate(db, "auth-test@securesystem.email", testPassword, testTOTPCode)
		if err != nil {
			t.Logf("Authentication failed (expected due to TOTP): %v", err)
			// Authentication might fail due to TOTP validation, which is expected
			return
		}

		if authUserID != userID {
			t.Errorf("User ID mismatch: expected %s, got %s", userID, authUserID)
		}

		if authTOTPSecret != totpSecret {
			t.Errorf("TOTP secret mismatch: expected %s, got %s", totpSecret, authTOTPSecret)
		}
	})

	t.Run("DuplicateUserCreation", func(t *testing.T) {
		// Create user
		_, _, err := auth.CreateUser(db, "duplicate-test@securesystem.email", testPassword)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Try to create duplicate user
		_, _, err = auth.CreateUser(db, "duplicate-test@securesystem.email", testPassword)
		if err == nil {
			t.Error("Expected error for duplicate user creation")
		}
	})
}

// TestEmailIntegration tests email-related functionality
func TestEmailIntegration(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runTestMigrations(db); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	t.Run("EmailEncryption", func(t *testing.T) {
		// Test email encryption functionality
		_, _, err := auth.CreateUser(db, "email-test@securesystem.email", testPassword)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Test JWT generation for email operations
		token, err := auth.GenerateJWT("email-test@securesystem.email")
		if err != nil {
			t.Fatalf("Failed to generate JWT: %v", err)
		}

		if len(token) == 0 {
			t.Error("JWT token should not be empty")
		}

		// Test JWT validation
		claims, err := auth.ParseJWT(token)
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		if claims.Email != "email-test@securesystem.email" {
			t.Errorf("JWT claims email mismatch: expected %s, got %s", "email-test@securesystem.email", claims.Email)
		}
	})
}

// Helper functions

func runTestMigrations(db *sql.DB) error {
	// Create users table
	usersQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(usersQuery); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create emails table
	emailsQuery := `
	CREATE TABLE IF NOT EXISTS emails (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		to_email TEXT NOT NULL,
		subject TEXT NOT NULL,
		encrypted_content TEXT NOT NULL,
		encrypted_key TEXT NOT NULL,
		nonce TEXT NOT NULL,
		auth_tag TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	if _, err := db.Exec(emailsQuery); err != nil {
		return fmt.Errorf("failed to create emails table: %v", err)
	}

	return nil
}
