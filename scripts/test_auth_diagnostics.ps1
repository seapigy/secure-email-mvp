# =============================================================================
# SECURE EMAIL MVP - AUTHENTICATION DIAGNOSTIC TESTS
# =============================================================================
# PowerShell script to test authentication components and identify issues

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",
    [string]$TestPassword = "TestPassword123!",
    [string]$TestTOTPSecret = "JBSWY3DPEHPK3PXP"
)

# Test configuration
$TestConfig = @{
    BaseUrl = $BaseUrl
    TestEmail = $TestEmail
    TestPassword = $TestPassword
    TestTOTPSecret = $TestTOTPSecret
}

# Utility functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Cyan"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Headers = @{},
        [string]$Body = ""
    )

    $uri = "$($TestConfig.BaseUrl)$Endpoint"

    $requestHeaders = @{
        "Content-Type" = "application/json"
    }

    foreach ($key in $Headers.Keys) {
        $requestHeaders[$key] = $Headers[$key]
    }

    try {
        if ($Method -eq "GET" -or $Body -eq "") {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $requestHeaders -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $requestHeaders -Body $Body -ErrorAction Stop
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
            $errorBody = "Unable to read error response"
        }

        return @{
            Success = $false
            StatusCode = $statusCode
            Error = $errorMessage
            ErrorBody = $errorBody
        }
    }
}

# Test functions
function Test-HashConsistency {
    Write-Output ""
    Write-Output "=" * 80 -ForegroundColor Cyan
    Write-Output " Testing Hash Consistency"
    Write-Output "=" * 80 -ForegroundColor Cyan

    Write-Info "Testing Argon2 hash consistency with known parameters"

    # Test 1: Create a new user and immediately try to authenticate
    Write-Info "Step 1: Creating test user for hash consistency test"

    $signupData = @{
        email = "hash_test@securesystem.email"
        password = $TestConfig.TestPassword
        fallback_email = "fallback@securesystem.email"
    }

    $signupResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body ($signupData | ConvertTo-Json)

    if ($signupResponse.Success) {
        Write-Success "Test user created successfully"

        # Test 2: Immediately try to authenticate with the same credentials
        Write-Info "Step 2: Attempting immediate authentication"

        Start-Sleep -Seconds 2  # Small delay to ensure user is fully created

        $loginData = @{
            email = "hash_test@securesystem.email"
            password = $TestConfig.TestPassword
            totp_code = "123456"  # Use test TOTP code
        }

        $loginResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body ($loginData | ConvertTo-Json)

        if ($loginResponse.Success) {
            Write-Success "Hash consistency test PASSED - Authentication successful"
            Write-Info "Token: $($loginResponse.Data.access_token.Substring(0, [Math]::Min(20, $loginResponse.Data.access_token.Length)))..."
        } else {
            Write-Error "Hash consistency test FAILED - Authentication failed"
            Write-Info "Status: $($loginResponse.StatusCode)"
            Write-Info "Error: $($loginResponse.Error)"
            if ($loginResponse.ErrorBody) {
                Write-Info "Error Body: $($loginResponse.ErrorBody)"
            }
        }
    } else {
        Write-Error "Failed to create test user for hash consistency test"
        Write-Info "Status: $($signupResponse.StatusCode)"
        Write-Info "Error: $($signupResponse.Error)"
    }
}

function Test-TOTPValidation {
    Write-Output ""
    Write-Output "=" * 80 -ForegroundColor Cyan
    Write-Output " Testing TOTP Validation"
    Write-Output "=" * 80 -ForegroundColor Cyan

    Write-Info "Testing TOTP code generation and validation"

    # Test 1: Generate TOTP code using the generator
    Write-Info "Step 1: Generating TOTP code from secret"

    try {
        $totpCode = & .\totp_generator.exe $TestConfig.TestTOTPSecret
        $totpCode = $totpCode.Trim()
        Write-Success "TOTP code generated: $totpCode"

        # Test 2: Try to authenticate with the generated TOTP code
        Write-Info "Step 2: Attempting authentication with generated TOTP code"

        $loginData = @{
            email = $TestConfig.TestEmail
            password = $TestConfig.TestPassword
            totp_code = $totpCode
        }

        $loginResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body ($loginData | ConvertTo-Json)

        if ($loginResponse.Success) {
            Write-Success "TOTP validation test PASSED - Authentication successful"
            Write-Info "Token: $($loginResponse.Data.access_token.Substring(0, [Math]::Min(20, $loginResponse.Data.access_token.Length)))..."
        } else {
            Write-Error "TOTP validation test FAILED - Authentication failed"
            Write-Info "Status: $($loginResponse.StatusCode)"
            Write-Info "Error: $($loginResponse.Error)"
            if ($loginResponse.ErrorBody) {
                Write-Info "Error Body: $($loginResponse.ErrorBody)"
            }
        }

    } catch {
        Write-Error "Failed to generate TOTP code: $($_.Exception.Message)"
    }
}

function Test-ExistingUserAuthentication {
    Write-Output ""
    Write-Output "=" * 80 -ForegroundColor Cyan
    Write-Output " Testing Existing User Authentication"
    Write-Output "=" * 80 -ForegroundColor Cyan

    Write-Info "Testing authentication with existing test user"

    # Test with the existing test user
    $loginData = @{
        email = $TestConfig.TestEmail
        password = $TestConfig.TestPassword
        totp_code = "123456"  # Use test TOTP code
    }

    $loginResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body ($loginData | ConvertTo-Json)

    if ($loginResponse.Success) {
        Write-Success "Existing user authentication PASSED"
        Write-Info "Token: $($loginResponse.Data.access_token.Substring(0, [Math]::Min(20, $loginResponse.Data.access_token.Length)))..."
    } else {
        Write-Error "Existing user authentication FAILED"
        Write-Info "Status: $($loginResponse.StatusCode)"
        Write-Info "Error: $($loginResponse.Error)"
        if ($loginResponse.ErrorBody) {
            Write-Info "Error Body: $($loginResponse.ErrorBody)"
        }
    }
}

function Test-APIHealth {
    Write-Output ""
    Write-Output "=" * 80 -ForegroundColor Cyan
    Write-Output " Testing API Health"
    Write-Output "=" * 80 -ForegroundColor Cyan

    Write-Info "Checking API server health"

    $healthResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/health"

    if ($healthResponse.Success) {
        Write-Success "API server is healthy"
        Write-Info "Response: $($healthResponse.Data | ConvertTo-Json)"
    } else {
        Write-Error "API server health check failed"
        Write-Info "Status: $($healthResponse.StatusCode)"
        Write-Info "Error: $($healthResponse.Error)"
    }
}

# Main test execution
function Start-AuthDiagnostics {
    Write-Output ""
    Write-Output "Starting Authentication Diagnostics"
    Write-Output "Base URL: $($TestConfig.BaseUrl)"
    Write-Output "Test Email: $($TestConfig.TestEmail)"
    Write-Output "Test TOTP Secret: $($TestConfig.TestTOTPSecret)"

    # Run diagnostic tests
    Test-APIHealth
    Test-ExistingUserAuthentication
    Test-TOTPValidation
    Test-HashConsistency

    Write-Output ""
    Write-Output "=" * 80 -ForegroundColor Green
    Write-Output " Authentication Diagnostics Complete"
    Write-Output "=" * 80 -ForegroundColor Green

    Write-Output ""
    Write-Output "Next Steps:"
    Write-Output "  1. Check server logs for [AUTH_DEBUG] messages"
    Write-Output "  2. Compare hash generation between signup and login"
    Write-Output "  3. Verify TOTP secret encoding and validation"
    Write-Output "  4. Check for email normalization differences"
}

# Script execution
Start-AuthDiagnostics
