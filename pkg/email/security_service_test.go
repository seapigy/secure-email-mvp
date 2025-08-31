package email

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupSecurityTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create emails table with all security fields
	_, err = db.Exec(`
		CREATE TABLE emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT,
			recipient TEXT,
			subject TEXT,
			body TEXT,
			time_lock BOOLEAN DEFAULT FALSE,
			unlock_after TEXT,
			expires_at TEXT,
			burn_after_read BOOLEAN DEFAULT FALSE,
			self_destruct_after_attempts BOOLEAN DEFAULT FALSE,
			max_attempts INTEGER DEFAULT 3,
			require_mfa BOOLEAN DEFAULT FALSE,
			mfa_type TEXT DEFAULT 'TOTP',
			mfa_on_open BOOLEAN DEFAULT FALSE,
			mfa_on_reply BOOLEAN DEFAULT FALSE,
			mfa_on_forward BOOLEAN DEFAULT FALSE,
			remote_revoke BOOLEAN DEFAULT FALSE,
			decoy_message BOOLEAN DEFAULT FALSE,
			decoy_secret TEXT,
			strip_metadata BOOLEAN DEFAULT FALSE,
			tamper_alerts BOOLEAN DEFAULT FALSE,
			geolocation_json TEXT,
			is_password_protected BOOLEAN DEFAULT FALSE,
			password_hash TEXT,
			password_salt TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create audit_log table
	_, err = db.Exec(`
		CREATE TABLE audit_log (
			log_id TEXT PRIMARY KEY,
			timestamp DATETIME,
			event_type TEXT,
			user_id TEXT,
			ip_address TEXT,
			user_agent TEXT,
			related_email_id TEXT,
			outcome TEXT,
			details TEXT,
			severity TEXT,
			session_id TEXT,
			request_id TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			created_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create audit_log table: %v", err)
	}

	return db
}

func TestNewEmailSecurityService(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.db == nil {
		t.Error("Expected database to be set")
	}
}

func TestApplySecurityFeatures_Basic(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-1")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		PasswordProtection: true,
		Password:          "testpassword123",
		BurnAfterRead:     true,
		RequireMFA:        true,
		MFAType:           "TOTP",
	}

	ctx := context.Background()
	err = service.ApplySecurityFeatures(ctx, "test-email-1", "test@securesystem.email", config)
	if err != nil {
		t.Fatalf("Failed to apply security features: %v", err)
	}

	// Verify the features were applied
	var timeLock, burnAfterRead, requireMFA bool
	var mfaType string
	err = db.QueryRow(`
		SELECT time_lock, burn_after_read, require_mfa, mfa_type 
		FROM emails WHERE email_id = ?
	`, "test-email-1").Scan(&timeLock, &burnAfterRead, &requireMFA, &mfaType)
	if err != nil {
		t.Fatalf("Failed to query applied features: %v", err)
	}

	if timeLock != false {
		t.Error("Expected time_lock to be false")
	}
	if !burnAfterRead {
		t.Error("Expected burn_after_read to be true")
	}
	if !requireMFA {
		t.Error("Expected require_mfa to be true")
	}
	if mfaType != "TOTP" {
		t.Errorf("Expected mfa_type to be 'TOTP', got '%s'", mfaType)
	}
}

func TestApplySecurityFeatures_TimeBasedControls(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-2")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		TimeLock:    true,
		UnlockAfter: "2024-12-31T23:59:59Z",
		ExpiresAt:   "2025-01-31T23:59:59Z",
	}

	ctx := context.Background()
	err = service.ApplySecurityFeatures(ctx, "test-email-2", "test@securesystem.email", config)
	if err != nil {
		t.Fatalf("Failed to apply time-based controls: %v", err)
	}

	// Verify time-based controls were applied
	var timeLock bool
	var unlockAfter, expiresAt sql.NullString
	err = db.QueryRow(`
		SELECT time_lock, unlock_after, expires_at 
		FROM emails WHERE email_id = ?
	`, "test-email-2").Scan(&timeLock, &unlockAfter, &expiresAt)
	if err != nil {
		t.Fatalf("Failed to query time-based controls: %v", err)
	}

	if !timeLock {
		t.Error("Expected time_lock to be true")
	}
	if !unlockAfter.Valid || unlockAfter.String != "2024-12-31T23:59:59Z" {
		t.Errorf("Expected unlock_after to be '2024-12-31T23:59:59Z', got '%v'", unlockAfter.String)
	}
	if !expiresAt.Valid || expiresAt.String != "2025-01-31T23:59:59Z" {
		t.Errorf("Expected expires_at to be '2025-01-31T23:59:59Z', got '%v'", expiresAt.String)
	}
}

func TestApplySecurityFeatures_AccessControl(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-3")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		BurnAfterRead:             true,
		SelfDestructAfterAttempts: true,
		MaxFailedAttempts:         5,
	}

	ctx := context.Background()
	err = service.ApplySecurityFeatures(ctx, "test-email-3", "test@securesystem.email", config)
	if err != nil {
		t.Fatalf("Failed to apply access control: %v", err)
	}

	// Verify access control was applied
	var burnAfterRead, selfDestructAfterAttempts bool
	var maxAttempts int
	err = db.QueryRow(`
		SELECT burn_after_read, self_destruct_after_attempts, max_attempts 
		FROM emails WHERE email_id = ?
	`, "test-email-3").Scan(&burnAfterRead, &selfDestructAfterAttempts, &maxAttempts)
	if err != nil {
		t.Fatalf("Failed to query access control: %v", err)
	}

	if !burnAfterRead {
		t.Error("Expected burn_after_read to be true")
	}
	if !selfDestructAfterAttempts {
		t.Error("Expected self_destruct_after_attempts to be true")
	}
	if maxAttempts != 5 {
		t.Errorf("Expected max_attempts to be 5, got %d", maxAttempts)
	}
}

func TestApplySecurityFeatures_GeolocationRestrictions(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-4")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		GeoVerificationType: "city_country",
		GeoCity:            "New York",
		GeoCountry:         "US",
	}

	ctx := context.Background()
	err = service.ApplySecurityFeatures(ctx, "test-email-4", "test@securesystem.email", config)
	if err != nil {
		t.Fatalf("Failed to apply geolocation restrictions: %v", err)
	}

	// Verify geolocation restrictions were applied
	var geoJSON sql.NullString
	err = db.QueryRow(`
		SELECT geolocation_json 
		FROM emails WHERE email_id = ?
	`, "test-email-4").Scan(&geoJSON)
	if err != nil {
		t.Fatalf("Failed to query geolocation restrictions: %v", err)
	}

	if !geoJSON.Valid {
		t.Error("Expected geolocation_json to be set")
	}

	// Verify JSON contains expected data
	expectedJSON := `{"verification_type":"city_country","city":"New York","country":"US"}`
	if geoJSON.String != expectedJSON {
		t.Errorf("Expected geolocation_json to be '%s', got '%s'", expectedJSON, geoJSON.String)
	}
}

func TestApplySecurityFeatures_MFA(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-5")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		RequireMFA:   true,
		MFAType:      "EMAIL_CODE",
		MFAOnOpen:    true,
		MFAOnReply:   true,
		MFAOnForward: true,
	}

	ctx := context.Background()
	err = service.ApplySecurityFeatures(ctx, "test-email-5", "test@securesystem.email", config)
	if err != nil {
		t.Fatalf("Failed to apply MFA settings: %v", err)
	}

	// Verify MFA settings were applied
	var requireMFA, mfaOnOpen, mfaOnReply, mfaOnForward bool
	var mfaType string
	err = db.QueryRow(`
		SELECT require_mfa, mfa_type, mfa_on_open, mfa_on_reply, mfa_on_forward 
		FROM emails WHERE email_id = ?
	`, "test-email-5").Scan(&requireMFA, &mfaType, &mfaOnOpen, &mfaOnReply, &mfaOnForward)
	if err != nil {
		t.Fatalf("Failed to query MFA settings: %v", err)
	}

	if !requireMFA {
		t.Error("Expected require_mfa to be true")
	}
	if mfaType != "EMAIL_CODE" {
		t.Errorf("Expected mfa_type to be 'EMAIL_CODE', got '%s'", mfaType)
	}
	if !mfaOnOpen {
		t.Error("Expected mfa_on_open to be true")
	}
	if !mfaOnReply {
		t.Error("Expected mfa_on_reply to be true")
	}
	if !mfaOnForward {
		t.Error("Expected mfa_on_forward to be true")
	}
}

func TestApplySecurityFeatures_AdvancedSecurity(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	// Insert test email
	_, err := db.Exec(`INSERT INTO emails (email_id) VALUES (?)`, "test-email-6")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		RemoteRevoke:  true,
		DecoyMessage:  true,
		DecoySecret:   "test-secret",
		StripMetadata: true,
		TamperAlerts:  true,
	}

	ctx := context.Background()
	err = service.ApplySecurityFeatures(ctx, "test-email-6", "test@securesystem.email", config)
	if err != nil {
		t.Fatalf("Failed to apply advanced security: %v", err)
	}

	// Verify advanced security features were applied
	var remoteRevoke, decoyMessage, stripMetadata, tamperAlerts bool
	var decoySecret sql.NullString
	err = db.QueryRow(`
		SELECT remote_revoke, decoy_message, decoy_secret, strip_metadata, tamper_alerts 
		FROM emails WHERE email_id = ?
	`, "test-email-6").Scan(&remoteRevoke, &decoyMessage, &decoySecret, &stripMetadata, &tamperAlerts)
	if err != nil {
		t.Fatalf("Failed to query advanced security: %v", err)
	}

	if !remoteRevoke {
		t.Error("Expected remote_revoke to be true")
	}
	if !decoyMessage {
		t.Error("Expected decoy_message to be true")
	}
	if !decoySecret.Valid {
		t.Error("Expected decoy_secret to be set")
	}
	if !stripMetadata {
		t.Error("Expected strip_metadata to be true")
	}
	if !tamperAlerts {
		t.Error("Expected tamper_alerts to be true")
	}
}

func TestSendEmailWithSecurity_Internal(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	deliveryConfig := EmailDeliveryConfig{
		DeliveryType:      "internal",
		InternalRecipient: "internal@securesystem.email",
		ExternalSecurityFeatures: SecurityFeatureConfig{
			PasswordProtection: true,
			Password:          "testpassword123",
			BurnAfterRead:     true,
		},
	}

	ctx := context.Background()
	err := service.SendEmailWithSecurity(ctx, "test@securesystem.email", deliveryConfig)
	if err != nil {
		t.Fatalf("Failed to send internal email: %v", err)
	}

	// Verify email was created with security features
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM emails WHERE recipient = ?`, "internal@securesystem.email").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count emails: %v", err)
	}

	if count == 0 {
		t.Error("Expected internal email to be created")
	}
}

func TestSendEmailWithSecurity_External(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	deliveryConfig := EmailDeliveryConfig{
		DeliveryType:      "external",
		ExternalRecipient: "external@example.com",
		ExternalSecurityFeatures: SecurityFeatureConfig{
			PasswordProtection: true,
			Password:          "testpassword123",
			BurnAfterRead:     true,
		},
	}

	ctx := context.Background()
	err := service.SendEmailWithSecurity(ctx, "test@securesystem.email", deliveryConfig)
	if err != nil {
		t.Fatalf("Failed to send external email: %v", err)
	}

	// Verify email was created with enhanced security features
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM emails WHERE recipient = ?`, "external@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count emails: %v", err)
	}

	if count == 0 {
		t.Error("Expected external email to be created")
	}
}

func TestSendEmailWithSecurity_Both(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	deliveryConfig := EmailDeliveryConfig{
		DeliveryType:      "both",
		InternalRecipient: "internal@securesystem.email",
		ExternalRecipient: "external@example.com",
		ExternalSecurityFeatures: SecurityFeatureConfig{
			PasswordProtection: true,
			Password:          "testpassword123",
			BurnAfterRead:     true,
		},
	}

	ctx := context.Background()
	err := service.SendEmailWithSecurity(ctx, "test@securesystem.email", deliveryConfig)
	if err != nil {
		t.Fatalf("Failed to send both emails: %v", err)
	}

	// Verify both emails were created
	var internalCount, externalCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM emails WHERE recipient = ?`, "internal@securesystem.email").Scan(&internalCount)
	if err != nil {
		t.Fatalf("Failed to count internal emails: %v", err)
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM emails WHERE recipient = ?`, "external@example.com").Scan(&externalCount)
	if err != nil {
		t.Fatalf("Failed to count external emails: %v", err)
	}

	if internalCount == 0 {
		t.Error("Expected internal email to be created")
	}
	if externalCount == 0 {
		t.Error("Expected external email to be created")
	}
}

func TestValidateSecurityConfig_Valid(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		PasswordProtection: true,
		Password:          "validpassword123",
		RequireMFA:        true,
		MFAType:           "TOTP",
		GeoVerificationType: "city",
		MaxFailedAttempts: 5,
	}

	err := service.ValidateSecurityConfig(config)
	if err != nil {
		t.Fatalf("Expected valid config, got error: %v", err)
	}
}

func TestValidateSecurityConfig_InvalidPassword(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		PasswordProtection: true,
		Password:          "123", // Too short
	}

	err := service.ValidateSecurityConfig(config)
	if err == nil {
		t.Error("Expected error for invalid password")
	}
}

func TestValidateSecurityConfig_InvalidMFAType(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		RequireMFA: true,
		MFAType:    "INVALID_TYPE",
	}

	err := service.ValidateSecurityConfig(config)
	if err == nil {
		t.Error("Expected error for invalid MFA type")
	}
}

func TestValidateSecurityConfig_InvalidGeoVerificationType(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		GeoVerificationType: "invalid_type",
	}

	err := service.ValidateSecurityConfig(config)
	if err == nil {
		t.Error("Expected error for invalid geolocation verification type")
	}
}

func TestValidateSecurityConfig_InvalidMaxAttempts(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		MaxFailedAttempts: 15, // Too high
	}

	err := service.ValidateSecurityConfig(config)
	if err == nil {
		t.Error("Expected error for invalid max attempts")
	}
}

func TestValidateSecurityConfig_TimeLockWithoutUnlockTime(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	config := SecurityFeatureConfig{
		TimeLock: true,
		// Missing UnlockAfter
	}

	err := service.ValidateSecurityConfig(config)
	if err == nil {
		t.Error("Expected error for time lock without unlock time")
	}
}

func TestGenerateEmailID(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	
	emailID1 := service.generateEmailID()
	emailID2 := service.generateEmailID()

	if emailID1 == emailID2 {
		t.Error("Expected different email IDs")
	}

	if len(emailID1) == 0 {
		t.Error("Expected non-empty email ID")
	}
}

func TestBuildSecurityFeaturesDescription(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	
	config := SecurityFeatureConfig{
		PasswordProtection: true,
		BurnAfterRead:     true,
		RequireMFA:        true,
		MFAType:           "TOTP",
		RemoteRevoke:      true,
		DecoyMessage:      true,
		StripMetadata:     true,
		TamperAlerts:      true,
	}

	description := service.buildSecurityFeaturesDescription(config)
	
	if len(description) == 0 {
		t.Error("Expected non-empty description")
	}

	// Verify key features are mentioned
	expectedFeatures := []string{
		"Password Protection Required",
		"Self-Destruct After First Read",
		"Multi-Factor Authentication (TOTP)",
		"Remote Revocation Enabled",
		"Decoy Message Protection",
		"Metadata Stripped",
		"Tamper Detection Enabled",
		"End-to-End Encryption",
	}

	for _, feature := range expectedFeatures {
		if !containsSecurityFeature(description, feature) {
			t.Errorf("Expected description to contain '%s'", feature)
		}
	}
}

func TestBuildSecurityFeaturesDescription_Empty(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	service := NewEmailSecurityService(db)
	
	config := SecurityFeatureConfig{}
	description := service.buildSecurityFeaturesDescription(config)
	
	expected := "• Standard Encryption"
	if description != expected {
		t.Errorf("Expected '%s', got '%s'", expected, description)
	}
}

// Helper function to check if a string contains a substring
func containsSecurityFeature(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSecuritySubstring(s, substr))))
}

func containsSecuritySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
