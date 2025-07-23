# Secure Email API – `/api/auth/login` Endpoint Documentation

---

## Endpoint Overview

`POST /api/auth/login`

Authenticates a user and returns a JWT token for session management. Requires a valid email and password. Fallback email must be confirmed for successful login.

---

## Request Example

```
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "StrongPassword123!"
}
```

---

## Response Examples

**Success (200 OK):**
```
{
  "token": "<JWT_TOKEN>",
  "message": "Login successful"
}
```

**Error (400 Bad Request):**
- Missing fields or invalid JSON:
```
{
  "error": "Email and password are required"
}
{
  "error": "Invalid JSON format"
}
```

**Error (401 Unauthorized):**
- Invalid credentials:
```
{
  "error": "Invalid email or password"
}
```

**Error (403 Forbidden):**
- Fallback email not confirmed:
```
{
  "error": "Fallback email not confirmed"
}
```

**Error (500 Internal Server Error):**
- Unexpected server/database error:
```
{
  "error": "Internal server error"
}
```

---

## Validation Rules
- **Email and password are required.**
- **Email must exist in the database.**
- **Password must match the stored hash.**
- **Fallback email must be confirmed.**

---

## Security Notes
- **Passwords are never stored or transmitted in plaintext.**
- **JWT tokens** are generated on successful login and should be used for authenticating protected endpoints.
- **Error messages** are generic for invalid credentials to prevent user enumeration.
- **All validation and error responses** are returned as JSON for easy client handling.

---

## Example PowerShell Test Command

```
$body = @{
  email = "user@example.com"
  password = "StrongPassword123!"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -ContentType "application/json" -Body $body
```

---

## Rate Limiting & Account Lockout

To protect against brute-force attacks, the login endpoint supports account lockout after repeated failed login attempts. This feature is controlled by the environment variable `LOGIN_RATE_LIMIT_ENABLED`.

- **How it works:**
  - After 5 failed login attempts within 15 minutes, the account is locked for 15 minutes.
  - During lockout, all login attempts return a 429 error with a clear message.
  - On successful login, failed attempt counters are reset.
- **Enable/Disable:**
  - Set `LOGIN_RATE_LIMIT_ENABLED=1` in your environment to enable lockout.
  - Unset or set to `0` to disable.
- **Error Response Example:**
  ```json
  {
    "error": "Account temporarily locked due to repeated failed login attempts. Please try again later."
  }
  ```
- **Error Code:**
  - `429 Too Many Requests` when account is locked.

---

## Error Codes (Updated)
| Status | Error Message Example                        | Description                        |
|--------|---------------------------------------------|------------------------------------|
| 200    | { "token": "<JWT_TOKEN>", "message": "Login successful" } | Login successful |
| 400    | { "error": "Email and password are required" } | Missing fields |
| 400    | { "error": "Invalid JSON format" }         | Malformed JSON                     |
| 401    | { "error": "Invalid email or password" }   | Invalid credentials                |
| 403    | { "error": "Fallback email not confirmed" }| Fallback not confirmed             |
| 429    | { "error": "Account temporarily locked due to repeated failed login attempts. Please try again later." } | Account locked |
| 500    | { "error": "Internal server error" }       | Unexpected server/database error    |

---

## Toggling Rate Limiting / Lockout

**Enable:**
```powershell
$env:LOGIN_RATE_LIMIT_ENABLED = '1'
```
**Disable:**
```powershell
Remove-Item Env:LOGIN_RATE_LIMIT_ENABLED
```
Restart the API server after changing this variable.

---

## See Also
- [Signup Endpoint Documentation](./email-auth-signup.md)
- [Email Send Endpoint Documentation](./email-send-endpoint.md)

---

*For questions or further improvements, contact the Secure Email API development team.* 