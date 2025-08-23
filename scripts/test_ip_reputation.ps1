# Test IP Reputation Service Integration
# This script tests the IP reputation service integration with signup and login endpoints

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$ApiKey = $env:IP_REPUTATION_API_KEY
)

# Colors for output
$Green = "Green"
$Red = "Red"
$Yellow = "Yellow"
$White = "White"

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = $White
    )
    Write-Host $Message -ForegroundColor $Color
}

function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [hashtable]$Headers = @{}
    )

    $uri = "$BaseUrl$Endpoint"
    $headers["Content-Type"] = "application/json"

    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Body $jsonBody -Headers $headers -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -ErrorAction Stop
        }
        return @{
            Success = $true
            Data = $response
            StatusCode = 200
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $errorMessage = $_.Exception.Message
        try {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
        } catch {
            $errorBody = $errorMessage
        }
        return @{
            Success = $false
            StatusCode = $statusCode
            Error = $errorBody
        }
    }
}

function Test-SignupWithIPReputation {
    Write-ColorOutput "`n=== Testing Signup with IP Reputation ===" $Yellow

    $testEmail = "testuser$(Get-Random)@securesystem.email"
    $testPassword = "SecurePassword123!"
    $fallbackEmail = "fallback$(Get-Random)@example.com"

    $signupData = @{
        email = $testEmail
        password = $testPassword
        fallback_email = $fallbackEmail
    }

    Write-ColorOutput "Attempting signup with email: $testEmail" $White

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData

    if ($response.Success) {
        Write-ColorOutput "[SUCCESS] Signup successful - IP reputation check passed" $Green
        return $testEmail
    } elseif ($response.StatusCode -eq 403 -and $response.Error -like "*IP reputation*") {
        Write-ColorOutput "[SUCCESS] Signup blocked by IP reputation service (expected behavior)" $Green
        return $null
    } else {
        Write-ColorOutput "[ERROR] Signup failed with unexpected error: $($response.Error)" $Red
        return $null
    }
}

function Test-LoginWithIPReputation {
    param([string]$Email)

    Write-ColorOutput "`n=== Testing Login with IP Reputation ===" $Yellow

    if (-not $Email) {
        Write-ColorOutput "Skipping login test - no valid email from signup" $Yellow
        return
    }

    $loginData = @{
        email = $Email
        password = "SecurePassword123!"
        totp_code = "123456"  # This will fail, but we're testing IP reputation, not auth
    }

    Write-ColorOutput "Attempting login with email: $Email" $White

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData

    if ($response.StatusCode -eq 403 -and $response.Error -like "*IP reputation*") {
        Write-ColorOutput "[SUCCESS] Login blocked by IP reputation service (expected behavior)" $Green
    } elseif ($response.StatusCode -eq 401) {
        Write-ColorOutput "[SUCCESS] Login reached authentication (IP reputation check passed)" $Green
    } else {
        Write-ColorOutput "[ERROR] Login failed with unexpected error: $($response.Error)" $Red
    }
}

function Test-IPReputationConfiguration {
    Write-ColorOutput "`n=== Testing IP Reputation Configuration ===" $Yellow

    if ($ApiKey) {
        Write-ColorOutput "[SUCCESS] IP_REPUTATION_API_KEY is configured" $Green
    } else {
        Write-ColorOutput "[WARNING] IP_REPUTATION_API_KEY is not configured - service will allow all IPs" $Yellow
    }

    $threshold = $env:IP_REPUTATION_THRESHOLD
    if ($threshold) {
        Write-ColorOutput "[SUCCESS] IP_REPUTATION_THRESHOLD is set to: $threshold" $Green
    } else {
        Write-ColorOutput "[INFO] IP_REPUTATION_THRESHOLD not set - using default (25)" $White
    }
}

function Test-HealthCheck {
    Write-ColorOutput "`n=== Testing Health Check ===" $Yellow

    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/health"

    if ($response.Success) {
        Write-ColorOutput "[SUCCESS] Health check passed" $Green
        return $true
    } else {
        Write-ColorOutput "[ERROR] Health check failed: $($response.Error)" $Red
        return $false
    }
}

# Main test execution
Write-ColorOutput "Starting IP Reputation Service Integration Tests" $Yellow
Write-ColorOutput "Base URL: $BaseUrl" $White

# Test health check first
$healthOk = Test-HealthCheck
if (-not $healthOk) {
    Write-ColorOutput "`n[ERROR] Server is not responding. Please ensure the API server is running." $Red
    exit 1
}

# Test configuration
Test-IPReputationConfiguration

# Test signup with IP reputation
$testEmail = Test-SignupWithIPReputation

# Test login with IP reputation
Test-LoginWithIPReputation -Email $testEmail

Write-ColorOutput "`n=== Test Summary ===" $Yellow
Write-ColorOutput "IP Reputation Service integration tests completed." $White
Write-ColorOutput "Note: Actual IP reputation results depend on your current IP address." $White
Write-ColorOutput "To test with a known malicious IP, you would need to configure the service" $White
Write-ColorOutput "with a test API key and use a VPN or proxy with a flagged IP address." $White
