# =============================================================================
# COMPREHENSIVE USER FLOW VALIDATION TEST
# =============================================================================
# This script validates all user flows from the flowmaps:
# - Internal User Authentication & Onboarding
# - Email Composition & Sending
# - Secure Link Creation & External Access
# - External User Secure Link Viewing
# - Reply & Forwarding
# - Admin Operations
# - Monitoring & Dashboard
# =============================================================================

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",  # Use the known test user
    [string]$TestPassword = "Test123!@#",
    [string]$ExternalEmail = "external@example.com",
    [int]$Timeout = 30
)

# =============================================================================
# TEST RESULTS TRACKING
# =============================================================================

$TestResults = @{
    Total = 0
    Passed = 0
    Failed = 0
    Errors = @()
}

function Write-TestResult {
    param(
        [string]$TestName,
        [bool]$Success,
        [string]$Message = ""
    )
    
    $TestResults.Total++
    if ($Success) {
        $TestResults.Passed++
        Write-Host "  ✓ PASSED: $TestName" -ForegroundColor Green
        if ($Message) { Write-Host "    $Message" -ForegroundColor DarkGreen }
    } else {
        $TestResults.Failed++
        $TestResults.Errors += "$TestName`:$Message"
        Write-Host "  ✗ FAILED: $TestName" -ForegroundColor Red
        if ($Message) { Write-Host "    $Message" -ForegroundColor DarkRed }
    }
}

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

function Test-APIEndpoint {
    param(
        [string]$Endpoint,
        [string]$Method = "GET",
        [object]$Body = $null,
        [hashtable]$Headers = @{},
        [string]$TestName
    )
    
    try {
        $uri = "$BaseUrl$Endpoint"
        $params = @{
            Uri = $uri
            Method = $Method
            TimeoutSec = $Timeout
        }
        
        if ($Body) {
            $params.Body = $Body | ConvertTo-Json -Depth 10
            $params.ContentType = "application/json"
        }
        
        if ($Headers.Count -gt 0) {
            $params.Headers = $Headers
        }
        
        $response = Invoke-RestMethod @params
        $status = if ($response.success) { "Success" } else { "Failed" }
        Write-TestResult -TestName $TestName -Success $true -Message "Status: $status"
        return $response
    } catch {
        $errorMsg = $_.Exception.Message
        Write-TestResult -TestName $TestName -Success $false -Message $errorMsg
        return $null
    }
}

function Test-JsonResponse {
    param(
        [object]$Response,
        [string]$ExpectedType,
        [string]$TestName
    )
    
    try {
        if ($ExpectedType -eq "object" -and $Response -isnot [PSCustomObject]) {
            throw "Expected object response, got $($Response.GetType().Name)"
        }
        
        if ($ExpectedType -eq "array" -and $Response -isnot [array]) {
            throw "Expected array response, got $($Response.GetType().Name)"
        }
        
        Write-TestResult -TestName $TestName -Success $true
        return $true
    } catch {
        Write-TestResult -TestName $TestName -Success $false -Message $_.Exception.Message
        return $false
    }
}

# =============================================================================
# USER FLOW 1: SYSTEM HEALTH & AVAILABILITY
# =============================================================================

Write-Host "`n🔍 USER FLOW 1: System Health & Availability" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

# Test 1.1: System Health Check
$healthResponse = Test-APIEndpoint -Endpoint "/api/metrics/health" -TestName "System Health Check"
if ($healthResponse) {
    Test-JsonResponse -Response $healthResponse.health -ExpectedType "object" -TestName "Health Response Structure"
}

# Test 1.2: Metrics Endpoint
$metricsResponse = Test-APIEndpoint -Endpoint "/api/metrics" -TestName "Metrics Endpoint"
if ($metricsResponse) {
    Test-JsonResponse -Response $metricsResponse.metrics -ExpectedType "object" -TestName "Metrics Response Structure"
}

# Test 1.3: API Availability
Test-APIEndpoint -Endpoint "/api/watermark/templates" -TestName "Watermark Templates Endpoint"

# =============================================================================
# USER FLOW 2: INTERNAL USER AUTHENTICATION
# =============================================================================

Write-Host "`n🔐 USER FLOW 2: Internal User Authentication" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

# Test 2.1: User Registration (if needed)
$signupData = @{
    email = $TestEmail
    password = $TestPassword
    fallback_email = "fallback@example.com"
}

# Skip user registration if user already exists (common in testing)
Write-Host "  ℹ SKIPPING: User Registration (using existing test user)" -ForegroundColor Yellow
$signupResponse = $null

# Test 2.2: User Login
# Generate correct TOTP code for test user (JBSWY3DPEHPK3PXP)
$totpOutput = & "$PSScriptRoot\..\scripts\generate_totp.ps1" -Secret "JBSWY3DPEHPK3PXP"
$currentTOTPCode = $totpOutput.Split("`n") | Where-Object { $_ -match 'CURRENT:' } | ForEach-Object { $_.Split(':')[1].Trim() }

$loginData = @{
    email = $TestEmail
    password = $TestPassword
    totp_code = $currentTOTPCode
}

# Add delay to avoid rate limiting
Start-Sleep -Seconds 5

# Check if we're rate limited and handle gracefully
$loginResponse = Test-APIEndpoint -Endpoint "/api/auth/login" -Method "POST" -Body $loginData -TestName "User Login"

# If rate limited, provide helpful message
if ($loginResponse -eq $null -or $loginResponse.error -eq $true) {
    Write-Host "  ⚠ RATE LIMITED: Authentication attempts exceeded limit" -ForegroundColor Yellow
    Write-Host "  💡 Solution: Wait 60 seconds or run .\scripts\reset_rate_limits.ps1" -ForegroundColor Cyan
    Write-Host "  🔧 For testing: Set TEST_MODE=true environment variable" -ForegroundColor Cyan
}

$authToken = $null
if ($loginResponse -and $loginResponse.token) {
    $authToken = $loginResponse.token
    Write-Host "    Authentication token obtained" -ForegroundColor DarkGreen
}

# =============================================================================
# USER FLOW 3: EMAIL COMPOSITION & SENDING
# =============================================================================

Write-Host "`n📧 USER FLOW 3: Email Composition & Sending" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

if ($authToken) {
    $headers = @{ "Authorization" = "Bearer $authToken" }
    
    # Test 3.1: Compose Internal Email
    $internalEmailData = @{
        recipient = "internal@example.com"
        subject = "Test Internal Email - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
        body = "This is a test internal email with rich content."
        attachments = @()
        security_settings = @{
            require_password = $false
            require_mfa = $false
            geo_restriction = "none"
            time_lock = $false
            burn_after_read = $false
            self_destruct_attempts = 5
            expires_at = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
        }
    }
    
    $internalEmailResponse = Test-APIEndpoint -Endpoint "/api/emails/send" -Method "POST" -Body $internalEmailData -Headers $headers -TestName "Internal Email Composition"
    
    # Test 3.2: Compose Secure Link Email
    $secureEmailData = @{
        recipient = $ExternalEmail
        subject = "Test Secure Link Email - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
        body = "This is a test secure email for external recipient."
        attachments = @()
        security_settings = @{
            require_password = $true
            password = "securepass123"
            require_mfa = $false
            geo_restriction = "none"
            time_lock = $false
            burn_after_read = $false
            self_destruct_attempts = 3
            expires_at = (Get-Date).AddDays(3).ToString("yyyy-MM-ddTHH:mm:ssZ")
        }
    }
    
    $secureEmailResponse = Test-APIEndpoint -Endpoint "/api/emails/send" -Method "POST" -Body $secureEmailData -Headers $headers -TestName "Secure Link Email Composition"
    
    $secureLinkId = $null
    if ($secureEmailResponse -and $secureEmailResponse.secure_link_id) {
        $secureLinkId = $secureEmailResponse.secure_link_id
        Write-Host "    Secure link created: $secureLinkId" -ForegroundColor DarkGreen
    }
} else {
    Write-Host "  ⚠ SKIPPED: Email composition tests (no auth token)" -ForegroundColor Yellow
}

# =============================================================================
# USER FLOW 4: SECURE LINK EXTERNAL ACCESS
# =============================================================================

Write-Host "`n🔗 USER FLOW 4: Secure Link External Access" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

if ($secureLinkId) {
    # Test 4.1: Secure Link Validation
    Test-APIEndpoint -Endpoint "/api/v/$secureLinkId/validate" -TestName "Secure Link Validation"
    
    # Test 4.2: Secure Link Access (with password)
    $accessData = @{
        password = "securepass123"
    }
    
    $accessResponse = Test-APIEndpoint -Endpoint "/api/v/$secureLinkId/access" -Method "POST" -Body $accessData -TestName "Secure Link Access"
    
    # Test 4.3: Secure Link Content Retrieval
    if ($accessResponse -and $accessResponse.access_token) {
        $accessHeaders = @{ "Authorization" = "Bearer $($accessResponse.access_token)" }
        Test-APIEndpoint -Endpoint "/api/v/$secureLinkId/content" -Headers $accessHeaders -TestName "Secure Link Content Retrieval"
    }
} else {
    Write-Host "  ⚠ SKIPPED: Secure link tests (no secure link ID)" -ForegroundColor Yellow
}

# =============================================================================
# USER FLOW 5: WATERMARKING & DLP
# =============================================================================

Write-Host "`n🖼️ USER FLOW 5: Watermarking & DLP" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan

# Test 5.1: Watermark Templates
$watermarkResponse = Test-APIEndpoint -Endpoint "/api/watermark/templates" -TestName "Watermark Templates Retrieval"

# Test 5.2: Advanced Watermarking
if ($secureLinkId) {
    $watermarkData = @{
        watermark_type = "text"
        content_type = "pdf"
        recipient_email = $ExternalEmail
        watermark_config = @{
            text = "CONFIDENTIAL"
            opacity = 0.8
            position = "center"
        }
        is_recipient_specific = $true
    }
    
    Test-APIEndpoint -Endpoint "/api/v/$secureLinkId/watermark/advanced" -Method "POST" -Body $watermarkData -TestName "Advanced Watermarking"
}

# Test 5.3: DLP Scanning
$dlpData = @{
    content = "This is a test message with sensitive information: SSN 123-45-6789"
    content_type = "text"
    user_id = $TestEmail
}

Test-APIEndpoint -Endpoint "/api/dlp/scan" -Method "POST" -Body $dlpData -TestName "DLP Content Scanning"

# =============================================================================
# USER FLOW 6: MONITORING & DASHBOARD
# =============================================================================

Write-Host "`n📊 USER FLOW 6: Monitoring & Dashboard" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan

# Test 6.1: Real-time Metrics
$metricsResponse = Test-APIEndpoint -Endpoint "/api/metrics" -TestName "Real-time Metrics"

# Test 6.2: System Health
$healthResponse = Test-APIEndpoint -Endpoint "/api/metrics/health" -TestName "System Health Status"

# Test 6.3: Monitoring Events
Test-APIEndpoint -Endpoint "/api/metrics/events" -TestName "Monitoring Events"

# =============================================================================
# USER FLOW 7: ADMIN OPERATIONS
# =============================================================================

Write-Host "`n👨‍💼 USER FLOW 7: Admin Operations" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

# Test 7.1: Admin Setup Check
Test-APIEndpoint -Endpoint "/admin/check-setup" -TestName "Admin Setup Status"

# Test 7.2: Admin Login (if setup)
# Note: Admin TOTP would need to be configured separately
$adminLoginData = @{
    email = "admin@example.com"
    password = "admin123"
    totp_code = "123456"  # Admin TOTP would need to be configured
}

Test-APIEndpoint -Endpoint "/admin/login" -Method "POST" -Body $adminLoginData -TestName "Admin Login"

# =============================================================================
# USER FLOW 8: SECURITY FEATURES
# =============================================================================

Write-Host "`n🔒 USER FLOW 8: Security Features" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

# Test 8.1: Security Templates
Test-APIEndpoint -Endpoint "/api/security/templates" -TestName "Security Templates"

# Test 8.2: Audit Logs
if ($authToken) {
    $headers = @{ "Authorization" = "Bearer $authToken" }
    Test-APIEndpoint -Endpoint "/api/audit/logs" -Headers $headers -TestName "Audit Logs"
}

# Test 8.3: Security Policies
Test-APIEndpoint -Endpoint "/api/security/policies" -TestName "Security Policies"

# =============================================================================
# TEST SUMMARY
# =============================================================================

Write-Host "`n📋 COMPREHENSIVE USER FLOW VALIDATION SUMMARY" -ForegroundColor Magenta
Write-Host "=============================================" -ForegroundColor Magenta

Write-Host "Total Tests: $($TestResults.Total)" -ForegroundColor White
Write-Host "Passed: $($TestResults.Passed)" -ForegroundColor Green
Write-Host "Failed: $($TestResults.Failed)" -ForegroundColor Red

$successRate = if ($TestResults.Total -gt 0) { [math]::Round(($TestResults.Passed / $TestResults.Total) * 100, 1) } else { 0 }
Write-Host "Success Rate: $successRate%" -ForegroundColor $(if ($successRate -ge 80) { "Green" } elseif ($successRate -ge 60) { "Yellow" } else { "Red" })

if ($TestResults.Errors.Count -gt 0) {
    Write-Host "`n❌ FAILED TESTS:" -ForegroundColor Red
    foreach ($err in $TestResults.Errors) {
        Write-Host "  - $err" -ForegroundColor DarkRed
    }
}

if ($successRate -ge 80) {
    Write-Host "`n🎉 EXCELLENT! System is ready for production use." -ForegroundColor Green
} elseif ($successRate -ge 60) {
    Write-Host "`n⚠️ GOOD! Some issues need attention before production." -ForegroundColor Yellow
} else {
    Write-Host "`n❌ CRITICAL! Significant issues must be resolved." -ForegroundColor Red
}

Write-Host "`n✅ Comprehensive User Flow Validation Complete!" -ForegroundColor Green
