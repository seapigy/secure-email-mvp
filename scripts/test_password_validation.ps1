# Test Password Validation & Breach Check Integration
# This script tests the password validation service integration with signup and password reset endpoints

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$ApiKey = $env:HIBP_API_KEY
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

function Test-SignupWithWeakPassword {
    Write-ColorOutput "`n=== Testing Signup with Weak Password ===" $Yellow

    $testEmail = "testuser$(Get-Random)@securesystem.email"
    $weakPassword = "weak"
    $fallbackEmail = "fallback$(Get-Random)@example.com"

    $signupData = @{
        email = $testEmail
        password = $weakPassword
        fallback_email = $fallbackEmail
    }

    Write-ColorOutput "Attempting signup with weak password: '$weakPassword'" $White

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData

    if ($response.StatusCode -eq 400 -and $response.Error -like "*security requirements*") {
        Write-ColorOutput "[SUCCESS] Signup correctly blocked due to weak password" $Green
        return $true
    } else {
        Write-ColorOutput "[ERROR] Signup should have been blocked for weak password" $Red
        return $false
    }
}

function Test-SignupWithCommonPassword {
    Write-ColorOutput "`n=== Testing Signup with Common Password ===" $Yellow

    $testEmail = "testuser$(Get-Random)@securesystem.email"
    $commonPassword = "password"
    $fallbackEmail = "fallback$(Get-Random)@example.com"

    $signupData = @{
        email = $testEmail
        password = $commonPassword
        fallback_email = $fallbackEmail
    }

    Write-ColorOutput "Attempting signup with common password: '$commonPassword'" $White

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData

    if ($response.StatusCode -eq 400 -and $response.Error -like "*security requirements*") {
        Write-ColorOutput "[SUCCESS] Signup correctly blocked due to common password" $Green
        return $true
    } else {
        Write-ColorOutput "[ERROR] Signup should have been blocked for common password" $Red
        return $false
    }
}

function Test-SignupWithStrongPassword {
    Write-ColorOutput "`n=== Testing Signup with Strong Password ===" $Yellow

    $testEmail = "testuser$(Get-Random)@securesystem.email"
    $strongPassword = "SecurePassword123!"
    $fallbackEmail = "fallback$(Get-Random)@example.com"

    $signupData = @{
        email = $testEmail
        password = $strongPassword
        fallback_email = $fallbackEmail
    }

    Write-ColorOutput "Attempting signup with strong password: '$strongPassword'" $White

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData

    if ($response.Success) {
        Write-ColorOutput "[SUCCESS] Signup successful with strong password" $Green
        return $testEmail
    } else {
        Write-ColorOutput "[ERROR] Signup failed with strong password: $($response.Error)" $Red
        return $null
    }
}

function Test-SignupWithMissingRequirements {
    Write-ColorOutput "`n=== Testing Signup with Missing Requirements ===" $Yellow

    $testCases = @(
        @{ Password = "nouppercase123!"; Description = "No uppercase letters" },
        @{ Password = "NOLOWERCASE123!"; Description = "No lowercase letters" },
        @{ Password = "NoNumbers!"; Description = "No numbers" },
        @{ Password = "NoSpecialChars123"; Description = "No special characters" },
        @{ Password = "Short1!"; Description = "Too short" }
    )

    $allPassed = $true

    foreach ($testCase in $testCases) {
        $testEmail = "testuser$(Get-Random)@securesystem.email"
        $fallbackEmail = "fallback$(Get-Random)@example.com"

        $signupData = @{
            email = $testEmail
            password = $testCase.Password
            fallback_email = $fallbackEmail
        }

        Write-ColorOutput "Testing: $($testCase.Description) - '$($testCase.Password)'" $White

        $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData

        if ($response.StatusCode -eq 400 -and $response.Error -like "*security requirements*") {
            Write-ColorOutput "[SUCCESS] Correctly blocked: $($testCase.Description)" $Green
        } else {
            Write-ColorOutput "[ERROR] Should have been blocked: $($testCase.Description)" $Red
            $allPassed = $false
        }
    }

    return $allPassed
}

function Test-PasswordResetValidation {
    Write-ColorOutput "`n=== Testing Password Reset Validation ===" $Yellow

    # First create a user with strong password
    $testEmail = "testuser$(Get-Random)@securesystem.email"
    $strongPassword = "SecurePassword123!"
    $fallbackEmail = "fallback$(Get-Random)@example.com"

    $signupData = @{
        email = $testEmail
        password = $strongPassword
        fallback_email = $fallbackEmail
    }

    $signupResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData
    if (-not $signupResponse.Success) {
        Write-ColorOutput "[ERROR] Failed to create test user for password reset test" $Red
        return $false
    }

    # Test password reset with weak password
    $weakPassword = "weak"
    $resetData = @{
        email = $testEmail
        new_password = $weakPassword
        reset_token = "fake-token"
    }

    Write-ColorOutput "Testing password reset with weak password: '$weakPassword'" $White

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/reset-password" -Body $resetData

    if ($response.StatusCode -eq 400 -and $response.Error -like "*security requirements*") {
        Write-ColorOutput "[SUCCESS] Password reset correctly blocked due to weak password" $Green
        return $true
    } else {
        Write-ColorOutput "[ERROR] Password reset should have been blocked for weak password" $Red
        return $false
    }
}

function Test-PasswordConfiguration {
    Write-ColorOutput "`n=== Testing Password Configuration ===" $Yellow

    if ($ApiKey) {
        Write-ColorOutput "[SUCCESS] HIBP_API_KEY is configured" $Green
    } else {
        Write-ColorOutput "[WARNING] HIBP_API_KEY is not configured - breach checking will be limited" $Yellow
    }

    Write-ColorOutput "[INFO] Password requirements:" $White
    Write-ColorOutput "  - Minimum length: 12 characters" $White
    Write-ColorOutput "  - Must contain uppercase letters" $White
    Write-ColorOutput "  - Must contain lowercase letters" $White
    Write-ColorOutput "  - Must contain numbers" $White
    Write-ColorOutput "  - Must contain special characters" $White
    Write-ColorOutput "  - Must not be in common password list" $White
    Write-ColorOutput "  - Must not be compromised (if API key configured)" $White
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
Write-ColorOutput "Starting Password Validation & Breach Check Integration Tests" $Yellow
Write-ColorOutput "Base URL: $BaseUrl" $White

# Test health check first
$healthOk = Test-HealthCheck
if (-not $healthOk) {
    Write-ColorOutput "`n[ERROR] Server is not responding. Please ensure the API server is running." $Red
    exit 1
}

# Test configuration
Test-PasswordConfiguration

# Test weak password rejection
$weakPasswordTest = Test-SignupWithWeakPassword

# Test common password rejection
$commonPasswordTest = Test-SignupWithCommonPassword

# Test missing requirements
$requirementsTest = Test-SignupWithMissingRequirements

# Test strong password acceptance
$strongPasswordTest = Test-SignupWithStrongPassword

# Test password reset validation
$resetTest = Test-PasswordResetValidation

Write-ColorOutput "`n=== Test Summary ===" $Yellow
Write-ColorOutput "Password validation integration tests completed." $White
Write-ColorOutput "Results:" $White
Write-ColorOutput "  - Weak password rejection: $(if ($weakPasswordTest) { '[SUCCESS]' } else { '[FAILED]' })" $White
Write-ColorOutput "  - Common password rejection: $(if ($commonPasswordTest) { '[SUCCESS]' } else { '[FAILED]' })" $White
Write-ColorOutput "  - Requirements validation: $(if ($requirementsTest) { '[SUCCESS]' } else { '[FAILED]' })" $White
Write-ColorOutput "  - Strong password acceptance: $(if ($strongPasswordTest) { '[SUCCESS]' } else { '[FAILED]' })" $White
Write-ColorOutput "  - Password reset validation: $(if ($resetTest) { '[SUCCESS]' } else { '[FAILED]' })" $White

Write-ColorOutput "`nNote: Breach checking requires a valid HIBP API key to be fully tested." $White
Write-ColorOutput "Get your free API key at: https://haveibeenpwned.com/API/Key" $White











