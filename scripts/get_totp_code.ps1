# Get TOTP Code Utility
# Returns a valid TOTP code for a given user email

param(
    [Parameter(Mandatory=$true)]
    [string]$Email,
    [string]$DbPath = "C:\var\db\secure-email.db"
)

function Get-TOTPCode {
    param([string]$Secret)

    try {
        $totpCode = & .\totp_generator.exe $Secret
        return $totpCode.Trim()
    } catch {
        Write-Error "Failed to generate TOTP code: $($_.Exception.Message)"
        return $null
    }
}

# Get TOTP secret from database
$totpSecret = sqlite3 $DbPath "SELECT totp_secret FROM users WHERE email = '$Email';"

if (-not $totpSecret) {
    Write-Error "User not found or no TOTP secret available for email: $Email"
    exit 1
}

# Generate TOTP code
$totpCode = Get-TOTPCode $totpSecret

if (-not $totpCode) {
    Write-Error "Failed to generate TOTP code"
    exit 1
}

# Return the TOTP code
Write-Output $totpCode






















