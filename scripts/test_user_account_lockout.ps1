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

        Write-Output "[SUCCESS] $Method $Endpoint - Status: OK"
        return @{
            Success = $true
            Response = $response
            StatusCode = 200
        }
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Output "[INFO] $Method $Endpoint - Status: $statusCode"

        if ($statusCode -eq $ExpectedStatus) {
            Write-Output "[SUCCESS] Expected status $ExpectedStatus received"
        } else {
            Write-Output "[WARNING] Unexpected status $statusCode (expected $ExpectedStatus)"
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
    Write-Output "`n=== Testing Lockout Configuration ==="

    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/auth/lockout/config"

    if ($result.Success) {
        $config = $result.Response
        Write-Output "Lockout Configuration:"
        Write-Output "  Max Attempts: $($config.max_attempts)"
        Write-Output "  Lockout Duration: $($config.lockout_duration)"
        Write-Output "  Attempt Window: $($config.attempt_window)"
        Write-Output "  Enabled: $($config.enabled)"
    } else {
        Write-Output "[ERROR] Failed to get lockout configuration"
    }
}

# Test user signup
function Test-UserSignup {
    Write-Output "`n=== Testing User Signup ==="

    $signupData = @{
        email = $TestEmail
        password = $TestPassword
        fallback_email = $FallbackEmail
    }

    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData -ExpectedStatus 201

    if ($result.Success) {
        Write-Output "[SUCCESS] User signup completed"
    } else {
        if ($result.StatusCode -eq 409) {
            Write-Output "[INFO] User already exists, continuing with tests"
        } else {
            Write-Output "[ERROR] User signup failed: $($result.Error)"
            return $false
        }
    }

    return $true
}

# Test successful login
function Test-SuccessfulLogin {
    Write-Output "`n=== Testing Successful Login ==="

    $loginData = @{
        email = $TestEmail
        password = $TestPassword
        totp_code = "123456"  # Default TOTP code for testing
    }

    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -ExpectedStatus 200

    if ($result.Success) {
        Write-Output "[SUCCESS] Login successful"
        return $result.Response.access_token
    } else {
        Write-Output "[ERROR] Login failed: $($result.Error)"
        return $null
    }
}

# Test failed login attempts
function Test-FailedLoginAttempts {
    Write-Output "`n=== Testing Failed Login Attempts ==="

    $wrongPassword = "WrongPassword123!"
    $loginData = @{
        email = $TestEmail
        password = $wrongPassword
        totp_code = "123456"
    }

    # Attempt 5 failed logins (should trigger lockout)
    for ($i = 1; $i -le 5; $i++) {
        Write-Output "Attempt $i of 5 failed logins..."

        $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -ExpectedStatus 401

        if ($result.StatusCode -eq 401) {
            Write-Output "[SUCCESS] Failed login attempt $i recorded"
        } elseif ($result.StatusCode -eq 429) {
            Write-Output "[SUCCESS] Account locked after attempt $i"
            return $true
        } else {
            Write-Output "[ERROR] Unexpected response on attempt $i"
        }

        # Small delay between attempts
        Start-Sleep -Milliseconds 100
    }

    return $false
}

# Test account lockout
function Test-AccountLockout {
    Write-Output "`n=== Testing Account Lockout ==="

    $loginData = @{
        email = $TestEmail
        password = $TestPassword  # Correct password
        totp_code = "123456"
    }

    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -ExpectedStatus 429

    if ($result.StatusCode -eq 429) {
        Write-Output "[SUCCESS] Account is properly locked"
        return $true
    } else {
        Write-Output "[ERROR] Account should be locked but isn't"
        return $false
    }
}

# Test lockout status endpoint
function Test-LockoutStatus {
    Write-Output "`n=== Testing Lockout Status Endpoint ==="

    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/auth/lockout/status?email=$TestEmail"

    if ($result.Success) {
        $status = $result.Response
        Write-Output "Lockout Status:"
        Write-Output "  Email: $($status.email)"
        Write-Output "  Is Locked Out: $($status.is_locked_out)"
        Write-Output "  Failed Attempts: $($status.failed_attempts)"
        Write-Output "  Max Attempts: $($status.max_attempts)"
        Write-Output "  Remaining Attempts: $($status.remaining_attempts)"

        if ($status.lockout_remaining) {
            Write-Output "  Lockout Remaining: $($status.lockout_remaining)"
        }

        return $status.is_locked_out
    } else {
        Write-Output "[ERROR] Failed to get lockout status"
        return $false
    }
}

# Test account unlock
function Test-AccountUnlock {
    Write-Output "`n=== Testing Account Unlock ==="

    $unlockData = @{
        email = $TestEmail
    }

    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/lockout/unlock" -Body $unlockData -ExpectedStatus 200

    if ($result.Success) {
        $response = $result.Response
        Write-Output "[SUCCESS] Account unlock response:"
        Write-Output "  Email: $($response.email)"
        Write-Output "  Unlocked: $($response.unlocked)"
        Write-Output "  Message: $($response.message)"

        return $response.unlocked
    } else {
        Write-Output "[ERROR] Failed to unlock account: $($result.Error)"
        return $false
    }
}

# Test login after unlock
function Test-LoginAfterUnlock {
    Write-Output "`n=== Testing Login After Unlock ==="

    $loginData = @{
        email = $TestEmail
        password = $TestPassword
        totp_code = "123456"
    }

    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -ExpectedStatus 200

    if ($result.Success) {
        Write-Output "[SUCCESS] Login successful after unlock"
        return $true
    } else {
        Write-Output "[ERROR] Login failed after unlock: $($result.Error)"
        return $false
    }
}

# Test lockout statistics
function Test-LockoutStatistics {
    Write-Output "`n=== Testing Lockout Statistics ==="

    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/auth/lockout/stats"

    if ($result.Success) {
        $stats = $result.Response
        Write-Output "Lockout Statistics:"
        Write-Output "  Currently Locked Accounts: $($stats.currently_locked_accounts)"
        Write-Output "  Accounts with Failed Attempts: $($stats.accounts_with_failed_attempts)"
        Write-Output "  Recent Lockout Events (24h): $($stats.recent_lockout_events_24h)"
        Write-Output "  Timestamp: $($stats.timestamp)"
    } else {
        Write-Output "[ERROR] Failed to get lockout statistics"
    }
}

# Test attempt window reset
function Test-AttemptWindowReset {
    Write-Output "`n=== Testing Attempt Window Reset ==="

    Write-Output "This test would require waiting for the attempt window to expire."
    Write-Output "In a real scenario, this would test that failed attempts reset after the window."
    Write-Output "For this test, we'll simulate by checking the status."

    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/auth/lockout/status?email=$TestEmail"

    if ($result.Success) {
        $status = $result.Response
        Write-Output "Current Status After Unlock:"
        Write-Output "  Failed Attempts: $($status.failed_attempts)"
        Write-Output "  Is Within Window: $($status.is_within_window)"
    }
}

# Main test execution
function Main {
    Write-Output "Starting User Account Lockout Tests (Micro-Iteration 4.6)"
    Write-Output "API Base: $ApiBase"
    Write-Output "Test Email: $TestEmail"

    # Test 1: Configuration
    Test-LockoutConfiguration

    # Test 2: User signup
    if (-not (Test-UserSignup)) {
        Write-Output "[ERROR] Cannot continue without user signup"
        return
    }

    # Test 3: Initial successful login
    $token = Test-SuccessfulLogin
    if (-not $token) {
        Write-Output "[WARNING] Initial login failed, but continuing with tests"
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

    Write-Output "`n=== Test Summary ==="
    Write-Output "User Account Lockout functionality has been tested."
    Write-Output "Check the logs above for detailed results."
}

# Run the tests
Main






















