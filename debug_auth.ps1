# Authentication System Debug Script
# Tests signup, login, and database operations step by step

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",
    [string]$TestPassword = "TestPassword123!",
    [string]$FallbackEmail = "fallback@example.com"
)

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "[SUCCESS] $Message" "Green" }
function Write-Error { param([string]$Message) Write-ColorOutput "[ERROR] $Message" "Red" }
function Write-Warning { param([string]$Message) Write-ColorOutput "[WARNING] $Message" "Yellow" }
function Write-Info { param([string]$Message) Write-ColorOutput "[INFO] $Message" "Blue" }

function Test-HealthCheck {
    Write-Info "Testing health check endpoint..."
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET
        Write-Success "Health check passed: $($response | ConvertTo-Json)"
        return $true
    } catch {
        Write-Error "Health check failed: $($_.Exception.Message)"
        return $false
    }
}

function Test-Signup {
    Write-Info "Testing user signup..."

    $signupData = @{
        email = $TestEmail
        password = $TestPassword
        fallback_email = $FallbackEmail
    }

    $jsonBody = $signupData | ConvertTo-Json
    Write-Info "Request body: $jsonBody"

    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method POST -Headers @{"Content-Type"="application/json"} -Body $jsonBody
        Write-Success "Signup successful: $($response | ConvertTo-Json)"
        return $true
    } catch {
        Write-Error "Signup failed: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Error "Response body: $errorBody"
        }
        return $false
    }
}

function Test-Login {
    Write-Info "Testing user login..."

    $loginData = @{
        email = $TestEmail
        password = $TestPassword
        totp_code = "123456"  # Test TOTP code
    }

    $jsonBody = $loginData | ConvertTo-Json
    Write-Info "Request body: $jsonBody"

    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $jsonBody
        Write-Success "Login successful: $($response | ConvertTo-Json)"
        return $response.access_token
    } catch {
        Write-Error "Login failed: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Error "Response body: $errorBody"
        }
        return $null
    }
}

function Test-DatabaseState {
    Write-Info "Checking database state..."

    # Check if user exists
    $userQuery = "SELECT email, totp_secret IS NOT NULL as has_totp, password_hash IS NOT NULL as has_password_hash, password IS NOT NULL as has_password FROM users WHERE email = '$TestEmail';"
    $userResult = sqlite3 secure-email.db $userQuery

    if ($userResult) {
        Write-Success "User found in database: $userResult"
    } else {
        Write-Warning "User not found in database"
    }

    # Check table schema
    Write-Info "Users table schema:"
    $schemaResult = sqlite3 secure-email.db ".schema users"
    Write-Host $schemaResult
}

function Test-EnvironmentVariables {
    Write-Info "Checking environment variables..."

    $requiredVars = @("JWT_SECRET", "SQLITE_DB", "DEBUG")

    foreach ($var in $requiredVars) {
        $value = [Environment]::GetEnvironmentVariable($var)
        if ($value) {
            Write-Success "$var is set: $value"
        } else {
            Write-Error "$var is not set"
        }
    }
}

# Main execution
Write-Info "=== Authentication System Debug ==="
Write-Info "Base URL: $BaseUrl"
Write-Info "Test Email: $TestEmail"

# Test environment variables
Test-EnvironmentVariables

# Test health check
if (-not (Test-HealthCheck)) {
    Write-Error "Health check failed. Cannot continue."
    exit 1
}

# Test database state before signup
Write-Info "=== Database State Before Signup ==="
Test-DatabaseState

# Test signup
Write-Info "=== Testing Signup ==="
$signupSuccess = Test-Signup

# Test database state after signup
Write-Info "=== Database State After Signup ==="
Test-DatabaseState

# Test login if signup was successful
if ($signupSuccess) {
    Write-Info "=== Testing Login ==="
    $token = Test-Login

    if ($token) {
        Write-Success "Authentication flow completed successfully!"
        Write-Info "Access token: $($token.Substring(0, [Math]::Min(20, $token.Length)))..."
    } else {
        Write-Error "Login failed after successful signup"
    }
} else {
    Write-Error "Signup failed. Cannot test login."
}

Write-Info "=== Debug Complete ==="











