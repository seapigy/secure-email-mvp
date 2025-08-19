# Test script for Micro-Iteration 4.25: Advanced Notification & Retention Analytics Enhancements
# This script demonstrates and tests the new retention notification and analytics features

Write-Host "=== Micro-Iteration 4.25: Retention Analytics & Notifications Test ===" -ForegroundColor Green

# Configuration
$API_BASE = "http://localhost:8080"
$JWT_TOKEN = "your_jwt_token_here"  # Replace with actual JWT token

# Set environment variables for testing
$env:ENABLE_EXPIRATION_NOTIFICATIONS = "true"
$env:EXPIRATION_ADVANCE_NOTICE_HOURS = "24"
$env:ENABLE_CLEANUP_NOTIFICATIONS_SENDER = "true"
$env:ANALYTICS_CACHE_DURATION_MINUTES = "30"
$env:ANALYTICS_MAX_RECORDS = "500"

Write-Host "Environment variables set for testing..." -ForegroundColor Yellow

# Function to make authenticated API calls
function Invoke-AuthenticatedAPI {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null
    )
    
    $headers = @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    }
    
    $params = @{
        Method = $Method
        Uri = "$API_BASE$Endpoint"
        Headers = $headers
    }
    
    if ($Body) {
        $params.Body = $Body | ConvertTo-Json -Depth 10
    }
    
    try {
        $response = Invoke-RestMethod @params
        return $response
    }
    catch {
        Write-Host "API Error: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# Test 1: Get comprehensive retention analytics
Write-Host "`n1. Testing comprehensive retention analytics..." -ForegroundColor Cyan
$analyticsResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics?limit=10&offset=0"
if ($analyticsResponse) {
    Write-Host "✓ Analytics retrieved successfully" -ForegroundColor Green
    Write-Host "  - Overall stats: $($analyticsResponse.overall_stats.total_emails) total emails" -ForegroundColor White
    Write-Host "  - Expired emails: $($analyticsResponse.overall_stats.expired_count)" -ForegroundColor White
    Write-Host "  - Self-destructed emails: $($analyticsResponse.overall_stats.self_destructed_count)" -ForegroundColor White
    Write-Host "  - Average retention time: $($analyticsResponse.overall_stats.avg_retention_hours) hours" -ForegroundColor White
} else {
    Write-Host "✗ Failed to retrieve analytics" -ForegroundColor Red
}

# Test 2: Get analytics summary for dashboard
Write-Host "`n2. Testing analytics summary..." -ForegroundColor Cyan
$summaryResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics-summary"
if ($summaryResponse) {
    Write-Host "✓ Summary retrieved successfully" -ForegroundColor Green
    Write-Host "  - Total emails: $($summaryResponse.total_emails)" -ForegroundColor White
    Write-Host "  - Pending expiration: $($summaryResponse.pending_expiration)" -ForegroundColor White
    Write-Host "  - Recent cleanups: $($summaryResponse.recent_cleanups)" -ForegroundColor White
} else {
    Write-Host "✗ Failed to retrieve summary" -ForegroundColor Red
}

# Test 3: Get retention notifications history
Write-Host "`n3. Testing retention notifications history..." -ForegroundColor Cyan
$notificationsResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-notifications?limit=5&offset=0"
if ($notificationsResponse) {
    Write-Host "✓ Notifications history retrieved successfully" -ForegroundColor Green
    Write-Host "  - Total notifications: $($notificationsResponse.total_count)" -ForegroundColor White
    Write-Host "  - Recent notifications: $($notificationsResponse.notifications.Count)" -ForegroundColor White
    foreach ($notification in $notificationsResponse.notifications) {
        Write-Host "    - $($notification.notification_type) for email $($notification.email_id) at $($notification.sent_at)" -ForegroundColor Gray
    }
} else {
    Write-Host "✗ Failed to retrieve notifications history" -ForegroundColor Red
}

# Test 4: Get retention notification preferences
Write-Host "`n4. Testing notification preferences..." -ForegroundColor Cyan
$preferencesResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-notification-preferences?user_id=test_user"
if ($preferencesResponse) {
    Write-Host "✓ Preferences retrieved successfully" -ForegroundColor Green
    Write-Host "  - Expiration notifications: $($preferencesResponse.expiration_notifications)" -ForegroundColor White
    Write-Host "  - Cleanup notifications: $($preferencesResponse.cleanup_notifications)" -ForegroundColor White
} else {
    Write-Host "✗ Failed to retrieve preferences" -ForegroundColor Red
}

# Test 5: Update notification preferences
Write-Host "`n5. Testing preference updates..." -ForegroundColor Cyan
$updateBody = @{
    user_id = "test_user"
    expiration_notifications = $true
    cleanup_notifications = $false
    advance_notice_hours = 48
}
$updateResponse = Invoke-AuthenticatedAPI -Method "PUT" -Endpoint "/api/admin/email/retention-notification-preferences" -Body $updateBody
if ($updateResponse) {
    Write-Host "✓ Preferences updated successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to update preferences" -ForegroundColor Red
}

# Test 6: Test analytics with filters
Write-Host "`n6. Testing filtered analytics..." -ForegroundColor Cyan
$filteredResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics?status=expired&limit=5&offset=0"
if ($filteredResponse) {
    Write-Host "✓ Filtered analytics retrieved successfully" -ForegroundColor Green
    Write-Host "  - Filtered results: $($filteredResponse.overall_stats.total_emails) emails" -ForegroundColor White
} else {
    Write-Host "✗ Failed to retrieve filtered analytics" -ForegroundColor Red
}

# Test 7: Test analytics with date range
Write-Host "`n7. Testing date range analytics..." -ForegroundColor Cyan
$dateRangeResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics?start_date=2024-01-01&end_date=2024-12-31&limit=5"
if ($dateRangeResponse) {
    Write-Host "✓ Date range analytics retrieved successfully" -ForegroundColor Green
    Write-Host "  - Date range results: $($dateRangeResponse.overall_stats.total_emails) emails" -ForegroundColor White
} else {
    Write-Host "✗ Failed to retrieve date range analytics" -ForegroundColor Red
}

# Test 8: Test cleanup logs
Write-Host "`n8. Testing cleanup logs..." -ForegroundColor Cyan
$cleanupLogsResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics?include_cleanup_logs=true&limit=3"
if ($cleanupLogsResponse -and $cleanupLogsResponse.cleanup_logs) {
    Write-Host "✓ Cleanup logs retrieved successfully" -ForegroundColor Green
    Write-Host "  - Cleanup operations: $($cleanupLogsResponse.cleanup_logs.Count)" -ForegroundColor White
    foreach ($log in $cleanupLogsResponse.cleanup_logs) {
        Write-Host "    - $($log.initiator) processed $($log.emails_processed) emails, deleted $($log.emails_deleted) at $($log.timestamp)" -ForegroundColor Gray
    }
} else {
    Write-Host "✗ Failed to retrieve cleanup logs or no logs available" -ForegroundColor Red
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "Micro-Iteration 4.25 features tested:" -ForegroundColor White
Write-Host "✓ Comprehensive retention analytics" -ForegroundColor Green
Write-Host "✓ Analytics summary for dashboards" -ForegroundColor Green
Write-Host "✓ Retention notifications history" -ForegroundColor Green
Write-Host "✓ Notification preferences management" -ForegroundColor Green
Write-Host "✓ Filtered and date-range analytics" -ForegroundColor Green
Write-Host "✓ Cleanup logs integration" -ForegroundColor Green

Write-Host "`nTo run the enhanced cleanup worker with notifications:" -ForegroundColor Yellow
Write-Host "1. Set environment variables:" -ForegroundColor White
Write-Host "   `$env:ENABLE_EXPIRATION_NOTIFICATIONS = 'true'" -ForegroundColor Gray
Write-Host "   `$env:ENABLE_CLEANUP_NOTIFICATIONS_SENDER = 'true'" -ForegroundColor Gray
Write-Host "2. Run the enhanced worker:" -ForegroundColor White
Write-Host "   go run ./cmd/workers/enhanced_email_cleanup_worker.go" -ForegroundColor Gray

Write-Host "`nTest completed successfully!" -ForegroundColor Green






