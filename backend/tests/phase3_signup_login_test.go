package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"securemail-backend/auth"
)

// Test enhanced signup flow with account type selection
func TestEnhancedSignupFlow(t *testing.T) {
	db := setupPhase3TestDB(t)
	defer db.Close()

	t.Run("Free Account Signup", func(t *testing.T) {
		reqBody := map[string]string{
			"username":      "testuser",
			"email":         "test@example.com",
			"password":      "TestPassword123!",
			"fallback_email": "backup@example.com",
			"account_type":  "free",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		auth.SignupHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		// Verify user was created with free account
		var accountType string
		err := db.QueryRow("SELECT account_type_new FROM users WHERE email = ?", "test@example.com").Scan(&accountType)
		if err != nil {
			t.Fatalf("Failed to check account type: %v", err)
		}
		if accountType != "free" {
			t.Errorf("Expected account type 'free', got '%s'", accountType)
		}

		// Verify no subscription was created for free account
		var subscriptionCount int
		err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = (SELECT id FROM users WHERE email = ?)", "test@example.com").Scan(&subscriptionCount)
		if err != nil {
			t.Fatalf("Failed to check subscription count: %v", err)
		}
		if subscriptionCount != 0 {
			t.Errorf("Expected 0 subscriptions for free account, got %d", subscriptionCount)
		}
	})

	t.Run("Premium Account Signup", func(t *testing.T) {
		reqBody := map[string]string{
			"username":      "premiumuser",
			"email":         "premium@example.com",
			"password":      "TestPassword123!",
			"fallback_email": "premiumbackup@example.com",
			"account_type":  "premium",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		auth.SignupHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		// Verify user was created with premium account
		var accountType string
		err := db.QueryRow("SELECT account_type_new FROM users WHERE email = ?", "premium@example.com").Scan(&accountType)
		if err != nil {
			t.Fatalf("Failed to check account type: %v", err)
		}
		if accountType != "premium" {
			t.Errorf("Expected account type 'premium', got '%s'", accountType)
		}

		// Verify placeholder subscription was created
		var subscriptionCount int
		err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = (SELECT id FROM users WHERE email = ?)", "premium@example.com").Scan(&subscriptionCount)
		if err != nil {
			t.Fatalf("Failed to check subscription count: %v", err)
		}
		if subscriptionCount != 1 {
			t.Errorf("Expected 1 subscription for premium account, got %d", subscriptionCount)
		}
	})

	t.Run("Enterprise Account Signup", func(t *testing.T) {
		reqBody := map[string]string{
			"username":      "enterpriseuser",
			"email":         "enterprise@example.com",
			"password":      "TestPassword123!",
			"fallback_email": "enterprisebackup@example.com",
			"account_type":  "enterprise",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		auth.SignupHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		// Verify user was created with enterprise account
		var accountType string
		err := db.QueryRow("SELECT account_type_new FROM users WHERE email = ?", "enterprise@example.com").Scan(&accountType)
		if err != nil {
			t.Fatalf("Failed to check account type: %v", err)
		}
		if accountType != "enterprise" {
			t.Errorf("Expected account type 'enterprise', got '%s'", accountType)
		}

		// Verify placeholder subscription was created
		var subscriptionCount int
		err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = (SELECT id FROM users WHERE email = ?)", "enterprise@example.com").Scan(&subscriptionCount)
		if err != nil {
			t.Fatalf("Failed to check subscription count: %v", err)
		}
		if subscriptionCount != 1 {
			t.Errorf("Expected 1 subscription for enterprise account, got %d", subscriptionCount)
		}
	})

	t.Run("Invalid Account Type", func(t *testing.T) {
		reqBody := map[string]string{
			"username":    "invaliduser",
			"email":       "invalid@example.com",
			"password":    "testpassword123",
			"account_type": "invalid",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		auth.SignupHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// Test enhanced login flow with organization information
func TestEnhancedLoginFlow(t *testing.T) {
	db := setupPhase3TestDB(t)
	defer db.Close()

	// Create test users with different account types (email verified for login tests)
	createTestUser(t, db, "freeuser", "free@example.com", "free", true)
	createTestUser(t, db, "premiumuser", "premium@example.com", "premium", true)
	createTestUser(t, db, "enterpriseuser", "enterprise@example.com", "enterprise", true)

	// Create organization and add enterprise user
	orgID := createTestOrganization(t, db, "Test Org", "enterprise@example.com")
	addUserToOrganization(t, db, orgID, "enterprise@example.com", "admin")

	t.Run("Free User Login", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "free@example.com",
			"password": "testpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		auth.LoginHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Parse response
		var response struct {
			Token            string `json:"token"`
			ExpiresAt        string `json:"expires_at"`
			AccountType      string `json:"account_type"`
			OrganizationID   string `json:"organization_id,omitempty"`
			OrganizationRole string `json:"organization_role,omitempty"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify response contains account type
		if response.AccountType != "free" {
			t.Errorf("Expected account type 'free', got '%s'", response.AccountType)
		}

		// Verify no organization information for free user
		if response.OrganizationID != "" {
			t.Errorf("Expected empty organization ID for free user, got '%s'", response.OrganizationID)
		}
	})

	t.Run("Premium User Login", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "premium@example.com",
			"password": "testpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		auth.LoginHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Parse response
		var response struct {
			Token            string `json:"token"`
			ExpiresAt        string `json:"expires_at"`
			AccountType      string `json:"account_type"`
			OrganizationID   string `json:"organization_id,omitempty"`
			OrganizationRole string `json:"organization_role,omitempty"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify response contains account type
		if response.AccountType != "premium" {
			t.Errorf("Expected account type 'premium', got '%s'", response.AccountType)
		}
	})

	t.Run("Enterprise User with Organization Login", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "enterprise@example.com",
			"password": "testpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		auth.LoginHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Parse response
		var response struct {
			Token            string `json:"token"`
			ExpiresAt        string `json:"expires_at"`
			AccountType      string `json:"account_type"`
			OrganizationID   string `json:"organization_id,omitempty"`
			OrganizationRole string `json:"organization_role,omitempty"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify response contains account type
		if response.AccountType != "enterprise" {
			t.Errorf("Expected account type 'enterprise', got '%s'", response.AccountType)
		}

		// Verify organization information is included
		if response.OrganizationID == "" {
			t.Errorf("Expected organization ID for enterprise user")
		}
		if response.OrganizationRole != "admin" {
			t.Errorf("Expected organization role 'admin', got '%s'", response.OrganizationRole)
		}
	})
}

// Test role enforcement middleware
func TestRoleEnforcementMiddleware(t *testing.T) {
	db := setupPhase3TestDB(t)
	defer db.Close()

	// Create test users
	createTestUser(t, db, "freeuser", "free@example.com", "free", true)
	createTestUser(t, db, "premiumuser", "premium@example.com", "premium", true)
	createTestUser(t, db, "enterpriseuser", "enterprise@example.com", "enterprise", true)

	// Create sessions for users
	freeToken := createTestSession(t, db, "free@example.com")
	premiumToken := createTestSession(t, db, "premium@example.com")
	enterpriseToken := createTestSession(t, db, "enterprise@example.com")

	t.Run("RequirePremiumOrEnterprise - Free User", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/domain/add", nil)
		req.Header.Set("Authorization", "Bearer "+freeToken)

		// Add user context
		userID := getUserIDByEmail(t, db, "free@example.com")
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler := auth.RequirePremiumOrEnterprise(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("RequirePremiumOrEnterprise - Premium User", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/domain/add", nil)
		req.Header.Set("Authorization", "Bearer "+premiumToken)

		// Add user context
		userID := getUserIDByEmail(t, db, "premium@example.com")
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler := auth.RequirePremiumOrEnterprise(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("RequireEnterprise - Premium User", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/org/create", nil)
		req.Header.Set("Authorization", "Bearer "+premiumToken)

		// Add user context
		userID := getUserIDByEmail(t, db, "premium@example.com")
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler := auth.RequireEnterprise(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("RequireEnterprise - Enterprise User", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/org/create", nil)
		req.Header.Set("Authorization", "Bearer "+enterpriseToken)

		// Add user context
		userID := getUserIDByEmail(t, db, "enterprise@example.com")
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler := auth.RequireEnterprise(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// Helper functions for test setup
func createTestUser(t *testing.T, db *sql.DB, username, email, accountType string, emailVerified bool) {
	hashedPassword, _ := auth.HashPassword("testpassword123")
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, hashed_password, account_type, account_type_new, email_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, username+"-id", username, email, hashedPassword, accountType, accountType, emailVerified, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
}

func createTestOrganization(t *testing.T, db *sql.DB, name, adminEmail string) string {
	userID := getUserIDByEmail(t, db, adminEmail)
	orgID := "test-org-id"
	_, err := db.Exec(`
		INSERT INTO organizations (id, name, admin_user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, orgID, name, userID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}
	return orgID
}

func addUserToOrganization(t *testing.T, db *sql.DB, orgID, userEmail, role string) {
	userID := getUserIDByEmail(t, db, userEmail)
	_, err := db.Exec(`
		INSERT INTO organization_members (id, organization_id, user_id, role, status, joined_at)
		VALUES (?, ?, ?, ?, 'active', ?)
	`, "test-member-id", orgID, userID, role, time.Now())
	if err != nil {
		t.Fatalf("Failed to add user to organization: %v", err)
	}
}

func createTestSession(t *testing.T, db *sql.DB, email string) string {
	userID := getUserIDByEmail(t, db, email)
	token := "test-token-" + email
	tokenHash := auth.HashToken(token)
	sessionID := "test-session-" + email
	_, err := db.Exec(`
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, userID, tokenHash, time.Now().Add(24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}
	return token
}

func getUserIDByEmail(t *testing.T, db *sql.DB, email string) string {
	var userID string
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}
	return userID
}
