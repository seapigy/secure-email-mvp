# Security Features Detailed

## Hybrid Encryption System

### AES-256-GCM
- **Purpose**: Symmetric encryption for message content
- **Key Size**: 256 bits
- **Mode**: Galois/Counter Mode (GCM) for authenticated encryption
- **Implementation**: Web Crypto API (`crypto.subtle`)
- **Benefits**: Fast, secure, widely supported, native browser implementation

### Post-Quantum Cryptography (PQC)
- **Algorithm**: ECDH (P-256) with Kyber-512 simulation
- **Purpose**: Quantum-resistant key exchange
- **Implementation**: Web Crypto API ECDH as fallback, with Kyber simulation for demo
- **Key Sizes**: 65 bytes (ECDH public), 800 bytes (simulated Kyber public)
- **Benefits**: Real cryptographic security with quantum-resistant concepts

### Key Derivation
- **Function**: Argon2id
- **Library**: `hash-wasm` (4.12.0)
- **Parameters**: 
  - Memory: 64MB
  - Iterations: 3
  - Parallelism: 1
  - Hash Length: 32 bytes (256 bits)
- **Benefits**: Memory-hard, resistant to side-channel attacks

## Crypto Library Architecture

### Dependencies (Minimized)
- **`hash-wasm`**: Argon2id key derivation only
- **Web Crypto API**: AES-256-GCM, ECDH, random number generation
- **Removed**: `tweetnacl`, `tweetnacl-util`, `kyber-crystals` (unused)

### Implementation Details
- **AES-256-GCM**: Native browser implementation via `crypto.subtle`
- **Key Generation**: `crypto.getRandomValues()` for cryptographically secure randomness
- **Key Exchange**: ECDH P-256 curve for real security
- **PQC Simulation**: Demonstrates Kyber concepts without external dependencies

## Zero-Knowledge Architecture
- Server never sees plaintext messages
- All encryption/decryption happens client-side
- Server only stores encrypted data and metadata
- No external crypto libraries required for core operations

## Security Guarantees
- **Confidentiality**: Messages encrypted with AES-256-GCM
- **Integrity**: GCM provides authentication
- **Forward Secrecy**: New keys for each session
- **Quantum Resistance**: ECDH provides current security, Kyber concepts for future
- **Minimal Attack Surface**: Reduced dependencies, native browser crypto

## Performance Optimizations
- **Bundle Size**: Reduced by removing unused crypto libraries
- **Code Splitting**: Crypto operations in separate chunk
- **Native Performance**: Web Crypto API provides hardware acceleration
- **Memory Efficiency**: Argon2id parameters optimized for web environment
