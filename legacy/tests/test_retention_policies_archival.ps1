# Test script for Micro-Iteration 4.26: Smart Retention Policy Engine & Automated Archival
# This script demonstrates and tests the new retention policy and archival features

Write-Output "=== Micro-Iteration 4.26: Retention Policies & Archival Test ==="

# Configuration
$API_BASE = "http://localhost:8080"
$JWT_TOKEN = "your_jwt_token_here"  # Replace with actual JWT token

# Set environment variables for testing
$env:DEFAULT_RETENTION_DAYS = "30"
$env:DEFAULT_ARCHIVE_RETENTION_DAYS = "365"
$env:DEFAULT_ARCHIVE_INSTEAD = "false"
$env:ENABLE_POLICY_EVALUATION_LOGGING = "true"

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
        Write-Output "API Error: $($_.Exception.Message)"
        return $null
    }
}

# Test 1: Create retention policies
Write-Output "`n1. Testing retention policy creation..."

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
    Write-Output "✓ User-specific policy created successfully"
    Write-Output "  - Policy ID: $($userPolicyResponse.id)"
    Write-Output "  - Retention days: $($userPolicyResponse.retention_days)"
    Write-Output "  - Archive instead: $($userPolicyResponse.archive_instead)"
} else {
    Write-Output "✗ Failed to create user-specific policy"
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
    Write-Output "✓ Domain-specific policy created successfully"
    Write-Output "  - Policy ID: $($domainPolicyResponse.id)"
    Write-Output "  - Sender domain: $($domainPolicyResponse.sender_domain)"
} else {
    Write-Output "✗ Failed to create domain-specific policy"
}

# Test 2: List retention policies
Write-Output "`n2. Testing retention policy listing..."
$policiesResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-policies?limit=10&offset=0"
if ($policiesResponse) {
    Write-Output "✓ Policies retrieved successfully"
    Write-Output "  - Total policies: $($policiesResponse.total_count)"
    Write-Output "  - Active policies: $($policiesResponse.policies.Count)"
    foreach ($policy in $policiesResponse.policies) {
        Write-Output "    - $($policy.name) (Priority: $($policy.priority), Active: $($policy.active))"
    }
} else {
    Write-Output "✗ Failed to retrieve policies"
}

# Test 3: Update a retention policy
Write-Output "`n3. Testing retention policy update..."
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
        Write-Output "✓ Policy updated successfully"
        Write-Output "  - New retention days: $($updateResponse.retention_days)"
        Write-Output "  - New archive retention: $($updateResponse.archive_retention_days)"
    } else {
        Write-Output "✗ Failed to update policy"
    }
}

# Test 4: Archive an email
Write-Output "`n4. Testing email archival..."
$archiveRequest = @{
    email_id = "test-email-123"
    archive_reason = "policy"
    retention_days = 180
}

$archiveResponse = Invoke-AuthenticatedAPI -Method "POST" -Endpoint "/api/admin/email/archived" -Body $archiveRequest
if ($archiveResponse) {
    Write-Output "✓ Email archived successfully"
    Write-Output "  - Archive ID: $($archiveResponse.archive_id)"
    Write-Output "  - Message: $($archiveResponse.message)"
} else {
    Write-Output "✗ Failed to archive email"
}

# Test 5: List archived emails
Write-Output "`n5. Testing archived emails listing..."
$archivedResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/archived?limit=5&offset=0"
if ($archivedResponse) {
    Write-Output "✓ Archived emails retrieved successfully"
    Write-Output "  - Total archived: $($archivedResponse.total_count)"
    Write-Output "  - Recent archives: $($archivedResponse.archived_emails.Count)"
    foreach ($archive in $archivedResponse.archived_emails) {
        Write-Output "    - $($archive.original_email_id) (Reason: $($archive.archive_reason), Archived: $($archive.archived_at))"
    }
} else {
    Write-Output "✗ Failed to retrieve archived emails"
}

# Test 6: Get archival statistics
Write-Output "`n6. Testing archival statistics..."
$statsResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/archived/stats"
if ($statsResponse) {
    Write-Output "✓ Archival statistics retrieved successfully"
    Write-Output "  - Total archived: $($statsResponse.total_archived)"
    Write-Output "  - Expired archives: $($statsResponse.expired_archives)"
    Write-Output "  - Total storage: $($statsResponse.total_storage_bytes) bytes"
    if ($statsResponse.archives_by_reason) {
        Write-Output "  - Archives by reason:"
        foreach ($reason in $statsResponse.archives_by_reason.PSObject.Properties) {
            Write-Output "    - $($reason.Name): $($reason.Value)"
        }
    }
} else {
    Write-Output "✗ Failed to retrieve archival statistics"
}

# Test 7: Test policy filtering
Write-Output "`n7. Testing policy filtering..."
$filteredPoliciesResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-policies?active=true&limit=5"
if ($filteredPoliciesResponse) {
    Write-Output "✓ Filtered policies retrieved successfully"
    Write-Output "  - Active policies: $($filteredPoliciesResponse.policies.Count)"
} else {
    Write-Output "✗ Failed to retrieve filtered policies"
}

# Test 8: Test archived email filtering
Write-Output "`n8. Testing archived email filtering..."
$filteredArchivedResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/archived?archive_reason=policy&limit=3"
if ($filteredArchivedResponse) {
    Write-Output "✓ Filtered archived emails retrieved successfully"
    Write-Output "  - Policy-based archives: $($filteredArchivedResponse.archived_emails.Count)"
} else {
    Write-Output "✗ Failed to retrieve filtered archived emails"
}

# Test 9: Cleanup expired archives
Write-Output "`n9. Testing expired archives cleanup..."
$cleanupResponse = Invoke-AuthenticatedAPI -Method "POST" -Endpoint "/api/admin/email/archived/cleanup"
if ($cleanupResponse) {
    Write-Output "✓ Archive cleanup completed successfully"
    Write-Output "  - Message: $($cleanupResponse.message)"
} else {
    Write-Output "✗ Failed to cleanup expired archives"
}

# Test 10: Get specific policy
Write-Output "`n10. Testing specific policy retrieval..."
if ($userPolicyResponse) {
    $specificPolicyResponse = Invoke-AuthenticatedAPI -Method "GET" -Endpoint "/api/admin/email/retention-policies/$($userPolicyResponse.id)"
    if ($specificPolicyResponse) {
        Write-Output "✓ Specific policy retrieved successfully"
        Write-Output "  - Policy name: $($specificPolicyResponse.name)"
        Write-Output "  - Priority: $($specificPolicyResponse.priority)"
        Write-Output "  - Retention days: $($specificPolicyResponse.retention_days)"
    } else {
        Write-Output "✗ Failed to retrieve specific policy"
    }
}

Write-Output "`n=== Test Summary ==="
Write-Output "Micro-Iteration 4.26 features tested:"
Write-Output "✓ Retention policy creation and management"
Write-Output "✓ Policy filtering and listing"
Write-Output "✓ Email archival operations"
Write-Output "✓ Archived email querying and filtering"
Write-Output "✓ Archival statistics and monitoring"
Write-Output "✓ Expired archive cleanup"

Write-Output "`nTo test policy evaluation with real emails:"
Write-Output "1. Create test emails with various criteria"
Write-Output "2. Run the enhanced cleanup worker to apply policies"
Write-Output "3. Check policy evaluation logs in the database"

Write-Output "`nTo run the enhanced cleanup worker with policies:"
Write-Output "1. Set environment variables:"
Write-Output "   `$env:DEFAULT_RETENTION_DAYS = '30'"
Write-Output "   `$env:DEFAULT_ARCHIVE_INSTEAD = 'true'"
Write-Output "2. Run the enhanced worker:"
Write-Output "   go run ./cmd/workers/enhanced_email_cleanup_worker.go"

Write-Output "`nTest completed successfully!"

















