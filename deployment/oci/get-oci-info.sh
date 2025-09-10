#!/bin/bash

# Script to help gather OCI configuration information
# Run this script to get the information needed for ~/.oci/config

echo "=== Oracle Cloud Infrastructure Configuration Helper ==="
echo ""

# Check if OCI CLI is installed
if ! command -v oci &> /dev/null; then
    echo "❌ OCI CLI is not installed. Please install it first:"
    echo "   https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm"
    exit 1
fi

echo "✅ OCI CLI is installed: $(oci --version)"
echo ""

# Check if config file exists
if [ -f ~/.oci/config ]; then
    echo "📁 Config file exists at: ~/.oci/config"
    echo ""
    echo "Current configuration:"
    echo "======================"
    cat ~/.oci/config
    echo ""
else
    echo "❌ Config file not found at: ~/.oci/config"
    echo ""
    echo "You need to create the config file. Here's what you need:"
    echo ""
    echo "1. User OCID: Get from OCI Console → Identity & Security → Users"
    echo "2. Tenancy OCID: Get from OCI Console → Administration → Tenancy Details"
    echo "3. Fingerprint: Get from OCI Console → Identity & Security → Users → API Keys"
    echo "4. Region: Your OCI region (e.g., us-ashburn-1, us-phoenix-1)"
    echo ""
    echo "Create ~/.oci/config with this template:"
    echo "========================================"
    cat << 'EOF'
[DEFAULT]
user=ocid1.user.oc1..aaaaaaaa...your-user-ocid-here
fingerprint=aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp
key_file=~/.oci/oci_api_key.pem
tenancy=ocid1.tenancy.oc1..aaaaaaaa...your-tenancy-ocid-here
region=us-ashburn-1
pass_phrase=
EOF
    echo ""
fi

# Check if private key exists
if [ -f ~/.oci/oci_api_key.pem ]; then
    echo "✅ Private key exists at: ~/.oci/oci_api_key.pem"
    
    # Get fingerprint
    echo ""
    echo "🔑 Your API key fingerprint:"
    openssl rsa -pubout -outform DER -in ~/.oci/oci_api_key.pem 2>/dev/null | openssl md5 -c 2>/dev/null || echo "Could not generate fingerprint"
else
    echo "❌ Private key not found at: ~/.oci/oci_api_key.pem"
    echo ""
    echo "Generate a new key pair:"
    echo "========================"
    echo "mkdir -p ~/.oci"
    echo "openssl genrsa -out ~/.oci/oci_api_key.pem 2048"
    echo "chmod 600 ~/.oci/oci_api_key.pem"
    echo "openssl rsa -pubout -in ~/.oci/oci_api_key.pem -out ~/.oci/oci_api_key_public.pem"
    echo ""
    echo "Then upload the public key to OCI Console → Identity & Security → Users → API Keys"
fi

echo ""
echo "=== GitHub Secrets Setup ==="
echo ""
echo "For the CI/CD pipeline, you need to set these GitHub secrets:"
echo ""
echo "1. OCI_CONFIG (base64 encoded config file):"
if [ -f ~/.oci/config ]; then
    echo "   cat ~/.oci/config | base64 -w 0"
    echo ""
    echo "   Current config (base64):"
    cat ~/.oci/config | base64 -w 0 2>/dev/null || cat ~/.oci/config | base64
else
    echo "   (Create config file first)"
fi

echo ""
echo "2. OCI_PRIVATE_KEY (base64 encoded private key):"
if [ -f ~/.oci/oci_api_key.pem ]; then
    echo "   cat ~/.oci/oci_api_key.pem | base64 -w 0"
    echo ""
    echo "   Current key (base64):"
    cat ~/.oci/oci_api_key.pem | base64 -w 0 2>/dev/null || cat ~/.oci/oci_api_key.pem | base64
else
    echo "   (Generate key pair first)"
fi

echo ""
echo "3. OCI_INSTANCE_ID: (Get after creating compute instance)"
echo "4. OCI_COMPARTMENT_ID: (Your compartment OCID)"

echo ""
echo "=== Test Configuration ==="
echo ""
if [ -f ~/.oci/config ] && [ -f ~/.oci/oci_api_key.pem ]; then
    echo "Testing OCI configuration..."
    if oci iam user list --limit 1 &>/dev/null; then
        echo "✅ OCI configuration is working!"
        
        echo ""
        echo "Current user info:"
        oci iam user get --user-id $(oci iam user list --query 'data[0].id' --raw-output) --query 'data.{name:name,id:id}' --output table 2>/dev/null || echo "Could not get user info"
        
        echo ""
        echo "Available compartments:"
        oci iam compartment list --query 'data[*].{name:name,id:id}' --output table 2>/dev/null || echo "Could not list compartments"
        
    else
        echo "❌ OCI configuration test failed. Check your credentials."
    fi
else
    echo "⚠️  Cannot test configuration - missing config file or private key"
fi

echo ""
echo "=== Next Steps ==="
echo ""
echo "1. Complete OCI CLI setup (if not done)"
echo "2. Set up GitHub secrets for CI/CD"
echo "3. Run: terraform init && terraform plan && terraform apply"
echo "4. Push to GitHub to trigger deployment"
