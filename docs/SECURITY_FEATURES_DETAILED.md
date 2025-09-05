# Security Features Detailed

## Hybrid Encryption System

### AES-256-GCM
- **Purpose**: Symmetric encryption for message content
- **Key Size**: 256 bits
- **Mode**: Galois/Counter Mode (GCM) for authenticated encryption
- **Benefits**: Fast, secure, widely supported

### Post-Quantum Cryptography (PQC)
- **Algorithm**: Kyber-512
- **Purpose**: Quantum-resistant key exchange
- **Key Sizes**: 800 bytes (public), 1632 bytes (private)
- **Benefits**: Future-proof against quantum attacks

### Key Derivation
- **Function**: Argon2id
- **Parameters**: 
  - Memory: 64MB
  - Iterations: 3
  - Parallelism: 1
- **Benefits**: Memory-hard, resistant to side-channel attacks

## Zero-Knowledge Architecture
- Server never sees plaintext messages
- All encryption/decryption happens client-side
- Server only stores encrypted data and metadata

## Security Guarantees
- **Confidentiality**: Messages encrypted with AES-256-GCM
- **Integrity**: GCM provides authentication
- **Forward Secrecy**: New keys for each session
- **Quantum Resistance**: PQC protects against future quantum attacks
