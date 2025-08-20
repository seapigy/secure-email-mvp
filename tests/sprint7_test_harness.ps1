# Sprint 7 Test Harness: Advanced PQC Features & Production Hardening
# Tests mixnet routing, hardware-backed keys, observability, and compliance automation

param(
    [string]$TestMode = "all",  # "all", "mixnet", "hardware", "observability", "compliance"
    [string]$Verbose = "false"
)

Write-Host "=== Sprint 7 Test Harness: Advanced PQC Features & Production Hardening ===" -ForegroundColor Cyan
Write-Host "Test Mode: $TestMode" -ForegroundColor Green
Write-Host "Timestamp: $(Get-Date)" -ForegroundColor Green
Write-Host ""

$ErrorActionPreference = "Continue"
$TestResults = @()
$PassedTests = 0
$FailedTests = 0

function Write-TestResult {
    param($TestName, $Result, $Details = "")
    
    $global:TestResults += [PSCustomObject]@{
        Test = $TestName
        Result = $Result
        Details = $Details
        Timestamp = Get-Date
    }
    
    if ($Result -eq "PASS") {
        Write-Host "✅ $TestName" -ForegroundColor Green
        $global:PassedTests++
    } else {
        Write-Host "❌ $TestName" -ForegroundColor Red
        if ($Details) {
            Write-Host "   Details: $Details" -ForegroundColor Yellow
        }
        $global:FailedTests++
    }
    
    if ($Verbose -eq "true" -and $Details) {
        Write-Host "   $Details" -ForegroundColor Gray
    }
}

function Test-FileExists {
    param($FilePath, $TestName)
    
    if (Test-Path $FilePath) {
        Write-TestResult $TestName "PASS" "File exists: $FilePath"
        return $true
    } else {
        Write-TestResult $TestName "FAIL" "File missing: $FilePath"
        return $false
    }
}

function Test-GoSyntax {
    param($FilePath, $TestName)
    
    try {
        $output = go run -o nul $FilePath 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-TestResult $TestName "PASS" "Go syntax valid"
            return $true
        } else {
            Write-TestResult $TestName "FAIL" "Go syntax error: $output"
            return $false
        }
    } catch {
        Write-TestResult $TestName "FAIL" "Failed to check syntax: $_"
        return $false
    }
}

function Test-StructDefinition {
    param($FilePath, $StructName, $TestName)
    
    try {
        $content = Get-Content $FilePath -Raw
        if ($content -match "type\s+$StructName\s+struct") {
            Write-TestResult $TestName "PASS" "Struct $StructName defined"
            return $true
        } else {
            Write-TestResult $TestName "FAIL" "Struct $StructName not found"
            return $false
        }
    } catch {
        Write-TestResult $TestName "FAIL" "Failed to read file: $_"
        return $false
    }
}

function Test-FunctionDefinition {
    param($FilePath, $FunctionName, $TestName)
    
    try {
        $content = Get-Content $FilePath -Raw
        if ($content -match "func\s+.*$FunctionName\s*\(") {
            Write-TestResult $TestName "PASS" "Function/Method $FunctionName defined"
            return $true
        } else {
            Write-TestResult $TestName "FAIL" "Function/Method $FunctionName not found"
            return $false
        }
    } catch {
        Write-TestResult $TestName "FAIL" "Failed to read file: $_"
        return $false
    }
}

function Test-InterfaceDefinition {
    param($FilePath, $InterfaceName, $TestName)
    
    try {
        $content = Get-Content $FilePath -Raw
        if ($content -match "type\s+$InterfaceName\s+interface") {
            Write-TestResult $TestName "PASS" "Interface $InterfaceName defined"
            return $true
        } else {
            Write-TestResult $TestName "FAIL" "Interface $InterfaceName not found"
            return $false
        }
    } catch {
        Write-TestResult $TestName "FAIL" "Failed to read file: $_"
        return $false
    }
}

function Test-ConstantDefinition {
    param($FilePath, $ConstantPattern, $TestName)
    
    try {
        $content = Get-Content $FilePath -Raw
        if ($content -match $ConstantPattern) {
            Write-TestResult $TestName "PASS" "Constants defined"
            return $true
        } else {
            Write-TestResult $TestName "FAIL" "Constants not found"
            return $false
        }
    } catch {
        Write-TestResult $TestName "FAIL" "Failed to read file: $_"
        return $false
    }
}

function Test-ImportStatement {
    param($FilePath, $ImportPath, $TestName)
    
    try {
        $content = Get-Content $FilePath -Raw
        if ($content -match "import.*`"$ImportPath`"" -or $content -match "_\s+`"$ImportPath`"" -or $content -match "import\s+`"$ImportPath`"" -or $content -match "`"$ImportPath`"") {
            Write-TestResult $TestName "PASS" "Import $ImportPath found"
            return $true
        } else {
            Write-TestResult $TestName "FAIL" "Import $ImportPath not found"
            return $false
        }
    } catch {
        Write-TestResult $TestName "FAIL" "Failed to read file: $_"
        return $false
    }
}

function Test-GoCompilation {
    param($Directory, $TestName)
    
    try {
        Push-Location $Directory
        $output = go build ./... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-TestResult $TestName "PASS" "Compilation successful"
            $result = $true
        } else {
            Write-TestResult $TestName "FAIL" "Compilation failed: $output"
            $result = $false
        }
        Pop-Location
        return $result
    } catch {
        Pop-Location
        Write-TestResult $TestName "FAIL" "Failed to compile: $_"
        return $false
    }
}

function Test-DesignDocument {
    param($FilePath, $TestName)
    
    try {
        if (Test-Path $FilePath) {
            $content = Get-Content $FilePath -Raw
            $requiredSections = @(
                "# Sprint 7",
                "## Architecture",
                "## Component Details",
                "### 1. Mixnet Routing System",
                "### 2. Hardware-Backed Key Management",
                "### 3. Advanced Observability",
                "### 4. Automated Compliance"
            )
            
            $missingCount = 0
            foreach ($section in $requiredSections) {
                if ($content -notmatch [regex]::Escape($section)) {
                    $missingCount++
                }
            }
            
            if ($missingCount -eq 0) {
                Write-TestResult $TestName "PASS" "All required sections present"
                return $true
            } else {
                Write-TestResult $TestName "FAIL" "$missingCount required sections missing"
                return $false
            }
        } else {
            Write-TestResult $TestName "FAIL" "Design document not found"
            return $false
        }
    } catch {
        Write-TestResult $TestName "FAIL" "Failed to validate design document: $_"
        return $false
    }
}

# Sprint 7 Core Tests
Write-Host "📋 Testing Sprint 7 Core Components..." -ForegroundColor Blue

# Design Document Tests
Write-Host "`n🔍 Design Document Validation..." -ForegroundColor Yellow
Test-DesignDocument "docs/sprint7_design.md" "Sprint 7 Design Document"

# Mixnet Routing System Tests
if ($TestMode -eq "all" -or $TestMode -eq "mixnet") {
    Write-Host "`n🌐 Mixnet Routing System Tests..." -ForegroundColor Yellow
    
    # File existence tests
    Test-FileExists "pkg/e2e/mixnet.go" "Mixnet Core File"
    Test-FileExists "pkg/e2e/mixnet_directory.go" "Mixnet Directory File"
    
    # Core structure tests
    Test-StructDefinition "pkg/e2e/mixnet.go" "MixnetRouter" "MixnetRouter Struct"
    Test-StructDefinition "pkg/e2e/mixnet.go" "MixnetConfig" "MixnetConfig Struct"
    Test-StructDefinition "pkg/e2e/mixnet.go" "MixnetNode" "MixnetNode Struct"
    Test-StructDefinition "pkg/e2e/mixnet.go" "MixnetRoute" "MixnetRoute Struct"
    Test-StructDefinition "pkg/e2e/mixnet.go" "MixnetMessage" "MixnetMessage Struct"
    Test-StructDefinition "pkg/e2e/mixnet.go" "OnionLayer" "OnionLayer Struct"
    
    # Function tests
    Test-FunctionDefinition "pkg/e2e/mixnet.go" "NewMixnetRouter" "NewMixnetRouter Function"
    Test-FunctionDefinition "pkg/e2e/mixnet.go" "RouteMessage" "RouteMessage Method"
    Test-FunctionDefinition "pkg/e2e/mixnet.go" "ProcessMessage" "ProcessMessage Method"
    Test-FunctionDefinition "pkg/e2e/mixnet.go" "GenerateCoverTraffic" "GenerateCoverTraffic Method"
    
    # Node Directory tests
    Test-StructDefinition "pkg/e2e/mixnet_directory.go" "NodeDirectory" "NodeDirectory Struct"
    Test-StructDefinition "pkg/e2e/mixnet_directory.go" "PathSelector" "PathSelector Struct"
    Test-StructDefinition "pkg/e2e/mixnet_directory.go" "MixnetCoverTrafficGenerator" "MixnetCoverTrafficGenerator Struct"
    
    Test-FunctionDefinition "pkg/e2e/mixnet_directory.go" "NewNodeDirectory" "NewNodeDirectory Function"
    Test-FunctionDefinition "pkg/e2e/mixnet_directory.go" "NewPathSelector" "NewPathSelector Function"
    Test-FunctionDefinition "pkg/e2e/mixnet_directory.go" "RefreshNodes" "RefreshNodes Method"
    Test-FunctionDefinition "pkg/e2e/mixnet_directory.go" "SelectPath" "SelectPath Method"
    
    # Import tests
    Test-ImportStatement "pkg/e2e/mixnet.go" "modernc.org/sqlite" "SQLite Import"
    Test-ImportStatement "pkg/e2e/mixnet.go" "context" "Context Import"
}

# Hardware-Backed Key Management Tests
if ($TestMode -eq "all" -or $TestMode -eq "hardware") {
    Write-Host "`n🔐 Hardware-Backed Key Management Tests..." -ForegroundColor Yellow
    
    # File existence tests
    Test-FileExists "pkg/e2e/hardware_security.go" "Hardware Security File"
    
    # Core structure tests
    Test-StructDefinition "pkg/e2e/hardware_security.go" "HardwareSecurityManager" "HardwareSecurityManager Struct"
    Test-StructDefinition "pkg/e2e/hardware_security.go" "HardwareSecurityConfig" "HardwareSecurityConfig Struct"
    Test-StructDefinition "pkg/e2e/hardware_security.go" "HardwareKey" "HardwareKey Struct"
    Test-StructDefinition "pkg/e2e/hardware_security.go" "AttestationReport" "AttestationReport Struct"
    Test-StructDefinition "pkg/e2e/hardware_security.go" "SecurityState" "SecurityState Struct"
    
    # Interface tests
    Test-InterfaceDefinition "pkg/e2e/hardware_security.go" "HardwareSecurityProvider" "HardwareSecurityProvider Interface"
    
    # Function tests
    Test-FunctionDefinition "pkg/e2e/hardware_security.go" "NewHardwareSecurityManager" "NewHardwareSecurityManager Function"
    Test-FunctionDefinition "pkg/e2e/hardware_security.go" "GenerateKey" "GenerateKey Function"
    Test-FunctionDefinition "pkg/e2e/hardware_security.go" "AttestKey" "AttestKey Function"
    Test-FunctionDefinition "pkg/e2e/hardware_security.go" "Sign" "Sign Function"
    
    # Platform-specific tests
    Test-StructDefinition "pkg/e2e/hardware_security.go" "MockTPMProvider" "MockTPMProvider Struct"
    Test-StructDefinition "pkg/e2e/hardware_security.go" "MockSecureEnclaveProvider" "MockSecureEnclaveProvider Struct"
    Test-StructDefinition "pkg/e2e/hardware_security.go" "MockPKCS11Provider" "MockPKCS11Provider Struct"
    
    # Constant tests
    Test-ConstantDefinition "pkg/e2e/hardware_security.go" "type\s+KeyProtectionMode\s+struct" "KeyProtectionMode Type"
}

# Advanced Observability Tests
if ($TestMode -eq "all" -or $TestMode -eq "observability") {
    Write-Host "`n📊 Advanced Observability Tests..." -ForegroundColor Yellow
    
    # File existence tests
    Test-FileExists "pkg/e2e/observability.go" "Observability File"
    
    # Core structure tests
    Test-StructDefinition "pkg/e2e/observability.go" "DistributedTracer" "DistributedTracer Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "AdvancedAlertManager" "AdvancedAlertManager Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "AnomalyDetector" "AnomalyDetector Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "AdvancedObservabilityConfig" "AdvancedObservabilityConfig Struct"
    
    # Tracing tests
    Test-StructDefinition "pkg/e2e/observability.go" "Span" "Span Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "Trace" "Trace Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "SpanLog" "SpanLog Struct"
    
    # Alerting tests
    Test-StructDefinition "pkg/e2e/observability.go" "AlertRule" "AlertRule Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "Alert" "Alert Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "AlertCondition" "AlertCondition Struct"
    
    # Anomaly detection tests
    Test-StructDefinition "pkg/e2e/observability.go" "AnomalyModel" "AnomalyModel Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "Baseline" "Baseline Struct"
    Test-StructDefinition "pkg/e2e/observability.go" "Anomaly" "Anomaly Struct"
    
    # Function tests
    Test-FunctionDefinition "pkg/e2e/observability.go" "NewDistributedTracer" "NewDistributedTracer Function"
    Test-FunctionDefinition "pkg/e2e/observability.go" "NewAdvancedAlertManager" "NewAdvancedAlertManager Function"
    Test-FunctionDefinition "pkg/e2e/observability.go" "NewAnomalyDetector" "NewAnomalyDetector Function"
    Test-FunctionDefinition "pkg/e2e/observability.go" "StartSpan" "StartSpan Method"
    Test-FunctionDefinition "pkg/e2e/observability.go" "DetectAnomalies" "DetectAnomalies Method"
    
    # Interface tests
    Test-InterfaceDefinition "pkg/e2e/observability.go" "TraceExporter" "TraceExporter Interface"
    Test-InterfaceDefinition "pkg/e2e/observability.go" "AlertProcessor" "AlertProcessor Interface"
    
    # Constant tests
    Test-ConstantDefinition "pkg/e2e/observability.go" "AlertSeverityCritical.*AlertSeverity" "Alert Severity Constants"
    Test-ConstantDefinition "pkg/e2e/observability.go" "SpanStatusOK.*SpanStatusCode" "Span Status Constants"
}

# Compliance Automation Tests
if ($TestMode -eq "all" -or $TestMode -eq "compliance") {
    Write-Host "`n✅ Compliance Automation Tests..." -ForegroundColor Yellow
    
    # File existence tests
    Test-FileExists "pkg/e2e/compliance_automation.go" "Compliance Automation File"
    
    # Core structure tests
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "ComplianceAutomationEngine" "ComplianceAutomationEngine Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "ComplianceConfig" "ComplianceConfig Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "ComplianceResult" "ComplianceResult Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "ComplianceControl" "ComplianceControl Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "ComplianceViolation" "ComplianceViolation Struct"
    
    # Validator interface tests
    Test-InterfaceDefinition "pkg/e2e/compliance_automation.go" "ComplianceValidator" "ComplianceValidator Interface"
    
    # Standard-specific validator tests
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "FIPS140Validator" "FIPS140Validator Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "GDPRValidator" "GDPRValidator Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "SOC2Validator" "SOC2Validator Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "HIPAAValidator" "HIPAAValidator Struct"
    
    # Function tests
    Test-FunctionDefinition "pkg/e2e/compliance_automation.go" "NewComplianceAutomationEngine" "NewComplianceAutomationEngine Function"
    Test-FunctionDefinition "pkg/e2e/compliance_automation.go" "ValidateCompliance" "ValidateCompliance Method"
    Test-FunctionDefinition "pkg/e2e/compliance_automation.go" "ValidateStandard" "ValidateStandard Method"
    Test-FunctionDefinition "pkg/e2e/compliance_automation.go" "GenerateComplianceReport" "GenerateComplianceReport Method"
    
    # Evidence and remediation tests
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "Evidence" "Evidence Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "RemediationPlan" "RemediationPlan Struct"
    Test-StructDefinition "pkg/e2e/compliance_automation.go" "TestDetails" "TestDetails Struct"
    
    # Constant tests
    Test-ConstantDefinition "pkg/e2e/compliance_automation.go" "ComplianceStatusCompliant.*ComplianceStatus" "Compliance Status Constants"
    Test-ConstantDefinition "pkg/e2e/compliance_automation.go" "ControlTypeTechnical.*ControlType" "Control Type Constants"
}

# Integration and Configuration Tests
Write-Host "`n🔧 Integration and Configuration Tests..." -ForegroundColor Yellow

# Configuration tests
$configTests = @(
    @{ File = "pkg/e2e/mixnet.go"; Function = "DefaultMixnetConfig"; Name = "Default Mixnet Config" },
    @{ File = "pkg/e2e/hardware_security.go"; Function = "DefaultHardwareSecurityConfig"; Name = "Default Hardware Security Config" },
    @{ File = "pkg/e2e/observability.go"; Function = "DefaultAdvancedObservabilityConfig"; Name = "Default Advanced Observability Config" },
    @{ File = "pkg/e2e/compliance_automation.go"; Function = "DefaultComplianceConfig"; Name = "Default Compliance Config" }
)

foreach ($test in $configTests) {
    Test-FunctionDefinition $test.File $test.Function $test.Name
}

# Database schema tests
$schemaTests = @(
    @{ File = "pkg/e2e/mixnet_directory.go"; Pattern = "CREATE TABLE.*mixnet_nodes"; Name = "Mixnet Nodes Schema" },
    @{ File = "pkg/e2e/observability.go"; Pattern = "CREATE TABLE.*alert_rules"; Name = "Alert Rules Schema" },
    @{ File = "pkg/e2e/compliance_automation.go"; Pattern = "CREATE TABLE.*compliance_results"; Name = "Compliance Results Schema" }
)

foreach ($test in $schemaTests) {
    try {
        $content = Get-Content $test.File -Raw
        if ($content -match $test.Pattern) {
            Write-TestResult $test.Name "PASS" "Database schema defined"
        } else {
            Write-TestResult $test.Name "FAIL" "Database schema not found"
        }
    } catch {
        Write-TestResult $test.Name "FAIL" "Failed to check schema: $_"
    }
}

# Compilation Tests
Write-Host "`n🔨 Compilation Tests..." -ForegroundColor Yellow
Test-GoCompilation "pkg/e2e" "Sprint 7 Package Compilation"

# Integration with Existing E2E System
Write-Host "`n🔗 Integration Tests..." -ForegroundColor Yellow

# Check if new components integrate with existing config
try {
    $configContent = Get-Content "pkg/e2e/config.go" -Raw
    if ($configContent -match "E2EConfig.*struct") {
        # Check for new configuration fields (these would be added in a real integration)
        $integrationChecks = @(
            "MixnetConfig",
            "HardwareSecurityConfig", 
            "ObservabilityConfig",
            "ComplianceConfig"
        )
        
        $integratedCount = 0
        foreach ($check in $integrationChecks) {
            if ($configContent -match $check) {
                $integratedCount++
            }
        }
        
        if ($integratedCount -gt 0) {
            Write-TestResult "E2E Config Integration" "PASS" "$integratedCount new config sections found"
        } else {
            Write-TestResult "E2E Config Integration" "PARTIAL" "New config sections not yet integrated"
        }
    } else {
        Write-TestResult "E2E Config Integration" "FAIL" "E2EConfig struct not found"
    }
} catch {
    Write-TestResult "E2E Config Integration" "FAIL" "Failed to check integration: $_"
}

# Test Results Summary
Write-Host "`n📊 Test Results Summary" -ForegroundColor Cyan
Write-Host "=====================================`n" -ForegroundColor Cyan

Write-Host "✅ Passed Tests: $PassedTests" -ForegroundColor Green
Write-Host "❌ Failed Tests: $FailedTests" -ForegroundColor Red
Write-Host "📋 Total Tests: $($PassedTests + $FailedTests)" -ForegroundColor Blue

$PassRate = if ($PassedTests + $FailedTests -gt 0) { 
    [math]::Round(($PassedTests / ($PassedTests + $FailedTests)) * 100, 2) 
} else { 
    0 
}
Write-Host "📈 Pass Rate: $PassRate%" -ForegroundColor Yellow

Write-Host "`n🎯 Sprint 7 Component Status:" -ForegroundColor Cyan
$componentStatus = @{
    "Mixnet Routing" = ($TestResults | Where-Object { $_.Test -like "*Mixnet*" -and $_.Result -eq "PASS" }).Count
    "Hardware Security" = ($TestResults | Where-Object { $_.Test -like "*Hardware*" -and $_.Result -eq "PASS" }).Count
    "Observability" = ($TestResults | Where-Object { $_.Test -like "*Observability*" -or $_.Test -like "*Alert*" -or $_.Test -like "*Anomaly*" -and $_.Result -eq "PASS" }).Count
    "Compliance" = ($TestResults | Where-Object { $_.Test -like "*Compliance*" -and $_.Result -eq "PASS" }).Count
}

foreach ($component in $componentStatus.GetEnumerator()) {
    $status = if ($component.Value -gt 0) { "✅ Implemented" } else { "❌ Missing" }
    Write-Host "  $($component.Key): $status ($($component.Value) tests passed)" -ForegroundColor White
}

if ($FailedTests -gt 0) {
    Write-Host "`n❌ Failed Tests Details:" -ForegroundColor Red
    $TestResults | Where-Object { $_.Result -eq "FAIL" } | ForEach-Object {
        Write-Host "  • $($_.Test): $($_.Details)" -ForegroundColor Yellow
    }
}

# Advanced Features Assessment
Write-Host "`n🚀 Advanced Features Assessment:" -ForegroundColor Cyan

$advancedFeatures = @{
    "Anonymous Mixnet Routing" = ($TestResults | Where-Object { $_.Test -like "*Mixnet*" -and $_.Result -eq "PASS" }).Count -gt 5
    "Hardware-Backed Keys" = ($TestResults | Where-Object { $_.Test -like "*Hardware*" -and $_.Result -eq "PASS" }).Count -gt 5
    "Distributed Tracing" = ($TestResults | Where-Object { $_.Test -like "*Tracing*" -or $_.Test -like "*Span*" -and $_.Result -eq "PASS" }).Count -gt 0
    "ML Anomaly Detection" = ($TestResults | Where-Object { $_.Test -like "*Anomaly*" -and $_.Result -eq "PASS" }).Count -gt 0
    "Automated Compliance" = ($TestResults | Where-Object { $_.Test -like "*Compliance*" -and $_.Result -eq "PASS" }).Count -gt 5
}

foreach ($feature in $advancedFeatures.GetEnumerator()) {
    $status = if ($feature.Value) { "✅ Ready" } else { "⚠️ Needs Work" }
    Write-Host "  $($feature.Key): $status" -ForegroundColor White
}

Write-Host "`n🎉 Sprint 7 Test Harness Completed!" -ForegroundColor Green
Write-Host "Timestamp: $(Get-Date)" -ForegroundColor Gray

# Exit with appropriate code
if ($FailedTests -eq 0) {
    Write-Host "🎊 All tests passed! Sprint 7 is ready for validation." -ForegroundColor Green
    exit 0
} else {
    Write-Host "⚠️ Some tests failed. Please review and fix before proceeding." -ForegroundColor Yellow
    exit 1
}
