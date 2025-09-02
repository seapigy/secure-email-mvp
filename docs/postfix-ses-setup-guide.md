# Postfix + Amazon SES Relay Configuration Guide

## Overview

This guide provides comprehensive instructions for configuring Postfix on Oracle Linux to relay all outgoing email through Amazon SES (Simple Email Service). This setup ensures reliable email delivery with proper authentication, encryption, and monitoring.

## Prerequisites

### AWS SES Setup
1. **SES Domain Verification**: Ensure `securesystem.email` is verified in SES
2. **DKIM Configuration**: Set up DKIM signing for the domain
3. **SPF Record**: Add SPF record to DNS: `v=spf1 include:amazonses.com ~all`
4. **SMTP Credentials**: Create SMTP credentials in SES Console
5. **Sending Limits**: Request production access if needed (default: sandbox mode)

### Server Requirements
- Oracle Linux 8 or 9
- Root access or sudo privileges
- Outbound internet access (port 587)
- DNS resolution working

## Configuration Steps

### Step 1: Environment Setup

1. **Copy environment template**:
   ```bash
   cp env.example .env
   ```

2. **Configure SES credentials** in `.env`:
   ```bash
   # Amazon SES SMTP Configuration
   SES_SMTP_USERNAME=your_ses_smtp_username_here
   SES_SMTP_PASSWORD=your_ses_smtp_password_here
   SES_SMTP_HOST=email-smtp.us-east-1.amazonaws.com
   SES_SMTP_PORT=587
   SES_SMTP_REGION=us-east-1
   SES_SMTP_RATE_LIMIT=14
   SES_SMTP_SANDBOX_MODE=true
   SES_DEFAULT_SENDER=noreply@securesystem.email
   ```

### Step 2: Run Setup Script

#### Option A: Direct SSH (Recommended)
```bash
# Make script executable
chmod +x deploy/setup_postfix_ses.sh

# Run on Oracle Linux server
sudo ./deploy/setup_postfix_ses.sh
```

#### Option B: PowerShell (Windows)
```powershell
# Run from Windows PowerShell
.\deploy\setup_postfix_ses.ps1 -SshHost "your-server-ip" -SshUser "your-username" -SshKeyPath "path\to\key.pem"
```

### Step 3: Verify Configuration

1. **Check Postfix status**:
   ```bash
   systemctl status postfix
   ```

2. **Test configuration**:
   ```bash
   postfix check
   ```

3. **View configuration**:
   ```bash
   postconf -n
   ```

## Testing Email Delivery

### Test Command
```bash
# Test to verified email address
sudo /usr/local/bin/test-ses-email.sh your-verified@email.com

# Test with custom subject and body
sudo /usr/local/bin/test-ses-email.sh recipient@example.com "Test Subject" "Test message body"
```

### Manual Test
```bash
# Create test email
cat > /tmp/test_email.txt << EOF
From: noreply@securesystem.email
To: your-verified@email.com
Subject: Test Email
Date: $(date -R)

This is a test email sent through Postfix + Amazon SES relay.
EOF

# Send email
sendmail -t < /tmp/test_email.txt
```

## Monitoring and Maintenance

### Available Commands

1. **Monitor Postfix**:
   ```bash
   sudo /usr/local/bin/monitor-postfix.sh
   ```

2. **View logs**:
   ```bash
   # Real-time logs
   journalctl -u postfix -f
   
   # Recent logs
   journalctl -u postfix --no-pager -n 50
   
   # Traditional log file
   tail -f /var/log/maillog
   ```

3. **Check queue**:
   ```bash
   mailq
   ```

4. **Flush queue** (if needed):
   ```bash
   postsuper -r ALL
   ```

### Credential Rotation

1. **Create new SES SMTP credentials** in AWS Console
2. **Update credentials**:
   ```bash
   sudo /usr/local/bin/rotate-ses-credentials.sh "new_username" "new_password"
   ```

3. **Test after rotation**:
   ```bash
   sudo /usr/local/bin/test-ses-email.sh your-verified@email.com
   ```

## Configuration Files

### Key Files Created/Modified

1. **`/etc/postfix/main.cf`** - Main Postfix configuration
2. **`/etc/postfix/sasl_passwd`** - SES SMTP credentials (600 permissions)
3. **`/etc/postfix/sasl/smtpd.conf`** - SASL authentication config
4. **`/usr/local/bin/test-ses-email.sh`** - Test script
5. **`/usr/local/bin/monitor-postfix.sh`** - Monitoring script
6. **`/usr/local/bin/rotate-ses-credentials.sh`** - Credential rotation script

### Backup Files
- Original configurations are backed up with timestamps
- Location: `/etc/postfix/main.cf.backup.YYYYMMDD_HHMMSS`

## Troubleshooting

### Common Issues

#### 1. Authentication Failures
**Symptoms**: `SASL authentication failed` in logs
**Solutions**:
- Verify SES SMTP credentials in AWS Console
- Check `/etc/postfix/sasl_passwd` file permissions (should be 600)
- Ensure credentials are correct in `.env` file

#### 2. Connection Timeouts
**Symptoms**: `Connection timed out` or `Network is unreachable`
**Solutions**:
- Check outbound connectivity: `telnet email-smtp.us-east-1.amazonaws.com 587`
- Verify firewall allows outbound port 587
- Check DNS resolution: `nslookup email-smtp.us-east-1.amazonaws.com`

#### 3. TLS/SSL Issues
**Symptoms**: `SSL_connect failed` or `certificate verify failed`
**Solutions**:
- Update CA certificates: `dnf update ca-certificates`
- Check TLS configuration in `/etc/postfix/main.cf`
- Verify system time is correct: `date`

#### 4. Rate Limiting
**Symptoms**: `Sending rate exceeded` or emails queued
**Solutions**:
- Check SES sending limits in AWS Console
- Monitor sending rate: `postconf smtp_destination_rate_delay`
- Implement proper rate limiting in applications

#### 5. Sandbox Mode Restrictions
**Symptoms**: `Email address not verified` errors
**Solutions**:
- Verify recipient email addresses in SES Console
- Request production access from AWS
- Use only verified email addresses for testing

### Diagnostic Commands

```bash
# Check Postfix configuration
postfix check

# View current configuration
postconf -n

# Test SMTP connection
telnet email-smtp.us-east-1.amazonaws.com 587

# Check DNS resolution
nslookup email-smtp.us-east-1.amazonaws.com

# View system logs
journalctl -u postfix --no-pager -n 100

# Check file permissions
ls -la /etc/postfix/sasl_passwd

# Test SASL configuration
postconf -h smtp_sasl_auth_enable
postconf -h smtp_sasl_password_maps
```

### Log Analysis

#### Common Log Entries

1. **Successful delivery**:
   ```
   postfix/smtp[12345]: ABC123456: to=<recipient@example.com>, relay=email-smtp.us-east-1.amazonaws.com[52.94.12.34]:587, delay=0.5, delays=0.01/0.01/0.3/0.2, dsn=2.0.0, status=sent (250 Ok 0123456789abcdef)
   ```

2. **Authentication failure**:
   ```
   postfix/smtp[12345]: ABC123456: SASL authentication failed; server email-smtp.us-east-1.amazonaws.com[52.94.12.34] said: 535 Authentication Credentials Invalid
   ```

3. **Connection timeout**:
   ```
   postfix/smtp[12345]: ABC123456: connect to email-smtp.us-east-1.amazonaws.com[52.94.12.34]:587: Connection timed out
   ```

## Security Considerations

### File Permissions
- `/etc/postfix/sasl_passwd`: 600 (root:root)
- `/etc/postfix/sasl_passwd.db`: 600 (root:root)
- All Postfix config files: 644 (root:root)

### Network Security
- Postfix only listens on loopback (127.0.0.1)
- No inbound SMTP ports exposed
- All outbound traffic encrypted with TLS

### Credential Management
- Credentials stored in environment variables
- SASL password file has restricted permissions
- Regular credential rotation recommended

## Performance Tuning

### Rate Limiting
```bash
# Current settings (optimized for SES)
smtp_destination_concurrency_limit = 2
smtp_destination_rate_delay = 1s
smtp_extra_recipient_limit = 10
```

### Queue Management
```bash
# Queue settings
maximal_queue_lifetime = 5d
minimal_backoff_time = 300s
maximal_backoff_time = 4000s
```

### Monitoring
- Queue monitoring: `mailq`
- Performance metrics: `postconf -h | grep -E "(queue|process|limit)"`
- Log rotation: Configured in `/etc/logrotate.d/postfix`

## Integration with Applications

### Environment Variables
Applications can use these environment variables for email configuration:
```bash
export MAIL_HOST=localhost
export MAIL_PORT=25
export MAIL_FROM=noreply@securesystem.email
```

### SMTP Configuration
Applications should configure SMTP to use:
- Host: `localhost`
- Port: `25` (or `587` for submission)
- Authentication: None (local delivery)
- Encryption: None (local connection)

### Testing Application Integration
```bash
# Test from application
echo "Subject: Test" | sendmail -f noreply@securesystem.email recipient@example.com
```

## Maintenance Schedule

### Daily
- Monitor queue: `mailq`
- Check logs: `journalctl -u postfix --no-pager -n 50`

### Weekly
- Review SES sending statistics in AWS Console
- Check for bounces and complaints
- Monitor sending reputation

### Monthly
- Rotate SES SMTP credentials
- Review and update rate limits
- Check for configuration updates

### Quarterly
- Review security settings
- Update Postfix and dependencies
- Review SES production access requirements

## Support and Resources

### AWS SES Resources
- [SES Developer Guide](https://docs.aws.amazon.com/ses/latest/dg/)
- [SES SMTP Configuration](https://docs.aws.amazon.com/ses/latest/dg/send-email-smtp.html)
- [SES Sending Limits](https://docs.aws.amazon.com/ses/latest/dg/manage-sending-quotas.html)

### Postfix Resources
- [Postfix Documentation](http://www.postfix.org/documentation.html)
- [Postfix Configuration](http://www.postfix.org/postconf.5.html)
- [Postfix SASL Configuration](http://www.postfix.org/SASL_README.html)

### Troubleshooting Resources
- [SES Troubleshooting](https://docs.aws.amazon.com/ses/latest/dg/troubleshooting.html)
- [Postfix Troubleshooting](http://www.postfix.org/TROUBLESHOOTING.html)

## Emergency Procedures

### Service Recovery
```bash
# Restart Postfix
systemctl restart postfix

# Check status
systemctl status postfix

# View recent logs
journalctl -u postfix --no-pager -n 100
```

### Configuration Recovery
```bash
# Restore from backup
cp /etc/postfix/main.cf.backup.YYYYMMDD_HHMMSS /etc/postfix/main.cf

# Reload configuration
systemctl reload postfix
```

### Credential Emergency
```bash
# Update credentials immediately
sudo /usr/local/bin/rotate-ses-credentials.sh "emergency_user" "emergency_pass"

# Test immediately
sudo /usr/local/bin/test-ses-email.sh verified@email.com
```

---

**Note**: This configuration is designed for production use with proper security measures. Always test in a staging environment first and monitor logs for any issues after deployment.












