# Comprehensive Test Runner for Secure Email MVP
# This script runs all tests and generates detailed reports

param(
    [switch]$Verbose,
    [switch]$Coverage,
    [switch]$Unit,
    [switch]$Integration,
    [switch]$E2E,
    [switch]$All
)

# Set error action preference
$ErrorActionPreference = "Continue"

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"
$White = "White"

# Test results tracking
$TestResults = @{
    Unit = @{ Passed = 0; Failed = 0; Total = 0 }
    Integration = @{ Passed = 0; Failed = 0; Total = 0 }
    E2E = @{ Passed = 0; Failed = 0; Total = 0 }
    Overall = @{ Passed = 0; Failed = 0; Total = 0 }
}

# Function to write colored output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = $White
    )
    Write-Host $Message -ForegroundColor $Color
}

# Function to run tests and capture results
function Run-TestSuite {
    param(
        [string]$TestType,
        [string]$TestPath,
        [string]$TestName
    )
    
    Write-ColorOutput "`n=== Running $TestName Tests ===" $Blue
    
    $startTime = Get-Date
    $output = & go test $TestPath -v 2>&1
    $endTime = Get-Date
    $duration = $endTime - $startTime
    
    # Parse test results
    $passed = ($output | Select-String "PASS:").Count
    $failed = ($output | Select-String "FAIL:").Count
    $total = $passed + $failed
    
    # Update test results
    $TestResults[$TestType].Passed += $passed
    $TestResults[$TestType].Failed += $failed
    $TestResults[$TestType].Total += $total
    $TestResults.Overall.Passed += $passed
    $TestResults.Overall.Failed += $failed
    $TestResults.Overall.Total += $total
    
    # Display results
    Write-ColorOutput "Duration: $($duration.TotalSeconds.ToString('F2')) seconds" $White
    Write-ColorOutput "Passed: $passed" $Green
    Write-ColorOutput "Failed: $failed" $(if ($failed -gt 0) { $Red } else { $Green })
    Write-ColorOutput "Total: $total" $White
    
    if ($Verbose) {
        Write-ColorOutput "`nDetailed Output:" $Yellow
        $output | ForEach-Object { Write-Host $_ }
    }
    
    return @{
        Passed = $passed
        Failed = $failed
        Total = $total
        Duration = $duration
        Output = $output
    }
}

# Function to generate coverage report
function Get-CoverageReport {
    Write-ColorOutput "`n=== Generating Coverage Report ===" $Blue
    
    $coverageOutput = & go test ./... -cover 2>&1
    $coverageOutput | ForEach-Object { Write-Host $_ }
    
    # Extract coverage percentages
    $coverageLines = $coverageOutput | Select-String "coverage:"
    $totalCoverage = 0
    $packageCount = 0
    
    foreach ($line in $coverageLines) {
        if ($line -match "coverage: (\d+\.\d+)%") {
            $coverage = [double]$matches[1]
            $totalCoverage += $coverage
            $packageCount++
        }
    }
    
    if ($packageCount -gt 0) {
        $averageCoverage = $totalCoverage / $packageCount
        Write-ColorOutput "`nAverage Coverage: $($averageCoverage.ToString('F2'))%" $(if ($averageCoverage -ge 80) { $Green } else { $Yellow })
    }
    
    return $coverageOutput
}

# Function to generate test summary
function Show-TestSummary {
    Write-ColorOutput "`n=== Test Summary ===" $Blue
    
    $summary = @"
┌─────────────────┬─────────┬─────────┬─────────┬─────────────┐
│ Test Type       │ Passed  │ Failed  │ Total   │ Success Rate │
├─────────────────┼─────────┼─────────┼─────────┼─────────────┤
"@
    
    foreach ($testType in @("Unit", "Integration", "E2E")) {
        $passed = $TestResults[$testType].Passed
        $failed = $TestResults[$testType].Failed
        $total = $TestResults[$testType].Total
        $successRate = if ($total -gt 0) { [math]::Round(($passed / $total) * 100, 1) } else { 0 }
        
        $summary += "`n│ $($testType.PadRight(15)) │ $($passed.ToString().PadLeft(7)) │ $($failed.ToString().PadLeft(7)) │ $($total.ToString().PadLeft(7)) │ $($successRate.ToString().PadLeft(10))% │"
    }
    
    $overallPassed = $TestResults.Overall.Passed
    $overallFailed = $TestResults.Overall.Failed
    $overallTotal = $TestResults.Overall.Total
    $overallSuccessRate = if ($overallTotal -gt 0) { [math]::Round(($overallPassed / $overallTotal) * 100, 1) } else { 0 }
    
    $summary += "`n├─────────────────┼─────────┼─────────┼─────────┼─────────────┤"
    $summary += "`n│ Overall         │ $($overallPassed.ToString().PadLeft(7)) │ $($overallFailed.ToString().PadLeft(7)) │ $($overallTotal.ToString().PadLeft(7)) │ $($overallSuccessRate.ToString().PadLeft(10))% │"
    $summary += "`n└─────────────────┴─────────┴─────────┴─────────┴─────────────┘"
    
    Write-ColorOutput $summary $White
    
    # Overall status
    if ($overallFailed -eq 0) {
        Write-ColorOutput "`n✅ All tests passed successfully!" $Green
    } else {
        Write-ColorOutput "`n❌ Some tests failed. Please review the output above." $Red
    }
}

# Function to save test results to file
function Save-TestResults {
    param([string]$Results)
    
    $timestamp = Get-Date -Format "yyyy-MM-dd_HH-mm-ss"
    $resultsFile = "tests/results_$timestamp.txt"
    
    try {
        $Results | Out-File -FilePath $resultsFile -Encoding UTF8
        Write-ColorOutput "`nTest results saved to: $resultsFile" $Green
    } catch {
        Write-ColorOutput "Failed to save test results: $($_.Exception.Message)" $Red
    }
}

# Main execution
Write-ColorOutput "🚀 Secure Email MVP - Comprehensive Test Suite" $Blue
Write-ColorOutput "===============================================" $Blue

# Determine which tests to run
$runUnit = $Unit -or $All -or (-not ($Unit -or $Integration -or $E2E))
$runIntegration = $Integration -or $All
$runE2E = $E2E -or $All

$allResults = @()

# Run unit tests
if ($runUnit) {
    $unitResults = Run-TestSuite -TestType "Unit" -TestPath "./tests/unit" -TestName "Unit"
    $allResults += "=== Unit Tests ==="
    $allResults += $unitResults.Output
}

# Run integration tests
if ($runIntegration) {
    $integrationResults = Run-TestSuite -TestType "Integration" -TestPath "./tests/integration" -TestName "Integration"
    $allResults += "=== Integration Tests ==="
    $allResults += $integrationResults.Output
}

# Run E2E tests
if ($runE2E) {
    $e2eResults = Run-TestSuite -TestType "E2E" -TestPath "./tests/e2e" -TestName "End-to-End"
    $allResults += "=== E2E Tests ==="
    $allResults += $e2eResults.Output
}

# Generate coverage report if requested
if ($Coverage) {
    $coverageResults = Get-CoverageReport
    $allResults += "=== Coverage Report ==="
    $allResults += $coverageResults
}

# Show test summary
Show-TestSummary

# Save results
Save-TestResults -Results $allResults

# Exit with appropriate code
if ($TestResults.Overall.Failed -gt 0) {
    Write-ColorOutput "`nExiting with code 1 due to test failures" $Red
    exit 1
} else {
    Write-ColorOutput "`nAll tests completed successfully!" $Green
    exit 0
}

