# Test Geo-Restrictions with Existing Email
# Tests geo-restriction functionality using an existing email from the database

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
            try {
                $errorJson = $errorBody | ConvertFrom-Json
                return @{ Success = $false; Error = $errorJson.error; StatusCode = $errorResponse.StatusCode }
            }
            catch {
                return @{ Success = $false; Error = $errorBody; StatusCode = $errorResponse.StatusCode }
            }
        }
        Write-Error "Request failed: $($_.Exception.Message)"
        return @{ Success = $false; Error = $_.Exception.Message }
    }
}

# Main test execution
Write-Info "=== Test Geo-Restrictions with Existing Email ==="

# Get an existing email from the database
Write-Info "Step 1: Getting existing email from database..."
$existingEmail = sqlite3 $DbPath "SELECT email_id FROM emails WHERE sender_id = (SELECT id FROM users WHERE email = '$TestEmail') LIMIT 1;"

if (-not $existingEmail) {
    Write-Error "No existing email found for user $TestEmail"
    exit 1
}

Write-Info "Using existing email ID: $existingEmail"

# Test user login
Write-Info "Step 2: Testing user login..."
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

# Test geo-restriction functionality
Write-Info "Step 3: Testing geo-restriction functionality..."

# Test get geo-restriction rules
Write-Info "3.1: Testing get geo-restriction rules..."
$rulesResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$existingEmail/geo-restrictions" -Token $accessToken
if ($rulesResult.Success) {
    Write-Success "Get geo-restriction rules successful"
    Write-Info "Rules: $($rulesResult.Data | ConvertTo-Json)"
} else {
    Write-Error "Get geo-restriction rules failed: $($rulesResult.Error)"
}

# Test create geo-restriction rule
Write-Info "3.2: Testing create geo-restriction rule..."
$ruleData = @{
    type = "allow"
    countries = @("US", "CA")
    cities = @("New York", "Toronto")
    description = "Allow US and Canada access"
}

$createRuleResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/$existingEmail/geo-restrictions" -Body $ruleData -Token $accessToken
if ($createRuleResult.Success) {
    $ruleId = $createRuleResult.Data.rules[0].id
    Write-Success "Create geo-restriction rule successful. Rule ID: $ruleId"
    
    # Test update geo-restriction rule
    Write-Info "3.3: Testing update geo-restriction rule..."
    $updateRuleData = @{
        type = "allow"
        countries = @("US", "CA", "UK")
        cities = @("New York", "Toronto", "London")
        description = "Updated: Allow US, Canada, and UK access"
    }
    
    $updateRuleResult = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/email/$existingEmail/geo-restrictions/$ruleId" -Body $updateRuleData -Token $accessToken
    if ($updateRuleResult.Success) {
        Write-Success "Update geo-restriction rule successful"
    } else {
        Write-Error "Update geo-restriction rule failed: $($updateRuleResult.Error)"
    }
    
    # Test delete geo-restriction rule
    Write-Info "3.4: Testing delete geo-restriction rule..."
    $deleteRuleResult = Invoke-ApiRequest -Method "DELETE" -Endpoint "/api/email/$existingEmail/geo-restrictions/$ruleId" -Token $accessToken
    if ($deleteRuleResult.Success) {
        Write-Success "Delete geo-restriction rule successful"
    } else {
        Write-Error "Delete geo-restriction rule failed: $($deleteRuleResult.Error)"
    }
} else {
    Write-Error "Create geo-restriction rule failed: $($createRuleResult.Error)"
}

# Test create block rule
Write-Info "3.5: Testing create block geo-restriction rule..."
$blockRuleData = @{
    type = "block"
    countries = @("XX", "YY")
    cities = @("BlockedCity")
    description = "Block specific countries"
}

$blockRuleResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/$existingEmail/geo-restrictions" -Body $blockRuleData -Token $accessToken
if ($blockRuleResult.Success) {
    Write-Success "Create block geo-restriction rule successful"
} else {
    Write-Error "Create block geo-restriction rule failed: $($blockRuleResult.Error)"
}

# Test get geo-restriction config
Write-Info "3.6: Testing get geo-restriction config..."
$configResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$existingEmail/geo-restrictions/config" -Token $accessToken
if ($configResult.Success) {
    Write-Success "Get geo-restriction config successful"
    Write-Info "Config: $($configResult.Data | ConvertTo-Json)"
} else {
    Write-Error "Get geo-restriction config failed: $($configResult.Error)"
}

# Test update geo-restriction config
Write-Info "3.7: Testing update geo-restriction config..."
$updateConfigData = @{
    enabled = $true
    default_action = "block"
    strict_mode = $false
    log_violations = $true
    block_on_geolocation_failure = $true
}

$updateConfigResult = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/email/$existingEmail/geo-restrictions/config" -Body $updateConfigData -Token $accessToken
if ($updateConfigResult.Success) {
    Write-Success "Update geo-restriction config successful"
} else {
    Write-Error "Update geo-restriction config failed: $($updateConfigResult.Error)"
}

# Test get geo-restriction status
Write-Info "3.8: Testing get geo-restriction status..."
$statusResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$existingEmail/geo-restrictions/status" -Token $accessToken
if ($statusResult.Success) {
    Write-Success "Get geo-restriction status successful"
    Write-Info "Status: $($statusResult.Data | ConvertTo-Json)"
} else {
    Write-Error "Get geo-restriction status failed: $($statusResult.Error)"
}

# Test view email with geo-restrictions
Write-Info "3.9: Testing view email with geo-restrictions..."
$viewEmailResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/$existingEmail" -Token $accessToken
if ($viewEmailResult.Success) {
    Write-Success "View email with geo-restrictions successful"
} else {
    Write-Error "View email with geo-restrictions failed: $($viewEmailResult.Error)"
}

Write-Success "=== Geo-Restrictions Test with Existing Email Completed ==="
