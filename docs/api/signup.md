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