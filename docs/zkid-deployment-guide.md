# ZKID Layer Deployment Guide

## Overview

This guide provides step-by-step instructions for deploying the Zero-Knowledge Identity Layer (ZKID) in production environments. The ZKID layer provides maximum privacy for user email mappings while maintaining full operational functionality.

## Prerequisites

### System Requirements
- Go 1.23 or later
- SQLite database (production-ready version)
- Secure key management system
- RBAC-enabled authentication system
- Audit logging infrastructure

### Security Requirements
- Secure environment variable management
- Hardware Security Module (HSM) or equivalent for key storage
- Network security and firewall configuration
- Access control and monitoring systems

## Environment Configuration

### Required Environment Variables

```bash
# ZKID Feature Flag (Required)
ZKID_ENABLED=true

# Cryptographic Keys (Required - 32 bytes each, hex encoded)
ZKID_MASTER_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
ZKID_EMAIL_HASH_PEPPER=deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef
ZKID_RECOVERY_PEPPER=feedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeef

# Optional: PQC Integration (Future)
ZKID_USE_PQC_KEY_WRAPPING=false
```

### Key Generation

Generate secure cryptographic keys using the following commands:

```bash
# Generate 32-byte master key
openssl rand -hex 32

# Generate 32-byte email hash pepper
openssl rand -hex 32

# Generate 32-byte recovery pepper
openssl rand -hex 32
```

### Environment Validation

Create a validation script to ensure all required variables are set:

```bash
#!/bin/bash
# validate_zkid_env.sh

required_vars=(
    "ZKID_ENABLED"
    "ZKID_MASTER_KEY"
    "ZKID_EMAIL_HASH_PEPPER"
    "ZKID_RECOVERY_PEPPER"
)

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "ERROR: Required environment variable $var is not set"
        exit 1
    fi
done

# Validate key lengths (32 bytes = 64 hex characters)
if [ ${#ZKID_MASTER_KEY} -ne 64 ]; then
    echo "ERROR: ZKID_MASTER_KEY must be 64 hex characters (32 bytes)"
    exit 1
fi

if [ ${#ZKID_EMAIL_HASH_PEPPER} -ne 64 ]; then
    echo "ERROR: ZKID_EMAIL_HASH_PEPPER must be 64 hex characters (32 bytes)"
    exit 1
fi

if [ ${#ZKID_RECOVERY_PEPPER} -ne 64 ]; then
    echo "ERROR: ZKID_RECOVERY_PEPPER must be 64 hex characters (32 bytes)"
    exit 1
fi

echo "ZKID environment validation passed"
```

## Database Migration

### Automatic Migration

The ZKID layer automatically applies database migrations on startup when `ZKID_ENABLED=true`. The migration creates:

- `zkid_email_mappings` table
- `zkid_recovery_codes` table
- Required indexes and foreign key constraints

### Manual Migration (Optional)

If you prefer manual migration control, run the migration script:

```bash
# Apply ZKID migration manually
sqlite3 your_database.db < migrations/xxxx_add_zkid_layer.sql
```

### Migration Verification

Verify the migration was successful:

```sql
-- Check tables exist
SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'zkid_%';

-- Check indexes
SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_zkid_%';

-- Verify table structure
PRAGMA table_info(zkid_email_mappings);
PRAGMA table_info(zkid_recovery_codes);
```

## Deployment Steps

### 1. Pre-Deployment Checklist

- [ ] Environment variables configured and validated
- [ ] Database backup completed
- [ ] Cryptographic keys securely stored
- [ ] RBAC roles configured (system_admin, enterprise_admin)
- [ ] Audit logging configured
- [ ] Monitoring and alerting set up
- [ ] Rollback plan prepared

### 2. Staging Deployment

```bash
# 1. Deploy to staging environment
export ZKID_ENABLED=true
export ZKID_MASTER_KEY=<staging_master_key>
export ZKID_EMAIL_HASH_PEPPER=<staging_email_pepper>
export ZKID_RECOVERY_PEPPER=<staging_recovery_pepper>

# 2. Start application
go run cmd/api/main.go

# 3. Verify ZKID initialization
grep "ZKID" application.log
```

### 3. Production Deployment

```bash
# 1. Deploy with ZKID disabled initially
export ZKID_ENABLED=false

# 2. Start application and verify normal operation
go run cmd/api/main.go

# 3. Enable ZKID layer
export ZKID_ENABLED=true

# 4. Restart application
# (Application will automatically apply migrations and enable ZKID)

# 5. Verify ZKID functionality
curl -X GET "http://localhost:8080/api/admin/zkid/stats" \
  -H "Authorization: Bearer <admin_token>"
```

### 4. Post-Deployment Verification

#### API Endpoint Testing

```bash
# Test internal ZKID operations
curl -X POST "http://localhost:8080/api/zkid/mapping" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-uuid","email":"test@example.com"}'

# Test admin operations (requires admin token)
curl -X GET "http://localhost:8080/api/admin/zkid/recovery-codes?user_id=test-uuid&count=5" \
  -H "Authorization: Bearer <admin_token>"

curl -X GET "http://localhost:8080/api/admin/zkid/stats" \
  -H "Authorization: Bearer <admin_token>"
```

#### Security Verification

```bash
# Verify UUID-only logging (no email exposure)
grep "ZKID_ADMIN" application.log | grep -v "@"

# Verify RBAC enforcement
curl -X GET "http://localhost:8080/api/admin/zkid/stats" \
  -H "Authorization: Bearer <user_token>"
# Should return 403 Forbidden
```

## Monitoring and Alerting

### Key Metrics to Monitor

1. **ZKID Layer Status**
   - ZKID enabled/disabled state
   - Database migration success/failure
   - Configuration loading errors

2. **Operational Metrics**
   - Email mapping creation rate
   - Recovery code generation rate
   - Admin operation frequency
   - Error rates by operation type

3. **Security Metrics**
   - Failed admin access attempts
   - RBAC violations
   - Cryptographic operation failures
   - Audit log completeness

### Alerting Rules

```yaml
# Example Prometheus alerting rules
groups:
  - name: zkid_alerts
    rules:
      - alert: ZKIDLayerDisabled
        expr: zkid_enabled == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "ZKID layer is disabled"
          
      - alert: ZKIDMigrationFailed
        expr: zkid_migration_errors > 0
        labels:
          severity: critical
        annotations:
          summary: "ZKID database migration failed"
          
      - alert: ZKIDAdminAccessDenied
        expr: rate(zkid_admin_access_denied_total[5m]) > 0.1
        labels:
          severity: warning
        annotations:
          summary: "High rate of ZKID admin access denials"
```

### Log Monitoring

Monitor these log patterns:

```bash
# ZKID initialization
grep "Initializing ZKID layer" application.log

# Admin operations (UUID-only)
grep "ZKID_ADMIN" application.log

# Security events
grep "ADMIN_REQUIRED\|AUTH_REQUIRED" application.log

# Migration events
grep "ZKID schema applied" application.log
```

## Rollback Procedures

### Emergency Rollback

```bash
# 1. Disable ZKID layer
export ZKID_ENABLED=false

# 2. Restart application
# (Application will continue with existing authentication)

# 3. Verify normal operation
curl -X POST "http://localhost:8080/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'
```

### Data Preservation

ZKID data is preserved during rollback:
- Email mappings remain encrypted in database
- Recovery codes remain stored
- No data loss occurs
- Can be re-enabled later

### Gradual Rollback

```bash
# 1. Disable new ZKID mappings
export ZKID_ENABLED=false

# 2. Monitor existing functionality
# 3. Gradually migrate users back to legacy system
# 4. Clean up ZKID data if needed
```

## Security Considerations

### Key Management

- Store cryptographic keys in HSM or secure key management system
- Rotate keys regularly (quarterly recommended)
- Use different keys for different environments
- Implement key backup and recovery procedures

### Access Control

- Ensure RBAC is properly configured
- Monitor admin access patterns
- Implement least-privilege access
- Regular access reviews

### Audit and Compliance

- All admin operations are logged with UUID-only identifiers
- No external email addresses are ever logged
- Maintain audit trail for compliance requirements
- Regular security assessments

### Network Security

- Use TLS 1.3 for all API communications
- Implement rate limiting on admin endpoints
- Network segmentation for admin access
- Intrusion detection and prevention

## Troubleshooting

### Common Issues

#### ZKID Layer Not Initializing

```bash
# Check environment variables
echo $ZKID_ENABLED
echo $ZKID_MASTER_KEY

# Check logs
grep "ZKID" application.log
```

#### Database Migration Failures

```bash
# Check database permissions
ls -la your_database.db

# Check migration file
cat migrations/xxxx_add_zkid_layer.sql

# Manual migration
sqlite3 your_database.db < migrations/xxxx_add_zkid_layer.sql
```

#### Admin Access Denied

```bash
# Verify RBAC configuration
# Check user role assignment
# Verify JWT token validity
# Check admin endpoint registration
```

#### Cryptographic Errors

```bash
# Verify key format (64 hex characters)
echo $ZKID_MASTER_KEY | wc -c

# Check key encoding
echo $ZKID_MASTER_KEY | xxd -r -p | wc -c
```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
export ZKID_DEBUG=true
export LOG_LEVEL=debug
```

## Performance Considerations

### Database Performance

- ZKID tables are indexed for optimal performance
- Email hash lookups are O(1) with proper indexing
- Recovery code validation uses efficient queries
- Consider database connection pooling

### Cryptographic Performance

- AES-256-GCM encryption is hardware-accelerated
- Argon2id hashing is memory-hard but optimized
- Key wrapping operations are minimal overhead
- Consider caching for frequently accessed data

### Monitoring Performance

- Monitor cryptographic operation latency
- Track database query performance
- Monitor memory usage for Argon2id operations
- Set up performance alerts

## Compliance and Auditing

### GDPR Compliance

- ZKID provides data minimization
- External emails are never stored in plaintext
- Right to be forgotten is supported
- Data portability maintained

### SOC 2 Compliance

- Access controls documented and tested
- Audit trails maintained
- Security controls implemented
- Regular assessments conducted

### Industry Standards

- Follows NIST cryptographic standards
- Implements OWASP security guidelines
- Uses industry-standard algorithms
- Regular security assessments

## Support and Maintenance

### Regular Maintenance

- Monthly security updates
- Quarterly key rotation
- Annual security assessments
- Continuous monitoring and alerting

### Support Contacts

- Security issues: security@company.com
- Technical support: support@company.com
- Emergency contact: oncall@company.com

### Documentation Updates

- Keep deployment guide current
- Update security procedures
- Maintain runbooks
- Regular team training

## Conclusion

The ZKID layer provides enterprise-grade privacy and security while maintaining full operational functionality. Follow this deployment guide carefully to ensure successful implementation and ongoing security.

For additional support or questions, refer to the main ZKID documentation or contact the development team.
