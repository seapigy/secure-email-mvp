# Developer Guide

## Overview

This guide provides comprehensive information for developers working on the Secure Email application, including architecture, development setup, coding standards, and deployment procedures.

## Architecture Overview

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   Database      │
│   (React/TS)    │◄──►│   (Go)          │◄──►│   (SQLite)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CDN/Static    │    │   Cloudflare R2 │    │   File System   │
│   Assets        │    │   Storage       │    │   Logs          │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Technology Stack

#### Frontend
- **Framework**: React 18.2.0 with TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **State Management**: React Hooks (useState, useEffect, useCallback, useMemo)
- **HTTP Client**: Axios
- **UI Components**: Lucide React Icons
- **Notifications**: React Toastify

#### Backend
- **Language**: Go 1.23.0
- **Framework**: Standard library with custom HTTP handlers
- **Database**: SQLite with GORM
- **Encryption**: Post-Quantum Cryptography (PQC) with Kyber + AES-256-GCM
- **Authentication**: JWT with RS256 signing
- **File Storage**: Cloudflare R2

#### Infrastructure
- **Hosting**: Oracle Linux VM
- **SSL/TLS**: Let's Encrypt certificates
- **Email**: Amazon SES
- **Monitoring**: Custom logging and monitoring

## Development Setup

### Prerequisites

#### Required Software
- **Node.js**: 18.x or higher
- **Go**: 1.23.0 or higher
- **Git**: Latest version
- **SQLite**: 3.x
- **AWS CLI**: For SES configuration
- **Cloudflare CLI**: For R2 configuration

#### Development Tools
- **VS Code**: Recommended IDE
- **Postman**: API testing
- **SQLite Browser**: Database management
- **Docker**: Containerization (optional)

### Environment Setup

#### 1. Clone the Repository
```bash
git clone https://github.com/your-org/secure-email.git
cd secure-email
```

#### 2. Frontend Setup
```bash
# Install dependencies
npm install

# Create environment file
cp .env.example .env.local

# Start development server
npm run dev
```

#### 3. Backend Setup
```bash
# Install Go dependencies
go mod download

# Create environment file
cp env.example .env

# Run database migrations
go run cmd/migrate/main.go

# Start development server
go run cmd/api/main.go
```

#### 4. Environment Configuration

**Frontend (.env.local):**
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_NAME=Secure Email
VITE_APP_VERSION=1.0.0
```

**Backend (.env):**
```env
# Server Configuration
PORT=8080
ENV=development

# Database Configuration
DB_PATH=./data/secure_email.db

# JWT Configuration
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRY=24h

# Encryption Configuration
PQC_PRIVATE_KEY_PATH=./keys/pqc_private.key
PQC_PUBLIC_KEY_PATH=./keys/pqc_public.key

# Cloudflare R2 Configuration
R2_ACCOUNT_ID=your-r2-account-id
R2_ACCESS_KEY_ID=your-r2-access-key
R2_SECRET_ACCESS_KEY=your-r2-secret-key
R2_BUCKET_NAME=your-r2-bucket-name

# Amazon SES Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-ses-access-key
AWS_SECRET_ACCESS_KEY=your-ses-secret-key
SES_FROM_EMAIL=noreply@securesystem.email
```

### Database Setup

#### 1. Initialize Database
```bash
# Create database directory
mkdir -p data

# Run migrations
go run cmd/migrate/main.go
```

#### 2. Seed Data (Optional)
```bash
# Run seed script
go run cmd/seed/main.go
```

### Key Generation

#### 1. Generate PQC Keys
```bash
# Create keys directory
mkdir -p keys

# Generate PQC key pair
go run cmd/generate-keys/main.go
```

## Project Structure

### Frontend Structure
```
src/
├── components/           # React components
│   ├── common/          # Shared components
│   ├── secure/          # Secure email components
│   └── external/        # External access components
├── lib/                 # Utility libraries
│   ├── api.ts          # API client
│   └── utils.ts        # Utility functions
├── hooks/               # Custom React hooks
├── types/               # TypeScript type definitions
├── styles/              # CSS and styling
└── App.tsx             # Main application component
```

### Backend Structure
```
cmd/                     # Application entry points
├── api/                # Main API server
├── migrate/            # Database migrations
├── seed/               # Database seeding
└── generate-keys/      # Key generation
pkg/                     # Package libraries
├── auth/               # Authentication
├── encryption/         # Encryption utilities
├── pqc/                # Post-Quantum Cryptography
├── securelinks/        # Secure link management
├── database/           # Database operations
└── utils/              # Utility functions
schema/                  # Database schemas
migrations/              # Database migration files
```

## Coding Standards

### TypeScript/JavaScript

#### Code Style
- **Indentation**: 2 spaces
- **Quotes**: Single quotes for strings
- **Semicolons**: Required
- **Trailing Commas**: Required in objects and arrays
- **Line Length**: Maximum 100 characters

#### Naming Conventions
- **Variables**: camelCase
- **Constants**: UPPER_SNAKE_CASE
- **Functions**: camelCase
- **Components**: PascalCase
- **Files**: kebab-case
- **Types/Interfaces**: PascalCase

#### Component Structure
```typescript
/**
 * Component description
 * @param props - Component props
 * @returns JSX element
 */
interface ComponentProps {
  // Prop definitions
}

const Component: React.FC<ComponentProps> = ({ prop1, prop2 }) => {
  // Hooks
  const [state, setState] = useState<StateType>(initialState);
  
  // Event handlers
  const handleEvent = useCallback((event: EventType) => {
    // Event handling logic
  }, [dependencies]);
  
  // Effects
  useEffect(() => {
    // Effect logic
    return () => {
      // Cleanup logic
    };
  }, [dependencies]);
  
  // Render
  return (
    <div className="component">
      {/* Component content */}
    </div>
  );
};

export default Component;
```

### Go

#### Code Style
- **Indentation**: Tabs
- **Line Length**: No strict limit, but keep readable
- **Package Naming**: Lowercase, single word
- **File Naming**: snake_case.go

#### Naming Conventions
- **Variables**: camelCase
- **Constants**: camelCase
- **Functions**: camelCase
- **Types**: PascalCase
- **Packages**: lowercase
- **Files**: snake_case

#### Function Structure
```go
// FunctionName performs a specific operation
// Parameters:
//   - param1: description of parameter
//   - param2: description of parameter
// Returns:
//   - result: description of return value
//   - error: description of error
func FunctionName(param1 string, param2 int) (result string, err error) {
    // Input validation
    if param1 == "" {
        return "", errors.New("param1 cannot be empty")
    }
    
    // Main logic
    result = processData(param1, param2)
    
    // Return result
    return result, nil
}
```

### Security Standards

#### Input Validation
- **All Inputs**: Validate and sanitize all user inputs
- **Type Checking**: Use TypeScript for compile-time type safety
- **Length Limits**: Enforce reasonable length limits
- **Format Validation**: Validate email, URL, and other formats

#### Authentication
- **JWT Tokens**: Use RS256 signing for JWT tokens
- **Password Hashing**: Use Argon2id for password hashing
- **Session Management**: Implement secure session handling
- **Rate Limiting**: Implement rate limiting for authentication endpoints

#### Encryption
- **Data in Transit**: Use TLS 1.3 for all communications
- **Data at Rest**: Encrypt sensitive data at rest
- **Key Management**: Use secure key management practices
- **PQC Implementation**: Use Post-Quantum Cryptography for future-proofing

## API Development

### RESTful API Design

#### Endpoint Structure
```
GET    /api/health                    # Health check
POST   /api/auth/login               # User login
POST   /api/auth/logout              # User logout
GET    /api/auth/user                # Get current user
POST   /api/secure-email/send        # Send secure email
GET    /api/secure-email/link/{id}   # Get link status
POST   /api/attachments/upload       # Upload attachment
GET    /api/security/access-events   # Get access events
```

#### Request/Response Format
```go
// Request structure
type Request struct {
    Field1 string `json:"field1" validate:"required"`
    Field2 int    `json:"field2" validate:"min=1,max=100"`
}

// Response structure
type Response struct {
    Status  string      `json:"status"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message,omitempty"`
    Error   *Error      `json:"error,omitempty"`
}

// Error structure
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

#### Error Handling
```go
// Standard error response
func handleError(w http.ResponseWriter, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    response := Response{
        Status: "error",
        Error: &Error{
            Code:    code,
            Message: message,
        },
    }
    
    json.NewEncoder(w).Encode(response)
}
```

### Middleware

#### Authentication Middleware
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from header
        token := extractToken(r)
        if token == "" {
            handleError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing token")
            return
        }
        
        // Validate token
        claims, err := validateToken(token)
        if err != nil {
            handleError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
            return
        }
        
        // Add claims to context
        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### Rate Limiting Middleware
```go
func RateLimitMiddleware(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 100)
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            handleError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests")
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

## Testing

### Frontend Testing

#### Unit Tests
```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { vi } from 'vitest';
import Component from './Component';

describe('Component', () => {
  it('renders correctly', () => {
    render(<Component />);
    expect(screen.getByText('Expected Text')).toBeInTheDocument();
  });
  
  it('handles user interactions', async () => {
    const user = userEvent.setup();
    render(<Component />);
    
    const button = screen.getByRole('button');
    await user.click(button);
    
    expect(mockFunction).toHaveBeenCalled();
  });
});
```

#### Integration Tests
```typescript
describe('API Integration', () => {
  it('sends secure email successfully', async () => {
    const emailData = {
      recipient: 'test@example.com',
      subject: 'Test Subject',
      body: 'Test message'
    };
    
    const response = await sendSecureEmail(emailData);
    expect(response.status).toBe('success');
    expect(response.secure_link_url).toBeDefined();
  });
});
```

### Backend Testing

#### Unit Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "processed_test",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

#### Integration Tests
```go
func TestAPIEndpoint(t *testing.T) {
    // Setup test server
    server := httptest.NewServer(http.HandlerFunc(handler))
    defer server.Close()
    
    // Make request
    resp, err := http.Post(server.URL+"/api/endpoint", "application/json", strings.NewReader(`{"field":"value"}`))
    assert.NoError(t, err)
    defer resp.Body.Close()
    
    // Assert response
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var response Response
    err = json.NewDecoder(resp.Body).Decode(&response)
    assert.NoError(t, err)
    assert.Equal(t, "success", response.Status)
}
```

## Deployment

### Development Deployment

#### Local Development
```bash
# Start frontend
npm run dev

# Start backend
go run cmd/api/main.go

# Run tests
npm run test
go test ./...
```

#### Docker Development
```bash
# Build and run with Docker Compose
docker-compose up --build

# Run tests in container
docker-compose exec app npm run test
docker-compose exec api go test ./...
```

### Production Deployment

#### Build Process
```bash
# Frontend build
npm run build

# Backend build
go build -o bin/api cmd/api/main.go

# Create deployment package
tar -czf deployment.tar.gz dist/ bin/ data/ keys/
```

#### Server Deployment
```bash
# Upload deployment package
scp deployment.tar.gz user@server:/opt/secure-email/

# Extract and setup
ssh user@server
cd /opt/secure-email
tar -xzf deployment.tar.gz

# Setup systemd service
sudo cp config/secure-email.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable secure-email
sudo systemctl start secure-email
```

#### Environment Configuration
```bash
# Production environment
export ENV=production
export PORT=443
export DB_PATH=/opt/secure-email/data/secure_email.db
export JWT_SECRET=your-production-jwt-secret
export PQC_PRIVATE_KEY_PATH=/opt/secure-email/keys/pqc_private.key
export PQC_PUBLIC_KEY_PATH=/opt/secure-email/keys/pqc_public.key
```

### Monitoring & Logging

#### Application Logging
```go
// Structured logging
logger := log.New(os.Stdout, "", log.LstdFlags)

logger.Printf("User %s logged in from %s", userID, ipAddress)
logger.Printf("Secure email sent: %s", emailID)
logger.Printf("Security violation detected: %s", violationType)
```

#### Health Checks
```go
func healthCheck(w http.ResponseWriter, r *http.Request) {
    health := HealthCheck{
        Status:    "healthy",
        Timestamp: time.Now().UTC(),
        Version:   "1.0.0",
        Services: map[string]string{
            "database":   "healthy",
            "encryption": "healthy",
            "storage":    "healthy",
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

## Performance Optimization

### Frontend Optimization

#### Code Splitting
```typescript
// Lazy load components
const SecureEmailViewer = lazy(() => import('./SecureEmailViewer'));

// Route-based code splitting
<Route path="/v/:linkID" element={
  <Suspense fallback={<LoadingSpinner />}>
    <SecureEmailViewer />
  </Suspense>
} />
```

#### Memoization
```typescript
// Memoize expensive calculations
const expensiveValue = useMemo(() => {
  return performExpensiveCalculation(data);
}, [data]);

// Memoize event handlers
const handleClick = useCallback((event: MouseEvent) => {
  handleAction(event);
}, [handleAction]);
```

#### Bundle Optimization
```typescript
// Tree shaking
import { Send, Lock, Shield } from 'lucide-react';

// Dynamic imports
const HeavyComponent = lazy(() => import('./HeavyComponent'));
```

### Backend Optimization

#### Database Optimization
```go
// Use indexes for frequently queried fields
CREATE INDEX idx_emails_user_id ON emails(user_id);
CREATE INDEX idx_emails_created_at ON emails(created_at);

// Use prepared statements
stmt, err := db.Prepare("SELECT * FROM emails WHERE user_id = ?")
if err != nil {
    return err
}
defer stmt.Close()

rows, err := stmt.Query(userID)
```

#### Caching
```go
// In-memory caching
var cache = make(map[string]interface{})
var cacheMutex sync.RWMutex

func getCachedValue(key string) (interface{}, bool) {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()
    value, exists := cache[key]
    return value, exists
}

func setCachedValue(key string, value interface{}) {
    cacheMutex.Lock()
    defer cacheMutex.Unlock()
    cache[key] = value
}
```

#### Connection Pooling
```go
// Database connection pool
db, err := sql.Open("sqlite3", dbPath)
if err != nil {
    return err
}

db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Security Considerations

### Code Security

#### Input Validation
```go
// Validate and sanitize all inputs
func validateEmail(email string) error {
    if email == "" {
        return errors.New("email cannot be empty")
    }
    
    if len(email) > 254 {
        return errors.New("email too long")
    }
    
    if !emailRegex.MatchString(email) {
        return errors.New("invalid email format")
    }
    
    return nil
}
```

#### SQL Injection Prevention
```go
// Use parameterized queries
func getUserByID(db *sql.DB, userID string) (*User, error) {
    query := "SELECT id, email, name FROM users WHERE id = ?"
    row := db.QueryRow(query, userID)
    
    var user User
    err := row.Scan(&user.ID, &user.Email, &user.Name)
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

#### XSS Prevention
```typescript
// Sanitize user input
const sanitizeInput = (input: string): string => {
  return input
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
    .replace(/javascript:/gi, '')
    .replace(/on\w+\s*=/gi, '');
};

// Use React's built-in XSS protection
const userInput = "<script>alert('xss')</script>";
return <div>{userInput}</div>; // React automatically escapes this
```

### Infrastructure Security

#### Network Security
```bash
# Firewall configuration
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

#### SSL/TLS Configuration
```nginx
# Nginx SSL configuration
server {
    listen 443 ssl http2;
    server_name securesystem.email;
    
    ssl_certificate /etc/letsencrypt/live/securesystem.email/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/securesystem.email/privkey.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
}
```

## Contributing

### Development Workflow

#### 1. Fork and Clone
```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/your-username/secure-email.git
cd secure-email

# Add upstream remote
git remote add upstream https://github.com/original-org/secure-email.git
```

#### 2. Create Feature Branch
```bash
# Create and checkout feature branch
git checkout -b feature/your-feature-name

# Make your changes
# Test your changes
npm run test
go test ./...
```

#### 3. Commit Changes
```bash
# Add your changes
git add .

# Commit with descriptive message
git commit -m "feat: add new security feature

- Add fingerprint hash generation
- Implement tamper detection
- Update documentation

Closes #123"
```

#### 4. Push and Create Pull Request
```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
# Include description of changes
# Link related issues
```

### Code Review Process

#### Review Checklist
- [ ] Code follows style guidelines
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] Security considerations addressed
- [ ] Performance impact assessed
- [ ] Backward compatibility maintained

#### Review Comments
```typescript
// Good: Clear, actionable feedback
// Consider using useCallback here to prevent unnecessary re-renders
const handleClick = useCallback(() => {
  // handler logic
}, [dependencies]);

// Bad: Vague feedback
// This doesn't look right
```

### Release Process

#### Version Management
```bash
# Update version in package.json
npm version patch  # 1.0.0 -> 1.0.1
npm version minor  # 1.0.0 -> 1.1.0
npm version major  # 1.0.0 -> 2.0.0

# Update version in Go
# Update VERSION file
echo "1.0.1" > VERSION
```

#### Release Notes
```markdown
# Release v1.0.1

## Features
- Add fingerprint hash generation
- Implement tamper detection
- Enhance security monitoring

## Bug Fixes
- Fix XSS vulnerability in email viewer
- Resolve authentication token refresh issue
- Fix file upload size validation

## Security
- Update dependencies to latest versions
- Implement additional input validation
- Enhance encryption key management

## Breaking Changes
None

## Migration Guide
No migration required
```

## Resources

### Documentation
- [React Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Go Documentation](https://golang.org/doc/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)

### Security Resources
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [OWASP Testing Guide](https://owasp.org/www-project-web-security-testing-guide/)

### Development Tools
- [VS Code Extensions](https://marketplace.visualstudio.com/)
- [Postman](https://www.postman.com/)
- [Docker](https://www.docker.com/)
- [Git](https://git-scm.com/)

### Community
- [GitHub Issues](https://github.com/your-org/secure-email/issues)
- [Discord Server](https://discord.gg/secure-email)
- [Stack Overflow](https://stackoverflow.com/questions/tagged/secure-email)
- [Reddit](https://www.reddit.com/r/secureemail/)

### Contact
- **Development Team**: dev@securesystem.email
- **Security Team**: security@securesystem.email
- **Architecture Team**: arch@securesystem.email
- **DevOps Team**: devops@securesystem.email
