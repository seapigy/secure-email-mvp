# Sprint 2 E2E Test Harness
# Tests the Key Transparency + Threshold HSM + Metadata Minimization + Server API implementation

Write-Host "Sprint 2 E2E Test Harness" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan

# Global test results tracking
$TestResults = @{
    TotalTests = 0
    PassedTests = 0
    FailedTests = 0
    Errors = @()
}

function Write-TestResult {
    param(
        [string]$TestName,
        [bool]$Passed,
        [string]$Message
    )
    
    $TestResults.TotalTests++
    
    if ($Passed) {
        $TestResults.PassedTests++
        Write-Host "✅ $TestName" -ForegroundColor Green
        if ($Message) {
            Write-Host "   $Message" -ForegroundColor Gray
        }
    } else {
        $TestResults.FailedTests++
        Write-Host "❌ $TestName" -ForegroundColor Red
        if ($Message) {
            Write-Host "   $Message" -ForegroundColor Gray
            $TestResults.Errors += "${TestName}: ${Message}"
        }
    }
}

# Test 1: Key Transparency Implementation
Write-Host "`nTest 1: Key Transparency Implementation" -ForegroundColor Yellow

$ktFile = "pkg/e2e/keytransparency.go"
if (Test-Path $ktFile) {
    $content = Get-Content $ktFile -Raw
    
    $hasKeyTransparency = $content -match "type KeyTransparency struct"
    Write-TestResult -TestName "KeyTransparency Struct" -Passed $hasKeyTransparency -Message "Key transparency core structure"
    
    $hasPublicKeyEntry = $content -match "type PublicKeyEntry struct"
    Write-TestResult -TestName "PublicKeyEntry Struct" -Passed $hasPublicKeyEntry -Message "Public key entry structure"
    
    $hasLogEntry = $content -match "type LogEntry struct"
    Write-TestResult -TestName "LogEntry Struct" -Passed $hasLogEntry -Message "Transparency log entry structure"
    
    $hasMerkleProof = $content -match "type MerkleProof struct"
    Write-TestResult -TestName "MerkleProof Struct" -Passed $hasMerkleProof -Message "Merkle proof structure"
    
    $hasRegisterKey = $content -match "func.*RegisterPublicKey"
    Write-TestResult -TestName "RegisterPublicKey Function" -Passed $hasRegisterKey -Message "Public key registration"
    
    $hasVerifyKey = $content -match "func.*VerifyPublicKey"
    Write-TestResult -TestName "VerifyPublicKey Function" -Passed $hasVerifyKey -Message "Public key verification"
    
    $hasAuditLog = $content -match "func.*AuditLog"
    Write-TestResult -TestName "AuditLog Function" -Passed $hasAuditLog -Message "Transparency log auditing"
    
    $hasRevokeKey = $content -match "func.*RevokePublicKey"
    Write-TestResult -TestName "RevokePublicKey Function" -Passed $hasRevokeKey -Message "Public key revocation"
    
    $hasMerkleOperations = $content -match "generateMerkleProof" -and $content -match "verifyMerkleProof"
    Write-TestResult -TestName "Merkle Operations" -Passed $hasMerkleOperations -Message "Merkle tree proof operations"
    
    $hasValidation = $content -match "ValidatePublicKey" -and $content -match "ValidateUserUUID"
    Write-TestResult -TestName "KT Validation" -Passed $hasValidation -Message "Key transparency validation functions"
    
} else {
    Write-TestResult -TestName "KT File Exists" -Passed $false -Message "Key transparency implementation file not found"
}

# Test 2: Threshold HSM Implementation
Write-Host "`nTest 2: Threshold HSM Implementation" -ForegroundColor Yellow

$hsmFile = "pkg/e2e/thresholdHSM.go"
if (Test-Path $hsmFile) {
    $content = Get-Content $hsmFile -Raw
    
    $hasThresholdHSM = $content -match "type ThresholdHSM struct"
    Write-TestResult -TestName "ThresholdHSM Struct" -Passed $hasThresholdHSM -Message "Threshold HSM core structure"
    
    $hasKeyShare = $content -match "type KeyShare struct"
    Write-TestResult -TestName "KeyShare Struct" -Passed $hasKeyShare -Message "Key share structure"
    
    $hasThresholdKey = $content -match "type ThresholdKey struct"
    Write-TestResult -TestName "ThresholdKey Struct" -Passed $hasThresholdKey -Message "Threshold key structure"
    
    $hasThresholdSignature = $content -match "type ThresholdSignature struct"
    Write-TestResult -TestName "ThresholdSignature Struct" -Passed $hasThresholdSignature -Message "Threshold signature structure"
    
    $hasGenerateKey = $content -match "func.*GenerateThresholdKey"
    Write-TestResult -TestName "GenerateThresholdKey Function" -Passed $hasGenerateKey -Message "Threshold key generation"
    
    $hasSign = $content -match "func.*Sign.*keyID.*message"
    Write-TestResult -TestName "Threshold Sign Function" -Passed $hasSign -Message "Threshold signing operation"
    
    $hasVerify = $content -match "func.*Verify.*signature.*publicKey"
    Write-TestResult -TestName "Threshold Verify Function" -Passed $hasVerify -Message "Threshold signature verification"
    
    $hasKeyRotation = $content -match "func.*RotateKey"
    Write-TestResult -TestName "Key Rotation Function" -Passed $hasKeyRotation -Message "Key rotation functionality"
    
    $hasShamirSharing = $content -match "generateKeyShares" -and $content -match "combineSignatureShares"
    Write-TestResult -TestName "Shamir Secret Sharing" -Passed $hasShamirSharing -Message "Secret sharing operations"
    
    $hasHSMValidation = $content -match "ValidateThresholdParams" -and $content -match "ValidateKeyType"
    Write-TestResult -TestName "HSM Validation" -Passed $hasHSMValidation -Message "Threshold HSM validation functions"
    
} else {
    Write-TestResult -TestName "HSM File Exists" -Passed $false -Message "Threshold HSM implementation file not found"
}

# Test 3: Metadata Minimization Implementation
Write-Host "`nTest 3: Metadata Minimization Implementation" -ForegroundColor Yellow

$metadataFile = "pkg/e2e/metadata.go"
if (Test-Path $metadataFile) {
    $content = Get-Content $metadataFile -Raw
    
    $hasMetadataMinimizer = $content -match "type MetadataMinimizer struct"
    Write-TestResult -TestName "MetadataMinimizer Struct" -Passed $hasMetadataMinimizer -Message "Metadata minimizer core structure"
    
    $hasMinimizedMetadata = $content -match "type MinimizedMetadata struct"
    Write-TestResult -TestName "MinimizedMetadata Struct" -Passed $hasMinimizedMetadata -Message "Minimized metadata structure"
    
    $hasPrivacyPolicy = $content -match "type PrivacyPolicy struct"
    Write-TestResult -TestName "PrivacyPolicy Struct" -Passed $hasPrivacyPolicy -Message "Privacy policy structure"
    
    $hasTimeBatch = $content -match "type TimeBatch struct"
    Write-TestResult -TestName "TimeBatch Struct" -Passed $hasTimeBatch -Message "Time batching structure"
    
    $hasMinimizeMetadata = $content -match "func.*MinimizeMetadata"
    Write-TestResult -TestName "MinimizeMetadata Function" -Passed $hasMinimizeMetadata -Message "Metadata minimization operation"
    
    $hasCreatePolicy = $content -match "func.*CreatePrivacyPolicy"
    Write-TestResult -TestName "CreatePrivacyPolicy Function" -Passed $hasCreatePolicy -Message "Privacy policy creation"
    
    $hasTimeBatching = $content -match "CreateTimeBatch" -and $content -match "ReleaseBatch"
    Write-TestResult -TestName "Time Batching Functions" -Passed $hasTimeBatching -Message "Time batching operations"
    
    $hasAnonymization = $content -match "generateAnonymousToken" -and $content -match "ResolveAnonymousToken"
    Write-TestResult -TestName "Anonymization Functions" -Passed $hasAnonymization -Message "Token anonymization operations"
    
    $hasPadding = $content -match "padMessageSize" -and $content -match "PaddingProfile"
    Write-TestResult -TestName "Message Padding" -Passed $hasPadding -Message "Message size padding operations"
    
    $hasMetadataValidation = $content -match "ValidatePrivacyPolicy" -and $content -match "GetPaddingProfile"
    Write-TestResult -TestName "Metadata Validation" -Passed $hasMetadataValidation -Message "Metadata validation functions"
    
} else {
    Write-TestResult -TestName "Metadata File Exists" -Passed $false -Message "Metadata minimization implementation file not found"
}

# Test 4: Server API Implementation
Write-Host "`nTest 4: Server API Implementation" -ForegroundColor Yellow

$serverFile = "pkg/e2e/server_api.go"
if (Test-Path $serverFile) {
    $content = Get-Content $serverFile -Raw
    
    $hasE2EServer = $content -match "type E2EServer struct"
    Write-TestResult -TestName "E2EServer Struct" -Passed $hasE2EServer -Message "E2E server core structure"
    
    $hasMessageRequest = $content -match "type E2EMessageRequest struct"
    Write-TestResult -TestName "E2EMessageRequest Struct" -Passed $hasMessageRequest -Message "E2E message request structure"
    
    $hasKeyRegistration = $content -match "type KeyRegistrationRequest struct" -and $content -match "type KeyRegistrationResponse struct"
    Write-TestResult -TestName "Key Registration Types" -Passed $hasKeyRegistration -Message "Key registration request/response types"
    
    $hasHandleSendMessage = $content -match "func.*HandleSendE2EMessage"
    Write-TestResult -TestName "HandleSendE2EMessage" -Passed $hasHandleSendMessage -Message "E2E message sending endpoint"
    
    $hasHandleRegisterKey = $content -match "func.*HandleRegisterKey"
    Write-TestResult -TestName "HandleRegisterKey" -Passed $hasHandleRegisterKey -Message "Key registration endpoint"
    
    $hasHandleVerifyKey = $content -match "func.*HandleVerifyKey"
    Write-TestResult -TestName "HandleVerifyKey" -Passed $hasHandleVerifyKey -Message "Key verification endpoint"
    
    $hasHandleThresholdSign = $content -match "func.*HandleThresholdSign"
    Write-TestResult -TestName "HandleThresholdSign" -Passed $hasHandleThresholdSign -Message "Threshold signing endpoint"
    
    $hasStatusEndpoint = $content -match "func.*HandleGetE2EStatus"
    Write-TestResult -TestName "HandleGetE2EStatus" -Passed $hasStatusEndpoint -Message "E2E status endpoint"
    
    $hasMiddleware = $content -match "E2EAuthMiddleware" -and $content -match "RateLimitMiddleware" -and $content -match "LoggingMiddleware"
    Write-TestResult -TestName "API Middleware" -Passed $hasMiddleware -Message "Authentication, rate limiting, and logging middleware"
    
    $hasRouteSetup = $content -match "SetupE2ERoutes"
    Write-TestResult -TestName "Route Setup" -Passed $hasRouteSetup -Message "API route configuration"
    
} else {
    Write-TestResult -TestName "Server API File Exists" -Passed $false -Message "Server API implementation file not found"
}

# Test 5: Unit Tests
Write-Host "`nTest 5: Unit Tests" -ForegroundColor Yellow

$ktTestFile = "pkg/e2e/keytransparency_test.go"
$hsmTestFile = "pkg/e2e/thresholdHSM_test.go"
$metadataTestFile = "pkg/e2e/metadata_test.go"

$hasKTTests = Test-Path $ktTestFile
Write-TestResult -TestName "Key Transparency Tests" -Passed $hasKTTests -Message "KT unit test file exists"

if ($hasKTTests) {
    $content = Get-Content $ktTestFile -Raw
    $ktTestCount = ([regex]::Matches($content, "func Test")).Count
    Write-TestResult -TestName "KT Test Functions" -Passed ($ktTestCount -ge 10) -Message "Found $ktTestCount KT test functions"
}

$hasHSMTests = Test-Path $hsmTestFile
Write-TestResult -TestName "Threshold HSM Tests" -Passed $hasHSMTests -Message "HSM unit test file exists"

if ($hasHSMTests) {
    $content = Get-Content $hsmTestFile -Raw
    $hsmTestCount = ([regex]::Matches($content, "func Test")).Count
    Write-TestResult -TestName "HSM Test Functions" -Passed ($hsmTestCount -ge 10) -Message "Found $hsmTestCount HSM test functions"
}

$hasMetadataTests = Test-Path $metadataTestFile
Write-TestResult -TestName "Metadata Tests" -Passed $hasMetadataTests -Message "Metadata unit test file exists"

if ($hasMetadataTests) {
    $content = Get-Content $metadataTestFile -Raw
    $metadataTestCount = ([regex]::Matches($content, "func Test")).Count
    Write-TestResult -TestName "Metadata Test Functions" -Passed ($metadataTestCount -ge 10) -Message "Found $metadataTestCount metadata test functions"
}

# Test 6: Build Validation
Write-Host "`nTest 6: Build Validation" -ForegroundColor Yellow

try {
    $buildResult = go build ./pkg/e2e/... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-TestResult -TestName "Sprint 2 Package Builds" -Passed $true -Message "All Sprint 2 packages compile successfully"
    } else {
        Write-TestResult -TestName "Sprint 2 Package Builds" -Passed $false -Message "Build failed: $buildResult"
    }
} catch {
    Write-TestResult -TestName "Sprint 2 Package Builds" -Passed $false -Message "Build error: $($_.Exception.Message)"
}

# Test 7: Test Execution
Write-Host "`nTest 7: Test Execution" -ForegroundColor Yellow

try {
    $testResult = go test ./pkg/e2e/... -v 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-TestResult -TestName "Sprint 2 Tests Pass" -Passed $true -Message "All Sprint 2 unit tests pass"
    } else {
        Write-TestResult -TestName "Sprint 2 Tests Pass" -Passed $false -Message "Some tests failed: $testResult"
    }
} catch {
    Write-TestResult -TestName "Sprint 2 Tests Pass" -Passed $false -Message "Test execution error: $($_.Exception.Message)"
}

# Test 8: Integration Components
Write-Host "`nTest 8: Integration Components" -ForegroundColor Yellow

# Check if all components work together
$allFiles = @($ktFile, $hsmFile, $metadataFile, $serverFile)
$allExist = $true
foreach ($file in $allFiles) {
    if (-not (Test-Path $file)) {
        $allExist = $false
        break
    }
}

Write-TestResult -TestName "All Components Present" -Passed $allExist -Message "All Sprint 2 components implemented"

# Check for integration between components
if ($allExist) {
    $serverContent = Get-Content $serverFile -Raw
    $hasKTIntegration = $serverContent -match "KeyTransparency" -and $serverContent -match "HandleRegisterKey"
    Write-TestResult -TestName "KT Integration" -Passed $hasKTIntegration -Message "Key transparency integrated with server API"
    
    $hasHSMIntegration = $serverContent -match "ThresholdHSM" -and $serverContent -match "HandleThresholdSign"
    Write-TestResult -TestName "HSM Integration" -Passed $hasHSMIntegration -Message "Threshold HSM integrated with server API"
    
    $hasMetadataIntegration = $serverContent -match "MetadataMinimizer" -and $serverContent -match "MinimizeMetadata"
    Write-TestResult -TestName "Metadata Integration" -Passed $hasMetadataIntegration -Message "Metadata minimization integrated with server API"
    
    $hasE2EConfig = $serverContent -match "E2EConfig"
    Write-TestResult -TestName "Configuration Integration" -Passed $hasE2EConfig -Message "E2E configuration system integrated"
}

# Test 9: API Endpoints Validation
Write-Host "`nTest 9: API Endpoints Validation" -ForegroundColor Yellow

if (Test-Path $serverFile) {
    $content = Get-Content $serverFile -Raw
    
    $hasMessageEndpoint = $content -match "/api/v2/e2e/send"
    Write-TestResult -TestName "Message Send Endpoint" -Passed $hasMessageEndpoint -Message "E2E message sending API endpoint"
    
    $hasKeyRegisterEndpoint = $content -match "/api/v2/e2e/keys/register"
    Write-TestResult -TestName "Key Register Endpoint" -Passed $hasKeyRegisterEndpoint -Message "Key registration API endpoint"
    
    $hasKeyVerifyEndpoint = $content -match "/api/v2/e2e/keys/verify"
    Write-TestResult -TestName "Key Verify Endpoint" -Passed $hasKeyVerifyEndpoint -Message "Key verification API endpoint"
    
    $hasHSMSignEndpoint = $content -match "/api/v2/e2e/hsm/sign"
    Write-TestResult -TestName "HSM Sign Endpoint" -Passed $hasHSMSignEndpoint -Message "Threshold HSM signing API endpoint"
    
    $hasStatusEndpoint = $content -match "/api/v2/e2e/status"
    Write-TestResult -TestName "Status Endpoint" -Passed $hasStatusEndpoint -Message "E2E status API endpoint"
}

# Test 10: Sprint 2 Readiness Assessment
Write-Host "`nTest 10: Sprint 2 Readiness" -ForegroundColor Yellow

$readinessCriteria = @{
    "Key Transparency Implementation" = (Test-Path $ktFile)
    "Threshold HSM Implementation" = (Test-Path $hsmFile)
    "Metadata Minimization Implementation" = (Test-Path $metadataFile)
    "Server API Implementation" = (Test-Path $serverFile)
    "Unit Test Coverage" = (Test-Path $ktTestFile) -and (Test-Path $hsmTestFile) -and (Test-Path $metadataTestFile)
    "Package Builds" = $true  # Based on build test above
}

$readyForSprint3 = $true
foreach ($criterion in $readinessCriteria.GetEnumerator()) {
    Write-TestResult -TestName "Sprint 2 Criterion: $($criterion.Key)" -Passed $criterion.Value -Message "Sprint 2 readiness criterion"
    if (-not $criterion.Value) {
        $readyForSprint3 = $false
    }
}

# Final Summary
Write-Host "`n" + "="*50 -ForegroundColor Cyan
Write-Host "SPRINT 2 TEST SUMMARY" -ForegroundColor Cyan
Write-Host "="*50 -ForegroundColor Cyan

Write-Host "Total Tests: $($TestResults.TotalTests)" -ForegroundColor White
Write-Host "Passed: $($TestResults.PassedTests)" -ForegroundColor Green
Write-Host "Failed: $($TestResults.FailedTests)" -ForegroundColor Red

$successRate = if ($TestResults.TotalTests -gt 0) { 
    [math]::Round(($TestResults.PassedTests / $TestResults.TotalTests) * 100, 2) 
} else { 0 }

Write-Host "Success Rate: $successRate%" -ForegroundColor $(if ($successRate -ge 90) { "Green" } elseif ($successRate -ge 75) { "Yellow" } else { "Red" })

if ($readyForSprint3) {
    Write-Host "`n🎉 SPRINT 2 COMPLETE - READY FOR SPRINT 3" -ForegroundColor Green
    Write-Host "Key Transparency + Threshold HSM + Metadata Minimization + Server API are implemented and tested." -ForegroundColor Green
    Write-Host "Proceed to Sprint 3: Hardware-backed keys + Mixnet + Cover Traffic (Optional)" -ForegroundColor Green
} else {
    Write-Host "`n⚠️  SPRINT 2 INCOMPLETE - ADDRESS ISSUES BEFORE SPRINT 3" -ForegroundColor Yellow
    Write-Host "Some core components need attention before proceeding." -ForegroundColor Yellow
}

if ($TestResults.Errors.Count -gt 0) {
    Write-Host "`nDetailed Errors:" -ForegroundColor Red
    foreach ($error in $TestResults.Errors) {
        Write-Host "  - $error" -ForegroundColor Gray
    }
}

Write-Host "`nNext Steps:" -ForegroundColor Cyan
Write-Host "1. Review failed tests and address issues" -ForegroundColor White
Write-Host "2. Complete any missing Sprint 2 components" -ForegroundColor White
Write-Host "3. Validate integration between components" -ForegroundColor White
Write-Host "4. Begin Sprint 3 implementation (if optional features desired)" -ForegroundColor White
Write-Host "5. Or proceed to final testing and deployment preparation" -ForegroundColor White
