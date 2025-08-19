# Test Password Validation Bypass
# Temporarily disable HIBP API to isolate the issue

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

# Temporarily clear HIBP API key
$originalHIBPKey = $env:HIBP_API_KEY
$env:HIBP_API_KEY = ""

Write-Info "Temporarily disabled HIBP API key"

# Test with a very strong password
$signupData = @{
    email = "test@securesystem.email"
    password = "SuperStrongPassword123!@#$%"
    fallback_email = "fallback@example.com"
}

$jsonBody = $signupData | ConvertTo-Json
Write-Info "Request body: $jsonBody"

try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method POST -Headers @{"Content-Type"="application/json"} -Body $jsonBody
    Write-Success "Signup successful: $($response | ConvertTo-Json)"
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
}

# Restore original HIBP API key
$env:HIBP_API_KEY = $originalHIBPKey
Write-Info "Restored HIBP API key"









