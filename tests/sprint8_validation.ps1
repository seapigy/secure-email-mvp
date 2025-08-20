# Sprint 8 Simple Validation Script
param([string]$TestMode = "all")

Write-Host "=== Sprint 8 Validation: Production Deployment & Enterprise Integration ===" -ForegroundColor Cyan
Write-Host "Timestamp: $(Get-Date)" -ForegroundColor Green
Write-Host ""

$PassedTests = 0
$FailedTests = 0

function Test-Component {
    param($FilePath, $ComponentName)
    
    if (Test-Path $FilePath) {
        Write-Host "✅ $ComponentName - File exists" -ForegroundColor Green
        $global:PassedTests++
        return $true
    } else {
        Write-Host "❌ $ComponentName - File missing" -ForegroundColor Red
        $global:FailedTests++
        return $false
    }
}

function Test-Compilation {
    Write-Host "`nTesting compilation..." -ForegroundColor Yellow
    try {
        $output = go build ./pkg/e2e/... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Compilation successful" -ForegroundColor Green
            $global:PassedTests++
            return $true
        } else {
            Write-Host "❌ Compilation failed: $output" -ForegroundColor Red
            $global:FailedTests++
            return $false
        }
    } catch {
        Write-Host "❌ Compilation error: $_" -ForegroundColor Red
        $global:FailedTests++
        return $false
    }
}

Write-Host "📋 Testing Sprint 8 Core Components..." -ForegroundColor Blue
Write-Host ""

# Component tests
Test-Component "docs/sprint8_design.md" "Sprint 8 Design Document"
Test-Component "pkg/e2e/deployment_automation.go" "Deployment Automation"
Test-Component "pkg/e2e/enterprise_apis.go" "Enterprise APIs"
Test-Component "pkg/e2e/scaling_infrastructure.go" "Scaling Infrastructure"

# Compilation test
Test-Compilation

Write-Host ""
Write-Host "📊 Validation Results" -ForegroundColor Cyan
Write-Host "=====================" -ForegroundColor Cyan
Write-Host "✅ Passed: $PassedTests" -ForegroundColor Green
Write-Host "❌ Failed: $FailedTests" -ForegroundColor Red
Write-Host "📋 Total: $($PassedTests + $FailedTests)" -ForegroundColor Blue

$PassRate = if ($PassedTests + $FailedTests -gt 0) { 
    [math]::Round(($PassedTests / ($PassedTests + $FailedTests)) * 100, 2) 
} else { 
    0 
}
Write-Host "📈 Pass Rate: $PassRate%" -ForegroundColor Yellow

Write-Host ""
if ($FailedTests -eq 0) {
    Write-Host "🎉 Sprint 8 validation passed! All components are ready." -ForegroundColor Green
    exit 0
} else {
    Write-Host "⚠️ Some validations failed. Please review the issues above." -ForegroundColor Yellow
    exit 1
}
