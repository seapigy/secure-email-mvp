package e2e

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityTestSuite(t *testing.T) {
	config := DefaultSecurityTestConfig()
	suite, err := NewSecurityTestSuite(config)

	require.NoError(t, err)
	assert.NotNil(t, suite)
	assert.NotNil(t, suite.CryptoValidator)
	assert.NotNil(t, suite.ProtocolAnalyzer)
	assert.NotNil(t, suite.PentestHooks)
	assert.NotNil(t, suite.ComplianceTests)
	assert.Equal(t, config.EnableCryptoTests, suite.Config.EnableCryptoTests)
}

func TestSecurityTestConfig_Defaults(t *testing.T) {
	config := DefaultSecurityTestConfig()

	assert.True(t, config.EnableCryptoTests)
	assert.True(t, config.EnableProtocolTests)
	assert.True(t, config.EnablePentestHooks)
	assert.True(t, config.EnableComplianceTests)

	// Check default security levels
	expectedLevels := []string{"128", "192", "256"}
	assert.Equal(t, expectedLevels, config.SecurityLevels)

	// Check default algorithms
	expectedAlgorithms := []string{"kyber512", "kyber768", "kyber1024", "dilithium2", "dilithium3", "dilithium5"}
	assert.Equal(t, expectedAlgorithms, config.Algorithms)

	// Check default compliance standards
	expectedStandards := []string{"FIPS-140-2", "GDPR"}
	assert.Equal(t, expectedStandards, config.ComplianceStandards)
}

func TestSecurityTestSuite_CryptographicTests(t *testing.T) {
	config := SecurityTestConfig{
		EnableCryptoTests:     true,
		EnableProtocolTests:   false,
		EnablePentestHooks:    false,
		EnableComplianceTests: false,
		Algorithms:            []string{"kyber768", "dilithium3"},
	}

	suite, err := NewSecurityTestSuite(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runCryptographicTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.NotEmpty(t, results)

	// Verify crypto test results
	cryptoTests := 0
	for _, result := range results {
		if result.Category == "cryptographic" {
			cryptoTests++
			assert.Equal(t, "pass", result.Status)
			assert.True(t, result.Score > 0)
		}
	}
	assert.True(t, cryptoTests > 0)
}

func TestSecurityTestSuite_KnownAnswerTests(t *testing.T) {
	config := SecurityTestConfig{
		Algorithms: []string{"kyber768", "dilithium3"},
	}

	suite, err := NewSecurityTestSuite(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runKnownAnswerTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "known_answer_tests", result.TestName)
	assert.Equal(t, "cryptographic", result.Category)
	assert.Equal(t, "high", result.Severity)
	assert.Equal(t, "pass", result.Status)
}

func TestSecurityTestSuite_RandomnessTests(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runRandomnessTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "randomness_quality", result.TestName)
	assert.Equal(t, "cryptographic", result.Category)
	assert.Equal(t, "pass", result.Status)
}

func TestSecurityTestSuite_KeyStrengthTests(t *testing.T) {
	config := SecurityTestConfig{
		Algorithms: []string{"kyber768"},
	}

	suite, err := NewSecurityTestSuite(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runKeyStrengthTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "key_strength_validation", result.TestName)
	assert.Equal(t, "pass", result.Status)
}

func TestSecurityTestSuite_ProtocolTests(t *testing.T) {
	config := SecurityTestConfig{
		EnableCryptoTests:   false,
		EnableProtocolTests: true,
	}

	suite, err := NewSecurityTestSuite(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runProtocolSecurityTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.NotEmpty(t, results)

	// Verify protocol test results
	protocolTests := 0
	for _, result := range results {
		if result.Category == "protocol" {
			protocolTests++
			assert.Equal(t, "pass", result.Status)
		}
	}
	assert.True(t, protocolTests > 0)
}

func TestSecurityTestSuite_ConfidentialityTests(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runConfidentialityTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "message_confidentiality", result.TestName)
	assert.Equal(t, "protocol", result.Category)
	assert.Equal(t, "pass", result.Status)
}

func TestSecurityTestSuite_IntegrityTests(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runIntegrityTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "message_integrity", result.TestName)
	assert.Equal(t, "protocol", result.Category)
	assert.Equal(t, "pass", result.Status)
}

func TestSecurityTestSuite_PenetrationTests(t *testing.T) {
	config := SecurityTestConfig{
		EnableCryptoTests:   false,
		EnableProtocolTests: false,
		EnablePentestHooks:  true,
	}

	suite, err := NewSecurityTestSuite(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runPenetrationTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.NotEmpty(t, results)

	// Verify penetration test results
	pentestTests := 0
	for _, result := range results {
		if result.Category == "implementation" {
			pentestTests++
			assert.Equal(t, "pass", result.Status)
		}
	}
	assert.True(t, pentestTests > 0)
}

func TestSecurityTestSuite_InputValidationTests(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runInputValidationTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "input_validation", result.TestName)
	assert.Equal(t, "implementation", result.Category)
	assert.Equal(t, "pass", result.Status)
}

func TestSecurityTestSuite_ComplianceTests(t *testing.T) {
	config := SecurityTestConfig{
		EnableCryptoTests:     false,
		EnableProtocolTests:   false,
		EnablePentestHooks:    false,
		EnableComplianceTests: true,
		ComplianceStandards:   []string{"FIPS-140-2"},
	}

	suite, err := NewSecurityTestSuite(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = suite.runComplianceTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.NotEmpty(t, results)

	// Verify compliance test results
	complianceTests := 0
	for _, result := range results {
		if result.Category == "compliance" {
			complianceTests++
			assert.Equal(t, "pass", result.Status)
		}
	}
	assert.True(t, complianceTests > 0)
}

func TestSecurityTestSuite_ValidationMethods(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	// Test KEM validation
	err = suite.validateKEMKnownAnswers("kyber768")
	assert.NoError(t, err)

	// Test DEM validation
	err = suite.validateDEMKnownAnswers("aes256gcm")
	assert.NoError(t, err)

	// Test signature validation
	err = suite.validateSignatureKnownAnswers("dilithium3")
	assert.NoError(t, err)

	// Test randomness validation
	randomData := make([]byte, 1024)
	for i := range randomData {
		randomData[i] = byte(i % 256) // Generate test data with good distribution
	}
	err = suite.validateRandomnessQuality(randomData)
	assert.NoError(t, err)

	// Test key strength validation
	keyPair, err := suite.CryptoValidator.cryptoProvider.GenerateKeyPair("kyber768")
	require.NoError(t, err)
	err = suite.validateKeyStrength(keyPair, "kyber768")
	assert.NoError(t, err)

	// Test NIST compliance validation
	err = suite.validateNISTCompliance("kyber768")
	assert.NoError(t, err)

	// Test non-compliant algorithm
	err = suite.validateNISTCompliance("invalid_algorithm")
	assert.Error(t, err)
}

func TestSecurityTestSuite_VulnerabilityDetection(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	// Test plaintext leakage detection
	plaintext := []byte("secret message")
	recipientKey := make([]byte, 32)
	message, err := suite.ProtocolAnalyzer.client.EncryptMessage(plaintext, recipientKey, "test_thread", "test_recipient")
	require.NoError(t, err)

	err = suite.validateNoPlaintextLeakage(message, plaintext)
	assert.NoError(t, err)

	// Test tamper detection
	err = suite.validateTamperDetection(message)
	assert.NoError(t, err) // Should detect tampering

	// Test metadata protection
	err = suite.validateMetadataProtection(message)
	assert.NoError(t, err)

	// Test replay protection
	err = suite.validateReplayProtection(message)
	assert.NoError(t, err)
}

func TestSecurityTestSuite_ComplianceValidation(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	// Test encryption compliance
	requirement := ComplianceRequirement{
		ID:        "TEST-ENC-1",
		Type:      "encryption",
		Mandatory: true,
	}
	err = suite.validateComplianceRequirement(requirement)
	assert.NoError(t, err)

	// Test key management compliance
	requirement.Type = "key_management"
	err = suite.validateComplianceRequirement(requirement)
	assert.NoError(t, err)

	// Test unknown compliance type
	requirement.Type = "unknown_type"
	err = suite.validateComplianceRequirement(requirement)
	assert.Error(t, err)
}

func TestSecurityTestSuite_ReportGeneration(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	// Run some tests to generate results
	ctx := context.Background()
	err = suite.runKnownAnswerTests(ctx)
	require.NoError(t, err)

	err = suite.runRandomnessTests(ctx)
	require.NoError(t, err)

	// Generate security report
	report := suite.GenerateSecurityReport()

	assert.True(t, report.TotalTests > 0)
	assert.Equal(t, report.PassedTests+report.FailedTests, report.TotalTests)
	assert.True(t, report.OverallScore >= 0.0)
	assert.True(t, report.OverallScore <= 100.0)
	assert.NotEmpty(t, report.Results)
	assert.NotEmpty(t, report.Summary)

	// Verify category summaries
	for category, summary := range report.Summary {
		assert.NotEmpty(t, category)
		assert.True(t, summary.TotalTests > 0)
		assert.Equal(t, summary.PassedTests+summary.FailedTests, summary.TotalTests)
		assert.True(t, summary.AverageScore >= 0.0)
		assert.True(t, summary.AverageScore <= 100.0)
	}
}

func TestSecurityTestSuite_FullSecurityTestRun(t *testing.T) {
	// Test with minimal configuration for faster execution
	config := SecurityTestConfig{
		EnableCryptoTests:     true,
		EnableProtocolTests:   true,
		EnablePentestHooks:    true,
		EnableComplianceTests: true,
		Algorithms:            []string{"kyber768"}, // Reduced set for testing
		ComplianceStandards:   []string{"FIPS-140-2"},
	}

	suite, err := NewSecurityTestSuite(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = suite.RunAllSecurityTests(ctx)
	require.NoError(t, err)

	results := suite.GetResults()
	assert.NotEmpty(t, results)

	// Verify we have results from all enabled test categories
	categories := make(map[string]bool)
	for _, result := range results {
		categories[result.Category] = true
		assert.NotEmpty(t, result.TestID)
		assert.NotEmpty(t, result.TestName)
		assert.Contains(t, []string{"pass", "fail", "warning"}, result.Status)
		assert.True(t, result.Score >= 0.0)
		assert.True(t, result.Score <= 100.0)
		assert.True(t, result.Duration >= 0) // Duration can be 0 for very fast operations
	}

	// Should have results from enabled categories
	assert.True(t, categories["cryptographic"])
	assert.True(t, categories["protocol"])
	assert.True(t, categories["implementation"])
	assert.True(t, categories["compliance"])

	// Generate final report
	report := suite.GenerateSecurityReport()
	assert.True(t, report.TotalTests > 0)
	assert.True(t, report.OverallScore > 0)
}

func TestSecurityTestSuite_ErrorHandling(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	// Test input validation with various edge cases
	testInputs := [][]byte{
		nil,                      // Null input
		make([]byte, 0),          // Empty input
		make([]byte, 1024*1024),  // Large input
		{0xFF, 0xFF, 0xFF, 0xFF}, // Invalid data
	}

	for _, input := range testInputs {
		err := suite.validateInputHandling(input)
		assert.NoError(t, err, "Input validation should handle edge cases gracefully")
	}
}

func TestSecurityTestSuite_ThreadSafety(t *testing.T) {
	suite, err := NewSecurityTestSuite(DefaultSecurityTestConfig())
	require.NoError(t, err)

	// Test concurrent access to results
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Add test results concurrently
				result := SecurityTestResult{
					TestID:    generateTestID(),
					TestName:  "concurrent_test",
					Category:  "test",
					Status:    "pass",
					Score:     100.0,
					Timestamp: time.Now(),
				}
				suite.addResult(result)

				// Read results concurrently
				results := suite.GetResults()
				assert.NotNil(t, results)
			}
		}()
	}

	wg.Wait()

	// Verify final result count
	results := suite.GetResults()
	assert.Len(t, results, numGoroutines*numOperations)
}

func TestUtilityFunctions(t *testing.T) {
	// Test ID generation
	id1 := generateTestID()
	id2 := generateTestID()
	assert.NotEqual(t, id1, id2)
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)

	// Test vulnerability ID generation
	vuln1 := generateVulnID()
	vuln2 := generateVulnID()
	assert.NotEqual(t, vuln1, vuln2)
	assert.NotEmpty(t, vuln1)
	assert.NotEmpty(t, vuln2)

	// Test compliance standards initialization
	standards := initializeComplianceStandards()
	assert.NotEmpty(t, standards)
	assert.Contains(t, standards, "FIPS-140-2")
	assert.Contains(t, standards, "GDPR")

	fips := standards["FIPS-140-2"]
	assert.Equal(t, "FIPS 140-2", fips.Name)
	assert.NotEmpty(t, fips.Requirements)

	gdpr := standards["GDPR"]
	assert.Equal(t, "General Data Protection Regulation", gdpr.Name)
	assert.NotEmpty(t, gdpr.Requirements)
}
