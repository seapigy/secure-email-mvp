# 🔒 Security Guide

## Overview
This document outlines security best practices for the Secure Email MVP project.

## 🚨 Critical Security Requirements

### 1. Environment Variables
- **NEVER commit `.env` files** to Git
- Use `env.example` as a template
- Generate unique credentials for each environment
- Keep credentials secure and never share them

### 2. Credential Management
- ✅ `.env` is properly listed in `.gitignore`
- ✅ JWT secrets are auto-generated securely
- ✅ SSH keys are excluded from Git
- ⚠️ **You must add your own Cloudflare R2 credentials**

### 3. Git Security Checklist
- [ ] `.env` file is not tracked by Git
- [ ] No sensitive credentials in commit history
- [ ] SSH keys are properly ignored
- [ ] Database files are excluded
- [ ] Compiled binaries are excluded

## 🔧 Secure Setup Process

### For New Deployments
1. Run the secure setup script:
   ```powershell
   .\scripts\secure_setup.ps1
   ```

2. Add your Cloudflare R2 credentials to `.env`:
   ```
   CLOUDFLARE_R2_ACCESS_KEY=your_access_key_here
   CLOUDFLARE_R2_SECRET_KEY=your_secret_key_here
   CLOUDFLARE_R2_BUCKET=your_bucket_name
   CLOUDFLARE_R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
   ```

3. Verify security:
   ```powershell
   git status
   git ls-files | Select-String "\.env"
   ```

## 🛡️ Security Features

### Frontend Security
- ✅ JWT token authentication
- ✅ Session management
- ✅ Password protection for emails
- ✅ Self-destruct after failed attempts
- ✅ Per-email password unlock

### Backend Security
- ✅ Rate limiting
- ✅ CORS protection
- ✅ Input validation
- ✅ SQL injection prevention
- ✅ Secure password hashing
- ✅ IP reputation checking (AbuseIPDB integration)
✅ Password strength validation & breach checking (HaveIBeenPwned integration)
✅ User account lockout after failed login attempts (configurable thresholds)
✅ Enhanced geo-restriction system (configurable allow/block lists)

## 🚨 Incident Response

### If Credentials Are Exposed
1. **Immediately revoke** exposed credentials
2. **Generate new credentials**
3. **Update all environments**
4. **Check Git history** for any commits
5. **Rotate all related secrets**

### Security Contact
For security issues, please:
1. Do not create public issues
2. Contact the development team privately
3. Provide detailed information about the vulnerability

## 📋 Pre-Deployment Checklist

- [ ] All credentials are secure
- [ ] `.env` file is not tracked by Git
- [ ] No sensitive data in commit history
- [ ] Security features are enabled
- [ ] Rate limiting is configured
- [ ] CORS is properly set up
- [ ] Database is secured
- [ ] SSL/TLS is enabled (production)

## 🛡️ Network-based Threat Mitigation

### IP Reputation Service
The Secure Email MVP integrates with AbuseIPDB to check IP reputation on all signup and login requests.

#### Configuration
- **IP_REPUTATION_API_KEY**: Your AbuseIPDB API key (free tier available)
- **IP_REPUTATION_THRESHOLD**: Malicious threshold (0-100, default: 25)

#### Features
- ✅ Real-time IP reputation checking
- ✅ Automatic blocking of known malicious IPs
- ✅ Graceful fallback on API failures
- ✅ Support for proxy headers (X-Forwarded-For, X-Real-IP, CF-Connecting-IP)
- ✅ IPv4 and IPv6 support
- ✅ Private IP filtering

#### Security Benefits
- Prevents signup/login from known malicious IPs
- Reduces automated attacks and bot registrations
- Maintains service availability during API outages
- No sensitive data sent to reputation service

#### Testing
```powershell
# Test IP reputation integration
.\scripts\test_ip_reputation.ps1
```

### Password Strength & Breach Check Service
The Secure Email MVP integrates comprehensive password validation and breach checking using the HaveIBeenPwned API.

#### Configuration
- **HIBP_API_KEY**: Your HaveIBeenPwned API key (optional, increases rate limits)
- **Password Requirements**:
  - Minimum length: 12 characters
  - Must contain uppercase letters
  - Must contain lowercase letters
  - Must contain numbers
  - Must contain special characters
  - Must not be in common password list
  - Must not be compromised (if API key configured)

#### Features
- ✅ Real-time password strength scoring (0-100)
- ✅ Comprehensive character requirement validation
- ✅ Common password blacklist checking
- ✅ HaveIBeenPwned breach detection using k-anonymity
- ✅ Password improvement suggestions
- ✅ Integration with signup and password reset flows
- ✅ Graceful fallback on API failures
- ✅ Audit logging of validation failures

#### Security Benefits
- Prevents use of weak or compromised passwords
- Reduces risk of credential stuffing attacks
- Provides user-friendly feedback without revealing specific requirements
- Maintains service availability during API outages
- Uses k-anonymity to protect user privacy

#### Testing
```powershell
# Test password validation integration
.\scripts\test_password_validation.ps1
```

### Enhanced Geo-Restriction Service
The Secure Email MVP implements a comprehensive location-based access control system that allows users to configure granular geo-restrictions for their emails.

#### Configuration
- **Rule Types**: Allow rules (whitelist) and block rules (blacklist)
- **Matching Logic**: Country and/or city-based restrictions
- **Strict Mode**: Option to require both country and city matches
- **Default Actions**: Configurable behavior when no rules exist
- **Violation Tracking**: Automatic tracking of access violations

#### Features
- ✅ Configurable allow/block lists for countries and cities
- ✅ Flexible rule types with different matching logic
- ✅ Strict mode requiring both country and city matches
- ✅ Default action configuration (allow/block when no rules exist)
- ✅ Violation tracking with timestamps
- ✅ Real-time enforcement during email access
- ✅ JSON-based rule and configuration storage
- ✅ Comprehensive rule validation and normalization
- ✅ Integration with existing geolocation service
- ✅ Audit logging of all geo-restriction events

#### Security Benefits
- Prevents unauthorized access from specific geographic locations
- Reduces risk of location-based attacks and data exfiltration
- Provides granular control over email access based on user location
- Maintains comprehensive audit trail of access attempts
- Integrates with existing security features for enhanced protection

#### API Endpoints
- `GET /api/email/{id}/geo-restrictions` - Retrieve rules
- `POST /api/email/{id}/geo-restrictions` - Create rule
- `PUT /api/email/{id}/geo-restrictions/{ruleId}` - Update rule
- `DELETE /api/email/{id}/geo-restrictions/{ruleId}` - Delete rule
- `GET /api/email/{id}/geo-restrictions/config` - Get configuration
- `PUT /api/email/{id}/geo-restrictions/config` - Update configuration
- `GET /api/email/{id}/geo-restrictions/status` - Get status

#### Testing
```powershell
# Test geo-restriction integration
.\scripts\test_geo_restrictions.ps1
```

### User Account Lockout Service
The Secure Email MVP implements temporary account lockout after failed login attempts to mitigate brute-force attacks.

#### Configuration
- **LOGIN_RATE_LIMIT_ENABLED**: Enable/disable lockout (1 = enabled, 0 = disabled)
- **LOGIN_MAX_ATTEMPTS**: Maximum failed attempts before lockout (default: 5)
- **LOGIN_LOCKOUT_MINUTES**: Lockout duration in minutes (default: 30)
- **LOGIN_ATTEMPT_WINDOW_MINUTES**: Time window for counting attempts (default: 15)

#### Features
- ✅ Configurable failed attempt thresholds
- ✅ Automatic account lockout after threshold exceeded
- ✅ Time-based attempt window with automatic reset
- ✅ Automatic lockout expiration after configured duration
- ✅ Manual account unlock via API
- ✅ Real-time lockout status checking
- ✅ System-wide lockout statistics
- ✅ Comprehensive audit logging
- ✅ Graceful degradation on service failures

#### Security Benefits
- Prevents brute-force password guessing attacks
- Configurable thresholds balance security and usability
- Time-based windows prevent permanent lockouts
- Generic error messages don't reveal lockout status
- Comprehensive audit logging for compliance
- Maintains service availability during lockout service failures

#### API Endpoints
- `GET /api/auth/lockout/status?email={email}` - Check lockout status
- `POST /api/auth/lockout/unlock` - Manually unlock account
- `GET /api/auth/lockout/config` - Get current configuration
- `GET /api/auth/lockout/stats` - System-wide statistics

#### Testing
```powershell
# Test user account lockout functionality
.\scripts\test_user_account_lockout.ps1
```

## 🔍 Security Monitoring

### Regular Checks
- Monitor for exposed credentials
- Review access logs
- Check for unusual activity
- Update dependencies regularly
- Audit security configurations
- Monitor IP reputation API usage and failures

### Automated Security
- Git hooks prevent credential commits
- Automated dependency scanning
- Security linting rules
- Environment validation

---

**Remember**: Security is everyone's responsibility. When in doubt, err on the side of caution.
