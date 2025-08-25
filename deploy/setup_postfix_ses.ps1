# =============================================================================
# Postfix + Amazon SES Relay Configuration Script (PowerShell)
# Oracle Linux 8/9 Compatible - Run via SSH
# =============================================================================
# This script configures Postfix to relay all outgoing mail through Amazon SES
# Domain: securesystem.email
# SES Region: us-east-1
# =============================================================================

param(
    [Parameter(Mandatory=$true)]
    [string]$SshHost,
    
    [Parameter(Mandatory=$true)]
    [string]$SshUser,
    
    [Parameter(Mandatory=$false)]
    [string]$SshKeyPath,
    
    [Parameter(Mandatory=$false)]
    [string]$SshPort = "22"
)

# Color codes for output
$Red = "`e[31m"
$Green = "`e[32m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$Reset = "`e[0m"

# Logging function
function Write-Log {
    param([string]$Message)
    Write-Host "$Green[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $Message$Reset"
}

function Write-Warning {
    param([string]$Message)
    Write-Host "$Yellow[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] WARNING: $Message$Reset"
}

function Write-Error {
    param([string]$Message)
    Write-Host "$Red[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] ERROR: $Message$Reset"
    exit 1
}

# Function to execute SSH command
function Invoke-SshCommand {
    param([string]$Command)
    
    $sshArgs = @(
        "-p", $SshPort,
        "-o", "StrictHostKeyChecking=no",
        "-o", "UserKnownHostsFile=/dev/null"
    )
    
    if ($SshKeyPath) {
        $sshArgs += "-i", $SshKeyPath
    }
    
    $sshArgs += "${SshUser}@${SshHost}", $Command
    
    $result = & ssh @sshArgs 2>&1
    return $result
}

# Function to copy file via SCP
function Copy-FileViaScp {
    param([string]$LocalPath, [string]$RemotePath)
    
    $scpArgs = @(
        "-P", $SshPort,
        "-o", "StrictHostKeyChecking=no",
        "-o", "UserKnownHostsFile=/dev/null"
    )
    
    if ($SshKeyPath) {
        $scpArgs += "-i", $SshKeyPath
    }
    
    $scpArgs += $LocalPath, "${SshUser}@${SshHost}:${RemotePath}"
    
    & scp @scpArgs
}

Write-Log "Starting Postfix + Amazon SES configuration via SSH..."

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Error ".env file not found. Please copy env.example to .env and configure your SES credentials."
}

# Read and validate environment variables
$envContent = Get-Content ".env" | Where-Object { $_ -match '^SES_SMTP_' }
$envVars = @{}

foreach ($line in $envContent) {
    if ($line -match '^([^=]+)=(.*)$') {
        $envVars[$matches[1]] = $matches[2]
    }
}

# Validate required variables
$requiredVars = @("SES_SMTP_USERNAME", "SES_SMTP_PASSWORD", "SES_SMTP_HOST", "SES_SMTP_PORT")
foreach ($var in $requiredVars) {
    if (-not $envVars.ContainsKey($var) -or [string]::IsNullOrEmpty($envVars[$var])) {
        Write-Error "$var must be set in .env file"
    }
}

Write-Log "Environment variables validated"

# Create the setup script content
$setupScript = @'
#!/bin/bash

# =============================================================================
# Postfix + Amazon SES Relay Configuration Script
# Oracle Linux 8/9 Compatible
# =============================================================================

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   error "This script must be run as root (use sudo)"
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
[email-smtp.us-east-1.amazonaws.com]:587 '$env:SES_SMTP_USERNAME':'$env:SES_SMTP_PASSWORD'
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
relayhost = [email-smtp.us-east-1.amazonaws.com]:587
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
# STEP 4: Test configuration and reload Postfix
# =============================================================================
log "Step 4: Testing configuration and reloading Postfix..."

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
# STEP 5: Create test script
# =============================================================================
log "Step 5: Creating test script..."

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
# STEP 6: Create monitoring script
# =============================================================================
log "Step 6: Creating monitoring script..."

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
# Final verification and summary
# =============================================================================
log "Final verification..."

echo -e "\n${BLUE}=== Postfix + SES Configuration Summary ===${NC}"
echo "✓ Postfix installed and configured"
echo "✓ SASL authentication configured"
echo "✓ TLS encryption enabled"
echo "✓ Relay host: email-smtp.us-east-1.amazonaws.com:587"
echo "✓ Domain: securesystem.email"
echo "✓ Logs: journalctl -u postfix or /var/log/maillog"

echo -e "\n${BLUE}=== Available Commands ===${NC}"
echo "• Test email: /usr/local/bin/test-ses-email.sh <recipient>"
echo "• Monitor: /usr/local/bin/monitor-postfix.sh"
echo "• View logs: journalctl -u postfix -f"
echo "• Check queue: mailq"

echo -e "\n${YELLOW}=== Important Notes ===${NC}"
echo "• If using SES sandbox mode, recipient emails must be verified"
echo "• Check SES console for sending statistics and bounces"
echo "• Monitor logs for authentication or delivery issues"

echo -e "\n${GREEN}✓ Postfix + Amazon SES configuration completed successfully!${NC}"

log "Configuration completed successfully!"
'@

# Create temporary script file
$tempScript = "setup_postfix_ses_temp.sh"
$setupScript | Out-File -FilePath $tempScript -Encoding UTF8

try {
    Write-Log "Copying setup script to remote server..."
    Copy-FileViaScp -LocalPath $tempScript -RemotePath "/tmp/setup_postfix_ses.sh"
    
    Write-Log "Making script executable..."
    Invoke-SshCommand "chmod +x /tmp/setup_postfix_ses.sh"
    
    Write-Log "Running setup script on remote server..."
    $result = Invoke-SshCommand "sudo /tmp/setup_postfix_ses.sh"
    
    Write-Host $result
    
    Write-Log "Cleaning up temporary files..."
    Invoke-SshCommand "rm -f /tmp/setup_postfix_ses.sh"
    
    Write-Log "Configuration completed successfully!"
    
} catch {
    Write-Error "Failed to execute setup script: $($_.Exception.Message)"
} finally {
    # Clean up local temporary file
    if (Test-Path $tempScript) {
        Remove-Item $tempScript
    }
}

Write-Log "Setup completed. You can now test email sending with:"
Write-Host "ssh $SshUser@$SshHost 'sudo /usr/local/bin/test-ses-email.sh your-verified@email.com'"


