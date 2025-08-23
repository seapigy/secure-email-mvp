# Admin Management Test Runner - Iteration 4
# Secure Email MVP - Multi-Admin Access and RBAC Enforcement

param(
    [string]$Environment = "staging",
    [switch]$SafeMode,
    [switch]$GenerateReports,
    [string]$OutputDir = "./admin_management_results"
)

# Configuration
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Colors for output
$Colors = @{
    Success = "Green"
    Warning = "Yellow"
    Error = "Red"
    Info = "Cyan"
    Header = "Magenta"
}

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Output $Message
}

function Write-Header {
    param([string]$Title)
    Write-ColorOutput "`n" -Color Info
    Write-ColorOutput "=" * 80 -Color Header
    Write-ColorOutput " $Title" -Color Header
    Write-ColorOutput "=" * 80 -Color Header
    Write-ColorOutput "`n" -Color Info
}

function Write-Section {
    param([string]$Title)
    Write-ColorOutput "`n--- $Title ---" -Color Info
}

function Test-Prerequisites {
    Write-Header "Checking Prerequisites"

    $prerequisites = @{
        "Go" = "go version"
        "Node.js" = "node --version"
        "SQLite" = "sqlite3 --version"
        "PowerShell" = "pwsh --version"
    }

    $missing = @()

    foreach ($tool in $prerequisites.Keys) {
        $command = $prerequisites[$tool]
        try {
            $result = Invoke-Expression $command 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-ColorOutput "✓ $tool is available" -Color Success
            } else {
                $missing += $tool
                Write-ColorOutput "✗ $tool is missing" -Color Error
            }
        } catch {
            $missing += $tool
            Write-ColorOutput "✗ $tool is missing" -Color Error
        }
    }

    if ($missing.Count -gt 0) {
        Write-ColorOutput "`nMissing prerequisites: $($missing -join ', ')" -Color Error
        Write-ColorOutput "Please install the missing tools before running admin management tests." -Color Error
        exit 1
    }

    Write-ColorOutput "`nAll prerequisites are satisfied!" -Color Success
}

function Test-Environment {
    Write-Header "Environment Validation"

    # Check if we're in staging environment
    if ($Environment -ne "staging") {
        Write-ColorOutput "⚠️  Warning: Running admin management tests in non-staging environment" -Color Warning
        Write-ColorOutput "Environment: $Environment" -Color Warning

        if (-not $SafeMode) {
            $response = Read-Host "Do you want to continue? (y/N)"
            if ($response -ne "y" -and $response -ne "Y") {
                Write-ColorOutput "Operation cancelled by user." -Color Info
                exit 0
            }
        }
    }

    # Check if application is running
    try {
        $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/health" -Method GET -TimeoutSec 10
        Write-ColorOutput "✓ Application is running and healthy" -Color Success
    } catch {
        Write-ColorOutput "✗ Application is not running or not accessible" -Color Error
        Write-ColorOutput "Please start the application before running admin management tests." -Color Error
        exit 1
    }

    # Check database connectivity
    try {
        $dbTest = sqlite3 "secure_email_mvp.db" "SELECT COUNT(*) FROM sqlite_master;" 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✓ Database is accessible" -Color Success
        } else {
            throw "Database test failed"
        }
    } catch {
        Write-ColorOutput "✗ Database is not accessible" -Color Error
        Write-ColorOutput "Please ensure the database is properly initialized." -Color Error
        exit 1
    }

    Write-ColorOutput "`nEnvironment validation completed successfully!" -Color Success
}

function Start-AdminManagementTests {
    Write-Header "Admin Management Test Suite - Iteration 4"

    Write-Section "Building Test Binary"

    # Create output directory
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    }

    # Build the test binary
    Set-Location "scripts/admin_management"
    Write-ColorOutput "Building admin management test binary..." -Color Info

    go build -o admin_management_test test_invitation_system.go

    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✓ Test binary built successfully" -Color Success
    } else {
        Write-ColorOutput "✗ Failed to build test binary" -Color Error
        Set-Location "../.."
        exit 1
    }

    Write-Section "Running Admin Management Tests"

    # Run the test suite
    Write-ColorOutput "Starting admin management test suite..." -Color Info
    Write-ColorOutput "This will test:" -Color Info
    Write-ColorOutput "  • Admin Setup and Initial Login" -Color Info
    Write-ColorOutput "  • Invitation Key Creation" -Color Info
    Write-ColorOutput "  • Invitation Key Validation" -Color Info
    Write-ColorOutput "  • Admin Account Creation via Invitation" -Color Info
    Write-ColorOutput "  • RBAC Enforcement - Role-based Access" -Color Info
    Write-ColorOutput "  • Session Management" -Color Info
    Write-ColorOutput "  • Invitation Key Revocation" -Color Info
    Write-ColorOutput "  • Multi-Admin Workflow" -Color Info
    Write-ColorOutput "  • User Recovery Testing" -Color Info
    Write-ColorOutput "  • Security Validation" -Color Info

    $startTime = Get-Date

    ./admin_management_test

    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✓ Admin management test suite completed successfully" -Color Success
    } else {
        Write-ColorOutput "✗ Admin management test suite failed" -Color Error
        Set-Location "../.."
        exit 1
    }

    $endTime = Get-Date
    $duration = $endTime - $startTime

    Write-ColorOutput "Test duration: $($duration.TotalSeconds.ToString('F2')) seconds" -Color Info

    Set-Location "../.."
}

function Generate-AdminManagementReport {
    Write-Header "Generating Admin Management Report"

    # Find the latest test report
    $reportFiles = Get-ChildItem -Path "." -Filter "admin_management_test_report_*.json" | Sort-Object LastWriteTime -Descending

    if ($reportFiles.Count -eq 0) {
        Write-ColorOutput "✗ No test report files found" -Color Error
        return
    }

    $latestReport = $reportFiles[0]
    Write-ColorOutput "Processing report: $($latestReport.Name)" -Color Info

    # Read and parse the report
    try {
        $reportData = Get-Content $latestReport.FullName | ConvertFrom-Json

        $totalTests = $reportData.summary.total_tests
        $passedTests = $reportData.summary.passed_tests
        $failedTests = $reportData.summary.failed_tests
        $successRate = $reportData.summary.success_rate
        $totalDuration = $reportData.summary.total_duration_ms

        Write-ColorOutput "`nTest Results Summary:" -Color Header
        Write-ColorOutput "  Total Tests: $totalTests" -Color Info
        Write-ColorOutput "  Passed: $passedTests" -Color Success
        Write-ColorOutput "  Failed: $failedTests" -Color Error
        Write-ColorOutput "  Success Rate: $($successRate.ToString('F1'))%" -Color Info
        Write-ColorOutput "  Total Duration: $($totalDuration.ToString('F0'))ms" -Color Info

        # Show detailed results
        Write-ColorOutput "`nDetailed Test Results:" -Color Header
        foreach ($test in $reportData.results) {
            $status = if ($test.success) { "✅ PASS" } else { "❌ FAIL" }
            $color = if ($test.success) { "Success" } else { "Error" }
            $duration = $test.duration.ToString('F0')

            Write-ColorOutput "  $status - $($test.test_name) (${duration}ms)" -Color $color

            if (-not $test.success -and $test.error) {
                Write-ColorOutput "    Error: $($test.error)" -Color Error
            }

            if ($test.details) {
                Write-ColorOutput "    Details: $($test.details)" -Color Info
            }
        }

        # Move report to output directory
        $outputPath = Join-Path $OutputDir $latestReport.Name
        Move-Item $latestReport.FullName $outputPath -Force
        Write-ColorOutput "`nReport saved to: $outputPath" -Color Success

    } catch {
        Write-ColorOutput "✗ Failed to process test report: $($_.Exception.Message)" -Color Error
    }
}

function Show-Summary {
    Write-Header "Admin Management Test Suite Summary"

    Write-ColorOutput "Iteration 4 - Admin & User Management completed!" -Color Success
    Write-ColorOutput "`nWhat was tested:" -Color Header
    Write-ColorOutput "  ✓ Admin Invitations: One-time links with 24-hour expiration" -Color Success
    Write-ColorOutput "  ✓ RBAC Enforcement: Role-based access for all admin panels" -Color Success
    Write-ColorOutput "  ✓ Session Management: Expiration, auto-logout, concurrent limits" -Color Success
    Write-ColorOutput "  ✓ User Recovery Testing: Recovery code generation under ZKID" -Color Success
    Write-ColorOutput "  ✓ Security Validation: Authentication and authorization checks" -Color Success

    Write-ColorOutput "`nKey Features Implemented:" -Color Header
    Write-ColorOutput "  • Secure invitation key system with role-based scope" -Color Info
    Write-ColorOutput "  • Comprehensive RBAC with root admin authority" -Color Info
    Write-ColorOutput "  • Session management with security controls" -Color Info
    Write-ColorOutput "  • Audit logging for all admin actions" -Color Info
    Write-ColorOutput "  • Multi-admin workflow validation" -Color Info

    Write-ColorOutput "`nNext Steps:" -Color Header
    Write-ColorOutput "  • Review test results for any failures" -Color Info
    Write-ColorOutput "  • Validate invitation workflow in production" -Color Info
    Write-ColorOutput "  • Monitor admin audit logs for security" -Color Info
    Write-ColorOutput "  • Consider additional RBAC refinements" -Color Info
}

# Main execution
try {
    Write-Header "Iteration 4 - Admin & User Management Test Suite"
    Write-ColorOutput "Testing secure multi-admin access and RBAC enforcement" -Color Info

    Test-Prerequisites
    Test-Environment
    Start-AdminManagementTests
    Generate-AdminManagementReport
    Show-Summary

    Write-ColorOutput "`n🎉 Admin Management Test Suite completed successfully!" -Color Success

} catch {
    Write-ColorOutput "`n💥 Admin Management Test Suite failed: $($_.Exception.Message)" -Color Error
    exit 1
}
