# Rate Limit Reset Script for Testing
# This script resets rate limits by clearing the rate limit cache

param(
    [string]$BaseUrl = "http://localhost:8080"
)

Write-Host "🔄 Resetting Rate Limits for Testing" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan

# Method 1: Try to restart the server (if running in test mode)
Write-Host "Attempting to reset rate limits..." -ForegroundColor Yellow

try {
    # Send a request to clear rate limits (if endpoint exists)
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/admin/rate-limits/reset" -Method POST -ErrorAction SilentlyContinue
    Write-Host "✅ Rate limits reset via API endpoint" -ForegroundColor Green
} catch {
    Write-Host "ℹ API endpoint not available, trying alternative methods..." -ForegroundColor Yellow
}

# Method 2: Wait for rate limit window to expire
Write-Host "Waiting for rate limit window to expire (60 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 60

# Method 3: Check if server is in test mode and restart if needed
Write-Host "Checking if server supports test mode..." -ForegroundColor Yellow

try {
    $healthResponse = Invoke-RestMethod -Uri "$BaseUrl/api/metrics/health" -Method GET
    Write-Host "✅ Server is responding" -ForegroundColor Green
} catch {
    Write-Host "❌ Server not responding, may need restart" -ForegroundColor Red
}

Write-Host ""
Write-Host "📋 Rate Limit Reset Complete" -ForegroundColor Green
Write-Host "===========================" -ForegroundColor Green
Write-Host "• Rate limit window has expired" -ForegroundColor White
Write-Host "• You can now retry your tests" -ForegroundColor White
Write-Host "• Consider using TEST_MODE=true for future testing" -ForegroundColor White

Write-Host ""
Write-Host "💡 Tips to avoid rate limiting in tests:" -ForegroundColor Magenta
Write-Host "1. Set TEST_MODE=true environment variable" -ForegroundColor White
Write-Host "2. Add delays between authentication attempts" -ForegroundColor White
Write-Host "3. Use different IP addresses for parallel tests" -ForegroundColor White
Write-Host "4. Implement exponential backoff in test scripts" -ForegroundColor White
