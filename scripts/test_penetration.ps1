# Penetration Testing Script for Secure Email MVP
# Tests various attack vectors and security vulnerabilities

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$OutputFile = "penetration_test_results.json",
    [switch]$Verbose
)

# Test results structure
$TestResults = @{
    timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    base_url = $BaseUrl
    tests = @()
    summary = @{
        total_tests = 0
        passed = 0
        failed = 0
        critical_findings = 0
        high_findings = 0
        medium_findings = 0
        low_findings = 0
    }
}

# Helper function to log test results
function Write-TestResult {
    param(
        [string]$TestName,
        [string]$Category,
        [string]$Severity,
        [bool]$Passed,
        [string]$Description,
        [object]$Details = $null
    )
    
    $result = @{
        test_name = $TestName
        category = $Category
        severity = $Severity
        passed = $Passed
        description = $Description
        timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        details = $Details
    }
    
    $TestResults.tests += $result
    $TestResults.summary.total_tests++
    
    if ($Passed) {
        $TestResults.summary.passed++
    } else {
        $TestResults.summary.failed++
        switch ($Severity) {
            "Critical" { $TestResults.summary.critical_findings++ }
            "High" { $TestResults.summary.high_findings++ }
            "Medium" { $TestResults.summary.medium_findings++ }
            "Low" { $TestResults.summary.low_findings++ }
        }
    }
    
    $status = if ($Passed) { "PASS" } else { "FAIL" }
    $color = if ($Passed) { "Green" } else { "Red" }
    
    Write-Host "[$status] $TestName ($Severity)" -ForegroundColor $color
    if ($Verbose -and $Details) {
        Write-Host "  Details: $($Details | ConvertTo-Json -Depth 3)" -ForegroundColor Yellow
    }
}

# Helper function to make HTTP requests
function Invoke-SecureRequest {
    param(
        [string]$Uri,
        [string]$Method = "GET",
        [hashtable]$Headers = @{},
        [string]$Body = $null,
        [int]$Timeout = 30
    )
    
    try {
        $params = @{
            Uri = $Uri
            Method = $Method
            Headers = $Headers
            TimeoutSec = $Timeout
            UseBasicParsing = $true
        }
        
        if ($Body) {
            $params.Body = $Body
        }
        
        $response = Invoke-WebRequest @params
        return @{
            success = $true
            status_code = $response.StatusCode
            content = $response.Content
            headers = $response.Headers
        }
    }
    catch {
        return @{
            success = $false
            error = $_.Exception.Message
            status_code = $_.Exception.Response.StatusCode.value__
        }
    }
}

# Test 1: SQL Injection Tests
function Test-SQLInjection {
    Write-Host "`n=== SQL Injection Tests ===" -ForegroundColor Cyan
    
    $sqlPayloads = @(
        "' OR 1=1 --",
        "'; DROP TABLE users; --",
        "' UNION SELECT * FROM users --",
        "admin'--",
        "1' OR '1'='1",
        "' AND (SELECT COUNT(*) FROM users) > 0 --"
    )
    
    $endpoints = @(
        "/api/auth/login",
        "/api/users/search",
        "/api/organizations",
        "/api/compliance/logs"
    )
    
    foreach ($endpoint in $endpoints) {
        foreach ($payload in $sqlPayloads) {
            $testName = "SQL Injection - $endpoint - $payload"
            
            $body = @{
                email = $payload
                password = "test123"
            } | ConvertTo-Json
            
            $response = Invoke-SecureRequest -Uri "$BaseUrl$endpoint" -Method "POST" -Body $body
            
            # Check for SQL error indicators
            $sqlErrorIndicators = @(
                "sqlite", "mysql", "postgresql", "oracle", "sql server",
                "syntax error", "unclosed quotation mark", "incorrect syntax",
                "ORA-", "MySQL", "SQLite", "PostgreSQL"
            )
            
            $hasError = $false
            foreach ($indicator in $sqlErrorIndicators) {
                if ($response.content -and $response.content.ToLower().Contains($indicator.ToLower())) {
                    $hasError = $true
                    break
                }
            }
            
            Write-TestResult -TestName $testName -Category "SQL Injection" -Severity "High" -Passed (-not $hasError) -Description "Tested SQL injection payload on $endpoint" -Details @{
                payload = $payload
                endpoint = $endpoint
                response_status = $response.status_code
                response_content = $response.content
                has_sql_error = $hasError
            }
        }
    }
}

# Test 2: XSS Tests
function Test-XSS {
    Write-Host "`n=== XSS Tests ===" -ForegroundColor Cyan
    
    $xssPayloads = @(
        "<script>alert('XSS')</script>",
        "javascript:alert('XSS')",
        "<img src=x onerror=alert('XSS')>",
        "<svg onload=alert('XSS')>",
        "';alert('XSS');//",
        "<iframe src=javascript:alert('XSS')>"
    )
    
    $endpoints = @(
        "/api/organizations",
        "/api/users",
        "/api/compliance/logs"
    )
    
    foreach ($endpoint in $endpoints) {
        foreach ($payload in $xssPayloads) {
            $testName = "XSS - $endpoint - $payload"
            
            $body = @{
                name = $payload
                description = $payload
            } | ConvertTo-Json
            
            $response = Invoke-SecureRequest -Uri "$BaseUrl$endpoint" -Method "POST" -Body $body
            
            # Check if payload is reflected in response
            $isReflected = $response.content -and $response.content.Contains($payload)
            
            Write-TestResult -TestName $testName -Category "XSS" -Severity "High" -Passed (-not $isReflected) -Description "Tested XSS payload on $endpoint" -Details @{
                payload = $payload
                endpoint = $endpoint
                response_status = $response.status_code
                is_reflected = $isReflected
            }
        }
    }
}

# Test 3: JWT Tampering Tests
function Test-JWTTampering {
    Write-Host "`n=== JWT Tampering Tests ===" -ForegroundColor Cyan
    
    # Test expired JWT
    $expiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
    
    $response = Invoke-SecureRequest -Uri "$BaseUrl/api/admin/organizations" -Method "GET" -Headers @{
        "Authorization" = "Bearer $expiredToken"
    }
    
    Write-TestResult -TestName "JWT Expired Token" -Category "JWT Security" -Severity "Medium" -Passed (-not $response.success) -Description "Tested expired JWT token" -Details @{
        expired_token = $expiredToken
        response_status = $response.status_code
        rejected_expired_token = -not $response.success
    }
}

# Test 4: TOTP Bypass Tests
function Test-TOTPBypass {
    Write-Host "`n=== TOTP Bypass Tests ===" -ForegroundColor Cyan
    
    $loginBody = @{
        email = "test@example.com"
        password = "test123"
    } | ConvertTo-Json
    
    # Test missing TOTP code
    $response = Invoke-SecureRequest -Uri "$BaseUrl/api/auth/login" -Method "POST" -Body $loginBody
    
    Write-TestResult -TestName "TOTP Missing Code" -Category "TOTP Security" -Severity "High" -Passed (-not $response.success) -Description "Tested login without TOTP code" -Details @{
        response_status = $response.status_code
        rejected_missing_totp = -not $response.success
    }
    
    # Test invalid TOTP codes
    $invalidCodes = @("000000", "123456", "999999", "abcdef", "12345", "1234567")
    
    foreach ($code in $invalidCodes) {
        $testName = "TOTP Invalid Code - $code"
        
        $body = @{
            email = "test@example.com"
            password = "test123"
            totp_code = $code
        } | ConvertTo-Json
        
        $response = Invoke-SecureRequest -Uri "$BaseUrl/api/auth/login" -Method "POST" -Body $body
        
        Write-TestResult -TestName $testName -Category "TOTP Security" -Severity "Medium" -Passed (-not $response.success) -Description "Tested invalid TOTP code" -Details @{
            invalid_code = $code
            response_status = $response.status_code
            rejected_invalid_totp = -not $response.success
        }
    }
}

# Test 5: Privilege Escalation Tests
function Test-PrivilegeEscalation {
    Write-Host "`n=== Privilege Escalation Tests ===" -ForegroundColor Cyan
    
    # Test access admin endpoints as regular user
    $regularUserToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJyZWd1bGFyX3VzZXIiLCJyb2xlIjoiZW50ZXJwcmlzZV91c2VyIn0.test"
    
    $adminEndpoints = @(
        "/api/admin/organizations",
        "/api/admin/users",
        "/api/admin/compliance/summary",
        "/api/admin/compliance/logs",
        "/api/admin/compliance/export"
    )
    
    foreach ($endpoint in $adminEndpoints) {
        $testName = "Privilege Escalation - $endpoint"
        
        $response = Invoke-SecureRequest -Uri "$BaseUrl$endpoint" -Method "GET" -Headers @{
            "Authorization" = "Bearer $regularUserToken"
        }
        
        Write-TestResult -TestName $testName -Category "Privilege Escalation" -Severity "Critical" -Passed (-not $response.success) -Description "Tested admin endpoint access with regular user token" -Details @{
            endpoint = $endpoint
            user_role = "enterprise_user"
            response_status = $response.status_code
            unauthorized_access = -not $response.success
        }
    }
}

# Test 6: Rate Limiting Tests
function Test-RateLimiting {
    Write-Host "`n=== Rate Limiting Tests ===" -ForegroundColor Cyan
    
    # Test rapid login attempts
    $testName = "Rate Limiting - Login Attempts"
    $rateLimitExceeded = $false
    
    for ($i = 1; $i -le 25; $i++) {
        $body = @{
            email = "test@example.com"
            password = "wrongpassword"
            totp_code = "000000"
        } | ConvertTo-Json
        
        $response = Invoke-SecureRequest -Uri "$BaseUrl/api/auth/login" -Method "POST" -Body $body
        
        if ($response.status_code -eq 429) {
            $rateLimitExceeded = $true
            break
        }
        
        # Also check if we get a rate limit error in the response
        if ($response.content -and $response.content.Contains("Too many requests")) {
            $rateLimitExceeded = $true
            break
        }
        
        Start-Sleep -Milliseconds 10
    }
    
    Write-TestResult -TestName $testName -Category "Rate Limiting" -Severity "Medium" -Passed $rateLimitExceeded -Description "Tested rate limiting on login endpoint" -Details @{
        attempts_made = $i
        rate_limit_exceeded = $rateLimitExceeded
        response_status = $response.status_code
    }
}

# Test 7: Security Headers Tests
function Test-SecurityHeaders {
    Write-Host "`n=== Security Headers Tests ===" -ForegroundColor Cyan
    
    $response = Invoke-SecureRequest -Uri "$BaseUrl/api/health" -Method "GET"
    
    $requiredHeaders = @{
        "X-Content-Type-Options" = "nosniff"
        "X-Frame-Options" = "DENY"
        "X-XSS-Protection" = "1; mode=block"
        "Strict-Transport-Security" = $true
        "Content-Security-Policy" = $true
        "Referrer-Policy" = $true
    }
    
    $missingHeaders = @()
    
    foreach ($header in $requiredHeaders.Keys) {
        if (-not $response.headers.ContainsKey($header)) {
            $missingHeaders += $header
        }
    }
    
    $allHeadersPresent = $missingHeaders.Count -eq 0
    
    Write-TestResult -TestName "Security Headers" -Category "Security Headers" -Severity "Medium" -Passed $allHeadersPresent -Description "Tested presence of security headers" -Details @{
        missing_headers = $missingHeaders
        all_headers_present = $allHeadersPresent
        response_headers = $response.headers
    }
}

# Main execution
Write-Host "Starting Penetration Testing for Secure Email MVP" -ForegroundColor Green
Write-Host "Base URL: $BaseUrl" -ForegroundColor Yellow
Write-Host "Output File: $OutputFile" -ForegroundColor Yellow

# Run all tests
Test-SQLInjection
Test-XSS
Test-JWTTampering
Test-TOTPBypass
Test-PrivilegeEscalation
Test-RateLimiting
Test-SecurityHeaders

# Generate summary
Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "Total Tests: $($TestResults.summary.total_tests)" -ForegroundColor White
Write-Host "Passed: $($TestResults.summary.passed)" -ForegroundColor Green
Write-Host "Failed: $($TestResults.summary.failed)" -ForegroundColor Red
Write-Host "Critical Findings: $($TestResults.summary.critical_findings)" -ForegroundColor Red
Write-Host "High Findings: $($TestResults.summary.high_findings)" -ForegroundColor Yellow
Write-Host "Medium Findings: $($TestResults.summary.medium_findings)" -ForegroundColor Cyan
Write-Host "Low Findings: $($TestResults.summary.low_findings)" -ForegroundColor Gray

# Save results to file
$TestResults | ConvertTo-Json -Depth 10 | Out-File -FilePath $OutputFile -Encoding UTF8

Write-Host "`nTest results saved to: $OutputFile" -ForegroundColor Green

# Exit with appropriate code
if ($TestResults.summary.failed -gt 0) {
    Write-Host "`nPenetration testing completed with security findings!" -ForegroundColor Red
    exit 1
} else {
    Write-Host "`nPenetration testing completed successfully - no security issues found!" -ForegroundColor Green
    exit 0
}
