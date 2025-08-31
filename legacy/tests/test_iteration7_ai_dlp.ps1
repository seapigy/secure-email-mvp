#!/usr/bin/env pwsh

# Iteration 7 - AI-Powered DLP Integration Test
# Tests AI-powered content classification, severity scoring, and override capabilities

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "TestPassword123!",
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

# Test counters
$script:totalTests = 0
$script:passedTests = 0
$script:failedTests = 0

function Write-TestLog {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $color = switch ($Level) {
        "PASS" { "Green" }
        "FAIL" { "Red" }
        "WARN" { "Yellow" }
        "INFO" { "Cyan" }
        default { "White" }
    }
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $color
}

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [object]$Body = $null,
        [scriptblock]$Validation
    )
    
    $script:totalTests++
    Write-TestLog "Testing: $Name" "INFO"
    
    try {
        $headers = @{
            "Content-Type" = "application/json"
        }
        
        $params = @{
            Uri = "$BaseUrl$Url"
            Method = $Method
            Headers = $headers
        }
        
        if ($Body) {
            $params.Body = $Body | ConvertTo-Json -Depth 10
        }
        
        $response = Invoke-RestMethod @params -ErrorAction Stop
        
        if ($Validation) {
            & $Validation $response
            $script:passedTests++
            Write-TestLog "✓ PASS: $Name" "PASS"
        } else {
            $script:passedTests++
            Write-TestLog "✓ PASS: $Name" "PASS"
        }
    }
    catch {
        $script:failedTests++
        Write-TestLog "✗ FAIL: $Name - $($_.Exception.Message)" "FAIL"
        if ($Verbose) {
            Write-TestLog "Error details: $($_.Exception.Response)" "WARN"
        }
    }
}

function Test-AIDLPScanning {
    Write-TestLog "=== Testing AI DLP Scanning ===" "INFO"
    
    # Test 1: PII Content Detection
    $piiContent = @{
        content = "My social security number is 123-45-6789 and my credit card is 4111-1111-1111-1111"
        content_type = "email_body"
        link_id = "test_link_pii_001"
    }
    
    Test-Endpoint -Name "AI DLP Scan - PII Detection" -Method "POST" -Url "/api/v/test_link_pii_001/ai-dlp/scan" -Body $piiContent -Validation {
        param($response)
        if (-not $response.success) { throw "AI DLP scan failed" }
        if ($response.classification.category -eq "none") { throw "PII content not detected" }
        if ($response.severity_score -lt 0.5) { throw "Severity score too low for PII content" }
        if ($response.risk_level -notin @("high", "critical")) { throw "Risk level not appropriate for PII" }
    }
    
    # Test 2: Financial Content Detection
    $financialContent = @{
        content = "Bank account number: 1234567890, routing number: 021000021, amount: $50,000"
        content_type = "email_body"
        link_id = "test_link_financial_001"
    }
    
    Test-Endpoint -Name "AI DLP Scan - Financial Detection" -Method "POST" -Url "/api/v/test_link_financial_001/ai-dlp/scan" -Body $financialContent -Validation {
        param($response)
        if (-not $response.success) { throw "AI DLP scan failed" }
        if ($response.classification.category -ne "financial") { throw "Financial content not properly classified" }
        if ($response.severity_score -lt 0.7) { throw "Severity score too low for financial content" }
        if ($response.risk_level -ne "critical") { throw "Risk level not critical for financial data" }
    }
    
    # Test 3: Healthcare Content Detection
    $healthcareContent = @{
        content = "Patient diagnosis: diabetes, treatment: insulin, medical record: PHI-12345"
        content_type = "email_body"
        link_id = "test_link_healthcare_001"
    }
    
    Test-Endpoint -Name "AI DLP Scan - Healthcare Detection" -Method "POST" -Url "/api/v/test_link_healthcare_001/ai-dlp/scan" -Body $healthcareContent -Validation {
        param($response)
        if (-not $response.success) { throw "AI DLP scan failed" }
        if ($response.classification.category -ne "healthcare") { throw "Healthcare content not properly classified" }
        if ($response.severity_score -lt 0.8) { throw "Severity score too low for healthcare content" }
        if ($response.risk_level -ne "critical") { throw "Risk level not critical for healthcare data" }
    }
    
    # Test 4: Clean Content
    $cleanContent = @{
        content = "Hello, this is a regular email about the weather and weekend plans."
        content_type = "email_body"
        link_id = "test_link_clean_001"
    }
    
    Test-Endpoint -Name "AI DLP Scan - Clean Content" -Method "POST" -Url "/api/v/test_link_clean_001/ai-dlp/scan" -Body $cleanContent -Validation {
        param($response)
        if (-not $response.success) { throw "AI DLP scan failed" }
        if ($response.classification.category -ne "none") { throw "Clean content incorrectly classified" }
        if ($response.severity_score -gt 0.1) { throw "Severity score too high for clean content" }
        if ($response.risk_level -ne "none") { throw "Risk level not none for clean content" }
        if ($response.action_taken -ne "allowed") { throw "Action should be allowed for clean content" }
    }
    
    # Test 5: Legal Content Detection
    $legalContent = @{
        content = "Attorney-client privileged communication regarding case number CR-2024-001"
        content_type = "email_body"
        link_id = "test_link_legal_001"
    }
    
    Test-Endpoint -Name "AI DLP Scan - Legal Detection" -Method "POST" -Url "/api/v/test_link_legal_001/ai-dlp/scan" -Body $legalContent -Validation {
        param($response)
        if (-not $response.success) { throw "AI DLP scan failed" }
        if ($response.classification.category -ne "legal") { throw "Legal content not properly classified" }
        if ($response.severity_score -lt 0.6) { throw "Severity score too low for legal content" }
        if ($response.risk_level -notin @("high", "critical")) { throw "Risk level not appropriate for legal content" }
    }
}

function Test-AIDLPOverrides {
    Write-TestLog "=== Testing AI DLP Overrides ===" "INFO"
    
    # Test 1: Valid Override Request
    $overrideRequest = @{
        scan_id = "ai_scan_test_override_001"
        override_reason = "Business necessity"
        user_id = "admin_user"
        user_role = "admin"
        justification = "This is a legitimate business communication that requires sending"
    }
    
    Test-Endpoint -Name "AI DLP Override - Valid Request" -Method "POST" -Url "/api/v/test_link_override_001/ai-dlp/override" -Body $overrideRequest -Validation {
        param($response)
        if (-not $response.success) { throw "Override request failed" }
        if ($response.action_taken -ne "overridden") { throw "Action not overridden" }
        if (-not $response.override_id) { throw "Override ID not returned" }
    }
    
    # Test 2: Invalid Override Request (Unauthorized Role)
    $invalidOverrideRequest = @{
        scan_id = "ai_scan_test_override_002"
        override_reason = "Business necessity"
        user_id = "regular_user"
        user_role = "user"
        justification = "This is a legitimate business communication"
    }
    
    Test-Endpoint -Name "AI DLP Override - Unauthorized Role" -Method "POST" -Url "/api/v/test_link_override_002/ai-dlp/override" -Body $invalidOverrideRequest -Validation {
        param($response)
        if ($response.success) { throw "Override should have failed for unauthorized role" }
        if ($response.error_code -ne "OVERRIDE_NOT_ALLOWED") { throw "Wrong error code for unauthorized override" }
    }
}

function Test-AIDLPPolicies {
    Write-TestLog "=== Testing AI DLP Policies ===" "INFO"
    
    # Test 1: Get AI DLP Policies
    Test-Endpoint -Name "Get AI DLP Policies" -Method "GET" -Url "/api/ai-dlp/policies" -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to get AI DLP policies" }
        if (-not $response.policies) { throw "No policies returned" }
        if ($response.policies.Count -eq 0) { throw "Empty policies list" }
        
        $defaultPolicy = $response.policies | Where-Object { $_.policy_id -eq "default_ai_dlp_policy" }
        if (-not $defaultPolicy) { throw "Default AI DLP policy not found" }
        if (-not $defaultPolicy.is_active) { throw "Default policy should be active" }
        if (-not $defaultPolicy.categories) { throw "Default policy missing categories" }
        if (-not $defaultPolicy.severity_thresholds) { throw "Default policy missing severity thresholds" }
        if (-not $defaultPolicy.actions) { throw "Default policy missing actions" }
    }
}

function Test-AIDLPMetrics {
    Write-TestLog "=== Testing AI DLP Metrics ===" "INFO"
    
    # Test 1: Get AI DLP Metrics
    Test-Endpoint -Name "Get AI DLP Metrics" -Method "GET" -Url "/api/ai-dlp/metrics" -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to get AI DLP metrics" }
        if (-not $response.metrics) { throw "No metrics returned" }
        if (-not $response.metrics.model_version) { throw "Model version not returned" }
        if (-not $response.metrics.last_updated) { throw "Last updated timestamp not returned" }
        
        # Verify metric fields exist
        $requiredFields = @("total_scans", "average_time", "accuracy", "false_positives", "false_negatives", "overrides", "blocked_content", "warned_content", "allowed_content")
        foreach ($field in $requiredFields) {
            if (-not ($response.metrics.PSObject.Properties.Name -contains $field)) {
                throw "Missing metric field: $field"
            }
        }
    }
}

function Test-AIDLPIntegration {
    Write-TestLog "=== Testing AI DLP Integration ===" "INFO"
    
    # Test 1: Complete AI DLP workflow
    Write-TestLog "Testing complete AI DLP workflow: Scan -> Classification -> Override -> Metrics" "INFO"
    
    # Step 1: AI DLP scan with sensitive content
    $sensitiveContent = @{
        content = "Patient SSN: 123-45-6789, Credit Card: 4111-1111-1111-1111, Amount: $25,000"
        content_type = "email_body"
        link_id = "workflow_ai_dlp_001"
    }
    
    Test-Endpoint -Name "Workflow - AI DLP Scan" -Method "POST" -Url "/api/v/workflow_ai_dlp_001/ai-dlp/scan" -Body $sensitiveContent -Validation {
        param($response)
        if (-not $response.success) { throw "AI DLP scan failed" }
        if ($response.classification.category -eq "none") { throw "Sensitive content not detected" }
        if ($response.severity_score -lt 0.8) { throw "Severity score too low for sensitive content" }
        if ($response.risk_level -ne "critical") { throw "Risk level not critical for sensitive content" }
        if ($response.action_taken -ne "blocked") { throw "Action should be blocked for critical content" }
        if (-not $response.can_override) { throw "Override should be available for blocked content" }
    }
    
    # Step 2: Override the blocked content
    $overrideRequest = @{
        scan_id = "ai_scan_workflow_001"
        override_reason = "Emergency medical communication"
        user_id = "admin_user"
        user_role = "admin"
        justification = "This is an emergency medical communication that requires immediate sending"
    }
    
    Test-Endpoint -Name "Workflow - AI DLP Override" -Method "POST" -Url "/api/v/workflow_ai_dlp_001/ai-dlp/override" -Body $overrideRequest -Validation {
        param($response)
        if (-not $response.success) { throw "Override failed" }
        if ($response.action_taken -ne "overridden") { throw "Action not overridden" }
    }
    
    # Step 3: Check metrics were updated
    Test-Endpoint -Name "Workflow - Check Metrics" -Method "GET" -Url "/api/ai-dlp/metrics" -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to get metrics" }
        if ($response.metrics.total_scans -lt 1) { throw "Total scans not updated" }
        if ($response.metrics.overrides -lt 1) { throw "Overrides count not updated" }
    }
}

function Main {
    Write-TestLog "Starting Iteration 7 - AI-Powered DLP Integration Tests" "INFO"
    Write-TestLog "Base URL: $BaseUrl" "INFO"
    Write-TestLog "Test Email: $TestEmail" "INFO"
    
    # Run test suites
    Test-AIDLPScanning
    Test-AIDLPOverrides
    Test-AIDLPPolicies
    Test-AIDLPMetrics
    Test-AIDLPIntegration
    
    # Summary
    Write-TestLog "=== Test Summary ===" "INFO"
    Write-TestLog "Total Tests: $script:totalTests" "INFO"
    Write-TestLog "Passed: $script:passedTests" "PASS"
    Write-TestLog "Failed: $script:failedTests" $(if ($script:failedTests -gt 0) { "FAIL" } else { "INFO" })
    
    if ($script:failedTests -gt 0) {
        Write-TestLog "Some tests failed. Please check the implementation." "FAIL"
        exit 1
    } else {
        Write-TestLog "All tests passed! AI DLP implementation is working correctly." "PASS"
    }
}

Main
