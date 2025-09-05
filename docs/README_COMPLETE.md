# SecureMail Complete Documentation

## Project Structure
```
securemail/
├── backend/           # Go API + DB
├── frontend/          # React app (SecureMail client)
├── website/           # Marketing landing page
├── docs/              # Project documentation
├── infra/             # Deployment, CI/CD, Dockerfiles
└── README.md          # Root README
```

## Quick Start

### Website (Marketing)
```bash
cd website
npm install
npm run dev
```

### Frontend (Client)
```bash
cd frontend
npm install
npm run dev
```

### Backend (API)
```bash
cd backend
go mod tidy
go run cmd/main.go
```

## Development Workflow
1. **Feature Development**: Create feature branches
2. **Testing**: Run tests before committing
3. **Code Review**: All changes require review
4. **Deployment**: Automated via CI/CD

## Security
- All communications encrypted with TLS
- Client-side encryption with hybrid crypto
- Zero-knowledge architecture
- Regular security audits

## Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License
Proprietary - All rights reserved
