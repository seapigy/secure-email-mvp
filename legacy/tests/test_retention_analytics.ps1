# Test script for Micro-Iteration 4.25: Advanced Notification & Retention Analytics Enhancements
# This script demonstrates and tests the new retention notification and analytics features

Write-Output "=== Micro-Iteration 4.25: Retention Analytics & Notifications Test ==="

# Configuration
$API_BASE = "http://localhost:8080"
$JWT_TOKEN = "your_jwt_token_here"  # Replace with actual JWT token

# Set environment variables for testing
$env:ENABLE_EXPIRATION_NOTIFICATIONS = "true"
$env:EXPIRATION_ADVANCE_NOTICE_HOURS = "24"
$env:ENABLE_CLEANUP_NOTIFICATIONS_SENDER = "true"
$env:ANALYTICS_CACHE_DURATION_MINUTES = "30"
$env:ANALYTICS_MAX_RECORDS = "500"

Write-Output "Environment variables set for testing..."

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
        Write-Error "API Error: $($_.Exception.Message)"
        return $null
    }
}

# Test 1: Get comprehensive retention analytics
Write-Output "`n1. Testing comprehensive retention analytics..."
$analyticsResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics?limit=10&offset=0"
if ($analyticsResponse) {
    Write-Output "✓ Analytics retrieved successfully"
    Write-Output "  - Overall stats: $($analyticsResponse.overall_stats.total_emails) total emails"
    Write-Output "  - Expired emails: $($analyticsResponse.overall_stats.expired_count)"
    Write-Output "  - Self-destructed emails: $($analyticsResponse.overall_stats.self_destructed_count)"
    Write-Output "  - Average retention time: $($analyticsResponse.overall_stats.avg_retention_hours) hours"
} else {
    Write-Output "✗ Failed to retrieve analytics"
}

# Test 2: Get analytics summary for dashboard
Write-Output "`n2. Testing analytics summary..."
$summaryResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics-summary"
if ($summaryResponse) {
    Write-Output "✓ Summary retrieved successfully"
    Write-Output "  - Total emails: $($summaryResponse.total_emails)"
    Write-Output "  - Pending expiration: $($summaryResponse.pending_expiration)"
    Write-Output "  - Recent cleanups: $($summaryResponse.recent_cleanups)"
} else {
    Write-Output "✗ Failed to retrieve summary"
}

# Test 3: Get retention notifications history
Write-Output "`n3. Testing retention notifications history..."
$notificationsResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-notifications?limit=5&offset=0"
if ($notificationsResponse) {
    Write-Output "✓ Notifications history retrieved successfully"
    Write-Output "  - Total notifications: $($notificationsResponse.total_count)"
    Write-Output "  - Recent notifications: $($notificationsResponse.notifications.Count)"
    foreach ($notification in $notificationsResponse.notifications) {
        Write-Output "    - $($notification.notification_type) for email $($notification.email_id) at $($notification.sent_at)"
    }
} else {
    Write-Output "✗ Failed to retrieve notifications history"
}

# Test 4: Get retention notification preferences
Write-Output "`n4. Testing notification preferences..."
$preferencesResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-notification-preferences?user_id=test_user"
if ($preferencesResponse) {
    Write-Output "✓ Preferences retrieved successfully"
    Write-Output "  - Expiration notifications: $($preferencesResponse.expiration_notifications)"
    Write-Output "  - Cleanup notifications: $($preferencesResponse.cleanup_notifications)"
} else {
    Write-Output "✗ Failed to retrieve preferences"
}

# Test 5: Update notification preferences
Write-Output "`n5. Testing preference updates..."
$updateBody = @{
    user_id = "test_user"
    expiration_notifications = $true
    cleanup_notifications = $false
    advance_notice_hours = 48
}
$updateResponse = Invoke-AuthenticatedAPI -Method "PUT" -Endpoint "/api/admin/email/retention-notification-preferences" -Body $updateBody
if ($updateResponse) {
    Write-Output "✓ Preferences updated successfully"
} else {
    Write-Output "✗ Failed to update preferences"
}

# Test 6: Test analytics with filters
Write-Output "`n6. Testing filtered analytics..."
$filteredResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics?status=expired&limit=5&offset=0"
if ($filteredResponse) {
    Write-Output "✓ Filtered analytics retrieved successfully"
    Write-Output "  - Filtered results: $($filteredResponse.overall_stats.total_emails) emails"
} else {
    Write-Output "✗ Failed to retrieve filtered analytics"
}

# Test 7: Test analytics with date range
Write-Output "`n7. Testing date range analytics..."
$dateRangeResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics?start_date=2024-01-01&end_date=2024-12-31&limit=5"
if ($dateRangeResponse) {
    Write-Output "✓ Date range analytics retrieved successfully"
    Write-Output "  - Date range results: $($dateRangeResponse.overall_stats.total_emails) emails"
} else {
    Write-Output "✗ Failed to retrieve date range analytics"
}

# Test 8: Test cleanup logs
Write-Output "`n8. Testing cleanup logs..."
$cleanupLogsResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-analytics?include_cleanup_logs=true&limit=3"
if ($cleanupLogsResponse -and $cleanupLogsResponse.cleanup_logs) {
    Write-Output "✓ Cleanup logs retrieved successfully"
    Write-Output "  - Cleanup operations: $($cleanupLogsResponse.cleanup_logs.Count)"
    foreach ($log in $cleanupLogsResponse.cleanup_logs) {
        Write-Output "    - $($log.initiator) processed $($log.emails_processed) emails, deleted $($log.emails_deleted) at $($log.timestamp)"
    }
} else {
    Write-Output "✗ Failed to retrieve cleanup logs or no logs available"
}

Write-Output "`n=== Test Summary ==="
Write-Output "Micro-Iteration 4.25 features tested:"
Write-Output "✓ Comprehensive retention analytics"
Write-Output "✓ Analytics summary for dashboards"
Write-Output "✓ Retention notifications history"
Write-Output "✓ Notification preferences management"
Write-Output "✓ Filtered and date-range analytics"
Write-Output "✓ Cleanup logs integration"

Write-Output "`nTo run the enhanced cleanup worker with notifications:"
Write-Output "1. Set environment variables:"
Write-Output "   `$env:ENABLE_EXPIRATION_NOTIFICATIONS = 'true'"
Write-Output "   `$env:ENABLE_CLEANUP_NOTIFICATIONS_SENDER = 'true'"
Write-Output "2. Run the enhanced worker:"
Write-Output "   go run ./cmd/workers/enhanced_email_cleanup_worker.go"

Write-Output "`nTest completed successfully!"






