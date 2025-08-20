# Sprint 4: Server API Integration + Migration Worker + Dual-mode Tests - Design Document

## 📋 **Executive Summary**

This document outlines the comprehensive server API integration for the E2E PQC system, including production server integration, migration worker implementation, and dual-mode testing strategy.

**Status**: Implementation Phase  
**Sprint**: 4 (Server Integration)  
**Timeline**: 2 weeks  
**Risk Level**: MEDIUM (integration with existing production system)

## 🎯 **Sprint 4 Goals**

### Primary Objectives
1. **Complete Server API Integration**: Integrate E2E endpoints into main API server with full backwards compatibility
2. **Migration Worker Implementation**: Build background migration system for legacy data conversion
3. **Dual-mode Testing**: Comprehensive testing of legacy/E2E message handling
4. **Production Readiness**: Ensure system can handle both modes seamlessly

### Success Criteria
- Server can handle both legacy and E2E messages simultaneously
- Migration worker successfully re-encrypts legacy data without downtime
- All dual-mode tests pass with 100% backwards compatibility
- Performance meets production requirements
- Zero data loss during migration

## 🏗️ **Architecture Overview**

```
┌─────────────────────────────────────────────────────────────┐
│                    Sprint 4 Architecture                    │
├─────────────────────────────────────────────────────────────┤
│  Main API Server  │    E2E Integration    │   Migration     │
│  ┌─────────────┐ │ ┌─────────────────────┐ │ ┌─────────────┐ │
│  │ Existing    │ │ │ E2E Endpoints       │ │ │ Background  │ │
│  │ Endpoints   │ │ │ (New Routes)        │ │ │ Worker      │ │
│  └─────────────┘ │ └─────────────────────┘ │ └─────────────┘ │
│  ┌─────────────┐ │ ┌─────────────────────┐ │ ┌─────────────┐ │
│  │ Dual-mode   │ │ │ Feature Flags       │ │ │ Job Queue   │ │
│  │ Handler     │ │ │ (Global/Org/User)   │ │ │ Management  │ │
│  └─────────────┘ │ └─────────────────────┘ │ └─────────────┘ │
│  ┌─────────────┐ │ ┌─────────────────────┐ │ ┌─────────────┐ │
│  │ Database    │ │ │ Migration Tracking  │ │ │ Progress    │ │
│  │ Integration │ │ │ (Schema Updates)    │ │ │ Monitoring  │ │
│  └─────────────┘ │ └─────────────────────┘ │ └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │   Dual-mode      │
                    │   Testing        │
                    │   Framework      │
                    └──────────────────┘
```

## 🔧 **Component 1: Server API Integration**

### Design Goals
- Seamless integration with existing API server
- Full backwards compatibility
- Feature flag controlled rollout
- Comprehensive error handling and logging

### Implementation Strategy

#### 1.1 E2E Endpoint Integration

**New API Routes to Add:**
```go
// E2E Message Operations
r.Handle("/api/e2e/messages/send", jwtMiddleware(http.HandlerFunc(e2eServer.HandleSendE2EMessage))).Methods("POST")
r.Handle("/api/e2e/messages/{id}", jwtMiddleware(http.HandlerFunc(e2eServer.HandleGetE2EMessage))).Methods("GET")
r.Handle("/api/e2e/messages/{id}", jwtMiddleware(http.HandlerFunc(e2eServer.HandleDeleteE2EMessage))).Methods("DELETE")
r.Handle("/api/e2e/messages", jwtMiddleware(http.HandlerFunc(e2eServer.HandleListE2EMessages))).Methods("GET")

// Key Management
r.Handle("/api/e2e/keys/register", jwtMiddleware(http.HandlerFunc(e2eServer.HandleKeyRegistration))).Methods("POST")
r.Handle("/api/e2e/keys/verify", jwtMiddleware(http.HandlerFunc(e2eServer.HandleKeyVerification))).Methods("POST")
r.Handle("/api/e2e/keys/{user_id}", jwtMiddleware(http.HandlerFunc(e2eServer.HandleGetUserKeys))).Methods("GET")
r.Handle("/api/e2e/keys/{user_id}/rotate", jwtMiddleware(http.HandlerFunc(e2eServer.HandleKeyRotation))).Methods("POST")

// Key Transparency
r.Handle("/api/e2e/kt/log", jwtMiddleware(http.HandlerFunc(e2eServer.HandleKTLogAppend))).Methods("POST")
r.Handle("/api/e2e/kt/proof/{entry_id}", jwtMiddleware(http.HandlerFunc(e2eServer.HandleKTProofVerification))).Methods("GET")
r.Handle("/api/e2e/kt/audit", jwtMiddleware(http.HandlerFunc(e2eServer.HandleKTAudit))).Methods("GET")

// Threshold HSM
r.Handle("/api/e2e/hsm/threshold-sign", jwtMiddleware(http.HandlerFunc(e2eServer.HandleThresholdSign))).Methods("POST")
r.Handle("/api/e2e/hsm/threshold-verify", jwtMiddleware(http.HandlerFunc(e2eServer.HandleThresholdVerify))).Methods("POST")
r.Handle("/api/e2e/hsm/status", jwtMiddleware(http.HandlerFunc(e2eServer.HandleHSMStatus))).Methods("GET")

// Migration Management
r.Handle("/api/e2e/migration/status", jwtMiddleware(http.HandlerFunc(e2eServer.HandleMigrationStatus))).Methods("GET")
r.Handle("/api/e2e/migration/start", jwtMiddleware(http.HandlerFunc(e2eServer.HandleMigrationStart))).Methods("POST")
r.Handle("/api/e2e/migration/pause", jwtMiddleware(http.HandlerFunc(e2eServer.HandleMigrationPause))).Methods("POST")
r.Handle("/api/e2e/migration/resume", jwtMiddleware(http.HandlerFunc(e2eServer.HandleMigrationResume))).Methods("POST")
r.Handle("/api/e2e/migration/rollback", jwtMiddleware(http.HandlerFunc(e2eServer.HandleMigrationRollback))).Methods("POST")

// Feature Flag Management
r.Handle("/api/e2e/features/status", jwtMiddleware(http.HandlerFunc(e2eServer.HandleFeatureStatus))).Methods("GET")
r.Handle("/api/e2e/features/enable", jwtMiddleware(http.HandlerFunc(e2eServer.HandleFeatureEnable))).Methods("POST")
r.Handle("/api/e2e/features/disable", jwtMiddleware(http.HandlerFunc(e2eServer.HandleFeatureDisable))).Methods("POST")
```

#### 1.2 Dual-mode Message Handler

**Enhanced Message Handler:**
```go
func (srv *Server) enhancedEmailHandler(w http.ResponseWriter, r *http.Request) {
    // Check E2E mode from headers
    e2eMode := r.Header.Get("X-E2E-Mode")
    
    switch e2eMode {
    case "hybrid", "e2e":
        // Route to E2E handler
        srv.e2eServer.HandleSendE2EMessage(w, r)
    case "legacy", "":
        // Route to legacy handler
        srv.sendEmailHandler(w, r)
    default:
        http.Error(w, "Invalid E2E mode", http.StatusBadRequest)
    }
}
```

#### 1.3 Feature Flag Integration

**Feature Flag System:**
```go
type FeatureFlags struct {
    Global struct {
        E2EEnabled bool `json:"e2e_enabled"`
        KTEnabled  bool `json:"kt_enabled"`
        HSMEnabled bool `json:"hsm_enabled"`
    } `json:"global"`
    
    Organizations map[string]OrgFlags `json:"organizations"`
    Users         map[string]UserFlags `json:"users"`
}

type OrgFlags struct {
    E2EEnabled bool `json:"e2e_enabled"`
    KTEnabled  bool `json:"kt_enabled"`
    HSMEnabled bool `json:"hsm_enabled"`
    RolloutPercentage float64 `json:"rollout_percentage"`
}

type UserFlags struct {
    E2EEnabled bool `json:"e2e_enabled"`
    KTEnabled  bool `json:"kt_enabled"`
    HSMEnabled bool `json:"hsm_enabled"`
    OptInDate  time.Time `json:"opt_in_date"`
}
```

### 1.4 Database Integration

**Enhanced Database Schema:**
```sql
-- E2E Message Storage
CREATE TABLE IF NOT EXISTS e2e_messages (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    sequence_number INTEGER NOT NULL,
    sender_uuid TEXT NOT NULL,
    recipient_uuid TEXT NOT NULL,
    envelope_hash TEXT NOT NULL,
    envelope_version TEXT NOT NULL DEFAULT '1.0',
    key_rotation_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    routing_token TEXT,
    delivery_status TEXT DEFAULT 'pending',
    kem_algorithm TEXT NOT NULL DEFAULT 'kyber768',
    dem_algorithm TEXT NOT NULL DEFAULT 'aes256gcm',
    signature_algorithm TEXT NOT NULL DEFAULT 'dilithium3',
    correlation_id TEXT,
    trace_id TEXT,
    e2e_enabled BOOLEAN DEFAULT FALSE,
    migration_status TEXT DEFAULT 'legacy',
    UNIQUE(thread_id, sequence_number)
);

-- Migration Tracking
CREATE TABLE IF NOT EXISTS e2e_migrations (
    id TEXT PRIMARY KEY,
    migration_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    progress INTEGER DEFAULT 0,
    total_items INTEGER DEFAULT 0,
    processed_items INTEGER DEFAULT 0,
    failed_items INTEGER DEFAULT 0,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    rollback_available BOOLEAN DEFAULT FALSE,
    created_by TEXT NOT NULL
);

-- Feature Flags
CREATE TABLE IF NOT EXISTS e2e_feature_flags (
    id TEXT PRIMARY KEY,
    scope TEXT NOT NULL, -- 'global', 'organization', 'user'
    scope_id TEXT, -- organization_id or user_id
    feature_name TEXT NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    rollout_percentage FLOAT DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope, scope_id, feature_name)
);
```

## 🔄 **Component 2: Migration Worker**

### Design Goals
- Background migration without service interruption
- Progress tracking and monitoring
- Rollback capability
- Error handling and retry logic

### Implementation Strategy

#### 2.1 Migration Worker Service

**Worker Service Structure:**
```go
type MigrationWorker struct {
    db              *sql.DB
    jobQueue        chan MigrationJob
    progressTracker *ProgressTracker
    errorHandler    *ErrorHandler
    config          MigrationConfig
    mu              sync.RWMutex
    isRunning       bool
    currentJob      *MigrationJob
}

type MigrationJob struct {
    ID              string    `json:"id"`
    Type            string    `json:"type"` // 'legacy_to_e2e', 'key_rotation', 'metadata_migration'
    Status          string    `json:"status"` // 'pending', 'running', 'completed', 'failed', 'paused'
    Progress        int       `json:"progress"`
    TotalItems      int       `json:"total_items"`
    ProcessedItems  int       `json:"processed_items"`
    FailedItems     int       `json:"failed_items"`
    StartedAt       time.Time `json:"started_at"`
    CompletedAt     *time.Time `json:"completed_at,omitempty"`
    ErrorMessage    string    `json:"error_message,omitempty"`
    RollbackAvailable bool    `json:"rollback_available"`
    CreatedBy       string    `json:"created_by"`
    
    // Job-specific data
    BatchSize       int       `json:"batch_size"`
    RetryCount      int       `json:"retry_count"`
    MaxRetries      int       `json:"max_retries"`
    RetryDelay      time.Duration `json:"retry_delay"`
}
```

#### 2.2 Migration Types

**Legacy to E2E Migration:**
```go
func (w *MigrationWorker) migrateLegacyToE2E(job *MigrationJob) error {
    // 1. Scan legacy messages
    legacyMessages, err := w.scanLegacyMessages(job.BatchSize)
    if err != nil {
        return fmt.Errorf("failed to scan legacy messages: %w", err)
    }
    
    // 2. Process each message
    for _, msg := range legacyMessages {
        // Check if user has E2E enabled
        if !w.isE2EEnabledForUser(msg.SenderID) {
            continue // Skip if E2E not enabled
        }
        
        // 3. Re-encrypt with E2E
        e2eEnvelope, err := w.reencryptWithE2E(msg)
        if err != nil {
            w.recordFailedItem(job, msg.ID, err)
            continue
        }
        
        // 4. Store E2E message
        err = w.storeE2EMessage(e2eEnvelope)
        if err != nil {
            w.recordFailedItem(job, msg.ID, err)
            continue
        }
        
        // 5. Mark legacy message as migrated
        err = w.markLegacyMessageMigrated(msg.ID)
        if err != nil {
            w.recordFailedItem(job, msg.ID, err)
            continue
        }
        
        w.recordProcessedItem(job, msg.ID)
    }
    
    return nil
}
```

#### 2.3 Progress Tracking

**Progress Tracker:**
```go
type ProgressTracker struct {
    db *sql.DB
    mu sync.RWMutex
}

func (pt *ProgressTracker) UpdateProgress(jobID string, progress int, processed, failed int) error {
    pt.mu.Lock()
    defer pt.mu.Unlock()
    
    query := `
        UPDATE e2e_migrations 
        SET progress = ?, processed_items = ?, failed_items = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ?
    `
    
    _, err := pt.db.Exec(query, progress, processed, failed, jobID)
    return err
}

func (pt *ProgressTracker) GetProgress(jobID string) (*MigrationProgress, error) {
    pt.mu.RLock()
    defer pt.mu.RUnlock()
    
    query := `
        SELECT progress, total_items, processed_items, failed_items, status, error_message
        FROM e2e_migrations 
        WHERE id = ?
    `
    
    var progress MigrationProgress
    err := pt.db.QueryRow(query, jobID).Scan(
        &progress.Progress,
        &progress.TotalItems,
        &progress.ProcessedItems,
        &progress.FailedItems,
        &progress.Status,
        &progress.ErrorMessage,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &progress, nil
}
```

## 🧪 **Component 3: Dual-mode Testing**

### Design Goals
- Comprehensive testing of both legacy and E2E modes
- Backwards compatibility verification
- Performance comparison testing
- Migration testing

### Implementation Strategy

#### 3.1 Dual-mode Test Framework

**Test Framework Structure:**
```go
type DualModeTestSuite struct {
    server     *Server
    e2eServer  *E2EServer
    testDB     *sql.DB
    testClient *http.Client
}

func (ts *DualModeTestSuite) TestLegacyToE2EMigration(t *testing.T) {
    // 1. Create legacy message
    legacyMsg := ts.createLegacyMessage(t)
    
    // 2. Enable E2E for user
    ts.enableE2EForUser(t, legacyMsg.SenderID)
    
    // 3. Trigger migration
    migrationJob := ts.startMigration(t, "legacy_to_e2e")
    
    // 4. Wait for completion
    ts.waitForMigrationCompletion(t, migrationJob.ID)
    
    // 5. Verify E2E message exists
    e2eMsg := ts.getE2EMessage(t, legacyMsg.ID)
    assert.NotNil(t, e2eMsg)
    assert.Equal(t, "e2e", e2eMsg.MigrationStatus)
    
    // 6. Verify legacy message still accessible
    legacyMsgAfter := ts.getLegacyMessage(t, legacyMsg.ID)
    assert.NotNil(t, legacyMsgAfter)
    
    // 7. Test decryption
    decrypted := ts.decryptE2EMessage(t, e2eMsg)
    assert.Equal(t, legacyMsg.Content, decrypted)
}
```

#### 3.2 Performance Testing

**Performance Test Suite:**
```go
func (ts *DualModeTestSuite) TestPerformanceComparison(t *testing.T) {
    // Test legacy mode performance
    legacyMetrics := ts.benchmarkLegacyMode(t, 1000)
    
    // Test E2E mode performance
    e2eMetrics := ts.benchmarkE2EMode(t, 1000)
    
    // Compare metrics
    assert.Less(t, e2eMetrics.AverageLatency, legacyMetrics.AverageLatency*1.5) // E2E should not be more than 50% slower
    assert.Less(t, e2eMetrics.ErrorRate, 0.01) // Error rate should be less than 1%
    assert.Greater(t, e2eMetrics.Throughput, legacyMetrics.Throughput*0.8) // Throughput should be at least 80%
}

func (ts *DualModeTestSuite) benchmarkLegacyMode(t *testing.T, messageCount int) *PerformanceMetrics {
    start := time.Now()
    errors := 0
    
    for i := 0; i < messageCount; i++ {
        msg := ts.createTestLegacyMessage(t)
        err := ts.sendLegacyMessage(t, msg)
        if err != nil {
            errors++
        }
    }
    
    duration := time.Since(start)
    
    return &PerformanceMetrics{
        MessageCount:   messageCount,
        Duration:       duration,
        AverageLatency: duration / time.Duration(messageCount),
        Throughput:     float64(messageCount) / duration.Seconds(),
        ErrorRate:      float64(errors) / float64(messageCount),
    }
}
```

#### 3.3 Integration Testing

**End-to-End Integration Tests:**
```go
func (ts *DualModeTestSuite) TestEndToEndMessageFlow(t *testing.T) {
    // 1. Setup test users
    sender := ts.createTestUser(t, "sender@test.com")
    recipient := ts.createTestUser(t, "recipient@test.com")
    
    // 2. Enable E2E for both users
    ts.enableE2EForUser(t, sender.ID)
    ts.enableE2EForUser(t, recipient.ID)
    
    // 3. Send E2E message
    message := ts.createTestMessage(t, sender.ID, recipient.ID, "Test E2E message")
    e2eResponse := ts.sendE2EMessage(t, message)
    
    // 4. Verify message stored
    assert.NotEmpty(t, e2eResponse.MessageID)
    assert.Equal(t, "delivered", e2eResponse.Status)
    
    // 5. Retrieve and decrypt message
    retrievedMsg := ts.getE2EMessage(t, e2eResponse.MessageID)
    decrypted := ts.decryptE2EMessage(t, retrievedMsg)
    
    // 6. Verify content
    assert.Equal(t, message.Content, decrypted)
    
    // 7. Verify server cannot decrypt
    serverDecryptAttempt := ts.attemptServerDecryption(t, retrievedMsg)
    assert.Error(t, serverDecryptAttempt)
}
```

## 🚀 **Implementation Plan**

### Week 1: Server API Integration
**Days 1-3: Core Integration**
- [ ] Integrate E2E endpoints into main API server
- [ ] Implement dual-mode message handler
- [ ] Add feature flag system
- [ ] Database schema integration

**Days 4-5: Authentication & Security**
- [ ] JWT token validation for E2E operations
- [ ] User permission checks
- [ ] Rate limiting for E2E endpoints
- [ ] Audit logging integration

### Week 2: Migration Worker + Testing
**Days 6-8: Migration Worker**
- [ ] Implement background migration system
- [ ] Job queue management
- [ ] Progress tracking and monitoring
- [ ] Error handling and retry logic

**Days 9-10: Dual-mode Testing**
- [ ] Comprehensive dual-mode tests
- [ ] Performance benchmarking
- [ ] Integration testing
- [ ] Security validation

## 📊 **Success Metrics**

### Functionality Metrics
- [ ] All E2E endpoints respond correctly
- [ ] Dual-mode message handling works seamlessly
- [ ] Migration worker successfully processes legacy data
- [ ] Feature flags control functionality correctly

### Performance Metrics
- [ ] E2E operations: < 200ms additional latency
- [ ] Migration throughput: > 100 messages/second
- [ ] API response time: < 500ms for all endpoints
- [ ] Memory usage: < 100MB additional for E2E features

### Security Metrics
- [ ] Server cannot decrypt E2E messages
- [ ] All authentication checks pass
- [ ] Audit logging captures all operations
- [ ] No sensitive data in logs

## 🎯 **Next Steps**

Upon completion of Sprint 4, the system will be ready for:
1. **Sprint 5**: Performance optimization and security testing
2. **Sprint 6**: Production deployment and monitoring
3. **Sprint 7**: Final validation and go-live

---

**Status**: Ready for implementation  
**Priority**: High (production integration)  
**Dependencies**: Sprints 0-3 completed
