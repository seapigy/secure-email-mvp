# SecureMail

**Quantum-Resistant Email Platform with Zero-Knowledge Architecture**

SecureMail provides end-to-end encrypted email communication using hybrid cryptography that combines AES-256-GCM with Post-Quantum Cryptography (PQC) to ensure your messages remain secure against both current and future threats.

## 🏗️ Project Structure

```
securemail/
├── backend/           # Go API + Database
│   ├── cmd/          # Application entry points
│   ├── internal/     # Private application code
│   ├── migrations/   # Database migrations
│   └── go.mod        # Go module definition
│
├── frontend/          # React Email Client
│   ├── src/          # React application source
│   ├── public/       # Static assets
│   ├── package.json  # Node.js dependencies
│   └── vite.config.ts # Vite configuration
│
├── website/           # Marketing Landing Page
│   ├── src/          # React website source
│   ├── public/       # Static assets
│   ├── package.json  # Node.js dependencies
│   └── vite.config.ts # Vite configuration
│
├── docs/              # Project Documentation
│   ├── SECURE_EMAIL_MVP_DESIGN_DOCUMENTATION.md
│   ├── SECURITY_FEATURES_DETAILED.md
│   ├── DEPLOYMENT_INFRASTRUCTURE.md
│   └── README_COMPLETE.md
│
├── infra/             # Deployment & Infrastructure
│   ├── docker-compose.yml
│   ├── k8s/          # Kubernetes manifests
│   └── github-actions/ # CI/CD workflows
│
└── README.md          # This file
```

## 🚀 Quick Start

### Marketing Website
```bash
cd website
npm install
npm run dev
# Visit http://localhost:3000
```

### Email Client (Frontend)
```bash
cd frontend
npm install
npm run dev
# Visit http://localhost:3000
```

### API Server (Backend)
```bash
cd backend
go mod tidy
go run cmd/main.go
# API available at http://localhost:8080
```

### Full Stack with Docker
```bash
cd infra
docker-compose up
# Website: http://localhost:3001
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
```

## 🔐 Security Features

- **Hybrid Encryption**: AES-256-GCM + PQC (Kyber-512)
- **Key Derivation**: Argon2id with memory-hard parameters
- **Zero-Knowledge**: Server never sees plaintext
- **Quantum-Resistant**: Future-proof against quantum attacks
- **Forward Secrecy**: New keys for each session

## 🛠️ Technology Stack

- **Backend**: Go, PostgreSQL, Redis
- **Frontend**: React, TypeScript, Vite
- **Website**: React, Tailwind CSS, Framer Motion
- **Infrastructure**: Docker, Kubernetes, GitHub Actions

## 📚 Documentation

- [MVP Design Documentation](docs/SECURE_EMAIL_MVP_DESIGN_DOCUMENTATION.md)
- [Security Features](docs/SECURITY_FEATURES_DETAILED.md)
- [Deployment Infrastructure](docs/DEPLOYMENT_INFRASTRUCTURE.md)
- [Complete Documentation](docs/README_COMPLETE.md)

## 🧪 Development

### Prerequisites
- Node.js 18+
- Go 1.21+
- Docker (optional)
- PostgreSQL (for backend)

### Development Workflow
1. Create feature branch from `develop`
2. Make changes and add tests
3. Run tests: `npm test` / `go test`
4. Submit pull request
5. Code review and merge

### Testing
```bash
# Website tests
cd website && npm test

# Frontend tests  
cd frontend && npm test

# Backend tests
cd backend && go test ./...
```

## 🚢 Deployment

### Local Development
```bash
docker-compose -f infra/docker-compose.yml up
```

### Production
- Kubernetes manifests in `infra/k8s/`
- CI/CD pipeline in `infra/github-actions/`
- Automated testing and deployment

## 📄 License

Proprietary - All rights reserved

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

---

**SecureMail** - Your messages, your privacy, your control.

🚀 **Live Demo**: [View the encryption demo](https://your-netlify-site.netlify.app)