# =============================================================================
# SECURE EMAIL MVP - TEMPORARY TOKEN GENERATOR (DEVELOPMENT ONLY)
# =============================================================================
# WARNING: This script is for development/testing only. Remove before production.
# Generates valid JWT tokens for testing Micro-Iteration 4.31 endpoints

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$UserEmail = "test@securesystem.email",
    [string]$UserID = "temp-user-id-123",
    [string]$UserRole = "user",
    [int]$ExpirationHours = 1
)

# Test configuration
$TestConfig = @{
    BaseUrl = $BaseUrl
    UserEmail = $UserEmail
    UserID = $UserID
    UserRole = $UserRole
    ExpirationHours = $ExpirationHours
}

# Utility functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Output $Message
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

function Generate-TempJWTToken {
    param(
        [string]$UserID,
        [string]$Email,
        [string]$Role = "user"
    )

    Write-Info "Generating temporary JWT token for development testing"

    # Get JWT secret from environment (same as main application)
    $jwtSecret = $env:JWT_SECRET
    if (-not $jwtSecret) {
        # Generate a secure random secret for development
        $randomBytes = New-Object byte[] 32
        $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
        $rng.GetBytes($randomBytes)
        $jwtSecret = [Convert]::ToBase64String($randomBytes)
        Write-Warning "JWT_SECRET not found in environment, generated secure development secret"
    }

    # Create JWT payload
    $now = [DateTimeOffset]::UtcNow
    $expiration = $now.AddHours($TestConfig.ExpirationHours)

    $payload = @{
        user_id = $UserID
        email = $Email
        role = $Role
        exp = $expiration.ToUnixTimeSeconds()
        iat = $now.ToUnixTimeSeconds()
        iss = "secure-email-dev"
        aud = "secure-email-api"
        temp_token = $true  # Mark as temporary token
    }

    # Convert to JSON
    $payloadJson = $payload | ConvertTo-Json -Compress

    # Create JWT header
    $header = @{
        alg = "HS256"
        typ = "JWT"
    }
    $headerJson = $header | ConvertTo-Json -Compress

    # Base64 encode header and payload
    $headerBytes = [System.Text.Encoding]::UTF8.GetBytes($headerJson)
    $payloadBytes = [System.Text.Encoding]::UTF8.GetBytes($payloadJson)

    $headerB64 = [Convert]::ToBase64String($headerBytes).Replace('+', '-').Replace('/', '_').TrimEnd('=')
    $payloadB64 = [Convert]::ToBase64String($payloadBytes).Replace('+', '-').Replace('/', '_').TrimEnd('=')

    # Create signature
    $signatureInput = "$headerB64.$payloadB64"
    $signatureBytes = [System.Security.Cryptography.HMACSHA256]::new([System.Text.Encoding]::UTF8.GetBytes($jwtSecret)).ComputeHash([System.Text.Encoding]::UTF8.GetBytes($signatureInput))
    $signatureB64 = [Convert]::ToBase64String($signatureBytes).Replace('+', '-').Replace('/', '_').TrimEnd('=')

    # Combine to create JWT
    $jwtToken = "$headerB64.$payloadB64.$signatureB64"

    Write-Success "Temporary JWT token generated successfully"
    Write-Info "Token expires: $($expiration.ToString('yyyy-MM-dd HH:mm:ss UTC'))"

    return $jwtToken
}

function Test-TempToken {
    param([string]$Token)

    Write-Info "Testing temporary token with API endpoints"

    # Test 1: User compliance status endpoint
    Write-Info "Testing user compliance status endpoint"
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/user/compliance/status" -Headers @{
        "Authorization" = "Bearer $Token"
    }

    if ($response.Success) {
        Write-Success "User compliance status endpoint test PASSED"
        Write-Info "Response: $($response.Data | ConvertTo-Json)"
    } else {
        Write-Error "User compliance status endpoint test FAILED"
        Write-Info "Status: $($response.StatusCode)"
        Write-Info "Error: $($response.Error)"
    }

    # Test 2: User compliance policies endpoint
    Write-Info "Testing user compliance policies endpoint"
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/user/compliance/policies" -Headers @{
        "Authorization" = "Bearer $Token"
    }

    if ($response.Success) {
        Write-Success "User compliance policies endpoint test PASSED"
        Write-Info "Response: $($response.Data | ConvertTo-Json)"
    } else {
        Write-Error "User compliance policies endpoint test FAILED"
        Write-Info "Status: $($response.StatusCode)"
        Write-Info "Error: $($response.Error)"
    }
}

function Generate-AdminToken {
    param([string]$UserID, [string]$Email)

    Write-Info "Generating temporary admin JWT token"

    # Generate admin token with admin role
    $adminToken = Generate-TempJWTToken -UserID $UserID -Email $Email -Role "admin"

    # Test admin endpoints
    Write-Info "Testing admin transparency settings endpoint"
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/admin/compliance/settings/user-transparency" -Headers @{
        "Authorization" = "Bearer $adminToken"
    }

    if ($response.Success) {
        Write-Success "Admin transparency settings endpoint test PASSED"
        Write-Info "Response: $($response.Data | ConvertTo-Json)"
    } else {
        Write-Error "Admin transparency settings endpoint test FAILED"
        Write-Info "Status: $($response.StatusCode)"
        Write-Info "Error: $($response.Error)"
    }

    return $adminToken
}

# Main function
function Start-TempTokenGeneration {
    Write-Output ""
    Write-Output "TEMPORARY TOKEN GENERATOR (DEVELOPMENT ONLY)"
    Write-Output "Base URL: $($TestConfig.BaseUrl)"
    Write-Output "User Email: $($TestConfig.UserEmail)"
    Write-Output "User ID: $($TestConfig.UserID)"
    Write-Output "Expiration: $($TestConfig.ExpirationHours) hours"

    Write-Output ""
    Write-Warning "WARNING: This generates temporary tokens for development testing only!"
    Write-Warning "Remove this script before production deployment."

    # Generate user token
    Write-Output ""
    Write-Output "=" * 80
    Write-Output " Generating User Token"
    Write-Output "=" * 80

    $userToken = Generate-TempJWTToken -UserID $TestConfig.UserID -Email $TestConfig.UserEmail -Role "user"

    # Test user token
    Test-TempToken -Token $userToken

    # Generate admin token
    Write-Output ""
    Write-Output "=" * 80
    Write-Output " Generating Admin Token"
    Write-Output "=" * 80

    $adminToken = Generate-AdminToken -UserID $TestConfig.UserID -Email $TestConfig.UserEmail

    # Display results
    Write-Output ""
    Write-Output "=" * 80
    Write-Output " Temporary Tokens Generated Successfully"
    Write-Output "=" * 80

    Write-Output ""
    Write-Output "User Token (first 50 chars):"
    Write-Output "  $($userToken.Substring(0, [Math]::Min(50, $userToken.Length)))..."

    Write-Output ""
    Write-Output "Admin Token (first 50 chars):"
    Write-Output "  $($adminToken.Substring(0, [Math]::Min(50, $adminToken.Length)))..."

    Write-Output ""
    Write-Output "Next Steps:"
    Write-Output "  Run Micro-Iteration 4.31 tests with these tokens:"
    Write-Output "  .\scripts\test_user_compliance_transparency.ps1 -UserToken $userToken -AdminToken $adminToken -EnableUserPortal"

    Write-Output ""
    Write-Output "Notes:"
    Write-Output "  - Tokens expire in $($TestConfig.ExpirationHours) hours"
    Write-Output "  - These are temporary tokens for development only"
    Write-Output "  - Remove this script before production deployment"

    # Save tokens to environment variables
    $env:TEMP_USER_TOKEN = $userToken
    $env:TEMP_ADMIN_TOKEN = $adminToken

    Write-Output ""
    Write-Output "Tokens saved to environment variables:"
    Write-Output "  `$env:TEMP_USER_TOKEN"
    Write-Output "  `$env:TEMP_ADMIN_TOKEN"
}

# Script execution
Start-TempTokenGeneration
