# Login API Documentation

## POST /api/auth/login

Authenticates a user with email, password, and TOTP code. Returns a JWT token for subsequent API calls.

### Endpoint
```
POST https://api.securesystem.email/api/auth/login
```

### Request Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "email": "user@securesystem.email",
  "password": "string",
  "totp_code": "123456"
}
```

#### Field Descriptions
- **email** (required): User's email address. Must be in the format `user@securesystem.email`
- **password** (required): User's password. Must be 8-128 characters long
- **totp_code** (required): 6-digit TOTP code from authenticator app

### Responses

#### 200 OK - Successful Authentication
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzNDU2Nzg5MCIsImVtYWlsIjoidXNlckBzZWN1cmVzeXN0ZW0uZW1haWwiLCJleHAiOjE2MzQ1Njc4OTAsImlhdCI6MTYzNDU2Nzg5MH0.signature",
  "user_id": "12345678-1234-1234-1234-123456789012"
}
```

#### 400 Bad Request - Invalid Input
```json
{
  "error": "Invalid request"
}
```

**Causes:**
- Missing required fields (email, password, totp_code)
- Invalid JSON format
- Email not in correct domain format
- Password length outside 8-128 character range
- TOTP code not 6 digits

#### 401 Unauthorized - Authentication Failed
```json
{
  "error": "Invalid credentials"
}
```

**Causes:**
- User not found
- Incorrect password
- Invalid TOTP code
- Expired TOTP code

#### 429 Too Many Requests - Rate Limited
```json
{
  "error": "Too many requests"
}
```

**Causes:**
- More than 10 requests per minute from the same IP address

### Security Features

#### TLS 1.3
- All requests must use HTTPS
- TLS 1.3 enforced by Cloudflare
- Certificate validation required

#### Rate Limiting
- 10 requests per minute per IP address
- In-memory tracking with automatic reset
- Applies to all authentication attempts

#### Secure Headers
The API returns the following security headers:
```
Content-Security-Policy: default-src 'self'
Strict-Transport-Security: max-age=31536000
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
```

#### Password Security
- Passwords hashed with Argon2 (time=1, memory=64 MiB, threads=4)
- Email used as salt for additional security
- Minimum 8 characters, maximum 128 characters

#### TOTP Authentication
- 6-digit codes with 30-second window
- Base32-encoded secrets stored securely
- Compatible with Google Authenticator, Authy, etc.

#### JWT Tokens
- Signed with HS256 algorithm
- 24-hour expiration
- Contains user_id and email claims
- Must be included in Authorization header for protected endpoints

### Error Logging
Failed authentication attempts are logged to `/var/log/api.log` with:
- Timestamp
- Email address (for tracking)
- IP address
- Error type (no sensitive data)

### Example Usage

#### cURL
```bash
curl -X POST https://api.securesystem.email/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@securesystem.email",
    "password": "securepassword123",
    "totp_code": "123456"
  }'
```

#### JavaScript (Fetch)
```javascript
const response = await fetch('https://api.securesystem.email/api/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@securesystem.email',
    password: 'securepassword123',
    totp_code: '123456'
  })
});

const data = await response.json();
if (response.ok) {
  const { token, user_id } = data;
  // Store token for subsequent API calls
  localStorage.setItem('auth_token', token);
} else {
  console.error('Login failed:', data.error);
}
```

#### Python (requests)
```python
import requests

response = requests.post('https://api.securesystem.email/api/auth/login', json={
    'email': 'user@securesystem.email',
    'password': 'securepassword123',
    'totp_code': '123456'
})

if response.status_code == 200:
    data = response.json()
    token = data['token']
    user_id = data['user_id']
    print(f"Login successful for user {user_id}")
else:
    print(f"Login failed: {response.json()['error']}")
```

### Testing
Use the provided test script to verify the API:
```bash
chmod +x tests/login_test.sh
./tests/login_test.sh
```

### Next Steps
After successful authentication:
1. Store the JWT token securely
2. Include token in Authorization header for protected endpoints
3. Handle token expiration (24 hours)
4. Implement logout by discarding token

### Related Endpoints
- `POST /api/auth/register` - User registration (future)
- `POST /api/auth/refresh` - Token refresh (future)
- `POST /api/auth/logout` - Token invalidation (future) 