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

Write-Host "Starting Audit Log Integration Tests..." -ForegroundColor Green
Write-Host "Base URL: $BaseUrl" -ForegroundColor Yellow

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
        Write-Host "API Error: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Body: $errorBody" -ForegroundColor Red
        }
        throw
    }
}

# Test 1: User Registration
Write-Host "`n1. Testing User Registration..." -ForegroundColor Cyan
try {
    $registerBody = @{
        email = $testUser.email
        password = $testUser.password
    }
    
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/register" -Body $registerBody
    Write-Host "✓ User registration successful" -ForegroundColor Green
}
catch {
    Write-Host "✗ User registration failed: $($_.Exception.Message)" -ForegroundColor Red
    # Continue if user already exists
}

# Test 2: User Login
Write-Host "`n2. Testing User Login..." -ForegroundColor Cyan
try {
    $loginBody = @{
        email = $testUser.email
        password = $testUser.password
        totp_code = "000000"  # Default TOTP for testing
    }
    
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginBody
    $global:authToken = $response.token
    $global:userId = $response.user_id
    Write-Host "✓ User login successful" -ForegroundColor Green
    Write-Host "  User ID: $global:userId" -ForegroundColor Yellow
}
catch {
    Write-Host "✗ User login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 3: Get Audit Event Types
Write-Host "`n3. Testing Get Audit Event Types..." -ForegroundColor Cyan
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/event-types"
    Write-Host "✓ Retrieved event types: $($response.event_types.Count) types" -ForegroundColor Green
    $response.event_types | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
}
catch {
    Write-Host "✗ Get event types failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Get User Audit Events
Write-Host "`n4. Testing Get User Audit Events..." -ForegroundColor Cyan
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/user-events?limit=10"
    Write-Host "✓ Retrieved user events: $($response.total) events" -ForegroundColor Green
    if ($response.events.Count -gt 0) {
        Write-Host "  Latest event: $($response.events[0].event_type) - $($response.events[0].outcome)" -ForegroundColor Gray
    }
}
catch {
    Write-Host "✗ Get user events failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: Query Audit Logs
Write-Host "`n5. Testing Query Audit Logs..." -ForegroundColor Cyan
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/logs?page=1&page_size=5"
    Write-Host "✓ Retrieved audit logs: $($response.total) total, $($response.events.Count) on page" -ForegroundColor Green
    if ($response.events.Count -gt 0) {
        Write-Host "  Sample event: $($response.events[0].event_type) by $($response.events[0].user_id)" -ForegroundColor Gray
    }
}
catch {
    Write-Host "✗ Query audit logs failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 6: Get Retention Policies
Write-Host "`n6. Testing Get Retention Policies..." -ForegroundColor Cyan
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/retention-policies"
    Write-Host "✓ Retrieved retention policies: $($response.policies.Count) policies" -ForegroundColor Green
    $response.policies | ForEach-Object { 
        Write-Host "  - $($_.event_type): $($_.retention_days) days (auto-purge: $($_.auto_purge))" -ForegroundColor Gray 
    }
}
catch {
    Write-Host "✗ Get retention policies failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 7: Create Export Request
Write-Host "`n7. Testing Create Export Request..." -ForegroundColor Cyan
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
    Write-Host "✓ Created export request: $exportId" -ForegroundColor Green
    Write-Host "  Status: $($response.status)" -ForegroundColor Yellow
}
catch {
    Write-Host "✗ Create export failed: $($_.Exception.Message)" -ForegroundColor Red
    $exportId = $null
}

# Test 8: Get Export Status
if ($exportId) {
    Write-Host "`n8. Testing Get Export Status..." -ForegroundColor Cyan
    try {
        Start-Sleep -Seconds 2  # Wait for processing
        $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/exports/$exportId"
        Write-Host "✓ Retrieved export status: $($response.status)" -ForegroundColor Green
        if ($response.status -eq "completed") {
            Write-Host "  File size: $($response.file_size) bytes" -ForegroundColor Yellow
        }
    }
    catch {
        Write-Host "✗ Get export status failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 9: Get User Exports
Write-Host "`n9. Testing Get User Exports..." -ForegroundColor Cyan
try {
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/exports?limit=5"
    Write-Host "✓ Retrieved user exports: $($response.total) exports" -ForegroundColor Green
    if ($response.exports.Count -gt 0) {
        Write-Host "  Latest export: $($response.exports[0].export_type) - $($response.exports[0].status)" -ForegroundColor Gray
    }
}
catch {
    Write-Host "✗ Get user exports failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 10: Download Export (if available)
if ($exportId) {
    Write-Host "`n10. Testing Download Export..." -ForegroundColor Cyan
    try {
        $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/exports/$exportId"
        if ($response.status -eq "completed") {
            Write-Host "✓ Export is ready for download" -ForegroundColor Green
            Write-Host "  Download URL: $BaseUrl/api/audit/exports/$exportId/download" -ForegroundColor Yellow
        } else {
            Write-Host "⚠ Export not ready yet: $($response.status)" -ForegroundColor Yellow
        }
    }
    catch {
        Write-Host "✗ Check export status failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 11: Test Filtering
Write-Host "`n11. Testing Audit Log Filtering..." -ForegroundColor Cyan
try {
    # Test filtering by event type
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/logs?event_types=email_creation,login_attempt&page_size=3"
    Write-Host "✓ Filtered by event types: $($response.total) events" -ForegroundColor Green
    
    # Test filtering by outcome
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/logs?outcomes=success&page_size=3"
    Write-Host "✓ Filtered by outcome: $($response.total) events" -ForegroundColor Green
    
    # Test date range filtering
    $dateFrom = (Get-Date).AddDays(-1).ToString("yyyy-MM-ddTHH:mm:ssZ")
    $dateTo = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/audit/logs?date_from=$dateFrom&date_to=$dateTo&page_size=3"
    Write-Host "✓ Filtered by date range: $($response.total) events" -ForegroundColor Green
}
catch {
    Write-Host "✗ Filtering tests failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 12: Cleanup - Delete Export
if ($exportId) {
    Write-Host "`n12. Testing Delete Export..." -ForegroundColor Cyan
    try {
        Invoke-ApiRequest -Method "DELETE" -Endpoint "/api/audit/exports/$exportId"
        Write-Host "✓ Deleted export: $exportId" -ForegroundColor Green
    }
    catch {
        Write-Host "✗ Delete export failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 13: Admin Endpoints (if user has admin privileges)
Write-Host "`n13. Testing Admin Endpoints..." -ForegroundColor Cyan
try {
    # Purge expired logs
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/audit/purge-expired"
    Write-Host "✓ Purged expired logs" -ForegroundColor Green
    
    # Cleanup expired exports
    $response = Invoke-ApiRequest -Method "POST" -Endpoint "/api/audit/cleanup-exports"
    Write-Host "✓ Cleaned up expired exports" -ForegroundColor Green
}
catch {
    Write-Host "✗ Admin endpoints failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nAudit Log Integration Tests Completed!" -ForegroundColor Green
Write-Host "`nNotes:" -ForegroundColor Yellow
Write-Host "- Audit logs are automatically created for login, email operations, etc." -ForegroundColor Gray
Write-Host "- Export files are automatically cleaned up after 24 hours" -ForegroundColor Gray
Write-Host "- Retention policies control how long different event types are kept" -ForegroundColor Gray
Write-Host "- The audit worker should be running to handle cleanup tasks" -ForegroundColor Gray







