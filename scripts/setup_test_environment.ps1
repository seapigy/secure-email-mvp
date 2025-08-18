# =============================================================================
# SECURE EMAIL MVP - TEST ENVIRONMENT SETUP SCRIPT
# =============================================================================
# PowerShell script to set up test environment for Micro-Iteration 4.31
# Creates test users and obtains authentication tokens

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestUserEmail = "testuser@securesystem.email",
    [string]$TestUserPassword = "TestPassword123!",
    [string]$AdminUserEmail = "admin@securesystem.email",
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
    Write-ColorOutput "✅ $Message" "Green"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "❌ $Message" "Red"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ️  $Message" "Cyan"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" "Yellow"
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
        $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $requestHeaders -Body $Body -ErrorAction Stop
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

function Test-HealthCheck {
    Write-Info "Testing API server health..."
    
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/health"
    
    if ($response.Success) {
        Write-Success "API server is healthy"
        return $true
    } else {
        Write-Error "API server health check failed: $($response.Error)"
        return $false
    }
}

function Create-TestUser {
    param(
        [string]$Email,
        [string]$Password,
        [string]$UserType
    )
    
    Write-Info "Creating $UserType user: $Email"
    
    $signupData = @{
        email = $Email
        password = $Password
    }
    
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body ($signupData | ConvertTo-Json)
    
    if ($response.Success) {
        Write-Success "$UserType user created successfully"
        return $true
    } else {
        if ($response.StatusCode -eq 409) {
            Write-Warning "$UserType user already exists"
            return $true
        } else {
            Write-Error "Failed to create $UserType user: $($response.Error)"
            return $false
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
        totp_code = "123456"  # Default TOTP for testing
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

function Setup-EnterpriseOrganization {
    param([string]$AdminToken)
    
    Write-Info "Setting up enterprise organization for admin user..."
    
    $enterpriseData = @{
        organization_name = "Test Enterprise Corp"
        domain = "securesystem.email"
        compliance_frameworks = @("GDPR", "HIPAA")
        contact_email = $TestConfig.AdminUserEmail
    }
    
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/admin/compliance/enterprise" -Headers @{
        "Authorization" = "Bearer $AdminToken"
    } -Body ($enterpriseData | ConvertTo-Json)
    
    if ($response.Success) {
        Write-Success "Enterprise organization created successfully"
        return $true
    } else {
        if ($response.StatusCode -eq 409) {
            Write-Warning "Enterprise organization already exists"
            return $true
        } else {
            Write-Error "Failed to create enterprise organization: $($response.Error)"
            return $false
        }
    }
}

# Main setup function
function Start-TestEnvironmentSetup {
    Write-Host "`n" -NoNewline
    Write-Host "🚀 Setting up test environment for Micro-Iteration 4.31" -ForegroundColor Yellow
    Write-Host "Base URL: $($TestConfig.BaseUrl)" -ForegroundColor Gray
    
    # Step 1: Health check
    if (-not (Test-HealthCheck)) {
        Write-Error "Cannot proceed without a healthy API server"
        Write-Host "Please start the API server with: go run cmd/api/main.go" -ForegroundColor Yellow
        exit 1
    }
    
    # Step 2: Create test users
    Write-Host "`n📝 Creating test users..." -ForegroundColor Cyan
    
    $userCreated = Create-TestUser -Email $TestConfig.TestUserEmail -Password $TestConfig.TestUserPassword -UserType "Test"
    $adminCreated = Create-TestUser -Email $TestConfig.AdminUserEmail -Password $TestConfig.AdminUserPassword -UserType "Admin"
    
    if (-not $userCreated -or -not $adminCreated) {
        Write-Error "Failed to create required test users"
        exit 1
    }
    
    # Step 3: Get authentication tokens
    Write-Host "`n🔐 Obtaining authentication tokens..." -ForegroundColor Cyan
    
    $userToken = Get-UserToken -Email $TestConfig.TestUserEmail -Password $TestConfig.TestUserPassword -UserType "Test"
    $adminToken = Get-UserToken -Email $TestConfig.AdminUserEmail -Password $TestConfig.AdminUserPassword -UserType "Admin"
    
    if (-not $userToken -or -not $adminToken) {
        Write-Error "Failed to obtain required authentication tokens"
        exit 1
    }
    
    # Step 4: Setup enterprise organization
    Write-Host "`n🏢 Setting up enterprise organization..." -ForegroundColor Cyan
    Setup-EnterpriseOrganization -AdminToken $adminToken
    
    # Step 5: Display results and next steps
    Write-Host "`n" -NoNewline
    Write-Host "=" * 80 -ForegroundColor Green
    Write-Host "✅ Test Environment Setup Complete" -ForegroundColor Green
    Write-Host "=" * 80 -ForegroundColor Green
    
    Write-Host "`n📋 Test Configuration:" -ForegroundColor White
    Write-Host "  Base URL: $($TestConfig.BaseUrl)" -ForegroundColor Gray
    Write-Host "  Test User: $($TestConfig.TestUserEmail)" -ForegroundColor Gray
    Write-Host "  Admin User: $($TestConfig.AdminUserEmail)" -ForegroundColor Gray
    
    Write-Host "`n🔑 Authentication Tokens:" -ForegroundColor White
    Write-Host "  User Token: $($userToken.Substring(0, [Math]::Min(20, $userToken.Length)))..." -ForegroundColor Gray
    Write-Host "  Admin Token: $($adminToken.Substring(0, [Math]::Min(20, $adminToken.Length)))..." -ForegroundColor Gray
    
    Write-Host "`n🚀 Next Steps:" -ForegroundColor White
    Write-Host "  Run the user compliance transparency tests:" -ForegroundColor Yellow
    Write-Host "  .\scripts\test_user_compliance_transparency.ps1 -UserToken '$userToken' -AdminToken '$adminToken' -EnableUserPortal" -ForegroundColor Cyan
    
    Write-Host "`n💡 Notes:" -ForegroundColor White
    Write-Host "  - TOTP code for testing: 123456" -ForegroundColor Gray
    Write-Host "  - Enterprise organization is set up for compliance testing" -ForegroundColor Gray
    Write-Host "  - Both users are ready for Micro-Iteration 4.31 testing" -ForegroundColor Gray
    
    # Save tokens to environment variables for easy access
    $env:TEST_USER_TOKEN = $userToken
    $env:TEST_ADMIN_TOKEN = $adminToken
    
    Write-Host "`nTokens saved to environment variables:" -ForegroundColor White
    Write-Host "  `$env:TEST_USER_TOKEN" -ForegroundColor Gray
    Write-Host "  `$env:TEST_ADMIN_TOKEN" -ForegroundColor Gray
}

# Script execution
Start-TestEnvironmentSetup
