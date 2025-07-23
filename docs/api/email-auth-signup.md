# Secure Email API – `/api/auth/signup` Endpoint Documentation

---

## Endpoint Overview

`POST /api/auth/signup`

Registers a new user account for the Secure Email API. Requires a valid email, a strong password, and a fallback email for account recovery.

---

## Request Example

```
POST /api/auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "StrongPassword123!",
  "fallback_email": "recovery@example.com"
}
```

---

## Response Examples

**Success (201 Created):**
```
{
  "message": "User created"
}
```

**Error (400 Bad Request):**
- Missing or invalid fields:
```
{
  "error": "Invalid email format"
}
{
  "error": "Password must be at least 8 characters long"
}
{
  "error": "Fallback email is required"
}
{
  "error": "Invalid fallback email format"
}
{
  "error": "User already exists"
}
```
- Invalid JSON:
```
{
  "error": "Invalid JSON format"
}
```
- Internal server error:
```
{
  "error": "Internal server error"
}
```

---

## Validation Rules
- **Email:** Must be a valid email address and unique.
- **Password:** Minimum 8 characters (enforced, can be strengthened further).
- **Fallback Email:** Required, must be a valid email address.

---

## Security Notes
- **Passwords are hashed** using bcrypt before storage; plaintext passwords are never stored.
- **Fallback email** is used for account recovery and must be valid and accessible by the user.
- **Duplicate emails** are not allowed; attempting to register with an existing email returns an error.
- **All validation and error responses** are returned as JSON for easy client handling.

---

## Example PowerShell Test Command

```
$body = @{
  email = "user@example.com"
  password = "StrongPassword123!"
  fallback_email = "recovery@example.com"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/auth/signup" -Method POST -ContentType "application/json" -Body $body
```

---

## Error Codes
| Status | Error Message Example                        | Description                        |
|--------|---------------------------------------------|------------------------------------|
| 201    | { "message": "User created" }              | Signup successful                  |
| 400    | { "error": "Invalid email format" }        | Invalid email                      |
| 400    | { "error": "Password must be at least 8 characters long" } | Weak password         |
| 400    | { "error": "Fallback email is required" }  | Missing fallback email             |
| 400    | { "error": "Invalid fallback email format" }| Invalid fallback email             |
| 400    | { "error": "User already exists" }         | Duplicate email                    |
| 400    | { "error": "Invalid JSON format" }         | Malformed JSON                     |
| 500    | { "error": "Internal server error" }       | Unexpected server/database error    |

---

## See Also
- [Email Send Endpoint Documentation](./email-send-endpoint.md)
- [Login Endpoint Documentation](./email-auth-login.md) *(to be created)*

---

*For questions or further improvements, contact the Secure Email API development team.* 