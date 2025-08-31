#!/usr/bin/env pwsh

# Iteration 6 - Advanced Security & Compliance Integration Test
# Tests DLP scanning, watermarking, security policies, and compliance audit logging

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
        default { "White" }
    }
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $color
}

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [hashtable]$Headers = @{},
        [string]$Body = $null,
        [scriptblock]$Validation
    )
    
    $script:totalTests++
    Write-TestLog "Running test: $Name"
    
    try {
        $params = @{
            Method = $Method
            Uri = "$BaseUrl$Url"
            Headers = $Headers
            ContentType = "application/json"
        }
        
        if ($Body) {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params -ErrorAction Stop
        
        if ($Validation) {
            & $Validation $response
            $script:passedTests++
            Write-TestLog "✓ $Name passed" "PASS"
        } else {
            $script:passedTests++
            Write-TestLog "✓ $Name passed" "PASS"
        }
    }
    catch {
        $script:failedTests++
        Write-TestLog "✗ $Name failed: $($_.Exception.Message)" "FAIL"
        if ($Verbose) {
            Write-TestLog "Response: $($_.Exception.Response)" "WARN"
        }
    }
}

function Test-DLPScanning {
    Write-TestLog "=== Testing DLP Scanning ===" "INFO"
    
    # Test 1: Scan content with credit card number
    $creditCardContent = @{
        content = "My credit card number is 4111-1111-1111-1111"
        content_type = "email_body"
        link_id = "test_link_001"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "DLP Scan - Credit Card Detection" -Method "POST" -Url "/api/v/test_link_001/dlp/scan" -Body $creditCardContent -Validation {
        param($response)
        if (-not $response.success) { throw "DLP scan failed" }
        if ($response.action_taken -ne "warned" -and $response.action_taken -ne "blocked") { throw "Expected warning or block for credit card" }
        if ($response.violations.Count -eq 0) { throw "No violations detected" }
    }
    
    # Test 2: Scan content with SSN
    $ssnContent = @{
        content = "My SSN is 123-45-6789"
        content_type = "reply_body"
        link_id = "test_link_002"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "DLP Scan - SSN Detection" -Method "POST" -Url "/api/v/test_link_002/dlp/scan" -Body $ssnContent -Validation {
        param($response)
        if (-not $response.success) { throw "DLP scan failed" }
        if ($response.violations.Count -eq 0) { throw "No violations detected" }
    }
    
    # Test 3: Scan clean content
    $cleanContent = @{
        content = "This is a normal message without sensitive data"
        content_type = "email_body"
        link_id = "test_link_003"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "DLP Scan - Clean Content" -Method "POST" -Url "/api/v/test_link_003/dlp/scan" -Body $cleanContent -Validation {
        param($response)
        if (-not $response.success) { throw "DLP scan failed" }
        if ($response.action_taken -ne "allowed") { throw "Expected allowed for clean content" }
    }
}

function Test-SecurityPolicies {
    Write-TestLog "=== Testing Security Policies ===" "INFO"
    
    # Test 1: Create security policy
    $policy = @{
        link_id = "test_link_004"
        dlp_enabled = $true
        watermark_enabled = $true
        download_disabled = $false
        forwarding_disabled = $true
        auto_revoke_after_reply = $false
        max_views = 5
        notify_on_expiry = $true
        notify_on_revoke = $true
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Create Security Policy" -Method "POST" -Url "/api/v/test_link_004/security/policy" -Body $policy -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to create security policy" }
        if (-not $response.policy_id) { throw "No policy ID returned" }
    }
    
    # Test 2: Get security policy
    Test-Endpoint -Name "Get Security Policy" -Method "GET" -Url "/api/v/test_link_004/security/policy" -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to get security policy" }
        if (-not $response.policy) { throw "No policy returned" }
    }
    
    # Test 3: Update security policy
    $updatedPolicy = @{
        policy_id = "policy_test_001"
        link_id = "test_link_004"
        dlp_enabled = $true
        watermark_enabled = $true
        download_disabled = $true
        forwarding_disabled = $true
        auto_revoke_after_reply = $true
        max_views = 3
        notify_on_expiry = $true
        notify_on_revoke = $true
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Update Security Policy" -Method "PUT" -Url "/api/v/test_link_004/security/policy" -Body $updatedPolicy -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to update security policy" }
    }
}

function Test-SecurityTemplates {
    Write-TestLog "=== Testing Security Policy Templates ===" "INFO"
    
    Test-Endpoint -Name "Get Security Templates" -Method "GET" -Url "/api/security/templates" -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to get security templates" }
        if (-not $response.templates) { throw "No templates returned" }
        if ($response.templates.Count -eq 0) { throw "No templates available" }
        
        # Check for default template
        $defaultTemplate = $response.templates | Where-Object { $_.is_default -eq $true }
        if (-not $defaultTemplate) { throw "No default template found" }
    }
}

function Test-Watermarking {
    Write-TestLog "=== Testing Watermarking ===" "INFO"
    
    # Test 1: Apply watermark to attachment
    $watermarkRequest = @{
        attachment_id = "att_test_001"
        watermark_text = "Confidential - {recipient_email} - {timestamp}"
        watermark_position = "bottom-right"
        watermark_opacity = 0.7
        watermark_font_size = 12
        watermark_color = "#FF0000"
        watermark_rotation = -45
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Apply Watermark" -Method "POST" -Url "/api/v/test_link_005/watermark" -Body $watermarkRequest -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to apply watermark" }
        if (-not $response.config_id) { throw "No config ID returned" }
    }
}

function Test-AccessControl {
    Write-TestLog "=== Testing Access Control ===" "INFO"
    
    # Test 1: Check access control
    $accessRequest = @{
        user_id = "test_user_001"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Check Access Control" -Method "POST" -Url "/api/v/test_link_004/security/access" -Body $accessRequest -Validation {
        param($response)
        if (-not $response.success) { throw "Access control check failed" }
    }
}

function Test-ComplianceAudit {
    Write-TestLog "=== Testing Compliance Audit Logging ===" "INFO"
    
    # Test 1: Log compliance audit event
    $auditEvent = @{
        event_type = "dlp_scan"
        link_id = "test_link_006"
        user_id = "test_user_002"
        event_details = @{
            content_type = "email_body"
            violations_count = 2
            action_taken = "warned"
        }
        severity = "warning"
        compliance_category = "dlp"
        retention_required = $true
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Log Compliance Audit Event" -Method "POST" -Url "/api/compliance/audit" -Body $auditEvent -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to log audit event" }
        if (-not $response.audit_id) { throw "No audit ID returned" }
    }
    
    # Test 2: Log policy enforcement event
    $policyEvent = @{
        event_type = "policy_enforced"
        link_id = "test_link_007"
        policy_id = "policy_test_002"
        user_id = "test_user_003"
        event_details = @{
            access_granted = $true
            policy_applied = "high_security"
        }
        severity = "info"
        compliance_category = "access_control"
        retention_required = $true
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Log Policy Enforcement Event" -Method "POST" -Url "/api/compliance/audit" -Body $policyEvent -Validation {
        param($response)
        if (-not $response.success) { throw "Failed to log policy enforcement event" }
    }
}

function Test-IntegrationScenarios {
    Write-TestLog "=== Testing Integration Scenarios ===" "INFO"
    
    # Test 1: Complete workflow with DLP and security policy
    Write-TestLog "Testing complete workflow: DLP scan -> Policy enforcement -> Audit logging" "INFO"
    
    # Step 1: Create security policy
    $workflowPolicy = @{
        link_id = "workflow_link_001"
        dlp_enabled = $true
        watermark_enabled = $true
        download_disabled = $false
        forwarding_disabled = $true
        auto_revoke_after_reply = $false
        max_views = 10
        notify_on_expiry = $true
        notify_on_revoke = $true
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Workflow - Create Policy" -Method "POST" -Url "/api/v/workflow_link_001/security/policy" -Body $workflowPolicy -Validation {
        param($response)
        if (-not $response.success) { throw "Workflow step 1 failed" }
    }
    
    # Step 2: DLP scan
    $workflowContent = @{
        content = "Please send payment to account 1234-5678-9012-3456"
        content_type = "email_body"
        link_id = "workflow_link_001"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Workflow - DLP Scan" -Method "POST" -Url "/api/v/workflow_link_001/dlp/scan" -Body $workflowContent -Validation {
        param($response)
        if (-not $response.success) { throw "Workflow step 2 failed" }
    }
    
    # Step 3: Access control check
    $workflowAccess = @{
        user_id = "workflow_user_001"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Workflow - Access Control" -Method "POST" -Url "/api/v/workflow_link_001/security/access" -Body $workflowAccess -Validation {
        param($response)
        if (-not $response.success) { throw "Workflow step 3 failed" }
    }
}

function Show-TestResults {
    Write-TestLog "=== Test Results Summary ===" "INFO"
    Write-TestLog "Total Tests: $script:totalTests" "INFO"
    Write-TestLog "Passed: $script:passedTests" "PASS"
    Write-TestLog "Failed: $script:failedTests" "FAIL"
    
    $successRate = if ($script:totalTests -gt 0) { [math]::Round(($script:passedTests / $script:totalTests) * 100, 2) } else { 0 }
    Write-TestLog "Success Rate: $successRate%" "INFO"
    
    if ($script:failedTests -eq 0) {
        Write-TestLog "🎉 All tests passed! Iteration 6 features are working correctly." "PASS"
        exit 0
    } else {
        Write-TestLog "❌ Some tests failed. Please review the errors above." "FAIL"
        exit 1
    }
}

# Main execution
function Main {
    Write-TestLog "Starting Iteration 6 - Advanced Security & Compliance Integration Tests" "INFO"
    Write-TestLog "Base URL: $BaseUrl" "INFO"
    Write-TestLog "Test Email: $TestEmail" "INFO"
    
    # Run all test suites
    Test-DLPScanning
    Test-SecurityPolicies
    Test-SecurityTemplates
    Test-Watermarking
    Test-AccessControl
    Test-ComplianceAudit
    Test-IntegrationScenarios
    
    Show-TestResults
}

# Run the tests
Main
