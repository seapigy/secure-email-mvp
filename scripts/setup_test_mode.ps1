# Test Mode Setup Script
# This script configures the environment for testing with relaxed rate limits

Write-Host "🧪 Setting Up Test Mode Environment" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan

# Set test mode environment variables
$env:TEST_MODE = "true"
$env:LOGIN_RATE_LIMIT_ENABLED = "0"  # Disable login rate limiting for testing
$env:RATE_LIMIT_REQUESTS = "100"     # Increase rate limit for testing
$env:RATE_LIMIT_WINDOW = "60"        # 60 second window
$env:DEBUG = "true"                  # Enable debug mode

Write-Host "✅ Environment variables set:" -ForegroundColor Green
Write-Host "  • TEST_MODE = $env:TEST_MODE" -ForegroundColor White
Write-Host "  • LOGIN_RATE_LIMIT_ENABLED = $env:LOGIN_RATE_LIMIT_ENABLED" -ForegroundColor White
Write-Host "  • RATE_LIMIT_REQUESTS = $env:RATE_LIMIT_REQUESTS" -ForegroundColor White
Write-Host "  • RATE_LIMIT_WINDOW = $env:RATE_LIMIT_WINDOW" -ForegroundColor White
Write-Host "  • DEBUG = $env:DEBUG" -ForegroundColor White

Write-Host ""
Write-Host "🔄 Restarting server with test mode configuration..." -ForegroundColor Yellow

# Stop existing server
try {
    Stop-Process -Name "api_server" -Force -ErrorAction SilentlyContinue
    Write-Host "✅ Stopped existing server" -ForegroundColor Green
} catch {
    Write-Host "ℹ No existing server to stop" -ForegroundColor Yellow
}

# Wait a moment
Start-Sleep -Seconds 2

# Start server with test mode
Write-Host "🚀 Starting server in test mode..." -ForegroundColor Yellow
Start-Process -FilePath ".\api_server.exe" -WindowStyle Minimized

# Wait for server to start
Write-Host "⏳ Waiting for server to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Test server health
try {
    $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/metrics/health" -Method GET
    Write-Host "✅ Server is running and healthy" -ForegroundColor Green
} catch {
    Write-Host "❌ Server health check failed" -ForegroundColor Red
    Write-Host "   You may need to start the server manually" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📋 Test Mode Setup Complete" -ForegroundColor Green
Write-Host "==========================" -ForegroundColor Green
Write-Host "• Rate limiting is relaxed for testing" -ForegroundColor White
Write-Host "• Debug mode is enabled" -ForegroundColor White
Write-Host "• You can now run tests without rate limiting issues" -ForegroundColor White

Write-Host ""
Write-Host "🧪 Next Steps:" -ForegroundColor Magenta
Write-Host "1. Run your test scripts" -ForegroundColor White
Write-Host "2. Monitor server logs for any issues" -ForegroundColor White
Write-Host "3. Use .\scripts\reset_rate_limits.ps1 if needed" -ForegroundColor White

Write-Host ""
Write-Host "⚠️  IMPORTANT: This configuration is for testing only!" -ForegroundColor Red
Write-Host "   Do not use these settings in production!" -ForegroundColor Red
