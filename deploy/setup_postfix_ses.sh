#!/bin/bash

# =============================================================================
# Postfix + Amazon SES Relay Configuration Script
# Oracle Linux 8/9 Compatible
# =============================================================================
# This script configures Postfix to relay all outgoing mail through Amazon SES
# Domain: securesystem.email
# SES Region: us-east-1
# =============================================================================

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   error "This script must be run as root (use sudo)"
fi

# Check if .env file exists and source it
if [[ ! -f ".env" ]]; then
    error ".env file not found. Please copy env.example to .env and configure your SES credentials."
fi

# Source environment variables
source .env

# Validate required environment variables
if [[ -z "$SES_SMTP_USERNAME" || -z "$SES_SMTP_PASSWORD" ]]; then
    error "SES_SMTP_USERNAME and SES_SMTP_PASSWORD must be set in .env file"
fi

if [[ -z "$SES_SMTP_HOST" || -z "$SES_SMTP_PORT" ]]; then
    error "SES_SMTP_HOST and SES_SMTP_PORT must be set in .env file"
fi

log "Starting Postfix + Amazon SES configuration..."

# =============================================================================
# STEP 1: Install and enable Postfix and required SASL libraries
# =============================================================================
log "Step 1: Installing Postfix and SASL libraries..."

# Update package cache
dnf update -y

# Install Postfix and SASL libraries
dnf install -y postfix cyrus-sasl cyrus-sasl-plain cyrus-sasl-md5

# Enable and start Postfix
systemctl enable postfix
systemctl start postfix

log "✓ Postfix installed and enabled"

# =============================================================================
# STEP 2: Configure SASL authentication
# =============================================================================
log "Step 2: Configuring SASL authentication..."

# Create SASL configuration directory
mkdir -p /etc/postfix/sasl

# Create SASL configuration file
cat > /etc/postfix/sasl/smtpd.conf << EOF
pwcheck_method: saslauthd
mech_list: PLAIN LOGIN
EOF

# Create SASL password file with secure permissions
cat > /etc/postfix/sasl_passwd << EOF
[$SES_SMTP_HOST]:$SES_SMTP_PORT $SES_SMTP_USERNAME:$SES_SMTP_PASSWORD
EOF

# Set secure permissions on SASL password file
chmod 600 /etc/postfix/sasl_passwd
chown root:root /etc/postfix/sasl_passwd

# Create Postfix lookup table
postmap /etc/postfix/sasl_passwd

log "✓ SASL authentication configured"

# =============================================================================
# STEP 3: Configure Postfix main.cf
# =============================================================================
log "Step 3: Configuring Postfix main.cf..."

# Backup original configuration
cp /etc/postfix/main.cf /etc/postfix/main.cf.backup.$(date +%Y%m%d_%H%M%S)

# Configure main.cf with SES settings
cat > /etc/postfix/main.cf << EOF
# =============================================================================
# Postfix Configuration for Amazon SES Relay
# =============================================================================

# Basic Postfix configuration
myhostname = mail.securesystem.email
mydomain = securesystem.email
myorigin = \$mydomain
inet_interfaces = loopback-only
inet_protocols = ipv4

# Relay configuration
relayhost = [$SES_SMTP_HOST]:$SES_SMTP_PORT
relay_domains = securesystem.email

# SASL authentication
smtp_sasl_auth_enable = yes
smtp_sasl_password_maps = hash:/etc/postfix/sasl_passwd
smtp_sasl_security_options = noanonymous
smtp_sasl_mechanism_filter = plain, login

# TLS configuration (AWS recommended settings)
smtp_tls_security_level = encrypt
smtp_tls_CAfile = /etc/ssl/certs/ca-bundle.crt
smtp_tls_session_cache_database = btree:/var/lib/postfix/smtp_tls_session_cache
smtp_tls_session_cache_timeout = 3600s
smtp_tls_loglevel = 1

# Rate limiting for SES
smtp_destination_concurrency_limit = 2
smtp_destination_rate_delay = 1s
smtp_extra_recipient_limit = 10

# Queue configuration
queue_directory = /var/spool/postfix
command_directory = /usr/sbin
daemon_directory = /usr/libexec/postfix
data_directory = /var/lib/postfix
mail_owner = postfix
myorigin = \$mydomain
mydestination = \$myhostname, localhost.\$mydomain, localhost, \$mydomain
local_recipient_maps =
local_transport = error:local delivery disabled

# Logging
maillog_file = /var/log/maillog
debug_peer_level = 2
debug_peer_list = 127.0.0.1

# Security settings
disable_vrfy_command = yes
strict_rfc821_envelopes = yes
receive_override_options = no_unknown_recipient_checks, no_header_body_checks

# Performance settings
default_process_limit = 100
smtp_client_restrictions = permit_mynetworks, permit_sasl_authenticated, reject
smtp_helo_required = yes
smtp_helo_name = \$myhostname

# Error handling
notify_classes = resource, software, bounce, delay, policy, protocol
delay_notice_recipient = postmaster
bounce_notice_recipient = postmaster
delay_warning_time = 4h
maximal_queue_lifetime = 5d
minimal_backoff_time = 300s
maximal_backoff_time = 4000s
EOF

log "✓ Postfix main.cf configured"

# =============================================================================
# STEP 4: Configure master.cf for submission
# =============================================================================
log "Step 4: Configuring master.cf..."

# Backup original master.cf
cp /etc/postfix/master.cf /etc/postfix/master.cf.backup.$(date +%Y%m%d_%H%M%S)

# Add submission service configuration
cat >> /etc/postfix/master.cf << EOF

# =============================================================================
# Submission service for authenticated clients
# =============================================================================
submission inet n       -       n       -       -       smtpd
  -o syslog_name=postfix/submission
  -o smtpd_tls_security_level=encrypt
  -o smtpd_sasl_auth_enable=yes
  -o smtpd_sasl_type=dovecot
  -o smtpd_sasl_path=private/auth
  -o smtpd_sasl_security_options=noanonymous,noplaintext
  -o smtpd_sasl_local_domain=\$myhostname
  -o smtpd_client_restrictions=permit_sasl_authenticated,reject
  -o smtpd_recipient_restrictions=reject_non_fqdn_recipient,permit_sasl_authenticated,reject
  -o smtpd_relay_restrictions=permit_sasl_authenticated,reject
  -o milter_macro_daemon_name=ORIGINATING
EOF

log "✓ Postfix master.cf configured"

# =============================================================================
# STEP 5: Create systemd service override for environment variables
# =============================================================================
log "Step 5: Creating systemd service override..."

# Create systemd override directory
mkdir -p /etc/systemd/system/postfix.service.d

# Create environment override file
cat > /etc/systemd/system/postfix.service.d/environment.conf << EOF
[Service]
Environment="SES_SMTP_USERNAME=$SES_SMTP_USERNAME"
Environment="SES_SMTP_PASSWORD=$SES_SMTP_PASSWORD"
Environment="SES_SMTP_HOST=$SES_SMTP_HOST"
Environment="SES_SMTP_PORT=$SES_SMTP_PORT"
EOF

# Reload systemd
systemctl daemon-reload

log "✓ Systemd service override created"

# =============================================================================
# STEP 6: Configure logging
# =============================================================================
log "Step 6: Configuring logging..."

# Create logrotate configuration for Postfix
cat > /etc/logrotate.d/postfix << EOF
/var/log/maillog {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        /usr/sbin/postfix reload > /dev/null 2>&1 || true
    endscript
}
EOF

log "✓ Logging configured"

# =============================================================================
# STEP 7: Test configuration and reload Postfix
# =============================================================================
log "Step 7: Testing configuration and reloading Postfix..."

# Test Postfix configuration
if postfix check; then
    log "✓ Postfix configuration is valid"
else
    error "Postfix configuration is invalid. Check the configuration files."
fi

# Reload Postfix
systemctl reload postfix

# Verify Postfix is running
if systemctl is-active --quiet postfix; then
    log "✓ Postfix is running"
else
    error "Postfix failed to start. Check logs with: journalctl -u postfix"
fi

# =============================================================================
# STEP 8: Create test script
# =============================================================================
log "Step 8: Creating test script..."

# Create test email script
cat > /usr/local/bin/test-ses-email.sh << 'EOF'
#!/bin/bash

# Test script for sending email through SES relay
# Usage: test-ses-email.sh <recipient_email> [subject] [body]

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <recipient_email> [subject] [body]"
    echo "Example: $0 test@example.com 'Test Subject' 'Test message body'"
    exit 1
fi

RECIPIENT="$1"
SUBJECT="${2:-Test email from Postfix + SES}"
BODY="${3:-This is a test email sent through Postfix relayed via Amazon SES.}"

# Create temporary email file
TEMP_EMAIL=$(mktemp)
cat > "$TEMP_EMAIL" << EMAIL
From: noreply@securesystem.email
To: $RECIPIENT
Subject: $SUBJECT
Date: $(date -R)
Content-Type: text/plain; charset=UTF-8

$BODY

---
Sent via Postfix + Amazon SES relay
EMAIL

# Send email
if sendmail -t < "$TEMP_EMAIL"; then
    echo "✓ Email sent successfully to $RECIPIENT"
    echo "Check Postfix logs: journalctl -u postfix -f"
else
    echo "✗ Failed to send email"
    echo "Check Postfix logs: journalctl -u postfix -f"
fi

# Clean up
rm -f "$TEMP_EMAIL"
EOF

chmod +x /usr/local/bin/test-ses-email.sh

log "✓ Test script created at /usr/local/bin/test-ses-email.sh"

# =============================================================================
# STEP 9: Create credential rotation script
# =============================================================================
log "Step 9: Creating credential rotation script..."

cat > /usr/local/bin/rotate-ses-credentials.sh << 'EOF'
#!/bin/bash

# SES Credential Rotation Script
# Usage: rotate-ses-credentials.sh <new_username> <new_password>

set -e

if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <new_username> <new_password>"
    echo "This script updates SES SMTP credentials and reloads Postfix"
    exit 1
fi

NEW_USERNAME="$1"
NEW_PASSWORD="$2"

echo "Rotating SES SMTP credentials..."

# Backup current configuration
cp /etc/postfix/sasl_passwd /etc/postfix/sasl_passwd.backup.$(date +%Y%m%d_%H%M%S)

# Update SASL password file
cat > /etc/postfix/sasl_passwd << EOF
[email-smtp.us-east-1.amazonaws.com]:587 $NEW_USERNAME:$NEW_PASSWORD
EOF

# Set secure permissions
chmod 600 /etc/postfix/sasl_passwd
chown root:root /etc/postfix/sasl_passwd

# Update Postfix lookup table
postmap /etc/postfix/sasl_passwd

# Reload Postfix
systemctl reload postfix

echo "✓ SES credentials rotated successfully"
echo "✓ Postfix reloaded"
echo "✓ Test with: test-ses-email.sh <recipient>"
EOF

chmod +x /usr/local/bin/rotate-ses-credentials.sh

log "✓ Credential rotation script created"

# =============================================================================
# STEP 10: Create monitoring script
# =============================================================================
log "Step 10: Creating monitoring script..."

cat > /usr/local/bin/monitor-postfix.sh << 'EOF'
#!/bin/bash

# Postfix Monitoring Script
# Shows queue status, recent logs, and connection status

echo "=== Postfix Status ==="
systemctl status postfix --no-pager -l

echo -e "\n=== Queue Status ==="
mailq

echo -e "\n=== Recent Logs (last 20 entries) ==="
journalctl -u postfix --no-pager -n 20

echo -e "\n=== Connection Status ==="
netstat -tlnp | grep :25
netstat -tlnp | grep :587

echo -e "\n=== SASL Status ==="
postconf -n | grep sasl

echo -e "\n=== Relay Host Status ==="
postconf relayhost
EOF

chmod +x /usr/local/bin/monitor-postfix.sh

log "✓ Monitoring script created"

# =============================================================================
# STEP 11: Final verification and summary
# =============================================================================
log "Step 11: Final verification..."

# Show final configuration summary
echo -e "\n${BLUE}=== Postfix + SES Configuration Summary ===${NC}"
echo "✓ Postfix installed and configured"
echo "✓ SASL authentication configured"
echo "✓ TLS encryption enabled"
echo "✓ Relay host: $SES_SMTP_HOST:$SES_SMTP_PORT"
echo "✓ Domain: securesystem.email"
echo "✓ Logs: journalctl -u postfix or /var/log/maillog"

echo -e "\n${BLUE}=== Available Commands ===${NC}"
echo "• Test email: /usr/local/bin/test-ses-email.sh <recipient>"
echo "• Monitor: /usr/local/bin/monitor-postfix.sh"
echo "• Rotate credentials: /usr/local/bin/rotate-ses-credentials.sh <user> <pass>"
echo "• View logs: journalctl -u postfix -f"
echo "• Check queue: mailq"

echo -e "\n${BLUE}=== Common Test Commands ===${NC}"
echo "• Test to verified email: /usr/local/bin/test-ses-email.sh your-verified@email.com"
echo "• Check Postfix status: systemctl status postfix"
echo "• View configuration: postconf -n"

echo -e "\n${YELLOW}=== Important Notes ===${NC}"
echo "• If using SES sandbox mode, recipient emails must be verified"
echo "• Check SES console for sending statistics and bounces"
echo "• Monitor logs for authentication or delivery issues"
echo "• Rate limits: $SES_SMTP_RATE_LIMIT emails per second"

echo -e "\n${GREEN}✓ Postfix + Amazon SES configuration completed successfully!${NC}"

# =============================================================================
# Troubleshooting section
# =============================================================================
echo -e "\n${BLUE}=== Troubleshooting Guide ===${NC}"
echo "If emails fail to send:"
echo "1. Check SES credentials: cat /etc/postfix/sasl_passwd"
echo "2. Verify SES domain verification: AWS SES Console"
echo "3. Check firewall: firewall-cmd --list-ports"
echo "4. Test connectivity: telnet $SES_SMTP_HOST $SES_SMTP_PORT"
echo "5. View detailed logs: journalctl -u postfix -f"
echo "6. Check queue: mailq"
echo "7. Flush queue if needed: postsuper -r ALL"

echo -e "\n${YELLOW}Common Issues:${NC}"
echo "• DNS resolution: Ensure $SES_SMTP_HOST resolves"
echo "• Port blocking: Ensure port 587 is open outbound"
echo "• Authentication: Verify SES SMTP credentials"
echo "• Sandbox mode: Recipients must be verified in SES"
echo "• Rate limiting: Respect SES sending limits"

log "Configuration completed successfully!"











