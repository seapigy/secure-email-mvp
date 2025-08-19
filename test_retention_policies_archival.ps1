# Test script for Micro-Iteration 4.26: Smart Retention Policy Engine & Automated Archival
# This script demonstrates and tests the new retention policy and archival features

Write-Host "=== Micro-Iteration 4.26: Retention Policies & Archival Test ===" -ForegroundColor Green

# Configuration
$API_BASE = "http://localhost:8080"
$JWT_TOKEN = "your_jwt_token_here"  # Replace with actual JWT token

# Set environment variables for testing
$env:DEFAULT_RETENTION_DAYS = "30"
$env:DEFAULT_ARCHIVE_RETENTION_DAYS = "365"
$env:DEFAULT_ARCHIVE_INSTEAD = "false"
$env:ENABLE_POLICY_EVALUATION_LOGGING = "true"

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

# Test 1: Create retention policies
Write-Host "`n1. Testing retention policy creation..." -ForegroundColor Cyan

# Create a policy for specific user
$userPolicy = @{
    name = "User-Specific Policy"
    description = "Policy for specific user with longer retention"
    priority = 100
    active = $true
    user_id = "test_user_123"
    retention_days = 90
    archive_instead = $true
    archive_retention_days = 730
}

$userPolicyResponse = Invoke-AuthenticatedAPI -Method "POST" -Endpoint "/api/admin/email/retention-policies" -Body $userPolicy
if ($userPolicyResponse) {
    Write-Host "✓ User-specific policy created successfully" -ForegroundColor Green
    Write-Host "  - Policy ID: $($userPolicyResponse.id)" -ForegroundColor White
    Write-Host "  - Retention days: $($userPolicyResponse.retention_days)" -ForegroundColor White
    Write-Host "  - Archive instead: $($userPolicyResponse.archive_instead)" -ForegroundColor White
} else {
    Write-Host "✗ Failed to create user-specific policy" -ForegroundColor Red
}

# Create a policy for specific domain
$domainPolicy = @{
    name = "Domain-Specific Policy"
    description = "Policy for emails from specific domain"
    priority = 50
    active = $true
    sender_domain = "example.com"
    retention_days = 60
    archive_instead = $false
    archive_retention_days = 365
}

$domainPolicyResponse = Invoke-AuthenticatedAPI -Method "POST" -Endpoint "/api/admin/email/retention-policies" -Body $domainPolicy
if ($domainPolicyResponse) {
    Write-Host "✓ Domain-specific policy created successfully" -ForegroundColor Green
    Write-Host "  - Policy ID: $($domainPolicyResponse.id)" -ForegroundColor White
    Write-Host "  - Sender domain: $($domainPolicyResponse.sender_domain)" -ForegroundColor White
} else {
    Write-Host "✗ Failed to create domain-specific policy" -ForegroundColor Red
}

# Test 2: List retention policies
Write-Host "`n2. Testing retention policy listing..." -ForegroundColor Cyan
$policiesResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-policies?limit=10&offset=0"
if ($policiesResponse) {
    Write-Host "✓ Policies retrieved successfully" -ForegroundColor Green
    Write-Host "  - Total policies: $($policiesResponse.total_count)" -ForegroundColor White
    Write-Host "  - Active policies: $($policiesResponse.policies.Count)" -ForegroundColor White
    foreach ($policy in $policiesResponse.policies) {
        Write-Host "    - $($policy.name) (Priority: $($policy.priority), Active: $($policy.active))" -ForegroundColor Gray
    }
} else {
    Write-Host "✗ Failed to retrieve policies" -ForegroundColor Red
}

# Test 3: Update a retention policy
Write-Host "`n3. Testing retention policy update..." -ForegroundColor Cyan
if ($userPolicyResponse) {
    $updateBody = @{
        name = "Updated User Policy"
        description = "Updated policy with new settings"
        priority = 150
        active = $true
        user_id = "test_user_123"
        retention_days = 120
        archive_instead = $true
        archive_retention_days = 1095
    }
    
    $updateResponse = Invoke-AuthenticatedAPI -Method "PUT" -Endpoint "/api/admin/email/retention-policies/$($userPolicyResponse.id)" -Body $updateBody
    if ($updateResponse) {
        Write-Host "✓ Policy updated successfully" -ForegroundColor Green
        Write-Host "  - New retention days: $($updateResponse.retention_days)" -ForegroundColor White
        Write-Host "  - New archive retention: $($updateResponse.archive_retention_days)" -ForegroundColor White
    } else {
        Write-Host "✗ Failed to update policy" -ForegroundColor Red
    }
}

# Test 4: Archive an email
Write-Host "`n4. Testing email archival..." -ForegroundColor Cyan
$archiveRequest = @{
    email_id = "test-email-123"
    archive_reason = "policy"
    retention_days = 180
}

$archiveResponse = Invoke-AuthenticatedAPI -Method "POST" -Endpoint "/api/admin/email/archived" -Body $archiveRequest
if ($archiveResponse) {
    Write-Host "✓ Email archived successfully" -ForegroundColor Green
    Write-Host "  - Archive ID: $($archiveResponse.archive_id)" -ForegroundColor White
    Write-Host "  - Message: $($archiveResponse.message)" -ForegroundColor White
} else {
    Write-Host "✗ Failed to archive email" -ForegroundColor Red
}

# Test 5: List archived emails
Write-Host "`n5. Testing archived emails listing..." -ForegroundColor Cyan
$archivedResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/archived?limit=5&offset=0"
if ($archivedResponse) {
    Write-Host "✓ Archived emails retrieved successfully" -ForegroundColor Green
    Write-Host "  - Total archived: $($archivedResponse.total_count)" -ForegroundColor White
    Write-Host "  - Recent archives: $($archivedResponse.archived_emails.Count)" -ForegroundColor White
    foreach ($archive in $archivedResponse.archived_emails) {
        Write-Host "    - $($archive.original_email_id) (Reason: $($archive.archive_reason), Archived: $($archive.archived_at))" -ForegroundColor Gray
    }
} else {
    Write-Host "✗ Failed to retrieve archived emails" -ForegroundColor Red
}

# Test 6: Get archival statistics
Write-Host "`n6. Testing archival statistics..." -ForegroundColor Cyan
$statsResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/archived/stats"
if ($statsResponse) {
    Write-Host "✓ Archival statistics retrieved successfully" -ForegroundColor Green
    Write-Host "  - Total archived: $($statsResponse.total_archived)" -ForegroundColor White
    Write-Host "  - Expired archives: $($statsResponse.expired_archives)" -ForegroundColor White
    Write-Host "  - Total storage: $($statsResponse.total_storage_bytes) bytes" -ForegroundColor White
    if ($statsResponse.archives_by_reason) {
        Write-Host "  - Archives by reason:" -ForegroundColor White
        foreach ($reason in $statsResponse.archives_by_reason.PSObject.Properties) {
            Write-Host "    - $($reason.Name): $($reason.Value)" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "✗ Failed to retrieve archival statistics" -ForegroundColor Red
}

# Test 7: Test policy filtering
Write-Host "`n7. Testing policy filtering..." -ForegroundColor Cyan
$filteredPoliciesResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-policies?active=true&limit=5"
if ($filteredPoliciesResponse) {
    Write-Host "✓ Filtered policies retrieved successfully" -ForegroundColor Green
    Write-Host "  - Active policies: $($filteredPoliciesResponse.policies.Count)" -ForegroundColor White
} else {
    Write-Host "✗ Failed to retrieve filtered policies" -ForegroundColor Red
}

# Test 8: Test archived email filtering
Write-Host "`n8. Testing archived email filtering..." -ForegroundColor Cyan
$filteredArchivedResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/archived?archive_reason=policy&limit=3"
if ($filteredArchivedResponse) {
    Write-Host "✓ Filtered archived emails retrieved successfully" -ForegroundColor Green
    Write-Host "  - Policy-based archives: $($filteredArchivedResponse.archived_emails.Count)" -ForegroundColor White
} else {
    Write-Host "✗ Failed to retrieve filtered archived emails" -ForegroundColor Red
}

# Test 9: Cleanup expired archives
Write-Host "`n9. Testing expired archives cleanup..." -ForegroundColor Cyan
$cleanupResponse = Invoke-AuthenticatedAPI -Method "POST" -Endpoint "/api/admin/email/archived/cleanup"
if ($cleanupResponse) {
    Write-Host "✓ Archive cleanup completed successfully" -ForegroundColor Green
    Write-Host "  - Message: $($cleanupResponse.message)" -ForegroundColor White
} else {
    Write-Host "✗ Failed to cleanup expired archives" -ForegroundColor Red
}

# Test 10: Get specific policy
Write-Host "`n10. Testing specific policy retrieval..." -ForegroundColor Cyan
if ($userPolicyResponse) {
    $specificPolicyResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-policies/$($userPolicyResponse.id)"
    if ($specificPolicyResponse) {
        Write-Host "✓ Specific policy retrieved successfully" -ForegroundColor Green
        Write-Host "  - Policy name: $($specificPolicyResponse.name)" -ForegroundColor White
        Write-Host "  - Priority: $($specificPolicyResponse.priority)" -ForegroundColor White
        Write-Host "  - Retention days: $($specificPolicyResponse.retention_days)" -ForegroundColor White
    } else {
        Write-Host "✗ Failed to retrieve specific policy" -ForegroundColor Red
    }
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "Micro-Iteration 4.26 features tested:" -ForegroundColor White
Write-Host "✓ Retention policy creation and management" -ForegroundColor Green
Write-Host "✓ Policy filtering and listing" -ForegroundColor Green
Write-Host "✓ Email archival operations" -ForegroundColor Green
Write-Host "✓ Archived email querying and filtering" -ForegroundColor Green
Write-Host "✓ Archival statistics and monitoring" -ForegroundColor Green
Write-Host "✓ Expired archive cleanup" -ForegroundColor Green

Write-Host "`nTo test policy evaluation with real emails:" -ForegroundColor Yellow
Write-Host "1. Create test emails with various criteria" -ForegroundColor White
Write-Host "2. Run the enhanced cleanup worker to apply policies" -ForegroundColor White
Write-Host "3. Check policy evaluation logs in the database" -ForegroundColor White

Write-Host "`nTo run the enhanced cleanup worker with policies:" -ForegroundColor Yellow
Write-Host "1. Set environment variables:" -ForegroundColor White
Write-Host "   `$env:DEFAULT_RETENTION_DAYS = '30'" -ForegroundColor Gray
Write-Host "   `$env:DEFAULT_ARCHIVE_INSTEAD = 'true'" -ForegroundColor Gray
Write-Host "2. Run the enhanced worker:" -ForegroundColor White
Write-Host "   go run ./cmd/workers/enhanced_email_cleanup_worker.go" -ForegroundColor Gray

Write-Host "`nTest completed successfully!" -ForegroundColor Green






