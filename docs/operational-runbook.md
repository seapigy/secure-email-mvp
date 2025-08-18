# Secure Email MVP - Operational Runbook
## Comprehensive Guide for Admins and Security Teams

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Monitoring Dashboards](#monitoring-dashboards)
3. [Alert Triage Procedures](#alert-triage-procedures)
4. [Penetration Testing Execution](#penetration-testing-execution)
5. [Admin Onboarding & Revocation](#admin-onboarding--revocation)
6. [Disaster Recovery Procedures](#disaster-recovery-procedures)
7. [Security Incident Response](#security-incident-response)
8. [Maintenance Procedures](#maintenance-procedures)
9. [Troubleshooting Guide](#troubleshooting-guide)
10. [Emergency Contacts](#emergency-contacts)

---

## System Overview

### Architecture Components

The Secure Email MVP consists of the following critical components:

- **ZKID Layer**: Zero-Knowledge Identity system for encrypted email mapping
- **PQC Encryption**: Post-Quantum Cryptography for future-proof encryption
- **Admin Dashboard**: Multi-admin management with RBAC
- **Email Pipeline**: Secure email delivery and storage
- **Monitoring System**: Real-time metrics and alerting
- **Audit Logging**: Comprehensive security event tracking

### Access Levels

| Role | Permissions | Dashboard Access |
|------|-------------|------------------|
| Root Admin | Full system access, admin management | All panels |
| Full Admin | Operational management, limited admin control | All panels except admin management |
| Read-Only Admin | View-only access to monitoring data | Monitoring panels only |

---

## Monitoring Dashboards

### Dashboard Access

**URL**: `https://your-domain.com/admin`  
**Authentication**: Multi-factor authentication required  
**Session Timeout**: 30 minutes of inactivity

### Key Dashboard Panels

#### 1. ZKID Layer Panel
**Purpose**: Monitor Zero-Knowledge Identity operations

**Critical Metrics to Monitor**:
- UUID mapping creation rate
- Recovery code usage (active/expired/revoked)
- Mapping retrieval success rate
- Security status indicators

**Normal Values**:
- Mapping creation: 10-100 per hour
- Recovery code usage: <5% of total codes
- Retrieval success rate: >99%

**Alert Thresholds**:
- Mapping creation >500/hour (potential abuse)
- Recovery code usage >20% (security concern)
- Retrieval success rate <95% (system issue)

#### 2. PQC Encryption Panel
**Purpose**: Monitor Post-Quantum Cryptography operations

**Critical Metrics to Monitor**:
- Key rotation schedule status
- AEAD encryption success/failure rates
- Tag verification errors
- Key health status

**Normal Values**:
- Key rotation: Every 24 hours
- Encryption success rate: >99.9%
- Tag verification errors: <0.1%

**Alert Thresholds**:
- Encryption failure rate >1%
- Tag verification errors >1%
- Key rotation overdue >2 hours

#### 3. Email Delivery Panel
**Purpose**: Monitor email system performance

**Critical Metrics to Monitor**:
- Email queue length
- Failed delivery attempts
- Storage usage
- Read-once enforcement counts

**Normal Values**:
- Queue length: <100 emails
- Failed deliveries: <5%
- Storage usage: <80%

**Alert Thresholds**:
- Queue length >1000 emails
- Failed deliveries >10%
- Storage usage >90%

#### 4. Performance & Operational Panel
**Purpose**: Monitor system performance and health

**Critical Metrics to Monitor**:
- API latency (P50, P95, P99)
- Concurrent user sessions
- Database response times
- System resource usage

**Normal Values**:
- P50 latency: <100ms
- P95 latency: <500ms
- P99 latency: <1000ms
- Database response: <50ms

**Alert Thresholds**:
- P95 latency >1000ms
- P99 latency >2000ms
- Database response >200ms

#### 5. Alerts Panel
**Purpose**: Centralized alert management

**Alert Severity Levels**:
- **Critical**: Immediate action required
- **High**: Action required within 1 hour
- **Medium**: Action required within 4 hours
- **Low**: Action required within 24 hours

### Dashboard Navigation

1. **Login**: Use admin credentials with MFA
2. **Role Selection**: Dashboard adapts based on admin role
3. **Panel Access**: Click on panel headers to expand/collapse
4. **Real-time Updates**: Data refreshes automatically every 30 seconds
5. **Export**: Use export buttons for detailed reports

---

## Alert Triage Procedures

### Alert Classification Matrix

| Alert Type | Severity | Response Time | Escalation |
|------------|----------|---------------|------------|
| ZKID/PQC Failures | Critical | 15 minutes | Immediate |
| System Outage | Critical | 15 minutes | Immediate |
| Security Breach | Critical | 5 minutes | Immediate |
| Performance Degradation | High | 1 hour | 30 minutes |
| High Resource Usage | Medium | 4 hours | 2 hours |
| Informational | Low | 24 hours | 12 hours |

### Step-by-Step Alert Triage Process

#### Step 1: Alert Assessment
1. **Identify Alert Type**: Check alert category and severity
2. **Verify Alert**: Confirm it's not a false positive
3. **Assess Impact**: Determine affected systems and users
4. **Check Recent Changes**: Review recent deployments or changes

#### Step 2: Initial Response
1. **Acknowledge Alert**: Mark alert as acknowledged in dashboard
2. **Gather Information**: Collect relevant logs and metrics
3. **Notify Stakeholders**: Alert appropriate team members
4. **Begin Investigation**: Start root cause analysis

#### Step 3: Investigation
1. **Review Logs**: Check system logs for errors
2. **Check Metrics**: Analyze performance and health metrics
3. **Test Functionality**: Verify affected system components
4. **Identify Root Cause**: Determine underlying issue

#### Step 4: Resolution
1. **Implement Fix**: Apply necessary changes or workarounds
2. **Verify Resolution**: Confirm issue is resolved
3. **Update Status**: Mark alert as resolved
4. **Document Actions**: Record all steps taken

#### Step 5: Post-Incident
1. **Review Process**: Conduct post-incident review
2. **Update Procedures**: Modify procedures if needed
3. **Monitor Recovery**: Ensure system stability
4. **Document Lessons**: Record lessons learned

### Common Alert Scenarios

#### Scenario 1: ZKID Mapping Creation Failure
**Symptoms**: High failure rate in UUID mapping creation
**Immediate Actions**:
1. Check database connectivity
2. Verify ZKID service status
3. Review recent code deployments
4. Check system resources

**Resolution Steps**:
1. Restart ZKID service if needed
2. Check database locks and deadlocks
3. Verify encryption key availability
4. Monitor for recurrence

#### Scenario 2: PQC Key Rotation Failure
**Symptoms**: Key rotation overdue or failed
**Immediate Actions**:
1. Check PQC service status
2. Verify HSM connectivity
3. Review key rotation logs
4. Check system time synchronization

**Resolution Steps**:
1. Manually trigger key rotation
2. Verify new keys are properly distributed
3. Check for key conflicts
4. Update rotation schedule if needed

#### Scenario 3: High API Latency
**Symptoms**: P95 latency >1000ms
**Immediate Actions**:
1. Check server resource usage
2. Review database performance
3. Analyze recent traffic patterns
4. Check for external dependencies

**Resolution Steps**:
1. Scale resources if needed
2. Optimize database queries
3. Implement caching if appropriate
4. Monitor performance improvements

---

## Penetration Testing Execution

### Pre-Test Preparation

#### 1. Environment Setup
```bash
# Ensure staging environment is ready
./scripts/pentest/validate_environment.sh

# Verify test tools are available
./scripts/pentest/check_prerequisites.sh

# Backup current system state
./scripts/operational/disaster_recovery.go --backup
```

#### 2. Test Configuration
```json
{
  "environment": "staging",
  "safe_mode": true,
  "max_concurrent_tests": 5,
  "test_timeout": 300,
  "report_format": "html,json"
}
```

### Standard Penetration Test Execution

#### Phase 1: Automated Security Tests
```bash
# Run standard security test suite
./scripts/pentest/run_security_tests.ps1 -Environment staging

# Expected duration: 15-30 minutes
# Tests: 45 automated security tests across 7 categories
```

**Test Categories**:
1. ZKID Layer Security (8 tests)
2. PQC Encryption Security (6 tests)
3. Authentication & Authorization (7 tests)
4. Admin Dashboard Security (6 tests)
5. Email Pipeline Security (6 tests)
6. Real-time Monitoring Security (6 tests)
7. Compliance & Audit Security (6 tests)

#### Phase 2: Level-5 Red Team Light Tests
```bash
# Run advanced red team tests
./scripts/pentest/run_redteam_pentest.ps1 -Environment staging -SafeMode

# Expected duration: 30-60 minutes
# Tests: Advanced attack simulation scenarios
```

**Red Team Test Categories**:
1. **UUID Enumeration Attacks**
   - Test UUID predictability
   - Brute force UUID discovery
   - UUID collision attacks

2. **Ciphertext Tampering**
   - AEAD tag manipulation
   - Encryption key compromise
   - Replay attack simulation

3. **Authentication Bypass**
   - Password spraying attacks
   - MFA bypass attempts
   - Session hijacking simulation

4. **Privilege Escalation**
   - Role manipulation attempts
   - Admin privilege escalation
   - Access control bypass

5. **System Flooding**
   - API rate limit bypass
   - WebSocket flooding
   - Database connection exhaustion

#### Phase 3: Manual Security Assessment
```bash
# Manual testing procedures
./scripts/pentest/manual_security_checks.sh

# Focus areas:
# - Business logic testing
# - Social engineering simulation
# - Physical security assessment
```

### Test Execution Checklist

#### Before Testing
- [ ] Environment is in staging mode
- [ ] Safe mode is enabled
- [ ] Backup is completed
- [ ] Team is notified
- [ ] Monitoring is active

#### During Testing
- [ ] Monitor system performance
- [ ] Watch for unexpected behavior
- [ ] Document any anomalies
- [ ] Maintain test logs
- [ ] Follow escalation procedures if needed

#### After Testing
- [ ] Review all test results
- [ ] Generate comprehensive report
- [ ] Analyze findings and recommendations
- [ ] Plan remediation actions
- [ ] Update security procedures

### Test Result Analysis

#### Critical Findings (Immediate Action Required)
- Security vulnerabilities with CVSS score >9.0
- Authentication bypass vulnerabilities
- Data exposure or leakage
- System compromise indicators

#### High Priority Findings (Action within 24 hours)
- Security vulnerabilities with CVSS score 7.0-8.9
- Performance degradation issues
- Configuration weaknesses
- Monitoring gaps

#### Medium Priority Findings (Action within 1 week)
- Security vulnerabilities with CVSS score 4.0-6.9
- Code quality issues
- Documentation gaps
- Process improvements

#### Low Priority Findings (Action within 1 month)
- Security vulnerabilities with CVSS score <4.0
- Minor configuration issues
- Documentation updates
- Future enhancements

---

## Admin Onboarding & Revocation

### Admin Onboarding Process

#### Step 1: Invitation Creation
1. **Login as Root Admin**
   ```bash
   curl -X POST http://localhost:8080/admin/login \
     -H "Content-Type: application/json" \
     -d '{
       "email": "rootadmin@example.com",
       "password": "SecurePassword123!"
     }'
   ```

2. **Create Invitation**
   ```bash
   curl -X POST http://localhost:8080/admin/invitations \
     -H "Authorization: Bearer <session_token>" \
     -H "Content-Type: application/json" \
     -d '{
       "email": "newadmin@example.com",
       "role": "full_admin",
       "max_uses": 1
     }'
   ```

3. **Send Invitation**
   - Copy invitation token from response
   - Send secure invitation link to new admin
   - Include role expectations and responsibilities

#### Step 2: Admin Account Creation
1. **New Admin Uses Invitation**
   ```bash
   curl -X POST http://localhost:8080/admin/invitations/use \
     -H "Content-Type: application/json" \
     -d '{
       "invitation_token": "invitation_token_here",
       "password": "SecurePassword123!"
     }'
   ```

2. **Verify Account Creation**
   - Check admin_users table
   - Verify role assignment
   - Confirm audit log entry

#### Step 3: Training and Access
1. **Provide Training**
   - Dashboard navigation
   - Alert response procedures
   - Security best practices
   - Emergency procedures

2. **Grant Access**
   - Dashboard access credentials
   - API access if needed
   - Documentation access
   - Emergency contact information

### Admin Revocation Process

#### Step 1: Revocation Decision
1. **Assess Situation**
   - Review admin activity logs
   - Check for policy violations
   - Consult with security team
   - Document reasons for revocation

2. **Notify Stakeholders**
   - Inform relevant team members
   - Prepare communication plan
   - Plan for service continuity

#### Step 2: Immediate Actions
1. **Disable Admin Account**
   ```bash
   # Update admin_users table
   UPDATE admin_users 
   SET is_active = FALSE, 
       updated_at = CURRENT_TIMESTAMP 
   WHERE email = 'admin_to_revoke@example.com';
   ```

2. **Revoke Active Sessions**
   ```bash
   # Invalidate all sessions for the admin
   UPDATE admin_sessions 
   SET is_active = FALSE 
   WHERE admin_id = (
     SELECT id FROM admin_users 
     WHERE email = 'admin_to_revoke@example.com'
   );
   ```

3. **Revoke Pending Invitations**
   ```bash
   # Revoke invitations created by the admin
   DELETE FROM admin_invitation_keys 
   WHERE created_by = (
     SELECT id FROM admin_users 
     WHERE email = 'admin_to_revoke@example.com'
   );
   ```

#### Step 3: Security Review
1. **Audit Recent Activity**
   - Review admin audit logs
   - Check for suspicious actions
   - Verify no unauthorized changes
   - Document all findings

2. **Update Access Controls**
   - Review and update RBAC policies
   - Verify no privilege escalation
   - Check for orphaned permissions
   - Update security procedures

#### Step 4: Communication
1. **Notify Affected Parties**
   - Inform the revoked admin
   - Update team documentation
   - Notify relevant stakeholders
   - Update emergency contacts

2. **Document Revocation**
   - Record revocation in audit logs
   - Update admin roster
   - Document lessons learned
   - Update procedures if needed

### Emergency Admin Access

#### Scenario: All Admins Unavailable
1. **Check Backup Admins**
   - Verify backup admin availability
   - Check emergency contact list
   - Attempt to reach primary admins

2. **Emergency Procedures**
   ```bash
   # Emergency admin creation (use with caution)
   ./scripts/admin_management/emergency_admin_setup.sh
   ```

3. **Documentation**
   - Record emergency actions
   - Update incident log
   - Plan for regular admin restoration

---

## Disaster Recovery Procedures

### Backup and Recovery Overview

#### Backup Types
1. **Full System Backup**: Complete system state
2. **Database Backup**: ZKID mappings, PQC keys, audit logs
3. **Configuration Backup**: System configuration files
4. **Application Backup**: Application code and assets

#### Backup Schedule
- **Full Backup**: Daily at 2:00 AM
- **Incremental Backup**: Every 4 hours
- **Configuration Backup**: Weekly
- **Application Backup**: On deployment

### Disaster Recovery Scenarios

#### Scenario 1: Database Corruption
**Symptoms**: Database errors, data inconsistency, application failures

**Immediate Actions**:
1. **Stop Application**
   ```bash
   sudo systemctl stop secure-email-mvp
   ```

2. **Assess Damage**
   ```bash
   # Check database integrity
   sqlite3 secure_email_mvp.db "PRAGMA integrity_check;"
   
   # Check for corruption
   sqlite3 secure_email_mvp.db "SELECT COUNT(*) FROM sqlite_master;"
   ```

3. **Restore from Backup**
   ```bash
   # Run disaster recovery script
   ./scripts/operational/disaster_recovery.go --restore --backup-id latest
   ```

4. **Verify Restoration**
   ```bash
   # Verify data integrity
   ./scripts/operational/verify_restoration.sh
   
   # Start application
   sudo systemctl start secure-email-mvp
   ```

#### Scenario 2: Encryption Key Loss
**Symptoms**: PQC encryption failures, ZKID mapping errors

**Immediate Actions**:
1. **Stop Encryption Services**
   ```bash
   # Stop PQC and ZKID services
   sudo systemctl stop pqc-service
   sudo systemctl stop zkid-service
   ```

2. **Restore Encryption Keys**
   ```bash
   # Restore PQC keys from backup
   ./scripts/operational/restore_pqc_keys.sh
   
   # Restore ZKID master key
   ./scripts/operational/restore_zkid_keys.sh
   ```

3. **Verify Key Restoration**
   ```bash
   # Test encryption operations
   ./scripts/operational/test_encryption.sh
   
   # Restart services
   sudo systemctl start pqc-service
   sudo systemctl start zkid-service
   ```

#### Scenario 3: Complete System Failure
**Symptoms**: System unavailable, all services down

**Immediate Actions**:
1. **Assess Infrastructure**
   ```bash
   # Check server status
   systemctl status secure-email-mvp
   
   # Check disk space
   df -h
   
   # Check memory usage
   free -h
   ```

2. **Full System Recovery**
   ```bash
   # Run complete recovery
   ./scripts/operational/full_system_recovery.sh
   ```

3. **Verify System Health**
   ```bash
   # Run health checks
   curl http://localhost:8080/api/health
   
   # Test all services
   ./scripts/operational/test_all_services.sh
   ```

### Recovery Testing Procedures

#### Monthly Recovery Test
1. **Schedule Test**
   - Notify team members
   - Prepare test environment
   - Backup current state

2. **Execute Test**
   ```bash
   # Run recovery test
   ./scripts/operational/test_recovery.sh --scenario database_corruption
   ```

3. **Document Results**
   - Record recovery time
   - Document any issues
   - Update procedures if needed

#### Quarterly Full Recovery Test
1. **Complete System Recovery**
   ```bash
   # Full system recovery test
   ./scripts/operational/full_recovery_test.sh
   ```

2. **Performance Validation**
   - Test system performance
   - Verify all functionality
   - Check data integrity

3. **Documentation Update**
   - Update recovery procedures
   - Record lessons learned
   - Update training materials

### Backup Verification

#### Daily Backup Verification
```bash
# Verify daily backup
./scripts/operational/verify_backup.sh --backup-id daily

# Check backup integrity
./scripts/operational/check_backup_integrity.sh
```

#### Weekly Backup Testing
```bash
# Test backup restoration
./scripts/operational/test_backup_restoration.sh

# Verify data consistency
./scripts/operational/verify_data_consistency.sh
```

---

## Security Incident Response

### Incident Classification

#### Critical Incidents
- **Data Breach**: Unauthorized access to sensitive data
- **System Compromise**: Malicious code or unauthorized access
- **Service Outage**: Complete system unavailability
- **Encryption Failure**: Loss of encryption keys or capabilities

#### High Priority Incidents
- **Authentication Bypass**: Successful bypass of security controls
- **Performance Attack**: DDoS or resource exhaustion
- **Configuration Breach**: Unauthorized configuration changes
- **Audit Tampering**: Attempts to modify audit logs

#### Medium Priority Incidents
- **Failed Login Attempts**: Multiple failed authentication attempts
- **Suspicious Activity**: Unusual system behavior
- **Performance Degradation**: System performance issues
- **Configuration Drift**: Unauthorized configuration changes

### Incident Response Process

#### Phase 1: Detection and Assessment
1. **Incident Detection**
   - Monitor alerts and dashboards
   - Review security logs
   - Check system health
   - Verify incident details

2. **Initial Assessment**
   - Determine incident severity
   - Identify affected systems
   - Assess potential impact
   - Notify appropriate personnel

#### Phase 2: Containment
1. **Immediate Containment**
   - Isolate affected systems
   - Block malicious traffic
   - Disable compromised accounts
   - Preserve evidence

2. **System Stabilization**
   - Restore critical services
   - Implement temporary fixes
   - Monitor system stability
   - Document containment actions

#### Phase 3: Eradication
1. **Root Cause Analysis**
   - Investigate incident cause
   - Identify vulnerabilities
   - Document findings
   - Plan remediation

2. **System Recovery**
   - Apply security patches
   - Update configurations
   - Restore from clean backups
   - Verify system integrity

#### Phase 4: Recovery
1. **Service Restoration**
   - Gradually restore services
   - Monitor system performance
   - Verify functionality
   - Update stakeholders

2. **Post-Incident Review**
   - Conduct incident review
   - Document lessons learned
   - Update procedures
   - Plan improvements

### Specific Incident Response Procedures

#### Data Breach Response
1. **Immediate Actions**
   - Isolate affected systems
   - Preserve evidence
   - Notify security team
   - Begin investigation

2. **Investigation**
   - Determine breach scope
   - Identify compromised data
   - Trace attack vector
   - Document findings

3. **Remediation**
   - Patch vulnerabilities
   - Update security controls
   - Restore from clean backups
   - Implement monitoring

#### Authentication Bypass Response
1. **Immediate Actions**
   - Disable affected accounts
   - Review access logs
   - Check for unauthorized changes
   - Notify affected users

2. **Investigation**
   - Identify bypass method
   - Review authentication code
   - Check for similar vulnerabilities
   - Document findings

3. **Remediation**
   - Fix authentication logic
   - Update security controls
   - Implement additional monitoring
   - Test fixes thoroughly

#### System Compromise Response
1. **Immediate Actions**
   - Isolate compromised systems
   - Preserve evidence
   - Notify security team
   - Begin investigation

2. **Investigation**
   - Determine compromise scope
   - Identify attack vector
   - Check for persistence
   - Document findings

3. **Remediation**
   - Remove malicious code
   - Patch vulnerabilities
   - Restore from clean backups
   - Implement monitoring

### Communication Procedures

#### Internal Communication
1. **Immediate Notification**
   - Security team
   - System administrators
   - Management team
   - Legal team (if needed)

2. **Status Updates**
   - Regular updates to stakeholders
   - Progress reports
   - Resolution timeline
   - Post-incident summary

#### External Communication
1. **Customer Notification**
   - Transparent communication
   - Impact assessment
   - Remediation steps
   - Contact information

2. **Regulatory Reporting**
   - Required notifications
   - Compliance reporting
   - Documentation requirements
   - Follow-up actions

---

## Maintenance Procedures

### Daily Maintenance Tasks

#### System Health Check
```bash
# Daily health check script
./scripts/maintenance/daily_health_check.sh

# Check system resources
./scripts/maintenance/check_system_resources.sh

# Verify service status
./scripts/maintenance/verify_services.sh
```

#### Log Review
```bash
# Review security logs
./scripts/maintenance/review_security_logs.sh

# Check for anomalies
./scripts/maintenance/check_log_anomalies.sh

# Archive old logs
./scripts/maintenance/archive_logs.sh
```

### Weekly Maintenance Tasks

#### Performance Review
```bash
# Performance analysis
./scripts/maintenance/performance_analysis.sh

# Database optimization
./scripts/maintenance/optimize_database.sh

# Clean up temporary files
./scripts/maintenance/cleanup_temp_files.sh
```

#### Security Review
```bash
# Security assessment
./scripts/maintenance/security_assessment.sh

# Update security patches
./scripts/maintenance/update_security_patches.sh

# Review access controls
./scripts/maintenance/review_access_controls.sh
```

### Monthly Maintenance Tasks

#### Comprehensive Review
```bash
# Full system review
./scripts/maintenance/full_system_review.sh

# Backup verification
./scripts/maintenance/verify_backups.sh

# Update documentation
./scripts/maintenance/update_documentation.sh
```

#### Disaster Recovery Test
```bash
# Recovery test
./scripts/maintenance/test_disaster_recovery.sh

# Update recovery procedures
./scripts/maintenance/update_recovery_procedures.sh
```

### Quarterly Maintenance Tasks

#### Security Assessment
```bash
# Penetration testing
./scripts/pentest/run_security_tests.ps1

# Security audit
./scripts/maintenance/security_audit.sh

# Update security policies
./scripts/maintenance/update_security_policies.sh
```

#### Performance Optimization
```bash
# Performance tuning
./scripts/maintenance/performance_tuning.sh

# Capacity planning
./scripts/maintenance/capacity_planning.sh

# Update monitoring
./scripts/maintenance/update_monitoring.sh
```

---

## Troubleshooting Guide

### Common Issues and Solutions

#### Issue 1: High API Latency
**Symptoms**: Slow response times, timeout errors

**Diagnosis**:
```bash
# Check system resources
top -p $(pgrep secure-email-mvp)

# Check database performance
sqlite3 secure_email_mvp.db "PRAGMA stats;"

# Check network connectivity
ping -c 5 localhost
```

**Solutions**:
1. Restart application services
2. Optimize database queries
3. Scale system resources
4. Check for external dependencies

#### Issue 2: ZKID Mapping Failures
**Symptoms**: UUID mapping creation errors

**Diagnosis**:
```bash
# Check ZKID service status
systemctl status zkid-service

# Check database connectivity
sqlite3 secure_email_mvp.db "SELECT COUNT(*) FROM zkid_mappings;"

# Check encryption keys
./scripts/troubleshooting/check_encryption_keys.sh
```

**Solutions**:
1. Restart ZKID service
2. Check database locks
3. Verify encryption key availability
4. Check system resources

#### Issue 3: PQC Key Rotation Failures
**Symptoms**: Key rotation errors, encryption failures

**Diagnosis**:
```bash
# Check PQC service status
systemctl status pqc-service

# Check key rotation logs
tail -f /var/log/pqc-service.log

# Verify HSM connectivity
./scripts/troubleshooting/check_hsm_connectivity.sh
```

**Solutions**:
1. Restart PQC service
2. Check HSM connectivity
3. Manually trigger key rotation
4. Verify system time synchronization

#### Issue 4: Admin Dashboard Access Issues
**Symptoms**: Dashboard login failures, session errors

**Diagnosis**:
```bash
# Check admin service status
systemctl status admin-service

# Check database connectivity
sqlite3 secure_email_mvp.db "SELECT COUNT(*) FROM admin_users;"

# Check session logs
tail -f /var/log/admin-service.log
```

**Solutions**:
1. Restart admin service
2. Check database connectivity
3. Clear expired sessions
4. Verify authentication configuration

### Emergency Procedures

#### Emergency System Restart
```bash
# Emergency restart script
./scripts/emergency/emergency_restart.sh

# This will:
# 1. Stop all services gracefully
# 2. Check system health
# 3. Restart services in order
# 4. Verify system functionality
```

#### Emergency Database Recovery
```bash
# Emergency database recovery
./scripts/emergency/emergency_db_recovery.sh

# This will:
# 1. Stop application services
# 2. Restore database from backup
# 3. Verify data integrity
# 4. Restart services
```

#### Emergency Admin Access
```bash
# Emergency admin access
./scripts/emergency/emergency_admin_access.sh

# This will:
# 1. Create temporary admin account
# 2. Grant emergency access
# 3. Log all actions
# 4. Set expiration time
```

---

## Emergency Contacts

### Primary Contacts

| Role | Name | Email | Phone | Availability |
|------|------|-------|-------|--------------|
| Root Admin | [Name] | rootadmin@example.com | [Phone] | 24/7 |
| Security Lead | [Name] | security@example.com | [Phone] | 24/7 |
| System Admin | [Name] | sysadmin@example.com | [Phone] | Business Hours |
| DevOps Lead | [Name] | devops@example.com | [Phone] | 24/7 |

### Escalation Procedures

#### Level 1: On-Call Team
- **Response Time**: 15 minutes
- **Contact**: On-call rotation
- **Actions**: Initial assessment and containment

#### Level 2: Security Team
- **Response Time**: 30 minutes
- **Contact**: Security team lead
- **Actions**: Security investigation and response

#### Level 3: Management
- **Response Time**: 1 hour
- **Contact**: CTO/CIO
- **Actions**: Strategic decisions and external communication

#### Level 4: Executive
- **Response Time**: 2 hours
- **Contact**: CEO/CTO
- **Actions**: Business impact assessment and external relations

### External Contacts

| Service | Contact | Phone | Email |
|---------|---------|-------|-------|
| Cloud Provider | [Provider] | [Phone] | [Email] |
| Security Vendor | [Vendor] | [Phone] | [Email] |
| Legal Counsel | [Law Firm] | [Phone] | [Email] |
| PR Agency | [Agency] | [Phone] | [Email] |

---

## Document Control

### Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2024-01-01 | [Author] | Initial version |
| 1.1 | 2024-01-15 | [Author] | Updated procedures |
| 1.2 | 2024-02-01 | [Author] | Added incident response |

### Review Schedule

- **Monthly**: Review and update procedures
- **Quarterly**: Comprehensive review and update
- **Annually**: Major revision and validation

### Distribution

- **Primary**: All admin users
- **Secondary**: Security team members
- **Tertiary**: Management team

---

**Document Status**: ✅ Complete  
**Last Updated**: 2024-01-01  
**Next Review**: 2024-02-01  
**Approved By**: [Name]  
**Version**: 1.0
