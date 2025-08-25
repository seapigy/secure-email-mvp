#!/usr/bin/env pwsh

# Generate Current TOTP Code for Testing
# This script generates the current TOTP code for the test user

param(
    [string]$Secret = "JBSWY3DPEHPK3PXP",
    [int]$Digits = 6,
    [int]$Period = 30
)

Write-Host "🔐 Generating Current TOTP Code" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

# Calculate current time step
$epoch = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$timeStep = [math]::Floor($epoch / $Period)

Write-Host "📅 Current Time Information:" -ForegroundColor Yellow
Write-Host "  - UTC Time: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss UTC')"
Write-Host "  - Epoch: $epoch"
Write-Host "  - Time Step: $timeStep"
Write-Host "  - Period: ${Period}s"

Write-Host ""
Write-Host "🔑 TOTP Configuration:" -ForegroundColor Yellow
Write-Host "  - Secret: $Secret"
Write-Host "  - Digits: $Digits"
Write-Host "  - Algorithm: SHA1"

# Generate TOTP code using PowerShell
# Note: This is a simplified implementation for testing
# In production, use a proper TOTP library

function Convert-Base32ToBytes {
    param([string]$Base32)
    
    $Base32 = $Base32.ToUpper()
    $Base32Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
    $bytes = @()
    
    $bits = 0
    $buffer = 0
    
    foreach ($char in $Base32.ToCharArray()) {
        $index = $Base32Chars.IndexOf($char)
        if ($index -eq -1) { continue }
        
        $buffer = ($buffer -shl 5) -bor $index
        $bits += 5
        
        if ($bits -ge 8) {
            $bytes += [byte](($buffer -shr ($bits - 8)) -band 0xFF)
            $bits -= 8
        }
    }
    
    return $bytes
}

function Generate-TOTP {
    param(
        [string]$Secret,
        [long]$TimeStep,
        [int]$Digits = 6
    )
    
    # Convert base32 secret to bytes
    $key = Convert-Base32ToBytes -Base32 $Secret
    
    # Convert time step to bytes (big-endian)
    $timeBytes = @()
    for ($i = 7; $i -ge 0; $i--) {
        $timeBytes += [byte](($TimeStep -shr ($i * 8)) -band 0xFF)
    }
    
    # Generate HMAC-SHA1
    $hmac = New-Object System.Security.Cryptography.HMACSHA1
    $hmac.Key = $key
    $hash = $hmac.ComputeHash($timeBytes)
    
    # Generate TOTP code
    $offset = $hash[-1] -band 0x0F
    $code = (($hash[$offset] -band 0x7F) -shl 24) -bor
            (($hash[$offset + 1] -band 0xFF) -shl 16) -bor
            (($hash[$offset + 2] -band 0xFF) -shl 8) -bor
            ($hash[$offset + 3] -band 0xFF)
    
    $code = $code % [math]::Pow(10, $Digits)
    return $code.ToString().PadLeft($Digits, '0')
}

# Generate current TOTP code
$currentCode = Generate-TOTP -Secret $Secret -TimeStep $timeStep -Digits $Digits

Write-Host ""
Write-Host "🎯 Current TOTP Code:" -ForegroundColor Green
Write-Host "  $currentCode" -ForegroundColor White -BackgroundColor Green

# Generate codes for adjacent time steps
$prevTimeStep = $timeStep - 1
$nextTimeStep = $timeStep + 1

$prevCode = Generate-TOTP -Secret $Secret -TimeStep $prevTimeStep -Digits $Digits
$nextCode = Generate-TOTP -Secret $Secret -TimeStep $nextTimeStep -Digits $Digits

Write-Host ""
Write-Host "⏰ Adjacent Time Steps (for testing):" -ForegroundColor Yellow
Write-Host "  Previous (-30s): $prevCode"
Write-Host "  Current:         $currentCode"
Write-Host "  Next (+30s):     $nextCode"

Write-Host ""
Write-Host "💡 Usage:" -ForegroundColor Cyan
Write-Host "  Use the 'Current' code for immediate testing"
Write-Host "  Use 'Previous' or 'Next' codes to test time window tolerance"
Write-Host ""
Write-Host "  Example curl command:"
Write-Host "  curl -X POST http://localhost:8080/api/auth/login \"
Write-Host "    -H 'Content-Type: application/json' \"
Write-Host "    -d '{\"email\":\"test@securesystem.email\",\"password\":\"Test123!@#\",\"totp_code\":\"$currentCode\"}'"
