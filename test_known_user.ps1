# Test Known User
# Tests login with the test user that has a known TOTP secret

param(
    [string]$BaseUrl = "http://localhost:8080"
)

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "[SUCCESS] $Message" "Green" }
function Write-Error { param([string]$Message) Write-ColorOutput "[ERROR] $Message" "Red" }
function Write-Info { param([string]$Message) Write-ColorOutput "[INFO] $Message" "Blue" }

# Test with the known test user
$testEmail = "test@securesystem.email"
$testPassword = "TestPassword123!"
$testTOTPCode = "123456"  # Known test TOTP code

Write-Info "Testing login with known test user..."
Write-Info "Email: $testEmail"
Write-Info "TOTP Code: $testTOTPCode"

$loginData = @{
    email = $testEmail
    password = $testPassword
    totp_code = $testTOTPCode
}

$loginJson = $loginData | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $loginJson
    Write-Success "Login successful!"
    Write-Info "Access Token: $($loginResponse.access_token.Substring(0, [Math]::Min(20, $loginResponse.access_token.Length)))..."
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












