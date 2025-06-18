#!/bin/bash

# setup_cloudflare.sh: Configure Cloudflare DNS and R2
set -e

# Check if required environment variables are set
if [ -z "$CLOUDFLARE_API_KEY" ] || [ -z "$CLOUDFLARE_ZONE_ID" ] || [ -z "$DOMAIN_NAME" ]; then
    echo "Error: Required environment variables not set"
    echo "Please set CLOUDFLARE_API_KEY, CLOUDFLARE_ZONE_ID, and DOMAIN_NAME in .env"
    exit 1
fi

# Function to make Cloudflare API calls
cloudflare_api() {
    local endpoint=$1
    local method=${2:-GET}
    local data=${3:-""}
    
    curl -s -X "$method" \
        "https://api.cloudflare.com/client/v4/$endpoint" \
        -H "Authorization: Bearer $CLOUDFLARE_API_KEY" \
        -H "Content-Type: application/json" \
        ${data:+-d "$data"}
}

echo "Setting up Cloudflare configuration..."

# 1. Configure DNS A records for VMs
echo "Configuring DNS A records..."
for vm in 1 2; do
    vm_ip=$(grep "VM${vm}_IP" .env | cut -d'=' -f2)
    if [ -n "$vm_ip" ]; then
        # Create A record for each VM
        data="{
            \"type\": \"A\",
            \"name\": \"api${vm}.$DOMAIN_NAME\",
            \"content\": \"$vm_ip\",
            \"ttl\": 1,
            \"proxied\": true
        }"
        
        cloudflare_api "zones/$CLOUDFLARE_ZONE_ID/dns_records" "POST" "$data"
        echo "Created A record for api${vm}.$DOMAIN_NAME"
    else
        echo "Warning: VM${vm}_IP not found in .env"
    fi
done

# 2. Configure SSL/TLS settings
echo "Configuring SSL/TLS settings..."
data="{
    \"value\": \"strict\",
    \"enabled\": true
}"
cloudflare_api "zones/$CLOUDFLARE_ZONE_ID/settings/ssl" "PATCH" "$data"

# 3. Configure HSTS
echo "Configuring HSTS..."
data="{
    \"enabled\": true,
    \"max_age\": 31536000,
    \"include_subdomains\": true,
    \"preload\": true
}"
cloudflare_api "zones/$CLOUDFLARE_ZONE_ID/settings/security_header" "PATCH" "$data"

# 4. Create R2 bucket if not exists
echo "Creating R2 bucket..."
if [ -z "$R2_BUCKET_NAME" ]; then
    echo "Error: R2_BUCKET_NAME not set in .env"
    exit 1
fi

# Check if bucket exists
bucket_exists=$(cloudflare_api "accounts/$CLOUDFLARE_ACCOUNT_ID/r2/buckets/$R2_BUCKET_NAME" | grep -q "not found" && echo "false" || echo "true")

if [ "$bucket_exists" = "false" ]; then
    data="{
        \"name\": \"$R2_BUCKET_NAME\"
    }"
    cloudflare_api "accounts/$CLOUDFLARE_ACCOUNT_ID/r2/buckets" "POST" "$data"
    echo "Created R2 bucket: $R2_BUCKET_NAME"
else
    echo "R2 bucket $R2_BUCKET_NAME already exists"
fi

# 5. Configure CORS for R2 bucket
echo "Configuring CORS for R2 bucket..."
data="{
    \"cors_rules\": [
        {
            \"allowed_origins\": [\"https://$DOMAIN_NAME\"],
            \"allowed_methods\": [\"GET\", \"PUT\", \"POST\", \"DELETE\"],
            \"allowed_headers\": [\"*\"],
            \"max_age_seconds\": 3600
        }
    ]
}"
cloudflare_api "accounts/$CLOUDFLARE_ACCOUNT_ID/r2/buckets/$R2_BUCKET_NAME/cors" "PUT" "$data"

echo "Cloudflare setup completed successfully!"
echo "Please verify the following:"
echo "1. DNS A records for api1.$DOMAIN_NAME and api2.$DOMAIN_NAME"
echo "2. SSL/TLS mode is set to Strict"
echo "3. HSTS is enabled"
echo "4. R2 bucket $R2_BUCKET_NAME is created and accessible" 