# /api/auth/signup
**POST** Initiate user sign-up with TOTP setup.

## Input
```json
{
  "email": "user@securesystem.email",
  "password": "string",
  "confirm_password": "string"
}
```

## Output
**200**: `{ "temp_id": "uuid", "totp_qr": "base64_png" }`

**400**: `{ "error": "Invalid email format" | "Passwords do not match" | "Email already exists" }`

**403**: `{ "error": "Max 100 users reached" }`

**500**: `{ "error": "Internal server error" }`

## Notes
- Email must end with `@securesystem.email`
- Password: 8–128 characters
- Temp ID expires in 5 minutes 

## Fallback Email Requirement & Confirmation

- Users must provide a fallback email during signup. After registration, a confirmation link is sent to the fallback email. The user must confirm this email before being able to log in. The confirmation link expires after 1 hour. If expired or lost, the user can request a new confirmation email via the `/resend-fallback` endpoint.
- This process ensures secure account recovery and prevents unauthorized access.

**Security:** Fallback tokens are HMAC-based, time-limited, and validated on the backend. All actions are logged for audit purposes. 