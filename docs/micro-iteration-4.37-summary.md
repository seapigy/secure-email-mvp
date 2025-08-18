# Micro-Iteration 4.37: Zero-Knowledge Identity Layer (ZKID)

## Overview

Micro-Iteration 4.37 implements a Zero-Knowledge Identity Layer that provides maximum privacy for user email mappings while maintaining operational functionality. The system ensures that internal staff can only see UUIDs and never have access to external email addresses, while still providing account recovery capabilities.

## Key Features

### Zero-Knowledge Email Mapping
- **Encrypted Storage**: All email mappings are encrypted using AES-256-GCM with per-record data keys
- **Master Key Wrapping**: Data keys are wrapped with a master key for secure storage
- **Peppered Hashing**: Email addresses are hashed with a secret pepper for lookup without revealing the original
- **UUID-Only Visibility**: Internal operations only see user UUIDs, never external emails

### Recovery Code System
- **Bitwarden-Style Codes**: One-time recovery codes for account recovery
- **Argon2id Hashing**: Secure storage of recovery code hashes with salt and pepper
- **Atomic Operations**: Code validation and consumption in single database transactions
- **Admin Management**: System admins can generate and revoke recovery codes

### Admin-Facing Endpoints
- **RBAC Protection**: All admin endpoints require system_admin or enterprise_admin role
- **UUID-Only Operations**: No external email addresses are ever exposed to staff
- **Audit Logging**: All admin actions are logged with UUID-only identifiers
- **Statistics**: Admin monitoring of ZKID usage and recovery code statistics

## Architecture

### Database Schema

#### zkid_email_mappings
```sql
CREATE TABLE zkid_email_mappings (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    email_hash TEXT NOT NULL UNIQUE, -- SHA-256 of normalized email with pepper
    email_ciphertext BLOB NOT NULL,
    email_nonce BLOB NOT NULL,
    email_tag BLOB NOT NULL,
    wrapped_key BLOB NOT NULL,           -- per-record data key wrapped with master key
    wrapped_key_nonce BLOB NOT NULL,
    wrapped_key_tag BLOB NOT NULL,
    fallback_email_ciphertext BLOB,
    fallback_email_nonce BLOB,
    fallback_email_tag BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### zkid_recovery_codes
```sql
CREATE TABLE zkid_recovery_codes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    salt BLOB NOT NULL,
    hash BLOB NOT NULL, -- Argon2id hash of code+pepper with salt
    used BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Core Components

#### ZKID Service (`pkg/zkid/zkid.go`)
- **Email Mapping**: Create, update, and retrieve encrypted email mappings
- **Key Management**: Generate and wrap data keys with master key
- **Encryption**: AES-256-GCM encryption/decryption with proper nonce handling
- **Hashing**: SHA-256 email hashing with pepper for lookup

#### Recovery System (`pkg/zkid/recovery.go`)
- **Code Generation**: Create one-time recovery codes with Argon2id hashing
- **Validation**: Atomic validation and consumption of recovery codes
- **Revocation**: Admin ability to revoke specific recovery codes

#### Configuration (`pkg/zkid/env.go`)
- **Environment Variables**: Load configuration from environment
- **Feature Flags**: Enable/disable ZKID functionality
- **Key Management**: Master key and pepper configuration

## API Endpoints

### Internal ZKID Operations
- `POST /api/zkid/mapping` - Create or update email mapping
- `GET /api/zkid/email?user_id=<uuid>` - Retrieve email by user ID

### Admin ZKID Operations (RBAC Protected)
- `GET /api/admin/zkid/recovery-codes?user_id=<uuid>&count=<n>` - Generate recovery codes
- `POST /api/admin/zkid/revoke-code` - Revoke specific recovery code
- `GET /api/admin/zkid/stats` - Get ZKID statistics

## Environment Variables

### Required for ZKID Operation
```bash
# Enable ZKID layer
ZKID_ENABLED=true

# 32-byte master key (hex encoded)
ZKID_MASTER_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef

# Email hash pepper (hex encoded)
ZKID_EMAIL_HASH_PEPPER=deadbeefdeadbeefdeadbeefdeadbeef

# Recovery code pepper (hex encoded)
ZKID_RECOVERY_PEPPER=feedbeeffeedbeeffeedbeeffeedbeef
```

## Security Features

### Zero-Knowledge Guarantees
- **Staff Isolation**: Internal staff never see external email addresses
- **Encrypted Storage**: All sensitive data encrypted at rest
- **Key Separation**: Master key separate from application keys
- **Audit Trail**: All operations logged with UUID-only identifiers

### Cryptographic Security
- **AES-256-GCM**: Authenticated encryption for email data
- **Argon2id**: Memory-hard hashing for recovery codes
- **Key Wrapping**: Secure storage of data keys
- **Peppered Hashing**: Protection against rainbow table attacks

### Access Control
- **RBAC Enforcement**: Admin endpoints require proper role
- **UUID-Only Context**: All operations use internal UUIDs
- **Audit Logging**: Comprehensive logging of all admin actions

## Integration Points

### Signup Flow
When ZKID is enabled, the signup handler automatically:
1. Creates the user account with UUID
2. Creates encrypted email mapping
3. Stores fallback email if provided
4. Logs operation with UUID-only identifiers

### Authentication Flow
The login flow remains unchanged but can optionally:
1. Use ZKID to retrieve email for external operations
2. Maintain UUID-only internal operations
3. Support recovery code validation

## Testing

### Unit Tests
- **Email Mapping**: Round-trip encryption/decryption
- **Recovery Codes**: Generation, validation, and revocation
- **Statistics**: Admin monitoring functionality
- **Configuration**: Environment variable loading

### Integration Tests
- **Admin Access**: RBAC enforcement for admin endpoints
- **UUID-Only Operations**: Verification of privacy guarantees
- **Recovery Flow**: End-to-end recovery code usage
- **Feature Flags**: Proper enable/disable behavior

## Migration Strategy

### Feature Flag Control
- **Environment Variable**: `ZKID_ENABLED` controls entire layer
- **Backward Compatibility**: Existing authentication continues to work
- **Gradual Rollout**: Can be enabled per environment
- **Rollback Capability**: Instant disable via environment variable

### Data Migration
- **Automatic Mapping**: New signups automatically create ZKID mappings
- **Existing Users**: Can be migrated via admin tools
- **Dual Mode**: System supports both ZKID and legacy modes
- **No Data Loss**: All existing functionality preserved

## Monitoring and Compliance

### Admin Statistics
- **Mapping Counts**: Total encrypted email mappings
- **Recovery Codes**: Total, used, and available codes
- **Usage Metrics**: Admin operation frequency
- **System Health**: ZKID layer status and errors

### Audit Logging
- **Admin Actions**: All admin operations logged with UUIDs
- **Security Events**: Failed access attempts and violations
- **Privacy Compliance**: Zero-knowledge guarantees maintained
- **Operational Monitoring**: System performance and errors

## Future Enhancements

### PQC Integration
- **Hybrid Encryption**: Combine with PQC layer for post-quantum security
- **Key Management**: Integrate with PQC key manager
- **Performance**: Optimize for quantum-resistant algorithms

### Advanced Features
- **Multi-Factor Recovery**: Additional recovery mechanisms
- **Temporal Access**: Time-limited admin access
- **Compliance Reporting**: Enhanced audit and compliance features
- **Performance Optimization**: Caching and optimization strategies

## Conclusion

Micro-Iteration 4.37 successfully implements a Zero-Knowledge Identity Layer that provides maximum privacy for user email mappings while maintaining full operational functionality. The system ensures that internal staff can perform all necessary operations without ever seeing external email addresses, while providing robust account recovery capabilities through secure recovery codes.

The implementation maintains backward compatibility, supports gradual rollout through feature flags, and provides comprehensive monitoring and audit capabilities. All security features are properly tested and documented, ensuring the system meets enterprise-grade privacy and security requirements.
