# Detailed Signup Debug Script
# Captures exact error responses

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

function Test-SignupDetailed {
    Write-Info "Testing user signup with detailed error capture..."
    
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
        Write-Error "Status Code: $($_.Exception.Response.StatusCode)"
        
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Error "Response body: '$errorBody'"
            
            # Try to parse as JSON
            try {
                $errorJson = $errorBody | ConvertFrom-Json
                Write-Error "Parsed error: $($errorJson | ConvertTo-Json)"
            } catch {
                Write-Error "Could not parse error as JSON"
            }
        }
        return $false
    }
}

# Test different email formats
function Test-EmailFormats {
    Write-Info "=== Testing Different Email Formats ==="
    
    $testEmails = @(
        "test@securesystem.email",
        "test@example.com",
        "test@test.com"
    )
    
    foreach ($email in $testEmails) {
        Write-Info "Testing email: $email"
        $signupData = @{
            email = $email
            password = $TestPassword
            fallback_email = $FallbackEmail
        }
        
        $jsonBody = $signupData | ConvertTo-Json
        
        try {
            $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method POST -Headers @{"Content-Type"="application/json"} -Body $jsonBody
            Write-Success "Signup successful for $email"
            break
        } catch {
            Write-Error "Signup failed for $email`: $($_.Exception.Response.StatusCode)"
            if ($_.Exception.Response) {
                $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
                $errorBody = $reader.ReadToEnd()
                Write-Error "Error: '$errorBody'"
            }
        }
    }
}

# Test different password formats
function Test-PasswordFormats {
    Write-Info "=== Testing Different Password Formats ==="
    
    $testPasswords = @(
        "TestPassword123!",
        "Password123!",
        "Test123!",
        "TestPassword123"
    )
    
    foreach ($password in $testPasswords) {
        Write-Info "Testing password: $password"
        $signupData = @{
            email = "test2@securesystem.email"
            password = $password
            fallback_email = $FallbackEmail
        }
        
        $jsonBody = $signupData | ConvertTo-Json
        
        try {
            $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method POST -Headers @{"Content-Type"="application/json"} -Body $jsonBody
            Write-Success "Signup successful with password: $password"
            break
        } catch {
            Write-Error "Signup failed with password $password`: $($_.Exception.Response.StatusCode)"
            if ($_.Exception.Response) {
                $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
                $errorBody = $reader.ReadToEnd()
                Write-Error "Error: '$errorBody'"
            }
        }
    }
}

# Main execution
Write-Info "=== Detailed Signup Debug ==="

# Test basic signup
Test-SignupDetailed

# Test different email formats
Test-EmailFormats

# Test different password formats
Test-PasswordFormats

Write-Info "=== Debug Complete ==="
