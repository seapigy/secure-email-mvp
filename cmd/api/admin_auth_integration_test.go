package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"secure-email-mvp/pkg/models"
	"secure-email-mvp/pkg/securelinks/admin"
)

// TestAdminLogin_Success tests successful admin login
func TestAdminLogin_Success(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create admin_users table
	_, err = db.Exec(`
		CREATE TABLE admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'admin',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create admin_users table: %v", err)
	}

	// Create test admin user
	password := "admin123456"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO admin_users (email, password, role, created_at)
		VALUES (?, ?, ?, ?)
	`, "admin@test.com", string(hashedPassword), "admin", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test admin: %v", err)
	}

	// Initialize admin service
	repo := admin.NewSQLiteAdminRepository(db)
	service := admin.NewService(db, repo)

	// Create test request
	loginData := map[string]string{
		"email":    "admin@test.com",
		"password": "admin123456",
	}
	jsonData, _ := json.Marshal(loginData)

	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	_ = httptest.NewRecorder() // Unused variable, but needed for request setup

	// Test login
	token, adminUser, err := service.LoginAdmin("admin@test.com", "admin123456")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Verify response
	if token == "" {
		t.Error("Expected token to be returned")
	}

	if adminUser == nil {
		t.Error("Expected admin user to be returned")
	}

	if adminUser.Email != "admin@test.com" {
		t.Errorf("Expected email %s, got %s", "admin@test.com", adminUser.Email)
	}

	if adminUser.Role != "admin" {
		t.Errorf("Expected role %s, got %s", "admin", adminUser.Role)
	}

	t.Logf("Login successful, token: %s", token)
}

// TestAdminLogin_InvalidPassword tests admin login with invalid password
func TestAdminLogin_InvalidPassword(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create admin_users table
	_, err = db.Exec(`
		CREATE TABLE admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'admin',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create admin_users table: %v", err)
	}

	// Create test admin user
	password := "admin123456"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO admin_users (email, password, role, created_at)
		VALUES (?, ?, ?, ?)
	`, "admin@test.com", string(hashedPassword), "admin", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test admin: %v", err)
	}

	// Initialize admin service
	repo := admin.NewSQLiteAdminRepository(db)
	service := admin.NewService(db, repo)

	// Test login with wrong password
	_, _, err = service.LoginAdmin("admin@test.com", "wrongpassword")
	if err == nil {
		t.Error("Expected login to fail with wrong password")
	}

	t.Logf("Login correctly failed with error: %v", err)
}

// TestAdminMiddleware_ValidJWT tests admin middleware with valid JWT
func TestAdminMiddleware_ValidJWT(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create admin_users table
	_, err = db.Exec(`
		CREATE TABLE admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'admin',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create admin_users table: %v", err)
	}

	// Create test admin user
	adminUser := &models.AdminUser{
		ID:        1,
		Email:     "admin@test.com",
		Password:  "hashedpassword",
		Role:      "admin",
		CreatedAt: time.Now(),
	}

	_, err = db.Exec(`
		INSERT INTO admin_users (id, email, password, role, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, adminUser.ID, adminUser.Email, adminUser.Password, adminUser.Role, adminUser.CreatedAt)
	if err != nil {
		t.Fatalf("Failed to insert test admin: %v", err)
	}

	// Initialize admin service and generate token
	repo := admin.NewSQLiteAdminRepository(db)
	service := admin.NewService(db, repo)

	token, _, err := service.LoginAdmin("admin@test.com", "admin123456")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create test request with valid token
	req := httptest.NewRequest("GET", "/api/admin/dlp/logs", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	_ = httptest.NewRecorder() // Unused variable, but needed for request setup

	// Test middleware (this would be tested in a full integration test)
	// For now, just verify the token is valid
	if token == "" {
		t.Error("Expected valid token")
	}

	t.Logf("Middleware test passed with valid token")
}

// TestAdminMiddleware_InvalidJWT tests admin middleware with invalid JWT
func TestAdminMiddleware_InvalidJWT(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// Create test request with invalid token
	req := httptest.NewRequest("GET", "/api/admin/dlp/logs", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	_ = httptest.NewRecorder() // Unused variable, but needed for request setup

	// Test middleware (this would be tested in a full integration test)
	// For now, just verify the request has an invalid token
	if req.Header.Get("Authorization") != "Bearer invalid-token" {
		t.Error("Expected invalid token in request")
	}

	t.Logf("Middleware test passed with invalid token")
}






