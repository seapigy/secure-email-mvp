# Test Login with Valid TOTP
# Tests login with a dynamically generated TOTP code

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$DbPath = "C:\var\db\secure-email.db"
)

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "[SUCCESS] $Message" "Green" }
function Write-Error { param([string]$Message) Write-ColorOutput "[ERROR] $Message" "Red" }
function Write-Info { param([string]$Message) Write-ColorOutput "[INFO] $Message" "Blue" }

function Get-TOTPCode {
    param([string]$Secret)
    
    try {
        $totpCode = & .\totp_generator.exe $Secret
        return $totpCode.Trim()
    } catch {
        Write-Error "Failed to generate TOTP code: $($_.Exception.Message)"
        return $null
    }
}

# Test with the known test user
$testEmail = "test@securesystem.email"
$testPassword = "TestPassword123!"

Write-Info "Testing login with dynamically generated TOTP code..."
Write-Info "Email: $testEmail"

# Step 1: Get TOTP secret from database
Write-Info "Step 1: Getting TOTP secret from database..."
$totpSecret = sqlite3 $DbPath "SELECT totp_secret FROM users WHERE email = '$testEmail';"

if (-not $totpSecret) {
    Write-Error "Could not retrieve TOTP secret from database"
    exit 1
}

Write-Info "TOTP Secret: $totpSecret"

# Step 2: Generate valid TOTP code
Write-Info "Step 2: Generating valid TOTP code..."
$totpCode = Get-TOTPCode $totpSecret

if (-not $totpCode) {
    Write-Error "Failed to generate TOTP code"
    exit 1
}

Write-Info "Generated TOTP Code: $totpCode"

# Step 3: Test login
Write-Info "Step 3: Testing login..."
$loginData = @{
    email = $testEmail
    password = $testPassword
    totp_code = $totpCode
}

$loginJson = $loginData | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $loginJson
    Write-Success "Login successful!"
    Write-Info "Access Token: $($loginResponse.access_token.Substring(0, [Math]::Min(20, $loginResponse.access_token.Length)))..."
    Write-Info "User ID: $($loginResponse.user_id)"
    Write-Info "Email: $($loginResponse.email)"
    Write-Success "=== Authentication System Test PASSED ==="
} catch {
    Write-Error "Login failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $errorBody = $reader.ReadToEnd()
        Write-Error "Response: $errorBody"
    }
    Write-Error "=== Authentication System Test FAILED ==="
}

