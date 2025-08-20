# Sprint 3: Hardware-backed Keys + Mixnet + Cover Traffic - Design Document

## 📋 **Executive Summary**

This document outlines the design and architecture for implementing hardware-backed key storage, mixnet traffic anonymization, and cover traffic generation for the Secure Email MVP's E2E PQC system.

**Status**: Design Phase  
**Sprint**: 3 (Advanced Security Features)  
**Timeline**: 2 weeks  
**Risk Level**: MEDIUM (optional features with fallback to software implementations)

## 🎯 **Sprint 3 Goals**

### Primary Objectives
1. **Hardware-backed Keys**: Secure key storage using platform-specific hardware security modules
2. **Mixnet Implementation**: Traffic anonymization through onion routing and mix networks
3. **Cover Traffic**: Dummy traffic generation to prevent traffic analysis
4. **Performance Optimization**: Ensure optional features don't impact core functionality

### Success Criteria
- Hardware key storage works on Windows (TPM), macOS (Secure Enclave), Linux (TPM/HSM)
- Mixnet provides traffic anonymization without breaking E2E encryption
- Cover traffic successfully obfuscates communication patterns
- All features are optional and gracefully degrade to software implementations
- Performance impact < 10% when features are enabled

## 🏗️ **Architecture Overview**

```
┌─────────────────────────────────────────────────────────────┐
│                    Sprint 3 Architecture                    │
├─────────────────────────────────────────────────────────────┤
│  Hardware Keys  │    Mixnet       │   Cover Traffic        │
│  ┌─────────────┐ │ ┌─────────────┐ │ ┌─────────────────────┐ │
│  │ TPM/Enclave │ │ │ Mix Nodes   │ │ │ Traffic Generator   │ │
│  │ Key Storage │ │ │ (Onion Rout)│ │ │ (Dummy Messages)    │ │
│  └─────────────┘ │ └─────────────┘ │ └─────────────────────┘ │
│  ┌─────────────┐ │ ┌─────────────┐ │ ┌─────────────────────┐ │
│  │ Fallback to │ │ │ Entry/Exit  │ │ │ Traffic Analysis    │ │
│  │ Software    │ │ │ Nodes       │ │ │ Protection          │ │
│  └─────────────┘ │ └─────────────┘ │ └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │   E2E Core       │
                    │ (Sprint 1 & 2)   │
                    └──────────────────┘
```

## 🔐 **Component 1: Hardware-backed Keys**

### Design Goals
- Secure key storage in hardware security modules
- Cross-platform compatibility (Windows TPM, macOS Secure Enclave, Linux TPM)
- Graceful fallback to software key storage
- Transparent integration with existing crypto operations

### Implementation Strategy

#### Hardware Key Manager Interface
```go
type HardwareKeyManager interface {
    // Key Generation
    GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error)
    
    // Key Operations
    Sign(keyRef *HardwareKeyRef, data []byte) ([]byte, error)
    Decrypt(keyRef *HardwareKeyRef, ciphertext []byte) ([]byte, error)
    
    // Key Management
    DeleteKey(keyRef *HardwareKeyRef) error
    ListKeys() ([]*HardwareKeyRef, error)
    
    // Platform Info
    IsAvailable() bool
    GetPlatformInfo() *PlatformInfo
}
```

#### Platform-specific Implementations

##### Windows TPM 2.0
```go
type WindowsTPMManager struct {
    tpmDevice *tpm2.TPMDevice
    config    TPMConfig
}

func (w *WindowsTPMManager) GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error) {
    // Use Windows TPM 2.0 API
    // Create non-extractable key in TPM
    // Return hardware reference
}
```

##### macOS Secure Enclave
```go
type MacOSEnclaveManager struct {
    keychainRef SecKeychainRef
    config      EnclaveConfig
}

func (m *MacOSEnclaveManager) GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error) {
    // Use Security Framework
    // Create key in Secure Enclave
    // Return keychain reference
}
```

##### Linux TPM/PKCS#11
```go
type LinuxHSMManager struct {
    pkcs11Module *pkcs11.Module
    config       HSMConfig
}

func (l *LinuxHSMManager) GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error) {
    // Use PKCS#11 interface
    // Support various HSM vendors
    // Return PKCS#11 handle
}
```

### Configuration
```go
type HardwareKeyConfig struct {
    Enabled          bool   `json:"enabled"`
    PreferredPlatform string `json:"preferred_platform"` // "tpm", "enclave", "pkcs11"
    FallbackToSoftware bool   `json:"fallback_to_software"`
    
    // Platform-specific configs
    TPMConfig     TPMConfig     `json:"tpm_config,omitempty"`
    EnclaveConfig EnclaveConfig `json:"enclave_config,omitempty"`
    HSMConfig     HSMConfig     `json:"hsm_config,omitempty"`
}
```

## 🌐 **Component 2: Mixnet Implementation**

### Design Goals
- Traffic anonymization through mix networks
- Onion routing for metadata protection
- Low-latency operation suitable for email
- Resistance to traffic analysis

### Architecture

#### Mix Network Structure
```
Client → Entry Node → Mix Node 1 → Mix Node 2 → Exit Node → Server
         (Encrypt)   (Re-encrypt)  (Re-encrypt)  (Decrypt)
```

#### Mix Node Implementation
```go
type MixNode struct {
    nodeID      string
    privateKey  *KeyPair
    nextHops    []*MixNodeInfo
    batchConfig BatchConfig
    
    // Message batching and mixing
    messageQueue chan *MixMessage
    batchTimer   *time.Timer
}

type MixMessage struct {
    MessageID   string            `json:"message_id"`
    LayerCount  int              `json:"layer_count"`
    EncryptedPayload []byte       `json:"encrypted_payload"`
    NextHop     *MixNodeInfo     `json:"next_hop,omitempty"`
    Timestamp   time.Time        `json:"timestamp"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}
```

#### Onion Routing Protocol
```go
type OnionRouter struct {
    mixNetwork  *MixNetwork
    pathFinder  *PathFinder
    encryption  *OnionEncryption
}

func (o *OnionRouter) CreateOnionPacket(
    message []byte, 
    recipientID string,
    path []*MixNodeInfo,
) (*OnionPacket, error) {
    // Layer encryption from exit to entry
    packet := &OnionPacket{Payload: message}
    
    for i := len(path) - 1; i >= 0; i-- {
        packet = o.encryption.AddLayer(packet, path[i])
    }
    
    return packet, nil
}
```

#### Mix Network Configuration
```go
type MixNetConfig struct {
    Enabled      bool              `json:"enabled"`
    MinPath      int               `json:"min_path_length"`  // Minimum 3 hops
    MaxPath      int               `json:"max_path_length"`  // Maximum 7 hops
    BatchSize    int               `json:"batch_size"`       // Messages per batch
    BatchDelay   time.Duration     `json:"batch_delay"`      // Max batch wait time
    Nodes        []*MixNodeInfo    `json:"mix_nodes"`
    
    // Directory service for node discovery
    DirectoryURL string            `json:"directory_url"`
    RefreshInterval time.Duration  `json:"refresh_interval"`
}
```

### Mix Network Directory Service
```go
type MixDirectory struct {
    nodes       map[string]*MixNodeInfo
    reputation  map[string]*NodeReputation
    lastUpdate  time.Time
}

type MixNodeInfo struct {
    NodeID      string    `json:"node_id"`
    PublicKey   []byte    `json:"public_key"`
    Address     string    `json:"address"`
    Bandwidth   int64     `json:"bandwidth"`
    Uptime      float64   `json:"uptime"`
    Version     string    `json:"version"`
    Capabilities []string `json:"capabilities"`
}
```

## 📡 **Component 3: Cover Traffic**

### Design Goals
- Generate dummy traffic to obfuscate real communications
- Adaptive traffic patterns based on user behavior
- Minimal bandwidth overhead
- Indistinguishable from real traffic

### Traffic Pattern Analysis
```go
type TrafficAnalyzer struct {
    patterns    map[string]*TrafficPattern
    baseline    *BaselineTraffic
    adaptive    *AdaptiveGenerator
}

type TrafficPattern struct {
    UserID          string        `json:"user_id"`
    AverageMessages int           `json:"avg_messages_per_hour"`
    PeakHours       []int         `json:"peak_hours"`
    MessageSizes    []int         `json:"typical_sizes"`
    Intervals       []time.Duration `json:"intervals"`
    LastUpdate      time.Time     `json:"last_update"`
}
```

### Cover Traffic Generator
```go
type CoverTrafficGenerator struct {
    enabled     bool
    intensity   float64  // 0.0 to 1.0
    analyzer    *TrafficAnalyzer
    scheduler   *TrafficScheduler
    encryptor   *CoverMessageEncryptor
}

func (c *CoverTrafficGenerator) GenerateCoverMessage(
    userID string,
    size int,
) (*CoverMessage, error) {
    // Create dummy message that looks real
    dummyContent := c.generateDummyContent(size)
    
    // Encrypt with same parameters as real messages
    envelope, err := c.encryptor.EncryptCover(dummyContent, userID)
    if err != nil {
        return nil, err
    }
    
    return &CoverMessage{
        MessageID: generateCoverMessageID(),
        UserID:    userID,
        Envelope:  envelope,
        Size:      size,
        Timestamp: time.Now(),
        IsCover:   true,  // Internal flag, not transmitted
    }, nil
}
```

### Adaptive Traffic Scheduling
```go
type TrafficScheduler struct {
    patterns     map[string]*TrafficPattern
    nextSchedule map[string]time.Time
    jitter       time.Duration
}

func (t *TrafficScheduler) ScheduleNextCover(userID string) time.Time {
    pattern := t.patterns[userID]
    if pattern == nil {
        return t.getDefaultSchedule()
    }
    
    // Calculate next send time based on pattern
    interval := t.calculateAdaptiveInterval(pattern)
    jitteredInterval := t.addJitter(interval)
    
    return time.Now().Add(jitteredInterval)
}
```

### Cover Traffic Configuration
```go
type CoverTrafficConfig struct {
    Enabled         bool          `json:"enabled"`
    Intensity       float64       `json:"intensity"`        // 0.0 = disabled, 1.0 = max
    MinInterval     time.Duration `json:"min_interval"`     // Minimum between messages
    MaxInterval     time.Duration `json:"max_interval"`     // Maximum between messages
    SizeVariation   int           `json:"size_variation"`   // Message size randomness
    AnalysisWindow  time.Duration `json:"analysis_window"`  // Traffic pattern analysis period
    
    // Adaptive behavior
    AdaptToUser     bool          `json:"adapt_to_user"`    // Learn user patterns
    GlobalBaseline  bool          `json:"global_baseline"`  // Use global traffic patterns
}
```

## 🔧 **Integration Architecture**

### Enhanced E2E Client
```go
type EnhancedE2EClient struct {
    *Client  // Embed existing client
    
    // Sprint 3 components
    hardwareKeys    HardwareKeyManager
    mixnetRouter    *OnionRouter
    coverTraffic    *CoverTrafficGenerator
    
    // Configuration
    sprint3Config   Sprint3Config
}

type Sprint3Config struct {
    HardwareKeys HardwareKeyConfig `json:"hardware_keys"`
    Mixnet       MixNetConfig      `json:"mixnet"`
    CoverTraffic CoverTrafficConfig `json:"cover_traffic"`
}
```

### Message Flow with Sprint 3 Features
```go
func (e *EnhancedE2EClient) SendMessage(
    plaintext []byte,
    recipientID string,
    options *SendOptions,
) error {
    // 1. Hardware-backed signing (if available)
    signingKey := e.getSigningKey(options.UseHardware)
    
    // 2. Standard E2E encryption
    message, err := e.EncryptMessage(plaintext, recipientID, signingKey)
    if err != nil {
        return err
    }
    
    // 3. Mixnet routing (if enabled)
    if e.sprint3Config.Mixnet.Enabled {
        return e.sendViaMixnet(message, recipientID)
    }
    
    // 4. Direct send with cover traffic
    if e.sprint3Config.CoverTraffic.Enabled {
        e.coverTraffic.ScheduleCoverTraffic(e.userID)
    }
    
    return e.sendDirect(message)
}
```

## 🧪 **Testing Strategy**

### Hardware Key Testing
```go
func TestHardwareKeyManager_CrossPlatform(t *testing.T) {
    platforms := []string{"windows_tpm", "macos_enclave", "linux_pkcs11"}
    
    for _, platform := range platforms {
        t.Run(platform, func(t *testing.T) {
            if !isPlatformAvailable(platform) {
                t.Skip("Platform not available")
            }
            
            manager := createHardwareManager(platform)
            testKeyOperations(t, manager)
        })
    }
}
```

### Mixnet Testing
```go
func TestMixNetwork_TrafficAnalysisResistance(t *testing.T) {
    // Create test mix network
    network := setupTestMixNetwork(t, 5) // 5 mix nodes
    
    // Send messages through different paths
    messages := generateTestMessages(100)
    paths := network.GenerateRandomPaths(messages)
    
    // Verify traffic patterns are mixed
    analyzer := &TrafficAnalyzer{}
    patterns := analyzer.AnalyzeTraffic(paths)
    
    // Assert that original patterns are obfuscated
    assert.True(t, patterns.IsObfuscated())
}
```

### Cover Traffic Testing
```go
func TestCoverTraffic_PatternObfuscation(t *testing.T) {
    generator := NewCoverTrafficGenerator(testConfig)
    
    // Simulate user sending pattern
    realPattern := simulateUserTraffic(t, "user123", 1*time.Hour)
    
    // Generate cover traffic
    coverPattern := generator.GenerateCoverTraffic("user123", 1*time.Hour)
    
    // Analyze combined pattern
    combined := combinePatterns(realPattern, coverPattern)
    
    // Verify pattern is obfuscated
    assert.True(t, isPatternObfuscated(realPattern, combined))
}
```

## 🚀 **Implementation Plan**

### Week 1: Hardware-backed Keys + Mixnet Foundation
**Days 1-3: Hardware Key Implementation**
- [ ] Design hardware key manager interface
- [ ] Implement Windows TPM integration
- [ ] Implement macOS Secure Enclave integration
- [ ] Implement Linux PKCS#11 integration
- [ ] Add fallback mechanisms

**Days 4-5: Mixnet Core**
- [ ] Design mix node architecture
- [ ] Implement onion routing protocol
- [ ] Create mix network directory service
- [ ] Basic message batching and mixing

### Week 2: Cover Traffic + Integration + Testing
**Days 6-8: Cover Traffic System**
- [ ] Implement traffic pattern analysis
- [ ] Create adaptive cover traffic generator
- [ ] Add traffic scheduling algorithms
- [ ] Integrate with existing message flow

**Days 9-10: Integration & Testing**
- [ ] Integrate all Sprint 3 components
- [ ] Comprehensive testing suite
- [ ] Performance benchmarking
- [ ] Security analysis and validation

## 📊 **Performance Considerations**

### Hardware Keys
- **Latency**: +2-5ms per crypto operation
- **Availability**: Graceful degradation to software
- **Compatibility**: Cross-platform testing required

### Mixnet
- **Latency**: +100-500ms per message (acceptable for email)
- **Bandwidth**: +20-50% overhead for onion routing
- **Reliability**: Multiple path redundancy

### Cover Traffic
- **Bandwidth**: +5-15% based on intensity setting
- **Pattern Analysis**: Minimal CPU overhead
- **Storage**: Traffic patterns require ~1KB per user

## 🔒 **Security Analysis**

### Threat Model
1. **Traffic Analysis**: Mixnet + cover traffic provide protection
2. **Key Extraction**: Hardware backing prevents key extraction
3. **Timing Attacks**: Batching and jitter provide mitigation
4. **Infrastructure Compromise**: Multiple mix nodes reduce risk

### Security Properties
- **Unlinkability**: Messages cannot be linked to senders
- **Unobservability**: Communication patterns are hidden
- **Forward Secrecy**: Hardware keys support key rotation
- **Availability**: Graceful degradation maintains functionality

## 📋 **Configuration Management**

### Environment Variables
```bash
# Hardware Keys
HARDWARE_KEYS_ENABLED=true
HARDWARE_KEYS_PLATFORM=auto  # auto, tpm, enclave, pkcs11
HARDWARE_KEYS_FALLBACK=true

# Mixnet
MIXNET_ENABLED=true
MIXNET_MIN_PATH_LENGTH=3
MIXNET_MAX_PATH_LENGTH=5
MIXNET_BATCH_SIZE=10
MIXNET_BATCH_DELAY=2s

# Cover Traffic
COVER_TRAFFIC_ENABLED=true
COVER_TRAFFIC_INTENSITY=0.3  # 30% of normal traffic
COVER_TRAFFIC_ADAPTIVE=true
```

### Feature Flags
```go
type Sprint3FeatureFlags struct {
    HardwareKeysEnabled    bool `json:"hardware_keys_enabled"`
    MixnetEnabled         bool `json:"mixnet_enabled"`
    CoverTrafficEnabled   bool `json:"cover_traffic_enabled"`
    
    // Gradual rollout
    HardwareKeysRollout   float64 `json:"hardware_keys_rollout"`   // 0.0-1.0
    MixnetRollout        float64 `json:"mixnet_rollout"`          // 0.0-1.0
    CoverTrafficRollout  float64 `json:"cover_traffic_rollout"`   // 0.0-1.0
}
```

## 🎯 **Success Metrics**

### Functionality Metrics
- [ ] Hardware key operations work on all target platforms
- [ ] Mixnet successfully anonymizes traffic flows
- [ ] Cover traffic obfuscates communication patterns
- [ ] Graceful degradation works when features are unavailable

### Performance Metrics
- [ ] Hardware key operations: < 10ms additional latency
- [ ] Mixnet routing: < 1s total message latency
- [ ] Cover traffic: < 20% bandwidth overhead
- [ ] Memory usage: < 50MB additional for all features

### Security Metrics
- [ ] Traffic analysis resistance verified through simulation
- [ ] Key extraction resistance verified on hardware platforms
- [ ] Pattern obfuscation effectiveness measured
- [ ] No reduction in core E2E security properties

## 🚨 **Risk Mitigation**

### Technical Risks
1. **Hardware Availability**: Fallback to software implementations
2. **Platform Compatibility**: Extensive cross-platform testing
3. **Performance Impact**: Optional features with kill switches
4. **Mixnet Reliability**: Multiple redundant paths

### Operational Risks
1. **Complexity**: Comprehensive monitoring and alerting
2. **Debugging**: Enhanced logging for troubleshooting
3. **Rollback**: Feature flags for instant disable
4. **Dependencies**: Minimal external dependencies

## 📚 **Documentation Requirements**

### Technical Documentation
- [ ] Hardware key setup guides per platform
- [ ] Mixnet configuration and monitoring
- [ ] Cover traffic tuning recommendations
- [ ] Troubleshooting guides

### User Documentation
- [ ] Privacy enhancement explanations
- [ ] Performance impact disclosure
- [ ] Optional feature descriptions
- [ ] Fallback behavior documentation

---

**Next Steps**: Upon approval, begin implementation of hardware key managers and mixnet foundation components.
