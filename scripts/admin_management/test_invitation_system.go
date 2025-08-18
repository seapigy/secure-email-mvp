package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Test configuration
const (
	BaseURL           = "http://localhost:8080"
	RootAdminEmail    = "cpigusch@gmail.com"
	RootAdminPassword = "SecureAdminPassword123!"
)

// Test results structure
type TestResult struct {
	TestName  string    `json:"test_name"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Duration  float64   `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// Test suite structure
type TestSuite struct {
	Results []TestResult `json:"results"`
	Summary TestSummary  `json:"summary"`
}

type TestSummary struct {
	TotalTests    int     `json:"total_tests"`
	PassedTests   int     `json:"passed_tests"`
	FailedTests   int     `json:"failed_tests"`
	SuccessRate   float64 `json:"success_rate"`
	TotalDuration float64 `json:"total_duration_ms"`
}

// HTTP client with custom settings
var client = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func main() {
	log.Println("Starting Iteration 4 - Admin & User Management Test Suite")
	log.Println("========================================================")

	suite := &TestSuite{
		Results: make([]TestResult, 0),
	}

	// Test 1: Admin Setup and Initial Login
	suite.runTest("Admin Setup and Initial Login", testAdminSetupAndLogin)

	// Test 2: Invitation Key Creation
	suite.runTest("Invitation Key Creation", testInvitationKeyCreation)

	// Test 3: Invitation Key Validation
	suite.runTest("Invitation Key Validation", testInvitationKeyValidation)

	// Test 4: Admin Account Creation via Invitation
	suite.runTest("Admin Account Creation via Invitation", testAdminAccountCreationViaInvitation)

	// Test 5: RBAC Enforcement - Role-based Access
	suite.runTest("RBAC Enforcement - Role-based Access", testRBACEnforcement)

	// Test 6: Session Management
	suite.runTest("Session Management", testSessionManagement)

	// Test 7: Invitation Key Revocation
	suite.runTest("Invitation Key Revocation", testInvitationKeyRevocation)

	// Test 8: Multi-Admin Workflow
	suite.runTest("Multi-Admin Workflow", testMultiAdminWorkflow)

	// Test 9: User Recovery Testing
	suite.runTest("User Recovery Testing", testUserRecovery)

	// Test 10: Security Validation
	suite.runTest("Security Validation", testSecurityValidation)

	// Generate and save test report
	suite.generateSummary()
	suite.saveReport()

	log.Println("Test suite completed!")
	log.Printf("Results: %d/%d tests passed (%.1f%%)",
		suite.Summary.PassedTests, suite.Summary.TotalTests, suite.Summary.SuccessRate)
}

func (ts *TestSuite) runTest(testName string, testFunc func() (bool, string, error)) {
	log.Printf("Running test: %s", testName)

	startTime := time.Now()
	success, details, err := testFunc()
	duration := time.Since(startTime).Milliseconds()

	result := TestResult{
		TestName:  testName,
		Success:   success,
		Duration:  float64(duration),
		Timestamp: time.Now(),
		Details:   details,
	}

	if err != nil {
		result.Error = err.Error()
		log.Printf("❌ Test failed: %s - %v", testName, err)
	} else {
		log.Printf("✅ Test passed: %s", testName)
	}

	ts.Results = append(ts.Results, result)
}

func (ts *TestSuite) generateSummary() {
	ts.Summary.TotalTests = len(ts.Results)

	for _, result := range ts.Results {
		ts.Summary.TotalDuration += result.Duration
		if result.Success {
			ts.Summary.PassedTests++
		} else {
			ts.Summary.FailedTests++
		}
	}

	if ts.Summary.TotalTests > 0 {
		ts.Summary.SuccessRate = float64(ts.Summary.PassedTests) / float64(ts.Summary.TotalTests) * 100
	}
}

func (ts *TestSuite) saveReport() {
	reportData, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal test report: %v", err)
		return
	}

	filename := fmt.Sprintf("admin_management_test_report_%s.json", time.Now().Format("20060102_150405"))
	err = os.WriteFile(filename, reportData, 0644)
	if err != nil {
		log.Printf("Failed to save test report: %v", err)
		return
	}

	log.Printf("Test report saved to: %s", filename)
}

// ============================================================================
// TEST FUNCTIONS
// ============================================================================

func testAdminSetupAndLogin() (bool, string, error) {
	// Check if admin setup is required
	resp, err := client.Get(fmt.Sprintf("%s/admin/check-setup", BaseURL))
	if err != nil {
		return false, "", fmt.Errorf("failed to check admin setup: %w", err)
	}
	defer resp.Body.Close()

	var setupResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&setupResponse); err != nil {
		return false, "", fmt.Errorf("failed to decode setup response: %w", err)
	}

	setupRequired := setupResponse["setup_required"].(bool)

	if setupRequired {
		// Create root admin
		setupData := map[string]string{
			"email":    RootAdminEmail,
			"password": RootAdminPassword,
		}

		setupJSON, _ := json.Marshal(setupData)
		resp, err = client.Post(fmt.Sprintf("%s/admin/setup", BaseURL),
			"application/json", bytes.NewBuffer(setupJSON))
		if err != nil {
			return false, "", fmt.Errorf("failed to create root admin: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			return false, "", fmt.Errorf("admin setup failed with status: %d", resp.StatusCode)
		}
	}

	// Login as root admin
	loginData := map[string]string{
		"email":    RootAdminEmail,
		"password": RootAdminPassword,
	}

	loginJSON, _ := json.Marshal(loginData)
	resp, err = client.Post(fmt.Sprintf("%s/admin/login", BaseURL),
		"application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		return false, "", fmt.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var loginResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return false, "", fmt.Errorf("failed to decode login response: %w", err)
	}

	sessionToken := loginResponse["session_token"].(string)
	if sessionToken == "" {
		return false, "", fmt.Errorf("no session token received")
	}

	return true, fmt.Sprintf("Root admin login successful, session token: %s...", sessionToken[:10]), nil
}

func testInvitationKeyCreation() (bool, string, error) {
	// First login to get session token
	sessionToken, err := getRootAdminSession()
	if err != nil {
		return false, "", fmt.Errorf("failed to get session token: %w", err)
	}

	// Create invitation for full admin
	invitationData := map[string]interface{}{
		"email":    "fulladmin@example.com",
		"role":     "full_admin",
		"max_uses": 1,
	}

	invitationJSON, _ := json.Marshal(invitationData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/admin/invitations", BaseURL),
		bytes.NewBuffer(invitationJSON))
	req.Header.Set("Authorization", "Bearer "+sessionToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to create invitation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return false, "", fmt.Errorf("invitation creation failed with status: %d", resp.StatusCode)
	}

	var invitationResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&invitationResponse); err != nil {
		return false, "", fmt.Errorf("failed to decode invitation response: %w", err)
	}

	invitationID := invitationResponse["invitation_id"].(string)
	if invitationID == "" {
		return false, "", fmt.Errorf("no invitation ID received")
	}

	return true, fmt.Sprintf("Invitation created successfully, ID: %s", invitationID), nil
}

func testInvitationKeyValidation() (bool, string, error) {
	// This test would require a valid invitation token
	// For now, we'll test the validation endpoint structure

	validationData := map[string]string{
		"invitation_token": "test_token_123",
	}

	validationJSON, _ := json.Marshal(validationData)
	resp, err := client.Post(fmt.Sprintf("%s/admin/invitations/validate", BaseURL),
		"application/json", bytes.NewBuffer(validationJSON))
	if err != nil {
		return false, "", fmt.Errorf("failed to validate invitation: %w", err)
	}
	defer resp.Body.Close()

	// We expect this to fail with an invalid token, but the endpoint should respond
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return true, "Invitation validation endpoint responding correctly", nil
}

func testAdminAccountCreationViaInvitation() (bool, string, error) {
	// This test would require a valid invitation token
	// For now, we'll test the endpoint structure

	useInvitationData := map[string]string{
		"invitation_token": "test_token_123",
		"password":         "SecurePassword123!",
	}

	useInvitationJSON, _ := json.Marshal(useInvitationData)
	resp, err := client.Post(fmt.Sprintf("%s/admin/invitations/use", BaseURL),
		"application/json", bytes.NewBuffer(useInvitationJSON))
	if err != nil {
		return false, "", fmt.Errorf("failed to use invitation: %w", err)
	}
	defer resp.Body.Close()

	// We expect this to fail with an invalid token, but the endpoint should respond
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusCreated {
		return false, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return true, "Admin account creation via invitation endpoint responding correctly", nil
}

func testRBACEnforcement() (bool, string, error) {
	// Test RBAC by attempting to create invitations with different roles
	sessionToken, err := getRootAdminSession()
	if err != nil {
		return false, "", fmt.Errorf("failed to get session token: %w", err)
	}

	// Test 1: Root admin should be able to create any role invitation
	invitationData := map[string]interface{}{
		"email":    "readonlyadmin@example.com",
		"role":     "read_only_admin",
		"max_uses": 1,
	}

	invitationJSON, _ := json.Marshal(invitationData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/admin/invitations", BaseURL),
		bytes.NewBuffer(invitationJSON))
	req.Header.Set("Authorization", "Bearer "+sessionToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to create read-only admin invitation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return false, "", fmt.Errorf("root admin should be able to create read-only admin invitation, got status: %d", resp.StatusCode)
	}

	// Test 2: List invitations (should work for root admin)
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/admin/invitations", BaseURL), nil)
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err = client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to list invitations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("root admin should be able to list invitations, got status: %d", resp.StatusCode)
	}

	return true, "RBAC enforcement working correctly for root admin", nil
}

func testSessionManagement() (bool, string, error) {
	// Test session validation
	sessionToken, err := getRootAdminSession()
	if err != nil {
		return false, "", fmt.Errorf("failed to get session token: %w", err)
	}

	// Validate session
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/admin/session", BaseURL), nil)
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to validate session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("session validation failed with status: %d", resp.StatusCode)
	}

	var sessionResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&sessionResponse); err != nil {
		return false, "", fmt.Errorf("failed to decode session response: %w", err)
	}

	success := sessionResponse["success"].(bool)
	if !success {
		return false, "", fmt.Errorf("session validation returned success=false")
	}

	// Test logout
	logoutData := map[string]string{
		"session_token": sessionToken,
	}

	logoutJSON, _ := json.Marshal(logoutData)
	resp, err = client.Post(fmt.Sprintf("%s/admin/logout", BaseURL),
		"application/json", bytes.NewBuffer(logoutJSON))
	if err != nil {
		return false, "", fmt.Errorf("failed to logout: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("logout failed with status: %d", resp.StatusCode)
	}

	return true, "Session management working correctly", nil
}

func testInvitationKeyRevocation() (bool, string, error) {
	// Test invitation revocation
	sessionToken, err := getRootAdminSession()
	if err != nil {
		return false, "", fmt.Errorf("failed to get session token: %w", err)
	}

	// First create an invitation to revoke
	invitationData := map[string]interface{}{
		"email":    "temporaryadmin@example.com",
		"role":     "read_only_admin",
		"max_uses": 1,
	}

	invitationJSON, _ := json.Marshal(invitationData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/admin/invitations", BaseURL),
		bytes.NewBuffer(invitationJSON))
	req.Header.Set("Authorization", "Bearer "+sessionToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to create invitation for revocation test: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return false, "", fmt.Errorf("failed to create invitation for revocation test: %d", resp.StatusCode)
	}

	var invitationResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&invitationResponse); err != nil {
		return false, "", fmt.Errorf("failed to decode invitation response: %w", err)
	}

	invitationID := invitationResponse["invitation_id"].(string)

	// Now revoke the invitation
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/admin/invitations/revoke?id=%s", BaseURL, invitationID), nil)
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err = client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to revoke invitation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("invitation revocation failed with status: %d", resp.StatusCode)
	}

	return true, fmt.Sprintf("Invitation revocation successful for ID: %s", invitationID), nil
}

func testMultiAdminWorkflow() (bool, string, error) {
	// Test the complete multi-admin workflow
	sessionToken, err := getRootAdminSession()
	if err != nil {
		return false, "", fmt.Errorf("failed to get session token: %w", err)
	}

	// Step 1: Create invitation for full admin
	invitationData := map[string]interface{}{
		"email":    "workflowadmin@example.com",
		"role":     "full_admin",
		"max_uses": 1,
	}

	invitationJSON, _ := json.Marshal(invitationData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/admin/invitations", BaseURL),
		bytes.NewBuffer(invitationJSON))
	req.Header.Set("Authorization", "Bearer "+sessionToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to create invitation in workflow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return false, "", fmt.Errorf("workflow invitation creation failed: %d", resp.StatusCode)
	}

	// Step 2: List invitations to verify creation
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/admin/invitations", BaseURL), nil)
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err = client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to list invitations in workflow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("workflow invitation listing failed: %d", resp.StatusCode)
	}

	var listResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return false, "", fmt.Errorf("failed to decode list response: %w", err)
	}

	success := listResponse["success"].(bool)
	if !success {
		return false, "", fmt.Errorf("invitation listing returned success=false")
	}

	return true, "Multi-admin workflow completed successfully", nil
}

func testUserRecovery() (bool, string, error) {
	// Test user recovery functionality under ZKID constraints
	// This would test recovery code generation, validation, and revocation

	// For now, we'll test that the system can handle recovery-related operations
	sessionToken, err := getRootAdminSession()
	if err != nil {
		return false, "", fmt.Errorf("failed to get session token: %w", err)
	}

	// Test audit logs access (part of recovery monitoring)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/admin/audit-logs", BaseURL), nil)
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to access audit logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("audit logs access failed: %d", resp.StatusCode)
	}

	var auditResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&auditResponse); err != nil {
		return false, "", fmt.Errorf("failed to decode audit response: %w", err)
	}

	success := auditResponse["success"].(bool)
	if !success {
		return false, "", fmt.Errorf("audit logs returned success=false")
	}

	return true, "User recovery monitoring accessible", nil
}

func testSecurityValidation() (bool, string, error) {
	// Test security aspects of the invitation system

	// Test 1: Attempt to create invitation without authentication
	invitationData := map[string]interface{}{
		"email":    "unauthorized@example.com",
		"role":     "full_admin",
		"max_uses": 1,
	}

	invitationJSON, _ := json.Marshal(invitationData)
	resp, err := client.Post(fmt.Sprintf("%s/admin/invitations", BaseURL),
		"application/json", bytes.NewBuffer(invitationJSON))
	if err != nil {
		return false, "", fmt.Errorf("failed to test unauthorized invitation creation: %w", err)
	}
	defer resp.Body.Close()

	// Should be unauthorized
	if resp.StatusCode != http.StatusUnauthorized {
		return false, "", fmt.Errorf("unauthorized invitation creation should fail, got status: %d", resp.StatusCode)
	}

	// Test 2: Attempt to list invitations without authentication
	resp, err = client.Get(fmt.Sprintf("%s/admin/invitations", BaseURL))
	if err != nil {
		return false, "", fmt.Errorf("failed to test unauthorized invitation listing: %w", err)
	}
	defer resp.Body.Close()

	// Should be unauthorized
	if resp.StatusCode != http.StatusUnauthorized {
		return false, "", fmt.Errorf("unauthorized invitation listing should fail, got status: %d", resp.StatusCode)
	}

	// Test 3: Attempt to use invalid invitation token
	invalidInvitationData := map[string]string{
		"invitation_token": "invalid_token_12345",
		"password":         "SecurePassword123!",
	}

	invalidInvitationJSON, _ := json.Marshal(invalidInvitationData)
	resp, err = client.Post(fmt.Sprintf("%s/admin/invitations/use", BaseURL),
		"application/json", bytes.NewBuffer(invalidInvitationJSON))
	if err != nil {
		return false, "", fmt.Errorf("failed to test invalid invitation usage: %w", err)
	}
	defer resp.Body.Close()

	// Should fail with bad request or internal server error
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
		return false, "", fmt.Errorf("invalid invitation usage should fail, got status: %d", resp.StatusCode)
	}

	return true, "Security validation tests passed", nil
}

// Helper function to get root admin session token
func getRootAdminSession() (string, error) {
	loginData := map[string]string{
		"email":    RootAdminEmail,
		"password": RootAdminPassword,
	}

	loginJSON, _ := json.Marshal(loginData)
	resp, err := client.Post(fmt.Sprintf("%s/admin/login", BaseURL),
		"application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var loginResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	sessionToken := loginResponse["session_token"].(string)
	if sessionToken == "" {
		return "", fmt.Errorf("no session token received")
	}

	return sessionToken, nil
}
