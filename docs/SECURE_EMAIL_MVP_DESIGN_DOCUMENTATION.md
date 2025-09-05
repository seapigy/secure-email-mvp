# SecureMail MVP Design Documentation

## Overview
SecureMail is a quantum-resistant email platform that provides end-to-end encryption using hybrid cryptography combining AES-256-GCM with Post-Quantum Cryptography (PQC).

## Architecture

### Core Components
- **Backend**: Go API server with PostgreSQL database
- **Frontend**: React-based email client
- **Website**: Marketing landing page with encryption demo

### Security Features
- AES-256-GCM symmetric encryption
- PQC (Kyber-512) for key exchange
- Argon2id for key derivation
- Zero-knowledge architecture

## Technology Stack
- **Backend**: Go, PostgreSQL, Redis
- **Frontend**: React, TypeScript, Vite
- **Website**: React, Tailwind CSS, Framer Motion
- **Infrastructure**: Docker, Kubernetes, GitHub Actions

## Development Status
- ✅ Marketing website with encryption demo
- 🚧 Backend API development
- 🚧 Frontend client development
- 🚧 Infrastructure setup
