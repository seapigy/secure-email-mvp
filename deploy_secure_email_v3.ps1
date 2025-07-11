# deploy_secure_email_v3.ps1: Deploy to Oracle Cloud VM instance secure-email-api-v3
param(
    [Parameter(Mandatory=$true)]
    [string]$VM_IP,
    [string]$VM_USER = "opc",
    [string]$SSH_KEY = ".\ssh-key-v3.key"
)

Write-Host "=== Secure Email MVP v3 Deployment ===" -ForegroundColor Green
Write-Host "Target VM: secure-email-api-v3" -ForegroundColor Cyan
Write-Host "Deploying to: $VM_USER@$VM_IP" -ForegroundColor Yellow
Write-Host "SSH Key: $SSH_KEY" -ForegroundColor Yellow

# Check if SSH key exists
if (-not (Test-Path $SSH_KEY)) {
    Write-Host "ERROR: SSH key not found at $SSH_KEY" -ForegroundColor Red
    exit 1
}

# Test SSH connection first
Write-Host "Testing SSH connection..." -ForegroundColor Yellow
$testCmd = "ssh -i `"$SSH_KEY`" $($VM_USER)@$($VM_IP) 'echo SSH connection successful'"
try {
    Invoke-Expression $testCmd
    if ($LASTEXITCODE -eq 0) {
        Write-Host "SSH connection successful!" -ForegroundColor Green
    } else {
        Write-Host "ERROR: SSH connection failed" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "ERROR: SSH connection failed: $_" -ForegroundColor Red
    exit 1
}

# Create project directory on VM
Write-Host "Creating project directory on VM..." -ForegroundColor Yellow
$createDirCmd = "ssh -i `"$SSH_KEY`" $($VM_USER)@$($VM_IP) 'sudo mkdir -p /home/opc/secure-email-mvp && sudo chown opc:opc /home/opc/secure-email-mvp'"
Invoke-Expression $createDirCmd

# Copy files individually
Write-Host "Copying files to VM..." -ForegroundColor Yellow

$filesToCopy = @(
    "cmd/",
    "pkg/", 
    "schema/",
    "go.mod",
    "go.sum",
    "api-server-linux",
    "env.example",
    "setup_api.sh",
    "setup_cloudflare.sh",
    "README.md",
    "vm_startup.sh",
    "vm_directory_check.sh"
)

foreach ($file in $filesToCopy) {
    if (Test-Path $file) {
        Write-Host "Copying $file..." -ForegroundColor Gray
        $scpCmd = "scp -i `"$SSH_KEY`" -r $file $($VM_USER)@$($VM_IP):/home/opc/secure-email-mvp/"
        Invoke-Expression $scpCmd
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Warning: Failed to copy $($file)" -ForegroundColor Yellow
        }
    } else {
        Write-Host "Warning: $($file) not found" -ForegroundColor Yellow
    }
}

# Set up environment on VM
Write-Host "Setting up environment on VM..." -ForegroundColor Yellow
$setupCmd = @"
ssh -i `"$SSH_KEY`" $($VM_USER)@$($VM_IP) '
set -e
cd /home/opc/secure-email-mvp

# Create necessary directories
sudo mkdir -p /var/db
sudo mkdir -p /var/log
sudo chown opc:opc /var/db
sudo chown opc:opc /var/log

# Rename Linux binary to api-server
mv api-server-linux api-server

# Set up .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from template..."
    cp env.example .env
    echo "Please update .env with your Cloudflare R2 credentials"
fi

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy

# Make scripts executable
chmod +x api-server
chmod +x vm_startup.sh
chmod +x vm_directory_check.sh

echo "=== Deployment Complete ==="
echo "Next steps:"
echo "1. Update .env with real credentials"
echo "2. Run: ./vm_startup.sh"
'
"@

Invoke-Expression $setupCmd

# Test Cloudflare R2 connectivity
Write-Host "Testing Cloudflare R2 connectivity..." -ForegroundColor Yellow
$r2TestCmd = @"
ssh -i `"$SSH_KEY`" $($VM_USER)@$($VM_IP) '
cd /home/opc/secure-email-mvp
echo "Testing Cloudflare R2 connectivity..."
# This will be tested when .env is properly configured
echo "R2 connectivity test ready - run after .env is configured"
'
"@

Invoke-Expression $r2TestCmd

Write-Host "=== Deployment Complete ===" -ForegroundColor Green
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. SSH to VM: ssh -i `"$SSH_KEY`" $($VM_USER)@$($VM_IP)" -ForegroundColor Cyan
Write-Host "2. Update .env: cd /home/opc/secure-email-mvp && nano .env" -ForegroundColor Cyan
Write-Host "3. Start server: ./vm_startup.sh" -ForegroundColor Cyan
Write-Host "4. Test R2 connectivity: ./setup_cloudflare.sh" -ForegroundColor Cyan 