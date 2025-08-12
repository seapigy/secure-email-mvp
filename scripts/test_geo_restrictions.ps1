# Geo-Restrictions Integration Test
# Tests the complete geo-restriction functionality with proper authentication

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",
    [string]$TestPassword = "TestPassword123!",
    [string]$DbPath = "C:\var\db\secure-email.db"
)

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "[SUCCESS] $Message" "Green" }
function Write-Error { param([string]$Message) Write-ColorOutput "[ERROR] $Message" "Red" }
function Write-Warning { param([string]$Message) Write-ColorOutput "[WARNING] $Message" "Yellow" }
function Write-Info { param([string]$Message) Write-ColorOutput "[INFO] $Message" "Blue" }

function Get-TOTPCode {
    param([string]$Email)
    
    try {
        $totpCode = & powershell -ExecutionPolicy Bypass -File "scripts/get_totp_code.ps1" -Email $Email
        return $totpCode.Trim()
    } catch {
        Write-Error "Failed to get TOTP code: $($_.Exception.Message)"
        return $null
    }
}

function Invoke-ApiRequest {
    param(
        [string]$Method = "GET",
        [string]$Endpoint,
        [object]$Body = $null,
        [string]$Token = $null
    )

    $headers = @{
        "Content-Type" = "application/json"
    }

    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }

    $uri = "$BaseUrl$Endpoint"

    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -Body $jsonBody
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }
        return @{ Success = $true; Data = $response }
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            try {
                $errorJson = $errorBody | ConvertFrom-Json
                return @{ Success = $false; Error = $errorJson.error; StatusCode = $errorResponse.StatusCode }
            }
            catch {
                return @{ Success = $false; Error = $errorBody; StatusCode = $errorResponse.StatusCode }
            }
        }
        return @{ Success = $false; Error = $_.Exception.Message }
    }
}

function Test-HealthCheck {
    Write-Info "Testing health check endpoint..."
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/health"
    if ($result.Success) {
        Write-Success "Health check passed"
        return $true
    } else {
        Write-Error "Health check failed: $($result.Error)"
        return $false
    }
}

function Test-UserSignup {
    Write-Info "Testing user signup..."
    $signupData = @{
        email = $TestEmail
        password = $TestPassword
        fallback_email = "fallback@example.com"
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData
    if ($result.Success) {
        Write-Success "User signup successful"
        return $true
    } else {
        if ($result.Error -like "*already exists*") {
            Write-Warning "User already exists, continuing with login"
            return $true
        } else {
            Write-Error "User signup failed: $($result.Error)"
            return $false
        }
    }
}

function Test-UserLogin {
    Write-Info "Testing user login..."
    
    # Get valid TOTP code
    $totpCode = Get-TOTPCode $TestEmail
    if (-not $totpCode) {
        Write-Error "Failed to get TOTP code for login"
        return $null
    }
    
    Write-Info "Using TOTP code: $totpCode"
    
    $loginData = @{
        email = $TestEmail
        password = $TestPassword
        totp_code = $totpCode
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData
    if ($result.Success) {
        Write-Success "User login successful"
        return $result.Data.access_token
    } else {
        Write-Error "User login failed: $($result.Error)"
        return $null
    }
}

function Test-SendEmail {
    param([string]$Token)
    
    Write-Info "Testing email sending..."
    $emailData = @{
        recipient_email = "recipient@example.com"
        subject = "Test Email with Geo-Restrictions"
        content = "This is a test email for geo-restriction testing."
        expires_in_hours = 24
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $emailData -Token $Token
    if ($result.Success) {
        Write-Success "Email sent successfully"
        return $result.Data.id
    } else {
        Write-Error "Email sending failed: $($result.Error)"
        return $null
    }
}

function Test-GetGeoRestrictionRules {
    param([string]$Token, [string]$EmailId)
    
    Write-Info "Testing get geo-restriction rules..."
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$EmailId/geo-restrictions" -Token $Token
    if ($result.Success) {
        Write-Success "Get geo-restriction rules successful"
        return $true
    } else {
        Write-Error "Get geo-restriction rules failed: $($result.Error)"
        return $false
    }
}

function Test-CreateGeoRestrictionRule {
    param([string]$Token, [string]$EmailId)
    
    Write-Info "Testing create geo-restriction rule..."
    $ruleData = @{
        type = "allow"
        countries = @("US", "CA")
        cities = @("New York", "Toronto")
        description = "Allow US and Canada access"
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/$EmailId/geo-restrictions" -Body $ruleData -Token $Token
    if ($result.Success) {
        Write-Success "Create geo-restriction rule successful"
        return $result.Data.id
    } else {
        Write-Error "Create geo-restriction rule failed: $($result.Error)"
        return $null
    }
}

function Test-CreateBlockRule {
    param([string]$Token, [string]$EmailId)
    
    Write-Info "Testing create block geo-restriction rule..."
    $ruleData = @{
        type = "block"
        countries = @("XX", "YY")
        cities = @("BlockedCity")
        description = "Block specific countries"
    }
    
    $result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/$EmailId/geo-restrictions" -Body $ruleData -Token $Token
    if ($result.Success) {
        Write-Success "Create block geo-restriction rule successful"
        return $result.Data.id
    } else {
        Write-Error "Create block geo-restriction rule failed: $($result.Error)"
        return $false
    }
}

function Test-UpdateGeoRestrictionRule {
    param([string]$Token, [string]$EmailId, [string]$RuleId)
    
    Write-Info "Testing update geo-restriction rule..."
    $ruleData = @{
        type = "allow"
        countries = @("US", "CA", "UK")
        cities = @("New York", "Toronto", "London")
        description = "Updated: Allow US, Canada, and UK access"
    }
    
    $result = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/email/$EmailId/geo-restrictions/$RuleId" -Body $ruleData -Token $Token
    if ($result.Success) {
        Write-Success "Update geo-restriction rule successful"
        return $true
    } else {
        Write-Error "Update geo-restriction rule failed: $($result.Error)"
        return $false
    }
}

function Test-GetGeoRestrictionConfig {
    param([string]$Token, [string]$EmailId)
    
    Write-Info "Testing get geo-restriction config..."
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$EmailId/geo-restrictions/config" -Token $Token
    if ($result.Success) {
        Write-Success "Get geo-restriction config successful"
        return $true
    } else {
        Write-Error "Get geo-restriction config failed: $($result.Error)"
        return $false
    }
}

function Test-UpdateGeoRestrictionConfig {
    param([string]$Token, [string]$EmailId)
    
    Write-Info "Testing update geo-restriction config..."
    $configData = @{
        enabled = $true
        default_action = "block"
        strict_mode = $false
        log_violations = $true
        block_on_geolocation_failure = $true
    }
    
    $result = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/email/$EmailId/geo-restrictions/config" -Body $configData -Token $Token
    if ($result.Success) {
        Write-Success "Update geo-restriction config successful"
        return $true
    } else {
        Write-Error "Update geo-restriction config failed: $($result.Error)"
        return $false
    }
}

function Test-GetGeoRestrictionStatus {
    param([string]$Token, [string]$EmailId)
    
    Write-Info "Testing get geo-restriction status..."
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$EmailId/geo-restrictions/status" -Token $Token
    if ($result.Success) {
        Write-Success "Get geo-restriction status successful"
        return $true
    } else {
        Write-Error "Get geo-restriction status failed: $($result.Error)"
        return $false
    }
}

function Test-DeleteGeoRestrictionRule {
    param([string]$Token, [string]$EmailId, [string]$RuleId)
    
    Write-Info "Testing delete geo-restriction rule..."
    $result = Invoke-ApiRequest -Method "DELETE" -Endpoint "/api/email/$EmailId/geo-restrictions/$RuleId" -Token $Token
    if ($result.Success) {
        Write-Success "Delete geo-restriction rule successful"
        return $true
    } else {
        Write-Error "Delete geo-restriction rule failed: $($result.Error)"
        return $false
    }
}

function Test-ViewEmailWithGeoRestrictions {
    param([string]$Token, [string]$EmailId)
    
    Write-Info "Testing view email with geo-restrictions..."
    $result = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$EmailId" -Token $Token
    if ($result.Success) {
        Write-Success "View email with geo-restrictions successful"
        return $true
    } else {
        Write-Error "View email with geo-restrictions failed: $($result.Error)"
        return $false
    }
}

# Main test execution
Write-Info "=== Geo-Restrictions Integration Test ==="

# Test health check
if (-not (Test-HealthCheck)) {
    Write-Error "Health check failed. Cannot continue with tests."
    exit 1
}

# Test user signup
if (-not (Test-UserSignup)) {
    Write-Error "User signup failed. Cannot continue with tests."
    exit 1
}

# Test user login
$accessToken = Test-UserLogin
if (-not $accessToken) {
    Write-Error "User login failed. Cannot continue with tests."
    exit 1
}

Write-Info "Authentication successful. Access token obtained."

# Test send email
$emailId = Test-SendEmail $accessToken
if (-not $emailId) {
    Write-Error "Email sending failed. Cannot continue with geo-restriction tests."
    exit 1
}

Write-Info "Email sent successfully. Email ID: $emailId"

# Test geo-restriction functionality
Write-Info "Testing geo-restriction functionality..."

# Test get geo-restriction rules
Test-GetGeoRestrictionRules $accessToken $emailId

# Test create geo-restriction rule
$ruleId = Test-CreateGeoRestrictionRule $accessToken $emailId

if ($ruleId) {
    # Test update geo-restriction rule
    Test-UpdateGeoRestrictionRule $accessToken $emailId $ruleId
    
    # Test delete geo-restriction rule
    Test-DeleteGeoRestrictionRule $accessToken $emailId $ruleId
}

# Test create block rule
Test-CreateBlockRule $accessToken $emailId

# Test get geo-restriction config
Test-GetGeoRestrictionConfig $accessToken $emailId

# Test update geo-restriction config
Test-UpdateGeoRestrictionConfig $accessToken $emailId

# Test get geo-restriction status
Test-GetGeoRestrictionStatus $accessToken $emailId

# Test view email with geo-restrictions
Test-ViewEmailWithGeoRestrictions $accessToken $emailId

Write-Success "=== Geo-Restrictions Integration Test Completed ==="
