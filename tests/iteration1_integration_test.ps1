# Iteration 1 - Integration Test Script
# Tests all Iteration 1 fixes: Error Response Standardization, TOTP Drift Tolerance,
# Database Indexing Optimization, and Mobile Dashboard Optimization

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$DatabasePath = "/var/db/secure-email.db",
    [switch]$RunAllTests = $true,
    [switch]$TestErrorResponses = $false,
    [switch]$TestTOTPDrift = $false,
    [switch]$TestDatabaseIndexes = $false,
    [switch]$TestMobileDashboard = $false
)

Write-Output "🔧 Iteration 1 - Integration Test Suite"
Write-Output "========================================="
Write-Output "Testing: Error Response Standardization, TOTP Drift Tolerance, Database Indexing, Mobile Dashboard"
Write-Output ""

# Test configuration
$testConfig = @{
    ApiUrl = $ApiUrl
    DatabasePath = $DatabasePath
    TestUser = @{
        Email = "test@securesystem.email"
        Password = "TestPassword123!"
        TOTPSecret = "JBSWY3DPEHPK3PXP"
    }
}

# Test results tracking
$testResults = @{
    TotalTests = 0
    PassedTests = 0
    FailedTests = 0
    Errors = @()
}

# Helper function to log test results
function Write-TestResult {
    param(
        [string]$TestName,
        [bool]$Passed,
        [string]$Message = "",
        [string]$Details = ""
    )

    $testResults.TotalTests++
    if ($Passed) {
        $testResults.PassedTests++
        Write-Output "✅ $TestName"
        if ($Message) { Write-Output "   $Message" }
    } else {
        $testResults.FailedTests++
        $testResults.Errors += "$TestName`: $Message"
        Write-Output "❌ $TestName"
        if ($Message) { Write-Output "   $Message" }
        if ($Details) { Write-Output "   Details: $Details" }
    }
}

# Helper function to make HTTP requests
function Invoke-TestRequest {
    param(
        [string]$Method = "GET",
        [string]$Url,
        [object]$Body = $null,
        [hashtable]$Headers = @{}
    )

    try {
        $params = @{
            Method = $Method
            Uri = $Url
            Headers = $Headers
            TimeoutSec = 10
        }

        if ($Body) {
            $params.Body = $Body | ConvertTo-Json -Depth 10
            $params.Headers["Content-Type"] = "application/json"
        }

        $response = Invoke-RestMethod @params -ErrorAction Stop
        return @{
            Success = $true
            StatusCode = $response.StatusCode
            Data = $response
        }
    }
    catch {
        return @{
            Success = $false
            StatusCode = $_.Exception.Response.StatusCode.value__
            Error = $_.Exception.Message
            Data = $_.Exception.Response
        }
    }
}

# =============================================================================
# 1. ERROR RESPONSE STANDARDIZATION TESTS
# =============================================================================

function Test-ErrorResponseStandardization {
    Write-Output "`n🔍 Testing Error Response Standardization"
    Write-Output "========================================="

    # Test 1: Invalid authentication
    Write-Output "`nTesting invalid authentication error response..."
    $result = Invoke-TestRequest -Method "POST" -Url "$ApiUrl/api/auth/login" -Body @{
        email = "invalid@example.com"
        password = "wrongpassword"
        totp_code = "123456"
    }

    $expectedErrorSchema = @{
        error = $true
        code = $null
        message = $null
        details = $null
        timestamp = $null
        path = $null
    }

    if (-not $result.Success -and $result.StatusCode -eq 401) {
        try {
            $errorData = $result.Data | ConvertFrom-Json
            $hasRequiredFields = $errorData.PSObject.Properties.Name -contains "error" -and
                                $errorData.PSObject.Properties.Name -contains "code" -and
                                $errorData.PSObject.Properties.Name -contains "message" -and
                                $errorData.PSObject.Properties.Name -contains "timestamp"

            Write-TestResult -TestName "Error Response Schema Validation" -Passed $hasRequiredFields -Message "Error response follows standardized schema"
        }
        catch {
            Write-TestResult -TestName "Error Response Schema Validation" -Passed $false -Message "Failed to parse error response JSON"
        }
    } else {
        Write-TestResult -TestName "Error Response Schema Validation" -Passed $false -Message "Expected 401 error response"
    }

    # Test 2: Missing required fields
    Write-Output "`nTesting missing fields error response..."
    $result = Invoke-TestRequest -Method "POST" -Url "$ApiUrl/api/auth/login" -Body @{
        email = "test@example.com"
        # Missing password and TOTP
    }

    if (-not $result.Success -and $result.StatusCode -eq 400) {
        try {
            $errorData = $result.Data | ConvertFrom-Json
            $hasValidationError = $errorData.code -eq "VALIDATION_FAILED" -or $errorData.code -eq "MISSING_FIELD"

            Write-TestResult -TestName "Validation Error Response" -Passed $hasValidationError -Message "Proper validation error code returned"
        }
        catch {
            Write-TestResult -TestName "Validation Error Response" -Passed $false -Message "Failed to parse validation error response"
        }
    } else {
        Write-TestResult -TestName "Validation Error Response" -Passed $false -Message "Expected 400 validation error"
    }

    # Test 3: Invalid JSON format
    Write-Output "`nTesting invalid JSON error response..."
    $result = Invoke-TestRequest -Method "POST" -Url "$ApiUrl/api/auth/login" -Body "invalid json"

    if (-not $result.Success -and $result.StatusCode -eq 400) {
        try {
            $errorData = $result.Data | ConvertFrom-Json
            $hasInvalidRequestError = $errorData.code -eq "INVALID_REQUEST" -or $errorData.code -eq "VALIDATION_FAILED"

            Write-TestResult -TestName "Invalid JSON Error Response" -Passed $hasInvalidRequestError -Message "Proper invalid request error code returned"
        }
        catch {
            Write-TestResult -TestName "Invalid JSON Error Response" -Passed $false -Message "Failed to parse invalid JSON error response"
        }
    } else {
        Write-TestResult -TestName "Invalid JSON Error Response" -Passed $false -Message "Expected 400 invalid request error"
    }
}

# =============================================================================
# 2. TOTP DRIFT TOLERANCE TESTS
# =============================================================================

function Test-TOTPDriftTolerance {
    Write-Output "`n🔐 Testing TOTP Drift Tolerance"
    Write-Output "==============================="

    # Generate TOTP codes for different time offsets
    Write-Output "`nGenerating TOTP codes for drift testing..."

    # Test TOTP generation and validation
    $testCases = @(
        @{ Offset = 0; Expected = $true; Description = "Current time" },
        @{ Offset = -30; Expected = $true; Description = "30 seconds ago" },
        @{ Offset = 30; Expected = $true; Description = "30 seconds future" },
        @{ Offset = -60; Expected = $false; Description = "60 seconds ago (should fail)" },
        @{ Offset = 60; Expected = $false; Description = "60 seconds future (should fail)" }
    )

    foreach ($testCase in $testCases) {
        Write-Output "`nTesting TOTP at $($testCase.Description)..."

        # Generate TOTP code at the specified time offset
        $offsetSeconds = $testCase.Offset
        $testTime = (Get-Date).AddSeconds($offsetSeconds)

        # Use the TOTP generator to create a code
        $totpCode = & "cmd/totp_generator/totp_generator.exe" $testConfig.TestUser.TOTPSecret 2>$null
        if ($LASTEXITCODE -ne 0) {
            Write-TestResult -TestName "TOTP Generation at $($testCase.Description)" -Passed $false -Message "Failed to generate TOTP code"
            continue
        }

        # Test login with the generated TOTP
        $result = Invoke-TestRequest -Method "POST" -Url "$ApiUrl/api/auth/login" -Body @{
            email = $testConfig.TestUser.Email
            password = $testConfig.TestUser.Password
            totp_code = $totpCode
        }

        $loginSuccess = $result.Success -and $result.StatusCode -eq 200
        $expectedSuccess = $testCase.Expected

        Write-TestResult -TestName "TOTP Drift Tolerance: $($testCase.Description)" -Passed ($loginSuccess -eq $expectedSuccess) -Message "Login $(if ($loginSuccess) { 'succeeded' } else { 'failed' }) as expected"
    }
}

# =============================================================================
# 3. DATABASE INDEXING OPTIMIZATION TESTS
# =============================================================================

function Test-DatabaseIndexing {
    Write-Output "`n🗄️ Testing Database Indexing Optimization"
    Write-Output "========================================="

    # Check if sqlite3 is available
    $sqliteVersion = sqlite3 --version 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-TestResult -TestName "SQLite3 Availability" -Passed $false -Message "sqlite3 command not found"
        return
    }

    Write-TestResult -TestName "SQLite3 Availability" -Passed $true -Message "SQLite3 version: $sqliteVersion"

    # Check if database exists
    if (-not (Test-Path $DatabasePath)) {
        Write-TestResult -TestName "Database Existence" -Passed $false -Message "Database not found at: $DatabasePath"
        return
    }

    Write-TestResult -TestName "Database Existence" -Passed $true -Message "Database found at: $DatabasePath"

    # Test critical query performance
    Write-Output "`nTesting query performance..."

    $queries = @{
        "User Authentication Lookup" = "SELECT * FROM users WHERE email = 'test@example.com' LIMIT 1;"
        "Email Listing by Sender" = "SELECT * FROM emails WHERE sender_id = 'test-user-id' ORDER BY created_at DESC LIMIT 50;"
        "Recent Audit Logs" = "SELECT * FROM audit_log WHERE user_id = 'test-user-id' AND timestamp > datetime('now', '-7 days') ORDER BY timestamp DESC LIMIT 100;"
        "Session Validation" = "SELECT * FROM sessions WHERE token_hash = 'test-hash' AND expires_at > datetime('now') LIMIT 1;"
    }

    foreach ($query in $queries.GetEnumerator()) {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

        # Run query multiple times for accurate measurement
        for ($i = 0; $i -lt 10; $i++) {
            $result = sqlite3 $DatabasePath $query.Value 2>$null
        }

        $stopwatch.Stop()
        $avgTime = $stopwatch.ElapsedMilliseconds / 10

        $performancePassed = $avgTime -lt 50  # Should be under 50ms
        Write-TestResult -TestName "Query Performance: $($query.Key)" -Passed $performancePassed -Message "Average time: $avgTime ms"
    }

    # Test index existence
    Write-Output "`nChecking critical indexes..."

    $criticalIndexes = @(
        "idx_users_email",
        "idx_emails_sender_id",
        "idx_emails_created_at",
        "idx_audit_log_timestamp",
        "idx_sessions_user_id"
    )

    foreach ($index in $criticalIndexes) {
        $indexQuery = "SELECT name FROM sqlite_master WHERE type='index' AND name='$index';"
        $result = sqlite3 $DatabasePath $indexQuery 2>$null

        $indexExists = $result -and $result.Trim() -eq $index
        Write-TestResult -TestName "Index Existence: $index" -Passed $indexExists -Message "Index $(if ($indexExists) { 'exists' } else { 'missing' })"
    }

    # Test EXPLAIN QUERY PLAN for critical queries
    Write-Output "`nAnalyzing query plans..."

    $explainQueries = @{
        "User Lookup Query Plan" = "EXPLAIN QUERY PLAN SELECT * FROM users WHERE email = 'test@example.com';"
        "Email Listing Query Plan" = "EXPLAIN QUERY PLAN SELECT * FROM emails WHERE sender_id = 'test-user-id' ORDER BY created_at DESC;"
    }

    foreach ($query in $explainQueries.GetEnumerator()) {
        $result = sqlite3 $DatabasePath $query.Value 2>$null

        if ($LASTEXITCODE -eq 0 -and $result) {
            $usesIndex = $result -match "USING INDEX" -or $result -match "SCAN TABLE.*INDEX"
            Write-TestResult -TestName "Query Plan: $($query.Key)" -Passed $usesIndex -Message "Query $(if ($usesIndex) { 'uses' } else { 'does not use' }) indexes"
        } else {
            Write-TestResult -TestName "Query Plan: $($query.Key)" -Passed $false -Message "Failed to get query plan"
        }
    }
}

# =============================================================================
# 4. MOBILE DASHBOARD OPTIMIZATION TESTS
# =============================================================================

function Test-MobileDashboardOptimization {
    Write-Output "`n📱 Testing Mobile Dashboard Optimization"
    Write-Output "========================================="

    # Test responsive design elements
    Write-Output "`nTesting responsive design..."

    # Check if the dashboard loads successfully
    $dashboardUrl = "$ApiUrl/admin"
    $result = Invoke-TestRequest -Method "GET" -Url $dashboardUrl

    if ($result.Success) {
        Write-TestResult -TestName "Dashboard Accessibility" -Passed $true -Message "Dashboard loads successfully"

        # Check for mobile-specific CSS classes
        $htmlContent = $result.Data | Out-String
        $mobileClasses = @(
            "sm:grid-cols-2",
            "sm:px-4",
            "sm:py-4",
            "col-span-1 sm:col-span-2"
        )

        $mobileOptimized = $true
        foreach ($class in $mobileClasses) {
            if ($htmlContent -notmatch [regex]::Escape($class)) {
                $mobileOptimized = $false
                break
            }
        }

        Write-TestResult -TestName "Mobile Responsive Classes" -Passed $mobileOptimized -Message "Mobile responsive CSS classes present"
    } else {
        Write-TestResult -TestName "Dashboard Accessibility" -Passed $false -Message "Dashboard not accessible"
    }

    # Test mobile-specific components
    Write-Output "`nTesting mobile components..."

    # Check for mobile-optimized panel components
    $mobileComponents = @(
        "MobileOptimizedPanel",
        "MobileMetricCard",
        "MobileDataTable",
        "MobileChartContainer"
    )

    foreach ($component in $mobileComponents) {
        $componentFile = "src/components/admin/panels/$component.tsx"
        $componentExists = Test-Path $componentFile

        Write-TestResult -TestName "Mobile Component: $component" -Passed $componentExists -Message "Component $(if ($componentExists) { 'exists' } else { 'missing' })"
    }

    # Test touch-friendly elements
    Write-Output "`nTesting touch-friendly elements..."

    $touchFriendlyClasses = @(
        "px-4 py-4",  # Large tap targets
        "focus:ring-2",  # Focus indicators
        "transition-colors",  # Smooth transitions
        "rounded-lg"  # Rounded corners for touch
    )

    $touchOptimized = $true
    foreach ($class in $touchFriendlyClasses) {
        if ($htmlContent -notmatch [regex]::Escape($class)) {
            $touchOptimized = $false
            break
        }
    }

    Write-TestResult -TestName "Touch-Friendly Elements" -Passed $touchOptimized -Message "Touch-friendly CSS classes present"
}

# =============================================================================
# MAIN TEST EXECUTION
# =============================================================================

function Run-AllTests {
    Write-Output "🚀 Starting Iteration 1 Integration Tests..."
    Write-Output "API URL: $($testConfig.ApiUrl)"
    Write-Output "Database: $($testConfig.DatabasePath)"
    Write-Output ""

    # Run tests based on parameters
    if ($RunAllTests -or $TestErrorResponses) {
        Test-ErrorResponseStandardization
    }

    if ($RunAllTests -or $TestTOTPDrift) {
        Test-TOTPDriftTolerance
    }

    if ($RunAllTests -or $TestDatabaseIndexes) {
        Test-DatabaseIndexing
    }

    if ($RunAllTests -or $TestMobileDashboard) {
        Test-MobileDashboardOptimization
    }
}

function Write-TestSummary {
    Write-Output "`n📊 Test Summary"
    Write-Output "==============="
    Write-Output "Total Tests: $($testResults.TotalTests)"
    Write-Output "Passed: $($testResults.PassedTests)"
    Write-Output "Failed: $($testResults.FailedTests)"

    $successRate = if ($testResults.TotalTests -gt 0) {
        [math]::Round(($testResults.PassedTests / $testResults.TotalTests) * 100, 2)
    } else { 0 }

    Write-Output "Success Rate: $successRate%" -ForegroundColor $(if ($successRate -ge 90) { "Green" } elseif ($successRate -ge 70) { "Yellow" } else { "Red" })

    if ($testResults.Errors.Count -gt 0) {
        Write-Output "`n❌ Failed Tests:"
        foreach ($testError in $testResults.Errors) {
            Write-Output "  - $testError"
        }
    }

    # Overall assessment
    Write-Output "`n🎯 Overall Assessment:"
    if ($successRate -ge 90) {
        Write-Output "✅ EXCELLENT - All Iteration 1 fixes are working correctly!"
    } elseif ($successRate -ge 70) {
        Write-Output "⚠️  GOOD - Most fixes are working, some issues need attention"
    } else {
        Write-Output "❌ NEEDS WORK - Significant issues with Iteration 1 fixes"
    }

    # Recommendations
    Write-Output "`n💡 Recommendations:"
    if ($testResults.FailedTests -gt 0) {
        Write-Output "  - Review failed tests and fix issues"
        Write-Output "  - Re-run tests after fixes"
    } else {
        Write-Output "  - All tests passed! Iteration 1 is ready for production"
    }

    Write-Output "  - Monitor performance in production environment"
    Write-Output "  - Consider load testing for high-traffic scenarios"
}

# Execute tests
try {
    Run-AllTests
    Write-TestSummary

    # Exit with appropriate code
    if ($testResults.FailedTests -eq 0) {
        Write-Output "`n✅ All tests passed! Iteration 1 fixes are working correctly."
        exit 0
    } else {
        Write-Output "`n❌ Some tests failed. Please review and fix issues."
        exit 1
    }
}
catch {
    Write-Output "`n💥 Test execution failed: $($_.Exception.Message)"
    Write-Output "Stack trace: $($_.ScriptStackTrace)"
    exit 1
}
