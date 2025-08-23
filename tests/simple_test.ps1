# Simple Test Script for Iteration 1 Fixes
# Tests basic functionality to ensure the system is working

Write-Output "Simple Test for Iteration 1 Fixes"
Write-Output "================================="

# Test configuration
$ApiUrl = "http://localhost:8080"
$DatabasePath = "secure-email.db"

# Test 1: Check if API server is running
Write-Output "Test 1: API Server Availability"
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/health" -Method GET -TimeoutSec 5
    Write-Output "API server is running"
} catch {
    Write-Output "API server is not running or not accessible"
    Write-Output "Error: $($_.Exception.Message)"
}

# Test 2: Check if database exists
Write-Output "Test 2: Database Existence"
if (Test-Path $DatabasePath) {
    Write-Output "Database exists at: $DatabasePath"
} else {
    Write-Output "Database not found at: $DatabasePath"
}

# Test 3: Check if Go server can be built
Write-Output "Test 3: Go Server Build"
try {
    $buildResult = go build -o test-server cmd/api/main.go 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Output "Go server builds successfully"
        Remove-Item "test-server.exe" -ErrorAction SilentlyContinue
    } else {
        Write-Output "Go server build failed"
        Write-Output "Error: $buildResult"
    }
} catch {
    Write-Output "Go build command failed"
    Write-Output "Error: $($_.Exception.Message)"
}

# Test 4: Check if React app can be built
Write-Output "Test 4: React App Build"
try {
    $npmResult = npm run build 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Output "React app builds successfully"
    } else {
        Write-Output "React app build failed"
        Write-Output "Error: $npmResult"
    }
} catch {
    Write-Output "npm build command failed"
    Write-Output "Error: $($_.Exception.Message)"
}

# Test 5: Check for key files from Iteration 1
Write-Output "Test 5: Iteration 1 Files"
$requiredFiles = @(
    "pkg/errors/errors.go",
    "pkg/auth/totp_drift_test.go",
    "src/components/admin/MobileOptimizedPanel.tsx",
    "tests/iteration1_integration_test.ps1",
    "docs/iteration1_implementation_summary.md"
)

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Output "$file exists"
    } else {
        Write-Output "$file missing"
    }
}

Write-Output "Simple Test Summary"
Write-Output "==================="
Write-Output "If all tests pass, the system is ready for full integration testing."
Write-Output "If any tests fail, please fix the issues before running the full test suite."
