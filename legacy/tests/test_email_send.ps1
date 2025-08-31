# Simple Email Send Test
# Tests email sending with fresh authentication

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

# Use the known test user
$testEmail = "test@securesystem.email"
$testPassword = "TestPassword123!"

Write-Info "Testing email sending with fresh authentication..."

# Get TOTP code
$totpCode = & powershell -ExecutionPolicy Bypass -File "scripts/get_totp_code.ps1" -Email $testEmail
Write-Info "TOTP Code: $totpCode"

# Login
$loginData = @{
    email = $testEmail
    password = $testPassword
    totp_code = $totpCode
}

$loginJson = $loginData | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $loginJson
    $accessToken = $loginResponse.access_token
    Write-Success "Login successful. Token: $($accessToken.Substring(0, 20))..."
} catch {
    Write-Error "Login failed: $($_.Exception.Message)"
    exit 1
}

# Send email
$emailData = @{
    recipient = "test@example.com"
    subject = "Test Email"
    body = "This is a test email body."
}

$emailJson = $emailData | ConvertTo-Json

try {
    $emailResponse = Invoke-RestMethod -Uri "$BaseUrl/api/email/send" -Method POST -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer $accessToken"} -Body $emailJson
    Write-Success "Email sent successfully!"
    Write-Info "Response: $($emailResponse | ConvertTo-Json)"
} catch {
    Write-Error "Email sending failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $errorBody = $reader.ReadToEnd()
        Write-Error "Response: $errorBody"
    }
}




















