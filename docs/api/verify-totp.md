# /api/auth/verify-totp
**POST** Verify TOTP code to finalize user creation.

## Input
```json
{
  "temp_id": "uuid",
  "totp_code": "123456"
}
```

## Output
**200**: `{ "token": "jwt" }`

**400**: `{ "error": "Invalid TOTP code" | "Invalid or expired temp ID" }`

**500**: `{ "error": "Internal server error" }`

## Notes
- Creates user in SQLite after TOTP validation
- JWT valid for 24 hours 