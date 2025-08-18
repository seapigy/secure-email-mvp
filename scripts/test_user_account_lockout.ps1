# Test script for User Account Lockout functionality (Micro-Iteration 4.6)
# Tests the temporary account lockout after failed login attempts

param(
    [string]$ApiBase = "http://localhost:8080",
    [string]$TestEmail = "lockout-test@example.com",
    [string]$TestPassword = "StrongPassword123!",
    [string]$FallbackEmail = "fallback-lockout@example.com"
)

# Helper function to make API requests
function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [int]$ExpectedStatus = 200
    )
    
    $uri = "$ApiBase$Endpoint"
    $headers = @{
        "Content-Type" = "application/json"
    }
    
    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -Body $jsonBody
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }
        
        Write-Host "[SUCCESS] $Method $Endpoint - Status: OK" -ForegroundColor Green
        return @{
            Success = $true
            Response = $response
            StatusCode = 200
        }
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "[INFO] $Method $Endpoint - Status: $statusCode" -ForegroundColor Yellow
        
        if ($statusCode -eq $ExpectedStatus) {
            Write-Host "[SUCCESS] Expected status $ExpectedStatus received" -ForegroundColor Green
        } else {
            Write-Host "[WARNING] Unexpected status $statusCode (expected $ExpectedStatus)" -ForegroundColor Yellow
        }
        
        return @{
            Success = $false
            StatusCode = $statusCode
            Error = $_.Exception.Message
        }
    }
}

# Test configuration
function Test-LockoutConfiguration {
    Write-Host "`n=== Testing Lockout Configuration ===" -ForegroundColor Cyan
    
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/auth/lockout/config"
    
    if ($result.Success) {
        $config = $result.Response
        Write-Host "Lockout Configuration:" -ForegroundColor Green
        Write-Host "  Max Attempts: $($config.max_attempts)" -ForegroundColor White
        Write-Host "  Lockout Duration: $($config.lockout_duration)" -ForegroundColor White
        Write-Host "  Attempt Window: $($config.attempt_window)" -ForegroundColor White
        Write-Host "  Enabled: $($config.enabled)" -ForegroundColor White
    } else {
        Write-Host "[ERROR] Failed to get lockout configuration" -ForegroundColor Red
    }
}

# Test user signup
function Test-UserSignup {
    Write-Host "`n=== Testing User Signup ===" -ForegroundColor Cyan
    
    $signupData = @{
        email = $TestEmail
        password = $TestPassword
        fallback_email = $FallbackEmail
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData -ExpectedStatus 201
    
    if ($result.Success) {
        Write-Host "[SUCCESS] User signup completed" -ForegroundColor Green
    } else {
        if ($result.StatusCode -eq 409) {
            Write-Host "[INFO] User already exists, continuing with tests" -ForegroundColor Yellow
        } else {
            Write-Host "[ERROR] User signup failed: $($result.Error)" -ForegroundColor Red
            return $false
        }
    }
    
    return $true
}

# Test successful login
function Test-SuccessfulLogin {
    Write-Host "`n=== Testing Successful Login ===" -ForegroundColor Cyan
    
    $loginData = @{
        email = $TestEmail
        password = $TestPassword
        totp_code = "123456"  # Default TOTP code for testing
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -ExpectedStatus 200
    
    if ($result.Success) {
        Write-Host "[SUCCESS] Login successful" -ForegroundColor Green
        return $result.Response.access_token
    } else {
        Write-Host "[ERROR] Login failed: $($result.Error)" -ForegroundColor Red
        return $null
    }
}

# Test failed login attempts
function Test-FailedLoginAttempts {
    Write-Host "`n=== Testing Failed Login Attempts ===" -ForegroundColor Cyan
    
    $wrongPassword = "WrongPassword123!"
    $loginData = @{
        email = $TestEmail
        password = $wrongPassword
        totp_code = "123456"
    }
    
    # Attempt 5 failed logins (should trigger lockout)
    for ($i = 1; $i -le 5; $i++) {
        Write-Host "Attempt $i of 5 failed logins..." -ForegroundColor Yellow
        
        $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -ExpectedStatus 401
        
        if ($result.StatusCode -eq 401) {
            Write-Host "[SUCCESS] Failed login attempt $i recorded" -ForegroundColor Green
        } elseif ($result.StatusCode -eq 429) {
            Write-Host "[SUCCESS] Account locked after attempt $i" -ForegroundColor Green
            return $true
        } else {
            Write-Host "[ERROR] Unexpected response on attempt $i" -ForegroundColor Red
        }
        
        # Small delay between attempts
        Start-Sleep -Milliseconds 100
    }
    
    return $false
}

# Test account lockout
function Test-AccountLockout {
    Write-Host "`n=== Testing Account Lockout ===" -ForegroundColor Cyan
    
    $loginData = @{
        email = $TestEmail
        password = $TestPassword  # Correct password
        totp_code = "123456"
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -ExpectedStatus 429
    
    if ($result.StatusCode -eq 429) {
        Write-Host "[SUCCESS] Account is properly locked" -ForegroundColor Green
        return $true
    } else {
        Write-Host "[ERROR] Account should be locked but isn't" -ForegroundColor Red
        return $false
    }
}

# Test lockout status endpoint
function Test-LockoutStatus {
    Write-Host "`n=== Testing Lockout Status Endpoint ===" -ForegroundColor Cyan
    
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/auth/lockout/status?email=$TestEmail"
    
    if ($result.Success) {
        $status = $result.Response
        Write-Host "Lockout Status:" -ForegroundColor Green
        Write-Host "  Email: $($status.email)" -ForegroundColor White
        Write-Host "  Is Locked Out: $($status.is_locked_out)" -ForegroundColor White
        Write-Host "  Failed Attempts: $($status.failed_attempts)" -ForegroundColor White
        Write-Host "  Max Attempts: $($status.max_attempts)" -ForegroundColor White
        Write-Host "  Remaining Attempts: $($status.remaining_attempts)" -ForegroundColor White
        
        if ($status.lockout_remaining) {
            Write-Host "  Lockout Remaining: $($status.lockout_remaining)" -ForegroundColor White
        }
        
        return $status.is_locked_out
    } else {
        Write-Host "[ERROR] Failed to get lockout status" -ForegroundColor Red
        return $false
    }
}

# Test account unlock
function Test-AccountUnlock {
    Write-Host "`n=== Testing Account Unlock ===" -ForegroundColor Cyan
    
    $unlockData = @{
        email = $TestEmail
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/lockout/unlock" -Body $unlockData -ExpectedStatus 200
    
    if ($result.Success) {
        $response = $result.Response
        Write-Host "[SUCCESS] Account unlock response:" -ForegroundColor Green
        Write-Host "  Email: $($response.email)" -ForegroundColor White
        Write-Host "  Unlocked: $($response.unlocked)" -ForegroundColor White
        Write-Host "  Message: $($response.message)" -ForegroundColor White
        
        return $response.unlocked
    } else {
        Write-Host "[ERROR] Failed to unlock account: $($result.Error)" -ForegroundColor Red
        return $false
    }
}

# Test login after unlock
function Test-LoginAfterUnlock {
    Write-Host "`n=== Testing Login After Unlock ===" -ForegroundColor Cyan
    
    $loginData = @{
        email = $TestEmail
        password = $TestPassword
        totp_code = "123456"
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -ExpectedStatus 200
    
    if ($result.Success) {
        Write-Host "[SUCCESS] Login successful after unlock" -ForegroundColor Green
        return $true
    } else {
        Write-Host "[ERROR] Login failed after unlock: $($result.Error)" -ForegroundColor Red
        return $false
    }
}

# Test lockout statistics
function Test-LockoutStatistics {
    Write-Host "`n=== Testing Lockout Statistics ===" -ForegroundColor Cyan
    
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/auth/lockout/stats"
    
    if ($result.Success) {
        $stats = $result.Response
        Write-Host "Lockout Statistics:" -ForegroundColor Green
        Write-Host "  Currently Locked Accounts: $($stats.currently_locked_accounts)" -ForegroundColor White
        Write-Host "  Accounts with Failed Attempts: $($stats.accounts_with_failed_attempts)" -ForegroundColor White
        Write-Host "  Recent Lockout Events (24h): $($stats.recent_lockout_events_24h)" -ForegroundColor White
        Write-Host "  Timestamp: $($stats.timestamp)" -ForegroundColor White
    } else {
        Write-Host "[ERROR] Failed to get lockout statistics" -ForegroundColor Red
    }
}

# Test attempt window reset
function Test-AttemptWindowReset {
    Write-Host "`n=== Testing Attempt Window Reset ===" -ForegroundColor Cyan
    
    Write-Host "This test would require waiting for the attempt window to expire." -ForegroundColor Yellow
    Write-Host "In a real scenario, this would test that failed attempts reset after the window." -ForegroundColor Yellow
    Write-Host "For this test, we'll simulate by checking the status." -ForegroundColor Yellow
    
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/auth/lockout/status?email=$TestEmail"
    
    if ($result.Success) {
        $status = $result.Response
        Write-Host "Current Status After Unlock:" -ForegroundColor Green
        Write-Host "  Failed Attempts: $($status.failed_attempts)" -ForegroundColor White
        Write-Host "  Is Within Window: $($status.is_within_window)" -ForegroundColor White
    }
}

# Main test execution
function Main {
    Write-Host "Starting User Account Lockout Tests (Micro-Iteration 4.6)" -ForegroundColor Magenta
    Write-Host "API Base: $ApiBase" -ForegroundColor White
    Write-Host "Test Email: $TestEmail" -ForegroundColor White
    
    # Test 1: Configuration
    Test-LockoutConfiguration
    
    # Test 2: User signup
    if (-not (Test-UserSignup)) {
        Write-Host "[ERROR] Cannot continue without user signup" -ForegroundColor Red
        return
    }
    
    # Test 3: Initial successful login
    $token = Test-SuccessfulLogin
    if (-not $token) {
        Write-Host "[WARNING] Initial login failed, but continuing with tests" -ForegroundColor Yellow
    }
    
    # Test 4: Failed login attempts
    $locked = Test-FailedLoginAttempts
    
    # Test 5: Account lockout verification
    if ($locked) {
        Test-AccountLockout
    }
    
    # Test 6: Lockout status
    $isLocked = Test-LockoutStatus
    
    # Test 7: Account unlock
    if ($isLocked) {
        Test-AccountUnlock
    }
    
    # Test 8: Login after unlock
    Test-LoginAfterUnlock
    
    # Test 9: Lockout statistics
    Test-LockoutStatistics
    
    # Test 10: Attempt window information
    Test-AttemptWindowReset
    
    Write-Host "`n=== Test Summary ===" -ForegroundColor Cyan
    Write-Host "User Account Lockout functionality has been tested." -ForegroundColor Green
    Write-Host "Check the logs above for detailed results." -ForegroundColor White
}

# Run the tests
Main







