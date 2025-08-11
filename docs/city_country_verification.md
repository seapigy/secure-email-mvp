# Enhanced Geolocation Verification for Secure Email Access

## Overview

**Micro-Iteration 4.15** implements enhanced geolocation verification functionality for the Secure Email MVP. This feature allows senders to optionally require city and/or country verification for email access, providing more flexible and granular location-based security controls compared to the previous single city/country restriction system.

## Key Features

- **Four Verification Types**: `none`, `country`, `city`, `city_country`
- **Flexible Configuration**: Choose the level of geolocation verification required
- **Case-Insensitive Matching**: Normalized comparison for both city and country
- **Whitespace Handling**: Automatic normalization of input values
- **Multi-Layer Integration**: Works alongside existing security features
- **Generic Error Messages**: Security-focused responses that don't reveal system details
- **Brute-Force Protection**: Integrates with existing lockout mechanisms
- **Backward Compatibility**: Coexists with existing geolocation restrictions

## Database Schema

### New Fields in `emails` Table

```sql
-- Migration: migrate_add_city_country_verification.sql

-- Add geolocation verification type field
ALTER TABLE emails ADD COLUMN geo_verification_type TEXT CHECK (geo_verification_type IN ('none', 'country', 'city', 'city_country')) DEFAULT 'none';

-- Add geolocation verification city and country fields
ALTER TABLE emails ADD COLUMN geo_city TEXT;
ALTER TABLE emails ADD COLUMN geo_country TEXT;

-- Add index for geolocation verification queries
CREATE INDEX IF NOT EXISTS idx_emails_geo_verification ON emails(geo_verification_type, geo_city, geo_country);
```

### Field Details

- **`geo_verification_type`**: Type of geolocation verification required
  - `'none'`: No geolocation verification required (default)
  - `'country'`: Only country verification required
  - `'city'`: Only city verification required
  - `'city_country'`: Both city and country verification required
- **`geo_city`**: City name for verification (case-insensitive, normalized)
- **`geo_country`**: ISO 3166-1 alpha-2 country code for verification (case-insensitive)

## Backend Implementation

### Geolocation Verification Package

The `pkg/geoverify` package provides the core functionality:

```go
// Initialize geolocation verifier
geoVerifier := geoverify.NewGeolocationVerifier()

// Verify location access
result := geoVerifier.VerifyLocation(
    geoverify.VerificationType("city_country"),
    clientLocation,
    "new york",
    "us",
)

if !result.Allowed {
    // Handle access denied
    log.Printf("Access denied: %s", result.Reason)
}
```

### Verification Types

#### 1. No Verification (`none`)
- **Description**: No geolocation verification required
- **Required Fields**: None
- **Behavior**: Access allowed regardless of location
- **Use Case**: Standard emails without location restrictions

#### 2. Country Only (`country`)
- **Description**: Only country verification required
- **Required Fields**: `geoCountry`
- **Behavior**: Access allowed if client is in the specified country
- **Use Case**: Restrict access to a specific country regardless of city

#### 3. City Only (`city`)
- **Description**: Only city verification required
- **Required Fields**: `geoCity`
- **Behavior**: Access allowed if client is in the specified city
- **Use Case**: Restrict access to a specific city regardless of country

#### 4. City + Country (`city_country`)
- **Description**: Both city and country verification required
- **Required Fields**: `geoCity`, `geoCountry`
- **Behavior**: Access allowed if client is in the specified city AND country
- **Use Case**: Restrict access to a specific city in a specific country

### Integration Points

#### 1. Email Send Flow

Enhanced geolocation verification is integrated into the email send flow in `send_email_handler.go`:

```go
// Validate enhanced geolocation verification (Micro-Iteration 4.15)
geoVerifier := geoverify.NewGeolocationVerifier()

// Set default verification type if not provided
geoVerificationType := req.GeoVerificationType
if geoVerificationType == "" {
    geoVerificationType = "none"
}

// Validate verification type
if err := geoVerifier.ValidateVerificationType(geoVerificationType); err != nil {
    // Return error response
}

// Validate verification fields based on type
if err := geoVerifier.ValidateVerificationFields(
    geoverify.VerificationType(geoVerificationType),
    req.GeoCity,
    req.GeoCountry,
); err != nil {
    // Return error response
}

// Normalize verification fields
normalizedGeoCity, normalizedGeoCountry := geoVerifier.NormalizeVerificationFields(
    geoverify.VerificationType(geoVerificationType),
    req.GeoCity,
    req.GeoCountry,
)
```

#### 2. Email Access Flow

Enhanced geolocation verification is integrated into the email access flow in `view_email_handler.go`:

```go
// Security flow order:
// 1. Authentication Check
// 2. IP-Based Lockout Check (Micro-Iteration 4.13)
// 3. Geolocation Check (if restrictions set) - Legacy
// 4. Enhanced Geolocation Verification (if enabled) - NEW
// 5. Per-Email Brute-Force Lockout Check (Micro-Iteration 4.12)
// 6. Password Check (if password-protected)
// 7. MFA Check (if enabled)
// 8. Email Decryption
```

#### 3. Enhanced Verification Logic

```go
// Check enhanced geolocation verification (Micro-Iteration 4.15)
if geoVerificationType != "" && geoVerificationType != "none" {
    // Get geolocation service
    geoService := geolocation.NewGeolocationService()
    
    // Get location for the client IP
    location, err := geoService.GetLocationByIP(clientIP)
    if err != nil {
        // Return access denied
    }

    // Use the geolocation verifier to check access
    geoVerifier := geoverify.NewGeolocationVerifier()
    verificationResult := geoVerifier.VerifyLocation(
        geoverify.VerificationType(geoVerificationType),
        location,
        geoCity,
        geoCountry,
    )

    if !verificationResult.Allowed {
        // Increment brute-force and IP tracking attempts
        // Return access denied
    }
}
```

## API Changes

### Send Email Endpoint

**POST** `/api/email/send`

New optional fields in request body:

```json
{
  "recipient": "user@example.com",
  "subject": "Secure Email",
  "body": "Email content",
  "geoVerificationType": "city_country",  // NEW: "none", "city", "city_country"
  "geoCity": "New York",                  // NEW: City for verification
  "geoCountry": "US"                      // NEW: Country for verification
}
```

### Field Validation

#### Verification Type Validation
- **Valid Values**: `"none"`, `"city"`, `"city_country"`
- **Default**: `"none"` if not provided
- **Case Sensitivity**: Case-sensitive

#### City Validation (when required)
- **Required**: When `geoVerificationType` is `"city"` or `"city_country"`
- **Format**: 2-100 characters, letters, spaces, hyphens, apostrophes only
- **Normalization**: Converted to lowercase, trimmed, single spaces

#### Country Validation (when required)
- **Required**: When `geoVerificationType` is `"city_country"`
- **Format**: ISO 3166-1 alpha-2 country code (e.g., "US", "CA", "GB")
- **Normalization**: Converted to lowercase, trimmed

### Error Responses

#### Invalid Verification Type (400 Bad Request)
```json
{
  "error": "invalid verification type: invalid_type. Must be 'none', 'city', or 'city_country'"
}
```

#### Missing Required Fields (400 Bad Request)
```json
{
  "error": "city is required when verification type is 'city'"
}
```

#### Invalid City Name (400 Bad Request)
```json
{
  "error": "invalid city name: N. Must be 2-100 characters, letters, spaces, hyphens, and apostrophes only"
}
```

#### Invalid Country Code (400 Bad Request)
```json
{
  "error": "invalid country code: USA. Must be ISO 3166-1 alpha-2 format (e.g., US, CA, GB)"
}
```

#### Geolocation Verification Failure (403 Forbidden)
```json
{
  "error": "Access denied"
}
```

## Security Features

### 1. Generic Error Messages

All geolocation verification failures return generic "Access denied" messages to prevent information leakage about the specific verification requirements.

### 2. Brute-Force Integration

Geolocation verification failures trigger existing brute-force protection:
- **Per-Email Lockout**: Failed attempts tracked per email ID
- **IP-Based Lockout**: Failed attempts tracked per IP address
- **Automatic Reset**: Failed attempts reset on successful access
- **Generic Responses**: No indication of lockout status

### 3. Multi-Layer Security

Enhanced geolocation verification works alongside existing security features:
- **MFA**: Verification check before MFA validation
- **Password Protection**: Verification check after password validation
- **Brute-Force**: Verification failures increment brute-force counters
- **IP Tracking**: Verification failures increment IP tracking counters

### 4. Case-Insensitive Matching

Both city and country matching are case-insensitive and handle whitespace:
- **City**: "New York" matches "new york", "  NEW YORK  "
- **Country**: "US" matches "us", "  US  "

## Configuration

### Default Settings

- **Default Verification Type**: `"none"`
- **City Validation**: 2-100 characters, letters, spaces, hyphens, apostrophes
- **Country Validation**: ISO 3166-1 alpha-2 format
- **Case Sensitivity**: Case-insensitive matching
- **Whitespace Handling**: Automatic normalization

### Environment Variables

Currently using database defaults, but can be extended to support environment variable configuration:

```go
// Future enhancement: Configurable via environment variables
defaultVerificationType := os.Getenv("DEFAULT_GEO_VERIFICATION_TYPE")
maxCityLength := os.Getenv("MAX_CITY_LENGTH")
```

## Testing

### Unit Tests

```bash
go test ./pkg/geoverify -v
```

Tests cover:
- Verification type validation
- Field validation based on verification type
- Location verification logic
- Case-insensitive and whitespace handling
- Error handling and edge cases
- Normalization functions

### Integration Tests

```bash
./scripts/test_city_country_verification.ps1
```

Tests cover:
- Email sending with different verification types
- Field validation scenarios
- Invalid input rejection
- Access control verification
- Integration with existing security layers

## Usage Examples

### Example 1: City-Only Verification

```bash
curl -X POST http://localhost:8080/api/email/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "recipient@example.com",
    "subject": "City-Restricted Email",
    "body": "This email can only be accessed from New York.",
    "geoVerificationType": "city",
    "geoCity": "New York"
  }'
```

### Example 2: City + Country Verification

```bash
curl -X POST http://localhost:8080/api/email/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "recipient@example.com",
    "subject": "Location-Restricted Email",
    "body": "This email can only be accessed from Los Angeles, US.",
    "geoVerificationType": "city_country",
    "geoCity": "Los Angeles",
    "geoCountry": "US"
  }'
```

### Example 3: No Verification (Default)

```bash
curl -X POST http://localhost:8080/api/email/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "recipient@example.com",
    "subject": "Standard Email",
    "body": "This email has no location restrictions.",
    "geoVerificationType": "none"
  }'
```

## Security Considerations

### 1. Geolocation Accuracy

- **IP-Based**: Accuracy depends on geolocation database
- **VPN Impact**: VPN usage may cause false rejections
- **Mobile Networks**: Mobile IPs may not reflect actual location
- **Corporate Networks**: Corporate IPs may show different locations

### 2. Attack Prevention

- **Brute-Force Protection**: Failed attempts trigger lockouts
- **IP Tracking**: Prevents attacks from specific IP addresses
- **Generic Responses**: No information leakage about verification requirements
- **Rate Limiting**: Integrates with existing rate limiting

### 3. User Experience

- **Optional Feature**: Only applies when verification is enabled
- **Clear Validation**: Users get clear feedback on validation errors
- **Automatic Reset**: Failed attempts reset on success
- **No Impact on Normal Emails**: Non-verified emails work normally

### 4. Integration

- **Seamless Integration**: Works with all existing security features
- **No Conflicts**: No interference with MFA, password protection, etc.
- **Consistent Behavior**: Follows same patterns as other security layers
- **Comprehensive Coverage**: All security failures tracked consistently

## Monitoring and Logging

### Log Messages

The system logs various geolocation verification events:

```
Enhanced geolocation verification passed for IP 192.168.1.1 (new york, US)
Enhanced geolocation verification failed for IP 192.168.1.2 (chicago, US): City verification failed: expected New York, got Chicago
Failed to get geolocation for enhanced verification for IP 192.168.1.3: geolocation service error
```

### Database Monitoring

Monitor the following fields for security analysis:

```sql
-- Check emails with enhanced geolocation verification
SELECT email_id, geo_verification_type, geo_city, geo_country, created_at 
FROM emails 
WHERE geo_verification_type != 'none';

-- Check verification type distribution
SELECT geo_verification_type, COUNT(*) as count
FROM emails 
GROUP BY geo_verification_type;
```

## Troubleshooting

### Common Issues

1. **Verification Not Working**
   - Check verification type is correctly set
   - Verify city and country values are provided when required
   - Check for typos or case sensitivity issues

2. **Invalid Field Errors**
   - Ensure city names meet length and character requirements
   - Verify country codes are in ISO 3166-1 alpha-2 format
   - Check that required fields are provided for the verification type

3. **Access Denied After Verification**
   - Check if your location matches the verification requirements
   - Consider VPN usage or corporate network routing
   - Verify geolocation service is working correctly

4. **Migration Issues**
   - Check database migration was applied
   - Verify geolocation verification package is imported
   - Check logs for initialization errors

### Debug Information

Enable debug logging to see:
- Verification type validation
- Field validation results
- Geolocation lookup results
- Verification matching logic
- Error conditions
- Security event tracking

## Future Enhancements

### Potential Improvements

1. **Additional Verification Types**: Region, state, or postal code verification
2. **Time-Based Restrictions**: Allow access only during specific time zones
3. **Multiple Locations**: Support for multiple allowed cities/countries
4. **Fuzzy Matching**: Fuzzy matching for city names with similar spellings
5. **Geolocation Caching**: Cache results to reduce API calls

### Configuration Options

1. **Per-System Settings**: System-wide verification policy
2. **Per-User Settings**: User-specific verification requirements
3. **Verification Complexity**: Configurable complexity rules
4. **Geolocation Provider**: Configurable geolocation service provider

## Conclusion

The enhanced geolocation verification feature provides flexible and granular location-based security controls for secure email access. The implementation is secure, performant, and fully integrated with existing security features.

### Key Benefits

- ✅ Provides flexible geolocation verification options
- ✅ Uses case-insensitive and normalized matching
- ✅ Integrates seamlessly with existing security layers
- ✅ Prevents brute-force attacks through lockout mechanisms
- ✅ Maintains security without revealing system details
- ✅ Provides comprehensive monitoring and logging
- ✅ Automatically handles failed attempts and resets
- ✅ Maintains backward compatibility with existing features

The feature successfully meets all acceptance criteria and provides a valuable enhancement to the Secure Email MVP's security capabilities. The implementation is production-ready and provides a solid foundation for future security enhancements.
