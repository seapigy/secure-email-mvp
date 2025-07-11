#!/bin/bash

# deploy_to_vm.sh: Deploy updated backend code to Linux VM
set -e

echo "=== Secure Email MVP Backend Deployment ==="

# Configuration
VM_USER="opc"
VM_IP="129.146.68.127"  # Update with your actual VM IP
PROJECT_DIR="/home/opc/secure-email-mvp"

echo "Deploying to VM: $VM_USER@$VM_IP"
echo "Target directory: $PROJECT_DIR"

# Create deployment package
echo "Creating deployment package..."
tar -czf secure-email-backend.tar.gz \
    cmd/ \
    pkg/ \
    schema/ \
    go.mod \
    go.sum \
    api-server \
    env.example \
    setup_api.sh \
    setup_cloudflare.sh \
    README.md

# Copy to VM
echo "Copying files to VM..."
scp -i "C:/Users/cpigu/OneDrive/Desktop/.ssh/ssh-key-2025-07-07.key" secure-email-backend.tar.gz $VM_USER@$VM_IP:/tmp/

# Execute deployment commands on VM
echo "Executing deployment on VM..."
ssh -i "C:/Users/cpigu/OneDrive/Desktop/.ssh/ssh-key-2025-07-07.key" $VM_USER@$VM_IP << 'EOF'
set -e

echo "=== VM Deployment Script ==="

# Create project directory
sudo mkdir -p /home/opc/secure-email-mvp
sudo chown opc:opc /home/opc/secure-email-mvp

# Extract deployment package
cd /home/opc/secure-email-mvp
tar -xzf /tmp/secure-email-backend.tar.gz

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
    echo "Required variables:"
    echo "  CLOUDFLARE_R2_ACCESS_KEY"
    echo "  CLOUDFLARE_R2_SECRET_KEY"
    echo "  CLOUDFLARE_R2_BUCKET"
    echo "  CLOUDFLARE_R2_ENDPOINT"
    echo "  JWT_SECRET"
fi

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy

# Build the server
echo "Building API server..."
go build -o api-server cmd/api/main.go

# Make executable
chmod +x api-server

echo "=== Deployment Complete ==="
echo "Next steps:"
echo "1. Update .env with real credentials"
echo "2. Run: ./api-server"
echo "3. Check logs for any errors"
EOF

# Clean up local files
rm secure-email-backend.tar.gz

echo "=== Deployment Complete ==="
echo "Please SSH to the VM and:"
echo "1. Update /home/opc/secure-email-mvp/.env with real R2 credentials"
echo "2. Run: cd /home/opc/secure-email-mvp && ./api-server"
echo "3. Monitor the detailed startup logs" 