#!/bin/bash
# setup_vm.sh: Copy project to VM and run setup

VM_IP="132.226.107.182"
SSH_KEY="vm_connection/private_key.txt"
PROJECT_DIR="secure-email-mvp"

echo "Setting up VM at $VM_IP..."

# Set proper permissions for SSH key
chmod 600 "$SSH_KEY"

# Copy project files to VM
echo "Copying project files to VM..."
scp -i "$SSH_KEY" -r . opc@"$VM_IP":~/"$PROJECT_DIR"

# Connect to VM and run setup
echo "Connecting to VM and running setup..."
ssh -i "$SSH_KEY" opc@"$VM_IP" << 'EOF'
cd secure-email-mvp
chmod +x setup_api.sh
./setup_api.sh
EOF

echo "VM setup complete!" 