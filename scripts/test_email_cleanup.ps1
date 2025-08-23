# Test Email Cleanup Worker - PowerShell Script
# Tests the automated email cleanup functionality

Write-Output "=== Email Cleanup Worker Test ==="
Write-Output ""

# Configuration
$API_BASE_URL = "http://localhost:8080"
$TEST_EMAIL = "test@example.com"
$TEST_PASSWORD = "testpassword123"

# Colors for output
function Write-Success { param($Message) Write-Output "✓ $Message" }
function Write-Error { param($Message) Write-Output "✗ $Message" }
function Write-Info { param($Message) Write-Output "ℹ $Message" }
function Write-Warning { param($Message) Write-Output "⚠ $Message" }

# Check if API server is running
Write-Info "Checking if API server is running..."
try {
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/health" -Method GET -TimeoutSec 5
    Write-Success "API server is running"
} catch {
    Write-Error "API server is not running. Please start the server first."
    Write-Output "Run: go run cmd/api/main.go"
    exit 1
}

# Step 1: Login to get JWT token
Write-Info "Step 1: Logging in to get JWT token..."
$loginBody = @{
    email = $TEST_EMAIL
    password = $TEST_PASSWORD
    totp_code = "123456"  # Default TOTP for testing
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$API_BASE_URL/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $jwtToken = $loginResponse.token
    Write-Success "Login successful, JWT token obtained"
} catch {
    Write-Error "Login failed. Please ensure the test user exists and TOTP is configured."
    Write-Output "You may need to create a test user first."
    exit 1
}

# Step 2: Get initial email retention statistics
Write-Info "Step 2: Getting initial email retention statistics..."
try {
    $headers = @{
        "Authorization" = "Bearer $jwtToken"
        "Content-Type" = "application/json"
    }

    $statsResponse = Invoke-RestMethod -Uri "$API_BASE_URL/admin/email-retention-stats" -Method GET -Headers $headers
    Write-Success "Retrieved email retention statistics"

    Write-Output "Initial Statistics:"
    Write-Output "  - Expired emails: $($statsResponse.stats.expired_emails)"
    Write-Output "  - Burn-after-read emails: $($statsResponse.stats.burn_after_read_emails)"
    Write-Output "  - Total emails with content: $($statsResponse.stats.total_emails_with_content)"
    Write-Output "  - Emails pending deletion: $($statsResponse.summary.emails_pending_deletion)"
    Write-Output ""
} catch {
    Write-Error "Failed to get email retention statistics"
    Write-Output "Error: $($_.Exception.Message)"
    exit 1
}

# Step 3: Run dry-run cleanup
Write-Info "Step 3: Running dry-run cleanup..."
try {
    $dryRunBody = @{
        dry_run = $true
    } | ConvertTo-Json

    $dryRunResponse = Invoke-RestMethod -Uri "$API_BASE_URL/admin/manual-cleanup" -Method POST -Body $dryRunBody -Headers $headers
    Write-Success "Dry-run cleanup completed"

    Write-Output "Dry-run Results:"
    Write-Output "  - Message: $($dryRunResponse.message)"
    Write-Output "  - Dry run: $($dryRunResponse.dry_run)"
    Write-Output ""
} catch {
    Write-Error "Failed to run dry-run cleanup"
    Write-Output "Error: $($_.Exception.Message)"
    exit 1
}

# Step 4: Ask user if they want to run actual cleanup
Write-Warning "Step 4: Manual cleanup confirmation"
Write-Output "Do you want to run actual cleanup (this will permanently delete emails)? (y/N)"
$response = Read-Host

if ($response -eq "y" -or $response -eq "Y") {
    Write-Info "Running actual cleanup..."
    try {
        $actualCleanupBody = @{
            dry_run = $false
        } | ConvertTo-Json

        $actualCleanupResponse = Invoke-RestMethod -Uri "$API_BASE_URL/admin/manual-cleanup" -Method POST -Body $actualCleanupBody -Headers $headers
        Write-Success "Actual cleanup completed"

        Write-Output "Actual Cleanup Results:"
        Write-Output "  - Message: $($actualCleanupResponse.message)"
        Write-Output "  - Dry run: $($actualCleanupResponse.dry_run)"

        if ($actualCleanupResponse.stats_before -and $actualCleanupResponse.stats_after) {
            Write-Output "  - Before cleanup:"
            Write-Output "    * Expired emails: $($actualCleanupResponse.stats_before.expired_emails)"
            Write-Output "    * Burn-after-read emails: $($actualCleanupResponse.stats_before.burn_after_read_emails)"
            Write-Output "    * Total emails with content: $($actualCleanupResponse.stats_before.total_emails_with_content)"

            Write-Output "  - After cleanup:"
            Write-Output "    * Expired emails: $($actualCleanupResponse.stats_after.expired_emails)"
            Write-Output "    * Burn-after-read emails: $($actualCleanupResponse.stats_after.burn_after_read_emails)"
            Write-Output "    * Total emails with content: $($actualCleanupResponse.stats_after.total_emails_with_content)"
        }
        Write-Output ""
    } catch {
        Write-Error "Failed to run actual cleanup"
        Write-Output "Error: $($_.Exception.Message)"
    }
} else {
    Write-Info "Skipping actual cleanup"
}

# Step 5: Get final statistics
Write-Info "Step 5: Getting final email retention statistics..."
try {
    $finalStatsResponse = Invoke-RestMethod -Uri "$API_BASE_URL/admin/email-retention-stats" -Method GET -Headers $headers
    Write-Success "Retrieved final email retention statistics"

    Write-Output "Final Statistics:"
    Write-Output "  - Expired emails: $($finalStatsResponse.stats.expired_emails)"
    Write-Output "  - Burn-after-read emails: $($finalStatsResponse.stats.burn_after_read_emails)"
    Write-Output "  - Total emails with content: $($finalStatsResponse.stats.total_emails_with_content)"
    Write-Output "  - Emails pending deletion: $($finalStatsResponse.summary.emails_pending_deletion)"
    Write-Output "  - Cleanup interval: $($finalStatsResponse.stats.cleanup_interval_minutes) minutes"
    Write-Output ""
} catch {
    Write-Error "Failed to get final email retention statistics"
    Write-Output "Error: $($_.Exception.Message)"
}

# Step 6: Test standalone worker (if available)
Write-Info "Step 6: Testing standalone worker availability..."
$workerPath = "cmd/workers/email_cleanup_worker.go"
if (Test-Path $workerPath) {
    Write-Success "Standalone worker source found"
    Write-Output "To build and run the standalone worker:"
    Write-Output "  go build -o email-cleanup-worker cmd/workers/email_cleanup_worker.go"
    Write-Output "  EMAIL_CLEANUP_INTERVAL_MINUTES=5 ./email-cleanup-worker"
    Write-Output ""
} else {
    Write-Warning "Standalone worker source not found at $workerPath"
}

# Step 7: Manual testing instructions
Write-Info "Step 7: Manual testing instructions"
Write-Output "To test the cleanup worker manually:"
Write-Output ""
Write-Output "1. Create test emails with expiration:"
Write-Output "   - Send emails with expiresAt set to past time"
Write-Output "   - Send burn-after-read emails and access them"
Write-Output ""
Write-Output "2. Check statistics via API:"
Write-Output "   curl -X GET http://localhost:8080/admin/email-retention-stats \\"
Write-Output "     -H \"Authorization: Bearer YOUR_JWT_TOKEN\"" -ForegroundColor Gray
Write-Output ""
Write-Output "3. Run manual cleanup:"
Write-Output "   curl -X POST http://localhost:8080/admin/manual-cleanup \\"
Write-Output "     -H \"Authorization: Bearer YOUR_JWT_TOKEN\" \\" -ForegroundColor Gray
Write-Output "     -H \"Content-Type: application/json\" \\" -ForegroundColor Gray
Write-Output "     -d '{\"dry_run\": false}'" -ForegroundColor Gray
Write-Output ""
Write-Output "4. Monitor logs for cleanup activity"
Write-Output ""

Write-Success "Email cleanup worker test completed!"
Write-Output ""
Write-Output "Next steps:"
Write-Output "1. Review the documentation in docs/email_cleanup_worker.md"
Write-Output "2. Configure EMAIL_CLEANUP_INTERVAL_MINUTES in your environment"
Write-Output "3. Deploy the worker as a service or integrate with the main API"
Write-Output "4. Set up monitoring and alerting for the cleanup process"














