# Manual Deployment Guide for Existing Oracle Cloud Instance

## 🎯 **Your Instance Details**
- **Instance OCID**: `ocid1.instance.oc1.phx.anyhqljt5omx3rycfotxxyrqmwosukmo5xop6dqtbmih3rkcdxg6zbxu2j7a`
- **Public IP**: `129.146.245.203`
- **Username**: `opc`
- **Region**: `US West (Phoenix)` - `us-phoenix-1`

## 🚀 **Deployment Options**

### Option 1: Connect via OCI Console (Recommended)
1. Go to [OCI Console](https://cloud.oracle.com)
2. Navigate to **Compute** → **Instances**
3. Click on your instance
4. Click **Connect** → **Cloud Shell Connection**
5. This will open a browser-based terminal

### Option 2: Set up SSH Keys
1. Generate SSH key on your local machine:
   ```bash
   ssh-keygen -t ed25519 -C "your-email@example.com"
   ```
2. Copy the public key:
   ```bash
   cat ~/.ssh/id_ed25519.pub
   ```
3. Add the key to your OCI instance via the console

## 📋 **Step-by-Step Deployment**

### 1. Connect to Your Instance
Use either the OCI Console Cloud Shell or SSH.

### 2. Update System and Install Dependencies
```bash
# Update system
sudo yum update -y

# Install Docker
sudo yum install -y docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker opc

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install additional tools
sudo yum install -y git curl wget
```

### 3. Clone and Set Up Application
```bash
# Create application directory
sudo mkdir -p /opt/secure-email
sudo chown opc:opc /opt/secure-email
cd /opt/secure-email

# Clone the repository
git clone https://github.com/seapigy/secure-email-mvp.git .

# Create environment file
cat > .env << 'EOF'
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
EOF
```

### 4. Configure Firewall
```bash
# Open required ports
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

### 5. Create Systemd Service
```bash
# Create service file
sudo tee /etc/systemd/system/secure-email-backend.service > /dev/null << 'EOF'
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
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable secure-email-backend.service
sudo systemctl start secure-email-backend.service
```

### 6. Verify Deployment
```bash
# Wait for services to start
sleep 30

# Check health
curl http://localhost:8080/health

# Check logs if needed
cd /opt/secure-email
docker-compose -f docker-compose.prod.yml logs
```

## 🧪 **Test Your Deployment**

### Health Check
```bash
curl http://129.146.245.203:8080/health
```

### Test Signup Endpoint
```bash
curl -X POST http://129.146.245.203:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@securesystem.email",
    "password": "password123",
    "tier": "free"
  }'
```

## 🔧 **Configuration Updates**

### Update SMTP Settings
Edit `/opt/secure-email/.env` and update:
```bash
SMTP_HOST=your-smtp-host
SMTP_USER=your-smtp-username
SMTP_PASS=your-smtp-password
EMAIL_FROM=no-reply@yourdomain.com
```

### Update Frontend URL
```bash
FRONTEND_URL=https://yourdomain.com
```

## 🔒 **Security Setup**

### Set up Nginx Reverse Proxy
```bash
# Install nginx
sudo yum install -y nginx

# Create nginx config
sudo tee /etc/nginx/conf.d/secure-email.conf > /dev/null << 'EOF'
server {
    listen 80;
    server_name 129.146.245.203;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# Start nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

## 📊 **Monitoring**

### Check Service Status
```bash
sudo systemctl status secure-email-backend.service
```

### View Logs
```bash
# Application logs
cd /opt/secure-email
docker-compose -f docker-compose.prod.yml logs -f

# System logs
sudo journalctl -u secure-email-backend.service -f
```

### Restart Services
```bash
sudo systemctl restart secure-email-backend.service
```

## 🎯 **Expected Results**

After successful deployment, you should have:

✅ **Backend API running on port 8080**  
✅ **MySQL database with users table**  
✅ **Health endpoint responding**  
✅ **Signup endpoint accepting requests**  
✅ **Automatic startup on boot**  
✅ **Firewall configured**  

## 🆘 **Troubleshooting**

### If services won't start:
```bash
# Check Docker status
sudo systemctl status docker

# Check logs
cd /opt/secure-email
docker-compose -f docker-compose.prod.yml logs

# Restart Docker
sudo systemctl restart docker
```

### If port 8080 is not accessible:
```bash
# Check firewall
sudo firewall-cmd --list-ports

# Check if service is listening
sudo netstat -tlnp | grep 8080
```

### If database connection fails:
```bash
# Check database container
docker ps | grep mysql

# Check database logs
docker logs securechat-email-db-1
```

## 🎉 **Success!**

Once everything is working, your Secure Email Backend will be available at:
- **Health Check**: http://129.146.245.203:8080/health
- **Signup API**: http://129.146.245.203:8080/api/signup

Next steps:
1. Configure your domain to point to `129.146.245.203`
2. Set up SSL/TLS certificates
3. Configure SMTP for email sending
4. Set up monitoring and backups
