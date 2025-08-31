# Signup V2 Endpoint Documentation

## Overview

The `/api/signup` endpoint (v2) provides a privacy-compliant user registration system that supports Free, Paid, and Company account types. This endpoint follows strict privacy rules and implements secure password hashing with Argon2id.

## Endpoint Details

- **URL**: `POST /api/signup`
- **Content-Type**: `application/json`
- **Rate Limited**: Yes (same as existing signup endpoint)

## Request Format

```json
{
  "plan": "free" | "paid" | "company",
  "email": "string",
  "password": "string",
  "company_code": "optional string"
}
```

### Field Descriptions

- **plan** (required): Account type
  - `"free"`: Free tier account
  - `"paid"`: Paid subscription account
  - `"company"`: Enterprise/company account
- **email** (required): User's email address
- **password** (required): User's password (must meet security requirements)
- **company_code** (optional): Required only for company plans

## Response Format

### Success Response (201 Created)

```json
{
  "status": "success",
  "user_id": "generated-uuid",
  "next_step": "verify_email"
}
```

### Error Response (400 Bad Request)

```json
{
  "error": "Error description"
}
```

## Privacy Compliance

This endpoint strictly follows privacy requirements:

- **No PII Logging**: Email addresses and passwords are never logged
- **No Request Payload Logging**: Request bodies are not logged
- **No Analytics**: No tracking identifiers or analytics data
- **Secure Storage**: Passwords are immediately hashed with Argon2id
- **Parameterized Queries**: All database queries use parameterized statements

## Security Features

### Password Requirements

Passwords must meet the following criteria:
- Minimum 8 characters
- Maximum 128 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special character (`!@#$%^&*()_+-=[]{}|;:,.<>?`)

### Password Hashing

- **Algorithm**: Argon2id
- **Parameters**: 
  - Time cost: 3
  - Memory cost: 64MB
  - Parallelism: 4
  - Key length: 32 bytes
- **Salt**: 16-byte random salt generated per password

### User ID Generation

- **Format**: UUID v4
- **Uniqueness**: Guaranteed by UUID algorithm
- **Security**: Cryptographically secure random generation

## Database Schema

The endpoint stores the following data:

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,                    -- UUID v4
    email TEXT UNIQUE NOT NULL,             -- User's email
    password_hash TEXT NOT NULL,            -- Argon2id hash
    plan TEXT DEFAULT 'free',               -- Account plan
    company_code TEXT,                      -- Company code (if applicable)
    status TEXT DEFAULT 'pending_verification', -- Account status
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Validation Rules

### Plan Validation
- Must be one of: `"free"`, `"paid"`, `"company"`
- Company plans require a non-empty `company_code`

### Email Validation
- Must contain `@` and `.` characters
- Length between 5 and 254 characters
- Basic format validation

### Company Code Validation
- Required for company plans
- Optional for free and paid plans

## Error Handling

| Error Condition | HTTP Status | Error Message |
|----------------|-------------|---------------|
| Invalid JSON | 400 | "Invalid JSON format" |
| Missing fields | 400 | "Missing required fields" |
| Invalid plan | 400 | "Invalid plan type" |
| Invalid email | 400 | "Invalid email format" |
| Weak password | 400 | "Password does not meet security requirements" |
| Missing company code | 400 | "Company code required for company plans" |
| User already exists | 400 | "User already exists" |
| Database error | 500 | "Internal server error" |

## Usage Examples

### Free Plan Signup

```bash
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{
    "plan": "free",
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

### Paid Plan Signup

```bash
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{
    "plan": "paid",
    "email": "premium@example.com",
    "password": "SecurePass123!"
  }'
```

### Company Plan Signup

```bash
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{
    "plan": "company",
    "email": "employee@company.com",
    "password": "SecurePass123!",
    "company_code": "COMP123"
  }'
```

## Next Steps

After successful signup:

1. **Email Verification**: The `next_step` will be `"verify_email"`
2. **Account Activation**: User account remains in `"pending_verification"` status
3. **Future Endpoints**: 
   - `/api/verify-email` (planned for next iteration)
   - Billing integration for paid plans
   - Bulk provisioning for company plans

## Testing

Run the test suite to verify functionality:

```bash
cd cmd/api
go test -v -run TestSignupHandlerV2
```

## Migration

To apply the required database changes:

```bash
sqlite3 your_database.db < migrations/xxxx_add_signup_v2_support.sql
```

## Security Considerations

- All passwords are hashed immediately upon receipt
- No raw passwords are stored or transmitted
- Database queries are parameterized to prevent SQL injection
- Rate limiting prevents abuse
- UUIDs prevent user enumeration attacks
- No sensitive data is logged or exposed in error messages

