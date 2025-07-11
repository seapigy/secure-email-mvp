#!/bin/bash
# connect.sh: Connect to Oracle Cloud VM using project SSH keys

VM_IP="132.226.107.182"
SSH_KEY="vm_connection/private_key.txt"

echo "Connecting to VM at $VM_IP..."
echo "Using SSH key: $SSH_KEY"

# Set proper permissions for SSH key
chmod 600 "$SSH_KEY"

# Connect to VM
ssh -i "$SSH_KEY" opc@"$VM_IP" 