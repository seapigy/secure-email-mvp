# Sprint 8 Simple Validation Script
param([string]$TestMode = "all")

Write-Output "=== Sprint 8 Validation: Production Deployment & Enterprise Integration ==="
Write-Output "Timestamp: $(Get-Date)"
Write-Output ""

$PassedTests = 0
$FailedTests = 0

function Test-Component {
    param($FilePath, $ComponentName)

    if (Test-Path $FilePath) {
        Write-Output "✅ $ComponentName - File exists"
        $global:PassedTests++
        return $true
    } else {
        Write-Output "❌ $ComponentName - File missing"
        $global:FailedTests++
        return $false
    }
}

function Test-Compilation {
    Write-Output "`nTesting compilation..."
    try {
        $output = go build ./pkg/e2e/... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Output "✅ Compilation successful"
            $global:PassedTests++
            return $true
        } else {
            Write-Output "❌ Compilation failed: $output"
            $global:FailedTests++
            return $false
        }
    } catch {
        Write-Output "❌ Compilation error: $_"
        $global:FailedTests++
        return $false
    }
}

Write-Output "📋 Testing Sprint 8 Core Components..."
Write-Output ""

# Component tests
Test-Component "docs/sprint8_design.md" "Sprint 8 Design Document"
Test-Component "pkg/e2e/deployment_automation.go" "Deployment Automation"
Test-Component "pkg/e2e/enterprise_apis.go" "Enterprise APIs"
Test-Component "pkg/e2e/scaling_infrastructure.go" "Scaling Infrastructure"

# Compilation test
Test-Compilation

Write-Output ""
Write-Output "📊 Validation Results"
Write-Output "====================="
Write-Output "✅ Passed: $PassedTests"
Write-Output "❌ Failed: $FailedTests"
Write-Output "📋 Total: $($PassedTests + $FailedTests)"

$PassRate = if ($PassedTests + $FailedTests -gt 0) {
    [math]::Round(($PassedTests / ($PassedTests + $FailedTests)) * 100, 2)
} else {
    0
}
Write-Output "📈 Pass Rate: $PassRate%"

Write-Output ""
if ($FailedTests -eq 0) {
    Write-Output "🎉 Sprint 8 validation passed! All components are ready."
    exit 0
} else {
    Write-Output "⚠️ Some validations failed. Please review the issues above."
    exit 1
}
