# Operational Readiness Test Runner
# Secure Email MVP - Iteration 3

param(
    [string]$Environment = "staging",
    [string]$TestType = "all", # all, disaster-recovery, load-testing, feature-rollback
    [int]$ConcurrentUsers = 100,
    [string]$TestDuration = "5m",
    [switch]$SafeMode,
    [switch]$GenerateReports,
    [string]$OutputDir = "./operational_readiness_results"
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
    Write-Host $Message -ForegroundColor $Colors[$Color]
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
        Write-ColorOutput "Please install the missing tools before running operational readiness tests." -Color Error
        exit 1
    }
    
    Write-ColorOutput "`nAll prerequisites are satisfied!" -Color Success
}

function Test-Environment {
    Write-Header "Environment Validation"
    
    # Check if we're in staging environment
    if ($Environment -ne "staging") {
        Write-ColorOutput "⚠️  Warning: Running operational readiness tests in non-staging environment" -Color Warning
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
        Write-ColorOutput "Please start the application before running operational readiness tests." -Color Error
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

function Start-DisasterRecoveryTest {
    Write-Header "Disaster Recovery Testing"
    
    Write-Section "Creating Disaster Recovery Manager"
    
    # Create backup directory
    $backupDir = "./backups"
    if (-not (Test-Path $backupDir)) {
        New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
    }
    
    # Run disaster recovery test
    try {
        Write-ColorOutput "Creating full system backup..." -Color Info
        
        # Build and run disaster recovery test
        Set-Location "scripts/operational"
        go build -o disaster_recovery_test disaster_recovery.go
        
        if ($LASTEXITCODE -eq 0) {
            ./disaster_recovery_test
            if ($LASTEXITCODE -eq 0) {
                Write-ColorOutput "✓ Disaster recovery test completed successfully" -Color Success
            } else {
                Write-ColorOutput "✗ Disaster recovery test failed" -Color Error
                return $false
            }
        } else {
            Write-ColorOutput "✗ Failed to build disaster recovery test" -Color Error
            return $false
        }
        
        Set-Location "../.."
        
    } catch {
        Write-ColorOutput "✗ Disaster recovery test failed: $($_.Exception.Message)" -Color Error
        return $false
    }
    
    Write-ColorOutput "✓ Disaster recovery testing completed!" -Color Success
    return $true
}

function Start-LoadTesting {
    Write-Header "Load Testing"
    
    Write-Section "Configuring Load Test"
    Write-ColorOutput "Concurrent Users: $ConcurrentUsers" -Color Info
    Write-ColorOutput "Test Duration: $TestDuration" -Color Info
    
    # Update load test configuration
    $loadTestConfig = Get-Content "scripts/operational/load_test_config.json" | ConvertFrom-Json
    $loadTestConfig.concurrent_users = $ConcurrentUsers
    $loadTestConfig.test_duration = $TestDuration
    $loadTestConfig | ConvertTo-Json -Depth 10 | Set-Content "scripts/operational/load_test_config.json"
    
    Write-Section "Running Load Test"
    
    try {
        # Build and run load test
        Set-Location "scripts/operational"
        go build -o load_test load_testing.go
        
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "Starting load test with $ConcurrentUsers concurrent users for $TestDuration..." -Color Info
            
            $startTime = Get-Date
            ./load_test
            $endTime = Get-Date
            $duration = $endTime - $startTime
            
            if ($LASTEXITCODE -eq 0) {
                Write-ColorOutput "✓ Load test completed successfully" -Color Success
                Write-ColorOutput "Test duration: $($duration.TotalSeconds.ToString('F2')) seconds" -Color Info
            } else {
                Write-ColorOutput "✗ Load test failed" -Color Error
                return $false
            }
        } else {
            Write-ColorOutput "✗ Failed to build load test" -Color Error
            return $false
        }
        
        Set-Location "../.."
        
    } catch {
        Write-ColorOutput "✗ Load test failed: $($_.Exception.Message)" -Color Error
        return $false
    }
    
    Write-ColorOutput "✓ Load testing completed!" -Color Success
    return $true
}

function Start-FeatureFlagRollbackTest {
    Write-Header "Feature Flag Rollback Testing"
    
    Write-Section "Preparing Feature Flag Rollback Test"
    
    # Check current feature flag states
    Write-ColorOutput "Current feature flag states:" -Color Info
    $envVars = @("ENABLE_ZKID_LAYER", "ENABLE_PQC_ENCRYPTION", "ENABLE_HYBRID_TLS")
    
    foreach ($envVar in $envVars) {
        $value = [Environment]::GetEnvironmentVariable($envVar)
        if ($value) {
            Write-ColorOutput "  $envVar = $value" -Color Info
        } else {
            Write-ColorOutput "  $envVar = not set (default: true)" -Color Info
        }
    }
    
    Write-Section "Running Feature Flag Rollback Test"
    
    try {
        # Build and run feature flag rollback test
        Set-Location "scripts/operational"
        go build -o feature_rollback_test feature_flag_rollback.go
        
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "Starting feature flag rollback test..." -Color Info
            
            $startTime = Get-Date
            ./feature_rollback_test
            $endTime = Get-Date
            $duration = $endTime - $startTime
            
            if ($LASTEXITCODE -eq 0) {
                Write-ColorOutput "✓ Feature flag rollback test completed successfully" -Color Success
                Write-ColorOutput "Test duration: $($duration.TotalSeconds.ToString('F2')) seconds" -Color Info
            } else {
                Write-ColorOutput "✗ Feature flag rollback test failed" -Color Error
                return $false
            }
        } else {
            Write-ColorOutput "✗ Failed to build feature flag rollback test" -Color Error
            return $false
        }
        
        Set-Location "../.."
        
    } catch {
        Write-ColorOutput "✗ Feature flag rollback test failed: $($_.Exception.Message)" -Color Error
        return $false
    }
    
    Write-ColorOutput "✓ Feature flag rollback testing completed!" -Color Success
    return $true
}

function Generate-OperationalReadinessReport {
    Write-Header "Generating Operational Readiness Report"
    
    if (-not $GenerateReports) {
        Write-ColorOutput "Report generation skipped (use -GenerateReports to enable)" -Color Info
        return
    }
    
    # Create output directory
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    }
    
    $reportData = @{
        timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        environment = $Environment
        test_type = $TestType
        concurrent_users = $ConcurrentUsers
        test_duration = $TestDuration
        safe_mode = $SafeMode
        results = @{
            disaster_recovery = $script:DisasterRecoverySuccess
            load_testing = $script:LoadTestingSuccess
            feature_rollback = $script:FeatureRollbackSuccess
        }
        summary = @{
            total_tests = 3
            passed_tests = 0
            failed_tests = 0
        }
    }
    
    # Calculate summary
    if ($script:DisasterRecoverySuccess) { $reportData.summary.passed_tests++ }
    if ($script:LoadTestingSuccess) { $reportData.summary.passed_tests++ }
    if ($script:FeatureRollbackSuccess) { $reportData.summary.passed_tests++ }
    $reportData.summary.failed_tests = $reportData.summary.total_tests - $reportData.summary.passed_tests
    
    # Generate JSON report
    $jsonReport = $reportData | ConvertTo-Json -Depth 10
    $jsonReportPath = Join-Path $OutputDir "operational_readiness_report.json"
    $jsonReport | Set-Content $jsonReportPath
    
    # Generate HTML report
    $htmlReport = @"
<!DOCTYPE html>
<html>
<head>
    <title>Operational Readiness Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; border-bottom: 2px solid #007acc; padding-bottom: 20px; margin-bottom: 30px; }
        .summary { display: flex; justify-content: space-around; margin: 20px 0; }
        .summary-item { text-align: center; padding: 20px; border-radius: 8px; }
        .passed { background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .failed { background-color: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .info { background-color: #d1ecf1; color: #0c5460; border: 1px solid #bee5eb; }
        .test-result { margin: 20px 0; padding: 15px; border-radius: 8px; border-left: 4px solid; }
        .test-passed { background-color: #f8fff8; border-left-color: #28a745; }
        .test-failed { background-color: #fff8f8; border-left-color: #dc3545; }
        .details { margin-top: 10px; font-size: 14px; color: #666; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f8f9fa; font-weight: bold; }
        .timestamp { color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Operational Readiness Report</h1>
            <p class="timestamp">Generated: $($reportData.timestamp)</p>
            <p><strong>Environment:</strong> $($reportData.environment) | <strong>Test Type:</strong> $($reportData.test_type)</p>
        </div>
        
        <div class="summary">
            <div class="summary-item info">
                <h3>Total Tests</h3>
                <h2>$($reportData.summary.total_tests)</h2>
            </div>
            <div class="summary-item passed">
                <h3>Passed</h3>
                <h2>$($reportData.summary.passed_tests)</h2>
            </div>
            <div class="summary-item failed">
                <h3>Failed</h3>
                <h2>$($reportData.summary.failed_tests)</h2>
            </div>
        </div>
        
        <h2>Test Results</h2>
        
        <div class="test-result $(if ($reportData.results.disaster_recovery) { 'test-passed' } else { 'test-failed' })">
            <h3>Disaster Recovery Test</h3>
            <div class="details">
                <p><strong>Status:</strong> $(if ($reportData.results.disaster_recovery) { 'PASSED' } else { 'FAILED' })</p>
                <p><strong>Description:</strong> Backup and restore workflow for ZKID encrypted mappings, PQC key backup and rotation recovery, audit logs and admin session data recovery procedure.</p>
            </div>
        </div>
        
        <div class="test-result $(if ($reportData.results.load_testing) { 'test-passed' } else { 'test-failed' })">
            <h3>Load Testing</h3>
            <div class="details">
                <p><strong>Status:</strong> $(if ($reportData.results.load_testing) { 'PASSED' } else { 'FAILED' })</p>
                <p><strong>Concurrent Users:</strong> $($reportData.concurrent_users)</p>
                <p><strong>Test Duration:</strong> $($reportData.test_duration)</p>
                <p><strong>Description:</strong> Test all endpoints under simulated high traffic, measure API latency, error rates, and database response times.</p>
            </div>
        </div>
        
        <div class="test-result $(if ($reportData.results.feature_rollback) { 'test-passed' } else { 'test-failed' })">
            <h3>Feature Flag Rollback Test</h3>
            <div class="details">
                <p><strong>Status:</strong> $(if ($reportData.results.feature_rollback) { 'PASSED' } else { 'FAILED' })</p>
                <p><strong>Description:</strong> Simulate disabling ZKID/PQC features using environment flags, ensure system falls back safely without data loss.</p>
            </div>
        </div>
        
        <h2>Test Configuration</h2>
        <table>
            <tr>
                <th>Parameter</th>
                <th>Value</th>
            </tr>
            <tr>
                <td>Environment</td>
                <td>$($reportData.environment)</td>
            </tr>
            <tr>
                <td>Test Type</td>
                <td>$($reportData.test_type)</td>
            </tr>
            <tr>
                <td>Concurrent Users</td>
                <td>$($reportData.concurrent_users)</td>
            </tr>
            <tr>
                <td>Test Duration</td>
                <td>$($reportData.test_duration)</td>
            </tr>
            <tr>
                <td>Safe Mode</td>
                <td>$($reportData.safe_mode)</td>
            </tr>
        </table>
        
        <h2>Recommendations</h2>
        <ul>
            $(if ($reportData.summary.failed_tests -gt 0) {
                "<li><strong>Critical:</strong> Address failed tests before proceeding to production deployment.</li>"
            })
            $(if ($reportData.results.disaster_recovery) {
                "<li><strong>Disaster Recovery:</strong> Regularly test backup and restore procedures.</li>"
            })
            $(if ($reportData.results.load_testing) {
                "<li><strong>Load Testing:</strong> Monitor performance metrics in production and adjust thresholds as needed.</li>"
            })
            $(if ($reportData.results.feature_rollback) {
                "<li><strong>Feature Flags:</strong> Maintain comprehensive rollback procedures for all critical features.</li>"
            })
            "<li><strong>Monitoring:</strong> Implement continuous monitoring and alerting for all critical system components.</li>"
            "<li><strong>Documentation:</strong> Keep operational procedures up to date with any system changes.</li>"
        </ul>
    </div>
</body>
</html>
"@
    
    $htmlReportPath = Join-Path $OutputDir "operational_readiness_report.html"
    $htmlReport | Set-Content $htmlReportPath
    
    Write-ColorOutput "✓ Reports generated successfully:" -Color Success
    Write-ColorOutput "  JSON Report: $jsonReportPath" -Color Info
    Write-ColorOutput "  HTML Report: $htmlReportPath" -Color Info
}

function Show-Summary {
    Write-Header "Operational Readiness Test Summary"
    
    $totalTests = 3
    $passedTests = 0
    $failedTests = 0
    
    if ($script:DisasterRecoverySuccess) { $passedTests++ } else { $failedTests++ }
    if ($script:LoadTestingSuccess) { $passedTests++ } else { $failedTests++ }
    if ($script:FeatureRollbackSuccess) { $passedTests++ } else { $failedTests++ }
    
    Write-ColorOutput "Test Results:" -Color Info
    Write-ColorOutput "  Disaster Recovery: $(if ($script:DisasterRecoverySuccess) { 'PASSED' } else { 'FAILED' })" -Color $(if ($script:DisasterRecoverySuccess) { 'Success' } else { 'Error' })
    Write-ColorOutput "  Load Testing: $(if ($script:LoadTestingSuccess) { 'PASSED' } else { 'FAILED' })" -Color $(if ($script:LoadTestingSuccess) { 'Success' } else { 'Error' })
    Write-ColorOutput "  Feature Flag Rollback: $(if ($script:FeatureRollbackSuccess) { 'PASSED' } else { 'FAILED' })" -Color $(if ($script:FeatureRollbackSuccess) { 'Success' } else { 'Error' })
    
    Write-ColorOutput "`nSummary:" -Color Info
    Write-ColorOutput "  Total Tests: $totalTests" -Color Info
    Write-ColorOutput "  Passed: $passedTests" -Color Success
    Write-ColorOutput "  Failed: $failedTests" -Color $(if ($failedTests -gt 0) { 'Error' } else { 'Success' })
    
    if ($failedTests -eq 0) {
        Write-ColorOutput "`n🎉 All operational readiness tests passed! The system is ready for production deployment." -Color Success
    } else {
        Write-ColorOutput "`n⚠️  Some operational readiness tests failed. Please address the issues before proceeding to production." -Color Warning
    }
}

# Main execution
try {
    Write-Header "Secure Email MVP - Operational Readiness Testing"
    Write-ColorOutput "Iteration 3: System Resiliency, Disaster Recovery, and Production Readiness" -Color Info
    
    # Initialize test result variables
    $script:DisasterRecoverySuccess = $false
    $script:LoadTestingSuccess = $false
    $script:FeatureRollbackSuccess = $false
    
    # Check prerequisites
    Test-Prerequisites
    
    # Validate environment
    Test-Environment
    
    # Run tests based on test type
    switch ($TestType.ToLower()) {
        "all" {
            $script:DisasterRecoverySuccess = Start-DisasterRecoveryTest
            $script:LoadTestingSuccess = Start-LoadTesting
            $script:FeatureRollbackSuccess = Start-FeatureFlagRollbackTest
        }
        "disaster-recovery" {
            $script:DisasterRecoverySuccess = Start-DisasterRecoveryTest
        }
        "load-testing" {
            $script:LoadTestingSuccess = Start-LoadTesting
        }
        "feature-rollback" {
            $script:FeatureRollbackSuccess = Start-FeatureFlagRollbackTest
        }
        default {
            Write-ColorOutput "Invalid test type: $TestType" -Color Error
            Write-ColorOutput "Valid options: all, disaster-recovery, load-testing, feature-rollback" -Color Error
            exit 1
        }
    }
    
    # Generate reports
    Generate-OperationalReadinessReport
    
    # Show summary
    Show-Summary
    
} catch {
    Write-ColorOutput "`n❌ Operational readiness testing failed: $($_.Exception.Message)" -Color Error
    Write-ColorOutput "Stack trace: $($_.ScriptStackTrace)" -Color Error
    exit 1
}

Write-ColorOutput "`nOperational readiness testing completed!" -Color Success
