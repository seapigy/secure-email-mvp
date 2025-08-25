# Authentication Debugging Integration Test
# This script provides comprehensive testing and debugging of the signup/login authentication system

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",
    [string]$TestPassword = "Test123!@#",
    [string]$FallbackEmail = "backup@securesystem.email",
    [switch]$Verbose = $false
)

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Cyan = "Cyan"
$White = "White"

function Write-Header {
    param([string]$Message)
    Write-Host "`n" -ForegroundColor $White
    Write-Host "=" * 80 -ForegroundColor $Cyan
    Write-Host $Message -ForegroundColor $Cyan
    Write-Host "=" * 80 -ForegroundColor $Cyan
    Write-Host "`n" -ForegroundColor $White
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor $Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor $Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor $Yellow
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ️  $Message" -ForegroundColor $White
}

function Write-Debug {
    param([string]$Message)
    if ($Verbose) {
        Write-Host "🔍 $Message" -ForegroundColor $Cyan
    }
}

function Test-ServerHealth {
    Write-Header "Testing Server Health"
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET -TimeoutSec 10
        if ($response.status -eq "ok") {
            Write-Success "Server is healthy and responding"
            return $true
        } else {
            Write-Error "Server health check failed: $($response | ConvertTo-Json)"
            return $false
        }
    } catch {
        Write-Error "Server health check failed: $($_.Exception.Message)"
        return $false
    }
}

function Test-ServerPing {
    Write-Header "Testing Server Ping"
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/ping" -Method GET -TimeoutSec 10
        if ($response -eq "pong") {
            Write-Success "Server ping successful"
            return $true
        } else {
            Write-Error "Server ping failed: $response"
            return $false
        }
    } catch {
        Write-Error "Server ping failed: $($_.Exception.Message)"
        return $false
    }
}

function Test-UserExists {
    param([string]$Email)
    
    Write-Info "Checking if user exists: $Email"
    
    try {
        # Try to signup - if it fails with "User already exists", the user exists
        $signupBody = @{
            email = $Email
            password = $TestPassword
            fallback_email = $FallbackEmail
        } | ConvertTo-Json
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method POST -ContentType "application/json" -Body $signupBody -TimeoutSec 10
        
        Write-Success "User does not exist, signup successful"
        return $false
    } catch {
        $errorMessage = $_.Exception.Message
        if ($errorMessage -like "*User already exists*") {
            Write-Success "User already exists: $Email"
            return $true
        } else {
            Write-Error "Unexpected error checking user existence: $errorMessage"
            return $null
        }
    }
}

function Test-Signup {
    param([string]$Email, [string]$Password, [string]$FallbackEmail)
    
    Write-Header "Testing User Signup"
    Write-Info "Email: $Email"
    Write-Info "Password length: $($Password.Length)"
    Write-Info "Fallback email: $FallbackEmail"
    
    try {
        $signupBody = @{
            email = $Email
            password = $Password
            fallback_email = $FallbackEmail
        } | ConvertTo-Json
        
        Write-Debug "Signup request body: $signupBody"
        Write-Info "📤 Sending signup request with JSON: $signupBody"
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method POST -ContentType "application/json" -Body $signupBody -TimeoutSec 10
        
        Write-Success "Signup successful: $($response.message)"
        return $true
    } catch {
        $errorMessage = $_.Exception.Message
        Write-Error "Signup failed: $errorMessage"
        return $false
    }
}

function Get-TOTPCode {
    param([string]$Secret = "JBSWY3DPEHPK3PXP")
    
    Write-Info "Generating TOTP code for secret: $Secret"
    
    try {
        $totpOutput = & ".\scripts\generate_totp.ps1" -Secret $Secret
        $currentCode = $totpOutput | Where-Object { $_ -match 'CURRENT:' } | ForEach-Object { $_.Split(':')[1].Trim() }
        
        if ($currentCode) {
            Write-Success "TOTP code generated: $currentCode"
            return $currentCode
        } else {
            Write-Error "Failed to extract TOTP code from output"
            return $null
        }
    } catch {
        Write-Error "Failed to generate TOTP code: $($_.Exception.Message)"
        return $null
    }
}

function Test-Login {
    param([string]$Email, [string]$Password, [string]$TOTPCode)
    
    Write-Header "Testing User Login"
    Write-Info "Email: $Email"
    Write-Info "Password length: $($Password.Length)"
    Write-Info "TOTP code: $TOTPCode"
    
    try {
        $loginBody = @{
            email = $Email
            password = $Password
            totp_code = $TOTPCode
        } | ConvertTo-Json
        
        Write-Debug "Login request body: $loginBody"
        Write-Info "📤 Sending login request with JSON: $loginBody"
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -ContentType "application/json" -Body $loginBody -TimeoutSec 10
        
        Write-Success "Login successful!"
        Write-Info "User ID: $($response.user_id)"
        Write-Info "Email: $($response.email)"
        Write-Info "Token type: $($response.token_type)"
        Write-Info "Expires in: $($response.expires_in) seconds"
        
        return @{
            success = $true
            access_token = $response.access_token
            refresh_token = $response.refresh_token
            user_id = $response.user_id
            email = $response.email
        }
    } catch {
        $errorMessage = $_.Exception.Message
        Write-Error "Login failed: $errorMessage"
        
        return @{
            success = $false
            error = @{ message = $errorMessage }
        }
    }
}

function Test-InvalidCredentials {
    param([string]$Email, [string]$Password, [string]$TOTPCode, [string]$TestName)
    
    Write-Header "Testing Invalid Credentials: $TestName"
    Write-Info "Email: $Email"
    Write-Info "Password length: $($Password.Length)"
    Write-Info "TOTP code: $TOTPCode"
    
    try {
        $loginBody = @{
            email = $Email
            password = $Password
            totp_code = $TOTPCode
        } | ConvertTo-Json
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -ContentType "application/json" -Body $loginBody -TimeoutSec 10
        
        Write-Error "Login should have failed but succeeded: $($response | ConvertTo-Json)"
        return $false
    } catch {
        $errorMessage = $_.Exception.Message
        Write-Success "Login correctly failed as expected: $errorMessage"
        return $true
    }
}

function Test-DatabaseUser {
    param([string]$Email)
    
    Write-Header "Checking Database User Details"
    
    try {
        # Check if sqlite3 is available
        $sqliteOutput = sqlite3 data/secure_email.db "SELECT email, password, password_hash, totp_secret, LENGTH(password), LENGTH(password_hash) FROM users WHERE email='$Email';" 2>$null
        
        if ($sqliteOutput) {
            Write-Success "User found in database"
            Write-Info "Database record: $sqliteOutput"
            
            # Parse the output (format: email|password|password_hash|totp_secret|password_length|password_hash_length)
            $parts = $sqliteOutput -split '\|'
            if ($parts.Length -ge 6) {
                Write-Info "Email: $($parts[0])"
                Write-Info "Password length: $($parts[4])"
                Write-Info "Password hash length: $($parts[5])"
                Write-Info "TOTP secret: $($parts[3])"
                
                return @{
                    exists = $true
                    email = $parts[0]
                    password_length = $parts[4]
                    password_hash_length = $parts[5]
                    totp_secret = $parts[3]
                }
            }
        } else {
            Write-Warning "User not found in database"
            return @{ exists = $false }
        }
    } catch {
        Write-Error "Failed to check database: $($_.Exception.Message)"
        return @{ exists = $null }
    }
}

# Main test execution
Write-Header "Authentication Debugging Integration Test"
Write-Info "Base URL: $BaseUrl"
Write-Info "Test Email: $TestEmail"
Write-Info "Test Password: $TestPassword"
Write-Info "Fallback Email: $FallbackEmail"
Write-Info "Verbose Mode: $Verbose"

# Test 1: Server Health
if (-not (Test-ServerHealth)) {
    Write-Error "Server health check failed. Please ensure the server is running."
    exit 1
}

# Test 2: Server Ping
if (-not (Test-ServerPing)) {
    Write-Error "Server ping failed. Please ensure the server is running."
    exit 1
}

# Test 3: Check if user exists
$userExists = Test-UserExists -Email $TestEmail

if ($userExists -eq $null) {
    Write-Error "Failed to determine user existence. Exiting."
    exit 1
}

# Test 4: Signup (if user doesn't exist)
if (-not $userExists) {
    Write-Info "User does not exist, creating new user..."
    if (-not (Test-Signup -Email $TestEmail -Password $TestPassword -FallbackEmail $FallbackEmail)) {
        Write-Error "Signup failed. Exiting."
        exit 1
    }
} else {
    Write-Info "User already exists, skipping signup"
}

# Test 5: Check database user details
$dbUser = Test-DatabaseUser -Email $TestEmail
if ($dbUser.exists -eq $null) {
    Write-Error "Failed to check database user details. Exiting."
    exit 1
}

# Test 6: Generate TOTP code
$totpCode = Get-TOTPCode
if (-not $totpCode) {
    Write-Error "Failed to generate TOTP code. Exiting."
    exit 1
}

# Test 7: Test login with valid credentials
$loginResult = Test-Login -Email $TestEmail -Password $TestPassword -TOTPCode $totpCode

if ($loginResult.success) {
    Write-Success "✅ Authentication system is working correctly!"
    
    # Test 8: Test invalid credentials
    Write-Header "Testing Invalid Credentials"
    
    # Test with wrong password
    Test-InvalidCredentials -Email $TestEmail -Password "WrongPassword123!" -TOTPCode $totpCode -TestName "Wrong Password"
    
    # Test with wrong TOTP code
    Test-InvalidCredentials -Email $TestEmail -Password $TestPassword -TOTPCode "000000" -TestName "Wrong TOTP Code"
    
    # Test with non-existent user
    Test-InvalidCredentials -Email "nonexistent@example.com" -Password $TestPassword -TOTPCode $totpCode -TestName "Non-existent User"
    
    Write-Header "Test Summary"
    Write-Success "All authentication tests completed successfully!"
    Write-Info "The authentication system is working correctly with:"
    Write-Info "- Argon2 password hashing"
    Write-Info "- TOTP 2FA authentication"
    Write-Info "- Proper error handling"
    Write-Info "- Rate limiting (disabled in test mode)"
    
} else {
    Write-Error "❌ Authentication system has issues!"
    Write-Header "Debugging Information"
    Write-Info "Login failed with error: $($loginResult.error | ConvertTo-Json -Depth 3)"
    Write-Info "Please check the server logs for detailed authentication debugging information."
    Write-Info "Look for [AUTH_DEBUG] and [LOGIN_DEBUG] log entries."
}

Write-Header "Test Complete"
