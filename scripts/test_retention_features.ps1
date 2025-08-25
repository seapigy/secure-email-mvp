# Test Script for Micro-Iteration 4.24: Email Retention & Auto-Deletion Enhancements
# This script demonstrates the new retention features

Write-Output "=== Micro-Iteration 4.24: Email Retention & Auto-Deletion Test ==="
Write-Output ""

# Configuration
$API_BASE_URL = "http://localhost:8080"
$JWT_TOKEN = "your_jwt_token_here"  # Replace with actual JWT token

Write-Output "1. Testing Admin Retention Statistics Endpoint"
Write-Output "GET /api/admin/email/retention-stats"

try {
    $headers = @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    }

    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/retention-stats" -Method GET -Headers $headers
    Write-Output "Response:"
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Output "Error: $($_.Exception.Message)"
}

Write-Output ""
Write-Output "2. Testing Admin Retention Query Endpoint"
Write-Output "GET /api/admin/email/retention?limit=10"

try {
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/retention?limit=10" -Method GET -Headers $headers
    Write-Output "Response:"
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Output "Error: $($_.Exception.Message)"
}

Write-Output ""
Write-Output "3. Testing Admin Retention Query with Filters"
Write-Output "GET /api/admin/email/retention?status=expired&limit=5"

try {
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/retention?status=expired&limit=5" -Method GET -Headers $headers
    Write-Output "Response:"
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Output "Error: $($_.Exception.Message)"
}

Write-Output ""
Write-Output "4. Testing Manual Retention Cleanup (Dry Run)"
Write-Output "POST /api/admin/email/retention-cleanup"

try {
    $body = @{
        dry_run = $true
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/retention-cleanup" -Method POST -Headers $headers -Body $body
    Write-Output "Response:"
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Output "Error: $($_.Exception.Message)"
}

Write-Output ""
Write-Output "5. Testing Set Email Expiration"
Write-Output "POST /api/admin/email/expiration?email_id=test-email-1"

try {
    $expirationDate = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
    $body = @{
        expires_at = $expirationDate
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/admin/email/expiration?email_id=test-email-1" -Method POST -Headers $headers -Body $body
    Write-Output "Response:"
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Output "Error: $($_.Exception.Message)"
}

Write-Output ""
Write-Output "=== Environment Configuration ==="
Write-Output ""

Write-Output "Current Retention Configuration:"
Write-Output "DEFAULT_EMAIL_EXPIRATION_DAYS: $env:DEFAULT_EMAIL_EXPIRATION_DAYS"
Write-Output "CLEANUP_AUDIT_LOGS: $env:CLEANUP_AUDIT_LOGS"
Write-Output "ENABLE_CLEANUP_NOTIFICATIONS: $env:ENABLE_CLEANUP_NOTIFICATIONS"
Write-Output "CLEANUP_BATCH_SIZE: $env:CLEANUP_BATCH_SIZE"
Write-Output "EMAIL_CLEANUP_INTERVAL_MINUTES: $env:EMAIL_CLEANUP_INTERVAL_MINUTES"

Write-Output ""
Write-Output "=== Usage Instructions ==="
Write-Output ""

Write-Output "To use this script:"
Write-Output "1. Replace 'your_jwt_token_here' with an actual JWT token"
Write-Output "2. Ensure the API server is running on localhost:8080"
Write-Output "3. Set appropriate environment variables for retention configuration"
Write-Output "4. Run the script to test all retention endpoints"

Write-Output ""
Write-Output "=== Environment Variables to Set ==="
Write-Output ""

Write-Output "# Retention Configuration"
Write-Host '$env:DEFAULT_EMAIL_EXPIRATION_DAYS = "30"' -ForegroundColor Cyan
Write-Host '$env:CLEANUP_AUDIT_LOGS = "false"' -ForegroundColor Cyan
Write-Host '$env:ENABLE_CLEANUP_NOTIFICATIONS = "true"' -ForegroundColor Cyan
Write-Host '$env:CLEANUP_BATCH_SIZE = "100"' -ForegroundColor Cyan
Write-Host '$env:EMAIL_CLEANUP_INTERVAL_MINUTES = "15"' -ForegroundColor Cyan

Write-Output ""
Write-Output "=== Enhanced Cleanup Worker ==="
Write-Output ""

Write-Output "To start the enhanced cleanup worker:"
Write-Output "go run cmd/workers/enhanced_email_cleanup_worker.go"

Write-Output ""
Write-Output "=== Test Complete ==="









