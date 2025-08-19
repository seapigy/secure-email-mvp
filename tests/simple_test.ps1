# Simple Test Script for Iteration 1 Fixes
# Tests basic functionality to ensure the system is working

Write-Host "Simple Test for Iteration 1 Fixes" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

# Test configuration
$ApiUrl = "http://localhost:8080"
$DatabasePath = "secure-email.db"

# Test 1: Check if API server is running
Write-Host "Test 1: API Server Availability" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/health" -Method GET -TimeoutSec 5
    Write-Host "API server is running" -ForegroundColor Green
} catch {
    Write-Host "API server is not running or not accessible" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Gray
}

# Test 2: Check if database exists
Write-Host "Test 2: Database Existence" -ForegroundColor Yellow
if (Test-Path $DatabasePath) {
    Write-Host "Database exists at: $DatabasePath" -ForegroundColor Green
} else {
    Write-Host "Database not found at: $DatabasePath" -ForegroundColor Red
}

# Test 3: Check if Go server can be built
Write-Host "Test 3: Go Server Build" -ForegroundColor Yellow
try {
    $buildResult = go build -o test-server cmd/api/main.go 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Go server builds successfully" -ForegroundColor Green
        Remove-Item "test-server.exe" -ErrorAction SilentlyContinue
    } else {
        Write-Host "Go server build failed" -ForegroundColor Red
        Write-Host "Error: $buildResult" -ForegroundColor Gray
    }
} catch {
    Write-Host "Go build command failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Gray
}

# Test 4: Check if React app can be built
Write-Host "Test 4: React App Build" -ForegroundColor Yellow
try {
    $npmResult = npm run build 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "React app builds successfully" -ForegroundColor Green
    } else {
        Write-Host "React app build failed" -ForegroundColor Red
        Write-Host "Error: $npmResult" -ForegroundColor Gray
    }
} catch {
    Write-Host "npm build command failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Gray
}

# Test 5: Check for key files from Iteration 1
Write-Host "Test 5: Iteration 1 Files" -ForegroundColor Yellow
$requiredFiles = @(
    "pkg/errors/errors.go",
    "pkg/auth/totp_drift_test.go",
    "src/components/admin/MobileOptimizedPanel.tsx",
    "tests/iteration1_integration_test.ps1",
    "docs/iteration1_implementation_summary.md"
)

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "$file exists" -ForegroundColor Green
    } else {
        Write-Host "$file missing" -ForegroundColor Red
    }
}

Write-Host "Simple Test Summary" -ForegroundColor Cyan
Write-Host "===================" -ForegroundColor Cyan
Write-Host "If all tests pass, the system is ready for full integration testing." -ForegroundColor White
Write-Host "If any tests fail, please fix the issues before running the full test suite." -ForegroundColor Yellow
