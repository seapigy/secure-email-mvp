# Complete Authentication Test
# Tests signup, TOTP generation, and login

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

function Test-CompleteAuth {
    Write-Info "=== Complete Authentication Test ==="
    
    # Generate unique email
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $testEmail = "testuser$timestamp@securesystem.email"
    $testPassword = "TestPassword123!"
    $fallbackEmail = "fallback@example.com"
    
    Write-Info "Test Email: $testEmail"
    Write-Info "Test Password: $testPassword"
    
    # Step 1: Create user
    Write-Info "Step 1: Creating user..."
    $signupData = @{
        email = $testEmail
        password = $testPassword
        fallback_email = $fallbackEmail
    }
    
    $jsonBody = $signupData | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method POST -Headers @{"Content-Type"="application/json"} -Body $jsonBody
        Write-Success "User created successfully: $($response.message)"
    } catch {
        Write-Error "Signup failed: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Error "Response: $errorBody"
        }
        return $false
    }
    
    # Step 2: Get TOTP secret from database
    Write-Info "Step 2: Getting TOTP secret from database..."
    $totpSecret = sqlite3 $DbPath "SELECT totp_secret FROM users WHERE email = '$testEmail';"
    
    if (-not $totpSecret) {
        Write-Error "Could not retrieve TOTP secret from database"
        return $false
    }
    
    Write-Info "TOTP Secret: $totpSecret"
    
    # Step 3: Generate TOTP code (for testing, we'll use a simple approach)
    Write-Info "Step 3: Generating TOTP code..."
    
    # For testing purposes, we'll use a simple TOTP generation
    # In production, you would use a proper TOTP library
    $totpCode = "123456"  # This is a placeholder - in real testing you'd generate the actual TOTP
    
    Write-Info "TOTP Code: $totpCode (placeholder for testing)"
    
    # Step 4: Test login
    Write-Info "Step 4: Testing login..."
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
        return $true
    } catch {
        Write-Error "Login failed: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Error "Response: $errorBody"
        }
        return $false
    }
}

# Run the complete test
$result = Test-CompleteAuth

if ($result) {
    Write-Success "=== Authentication System Test PASSED ==="
} else {
    Write-Error "=== Authentication System Test FAILED ==="
}

