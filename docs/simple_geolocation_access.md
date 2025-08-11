# Simple Geolocation-Based Email Access Restrictions

## Overview

**Micro-Iteration 4.10** implements simple geolocation-based access restrictions for secure emails. This feature allows senders to restrict access to encrypted emails based on a specific city, with an optional country selection as an additional filter.

## Key Features

- **Single City Restriction**: Restrict access to a specific city only
- **Single Country Restriction**: Restrict access to a specific country only  
- **Combined Restrictions**: Require both city AND country to match
- **Case-Insensitive Matching**: Normalized comparison for both city and country
- **Generic Error Messages**: Security-focused responses that don't reveal specific restriction details
- **VPN Awareness**: Designed to work with VPN usage considerations

## Database Schema

### New Fields Added to `emails` Table

```sql
-- Single city name (case-insensitive, normalized)
ALTER TABLE emails ADD COLUMN allowed_city TEXT;

-- Single ISO 3166-1 alpha-2 country code (case-insensitive)
ALTER TABLE emails ADD COLUMN allowed_country TEXT;

-- Index for geolocation queries
CREATE INDEX IF NOT EXISTS idx_emails_simple_geolocation ON emails(allowed_city, allowed_country);
```

### Field Details

- **`allowed_city`**: Single city name (optional)
  - Case-insensitive matching
  - Normalized (lowercase, trimmed, single spaces)
  - Validates: 2-100 characters, letters, spaces, hyphens, apostrophes only
  
- **`allowed_country`**: Single ISO 3166-1 alpha-2 country code (optional)
  - Case-insensitive matching
  - Validates: Exactly 2 letters (a-z, A-Z)
  - Examples: "US", "CA", "GB", "DE", "FR"

## API Changes

### Send Email Endpoint (`POST /api/email/send`)

#### Request Body Changes

```json
{
  "recipient": "user@example.com",
  "subject": "Secure Email",
  "body": "Email content",
  "allowedCity": "New York",        // Optional: single city
  "allowedCountry": "US"            // Optional: single country code
}
```

#### Validation Rules

- **Country Code**: Must be valid ISO 3166-1 alpha-2 format (2 letters)
- **City Name**: Must be 2-100 characters, letters/spaces/hyphens/apostrophes only
- **Both Optional**: Can set city only, country only, both, or neither

### View Email Endpoint (`GET /api/email/view/{id}`)

#### Geolocation Enforcement

1. **IP Detection**: Extracts client IP from headers (X-Forwarded-For, X-Real-IP, CF-Connecting-IP)
2. **Geolocation Lookup**: Uses ip-api.com to resolve country and city
3. **Exact Matching**: 
   - Country: Case-insensitive exact match
   - City: Normalized case-insensitive exact match
4. **AND Logic**: If both restrictions set, both must match
5. **Generic Response**: Returns `403 Forbidden` with `{"error":"Access denied"}` on failure

#### Security Flow

```
1. Authentication Check
2. Geolocation Check (if restrictions set)
3. MFA Check (if enabled)
4. Password Check (if required)
5. Email Decryption
```

## Frontend Changes

### Compose Modal Updates

#### New UI Elements

```tsx
// Country Restriction Dropdown
<select value={allowedCountry} onChange={handleCountryChange}>
  <option value="">No country restriction</option>
  <option value="US">United States</option>
  <option value="CA">Canada</option>
  <option value="GB">United Kingdom</option>
  // ... more countries
</select>

// City Restriction Text Input
<input 
  type="text"
  placeholder="Enter city name (e.g., New York)"
  value={allowedCity}
  onChange={handleCityChange}
/>
```

#### Form Data Structure

```typescript
interface ComposeFormData {
  securitySettings: {
    // ... existing fields
    allowedCountry?: string;  // Single country code
    allowedCity?: string;     // Single city name
  }
}
```

## Geolocation Service

### IP Geolocation Provider

- **Service**: ip-api.com (free tier: 1,000 requests/day)
- **Endpoint**: `http://ip-api.com/json/{ip}?fields=status,message,countryCode,city,query`
- **Response**: JSON with country code and city name

### Normalization Functions

```go
// Country normalization
strings.ToLower(strings.TrimSpace(countryCode))

// City normalization  
func NormalizeCityName(city string) string {
    normalized := strings.ToLower(strings.TrimSpace(city))
    return strings.Join(strings.Fields(normalized), " ")
}
```

## Usage Examples

### Example 1: City-Only Restriction

```json
{
  "recipient": "colleague@company.com",
  "subject": "Office Meeting Details",
  "body": "Meeting details for tomorrow...",
  "allowedCity": "New York",
  "allowedCountry": ""
}
```

**Result**: Only users accessing from New York can view the email.

### Example 2: Country-Only Restriction

```json
{
  "recipient": "client@company.com", 
  "subject": "Confidential Report",
  "body": "Quarterly report...",
  "allowedCity": "",
  "allowedCountry": "US"
}
```

**Result**: Only users accessing from the United States can view the email.

### Example 3: Combined Restriction

```json
{
  "recipient": "partner@company.com",
  "subject": "Local Partnership Details", 
  "body": "Partnership terms...",
  "allowedCity": "San Francisco",
  "allowedCountry": "US"
}
```

**Result**: Only users accessing from San Francisco, US can view the email.

### Example 4: No Restrictions

```json
{
  "recipient": "general@company.com",
  "subject": "Public Announcement",
  "body": "Public information...",
  "allowedCity": "",
  "allowedCountry": ""
}
```

**Result**: No location restrictions applied.

## Security Considerations

### VPN and Proxy Usage

- **VPN Impact**: VPN usage may cause legitimate access failures
- **Proxy Detection**: System detects and uses real client IP when possible
- **Graceful Handling**: Generic error messages prevent information leakage

### Error Handling

- **Geolocation Failures**: Access denied when geolocation service unavailable
- **Generic Messages**: No specific restriction details revealed to users
- **Logging**: All blocked attempts logged with IP and location details

### Validation

- **Input Sanitization**: All inputs validated and normalized
- **SQL Injection**: Parameterized queries prevent injection attacks
- **XSS Prevention**: Output properly escaped in frontend

## Testing

### Unit Tests

```bash
go test ./pkg/geolocation
```

Tests cover:
- City name normalization
- Country code validation  
- City name validation
- Exact matching logic

### Integration Tests

```bash
./scripts/test_simple_geolocation.ps1
```

Tests cover:
- Email sending with various restriction combinations
- Invalid input validation
- Case sensitivity and normalization
- API endpoint functionality

### Manual Testing

To test actual geolocation enforcement:

1. **VPN Testing**: Use VPN services to simulate different locations
2. **IP Testing**: Access from different physical locations
3. **Edge Cases**: Test with various city/country combinations

## Migration

### Database Migration

The migration is automatically applied on server startup:

```sql
-- schema/migrate_add_simple_geolocation.sql
ALTER TABLE emails ADD COLUMN allowed_city TEXT;
ALTER TABLE emails ADD COLUMN allowed_country TEXT;
CREATE INDEX IF NOT EXISTS idx_emails_simple_geolocation ON emails(allowed_city, allowed_country);
```

### Backward Compatibility

- **Existing Emails**: No restrictions applied (fields are NULL)
- **API Compatibility**: New fields are optional
- **Frontend**: Gracefully handles missing fields

## Limitations

### Geolocation Accuracy

- **IP-Based**: Accuracy depends on IP geolocation database
- **VPN Impact**: VPN usage may cause false rejections
- **Mobile Networks**: Mobile IPs may not reflect actual location

### Service Dependencies

- **External API**: Depends on ip-api.com availability
- **Rate Limits**: 1,000 requests/day free tier
- **Network**: Requires internet connectivity

### User Experience

- **Generic Errors**: Users don't know specific restriction details
- **VPN Users**: May experience legitimate access failures
- **Mobile Users**: Location may not match expected city

## Future Enhancements

### Potential Improvements

1. **Multiple Geolocation Providers**: Fallback providers for reliability
2. **Geolocation Caching**: Cache results to reduce API calls
3. **More Granular Restrictions**: State/province level restrictions
4. **User Feedback**: Allow users to report incorrect restrictions
5. **Advanced Matching**: Fuzzy matching for city names

### Configuration Options

1. **Provider Selection**: Choose geolocation provider
2. **Cache Settings**: Configure caching behavior
3. **Fallback Behavior**: Define behavior when geolocation fails
4. **Logging Levels**: Configure detailed logging

## Troubleshooting

### Common Issues

1. **Access Denied Errors**
   - Check if VPN is being used
   - Verify geolocation service is available
   - Confirm restriction settings are correct

2. **Invalid Input Errors**
   - Ensure country codes are 2 letters (e.g., "US" not "USA")
   - Check city names don't contain special characters
   - Verify input lengths are within limits

3. **Geolocation Failures**
   - Check internet connectivity
   - Verify ip-api.com is accessible
   - Review rate limit usage

### Debug Information

Enable debug logging to see:
- Client IP addresses
- Geolocation lookup results
- Restriction matching details
- Access decision logic

## API Reference

### Send Email Request

```typescript
interface SendEmailRequest {
  recipient: string;
  subject: string;
  body: string;
  allowedCity?: string;     // Optional: single city name
  allowedCountry?: string;  // Optional: single country code
  // ... other fields
}
```

### View Email Response (Error)

```json
{
  "error": "Access denied"
}
```

### Geolocation Response

```json
{
  "country": "us",
  "city": "new york", 
  "ip": "192.168.1.1"
}
```

## Conclusion

The simple geolocation access restrictions provide a straightforward way to limit email access based on location while maintaining security and user privacy. The implementation is robust, secure, and designed to handle real-world scenarios including VPN usage and geolocation service failures.
