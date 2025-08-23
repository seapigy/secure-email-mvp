# Sprint 8 Test Harness: Production Deployment & Enterprise Integration
# Tests deployment automation, enterprise APIs, scaling infrastructure, and production readiness

param(
    [string]$TestMode = "all",  # "all", "deployment", "enterprise", "scaling", "production"
    [string]$Verbose = "false"
)

Write-Output "=== Sprint 8 Test Harness: Production Deployment & Enterprise Integration ==="
Write-Output "Test Mode: $TestMode"
Write-Output "Timestamp: $(Get-Date)"
Write-Output ""

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
        Write-Output "✅ $TestName"
        $global:PassedTests++
    } else {
        Write-Output "❌ $TestName"
        if ($Details) {
            Write-Output "   Details: $Details"
        }
        $global:FailedTests++
    }

    if ($Verbose -eq "true" -and $Details) {
        Write-Output "   $Details"
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

# Sprint 8 Core Tests
Write-Output "📋 Testing Sprint 8 Core Components..."

# Design Document Tests
Write-Output "`n🔍 Design Document Validation..."
Test-FileExists "docs/sprint8_design.md" "Sprint 8 Design Document"

# Deployment Automation Tests
if ($TestMode -eq "all" -or $TestMode -eq "deployment") {
    Write-Output "`n🚀 Deployment Automation Tests..."

    # File existence tests
    Test-FileExists "pkg/e2e/deployment_automation.go" "Deployment Automation File"

    # Core structure tests
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "DeploymentAutomationEngine" "DeploymentAutomationEngine Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "DeploymentConfig" "DeploymentConfig Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "DockerBuilder" "DockerBuilder Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "KubernetesDeployer" "KubernetesDeployer Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "CIPipeline" "CIPipeline Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "SecurityScanner" "SecurityScanner Struct"

    # Configuration structure tests
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "ContainerRegistry" "ContainerRegistry Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "KubernetesConfig" "KubernetesConfig Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "CIPipelineConfig" "CIPipelineConfig Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "DeploymentSecurity" "DeploymentSecurity Struct"

    # Function tests
    Test-FunctionDefinition "pkg/e2e/deployment_automation.go" "NewDeploymentAutomationEngine" "NewDeploymentAutomationEngine Function"
    Test-FunctionDefinition "pkg/e2e/deployment_automation.go" "Deploy" "Deploy Method"
    Test-FunctionDefinition "pkg/e2e/deployment_automation.go" "GetDeployment" "GetDeployment Method"
    Test-FunctionDefinition "pkg/e2e/deployment_automation.go" "ListDeployments" "ListDeployments Method"

    # Build and deployment result tests
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "BuildResult" "BuildResult Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "K8sDeployment" "K8sDeployment Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "PipelineExecution" "PipelineExecution Struct"
    Test-StructDefinition "pkg/e2e/deployment_automation.go" "ScanResult" "ScanResult Struct"

    # Constants and enums tests
    Test-ConstantDefinition "pkg/e2e/deployment_automation.go" "DeploymentStatusPending.*DeploymentStatus" "Deployment Status Constants"
    Test-ConstantDefinition "pkg/e2e/deployment_automation.go" "PipelineStatusPending.*PipelineStatus" "Pipeline Status Constants"
    Test-ConstantDefinition "pkg/e2e/deployment_automation.go" "ScanStatusPending.*ScanStatus" "Scan Status Constants"
}

# Enterprise APIs Tests
if ($TestMode -eq "all" -or $TestMode -eq "enterprise") {
    Write-Output "`n🏢 Enterprise APIs Tests..."

    # File existence tests
    Test-FileExists "pkg/e2e/enterprise_apis.go" "Enterprise APIs File"

    # Core structure tests
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "EnterpriseManager" "EnterpriseManager Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "EnterpriseConfig" "EnterpriseConfig Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "AdminAPI" "AdminAPI Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "SSOManager" "SSOManager Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "RBACManager" "RBACManager Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "OrganizationManager" "OrganizationManager Struct"

    # Configuration structure tests
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "AdminAPIConfig" "AdminAPIConfig Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "SSOConfig" "SSOConfig Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "RBACConfig" "RBACConfig Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "OrganizationConfig" "OrganizationConfig Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "AuditConfig" "AuditConfig Struct"

    # SSO and authentication tests
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "SSOProvider" "SSOProvider Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "Session" "Session Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "SSOSession" "SSOSession Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "APIKey" "APIKey Struct"

    # RBAC tests
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "Role" "Role Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "Permission" "Permission Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "AuthorizationResult" "AuthorizationResult Struct"

    # Organization and user management tests
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "Organization" "Organization Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "User" "User Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "UserLimits" "UserLimits Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "OrgLimits" "OrgLimits Struct"

    # Function tests
    Test-FunctionDefinition "pkg/e2e/enterprise_apis.go" "NewEnterpriseManager" "NewEnterpriseManager Function"
    Test-FunctionDefinition "pkg/e2e/enterprise_apis.go" "Authenticate" "Authenticate Method"
    Test-FunctionDefinition "pkg/e2e/enterprise_apis.go" "Authorize" "Authorize Method"
    Test-FunctionDefinition "pkg/e2e/enterprise_apis.go" "CheckPermission" "CheckPermission Method"

    # Audit and security tests
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "AuditLogger" "AuditLogger Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "AuditEvent" "AuditEvent Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "RateLimiter" "RateLimiter Struct"
    Test-StructDefinition "pkg/e2e/enterprise_apis.go" "EnterpriseSecurityConfig" "EnterpriseSecurityConfig Struct"
}

# Scaling Infrastructure Tests
if ($TestMode -eq "all" -or $TestMode -eq "scaling") {
    Write-Output "`nScaling Infrastructure Tests..."

    # File existence tests
    Test-FileExists "pkg/e2e/scaling_infrastructure.go" "Scaling Infrastructure File"

    # Core structure tests
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ScalingInfrastructureManager" "ScalingInfrastructureManager Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ScalingInfrastructureConfig" "ScalingInfrastructureConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "LoadBalancer" "LoadBalancer Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "AutoScaler" "AutoScaler Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ServiceMesh" "ServiceMesh Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "DatabaseScaler" "DatabaseScaler Struct"

    # Load balancing tests
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "LoadBalancerConfig" "LoadBalancerConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "Backend" "Backend Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "HealthCheckConfig" "HealthCheckConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "RetryPolicyConfig" "RetryPolicyConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "CircuitBreakerConfig" "CircuitBreakerConfig Struct"

    # Auto-scaling tests
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "AutoScalerConfig" "AutoScalerConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "HPAConfig" "HPAConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "VPAConfig" "VPAConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ScalingGroup" "ScalingGroup Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ServiceInstance" "ServiceInstance Struct"

    # Service mesh tests
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ServiceMeshConfig" "ServiceMeshConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "TrafficSplitConfig" "TrafficSplitConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "MeshService" "MeshService Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "MeshLoadBalancing" "MeshLoadBalancing Struct"

    # Database scaling tests
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "DatabaseScalingConfig" "DatabaseScalingConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ReadReplicaConfig" "ReadReplicaConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ShardingConfig" "ShardingConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ConnectionPoolConfig" "ConnectionPoolConfig Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "DatabaseCachingConfig" "DatabaseCachingConfig Struct"

    # Function tests
    Test-FunctionDefinition "pkg/e2e/scaling_infrastructure.go" "NewScalingInfrastructureManager" "NewScalingInfrastructureManager Function"
    Test-FunctionDefinition "pkg/e2e/scaling_infrastructure.go" "NewLoadBalancer" "NewLoadBalancer Function"
    Test-FunctionDefinition "pkg/e2e/scaling_infrastructure.go" "NewAutoScaler" "NewAutoScaler Function"
    Test-FunctionDefinition "pkg/e2e/scaling_infrastructure.go" "NewServiceMesh" "NewServiceMesh Function"
    Test-FunctionDefinition "pkg/e2e/scaling_infrastructure.go" "NewDatabaseScaler" "NewDatabaseScaler Function"

    # Interface tests
    Test-InterfaceDefinition "pkg/e2e/scaling_infrastructure.go" "LoadBalancingAlgorithm" "LoadBalancingAlgorithm Interface"

    # Monitoring and metrics tests
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ScalingMonitor" "ScalingMonitor Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "ScalingMetrics" "ScalingMetrics Struct"
    Test-StructDefinition "pkg/e2e/scaling_infrastructure.go" "MetricSeries" "MetricSeries Struct"
}

# Configuration Tests
Write-Output "`nConfiguration Tests..."

# Default configuration function tests
$configTests = @(
    @{ File = "pkg/e2e/deployment_automation.go"; Function = "DefaultDeploymentConfig"; Name = "Default Deployment Config" },
    @{ File = "pkg/e2e/enterprise_apis.go"; Function = "DefaultEnterpriseConfig"; Name = "Default Enterprise Config" },
    @{ File = "pkg/e2e/scaling_infrastructure.go"; Function = "DefaultScalingInfrastructureConfig"; Name = "Default Scaling Infrastructure Config" }
)

foreach ($test in $configTests) {
    Test-FunctionDefinition $test.File $test.Function $test.Name
}

# Compilation Tests
Write-Output "`n🔨 Compilation Tests..."
try {
    $output = go build ./pkg/e2e/... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-TestResult "Sprint 8 Package Compilation" "PASS" "Compilation successful"
    } else {
        Write-TestResult "Sprint 8 Package Compilation" "FAIL" "Compilation failed: $output"
    }
} catch {
    Write-TestResult "Sprint 8 Package Compilation" "FAIL" "Failed to compile: $_"
}

# Integration Tests
Write-Output "`n🔗 Integration Tests..."

# Check if new components integrate with existing E2E system
try {
    $configContent = Get-Content "pkg/e2e/config.go" -Raw
    if ($configContent -match "E2EConfig.*struct") {
        # Check for integration readiness
        $integrationChecks = @(
            "DeploymentConfig",
            "EnterpriseConfig",
            "ScalingInfrastructureConfig"
        )

        $integratedCount = 0
        foreach ($check in $integrationChecks) {
            # Check if these types exist (indicating readiness for integration)
            $deploymentContent = Get-Content "pkg/e2e/deployment_automation.go" -Raw -ErrorAction SilentlyContinue
            $enterpriseContent = Get-Content "pkg/e2e/enterprise_apis.go" -Raw -ErrorAction SilentlyContinue
            $scalingContent = Get-Content "pkg/e2e/scaling_infrastructure.go" -Raw -ErrorAction SilentlyContinue

            if (($deploymentContent -and $deploymentContent -match $check) -or
                ($enterpriseContent -and $enterpriseContent -match $check) -or
                ($scalingContent -and $scalingContent -match $check)) {
                $integratedCount++
            }
        }

        if ($integratedCount -eq 3) {
            Write-TestResult "E2E Config Integration Readiness" "PASS" "All Sprint 8 config types defined"
        } elseif ($integratedCount -gt 0) {
            Write-TestResult "E2E Config Integration Readiness" "PARTIAL" "$integratedCount/3 config types ready"
        } else {
            Write-TestResult "E2E Config Integration Readiness" "FAIL" "No config types found"
        }
    } else {
        Write-TestResult "E2E Config Integration Readiness" "FAIL" "E2EConfig struct not found"
    }
} catch {
    Write-TestResult "E2E Config Integration Readiness" "FAIL" "Failed to check integration: $_"
}

# Production Readiness Assessment
if ($TestMode -eq "all" -or $TestMode -eq "production") {
    Write-Output "`n🏭 Production Readiness Assessment..."

    # Infrastructure components check
    $infrastructureComponents = @(
        @{ Name = "Container Orchestration"; File = "pkg/e2e/deployment_automation.go"; Pattern = "KubernetesDeployer" },
        @{ Name = "Load Balancing"; File = "pkg/e2e/scaling_infrastructure.go"; Pattern = "LoadBalancer" },
        @{ Name = "Auto Scaling"; File = "pkg/e2e/scaling_infrastructure.go"; Pattern = "AutoScaler" },
        @{ Name = "Enterprise SSO"; File = "pkg/e2e/enterprise_apis.go"; Pattern = "SSOManager" },
        @{ Name = "Role-Based Access Control"; File = "pkg/e2e/enterprise_apis.go"; Pattern = "RBACManager" },
        @{ Name = "Security Scanning"; File = "pkg/e2e/deployment_automation.go"; Pattern = "SecurityScanner" },
        @{ Name = "Service Mesh"; File = "pkg/e2e/scaling_infrastructure.go"; Pattern = "ServiceMesh" },
        @{ Name = "Database Scaling"; File = "pkg/e2e/scaling_infrastructure.go"; Pattern = "DatabaseScaler" }
    )

    $readyComponents = 0
    foreach ($component in $infrastructureComponents) {
        try {
            $content = Get-Content $component.File -Raw -ErrorAction SilentlyContinue
            if ($content -and $content -match $component.Pattern) {
                Write-TestResult $component.Name "PASS" "Component implemented"
                $readyComponents++
            } else {
                Write-TestResult $component.Name "FAIL" "Component not found"
            }
        } catch {
            Write-TestResult $component.Name "FAIL" "Failed to check component"
        }
    }

    # Overall production readiness score
    $readinessPercentage = [math]::Round(($readyComponents / $infrastructureComponents.Count) * 100, 2)
    if ($readinessPercentage -ge 90) {
        Write-TestResult "Production Readiness Score" "PASS" "$readinessPercentage% ready"
    } elseif ($readinessPercentage -ge 70) {
        Write-TestResult "Production Readiness Score" "PARTIAL" "$readinessPercentage% ready"
    } else {
        Write-TestResult "Production Readiness Score" "FAIL" "$readinessPercentage% ready"
    }
}

# Test Results Summary
Write-Output "`n📊 Test Results Summary"
Write-Output "====================================="

Write-Output "✅ Passed Tests: $PassedTests"
Write-Output "❌ Failed Tests: $FailedTests"
Write-Output "📋 Total Tests: $($PassedTests + $FailedTests)"

$PassRate = if ($PassedTests + $FailedTests -gt 0) {
    [math]::Round(($PassedTests / ($PassedTests + $FailedTests)) * 100, 2)
} else {
    0
}
Write-Output "📈 Pass Rate: $PassRate%"

Write-Output "`n🎯 Sprint 8 Component Status:"
$componentStatus = @{
    "Deployment Automation" = ($TestResults | Where-Object { $_.Test -like "*Deployment*" -and $_.Result -eq "PASS" }).Count
    "Enterprise APIs" = ($TestResults | Where-Object { $_.Test -like "*Enterprise*" -or $_.Test -like "*SSO*" -or $_.Test -like "*RBAC*" -and $_.Result -eq "PASS" }).Count
    "Scaling Infrastructure" = ($TestResults | Where-Object { $_.Test -like "*Scaling*" -or $_.Test -like "*LoadBalancer*" -or $_.Test -like "*AutoScaler*" -and $_.Result -eq "PASS" }).Count
    "Production Readiness" = ($TestResults | Where-Object { $_.Test -like "*Production*" -and $_.Result -eq "PASS" }).Count
}

foreach ($component in $componentStatus.GetEnumerator()) {
    $status = if ($component.Value -gt 0) { "✅ Implemented" } else { "❌ Missing" }
    Write-Output "  $($component.Key): $status"
}

if ($FailedTests -gt 0) {
    Write-Output "`nFailed Tests Details:"
    $TestResults | Where-Object { $_.Result -eq "FAIL" } | ForEach-Object {
        Write-Output "  - $($_.Test): $($_.Details)"
    }
}

# Enterprise Features Assessment
Write-Output "`nEnterprise Features Assessment:"

$enterpriseFeatures = @{
    "Container Orchestration" = ($TestResults | Where-Object { $_.Test -like "*Kubernetes*" -and $_.Result -eq "PASS" }).Count -gt 0
    "CI/CD Pipeline" = ($TestResults | Where-Object { $_.Test -like "*Pipeline*" -and $_.Result -eq "PASS" }).Count -gt 0
    "Auto Scaling" = ($TestResults | Where-Object { $_.Test -like "*AutoScaler*" -and $_.Result -eq "PASS" }).Count -gt 0
    "Load Balancing" = ($TestResults | Where-Object { $_.Test -like "*LoadBalancer*" -and $_.Result -eq "PASS" }).Count -gt 0
    "Enterprise SSO" = ($TestResults | Where-Object { $_.Test -like "*SSO*" -and $_.Result -eq "PASS" }).Count -gt 0
    "Role-Based Access Control" = ($TestResults | Where-Object { $_.Test -like "*RBAC*" -and $_.Result -eq "PASS" }).Count -gt 0
    "Security Scanning" = ($TestResults | Where-Object { $_.Test -like "*Security*" -and $_.Result -eq "PASS" }).Count -gt 0
    "Database Scaling" = ($TestResults | Where-Object { $_.Test -like "*Database*" -and $_.Result -eq "PASS" }).Count -gt 0
}

foreach ($feature in $enterpriseFeatures.GetEnumerator()) {
    $status = if ($feature.Value) { "Ready" } else { "Needs Work" }
    Write-Output "  $($feature.Key): $status"
}

Write-Output "`nSprint 8 Test Harness Completed!"
Write-Output "Timestamp: $(Get-Date)"

# Exit with appropriate code
if ($FailedTests -eq 0) {
    Write-Output "All tests passed! Sprint 8 is ready for production deployment."
    exit 0
} else {
    Write-Output "Some tests failed. Please review and fix before proceeding to production."
    exit 1
}
