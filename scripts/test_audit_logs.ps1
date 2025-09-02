# =============================================================================
# SECURE EMAIL MVP - AUDIT LOG INTEGRATION TESTS
# =============================================================================
# PowerShell script to test the audit log system API endpoints.
# =============================================================================

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "TestPassword123!"
)

Write-Output "Starting Audit Log Integration Tests..."
Write-Output "Base URL: $BaseUrl"

# Test data
$testUser = @{
    email = $TestEmail
    password = $TestPassword
}

# Global variables
$global:authToken = $null
$global:userId = $null

# Helper function to make HTTP requests
function Invoke-ApiRequest {
    param(
        [string]$Method = "GET",
        [string]$Endpoint,
        [object]$Body = $null,
        [hashtable]$Headers = @{}
    )

    $uri = "$BaseUrl$Endpoint"
    $headers["Content-Type"] = "application/json"

    if ($global:authToken) {
        $headers["Authorization"] = "Bearer $global:authToken"
    }

    $params = @{
        Method = $Method
        Uri = $uri
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
        Write-Output "API Error: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Output "Error Body: $errorBody"
        }
        throw
    }
}

# Test 1: User Registration
Write-Output "`n1. Testing User Registration..."
try {
    $registerBody = @{
        email = $testUser.email
        password = $testUser.password
    }

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/register" -Body $registerBody
    Write-Output "✓ User registration successful"
}
catch {
    Write-Output "✗ User registration failed: $($_.Exception.Message)"
    # Continue if user already exists
}

# Test 2: User Login
Write-Output "`n2. Testing User Login..."
try {
    $loginBody = @{
        email = $testUser.email
        password = $testUser.password
        totp_code = "000000"  # Default TOTP for testing
    }

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginBody
    $global:authToken = $response.token
    $global:userId = $response.user_id
    Write-Output "✓ User login successful"
    Write-Output "  User ID: $global:userId"
}
catch {
    Write-Output "✗ User login failed: $($_.Exception.Message)"
    exit 1
}

# Test 3: Get Audit Event Types
Write-Output "`n3. Testing Get Audit Event Types..."
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/event-types"
    Write-Output "✓ Retrieved event types: $($response.event_types.Count) types"
    $response.event_types | ForEach-Object { Write-Output "  - $_" }
}
catch {
    Write-Output "✗ Get event types failed: $($_.Exception.Message)"
}

# Test 4: Get User Audit Events
Write-Output "`n4. Testing Get User Audit Events..."
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/user-events?limit=10"
    Write-Output "✓ Retrieved user events: $($response.total) events"
    if ($response.events.Count -gt 0) {
        Write-Output "  Latest event: $($response.events[0].event_type) - $($response.events[0].outcome)"
    }
}
catch {
    Write-Output "✗ Get user events failed: $($_.Exception.Message)"
}

# Test 5: Query Audit Logs
Write-Output "`n5. Testing Query Audit Logs..."
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/logs?page=1&page_size=5"
    Write-Output "✓ Retrieved audit logs: $($response.total) total, $($response.events.Count) on page"
    if ($response.events.Count -gt 0) {
        Write-Output "  Sample event: $($response.events[0].event_type) by $($response.events[0].user_id)"
    }
}
catch {
    Write-Output "✗ Query audit logs failed: $($_.Exception.Message)"
}

# Test 6: Get Retention Policies
Write-Output "`n6. Testing Get Retention Policies..."
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/retention-policies"
    Write-Output "✓ Retrieved retention policies: $($response.policies.Count) policies"
    $response.policies | ForEach-Object {
        Write-Output "  - $($_.event_type): $($_.retention_days) days (auto-purge: $($_.auto_purge))"
    }
}
catch {
    Write-Output "✗ Get retention policies failed: $($_.Exception.Message)"
}

# Test 7: Create Export Request
Write-Output "`n7. Testing Create Export Request..."
try {
    $exportBody = @{
        export_type = "json"
        filter = @{
            date_from = (Get-Date).AddDays(-7).ToString("yyyy-MM-ddTHH:mm:ssZ")
            date_to = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
            event_types = @("email_creation", "email_access", "login_attempt")
        }
    }

    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/audit/exports" -Body $exportBody
    $exportId = $response.export_id
    Write-Output "✓ Created export request: $exportId"
    Write-Output "  Status: $($response.status)"
}
catch {
    Write-Output "✗ Create export failed: $($_.Exception.Message)"
    $exportId = $null
}

# Test 8: Get Export Status
if ($exportId) {
    Write-Output "`n8. Testing Get Export Status..."
    try {
        Start-Sleep -Seconds 2  # Wait for processing
        $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/exports/$exportId"
        Write-Output "✓ Retrieved export status: $($response.status)"
        if ($response.status -eq "completed") {
            Write-Output "  File size: $($response.file_size) bytes"
        }
    }
    catch {
        Write-Output "✗ Get export status failed: $($_.Exception.Message)"
    }
}

# Test 9: Get User Exports
Write-Output "`n9. Testing Get User Exports..."
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/exports?limit=5"
    Write-Output "✓ Retrieved user exports: $($response.total) exports"
    if ($response.exports.Count -gt 0) {
        Write-Output "  Latest export: $($response.exports[0].export_type) - $($response.exports[0].status)"
    }
}
catch {
    Write-Output "✗ Get user exports failed: $($_.Exception.Message)"
}

# Test 10: Download Export (if available)
if ($exportId) {
    Write-Output "`n10. Testing Download Export..."
    try {
        $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/exports/$exportId"
        if ($response.status -eq "completed") {
            Write-Output "✓ Export is ready for download"
            Write-Output "  Download URL: $BaseUrl/api/audit/exports/$exportId/download"
        } else {
            Write-Output "⚠ Export not ready yet: $($response.status)"
        }
    }
    catch {
        Write-Output "✗ Check export status failed: $($_.Exception.Message)"
    }
}

# Test 11: Test Filtering
Write-Output "`n11. Testing Audit Log Filtering..."
try {
    # Test filtering by event type
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/logs?event_types=email_creation,login_attempt&page_size=3"
    Write-Output "✓ Filtered by event types: $($response.total) events"

    # Test filtering by outcome
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/logs?outcomes=success&page_size=3"
    Write-Output "✓ Filtered by outcome: $($response.total) events"

    # Test date range filtering
    $dateFrom = (Get-Date).AddDays(-1).ToString("yyyy-MM-ddTHH:mm:ssZ")
    $dateTo = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/logs?date_from=$dateFrom&date_to=$dateTo&page_size=3"
    Write-Output "✓ Filtered by date range: $($response.total) events"
}
catch {
    Write-Output "✗ Filtering tests failed: $($_.Exception.Message)"
}

# Test 12: Cleanup - Delete Export
if ($exportId) {
    Write-Output "`n12. Testing Delete Export..."
    try {
        Invoke-ApiRequest -Method "DELETE" -Endpoint "/api/audit/exports/$exportId"
        Write-Output "✓ Deleted export: $exportId"
    }
    catch {
        Write-Output "✗ Delete export failed: $($_.Exception.Message)"
    }
}

# Test 13: Admin Endpoints (if user has admin privileges)
Write-Output "`n13. Testing Admin Endpoints..."
try {
    # Purge expired logs
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/audit/purge-expired"
    Write-Output "✓ Purged expired logs"

    # Cleanup expired exports
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/audit/cleanup-exports"
    Write-Output "✓ Cleaned up expired exports"
}
catch {
    Write-Output "✗ Admin endpoints failed: $($_.Exception.Message)"
}

Write-Output "`nAudit Log Integration Tests Completed!"
Write-Output "`nNotes:"
Write-Output "- Audit logs are automatically created for login, email operations, etc."
Write-Output "- Export files are automatically cleaned up after 24 hours"
Write-Output "- Retention policies control how long different event types are kept"
Write-Output "- The audit worker should be running to handle cleanup tasks"






















