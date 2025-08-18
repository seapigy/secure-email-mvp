# Test Password Validation
# Check if password validation is causing the signup failure

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

# Test with different passwords
$testCases = @(
    @{email = "test1@example.com"; password = "TestPassword123!"; description = "Strong password"},
    @{email = "test2@example.com"; password = "Password123!"; description = "Medium password"},
    @{email = "test3@example.com"; password = "Test123!"; description = "Short password"},
    @{email = "test4@example.com"; password = "password"; description = "Weak password"},
    @{email = "test5@example.com"; password = "123456"; description = "Very weak password"}
)

foreach ($testCase in $testCases) {
    Write-Info "Testing: $($testCase.description)"
    Write-Info "Email: $($testCase.email)"
    Write-Info "Password: $($testCase.password)"
    
    $signupData = @{
        email = $testCase.email
        password = $testCase.password
        fallback_email = "fallback@example.com"
    }
    
    $jsonBody = $signupData | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method POST -Headers @{"Content-Type"="application/json"} -Body $jsonBody
        Write-Success "Signup successful for $($testCase.description)"
    } catch {
        Write-Error "Signup failed for $($testCase.description): $($_.Exception.Response.StatusCode)"
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Error "Error: '$errorBody'"
        }
    }
    
    Write-Host ""
}







