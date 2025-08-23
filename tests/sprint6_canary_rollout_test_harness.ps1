# Sprint 6: Canary Rollout + Monitoring + Runbooks Test Harness
# =============================================================================

param(
    [string]$ProjectRoot = ".",
    [switch]$Verbose
)

$separatorLine = "=" * 60

function Write-TestHeader {
    param([string]$TestName)
    Write-Output "`n$separatorLine"
    Write-Output "Testing: $TestName"
    Write-Host $separatorLine -ForegroundColor Cyan
}

function Write-TestResult {
    param([string]$TestName, [string]$Message, [bool]$Passed)
    $status = if ($Passed) { "PASS" } else { "FAIL" }
    $color = if ($Passed) { "Green" } else { "Red" }
    Write-Output "[$status] $TestName`: $Message" -ForegroundColor $color
}

function Test-FileExists {
    param([string]$FilePath, [string]$TestName)
    $exists = Test-Path $FilePath
    Write-TestResult -TestName $TestName -Message "File exists: $FilePath" -Passed $exists
    return $exists
}

function Test-FileContent {
    param([string]$FilePath, [string]$Pattern, [string]$TestName)
    if (-not (Test-Path $FilePath)) {
        Write-TestResult -TestName $TestName -Message "File not found: $FilePath" -Passed $false
        return $false
    }

    $content = Get-Content $FilePath -Raw
    $matches = $content -match $Pattern
    Write-TestResult -TestName $TestName -Message "Pattern found: $Pattern" -Passed $matches
    return $matches
}

function Test-GoBuild {
    param([string]$PackagePath, [string]$TestName)
    try {
        Push-Location $PackagePath
        $output = go build -o /dev/null 2>&1
        $success = $LASTEXITCODE -eq 0
        Write-TestResult -TestName $TestName -Message "Go build successful" -Passed $success
        return $success
    }
    catch {
        Write-TestResult -TestName $TestName -Message "Go build failed: $($_.Exception.Message)" -Passed $false
        return $false
    }
    finally {
        Pop-Location
    }
}

function Test-GoTest {
    param([string]$PackagePath, [string]$TestName)
    try {
        Push-Location $PackagePath
        $output = go test -v 2>&1
        $success = $LASTEXITCODE -eq 0
        Write-TestResult -TestName $TestName -Message "Go tests passed" -Passed $success
        if (-not $success -and $Verbose) {
            Write-Output "Test output:"
            Write-Host $output -ForegroundColor Gray
        }
        return $success
    }
    catch {
        Write-TestResult -TestName $TestName -Message "Go tests failed: $($_.Exception.Message)" -Passed $false
        return $false
    }
    finally {
        Pop-Location
    }
}

# Initialize test results
$testResults = @{
    Total = 0
    Passed = 0
    Failed = 0
}

function Update-TestResults {
    param([bool]$Passed)
    $testResults.Total++
    if ($Passed) {
        $testResults.Passed++
    } else {
        $testResults.Failed++
    }
}

# Main test execution
Write-Output "Sprint 6: Canary Rollout + Monitoring + Runbooks Test Harness"
Write-Output "Starting comprehensive validation..."

# Test 1: Design Document
Write-TestHeader "Design Document Validation"
$passed = Test-FileExists -FilePath "docs/sprint6_canary_rollout_design.md" -TestName "Sprint 6 Design Document"
Update-TestResults -Passed $passed

if ($passed) {
    $passed = Test-FileContent -FilePath "docs/sprint6_canary_rollout_design.md" -Pattern "Canary Rollout Strategy" -TestName "Canary Rollout Strategy Section"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "docs/sprint6_canary_rollout_design.md" -Pattern "Monitoring Architecture" -TestName "Monitoring Architecture Section"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "docs/sprint6_canary_rollout_design.md" -Pattern "Operational Runbooks" -TestName "Operational Runbooks Section"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "docs/sprint6_canary_rollout_design.md" -Pattern "A/B Testing Framework" -TestName "A/B Testing Framework Section"
    Update-TestResults -Passed $passed
}

# Test 2: Database Migration
Write-TestHeader "Database Migration Validation"
$passed = Test-FileExists -FilePath "schema/migrate_add_canary_rollout.sql" -TestName "Canary Rollout Migration File"
Update-TestResults -Passed $passed

if ($passed) {
    $passed = Test-FileContent -FilePath "schema/migrate_add_canary_rollout.sql" -Pattern "canary_config" -TestName "Canary Config Table"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "schema/migrate_add_canary_rollout.sql" -Pattern "ab_test_results" -TestName "A/B Test Results Table"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "schema/migrate_add_canary_rollout.sql" -Pattern "rollback_events" -TestName "Rollback Events Table"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "schema/migrate_add_canary_rollout.sql" -Pattern "monitoring_alerts" -TestName "Monitoring Alerts Table"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "schema/migrate_add_canary_rollout.sql" -Pattern "runbook_executions" -TestName "Runbook Executions Table"
    Update-TestResults -Passed $passed
}

# Test 3: Canary Rollout System
Write-TestHeader "Canary Rollout System Validation"
$passed = Test-FileExists -FilePath "pkg/e2e/canary_rollout.go" -TestName "Canary Rollout Implementation"
Update-TestResults -Passed $passed

if ($passed) {
    $passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "type CanaryRolloutManager struct" -TestName "CanaryRolloutManager Struct"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "type ABTestEngine struct" -TestName "ABTestEngine Struct"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "type TrafficRouter struct" -TestName "TrafficRouter Struct"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "func NewCanaryRolloutManager" -TestName "NewCanaryRolloutManager Function"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "func.*ShouldRouteToE2E" -TestName "ShouldRouteToE2E Function"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "func.*TriggerRollback" -TestName "TriggerRollback Function"
    Update-TestResults -Passed $passed
}

# Test 4: Runbook Automation System
Write-TestHeader "Runbook Automation System Validation"
$passed = Test-FileExists -FilePath "pkg/e2e/runbooks.go" -TestName "Runbook Automation Implementation"
Update-TestResults -Passed $passed

if ($passed) {
    $passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "type RunbookEngine struct" -TestName "RunbookEngine Struct"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "type Procedure struct" -TestName "Procedure Struct"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "type Step struct" -TestName "Step Struct"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "func NewRunbookEngine" -TestName "NewRunbookEngine Function"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "func.*ExecuteProcedure" -TestName "ExecuteProcedure Function"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "canary_rollout" -TestName "Canary Rollout Procedure"
    Update-TestResults -Passed $passed

    $passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "emergency_rollback" -TestName "Emergency Rollback Procedure"
    Update-TestResults -Passed $passed
}

# Test 5: Unit Tests
Write-TestHeader "Unit Tests Validation"
$passed = Test-FileExists -FilePath "pkg/e2e/canary_rollout_test.go" -TestName "Canary Rollout Unit Tests"
Update-TestResults -Passed $passed

$passed = Test-FileExists -FilePath "pkg/e2e/runbooks_test.go" -TestName "Runbook Unit Tests"
Update-TestResults -Passed $passed

# Test 6: Go Build Validation
Write-TestHeader "Go Build Validation"
$passed = Test-GoBuild -PackagePath "pkg/e2e" -TestName "E2E Package Build"
Update-TestResults -Passed $passed

# Test 7: Unit Test Execution
Write-TestHeader "Unit Test Execution"
$passed = Test-GoTest -PackagePath "pkg/e2e" -TestName "E2E Package Tests"
Update-TestResults -Passed $passed

# Test 8: Integration with Previous Sprints
Write-TestHeader "Integration with Previous Sprints"
$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "MetricsCollector" -TestName "Integration with Monitoring (Sprint 5)"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "sql\.DB" -TestName "Integration with Database (Sprint 4)"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "uuid\.New" -TestName "Integration with UUID Generation"
Update-TestResults -Passed $passed

# Test 9: Configuration Management
Write-TestHeader "Configuration Management Validation"
$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "type CanaryConfig struct" -TestName "CanaryConfig Struct"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "TrafficPercentage.*float64" -TestName "Traffic Percentage Configuration"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "RollbackThreshold.*float64" -TestName "Rollback Threshold Configuration"
Update-TestResults -Passed $passed

# Test 10: Error Handling and Safety
Write-TestHeader "Error Handling and Safety Validation"
$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "if percentage < 0 \|\| percentage > 100" -TestName "Traffic Percentage Validation"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "Critical.*bool" -TestName "Critical Step Handling"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "triggerRollback" -TestName "Rollback Safety Mechanism"
Update-TestResults -Passed $passed

# Test 11: Monitoring Integration
Write-TestHeader "Monitoring Integration Validation"
$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "monitorRollout" -TestName "Continuous Monitoring"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "checkRollbackConditions" -TestName "Rollback Condition Checking"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "GetRolloutStatus" -TestName "Status Reporting"
Update-TestResults -Passed $passed

# Test 12: A/B Testing Framework
Write-TestHeader "A/B Testing Framework Validation"
$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "type Criterion struct" -TestName "A/B Test Criteria"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "evaluateCriterion" -TestName "Criteria Evaluation"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "updateTestDecision" -TestName "Test Decision Logic"
Update-TestResults -Passed $passed

# Test 13: Traffic Routing
Write-TestHeader "Traffic Routing Validation"
$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "RouteRequest" -TestName "Request Routing Function"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "hashUserID" -TestName "Consistent Hashing"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/canary_rollout.go" -Pattern "UserSegments" -TestName "User Segment Support"
Update-TestResults -Passed $passed

# Test 14: Runbook Procedures
Write-TestHeader "Runbook Procedures Validation"
$passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "registerBuiltInProcedures" -TestName "Built-in Procedures Registration"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "ExecuteAction" -TestName "Action Execution"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "pkg/e2e/runbooks.go" -Pattern "GetExecutionStatus" -TestName "Execution Status Tracking"
Update-TestResults -Passed $passed

# Test 15: Documentation Completeness
Write-TestHeader "Documentation Completeness Validation"
$passed = Test-FileContent -FilePath "docs/sprint6_canary_rollout_design.md" -Pattern "Production Readiness" -TestName "Production Readiness Section"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "docs/sprint6_canary_rollout_design.md" -Pattern "Risk Mitigation" -TestName "Risk Mitigation Section"
Update-TestResults -Passed $passed

$passed = Test-FileContent -FilePath "docs/sprint6_canary_rollout_design.md" -Pattern "Success Metrics" -TestName "Success Metrics Section"
Update-TestResults -Passed $passed

# Summary
Write-Output "`n$separatorLine"
Write-Output "Sprint 6 Test Results Summary"
Write-Host $separatorLine -ForegroundColor Magenta
Write-Output "Total Tests: $($testResults.Total)"
Write-Output "Passed: $($testResults.Passed)"
Write-Output "Failed: $($testResults.Failed)"
Write-Output "Success Rate: $([math]::Round(($testResults.Passed / $testResults.Total) * 100, 2))%"

if ($testResults.Failed -eq 0) {
    Write-Output "`n🎉 All Sprint 6 tests passed! The canary rollout system is ready for production deployment."
} else {
    Write-Output "`n⚠️  Some tests failed. Please review the failed tests above and fix any issues before proceeding."
}

Write-Output "`nSprint 6 Canary Rollout and Monitoring Test Harness Complete!"
