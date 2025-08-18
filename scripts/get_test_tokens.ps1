# =============================================================================
# SECURE EMAIL MVP - GET TEST TOKENS SCRIPT
# =============================================================================
# PowerShell script to get authentication tokens for existing test users

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestUserEmail = "testuser431@securesystem.email",
    [string]$TestUserPassword = "TestPassword123!",
    [string]$AdminUserEmail = "admin431@securesystem.email",
    [string]$AdminUserPassword = "AdminPassword123!"
)

# Test configuration
$TestConfig = @{
    BaseUrl = $BaseUrl
    TestUserEmail = $TestUserEmail
    TestUserPassword = $TestUserPassword
    AdminUserEmail = $AdminUserEmail
    AdminUserPassword = $AdminUserPassword
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

function Get-UserToken {
    param(
        [string]$Email,
        [string]$Password,
        [string]$UserType
    )
    
    Write-Info "Logging in $UserType user: $Email"
    
    $loginData = @{
        email = $Email
        password = $Password
        totp_code = "369496"  # Correct TOTP for JBSWY3DPEHPK3PXP secret
    }
    
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body ($loginData | ConvertTo-Json)
    
    if ($response.Success) {
        $token = $response.Data.access_token
        if (-not $token) {
            $token = $response.Data.token  # Try alternative field name
        }
        
        if ($token) {
            Write-Success "$UserType login successful"
            Write-Info "Token: $($token.Substring(0, [Math]::Min(20, $token.Length)))..."
            return $token
        } else {
            Write-Error "$UserType login response missing token"
            return $null
        }
    } else {
        Write-Error "$UserType login failed: $($response.Error)"
        return $null
    }
}

# Main function
function Start-GetTokens {
    Write-Host ""
    Write-Host "Getting authentication tokens for existing test users" -ForegroundColor Yellow
    Write-Host "Base URL: $($TestConfig.BaseUrl)" -ForegroundColor Gray
    
    # Get authentication tokens
    Write-Host ""
    Write-Host "Obtaining authentication tokens..." -ForegroundColor Cyan
    
    $userToken = Get-UserToken -Email $TestConfig.TestUserEmail -Password $TestConfig.TestUserPassword -UserType "Test"
    Start-Sleep -Seconds 2  # Wait between requests to avoid rate limiting
    $adminToken = Get-UserToken -Email $TestConfig.AdminUserEmail -Password $TestConfig.AdminUserPassword -UserType "Admin"
    
    if (-not $userToken -or -not $adminToken) {
        Write-Error "Failed to obtain required authentication tokens"
        exit 1
    }
    
    # Display results and next steps
    Write-Host ""
    Write-Host "=" * 80 -ForegroundColor Green
    Write-Host "Authentication Tokens Obtained" -ForegroundColor Green
    Write-Host "=" * 80 -ForegroundColor Green
    
    Write-Host ""
    Write-Host "Test Configuration:" -ForegroundColor White
    Write-Host "  Base URL: $($TestConfig.BaseUrl)" -ForegroundColor Gray
    Write-Host "  Test User: $($TestConfig.TestUserEmail)" -ForegroundColor Gray
    Write-Host "  Admin User: $($TestConfig.AdminUserEmail)" -ForegroundColor Gray
    
    Write-Host ""
    Write-Host "Authentication Tokens:" -ForegroundColor White
    Write-Host "  User Token: $($userToken.Substring(0, [Math]::Min(20, $userToken.Length)))..." -ForegroundColor Gray
    Write-Host "  Admin Token: $($adminToken.Substring(0, [Math]::Min(20, $adminToken.Length)))..." -ForegroundColor Gray
    
    Write-Host ""
    Write-Host "Next Steps:" -ForegroundColor White
    Write-Host "  Run the user compliance transparency tests:" -ForegroundColor Yellow
    Write-Host "  .\scripts\test_user_compliance_transparency.ps1 -UserToken '$userToken' -AdminToken '$adminToken' -EnableUserPortal" -ForegroundColor Cyan
    
    Write-Host ""
    Write-Host "Notes:" -ForegroundColor White
    Write-Host "  - TOTP code for testing: 123456" -ForegroundColor Gray
    Write-Host "  - Both users are ready for Micro-Iteration 4.31 testing" -ForegroundColor Gray
    
    # Save tokens to environment variables for easy access
    $env:TEST_USER_TOKEN = $userToken
    $env:TEST_ADMIN_TOKEN = $adminToken
    
    Write-Host ""
    Write-Host "Tokens saved to environment variables:" -ForegroundColor White
    Write-Host "  `$env:TEST_USER_TOKEN" -ForegroundColor Gray
    Write-Host "  `$env:TEST_ADMIN_TOKEN" -ForegroundColor Gray
}

# Script execution
Start-GetTokens
