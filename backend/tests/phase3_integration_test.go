package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"securemail-backend/auth"
)

// Test database setup for Phase 3
func setupPhase3TestDB(t *testing.T) *sql.DB {
	// Create in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Set the global DB variable
	auth.DB = db

	// Run all migrations
	migrations := []string{
		// Phase 1 migrations
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			email TEXT NOT NULL,
			hashed_password TEXT NOT NULL,
			account_type TEXT NOT NULL DEFAULT 'free',
			account_type_new TEXT DEFAULT 'free',
			account_status TEXT NOT NULL DEFAULT 'pending_verification',
			email_verified BOOLEAN DEFAULT FALSE,
			verification_code TEXT NULL,
			verification_code_expires_at TIMESTAMP NULL,
			totp_secret TEXT NULL,
			mfa_enabled BOOLEAN DEFAULT FALSE,
			backup_codes_hashed JSON NULL,
			organization_id TEXT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// Phase 2 migrations
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_verification_code ON users(verification_code)`,
		`CREATE INDEX IF NOT EXISTS idx_users_mfa_enabled ON users(mfa_enabled)`,
		// Phase 3 migrations
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			stripe_customer_id TEXT NULL,
			stripe_subscription_id TEXT NULL,
			status TEXT NOT NULL DEFAULT 'inactive',
			plan TEXT NOT NULL DEFAULT 'free',
			start_date TIMESTAMP NULL,
			end_date TIMESTAMP NULL,
			trial_end_date TIMESTAMP NULL,
			cancel_at_period_end BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS domains (
			id TEXT PRIMARY KEY,
			user_id TEXT NULL,
			organization_id TEXT NULL,
			domain_name TEXT NOT NULL,
			verified BOOLEAN DEFAULT FALSE,
			verification_code TEXT NULL,
			verification_method TEXT NULL,
			dns_record_value TEXT NULL,
			verification_attempts INTEGER DEFAULT 0,
			last_verification_attempt TIMESTAMP NULL,
			is_primary BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			admin_user_id TEXT NOT NULL,
			domain TEXT NULL,
			max_users INTEGER DEFAULT 10,
			subscription_id TEXT NULL,
			settings JSON NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS organization_members (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			invited_by TEXT NULL,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status TEXT NOT NULL DEFAULT 'active'
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_domains_domain_name ON domains(domain_name)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_organizations_admin ON organizations(admin_user_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_org_members_org_user ON organization_members(organization_id, user_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			t.Fatalf("Failed to run migration: %v", err)
		}
	}

	return db
}

// Test subscription upgrade flow
func TestSubscriptionUpgradeFlow(t *testing.T) {
	db := setupPhase3TestDB(t)
	defer db.Close()

	// Create a test user
	userID := "test-user-id"
	hashedPassword, _ := auth.HashPassword("testpassword")
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, hashed_password, account_type, account_type_new, email_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'free', 'free', TRUE, ?, ?)
	`, userID, "testuser", "test@example.com", hashedPassword, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a mock session for the user
	sessionID := "test-session-id"
	tokenHash := auth.HashToken("test-token")
	_, err = db.Exec(`
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, userID, tokenHash, time.Now().Add(24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Upgrade to Premium", func(t *testing.T) {
		// Create upgrade request
		reqBody := map[string]string{"plan": "premium"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/upgrade-account", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.UpgradeAccountHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify subscription was created
		var subscriptionCount int
		err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = ? AND plan = 'premium'", userID).Scan(&subscriptionCount)
		if err != nil {
			t.Fatalf("Failed to check subscription: %v", err)
		}
		if subscriptionCount != 1 {
			t.Errorf("Expected 1 subscription, got %d", subscriptionCount)
		}

		// Verify user account type was updated
		var accountType string
		err = db.QueryRow("SELECT account_type_new FROM users WHERE id = ?", userID).Scan(&accountType)
		if err != nil {
			t.Fatalf("Failed to check account type: %v", err)
		}
		if accountType != "premium" {
			t.Errorf("Expected account type 'premium', got '%s'", accountType)
		}
	})

	t.Run("Upgrade to Enterprise", func(t *testing.T) {
		// Create upgrade request
		reqBody := map[string]string{"plan": "enterprise"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/upgrade-account", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.UpgradeAccountHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify subscription was updated
		var plan string
		err = db.QueryRow("SELECT plan FROM subscriptions WHERE user_id = ?", userID).Scan(&plan)
		if err != nil {
			t.Fatalf("Failed to check subscription plan: %v", err)
		}
		if plan != "enterprise" {
			t.Errorf("Expected plan 'enterprise', got '%s'", plan)
		}

		// Verify user account type was updated
		var accountType string
		err = db.QueryRow("SELECT account_type_new FROM users WHERE id = ?", userID).Scan(&accountType)
		if err != nil {
			t.Fatalf("Failed to check account type: %v", err)
		}
		if accountType != "enterprise" {
			t.Errorf("Expected account type 'enterprise', got '%s'", accountType)
		}
	})
}

// Test domain management flow
func TestDomainManagementFlow(t *testing.T) {
	db := setupPhase3TestDB(t)
	defer db.Close()

	// Create a premium user
	userID := "test-user-id"
	hashedPassword, _ := auth.HashPassword("testpassword")
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, hashed_password, account_type, account_type_new, email_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'premium', 'premium', TRUE, ?, ?)
	`, userID, "testuser", "test@example.com", hashedPassword, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a mock session for the user
	sessionID := "test-session-id"
	tokenHash := auth.HashToken("test-token")
	_, err = db.Exec(`
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, userID, tokenHash, time.Now().Add(24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Add Domain", func(t *testing.T) {
		// Create add domain request
		reqBody := map[string]string{"domain_name": "example.com"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/domain/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.AddDomainHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify domain was added
		var domainCount int
		err = db.QueryRow("SELECT COUNT(*) FROM domains WHERE user_id = ? AND domain_name = 'example.com'", userID).Scan(&domainCount)
		if err != nil {
			t.Fatalf("Failed to check domain: %v", err)
		}
		if domainCount != 1 {
			t.Errorf("Expected 1 domain, got %d", domainCount)
		}
	})

	t.Run("Verify Domain", func(t *testing.T) {
		// Create verify domain request
		reqBody := map[string]string{"domain_name": "example.com"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/domain/verify", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.VerifyDomainHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify domain was marked as verified
		var verified bool
		err = db.QueryRow("SELECT verified FROM domains WHERE user_id = ? AND domain_name = 'example.com'", userID).Scan(&verified)
		if err != nil {
			t.Fatalf("Failed to check domain verification: %v", err)
		}
		if !verified {
			t.Errorf("Expected domain to be verified")
		}
	})

	t.Run("Remove Domain", func(t *testing.T) {
		// Create remove domain request
		reqBody := map[string]string{"domain_name": "example.com"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/domain/remove", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.RemoveDomainHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify domain was removed
		var domainCount int
		err = db.QueryRow("SELECT COUNT(*) FROM domains WHERE user_id = ? AND domain_name = 'example.com'", userID).Scan(&domainCount)
		if err != nil {
			t.Fatalf("Failed to check domain: %v", err)
		}
		if domainCount != 0 {
			t.Errorf("Expected 0 domains, got %d", domainCount)
		}
	})
}

// Test organization management flow
func TestOrganizationManagementFlow(t *testing.T) {
	db := setupPhase3TestDB(t)
	defer db.Close()

	// Create an enterprise user
	userID := "test-admin-id"
	hashedPassword, _ := auth.HashPassword("testpassword")
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, hashed_password, account_type, account_type_new, email_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'enterprise', 'enterprise', TRUE, ?, ?)
	`, userID, "admin", "admin@example.com", hashedPassword, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test admin user: %v", err)
	}

	// Create a regular user to add to organization
	memberUserID := "test-member-id"
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, hashed_password, account_type, account_type_new, email_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'free', 'free', TRUE, ?, ?)
	`, memberUserID, "member", "member@example.com", hashedPassword, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test member user: %v", err)
	}

	// Create a mock session for the admin user
	sessionID := "test-session-id"
	tokenHash := auth.HashToken("test-token")
	_, err = db.Exec(`
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, userID, tokenHash, time.Now().Add(24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Create Organization", func(t *testing.T) {
		// Create organization request
		reqBody := map[string]string{"name": "Test Organization"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/org/create", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.CreateOrganizationHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify organization was created
		var orgCount int
		err = db.QueryRow("SELECT COUNT(*) FROM organizations WHERE admin_user_id = ?", userID).Scan(&orgCount)
		if err != nil {
			t.Fatalf("Failed to check organization: %v", err)
		}
		if orgCount != 1 {
			t.Errorf("Expected 1 organization, got %d", orgCount)
		}
	})

	t.Run("Add User to Organization", func(t *testing.T) {
		// Get organization ID
		var orgID string
		err = db.QueryRow("SELECT id FROM organizations WHERE admin_user_id = ?", userID).Scan(&orgID)
		if err != nil {
			t.Fatalf("Failed to get organization ID: %v", err)
		}

		// Create add user request
		reqBody := map[string]string{
			"organization_id": orgID,
			"user_email":      "member@example.com",
			"role":            "member",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/org/add-user", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.AddUserToOrgHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify user was added to organization
		var memberCount int
		err = db.QueryRow("SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ?", orgID, memberUserID).Scan(&memberCount)
		if err != nil {
			t.Fatalf("Failed to check organization member: %v", err)
		}
		if memberCount != 1 {
			t.Errorf("Expected 1 organization member, got %d", memberCount)
		}
	})

	t.Run("List Organization Users", func(t *testing.T) {
		// Get organization ID
		var orgID string
		err = db.QueryRow("SELECT id FROM organizations WHERE admin_user_id = ?", userID).Scan(&orgID)
		if err != nil {
			t.Fatalf("Failed to get organization ID: %v", err)
		}

		// Create list users request
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/org/list-users?organization_id=%s", orgID), nil)
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.ListOrgUsersHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Parse response
		var response struct {
			Success bool `json:"success"`
			Users   []struct {
				ID       string `json:"id"`
				Email    string `json:"email"`
				Role     string `json:"role"`
				Status   string `json:"status"`
				JoinedAt string `json:"joined_at"`
			} `json:"users"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify response contains users
		if len(response.Users) != 2 { // admin + member
			t.Errorf("Expected 2 users, got %d", len(response.Users))
		}
	})
}

// Test account downgrade flow
func TestAccountDowngradeFlow(t *testing.T) {
	db := setupPhase3TestDB(t)
	defer db.Close()

	// Create a premium user with subscription
	userID := "test-user-id"
	hashedPassword, _ := auth.HashPassword("testpassword")
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, hashed_password, account_type, account_type_new, email_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'premium', 'premium', TRUE, ?, ?)
	`, userID, "testuser", "test@example.com", hashedPassword, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a subscription
	subscriptionID := "test-subscription-id"
	_, err = db.Exec(`
		INSERT INTO subscriptions (id, user_id, status, plan, created_at, updated_at)
		VALUES (?, ?, 'active', 'premium', ?, ?)
	`, subscriptionID, userID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test subscription: %v", err)
	}

	// Create a mock session for the user
	sessionID := "test-session-id"
	tokenHash := auth.HashToken("test-token")
	_, err = db.Exec(`
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, userID, tokenHash, time.Now().Add(24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("Downgrade to Free", func(t *testing.T) {
		// Create downgrade request
		reqBody := map[string]string{"plan": "free"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/downgrade-account", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		auth.DowngradeAccountHandler(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify subscription was canceled
		var status string
		err = db.QueryRow("SELECT status FROM subscriptions WHERE user_id = ?", userID).Scan(&status)
		if err != nil {
			t.Fatalf("Failed to check subscription status: %v", err)
		}
		if status != "canceled" {
			t.Errorf("Expected subscription status 'canceled', got '%s'", status)
		}

		// Verify user account type was updated
		var accountType string
		err = db.QueryRow("SELECT account_type_new FROM users WHERE id = ?", userID).Scan(&accountType)
		if err != nil {
			t.Fatalf("Failed to check account type: %v", err)
		}
		if accountType != "free" {
			t.Errorf("Expected account type 'free', got '%s'", accountType)
		}
	})
}
