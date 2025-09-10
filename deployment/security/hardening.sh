#!/bin/bash

# Security Hardening Script for Secure Email Backend
set -e

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

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    error "Please run as root"
fi

log "Starting security hardening..."

# 1. System Updates
log "Updating system packages..."
apt-get update
apt-get upgrade -y
apt-get autoremove -y
apt-get autoclean

# 2. Install security tools
log "Installing security tools..."
apt-get install -y \
    fail2ban \
    ufw \
    unattended-upgrades \
    apt-listchanges \
    rkhunter \
    chkrootkit \
    lynis

# 3. Configure automatic security updates
log "Configuring automatic security updates..."
cat > /etc/apt/apt.conf.d/50unattended-upgrades << 'EOF'
Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}-security";
    "${distro_id}ESMApps:${distro_codename}-apps-security";
    "${distro_id}ESM:${distro_codename}-infra-security";
};

Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
EOF

cat > /etc/apt/apt.conf.d/20auto-upgrades << 'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF

# 4. Configure firewall
log "Configuring firewall..."
ufw --force reset
ufw default deny incoming
ufw default allow outgoing

# Allow SSH (be careful with this)
ufw allow 22/tcp

# Allow HTTP/HTTPS
ufw allow 80/tcp
ufw allow 443/tcp

# Allow specific IP ranges if needed (uncomment and modify)
# ufw allow from 192.168.1.0/24 to any port 22

ufw --force enable
success "Firewall configured"

# 5. Configure fail2ban
log "Configuring fail2ban..."
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3
backend = systemd

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 10

[nginx-botsearch]
enabled = true
filter = nginx-botsearch
port = http,https
logpath = /var/log/nginx/access.log
maxretry = 2
EOF

systemctl enable fail2ban
systemctl restart fail2ban
success "Fail2ban configured"

# 6. Secure SSH
log "Securing SSH configuration..."
cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup

cat > /etc/ssh/sshd_config << 'EOF'
# SSH Security Configuration
Port 22
Protocol 2
HostKey /etc/ssh/ssh_host_rsa_key
HostKey /etc/ssh/ssh_host_ecdsa_key
HostKey /etc/ssh/ssh_host_ed25519_key

# Authentication
LoginGraceTime 60
PermitRootLogin no
StrictModes yes
MaxAuthTries 3
MaxSessions 3

# Key-based authentication
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
PasswordAuthentication no
PermitEmptyPasswords no
ChallengeResponseAuthentication no

# Security settings
X11Forwarding no
X11DisplayOffset 10
PrintMotd no
PrintLastLog yes
TCPKeepAlive yes
ClientAliveInterval 300
ClientAliveCountMax 2

# Logging
SyslogFacility AUTH
LogLevel INFO

# Banner
Banner /etc/ssh/banner

# Allow specific users (modify as needed)
AllowUsers ubuntu

# Disable unused authentication methods
KerberosAuthentication no
GSSAPIAuthentication no
UsePAM yes
EOF

# Create SSH banner
cat > /etc/ssh/banner << 'EOF'
***************************************************************************
                    AUTHORIZED ACCESS ONLY
***************************************************************************
This system is for the use of authorized users only. Individuals using
this computer system without authority, or in excess of their authority,
are subject to having all of their activities on this system monitored
and recorded by system personnel.

In the course of monitoring individuals improperly using this system,
or in the course of system maintenance, the activities of authorized
users may also be monitored.

Anyone using this system expressly consents to such monitoring and is
advised that if such monitoring reveals possible evidence of criminal
activity, system personnel may provide the evidence of such monitoring
to law enforcement officials.
***************************************************************************
EOF

systemctl restart ssh
success "SSH secured"

# 7. Set up intrusion detection
log "Setting up intrusion detection..."
rkhunter --update
rkhunter --propupd

# Create rkhunter cron job
echo "0 2 * * * /usr/bin/rkhunter --cronjob --update --quiet" | crontab -

# 8. Configure system limits
log "Configuring system limits..."
cat >> /etc/security/limits.conf << 'EOF'
# Security limits
* soft nofile 65536
* hard nofile 65536
* soft nproc 32768
* hard nproc 32768
EOF

# 9. Disable unnecessary services
log "Disabling unnecessary services..."
systemctl disable bluetooth 2>/dev/null || true
systemctl disable cups 2>/dev/null || true
systemctl disable avahi-daemon 2>/dev/null || true
systemctl disable cups-browsed 2>/dev/null || true

# 10. Set up log monitoring
log "Setting up log monitoring..."
cat > /etc/logrotate.d/secure-logs << 'EOF'
/var/log/auth.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 640 root adm
}

/var/log/syslog {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 640 root adm
}

/var/log/nginx/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 640 www-data adm
}
EOF

# 11. Create security monitoring script
log "Creating security monitoring script..."
cat > /opt/secure-email/security-monitor.sh << 'EOF'
#!/bin/bash

# Security monitoring script
LOG_FILE="/opt/secure-email/logs/security.log"
ALERT_EMAIL="admin@securesystem.email"

# Create logs directory if it doesn't exist
mkdir -p /opt/secure-email/logs

# Check for failed login attempts
FAILED_LOGINS=$(grep "Failed password" /var/log/auth.log | grep "$(date '+%b %d')" | wc -l)
if [ "$FAILED_LOGINS" -gt 10 ]; then
    echo "$(date): High number of failed login attempts: $FAILED_LOGINS" >> "$LOG_FILE"
fi

# Check for root login attempts
ROOT_LOGINS=$(grep "root" /var/log/auth.log | grep "$(date '+%b %d')" | wc -l)
if [ "$ROOT_LOGINS" -gt 0 ]; then
    echo "$(date): Root login attempt detected" >> "$LOG_FILE"
fi

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 80 ]; then
    echo "$(date): High disk usage: ${DISK_USAGE}%" >> "$LOG_FILE"
fi

# Check memory usage
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
if [ "$MEMORY_USAGE" -gt 80 ]; then
    echo "$(date): High memory usage: ${MEMORY_USAGE}%" >> "$LOG_FILE"
fi
EOF

chmod +x /opt/secure-email/security-monitor.sh

# Add security monitoring cron job
echo "*/15 * * * * /opt/secure-email/security-monitor.sh" | crontab -

# 12. Set up file integrity monitoring
log "Setting up file integrity monitoring..."
apt-get install -y aide
aideinit
mv /var/lib/aide/aide.db.new /var/lib/aide/aide.db

# Create aide cron job
echo "0 3 * * * /usr/bin/aide --check" | crontab -

success "Security hardening completed!"

log "Security hardening summary:"
log "- System updated and unnecessary services disabled"
log "- Firewall configured (UFW)"
log "- Fail2ban configured for intrusion prevention"
log "- SSH secured with key-based authentication only"
log "- Intrusion detection tools installed (rkhunter, chkrootkit)"
log "- Log monitoring and rotation configured"
log "- File integrity monitoring set up (AIDE)"
log "- Security monitoring script created"

warning "Important security reminders:"
warning "- Ensure SSH keys are properly configured"
warning "- Test SSH access before disconnecting"
warning "- Regularly review security logs"
warning "- Keep system updated"
warning "- Monitor fail2ban status"
