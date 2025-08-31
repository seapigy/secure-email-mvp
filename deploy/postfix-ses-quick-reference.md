# Postfix + Amazon SES Quick Reference

## 🚀 Quick Setup

```bash
# 1. Configure .env with SES credentials
cp env.example .env
# Edit .env with your SES SMTP credentials

# 2. Run setup script
chmod +x deploy/setup_postfix_ses.sh
sudo ./deploy/setup_postfix_ses.sh
```

## 📧 Test Email Sending

```bash
# Test to verified email
sudo /usr/local/bin/test-ses-email.sh your-verified@email.com

# Test with custom subject/body
sudo /usr/local/bin/test-ses-email.sh recipient@example.com "Subject" "Message"
```

## 🔍 Monitoring Commands

```bash
# Check Postfix status
systemctl status postfix

# Monitor queue
mailq

# View real-time logs
journalctl -u postfix -f

# Comprehensive monitoring
sudo /usr/local/bin/monitor-postfix.sh
```

## 🔧 Configuration

```bash
# View current config
postconf -n

# Check configuration
postfix check

# Reload Postfix
systemctl reload postfix
```

## 🔄 Credential Rotation

```bash
# Update SES credentials
sudo /usr/local/bin/rotate-ses-credentials.sh "new_user" "new_pass"

# Test after rotation
sudo /usr/local/bin/test-ses-email.sh verified@email.com
```

## 🚨 Troubleshooting

### Common Issues & Solutions

| Issue | Symptom | Solution |
|-------|---------|----------|
| **Authentication Failed** | `SASL authentication failed` | Check SES credentials in AWS Console |
| **Connection Timeout** | `Connection timed out` | Test: `telnet email-smtp.us-east-1.amazonaws.com 587` |
| **TLS Error** | `SSL_connect failed` | Update CA certs: `dnf update ca-certificates` |
| **Rate Limited** | `Sending rate exceeded` | Check SES limits in AWS Console |
| **Sandbox Mode** | `Email address not verified` | Verify recipient in SES Console |

### Diagnostic Commands

```bash
# Test SMTP connectivity
telnet email-smtp.us-east-1.amazonaws.com 587

# Check DNS resolution
nslookup email-smtp.us-east-1.amazonaws.com

# View recent logs
journalctl -u postfix --no-pager -n 50

# Check file permissions
ls -la /etc/postfix/sasl_passwd

# Flush queue (if needed)
postsuper -r ALL
```

## 📁 Key Files

| File | Purpose | Permissions |
|------|---------|-------------|
| `/etc/postfix/main.cf` | Main configuration | 644 |
| `/etc/postfix/sasl_passwd` | SES credentials | 600 |
| `/usr/local/bin/test-ses-email.sh` | Test script | 755 |
| `/usr/local/bin/monitor-postfix.sh` | Monitoring script | 755 |

## 🔐 Security Checklist

- [ ] SES SMTP credentials configured in `.env`
- [ ] Domain `securesystem.email` verified in SES
- [ ] DKIM configured for domain
- [ ] SPF record added to DNS
- [ ] `/etc/postfix/sasl_passwd` has 600 permissions
- [ ] Postfix only listening on loopback (127.0.0.1)

## 📊 SES Configuration

```bash
# Environment variables needed in .env
SES_SMTP_USERNAME=your_ses_smtp_username
SES_SMTP_PASSWORD=your_ses_smtp_password
SES_SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SES_SMTP_PORT=587
SES_SMTP_REGION=us-east-1
SES_DEFAULT_SENDER=noreply@securesystem.email
```

## 🚀 Emergency Procedures

```bash
# Service recovery
systemctl restart postfix

# Configuration recovery
cp /etc/postfix/main.cf.backup.YYYYMMDD_HHMMSS /etc/postfix/main.cf
systemctl reload postfix

# Emergency credential update
sudo /usr/local/bin/rotate-ses-credentials.sh "emergency_user" "emergency_pass"
```

## 📞 Support Resources

- **AWS SES Console**: https://console.aws.amazon.com/ses/
- **SES Documentation**: https://docs.aws.amazon.com/ses/
- **Postfix Documentation**: http://www.postfix.org/documentation.html
- **Logs**: `journalctl -u postfix -f` or `/var/log/maillog`

---

**Remember**: Always test with verified email addresses in SES sandbox mode!











