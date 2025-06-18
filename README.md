# Secure Email System MVP

A web-based secure email system with end-to-end encryption, built with Go and React.

## Features

- End-to-end encryption (AES-256-GCM)
- Secure link-based delivery
- Password and geolocation-based access verification
- Modern, responsive UI with dark/light modes
- TOTP authentication
- Folder-based organization

## Tech Stack

- Backend: Go 1.23
- Frontend: React 18
- Database: SQLite
- Storage: Cloudflare R2
- Hosting: Oracle Cloud Free Tier, Netlify

## Prerequisites

- Node.js v20
- Go v1.23
- Git

## Development Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd secure-email-mvp
   ```

2. Install dependencies:
   ```bash
   # Backend
   go mod init secure-email-mvp
   go mod tidy

   # Frontend
   cd src
   npm install
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run the development servers:
   ```bash
   # Backend
   go run cmd/api/main.go

   # Frontend
   cd src
   npm run dev
   ```

## Project Structure

```
.
├── cmd/
│   └── api/          # Backend entry point
├── pkg/              # Backend packages
├── src/              # Frontend source
│   ├── components/   # React components
│   ├── styles/       # CSS and Tailwind
│   └── lib/          # Frontend utilities
├── docs/             # Documentation
└── tests/            # Test files
```

## License

MIT License 