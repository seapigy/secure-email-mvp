# Sprint 0: Full E2E PQC & Operational Hardening - Design Document

## 📋 **Executive Summary**

This document outlines the design and architecture for implementing a production-grade end-to-end PQC (Post-Quantum Cryptography) system with metadata minimization, per-thread encryption, key transparency, threshold HSM, and comprehensive operational hardening for the Secure Email MVP.

**Status**: Design Phase  
**Sprint**: 0 (Design + Prototyping)  
**Timeline**: 1 week  
**Risk Level**: HIGH (requires careful implementation with safety measures)

## 🎯 **Goals & Objectives**

### Primary Goals
1. **End-to-End Encryption**: Implement client → client encryption using hybrid PQC + AEAD scheme
2. **Metadata Minimization**: Server stores only ciphertext + minimal routing metadata
3. **Key Transparency**: Auditable and tamper-evident public key mappings
4. **Threshold HSM**: Multi-operator key management to prevent single-operator compromise
5. **Operational Safety**: Feature flags, dual-mode compatibility, exhaustive testing, observability

### Success Criteria
- ✅ Server cannot decrypt E2E messages when E2E is enabled
- ✅ 100% backward compatibility with existing AES-256-GCM system
- ✅ Zero plaintext exposure in logs or server storage
- ✅ Comprehensive test coverage with race detection
- ✅ Instant rollback capability via feature flags

## 🏗️ **System Architecture**

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client SDK    │    │   Key Transp.   │    │   Threshold     │
│   (JS/Go)       │◄──►│   Service       │◄──►│   HSM           │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   E2E Protocol  │    │   Merkle Tree   │    │   Key Manager   │
│   (PQC Hybrid)  │    │   (CT-like)     │    │   (M-of-N)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Server API    │    │   Migration     │    │   Observability │
│   (Dual Mode)   │    │   Worker        │    │   & Monitoring  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Component Breakdown

#### 1. **Crypto & Protocol Layer**
- **KEM**: Kyber (liboqs binding) with classical fallback
- **DEM**: AES-256-GCM + ChaCha20-Poly1305 (dual encryption)
- **Signatures**: Dilithium for post-quantum authenticity
- **Key Derivation**: HKDF with thread context and sequence numbers

#### 2. **Key Transparency Service**
- **Append-only Log**: Merkle tree structure (CT-like)
- **Verification**: REST endpoints for proof verification
- **Monitoring**: Tamper-evident logging and alerting

#### 3. **Threshold HSM Integration**
- **Key Wrapping**: M-of-N threshold operations
- **HSM Mock**: Software fallback for development
- **Audit Trail**: All key operations logged and monitored

#### 4. **Client SDK (JS/Go)**
- **Encrypt-before-upload**: Client-side encryption
- **Decrypt-after-download**: Client-side decryption
- **KT Verification**: Automatic key verification
- **Retry Logic**: Exponential backoff and error handling

## 🔐 **Security Design**

### Encryption Protocol

#### Message Envelope Format (Versioned)
```json
{
  "envelope_version": "1.0",
  "ciphertext_body": "base64_encoded",
  "ciphertext_headers": "base64_encoded",
  "pub_kem": "base64_encoded",
  "enc_metadata": "base64_encoded",
  "aad_hash": "sha256_hash",
  "thread_id": "uuid",
  "sequence_number": 123,
  "key_rotation_id": "uuid",
  "signature": "base64_encoded_dilithium"
}
```

#### Per-Thread Key Derivation
```
Thread Key = HKDF(Thread Secret, "thread_key", Thread ID)
Message Key = HKDF(Thread Key, "message_key", Sequence Number)
```

#### Hybrid Encryption Flow
1. **Key Encapsulation**: Kyber KEM generates shared secret
2. **Key Derivation**: HKDF derives symmetric keys
3. **Dual Encryption**: AES-256-GCM + ChaCha20-Poly1305
4. **Authentication**: Dilithium signature for authenticity
5. **Metadata Protection**: Encrypted headers and minimal routing

### Metadata Minimization Strategy

#### Server-Side Storage (Minimal)
```sql
-- Only essential routing information
CREATE TABLE e2e_messages (
    id TEXT PRIMARY KEY,
    recipient_uuid TEXT NOT NULL,  -- ZKID UUID only
    routing_token TEXT,            -- Ephemeral routing token
    envelope_hash TEXT NOT NULL,   -- Hash of encrypted envelope
    created_at TIMESTAMP,
    expires_at TIMESTAMP
);
```

#### Encrypted Metadata (Client-Side)
- Subject lines
- Attachment metadata
- Sensitive headers
- Search tokens (hashed/blinded)

## 🔧 **Technical Implementation**

### Feature Flags & Configuration

#### Environment Variables
```bash
# Global Feature Flags
E2E_ENABLED=false                    # Global E2E enable/disable
E2E_ORG_ENABLED_<org_id>=false       # Per-org enable
E2E_USER_ENABLED_<user_id>=false     # Per-user opt-in

# Crypto Configuration
PQC_LIBRARY=liboqs                   # PQC library to use
KEM_PARAMS=kyber768                  # KEM parameters
DEMO_PLAINTEXT_MODE=false            # NEVER true in production

# Key Transparency
KT_ENABLED=false                     # Key Transparency service
KT_LOG_URL=https://kt.example.com    # KT log endpoint

# HSM Configuration
HSM_ENABLED=false                    # HSM integration
HSM_THRESHOLD_M=3                    # M-of-N threshold
HSM_THRESHOLD_N=5                    # Total operators

# Debug & Monitoring
E2E_DEBUG=false                      # Debug logging
E2E_OBSERVABILITY=true               # Observability features
```

### API Design

#### New Endpoints (Backwards Compatible)
```go
// Key Management
POST   /api/e2e/keys/publish          // Publish public key to KT
GET    /api/e2e/keys/{user_id}        // Get user's public key
GET    /api/e2e/kt/proof/{user_id}    // Get KT proof for user

// Message Operations
POST   /api/e2e/messages/send         // Send E2E encrypted message
GET    /api/e2e/messages/{id}         // Get E2E encrypted message
POST   /api/e2e/messages/migrate      // Migrate legacy message to E2E

// Migration & Management
GET    /api/e2e/migration/status      // Migration status
POST   /api/e2e/migration/start       // Start background migration
POST   /api/e2e/migration/rollback    // Rollback migration
```

#### Request/Response Headers
```http
X-E2E-Mode: hybrid|legacy|dual        # Encryption mode
X-Correlation-ID: uuid                # Request correlation
X-Trace-ID: uuid                      # Distributed tracing
```

### Database Schema Extensions

#### E2E Messages Table
```sql
CREATE TABLE e2e_messages (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    sequence_number INTEGER NOT NULL,
    sender_uuid TEXT NOT NULL,
    recipient_uuid TEXT NOT NULL,
    envelope_hash TEXT NOT NULL,
    envelope_version TEXT NOT NULL,
    key_rotation_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    
    -- Routing metadata (minimal)
    routing_token TEXT,
    delivery_status TEXT DEFAULT 'pending',
    
    -- Encryption metadata
    kem_algorithm TEXT NOT NULL,
    dem_algorithm TEXT NOT NULL,
    signature_algorithm TEXT NOT NULL,
    
    -- Audit fields
    correlation_id TEXT,
    trace_id TEXT,
    
    UNIQUE(thread_id, sequence_number)
);

-- Indexes for performance
CREATE INDEX idx_e2e_messages_thread ON e2e_messages(thread_id);
CREATE INDEX idx_e2e_messages_recipient ON e2e_messages(recipient_uuid);
CREATE INDEX idx_e2e_messages_created ON e2e_messages(created_at);
```

#### Key Transparency Tables
```sql
CREATE TABLE kt_public_keys (
    id TEXT PRIMARY KEY,
    user_uuid TEXT NOT NULL,
    public_key TEXT NOT NULL,
    key_type TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    merkle_index INTEGER,
    merkle_proof TEXT,
    
    UNIQUE(user_uuid, key_type)
);

CREATE TABLE kt_log_entries (
    id TEXT PRIMARY KEY,
    entry_hash TEXT NOT NULL,
    merkle_index INTEGER NOT NULL,
    merkle_root TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    signature TEXT NOT NULL
);
```

#### Threshold HSM Tables
```sql
CREATE TABLE hsm_key_operations (
    id TEXT PRIMARY KEY,
    operation_type TEXT NOT NULL,
    key_id TEXT NOT NULL,
    threshold_m INTEGER NOT NULL,
    threshold_n INTEGER NOT NULL,
    operator_signatures TEXT, -- JSON array of signatures
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);
```

## 🧪 **Testing Strategy**

### Test Categories

#### 1. **Unit Tests**
- Crypto primitives (KEM, DEM, signatures)
- Envelope serialization/deserialization
- Key derivation functions
- KT proof verification

#### 2. **Integration Tests**
- Client-to-client encryption roundtrip
- Server cannot decrypt E2E messages
- KT log append/verify operations
- Threshold unwrap operations

#### 3. **Concurrency Tests**
- Race detection on key operations
- Concurrent message encryption/decryption
- KT log concurrent updates

#### 4. **Performance Tests**
- PQC operation latency benchmarks
- Load testing under high concurrency
- Memory usage profiling

#### 5. **Security Tests**
- Penetration testing integration
- Fuzz testing for malformed inputs
- Side-channel analysis

### Test Infrastructure

#### Local Development Environment
```yaml
# docker-compose.test.yml
version: '3.8'
services:
  e2e-test-harness:
    build: ./test/harness
    environment:
      - E2E_DEBUG=true
      - DEMO_PLAINTEXT_MODE=false
      - HSM_ENABLED=false
    ports:
      - "8080:8080"
  
  kt-service:
    build: ./kt-service
    environment:
      - KT_LOG_URL=http://localhost:8081
    ports:
      - "8081:8081"
```

## 📊 **Observability & Monitoring**

### Structured Logging

#### Log Format (JSON)
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "correlation_id": "uuid",
  "trace_id": "uuid",
  "service": "e2e-encryption",
  "operation": "encrypt_message",
  "user_id": "uuid_only",
  "thread_id": "uuid",
  "sequence_number": 123,
  "encryption_method": "hybrid_pqc",
  "kyber_level": 768,
  "duration_ms": 45,
  "success": true,
  "error_code": null,
  "metadata": {
    "aad_hash": "sha256_hash",
    "envelope_version": "1.0"
  }
}
```

### Metrics & Alerting

#### Critical Metrics
- PQC encryption/decryption latency
- KT append/verify success rates
- HSM operation success rates
- E2E message delivery success rates

#### Alert Rules
```yaml
# Prometheus Alert Rules
groups:
  - name: e2e_critical_alerts
    rules:
      - alert: KTAppendFailureRate
        expr: rate(kt_append_failures_total[5m]) > 0.005
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "KT append failure rate > 0.5%"
          
      - alert: PQCDecryptionFailures
        expr: rate(pqc_decrypt_failures_total[5m]) > 10
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "PQC decryption failures spiking"
```

### Dashboards

#### Grafana Dashboard Panels
1. **E2E Encryption Overview**
   - Messages encrypted per minute
   - Encryption latency percentiles
   - Success/failure rates

2. **Key Transparency Health**
   - KT log append rate
   - Proof verification success rate
   - Merkle tree size and growth

3. **HSM Operations**
   - Threshold operation success rate
   - Key rotation status
   - Operator availability

## 🔄 **Migration & Rollback Strategy**

### Migration Phases

#### Phase 1: Dual-Mode Deployment
- Deploy code that understands both legacy and E2E
- Feature flags default to OFF
- No impact on existing functionality

#### Phase 2: Key Publication
- Publish public keys to KT
- Verify KT proofs
- Enable E2E for test users

#### Phase 3: Gradual Rollout
- Canary deployment (5% → 25% → 100%)
- Monitor health metrics
- Automatic rollback on issues

#### Phase 4: Background Migration
- Migrate existing messages to E2E
- Dual-mode reading during migration
- Verify no data loss

### Rollback Procedures

#### Instant Rollback (Feature Flag)
```bash
# Disable E2E globally
export E2E_ENABLED=false

# Disable per-org
export E2E_ORG_ENABLED_org123=false

# Disable per-user
export E2E_USER_ENABLED_user456=false
```

#### Database Rollback
```sql
-- Restore from backup
RESTORE DATABASE FROM 'backup_before_e2e.sql';

-- Verify data integrity
SELECT COUNT(*) FROM emails WHERE e2e_enabled = true;
```

## 🚨 **Risk Assessment & Mitigation**

### High-Risk Areas

#### 1. **Data Loss During Migration**
- **Risk**: Migration process corrupts existing data
- **Mitigation**: Comprehensive backups, dry-run testing, dual-mode reading

#### 2. **Performance Degradation**
- **Risk**: PQC operations significantly slower than AES
- **Mitigation**: Performance benchmarks, gradual rollout, fallback mechanisms

#### 3. **Key Management Failures**
- **Risk**: HSM or KT service failures
- **Mitigation**: Redundant systems, monitoring, automatic failover

#### 4. **Logging Plaintext**
- **Risk**: Accidental plaintext exposure in logs
- **Mitigation**: Automated log scanning, structured logging, UUID-only policies

### Safety Measures

#### 1. **Defensive Programming**
- All crypto operations wrapped in try-catch
- Graceful degradation on failures
- Comprehensive error handling

#### 2. **Monitoring & Alerting**
- Real-time monitoring of all critical operations
- Automated alerting for anomalies
- Detailed failure reporting

#### 3. **Testing & Validation**
- Exhaustive test coverage
- Automated security scanning
- Penetration testing integration

## 📅 **Implementation Timeline**

### Sprint Breakdown

#### Sprint 0 (Current): Design + Prototyping (1 week)
- ✅ Finalize envelope specification
- ✅ Design KT API and key management
- ✅ Create test plan and infrastructure
- ✅ Set up development environment

#### Sprint 1: Core KEM/DEM + Envelope + Client SDK (2 weeks)
- Implement basic hybrid encrypt/decrypt
- Create message envelope format
- Develop client SDK (JS/Go)
- Unit tests for crypto primitives

#### Sprint 2: KT & Publish/Verify + CI Tests (1 week)
- Implement Key Transparency service
- Create KT verification endpoints
- Add CI integration tests
- Performance benchmarking

#### Sprint 3: Key Manager + Rotation + HSM Simulation (2 weeks)
- Implement key management system
- Add key rotation logic
- Create HSM simulation for testing
- Threshold operations

#### Sprint 4: Server API + Migration Worker + Dual-Mode (2 weeks)
- Extend server API for E2E
- Implement background migration worker
- Add dual-mode reading capability
- Integration testing

#### Sprint 5: Performance + Load Tests + Pentest Integration (1 week)
- Performance optimization
- Load testing under stress
- Integrate with existing pentest framework
- Security validation

#### Sprint 6: Canary Rollout + Monitoring + Runbooks (1 week)
- Implement canary deployment automation
- Create monitoring dashboards
- Write operational runbooks
- Documentation

#### Sprint 7: Staging Pentest + Final Adjustments (1 week)
- Complete staging penetration testing
- Address security findings
- Final performance tuning
- Production readiness validation

## 📋 **Acceptance Criteria**

### Must Pass All Tests
- [ ] All unit/integration/e2e tests pass in CI
- [ ] Race detector enabled and passing
- [ ] Server cannot decrypt E2E messages (asserted in CI)
- [ ] KT log append & verify tests pass
- [ ] Key rotation and threshold unwrap tests succeed

### Security Requirements
- [ ] No plaintext exposure in logs
- [ ] KT equivocation detection working
- [ ] HSM threshold operator alerts configured
- [ ] Penetration test results gate promotion

### Performance Requirements
- [ ] PQC operations < 100ms average latency
- [ ] No significant performance regression
- [ ] Memory usage within acceptable limits
- [ ] Database query performance maintained

### Operational Requirements
- [ ] Feature flags working correctly
- [ ] Instant rollback capability
- [ ] Monitoring dashboards populated
- [ ] Alert rules firing correctly

## 🔧 **Development Environment Setup**

### Prerequisites
- Go 1.23+
- Node.js 18+
- Docker & Docker Compose
- SQLite3
- liboqs (PQC library)

### Local Development
```bash
# Clone and setup
git clone <repository>
cd secure-email-mvp

# Install dependencies
go mod download
npm install

# Setup development environment
cp .env.example .env
# Edit .env with development settings

# Run tests
go test ./... -race
npm test

# Start development server
go run cmd/api/main.go
```

### Testing Environment
```bash
# Run full test suite
./scripts/run_e2e_tests.sh

# Run performance benchmarks
go test ./pkg/pqc -bench=.

# Run security tests
./scripts/run_pentest.sh
```

## 📚 **Documentation & References**

### Technical References
- [RFC 6238: TOTP](https://tools.ietf.org/html/rfc6238)
- [Kyber Specification](https://pq-crystals.org/kyber/)
- [Dilithium Specification](https://pq-crystals.org/dilithium/)
- [Key Transparency RFC](https://tools.ietf.org/html/draft-ietf-trans-rfc6962-bis)

### Security Standards
- [NIST Post-Quantum Cryptography](https://csrc.nist.gov/projects/post-quantum-cryptography)
- [OWASP Cryptographic Storage](https://owasp.org/www-project-cheat-sheets/cheatsheets/Cryptographic_Storage_Cheat_Sheet)
- [Cloudflare Key Transparency](https://developers.cloudflare.com/ssl/key-transparency/)

### Implementation Guides
- [liboqs Go Bindings](https://github.com/open-quantum-safe/liboqs-go)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Prometheus Alerting](https://prometheus.io/docs/alerting/latest/overview/)

---

**Document Version**: 1.0  
**Last Updated**: 2024-01-15  
**Next Review**: Sprint 1 Planning  
**Approval Status**: Pending Security Review
