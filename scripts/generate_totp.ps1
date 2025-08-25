# TOTP Code Generator for Testing
# This script generates TOTP codes for known test secrets

param(
    [string]$Secret = "JBSWY3DPEHPK3PXP",  # Default test user secret
    [int]$Steps = 3  # Number of time steps to generate (current + previous/next)
)

# Function to generate TOTP code using PowerShell
function Get-TOTPCode {
    param([string]$Secret, [long]$TimeStep)
    
    # Base32 decode the secret
    $base32Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
    $bits = ""
    foreach ($char in $Secret.ToUpper().ToCharArray()) {
        $index = $base32Chars.IndexOf($char)
        if ($index -ge 0) {
            $bits += [Convert]::ToString($index, 2).PadLeft(5, '0')
        }
    }
    
    # Convert bits to bytes
    $bytes = @()
    for ($i = 0; $i -lt $bits.Length; $i += 8) {
        if ($i + 8 -le $bits.Length) {
            $byte = [Convert]::ToByte($bits.Substring($i, 8), 2)
            $bytes += $byte
        }
    }
    
    # Create HMAC-SHA1
    $hmac = New-Object System.Security.Cryptography.HMACSHA1
    $hmac.Key = $bytes
    
    # Convert time step to bytes (big-endian)
    $timeBytes = [BitConverter]::GetBytes($TimeStep)
    if ([BitConverter]::IsLittleEndian) {
        [Array]::Reverse($timeBytes)
    }
    
    # Generate HMAC
    $hash = $hmac.ComputeHash($timeBytes)
    
    # Generate 6-digit code
    $offset = $hash[-1] -band 0x0F
    $code = (($hash[$offset] -band 0x7F) -shl 24) -bor (($hash[$offset + 1] -band 0xFF) -shl 16) -bor (($hash[$offset + 2] -band 0xFF) -shl 8) -bor ($hash[$offset + 3] -band 0xFF)
    $code = $code % 1000000
    
    return $code.ToString("D6")
}

# Get current Unix timestamp (30-second intervals)
$currentTime = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$timeStep = [Math]::Floor($currentTime / 30)

Write-Host "TOTP Code Generator for Secret: $Secret" -ForegroundColor Cyan
Write-Host "Current Time Step: $timeStep" -ForegroundColor Yellow
Write-Host ""

# Generate codes for current and surrounding time steps
for ($i = -$Steps; $i -le $Steps; $i++) {
    $stepTime = $timeStep + $i
    $code = Get-TOTPCode -Secret $Secret -TimeStep $stepTime
    $stepLabel = if ($i -eq 0) { "CURRENT" } elseif ($i -lt 0) { "PREVIOUS" } else { "NEXT" }
    $timeLabel = if ($i -eq 0) { "" } else { " (step $i)" }
    
    Write-Host "$stepLabel$timeLabel`: $code" -ForegroundColor $(if ($i -eq 0) { "Green" } else { "Gray" })
}

Write-Host ""
Write-Host "Usage in test scripts:" -ForegroundColor Magenta
Write-Host "  `$totpCode = & `"$PSScriptRoot\generate_totp.ps1`" -Secret `"$Secret`"" -ForegroundColor White
Write-Host "  `$currentCode = `$totpCode.Split(`"`n`") | Where-Object { `$_ -match 'CURRENT:' } | ForEach-Object { `$_.Split(':')[1].Trim() }" -ForegroundColor White
