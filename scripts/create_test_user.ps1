# Create Test User Script
# Creates a test user with known credentials for testing

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",
    [string]$TestPassword = "TestPassword123!",
    [string]$TestTOTPSecret = "JBSWY3DPEHPK3PXP"
)

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Invoke-ApiRequest {
    param(
        [string]$Method = "GET",
        [string]$Endpoint,
        [object]$Body = $null
    )

    $headers = @{
        "Content-Type" = "application/json"
    }

    $uri = "$BaseUrl$Endpoint"

    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -Body $jsonBody
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }
        return @{ Success = $true; Data = $response }
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            try {
                $errorJson = $errorBody | ConvertFrom-Json
                return @{ Success = $false; Error = $errorJson.error; StatusCode = $errorResponse.StatusCode }
            }
            catch {
                return @{ Success = $false; Error = $errorBody; StatusCode = $errorResponse.StatusCode }
            }
        }
        return @{ Success = $false; Error = $_.Exception.Message }
    }
}

Write-Info "Creating test user with known TOTP secret..."

# Test health check first
$healthResult = Invoke-ApiRequest -Method "GET" -Endpoint "/health"
if (-not $healthResult.Success) {
    Write-Error "Health check failed. Make sure the API server is running."
    exit 1
}

Write-Success "API server is running"

# Create test user
$signupData = @{
    email = $TestEmail
    password = $TestPassword
}

$result = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $signupData

if ($result.Success) {
    Write-Success "Test user created successfully"
    Write-Info "Email: $TestEmail"
    Write-Info "Password: $TestPassword"
    Write-Info "TOTP Secret: $TestTOTPSecret"
    Write-Info "TOTP Code: 123456 (for testing)"
} else {
    Write-Error "Failed to create test user: $($result.Error)"
    exit 1
}











