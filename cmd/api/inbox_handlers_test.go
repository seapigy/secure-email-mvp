package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDatabase creates a test database with both emails and users schemas
func setupTestDatabase(t *testing.T) *sql.DB {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-inbox-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Apply emails schema
	emailsSchema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read emails schema: %v", err)
	}

	_, err = db.Exec(string(emailsSchema))
	if err != nil {
		t.Fatalf("Failed to apply emails schema: %v", err)
	}

	// Apply users schema
	usersSchema, err := os.ReadFile("../../schema/users.sql")
	if err != nil {
		t.Fatalf("Failed to read users schema: %v", err)
	}

	_, err = db.Exec(string(usersSchema))
	if err != nil {
		t.Fatalf("Failed to apply users schema: %v", err)
	}

	// Apply inbox indexes
	indexes, err := os.ReadFile("../../schema/migrate_add_inbox_indexes.sql")
	if err != nil {
		t.Fatalf("Failed to read indexes: %v", err)
	}

	_, err = db.Exec(string(indexes))
	if err != nil {
		t.Fatalf("Failed to apply indexes: %v", err)
	}

	// Apply inbox_messages table
	inboxMessages, err := os.ReadFile("../../schema/migrate_add_inbox_messages_table.sql")
	if err != nil {
		t.Fatalf("Failed to read inbox_messages schema: %v", err)
	}

	_, err = db.Exec(string(inboxMessages))
	if err != nil {
		t.Fatalf("Failed to apply inbox_messages schema: %v", err)
	}

	return db
}

// createInboxMessageForTest creates an inbox message for testing
func createInboxMessageForTest(db *sql.DB, userID, emailID string) error {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO inbox_messages (id, user_id, email_id, is_read, is_deleted, created_at, updated_at)
		VALUES (lower(hex(randomblob(16))), ?, ?, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID, emailID,
	)
	return err
}

// TestListInboxHandler tests the inbox list functionality
func TestListInboxHandler(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create test users
	testUserID := "test-user-inbox-list"
	testUserEmail := "test@example.com"
	otherUserID := "other-user-inbox-list"
	otherUserEmail := "other@example.com"

	// Insert test users
	_, err := db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", testUserID, testUserEmail, "hash", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", otherUserID, otherUserEmail, "hash", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Failed to insert other user: %v", err)
	}

	// Insert test emails - some sent TO test user, some sent BY test user
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
		) VALUES 
		('email-to-test-1', ?, ?, 'Email to test user 1', 'blob-1', 'key-1', 'nonce-1', 'tag-1', 'gzip', 'hash-1', ?),
		('email-to-test-2', ?, ?, 'Email to test user 2', 'blob-2', 'key-2', 'nonce-2', 'tag-2', 'gzip', 'hash-2', ?),
		('email-from-test-1', ?, ?, 'Email from test user 1', 'blob-3', 'key-3', 'nonce-3', 'tag-3', 'gzip', 'hash-3', ?),
		('email-to-other-1', ?, ?, 'Email to other user 1', 'blob-4', 'key-4', 'nonce-4', 'tag-4', 'gzip', 'hash-4', ?)`,
		otherUserID, testUserEmail, time.Now(),
		otherUserID, testUserEmail, time.Now(),
		testUserID, otherUserEmail, time.Now(),
		testUserID, otherUserEmail, time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to insert test emails: %v", err)
	}

	// Create inbox messages for emails sent TO test user
	err = createInboxMessageForTest(db, testUserID, "email-to-test-1")
	if err != nil {
		t.Fatalf("Failed to create inbox message for email-to-test-1: %v", err)
	}

	err = createInboxMessageForTest(db, testUserID, "email-to-test-2")
	if err != nil {
		t.Fatalf("Failed to create inbox message for email-to-test-2: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   testUserEmail,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Test 1: Valid JWT token should return user's inbox
	t.Run("ValidJWTToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/inbox/list", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.listInboxHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())
			return
		}

		var response ListInboxResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", response.Status)
		}

		// Should only see emails sent TO the test user (2 emails)
		if len(response.Emails) != 2 {
			t.Errorf("Expected 2 emails in inbox, got %d", len(response.Emails))
		}

		// Verify emails are the ones sent TO the test user
		expectedEmailIDs := map[string]bool{
			"email-to-test-1": true,
			"email-to-test-2": true,
		}

		for _, email := range response.Emails {
			if !expectedEmailIDs[email.EmailID] {
				t.Errorf("Unexpected email ID in inbox: %s", email.EmailID)
			}
		}
	})

	// Test 2: Missing authorization header should return 401
	t.Run("MissingAuthorization", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/inbox/list", nil)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.listInboxHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	// Test 3: Invalid JWT token should return 401
	t.Run("InvalidJWTToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/inbox/list", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.listInboxHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	// Test 4: User with no emails should get empty inbox
	t.Run("EmptyInbox", func(t *testing.T) {
		// Generate JWT token for user with no emails
		emptyUserID := "empty-user"
		emptyUserEmail := "empty@example.com"

		_, err := db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", emptyUserID, emptyUserEmail, "hash", "JBSWY3DPEHPK3PXP")
		if err != nil {
			t.Fatalf("Failed to insert empty user: %v", err)
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": emptyUserID,
			"email":   emptyUserEmail,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
			"iat":     time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
		if err != nil {
			t.Fatalf("Failed to generate JWT token: %v", err)
		}

		req := httptest.NewRequest("GET", "/api/inbox/list", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.listInboxHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
			return
		}

		var response ListInboxResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Emails) != 0 {
			t.Errorf("Expected 0 emails in empty inbox, got %d", len(response.Emails))
		}
	})
}

// TestGetInboxEmailHandler tests the inbox email retrieval functionality
func TestGetInboxEmailHandler(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create test users
	testUserID := "test-user-inbox-get"
	testUserEmail := "test-get@example.com"
	otherUserID := "other-user-inbox-get"
	otherUserEmail := "other-get@example.com"

	// Insert test users
	_, err := db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", testUserID, testUserEmail, "hash", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", otherUserID, otherUserEmail, "hash", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Failed to insert other user: %v", err)
	}

	// Insert test email sent TO test user
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
		) VALUES 
		('email-to-test-get', ?, ?, 'Email to test user for get', 'blob-1', 'key-1', 'nonce-1', 'tag-1', 'gzip', 'hash-1', ?)`,
		otherUserID, testUserEmail, time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create inbox message for the test email
	err = createInboxMessageForTest(db, testUserID, "email-to-test-get")
	if err != nil {
		t.Fatalf("Failed to create inbox message for email-to-test-get: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   testUserEmail,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Test 1: Valid JWT token should return user's email
	t.Run("ValidJWTToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/inbox/email-to-test-get", nil)
		// Set up the URL variables for mux.Vars to work
		req = mux.SetURLVars(req, map[string]string{"id": "email-to-test-get"})
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.getInboxEmailHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())
			return
		}

		var response GetInboxEmailResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", response.Status)
		}

		if response.Email.EmailID != "email-to-test-get" {
			t.Errorf("Expected email ID 'email-to-test-get', got '%s'", response.Email.EmailID)
		}
	})

	// Test 2: User cannot access another user's email
	t.Run("CannotAccessOtherUserEmail", func(t *testing.T) {
		// Insert email sent TO other user
		_, err := db.Exec(`
			INSERT INTO emails (
				email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
				encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
			) VALUES 
			('email-to-other-get', ?, ?, 'Email to other user for get', 'blob-2', 'key-2', 'nonce-2', 'tag-2', 'gzip', 'hash-2', ?)`,
			testUserID, otherUserEmail, time.Now(),
		)
		if err != nil {
			t.Fatalf("Failed to insert other user email: %v", err)
		}

		// Create inbox message for the other user's email
		err = createInboxMessageForTest(db, otherUserID, "email-to-other-get")
		if err != nil {
			t.Fatalf("Failed to create inbox message for email-to-other-get: %v", err)
		}

		req := httptest.NewRequest("GET", "/api/inbox/email-to-other-get", nil)
		// Set up the URL variables for mux.Vars to work
		req = mux.SetURLVars(req, map[string]string{"id": "email-to-other-get"})
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.getInboxEmailHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	// Test 3: Non-existent email should return 404
	t.Run("NonExistentEmail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/inbox/non-existent-email", nil)
		// Set up the URL variables for mux.Vars to work
		req = mux.SetURLVars(req, map[string]string{"id": "non-existent-email"})
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.getInboxEmailHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

// TestDeleteInboxEmailHandler tests the inbox email deletion functionality
func TestDeleteInboxEmailHandler(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create test users
	testUserID := "test-user-inbox-delete"
	testUserEmail := "test-delete@example.com"
	otherUserID := "other-user-inbox-delete"
	otherUserEmail := "other-delete@example.com"

	// Insert test users
	_, err := db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", testUserID, testUserEmail, "hash", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", otherUserID, otherUserEmail, "hash", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Failed to insert other user: %v", err)
	}

	// Insert test email sent TO test user
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
		) VALUES 
		('email-to-test-delete', ?, ?, 'Email to test user for delete', 'blob-1', 'key-1', 'nonce-1', 'tag-1', 'gzip', 'hash-1', ?)`,
		otherUserID, testUserEmail, time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create inbox message for the test email
	err = createInboxMessageForTest(db, testUserID, "email-to-test-delete")
	if err != nil {
		t.Fatalf("Failed to create inbox message for email-to-test-delete: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   testUserEmail,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Test 1: Valid JWT token should delete user's email
	t.Run("ValidJWTToken", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/inbox/email-to-test-delete", nil)
		// Set up the URL variables for mux.Vars to work
		req = mux.SetURLVars(req, map[string]string{"id": "email-to-test-delete"})
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.deleteInboxEmailHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())
			return
		}

		var response DeleteInboxEmailResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", response.Status)
		}

		// Verify email was soft deleted from inbox_messages
		var isDeleted bool
		err := db.QueryRow("SELECT is_deleted FROM inbox_messages WHERE user_id = ? AND email_id = ?", testUserID, "email-to-test-delete").Scan(&isDeleted)
		if err != nil {
			t.Fatalf("Failed to check email deletion status: %v", err)
		}

		if !isDeleted {
			t.Errorf("Expected email to be soft deleted (is_deleted = true), got %v", isDeleted)
		}
	})

	// Test 2: User cannot delete another user's email
	t.Run("CannotDeleteOtherUserEmail", func(t *testing.T) {
		// Insert email sent TO other user
		_, err := db.Exec(`
			INSERT INTO emails (
				email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
				encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
			) VALUES 
			('email-to-other-delete', ?, ?, 'Email to other user for delete', 'blob-2', 'key-2', 'nonce-2', 'tag-2', 'gzip', 'hash-2', ?)`,
			testUserID, otherUserEmail, time.Now(),
		)
		if err != nil {
			t.Fatalf("Failed to insert other user email: %v", err)
		}

		// Create inbox message for the other user's email
		err = createInboxMessageForTest(db, otherUserID, "email-to-other-delete")
		if err != nil {
			t.Fatalf("Failed to create inbox message for email-to-other-delete: %v", err)
		}

		req := httptest.NewRequest("DELETE", "/api/inbox/email-to-other-delete", nil)
		// Set up the URL variables for mux.Vars to work
		req = mux.SetURLVars(req, map[string]string{"id": "email-to-other-delete"})
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.deleteInboxEmailHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}

		// Verify email was NOT deleted from inbox_messages
		var isDeleted bool
		err = db.QueryRow("SELECT is_deleted FROM inbox_messages WHERE user_id = ? AND email_id = ?", otherUserID, "email-to-other-delete").Scan(&isDeleted)
		if err != nil {
			t.Fatalf("Failed to check email deletion status: %v", err)
		}

		if isDeleted {
			t.Errorf("Expected email to NOT be deleted (is_deleted = false), got %v", isDeleted)
		}
	})

	// Test 3: Non-existent email should return 404
	t.Run("NonExistentEmail", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/inbox/non-existent-email", nil)
		// Set up the URL variables for mux.Vars to work
		req = mux.SetURLVars(req, map[string]string{"id": "non-existent-email"})
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.deleteInboxEmailHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

// TestInboxUserIsolation tests that users cannot access each other's inboxes
func TestInboxUserIsolation(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Create test users
	user1ID := "user1-isolation"
	user1Email := "user1@example.com"
	user2ID := "user2-isolation"
	user2Email := "user2@example.com"

	// Insert test users
	_, err := db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", user1ID, user1Email, "hash", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Failed to insert user1: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)", user2ID, user2Email, "hash", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Failed to insert user2: %v", err)
	}

	// Insert test emails - each user has emails sent TO them
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
		) VALUES 
		('email-to-user1', ?, ?, 'Email to user1', 'blob-1', 'key-1', 'nonce-1', 'tag-1', 'gzip', 'hash-1', ?),
		('email-to-user2', ?, ?, 'Email to user2', 'blob-2', 'key-2', 'nonce-2', 'tag-2', 'gzip', 'hash-2', ?)`,
		user2ID, user1Email, time.Now(),
		user1ID, user2Email, time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to insert test emails: %v", err)
	}

	// Create inbox messages for each user
	err = createInboxMessageForTest(db, user1ID, "email-to-user1")
	if err != nil {
		t.Fatalf("Failed to create inbox message for email-to-user1: %v", err)
	}

	err = createInboxMessageForTest(db, user2ID, "email-to-user2")
	if err != nil {
		t.Fatalf("Failed to create inbox message for email-to-user2: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Test: User1 cannot access User2's inbox
	t.Run("User1CannotAccessUser2Inbox", func(t *testing.T) {
		// Generate JWT token for user1
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user1ID,
			"email":   user1Email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
			"iat":     time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
		if err != nil {
			t.Fatalf("Failed to generate JWT token: %v", err)
		}

		// Try to get user2's email
		req := httptest.NewRequest("GET", "/api/inbox/email-to-user2", nil)
		// Set up the URL variables for mux.Vars to work
		req = mux.SetURLVars(req, map[string]string{"id": "email-to-user2"})
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.getInboxEmailHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	// Test: User2 cannot access User1's inbox
	t.Run("User2CannotAccessUser1Inbox", func(t *testing.T) {
		// Generate JWT token for user2
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user2ID,
			"email":   user2Email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
			"iat":     time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
		if err != nil {
			t.Fatalf("Failed to generate JWT token: %v", err)
		}

		// Try to get user1's email
		req := httptest.NewRequest("GET", "/api/inbox/email-to-user1", nil)
		// Set up the URL variables for mux.Vars to work
		req = mux.SetURLVars(req, map[string]string{"id": "email-to-user1"})
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()

		handler := jwtMiddleware(http.HandlerFunc(srv.getInboxEmailHandler))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}
