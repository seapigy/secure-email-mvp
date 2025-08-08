# Test Email Cleanup Worker - PowerShell Script
# Tests the automated email cleanup functionality

Write-Host "=== Email Cleanup Worker Test ===" -ForegroundColor Green
Write-Host ""

# Configuration
$API_BASE_URL = "http://localhost:8080"
$TEST_EMAIL = "test@example.com"
$TEST_PASSWORD = "testpassword123"

# Colors for output
function Write-Success { param($Message) Write-Host "✓ $Message" -ForegroundColor Green }
function Write-Error { param($Message) Write-Host "✗ $Message" -ForegroundColor Red }
function Write-Info { param($Message) Write-Host "ℹ $Message" -ForegroundColor Cyan }
function Write-Warning { param($Message) Write-Host "⚠ $Message" -ForegroundColor Yellow }

# Check if API server is running
Write-Info "Checking if API server is running..."
try {
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/health" -Method GET -TimeoutSec 5
    Write-Success "API server is running"
} catch {
    Write-Error "API server is not running. Please start the server first."
    Write-Host "Run: go run cmd/api/main.go" -ForegroundColor Yellow
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
    Write-Host "You may need to create a test user first." -ForegroundColor Yellow
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
    
    Write-Host "Initial Statistics:" -ForegroundColor Yellow
    Write-Host "  - Expired emails: $($statsResponse.stats.expired_emails)" -ForegroundColor White
    Write-Host "  - Burn-after-read emails: $($statsResponse.stats.burn_after_read_emails)" -ForegroundColor White
    Write-Host "  - Total emails with content: $($statsResponse.stats.total_emails_with_content)" -ForegroundColor White
    Write-Host "  - Emails pending deletion: $($statsResponse.summary.emails_pending_deletion)" -ForegroundColor White
    Write-Host ""
} catch {
    Write-Error "Failed to get email retention statistics"
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
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
    
    Write-Host "Dry-run Results:" -ForegroundColor Yellow
    Write-Host "  - Message: $($dryRunResponse.message)" -ForegroundColor White
    Write-Host "  - Dry run: $($dryRunResponse.dry_run)" -ForegroundColor White
    Write-Host ""
} catch {
    Write-Error "Failed to run dry-run cleanup"
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 4: Ask user if they want to run actual cleanup
Write-Warning "Step 4: Manual cleanup confirmation"
Write-Host "Do you want to run actual cleanup (this will permanently delete emails)? (y/N)" -ForegroundColor Yellow
$response = Read-Host

if ($response -eq "y" -or $response -eq "Y") {
    Write-Info "Running actual cleanup..."
    try {
        $actualCleanupBody = @{
            dry_run = $false
        } | ConvertTo-Json
        
        $actualCleanupResponse = Invoke-RestMethod -Uri "$API_BASE_URL/admin/manual-cleanup" -Method POST -Body $actualCleanupBody -Headers $headers
        Write-Success "Actual cleanup completed"
        
        Write-Host "Actual Cleanup Results:" -ForegroundColor Yellow
        Write-Host "  - Message: $($actualCleanupResponse.message)" -ForegroundColor White
        Write-Host "  - Dry run: $($actualCleanupResponse.dry_run)" -ForegroundColor White
        
        if ($actualCleanupResponse.stats_before -and $actualCleanupResponse.stats_after) {
            Write-Host "  - Before cleanup:" -ForegroundColor White
            Write-Host "    * Expired emails: $($actualCleanupResponse.stats_before.expired_emails)" -ForegroundColor White
            Write-Host "    * Burn-after-read emails: $($actualCleanupResponse.stats_before.burn_after_read_emails)" -ForegroundColor White
            Write-Host "    * Total emails with content: $($actualCleanupResponse.stats_before.total_emails_with_content)" -ForegroundColor White
            
            Write-Host "  - After cleanup:" -ForegroundColor White
            Write-Host "    * Expired emails: $($actualCleanupResponse.stats_after.expired_emails)" -ForegroundColor White
            Write-Host "    * Burn-after-read emails: $($actualCleanupResponse.stats_after.burn_after_read_emails)" -ForegroundColor White
            Write-Host "    * Total emails with content: $($actualCleanupResponse.stats_after.total_emails_with_content)" -ForegroundColor White
        }
        Write-Host ""
    } catch {
        Write-Error "Failed to run actual cleanup"
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    }
} else {
    Write-Info "Skipping actual cleanup"
}

# Step 5: Get final statistics
Write-Info "Step 5: Getting final email retention statistics..."
try {
    $finalStatsResponse = Invoke-RestMethod -Uri "$API_BASE_URL/admin/email-retention-stats" -Method GET -Headers $headers
    Write-Success "Retrieved final email retention statistics"
    
    Write-Host "Final Statistics:" -ForegroundColor Yellow
    Write-Host "  - Expired emails: $($finalStatsResponse.stats.expired_emails)" -ForegroundColor White
    Write-Host "  - Burn-after-read emails: $($finalStatsResponse.stats.burn_after_read_emails)" -ForegroundColor White
    Write-Host "  - Total emails with content: $($finalStatsResponse.stats.total_emails_with_content)" -ForegroundColor White
    Write-Host "  - Emails pending deletion: $($finalStatsResponse.summary.emails_pending_deletion)" -ForegroundColor White
    Write-Host "  - Cleanup interval: $($finalStatsResponse.stats.cleanup_interval_minutes) minutes" -ForegroundColor White
    Write-Host ""
} catch {
    Write-Error "Failed to get final email retention statistics"
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Step 6: Test standalone worker (if available)
Write-Info "Step 6: Testing standalone worker availability..."
$workerPath = "cmd/workers/email_cleanup_worker.go"
if (Test-Path $workerPath) {
    Write-Success "Standalone worker source found"
    Write-Host "To build and run the standalone worker:" -ForegroundColor Yellow
    Write-Host "  go build -o email-cleanup-worker cmd/workers/email_cleanup_worker.go" -ForegroundColor White
    Write-Host "  EMAIL_CLEANUP_INTERVAL_MINUTES=5 ./email-cleanup-worker" -ForegroundColor White
    Write-Host ""
} else {
    Write-Warning "Standalone worker source not found at $workerPath"
}

# Step 7: Manual testing instructions
Write-Info "Step 7: Manual testing instructions"
Write-Host "To test the cleanup worker manually:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Create test emails with expiration:" -ForegroundColor White
Write-Host "   - Send emails with expiresAt set to past time" -ForegroundColor Gray
Write-Host "   - Send burn-after-read emails and access them" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Check statistics via API:" -ForegroundColor White
Write-Host "   curl -X GET http://localhost:8080/admin/email-retention-stats \\" -ForegroundColor Gray
Write-Host "     -H \"Authorization: Bearer YOUR_JWT_TOKEN\"" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Run manual cleanup:" -ForegroundColor White
Write-Host "   curl -X POST http://localhost:8080/admin/manual-cleanup \\" -ForegroundColor Gray
Write-Host "     -H \"Authorization: Bearer YOUR_JWT_TOKEN\" \\" -ForegroundColor Gray
Write-Host "     -H \"Content-Type: application/json\" \\" -ForegroundColor Gray
Write-Host "     -d '{\"dry_run\": false}'" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Monitor logs for cleanup activity" -ForegroundColor White
Write-Host ""

Write-Success "Email cleanup worker test completed!"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Review the documentation in docs/email_cleanup_worker.md" -ForegroundColor White
Write-Host "2. Configure EMAIL_CLEANUP_INTERVAL_MINUTES in your environment" -ForegroundColor White
Write-Host "3. Deploy the worker as a service or integrate with the main API" -ForegroundColor White
Write-Host "4. Set up monitoring and alerting for the cleanup process" -ForegroundColor White

