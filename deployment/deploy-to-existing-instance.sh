#!/bin/bash

# Deploy Secure Email Backend to Existing Oracle Cloud Instance
# Usage: ./deploy-to-existing-instance.sh

set -e

# Configuration
INSTANCE_IP="129.146.245.203"
INSTANCE_USER="opc"
INSTANCE_OCID="ocid1.instance.oc1.phx.anyhqljt5omx3rycfotxxyrqmwosukmo5xop6dqtbmih3rkcdxg6zbxu2j7a"
REGION="us-phoenix-1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log "Starting deployment to existing Oracle Cloud instance..."
log "Instance IP: $INSTANCE_IP"
log "Instance OCID: $INSTANCE_OCID"
log "Region: $REGION"

# Check if SSH key exists
if [ ! -f ~/.ssh/id_rsa ] && [ ! -f ~/.ssh/id_ed25519 ]; then
    warning "No SSH key found. You'll need to:"
    echo "1. Generate an SSH key: ssh-keygen -t ed25519 -C 'your-email@example.com'"
    echo "2. Add the public key to your OCI instance"
    echo "3. Or use the OCI Console to connect via browser"
    exit 1
fi

# Test SSH connection
log "Testing SSH connection..."
if ssh -o ConnectTimeout=10 -o BatchMode=yes $INSTANCE_USER@$INSTANCE_IP "echo 'SSH connection successful'" 2>/dev/null; then
    success "SSH connection successful"
else
    error "SSH connection failed. Please ensure:"
    echo "1. Your SSH key is added to the instance"
    echo "2. The instance is running"
    echo "3. Security groups allow SSH (port 22)"
    echo ""
    echo "You can also connect via OCI Console:"
    echo "1. Go to Compute → Instances"
    echo "2. Click on your instance"
    echo "3. Click 'Connect' → 'Cloud Shell Connection'"
fi

# Create deployment script for the instance
log "Creating deployment script for the instance..."
cat > /tmp/instance-setup.sh << 'EOF'
#!/bin/bash

# Instance Setup Script
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log "Starting instance setup..."

# Update system
log "Updating system packages..."
sudo yum update -y

# Install Docker
log "Installing Docker..."
sudo yum install -y docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker opc

# Install Docker Compose
log "Installing Docker Compose..."
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install additional tools
log "Installing additional tools..."
sudo yum install -y git curl wget

# Create application directory
log "Creating application directory..."
sudo mkdir -p /opt/secure-email
sudo chown opc:opc /opt/secure-email
cd /opt/secure-email

# Clone the repository
log "Cloning repository..."
git clone https://github.com/seapigy/secure-email-mvp.git .

# Create production environment file
log "Creating environment configuration..."
cat > .env << 'ENVEOF'
# Database Configuration
DB_DSN=mysql://secureuser:securepass@tcp(localhost:3306)/securesystem?parseTime=true

# Email Configuration (UPDATE THESE)
SMTP_HOST=CHANGE_ME
SMTP_PORT=587
SMTP_USER=CHANGE_ME
SMTP_PASS=CHANGE_ME
EMAIL_FROM=no-reply@securesystem.email
FRONTEND_URL=https://129.146.245.203

# Token & Security Configuration
RECOVERY_TOKEN_EXP_DAYS=7
VERIFICATION_TOKEN_EXP_HOURS=24

# Argon2 Password Hashing Configuration
ARGON2_MEMORY_KB=131072
ARGON2_ITERATIONS=4
ARGON2_PARALLELISM=4
ARGON2_SALT_LEN=16
ARGON2_KEY_LEN=32

# Server Configuration
PORT=8080
HOST=0.0.0.0

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Environment
ENV=production
DEBUG=false
ENVEOF

# Set up firewall
log "Configuring firewall..."
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# Create systemd service
log "Creating systemd service..."
sudo tee /etc/systemd/system/secure-email-backend.service > /dev/null << 'SERVICEEOF'
[Unit]
Description=Secure Email Backend
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/secure-email
ExecStart=/usr/local/bin/docker-compose -f docker-compose.prod.yml up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose.prod.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
SERVICEEOF

sudo systemctl daemon-reload
sudo systemctl enable secure-email-backend.service

# Start the services
log "Starting services..."
sudo systemctl start secure-email-backend.service

# Wait for services to be ready
log "Waiting for services to start..."
sleep 30

# Health check
log "Performing health check..."
if curl -f -s http://localhost:8080/health >/dev/null; then
    success "Backend is healthy"
else
    warning "Backend health check failed, checking logs..."
    docker-compose -f docker-compose.prod.yml logs
fi

success "Instance setup completed!"
log "Your application should be available at:"
log "  - Backend API: http://129.146.245.203:8080"
log "  - Health check: http://129.146.245.203:8080/health"
log "  - Signup endpoint: http://129.146.245.203:8080/api/signup"

log "Next steps:"
log "1. Configure SMTP settings in /opt/secure-email/.env"
log "2. Set up SSL/TLS with a reverse proxy (nginx)"
log "3. Configure your domain to point to this IP"
log "4. Test the signup endpoint"
EOF

# Copy and run the setup script on the instance
log "Copying setup script to instance..."
scp /tmp/instance-setup.sh $INSTANCE_USER@$INSTANCE_IP:/tmp/

log "Running setup script on instance..."
ssh $INSTANCE_USER@$INSTANCE_IP "chmod +x /tmp/instance-setup.sh && /tmp/instance-setup.sh"

# Test the deployment
log "Testing deployment..."
sleep 10

if curl -f -s http://$INSTANCE_IP:8080/health >/dev/null; then
    success "Deployment successful!"
    log ""
    log "🎉 Your Secure Email Backend is now running!"
    log ""
    log "Access URLs:"
    log "  - Health check: http://$INSTANCE_IP:8080/health"
    log "  - Signup API: http://$INSTANCE_IP:8080/api/signup"
    log ""
    log "Test the signup endpoint:"
    log "curl -X POST http://$INSTANCE_IP:8080/api/signup \\"
    log "  -H 'Content-Type: application/json' \\"
    log "  -d '{\"email\":\"test@securesystem.email\",\"password\":\"password123\",\"tier\":\"free\"}'"
    log ""
    log "Next steps:"
    log "1. Configure SMTP settings in /opt/secure-email/.env on the instance"
    log "2. Set up a reverse proxy (nginx) for SSL/TLS"
    log "3. Configure your domain to point to $INSTANCE_IP"
    log "4. Set up monitoring and backups"
else
    warning "Deployment test failed. Check the instance logs:"
    log "ssh $INSTANCE_USER@$INSTANCE_IP 'cd /opt/secure-email && docker-compose -f docker-compose.prod.yml logs'"
fi

# Clean up
rm -f /tmp/instance-setup.sh

success "Deployment process completed!"
