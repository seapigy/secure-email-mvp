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


// Test complete Phase 3 signup/login/inbox flow
func TestCompletePhase3Flow(t *testing.T) {
	db := setupCompletePhase3TestDB(t)
	defer db.Close()

	t.Run("Complete Free User Flow", func(t *testing.T) {
		// 1. Signup
		reqBody := map[string]string{
			"username":      "freeuser",
			"email":         "free@example.com",
			"password":      "TestPassword123!",
			"fallback_email": "freebackup@example.com",
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

		// Verify user was created
		var userID string
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", "free@example.com").Scan(&userID)
		if err != nil {
			t.Fatalf("Failed to get user ID: %v", err)
		}

		// Verify default folders were created
		var folderCount int
		err = db.QueryRow("SELECT COUNT(*) FROM mailbox_folders WHERE user_id = ?", userID).Scan(&folderCount)
		if err != nil {
			t.Fatalf("Failed to count folders: %v", err)
		}
		if folderCount != 4 { // Inbox, Sent, Trash, Drafts
			t.Errorf("Expected 4 default folders, got %d", folderCount)
		}

		// Verify welcome message was created
		var messageCount int
		err = db.QueryRow("SELECT COUNT(*) FROM email_messages WHERE user_id = ?", userID).Scan(&messageCount)
		if err != nil {
			t.Fatalf("Failed to count messages: %v", err)
		}
		if messageCount != 1 {
			t.Errorf("Expected 1 welcome message, got %d", messageCount)
		}

		// Verify no subscription for free user
		var subscriptionCount int
		err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = ?", userID).Scan(&subscriptionCount)
		if err != nil {
			t.Fatalf("Failed to count subscriptions: %v", err)
		}
		if subscriptionCount != 0 {
			t.Errorf("Expected 0 subscriptions for free user, got %d", subscriptionCount)
		}
	})

	t.Run("Complete Premium User Flow", func(t *testing.T) {
		// 1. Signup
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

		// Verify user was created
		var userID string
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", "premium@example.com").Scan(&userID)
		if err != nil {
			t.Fatalf("Failed to get user ID: %v", err)
		}

		// Verify placeholder subscription was created
		var subscriptionCount int
		err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = ?", userID).Scan(&subscriptionCount)
		if err != nil {
			t.Fatalf("Failed to count subscriptions: %v", err)
		}
		if subscriptionCount != 1 {
			t.Errorf("Expected 1 subscription for premium user, got %d", subscriptionCount)
		}

		// Verify subscription details
		var plan string
		var status string
		err = db.QueryRow("SELECT plan, status FROM subscriptions WHERE user_id = ?", userID).Scan(&plan, &status)
		if err != nil {
			t.Fatalf("Failed to get subscription details: %v", err)
		}
		if plan != "premium" {
			t.Errorf("Expected plan 'premium', got '%s'", plan)
		}
		if status != "active" {
			t.Errorf("Expected status 'active', got '%s'", status)
		}

		// 2. Mark email as verified for testing
		_, err = db.Exec("UPDATE users SET email_verified = TRUE WHERE email = ?", "premium@example.com")
		if err != nil {
			t.Fatalf("Failed to verify email: %v", err)
		}

		// 3. Login
		loginReqBody := map[string]string{
			"email":    "premium@example.com",
			"password": "testpassword123",
		}
		loginJsonBody, _ := json.Marshal(loginReqBody)
		loginReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginJsonBody))
		loginReq.Header.Set("Content-Type", "application/json")

		loginW := httptest.NewRecorder()
		auth.LoginHandler(loginW, loginReq)

		if loginW.Code != http.StatusOK {
			t.Errorf("Expected login status 200, got %d", loginW.Code)
		}

		// Parse login response
		var loginResponse struct {
			Token       string `json:"token"`
			ExpiresAt   string `json:"expires_at"`
			AccountType string `json:"account_type"`
		}
		err = json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
		if err != nil {
			t.Fatalf("Failed to parse login response: %v", err)
		}

		if loginResponse.AccountType != "premium" {
			t.Errorf("Expected account type 'premium', got '%s'", loginResponse.AccountType)
		}
	})

	t.Run("Complete Enterprise User Flow", func(t *testing.T) {
		// 1. Signup
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

		// Verify user was created
		var userID string
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", "enterprise@example.com").Scan(&userID)
		if err != nil {
			t.Fatalf("Failed to get user ID: %v", err)
		}

		// Verify placeholder subscription was created
		var subscriptionCount int
		err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = ?", userID).Scan(&subscriptionCount)
		if err != nil {
			t.Fatalf("Failed to count subscriptions: %v", err)
		}
		if subscriptionCount != 1 {
			t.Errorf("Expected 1 subscription for enterprise user, got %d", subscriptionCount)
		}

		// 2. Create organization
		orgReqBody := map[string]string{
			"name": "Test Enterprise Org",
		}
		orgJsonBody, _ := json.Marshal(orgReqBody)
		orgReq := httptest.NewRequest("POST", "/api/org/create", bytes.NewBuffer(orgJsonBody))
		orgReq.Header.Set("Content-Type", "application/json")
		orgReq.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(orgReq.Context(), userID)
		orgReq = orgReq.WithContext(ctx)

		orgW := httptest.NewRecorder()
		auth.CreateOrganizationHandler(orgW, orgReq)

		if orgW.Code != http.StatusOK && orgW.Code != http.StatusCreated {
			t.Errorf("Expected org creation status 200 or 201, got %d", orgW.Code)
		}

		// Verify organization was created
		var orgCount int
		err = db.QueryRow("SELECT COUNT(*) FROM organizations WHERE admin_user_id = ?", userID).Scan(&orgCount)
		if err != nil {
			t.Fatalf("Failed to count organizations: %v", err)
		}
		if orgCount != 1 {
			t.Errorf("Expected 1 organization, got %d", orgCount)
		}
	})
}

// Test inbox functionality
func TestInboxFunctionality(t *testing.T) {
	db := setupCompletePhase3TestDB(t)
	defer db.Close()

	// Create test user
	userID := createTestUserWithInbox(t, db, "inboxtest", "inbox@example.com", "free", true)

	t.Run("Get Inbox Folders", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/inbox/folders", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		auth.InboxFoldersHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Parse response
		var response struct {
			Folders []map[string]interface{} `json:"folders"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if len(response.Folders) != 4 {
			t.Errorf("Expected 4 folders, got %d", len(response.Folders))
		}

		// Check folder types
		folderTypes := make(map[string]bool)
		for _, folder := range response.Folders {
			if folderType, ok := folder["folder_type"].(string); ok {
				folderTypes[folderType] = true
			}
		}

		expectedTypes := []string{"inbox", "sent", "trash", "drafts"}
		for _, expectedType := range expectedTypes {
			if !folderTypes[expectedType] {
				t.Errorf("Expected folder type '%s' not found", expectedType)
			}
		}
	})

	t.Run("Get Inbox Messages", func(t *testing.T) {
		// Get inbox folder ID
		var inboxFolderID string
		err := db.QueryRow("SELECT id FROM mailbox_folders WHERE user_id = ? AND folder_type = 'inbox'", userID).Scan(&inboxFolderID)
		if err != nil {
			t.Fatalf("Failed to get inbox folder ID: %v", err)
		}

		req := httptest.NewRequest("GET", "/api/inbox/messages?folder_id="+inboxFolderID, nil)
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		auth.InboxMessagesHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Parse response
		var response struct {
			Messages []map[string]interface{} `json:"messages"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if len(response.Messages) != 1 {
			t.Errorf("Expected 1 welcome message, got %d", len(response.Messages))
		}

		// Check welcome message subject
		if len(response.Messages) > 0 {
			subject := response.Messages[0]["subject"].(string)
			if subject != "Welcome to Secure Email!" {
				t.Errorf("Expected welcome message subject, got '%s'", subject)
			}
		}
	})
}

// Test trial warning system
func TestTrialWarningSystem(t *testing.T) {
	db := setupCompletePhase3TestDB(t)
	defer db.Close()

	// Create premium user with expiring trial
	userID := createTestUserWithInbox(t, db, "trialuser", "trial@example.com", "premium", true)
	
	// Create subscription with trial expiring in 3 days
	_, err := db.Exec(`
		INSERT INTO subscriptions (id, user_id, status, plan, start_date, end_date, created_at, updated_at)
		VALUES (?, ?, 'active', 'premium', ?, ?, ?, ?)
	`, "test-subscription-id", userID, time.Now(), time.Now().AddDate(0, 0, 3), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create trial subscription: %v", err)
	}

	t.Run("Check Trial Expiration Warning", func(t *testing.T) {
		warning, err := auth.CheckTrialExpiration(userID)
		if err != nil {
			t.Fatalf("Failed to check trial expiration: %v", err)
		}

		if warning == nil {
			t.Errorf("Expected trial warning, got nil")
			return
		}

		if warning.Plan != "premium" {
			t.Errorf("Expected plan 'premium', got '%s'", warning.Plan)
		}

		if warning.DaysRemaining < 2 || warning.DaysRemaining > 4 {
			t.Errorf("Expected 2-4 days remaining, got %d", warning.DaysRemaining)
		}

		if warning.WarningLevel != "warning" {
			t.Errorf("Expected warning level 'warning', got '%s'", warning.WarningLevel)
		}
	})

	t.Run("Trial Warning Handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/trial/warning", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		// Add user context
		ctx := auth.ContextWithUserID(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		auth.TrialWarningHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Parse response
		var response struct {
			HasWarning bool                    `json:"has_warning"`
			Warning    *auth.TrialWarning      `json:"warning,omitempty"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if !response.HasWarning {
			t.Errorf("Expected has_warning to be true")
		}

		if response.Warning == nil {
			t.Errorf("Expected warning object, got nil")
		}
	})
}

// Test analytics system
func TestAnalyticsSystem(t *testing.T) {
	db := setupCompletePhase3TestDB(t)
	defer db.Close()

	userID := "test-analytics-user"

	t.Run("Log Analytics Event", func(t *testing.T) {
		// Log a test event
		auth.LogAnalyticsEvent(userID, "test_event", map[string]interface{}{
			"account_type": "premium",
			"success":      true,
		})

		// Verify event was logged
		var eventCount int
		err := db.QueryRow("SELECT COUNT(*) FROM analytics_events WHERE event_type = 'test_event'").Scan(&eventCount)
		if err != nil {
			t.Fatalf("Failed to count analytics events: %v", err)
		}
		if eventCount != 1 {
			t.Errorf("Expected 1 analytics event, got %d", eventCount)
		}

		// Verify user hash was created (privacy protection)
		var userHash string
		err = db.QueryRow("SELECT user_hash FROM analytics_events WHERE event_type = 'test_event'").Scan(&userHash)
		if err != nil {
			t.Fatalf("Failed to get user hash: %v", err)
		}
		if userHash == userID {
			t.Errorf("User hash should not be the same as user ID for privacy")
		}
		if len(userHash) != 64 { // SHA256 hex length
			t.Errorf("Expected 64-character hash, got %d", len(userHash))
		}
	})
}

// Helper functions - using shared setupCompletePhase3TestDB function from test_setup.go

func createTestUserWithInbox(t *testing.T, db *sql.DB, username, email, accountType string, emailVerified bool) string {
	hashedPassword, _ := auth.HashPassword("testpassword123")
	userID := username + "-id"
	
	_, err := db.Exec(`
		INSERT INTO users (id, username, email, hashed_password, account_type, account_type_new, email_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, username, email, hashedPassword, accountType, accountType, emailVerified, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create default inbox
	err = auth.CreateDefaultInbox(userID)
	if err != nil {
		t.Fatalf("Failed to create default inbox: %v", err)
	}

	return userID
}
