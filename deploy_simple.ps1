# deploy_simple.ps1: Simple deployment to Oracle Linux VM
param(
    [string]$VM_IP = "129.146.68.127",
    [string]$VM_USER = "opc",
    [string]$SSH_KEY = "C:\Users\cpigu\OneDrive\Desktop\.ssh\ssh-key-2025-07-07.key"
)

Write-Host "=== Secure Email MVP Deployment ===" -ForegroundColor Green
Write-Host "Target: $VM_USER@$VM_IP" -ForegroundColor Yellow

# Check SSH key
if (-not (Test-Path $SSH_KEY)) {
    Write-Host "ERROR: SSH key not found at $SSH_KEY" -ForegroundColor Red
    exit 1
}

# Test SSH connection first
Write-Host "Testing SSH connection..." -ForegroundColor Yellow
$testCmd = "ssh -i `"$SSH_KEY`" $VM_USER@$VM_IP 'echo SSH connection successful'"
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
$createDirCmd = "ssh -i `"$SSH_KEY`" $VM_USER@$VM_IP 'sudo mkdir -p /home/opc/secure-email-mvp && sudo chown opc:opc /home/opc/secure-email-mvp'"
Invoke-Expression $createDirCmd

# Copy files individually (avoiding tar issues)
Write-Host "Copying files to VM..." -ForegroundColor Yellow

$filesToCopy = @(
    "cmd/",
    "pkg/", 
    "schema/",
    "go.mod",
    "go.sum",
    "api-server",
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
        $scpCmd = "scp -i `"$SSH_KEY`" -r $file $VM_USER@$VM_IP:/home/opc/secure-email-mvp/"
        Invoke-Expression $scpCmd
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Warning: Failed to copy $file" -ForegroundColor Yellow
        }
    } else {
        Write-Host "Warning: $file not found" -ForegroundColor Yellow
    }
}

# Set up environment on VM
Write-Host "Setting up environment on VM..." -ForegroundColor Yellow
$setupCmd = @"
ssh -i `"$SSH_KEY`" $VM_USER@$VM_IP '
set -e
cd /home/opc/secure-email-mvp

# Create necessary directories
sudo mkdir -p /var/db
sudo mkdir -p /var/log
sudo chown opc:opc /var/db
sudo chown opc:opc /var/log

# Set up .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from template..."
    cp env.example .env
    echo "Please update .env with your Cloudflare R2 credentials"
fi

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy

# Build the server
echo "Building API server..."
go build -o api-server cmd/api/main.go

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

Write-Host "=== Deployment Complete ===" -ForegroundColor Green
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. SSH to VM: ssh -i `"$SSH_KEY`" $VM_USER@$VM_IP" -ForegroundColor Cyan
Write-Host "2. Update .env: cd /home/opc/secure-email-mvp && nano .env" -ForegroundColor Cyan
Write-Host "3. Start server: ./vm_startup.sh" -ForegroundColor Cyan 

#
# To run the Go API server in development mode, use:
#   go run ./cmd/api/*.go
#
# Do NOT use 'go run cmd/api/main.go' directly, as it will not include all necessary files.
#

go run cmd/api/main.go cmd/api/rate_limit.go cmd/api/login_handler.go cmd/api/signup_handler.go cmd/api/fallback_handler.go cmd/api/resend_fallback_handler.go

Let me know if you want to proceed or if you want to adjust the test setup! 