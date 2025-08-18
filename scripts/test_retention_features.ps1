# Test Script for Micro-Iteration 4.24: Email Retention & Auto-Deletion Enhancements
# This script demonstrates the new retention features

Write-Host "=== Micro-Iteration 4.24: Email Retention & Auto-Deletion Test ===" -ForegroundColor Green
Write-Host ""

# Configuration
$API_BASE_URL = "http://localhost:8080"
$JWT_TOKEN = "your_jwt_token_here"  # Replace with actual JWT token

Write-Host "1. Testing Admin Retention Statistics Endpoint" -ForegroundColor Yellow
Write-Host "GET /api/admin/email/retention-stats" -ForegroundColor Cyan

try {
    $headers = @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    }
    
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/retention-stats" -Method GET -Headers $headers
    Write-Host "Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "2. Testing Admin Retention Query Endpoint" -ForegroundColor Yellow
Write-Host "GET /api/admin/email/retention?limit=10" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/retention?limit=10" -Method GET -Headers $headers
    Write-Host "Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "3. Testing Admin Retention Query with Filters" -ForegroundColor Yellow
Write-Host "GET /api/admin/email/retention?status=expired&limit=5" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/retention?status=expired&limit=5" -Method GET -Headers $headers
    Write-Host "Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "4. Testing Manual Retention Cleanup (Dry Run)" -ForegroundColor Yellow
Write-Host "POST /api/admin/email/retention-cleanup" -ForegroundColor Cyan

try {
    $body = @{
        dry_run = $true
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/retention-cleanup" -Method POST -Headers $headers -Body $body
    Write-Host "Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "5. Testing Set Email Expiration" -ForegroundColor Yellow
Write-Host "POST /api/admin/email/expiration?email_id=test-email-1" -ForegroundColor Cyan

try {
    $expirationDate = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
    $body = @{
        expires_at = $expirationDate
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/expiration?email_id=test-email-1" -Method POST -Headers $headers -Body $body
    Write-Host "Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Environment Configuration ===" -ForegroundColor Green
Write-Host ""

Write-Host "Current Retention Configuration:" -ForegroundColor Yellow
Write-Host "DEFAULT_EMAIL_EXPIRATION_DAYS: $env:DEFAULT_EMAIL_EXPIRATION_DAYS" -ForegroundColor Cyan
Write-Host "CLEANUP_AUDIT_LOGS: $env:CLEANUP_AUDIT_LOGS" -ForegroundColor Cyan
Write-Host "ENABLE_CLEANUP_NOTIFICATIONS: $env:ENABLE_CLEANUP_NOTIFICATIONS" -ForegroundColor Cyan
Write-Host "CLEANUP_BATCH_SIZE: $env:CLEANUP_BATCH_SIZE" -ForegroundColor Cyan
Write-Host "EMAIL_CLEANUP_INTERVAL_MINUTES: $env:EMAIL_CLEANUP_INTERVAL_MINUTES" -ForegroundColor Cyan

Write-Host ""
Write-Host "=== Usage Instructions ===" -ForegroundColor Green
Write-Host ""

Write-Host "To use this script:" -ForegroundColor Yellow
Write-Host "1. Replace 'your_jwt_token_here' with an actual JWT token" -ForegroundColor White
Write-Host "2. Ensure the API server is running on localhost:8080" -ForegroundColor White
Write-Host "3. Set appropriate environment variables for retention configuration" -ForegroundColor White
Write-Host "4. Run the script to test all retention endpoints" -ForegroundColor White

Write-Host ""
Write-Host "=== Environment Variables to Set ===" -ForegroundColor Green
Write-Host ""

Write-Host "# Retention Configuration" -ForegroundColor Yellow
Write-Host '$env:DEFAULT_EMAIL_EXPIRATION_DAYS = "30"' -ForegroundColor Cyan
Write-Host '$env:CLEANUP_AUDIT_LOGS = "false"' -ForegroundColor Cyan
Write-Host '$env:ENABLE_CLEANUP_NOTIFICATIONS = "true"' -ForegroundColor Cyan
Write-Host '$env:CLEANUP_BATCH_SIZE = "100"' -ForegroundColor Cyan
Write-Host '$env:EMAIL_CLEANUP_INTERVAL_MINUTES = "15"' -ForegroundColor Cyan

Write-Host ""
Write-Host "=== Enhanced Cleanup Worker ===" -ForegroundColor Green
Write-Host ""

Write-Host "To start the enhanced cleanup worker:" -ForegroundColor Yellow
Write-Host "go run cmd/workers/enhanced_email_cleanup_worker.go" -ForegroundColor Cyan

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Green




