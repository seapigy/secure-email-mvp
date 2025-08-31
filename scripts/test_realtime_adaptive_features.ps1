# =============================================================================
# SECURE EMAIL MVP - MICRO-ITERATION 4.28 TEST SCRIPT
# =============================================================================
# Real-Time Retention Monitoring & Adaptive Policy Enforcement
# =============================================================================
# This script tests the new real-time monitoring and adaptive policy features
# including real-time metrics, adaptive policy changes, and policy performance analysis
# =============================================================================

param(
    [string]$ApiHost = "http://localhost:8080",
    [string]$AdminToken = "",
    [switch]$Verbose
)

# Color functions for output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "✅ $Message" "Green" }
function Write-Error { param([string]$Message) Write-ColorOutput "❌ $Message" "Red" }
function Write-Info { param([string]$Message) Write-ColorOutput "ℹ️  $Message" "Cyan" }
function Write-Warning { param([string]$Message) Write-ColorOutput "⚠️  $Message" "Yellow" }

# Test endpoint function
function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [string]$Description
    )

    Write-Info "Testing: $Description"
    Write-Info "  $Method $ApiHost$Endpoint"

    $headers = @{
        "Content-Type" = "application/json"
    }

    if ($AdminToken) {
        $headers["Authorization"] = "Bearer $AdminToken"
    }

    try {
        $params = @{
            Uri = "$ApiHost$Endpoint"
            Method = $Method
            Headers = $headers
        }

        if ($Body) {
            $params.Body = $Body | ConvertTo-Json -Depth 10
        }

        $response = Invoke-RestMethod @params -ErrorAction Stop

        if ($Verbose) {
            Write-Info "Response: $($response | ConvertTo-Json -Depth 10)"
        }

        Write-Success "✅ $Description - SUCCESS"
        return $response
    }
    catch {
        $errorMessage = $_.Exception.Message
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorMessage = $reader.ReadToEnd()
        }

        Write-Error "❌ $Description - FAILED: $errorMessage"
        return $null
    }
}

# Test real-time metrics endpoints
function Test-RealtimeMetricsEndpoints {
    Write-Info "`n=== Testing Real-Time Metrics Endpoints ==="

    # Test global real-time metrics
    $globalMetrics = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/retention-realtime" -Description "Get global real-time metrics"
    if ($globalMetrics) {
        Write-Info "Global metrics retrieved successfully"
        Write-Info "  Active emails: $($globalMetrics.data.active_emails_count)"
        Write-Info "  Archived emails: $($globalMetrics.data.archived_emails_count)"
        Write-Info "  Deleted emails: $($globalMetrics.data.deleted_emails_count)"
        Write-Info "  Total storage: $($globalMetrics.data.total_storage_bytes) bytes"
    }

    # Test user-specific metrics
    $userMetrics = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/retention-realtime?metric_type=user&metric_key=test@example.com" -Description "Get user-specific real-time metrics"
    if ($userMetrics) {
        Write-Info "User metrics retrieved successfully"
    }

    # Test domain-specific metrics
    $domainMetrics = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/retention-realtime?metric_type=domain&metric_key=example.com" -Description "Get domain-specific real-time metrics"
    if ($domainMetrics) {
        Write-Info "Domain metrics retrieved successfully"
    }

    # Test policy-specific metrics
    $policyMetrics = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/retention-realtime?metric_type=policy&metric_key=1" -Description "Get policy-specific real-time metrics"
    if ($policyMetrics) {
        Write-Info "Policy metrics retrieved successfully"
    }
}

# Test adaptive policy changes endpoints
function Test-AdaptivePolicyChangesEndpoints {
    Write-Info "`n=== Testing Adaptive Policy Changes Endpoints ==="

    # Get adaptive policy changes
    $changes = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/adaptive-policy-changes" -Description "Get adaptive policy changes"
    if ($changes) {
        Write-Info "Adaptive changes retrieved successfully"
        Write-Info "  Total changes: $($changes.total)"
        if ($changes.data -and $changes.data.Count -gt 0) {
            Write-Info "  Latest change: $($changes.data[0].change_type) for policy $($changes.data[0].policy_id)"
        }
    }

    # Get adaptive policy changes with filters
    $filteredChanges = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/adaptive-policy-changes?status=pending" -Description "Get pending adaptive policy changes"
    if ($filteredChanges) {
        Write-Info "Filtered changes retrieved successfully"
    }
}

# Test adaptive policy enable/disable endpoints
function Test-AdaptivePolicyControlEndpoints {
    Write-Info "`n=== Testing Adaptive Policy Control Endpoints ==="

    # Enable adaptive policy for a test policy
    $enableConfig = @{
        policy_id = 1
        max_change_percentage = 15.0
        cooldown_days = 5
        requires_admin_approval = $true
        min_retention_days = 1
        max_retention_days = 365
        min_archive_retention_days = 30
        max_archive_retention_days = 2555
        max_storage_impact_bytes = 1073741824
        max_archival_load_impact = 0.5
    }

    $enableResult = Test-Endpoint -Method "POST" -Endpoint "/api/admin/email/adaptive-policy/enable" -Body $enableConfig -Description "Enable adaptive policy"
    if ($enableResult) {
        Write-Info "Adaptive policy enabled successfully"
    }

    # Disable adaptive policy
    $disableConfig = @{
        policy_id = 1
    }

    $disableResult = Test-Endpoint -Method "POST" -Endpoint "/api/admin/email/adaptive-policy/disable" -Body $disableConfig -Description "Disable adaptive policy"
    if ($disableResult) {
        Write-Info "Adaptive policy disabled successfully"
    }
}

# Test policy performance analysis
function Test-PolicyPerformanceEndpoints {
    Write-Info "`n=== Testing Policy Performance Analysis Endpoints ==="

    # Analyze policy performance
    $performance = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/policy-performance?policy_id=1" -Description "Analyze policy performance"
    if ($performance) {
        Write-Info "Policy performance analyzed successfully"
        Write-Info "  Total evaluations: $($performance.data.total_evaluations)"
        Write-Info "  Match rate: $($performance.data.match_rate)%"
        Write-Info "  Application rate: $($performance.data.application_rate)%"
        Write-Info "  Average impact score: $($performance.data.avg_impact_score)"
        Write-Info "  Storage savings: $($performance.data.storage_savings_bytes) bytes"
    }
}

# Test adaptive recommendations generation
function Test-AdaptiveRecommendationsEndpoints {
    Write-Info "`n=== Testing Adaptive Recommendations Endpoints ==="

    # Generate adaptive recommendations
    $recommendations = Test-Endpoint -Method "POST" -Endpoint "/api/admin/email/adaptive-policy/generate-recommendations" -Description "Generate adaptive recommendations"
    if ($recommendations) {
        Write-Info "Adaptive recommendations generated successfully"
        Write-Info "  Total recommendations: $($recommendations.count)"
        if ($recommendations.data -and $recommendations.data.Count -gt 0) {
            Write-Info "  Latest recommendation: $($recommendations.data[0].change_type) for policy $($recommendations.data[0].policy_id)"
        }
    }
}

# Test adaptive change application
function Test-AdaptiveChangeApplication {
    Write-Info "`n=== Testing Adaptive Change Application ==="

    # First, get some adaptive changes
    $changes = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/adaptive-policy-changes?status=pending&limit=1" -Description "Get pending adaptive changes for testing"

    if ($changes -and $changes.data -and $changes.data.Count -gt 0) {
        $changeId = $changes.data[0].id

        # Preview the change
        $previewConfig = @{
            change_id = $changeId
            preview = $true
            applied_by = "test-admin"
        }

        $previewResult = Test-Endpoint -Method "POST" -Endpoint "/api/admin/email/adaptive-policy/apply" -Body $previewConfig -Description "Preview adaptive change"
        if ($previewResult) {
            Write-Info "Change preview generated successfully"
            Write-Info "  Change type: $($previewResult.data.change_type)"
            Write-Info "  Old value: $($previewResult.data.old_value)"
            Write-Info "  New value: $($previewResult.data.new_value)"
            Write-Info "  Expected savings: $($previewResult.data.expected_storage_savings) bytes"
        }

        # Apply the change (if not in preview mode)
        if (-not $previewConfig.preview) {
            $applyConfig = @{
                change_id = $changeId
                preview = $false
                applied_by = "test-admin"
            }

            $applyResult = Test-Endpoint -Method "POST" -Endpoint "/api/admin/email/adaptive-policy/apply" -Body $applyConfig -Description "Apply adaptive change"
            if ($applyResult) {
                Write-Info "Change applied successfully"
                Write-Info "  Applied result: $($applyResult.data.applied_result)"
            }
        }
    } else {
        Write-Warning "No pending adaptive changes found for testing"
    }
}

# Test data validation
function Test-DataValidation {
    Write-Info "`n=== Testing Data Validation ==="

    # Test invalid policy ID
    $invalidPerformance = Test-Endpoint -Method "GET" -Endpoint "/api/admin/email/policy-performance?policy_id=999999" -Description "Test invalid policy ID"
    if (-not $invalidPerformance) {
        Write-Success "Invalid policy ID correctly rejected"
    }

    # Test invalid change ID
    $invalidChangeConfig = @{
        change_id = 999999
        preview = $true
        applied_by = "test-admin"
    }

    $invalidChange = Test-Endpoint -Method "POST" -Endpoint "/api/admin/email/adaptive-policy/apply" -Body $invalidChangeConfig -Description "Test invalid change ID"
    if (-not $invalidChange) {
        Write-Success "Invalid change ID correctly rejected"
    }

    # Test missing required fields
    $invalidConfig = @{
        # Missing policy_id
        max_change_percentage = 15.0
    }

    $invalidConfigResult = Test-Endpoint -Method "POST" -Endpoint "/api/admin/email/adaptive-policy/enable" -Body $invalidConfig -Description "Test missing required fields"
    if (-not $invalidConfigResult) {
        Write-Success "Missing required fields correctly rejected"
    }
}

# Test performance and load handling
function Test-PerformanceAndLoad {
    Write-Info "`n=== Testing Performance and Load Handling ==="

    # Test concurrent requests to real-time metrics
    Write-Info "Testing concurrent real-time metrics requests..."
    $jobs = @()

    for ($i = 1; $i -le 5; $i++) {
        $jobs += Start-Job -ScriptBlock {
            param($ApiHost, $AdminToken)
            $headers = @{ "Authorization" = "Bearer $AdminToken" }
            try {
                $response = Invoke-RestMethod -Uri "$ApiHost/api/admin/email/retention-realtime" -Headers $headers -Method GET
                return "Success"
            } catch {
                return "Failed: $($_.Exception.Message)"
            }
        } -ArgumentList $ApiHost, $AdminToken
    }

    $results = $jobs | Wait-Job | Receive-Job
    $jobs | Remove-Job

    $successCount = ($results | Where-Object { $_ -eq "Success" }).Count
    Write-Info "Concurrent requests completed: $successCount/5 successful"

    if ($successCount -eq 5) {
        Write-Success "Concurrent request handling working correctly"
    } else {
        Write-Warning "Some concurrent requests failed"
    }
}

# Main test execution
function Main {
    Write-ColorOutput "`n🚀 Starting Micro-Iteration 4.28 Tests" "Magenta"
    Write-ColorOutput "API Host: $ApiHost" "White"
    Write-ColorOutput "Admin Token: $($AdminToken ? 'Provided' : 'Not provided')" "White"
    Write-ColorOutput "Verbose Mode: $($Verbose ? 'Enabled' : 'Disabled')" "White"

    if (-not $AdminToken) {
        Write-Warning "No admin token provided. Some endpoints may fail authentication."
    }

    # Run all test suites
    Test-RealtimeMetricsEndpoints
    Test-AdaptivePolicyChangesEndpoints
    Test-AdaptivePolicyControlEndpoints
    Test-PolicyPerformanceEndpoints
    Test-AdaptiveRecommendationsEndpoints
    Test-AdaptiveChangeApplication
    Test-DataValidation
    Test-PerformanceAndLoad

    Write-ColorOutput "`n🎉 Micro-Iteration 4.28 Tests Completed!" "Magenta"
    Write-Info "Check the output above for test results and any issues."
}

# Run the main function
Main

















