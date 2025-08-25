# =============================================================================
# Postfix + SES Setup Launcher
# =============================================================================
# This script loads the server configuration and runs the Postfix setup

# Import the server configuration
$configPath = Join-Path $PSScriptRoot "server-config.ps1"
if (Test-Path $configPath) {
    . $configPath
    
    # Test configuration and run setup
    if (Test-ServerConfig) {
        Start-PostfixSetup
    } else {
        Write-Host "`n📝 To configure your server details:" -ForegroundColor Yellow
        Write-Host "   1. Edit: deploy\server-config.ps1" -ForegroundColor Cyan
        Write-Host "   2. Replace YOUR_SERVER_IP_HERE with your server IP" -ForegroundColor White
        Write-Host "   3. Replace YOUR_SSH_USERNAME_HERE with your SSH username" -ForegroundColor White
        Write-Host "   4. Run this script again" -ForegroundColor White
    }
} else {
    Write-Host "❌ Server configuration file not found!" -ForegroundColor Red
    Write-Host "   Please ensure deploy\server-config.ps1 exists" -ForegroundColor Yellow
}


