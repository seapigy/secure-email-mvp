# Debug Geo-Restrictions Test
# Tests the geo-restriction functionality with detailed error reporting

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "geotest$(Get-Date -Format 'yyyyMMddHHmmss')@securesystem.email",
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
    Write-Info "Making $Method request to: $uri"

    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            Write-Info "Request body: $jsonBody"
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -Body $jsonBody
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }
        Write-Success "Request successful"
        return @{ Success = $true; Data = $response }
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Error "HTTP $($errorResponse.StatusCode): $errorBody"
            Write-Error "Full error response: $errorBody"
            try {
                $errorJson = $errorBody | ConvertFrom-Json
                return @{ Success = $false; Error = $errorJson.error; StatusCode = $errorResponse.StatusCode; FullError = $errorBody }
            }
            catch {
                return @{ Success = $false; Error = $errorBody; StatusCode = $errorResponse.StatusCode; FullError = $errorBody }
            }
        }
        Write-Error "Request failed: $($_.Exception.Message)"
        return @{ Success = $false; Error = $_.Exception.Message }
    }
}

# Main test execution
Write-Info "=== Debug Geo-Restrictions Test ==="

# Test health check
Write-Info "Step 1: Testing health check..."
$healthResult = Invoke-ApiRequest -Method "GET" -Endpoint "/health"
if (-not $healthResult.Success) {
    Write-Error "Health check failed. Cannot continue."
    exit 1
}
Write-Success "Health check passed"

# Test user signup
Write-Info "Step 2: Testing user signup..."
$signupData = @{
    email = $TestEmail
    password = $TestPassword
    fallback_email = "fallback@example.com"
}

$signupResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData
if (-not $signupResult.Success) {
    Write-Error "Signup failed: $($signupResult.Error)"
    exit 1
}
Write-Success "User signup successful"

# Test user login
Write-Info "Step 3: Testing user login..."
$totpCode = Get-TOTPCode $TestEmail
if (-not $totpCode) {
    Write-Error "Failed to get TOTP code for login"
    exit 1
}

Write-Info "Using TOTP code: $totpCode"

$loginData = @{
    email = $TestEmail
    password = $TestPassword
    totp_code = $totpCode
}

$loginResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData
if (-not $loginResult.Success) {
    Write-Error "Login failed: $($loginResult.Error)"
    exit 1
}

$accessToken = $loginResult.Data.access_token
Write-Success "User login successful. Access token obtained."

# Test send email
Write-Info "Step 4: Testing email sending..."
$emailData = @{
    recipient = "recipient@example.com"
    subject = "Test Email with Geo-Restrictions"
    body = "This is a test email for geo-restriction testing."
    expiresAt = (Get-Date).AddHours(24).ToString("yyyy-MM-ddTHH:mm:ssZ")
}

$emailResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $emailData -Token $accessToken
if (-not $emailResult.Success) {
    Write-Error "Email sending failed: $($emailResult.Error)"
    Write-Error "Full error: $($emailResult.FullError)"
    exit 1
}

$emailId = $emailResult.Data.id
Write-Success "Email sent successfully. Email ID: $emailId"

# Test geo-restriction functionality
Write-Info "Step 5: Testing geo-restriction functionality..."

# Test get geo-restriction rules
Write-Info "5.1: Testing get geo-restriction rules..."
$rulesResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$emailId/geo-restrictions" -Token $accessToken
if ($rulesResult.Success) {
    Write-Success "Get geo-restriction rules successful"
} else {
    Write-Error "Get geo-restriction rules failed: $($rulesResult.Error)"
}

# Test create geo-restriction rule
Write-Info "5.2: Testing create geo-restriction rule..."
$ruleData = @{
    type = "allow"
    countries = @("US", "CA")
    cities = @("New York", "Toronto")
    description = "Allow US and Canada access"
}

$createRuleResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/$emailId/geo-restrictions" -Body $ruleData -Token $accessToken
if ($createRuleResult.Success) {
    $ruleId = $createRuleResult.Data.id
    Write-Success "Create geo-restriction rule successful. Rule ID: $ruleId"

    # Test update geo-restriction rule
    Write-Info "5.3: Testing update geo-restriction rule..."
    $updateRuleData = @{
        type = "allow"
        countries = @("US", "CA", "UK")
        cities = @("New York", "Toronto", "London")
        description = "Updated: Allow US, Canada, and UK access"
    }

    $updateRuleResult = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/email/$emailId/geo-restrictions/$ruleId" -Body $updateRuleData -Token $accessToken
    if ($updateRuleResult.Success) {
        Write-Success "Update geo-restriction rule successful"
    } else {
        Write-Error "Update geo-restriction rule failed: $($updateRuleResult.Error)"
    }

    # Test delete geo-restriction rule
    Write-Info "5.4: Testing delete geo-restriction rule..."
    $deleteRuleResult = Invoke-ApiRequest -Method "DELETE" -Endpoint "/api/email/$emailId/geo-restrictions/$ruleId" -Token $accessToken
    if ($deleteRuleResult.Success) {
        Write-Success "Delete geo-restriction rule successful"
    } else {
        Write-Error "Delete geo-restriction rule failed: $($deleteRuleResult.Error)"
    }
} else {
    Write-Error "Create geo-restriction rule failed: $($createRuleResult.Error)"
}

# Test create block rule
Write-Info "5.5: Testing create block geo-restriction rule..."
$blockRuleData = @{
    type = "block"
    countries = @("XX", "YY")
    cities = @("BlockedCity")
    description = "Block specific countries"
}

$blockRuleResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/$emailId/geo-restrictions" -Body $blockRuleData -Token $accessToken
if ($blockRuleResult.Success) {
    Write-Success "Create block geo-restriction rule successful"
} else {
    Write-Error "Create block geo-restriction rule failed: $($blockRuleResult.Error)"
}

# Test get geo-restriction config
Write-Info "5.6: Testing get geo-restriction config..."
$configResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$emailId/geo-restrictions/config" -Token $accessToken
if ($configResult.Success) {
    Write-Success "Get geo-restriction config successful"
} else {
    Write-Error "Get geo-restriction config failed: $($configResult.Error)"
}

# Test update geo-restriction config
Write-Info "5.7: Testing update geo-restriction config..."
$updateConfigData = @{
    enabled = $true
    default_action = "block"
    strict_mode = $false
    log_violations = $true
    block_on_geolocation_failure = $true
}

$updateConfigResult = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/email/$emailId/geo-restrictions/config" -Body $updateConfigData -Token $accessToken
if ($updateConfigResult.Success) {
    Write-Success "Update geo-restriction config successful"
} else {
    Write-Error "Update geo-restriction config failed: $($updateConfigResult.Error)"
}

# Test get geo-restriction status
Write-Info "5.8: Testing get geo-restriction status..."
$statusResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$emailId/geo-restrictions/status" -Token $accessToken
if ($statusResult.Success) {
    Write-Success "Get geo-restriction status successful"
} else {
    Write-Error "Get geo-restriction status failed: $($statusResult.Error)"
}

# Test view email with geo-restrictions
Write-Info "5.9: Testing view email with geo-restrictions..."
$viewEmailResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$emailId" -Token $accessToken
if ($viewEmailResult.Success) {
    Write-Success "View email with geo-restrictions successful"
} else {
    Write-Error "View email with geo-restrictions failed: $($viewEmailResult.Error)"
}

Write-Success "=== Debug Geo-Restrictions Test Completed ==="
