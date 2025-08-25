# =============================================================================
# Server Configuration for Postfix + SES Setup
# =============================================================================
# Edit this file with your Oracle Linux server details
# This file is for convenience - you can also provide credentials directly

# Oracle Linux Server Configuration
$ServerConfig = @{
    # Replace with your actual server details
    SshHost = "YOUR_SERVER_IP_HERE"           # e.g., "192.168.1.100" or "your-server.com"
    SshUser = "YOUR_SSH_USERNAME_HERE"        # e.g., "oracle", "root", or your custom username
    SshPort = "22"                            # Default SSH port (change if different)
    SshKeyPath = $null                        # Path to SSH key if using key-based auth (optional)
}

# Function to validate configuration
function Test-ServerConfig {
    if ($ServerConfig.SshHost -eq "YOUR_SERVER_IP_HERE" -or $ServerConfig.SshUser -eq "YOUR_SSH_USERNAME_HERE") {
        Write-Host "❌ Please edit this file with your actual server details!" -ForegroundColor Red
        Write-Host "   - Replace YOUR_SERVER_IP_HERE with your server IP" -ForegroundColor Yellow
        Write-Host "   - Replace YOUR_SSH_USERNAME_HERE with your SSH username" -ForegroundColor Yellow
        return $false
    }
    return $true
}

# Function to run Postfix setup
function Start-PostfixSetup {
    if (-not (Test-ServerConfig)) {
        return
    }
    
    Write-Host "🚀 Starting Postfix + SES setup..." -ForegroundColor Green
    Write-Host "   Server: $($ServerConfig.SshUser)@$($ServerConfig.SshHost)" -ForegroundColor Cyan
    
    # Run the setup script
    $scriptPath = Join-Path $PSScriptRoot "setup_postfix_ses.ps1"
    & $scriptPath -SshHost $ServerConfig.SshHost -SshUser $ServerConfig.SshUser -SshPort $ServerConfig.SshPort -SshKeyPath $ServerConfig.SshKeyPath
}

# Export the configuration for use in other scripts
Export-ModuleMember -Variable ServerConfig -Function Test-ServerConfig, Start-PostfixSetup


