# Geolocation Restrictions Feature

## Overview

The Secure Email MVP now supports country and city-level geolocation restrictions, allowing senders to restrict access to their emails based on the recipient's physical location. This feature provides an additional layer of security by ensuring that sensitive emails can only be accessed from specific geographic locations.

## Features

### Country-Level Restrictions
- **ISO 3166-1 alpha-2 country codes**: Support for standard two-letter country codes (e.g., "US", "CA", "GB")
- **Multiple countries**: Allow access from multiple countries simultaneously
- **Case-insensitive**: Country codes are automatically normalized to lowercase

### City-Level Restrictions
- **Normalized city names**: City names are automatically normalized for comparison
- **Multiple cities**: Allow access from multiple cities simultaneously
- **Flexible naming**: Supports cities with spaces, hyphens, and apostrophes

### Combined Restrictions
- **AND logic**: When both country and city restrictions are set, the recipient must be in an allowed country AND an allowed city
- **Flexible configuration**: Can set only country restrictions, only city restrictions, or both

## Implementation Details

### Backend Architecture

#### Database Schema
```sql
-- Added to emails table
ALTER TABLE emails ADD COLUMN allowed_countries TEXT;  -- JSON array of country codes
ALTER TABLE emails ADD COLUMN allowed_cities TEXT;     -- JSON array of city names
```

#### Geolocation Service
- **Provider**: ipapi.co (free tier: 1,000 requests/day)
- **Data**: Country code and city name for IP addresses
- **Fallback**: Denies access if geolocation fails (security-first approach)

#### Enforcement Logic
1. **IP Resolution**: Extracts client IP from request headers (X-Forwarded-For, X-Real-IP, CF-Connecting-IP)
2. **Geolocation Lookup**: Queries ipapi.co for country and city information
3. **Restriction Check**: Validates location against allowed countries and cities
4. **Access Decision**: Allows or denies access based on location match

### Frontend Implementation

#### Compose Modal
- **Location Restrictions Toggle**: Enable/disable geolocation restrictions
- **Country Input**: Multi-select input for ISO country codes
- **City Input**: Multi-select input for city names
- **Validation**: Real-time validation of country codes and city names
- **User Guidance**: Clear instructions and examples

#### User Experience
- **Clear Feedback**: Shows restriction logic and requirements
- **Input Validation**: Prevents invalid country codes or city names
- **Flexible Input**: Supports comma-separated values for multiple locations

## API Endpoints

### Send Email with Geolocation Restrictions
```http
POST /api/email/send
Content-Type: application/json
Authorization: Bearer <jwt_token>

{
  "recipient": "recipient@example.com",
  "subject": "Restricted Email",
  "body": "This email has location restrictions.",
  "allowedCountries": ["us", "ca"],
  "allowedCities": ["new york", "toronto"]
}
```

### View Email (Geolocation Enforcement)
```http
GET /api/email/view/{email_id}
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
{
  "email_id": "uuid",
  "recipient": "recipient@example.com",
  "subject": "Restricted Email",
  "body": "This email has location restrictions.",
  "status": "success"
}
```

**Blocked Response (403 Forbidden):**
```json
{
  "error": "Access blocked: Your location (chicago, US) is not in the allowed countries or cities.",
  "code": "geo_restricted",
  "message": "Access blocked based on sender's location restrictions."
}
```

## Configuration

### Environment Variables
```bash
# No additional environment variables required
# Uses ipapi.co free tier by default
```

### Geolocation Provider
- **Service**: ipapi.co
- **Rate Limit**: 1,000 requests/day (free tier)
- **Data**: Country code, city name, IP address
- **Privacy**: GDPR-compliant, no personal data stored

## Security Considerations

### IP Address Handling
- **Proxy Support**: Handles X-Forwarded-For, X-Real-IP, CF-Connecting-IP headers
- **Fallback**: Uses RemoteAddr if no proxy headers are present
- **Validation**: Basic IP format validation

### Geolocation Accuracy
- **VPN Detection**: May block legitimate users using VPNs
- **Mobile Networks**: May show different locations for mobile users
- **Corporate Networks**: May show corporate location instead of user location

### Privacy and Compliance
- **No Storage**: Geolocation data is not stored permanently
- **Logging**: Access attempts are logged for security monitoring
- **GDPR**: Compliant with data protection regulations

## Usage Examples

### Country-Only Restrictions
```json
{
  "allowedCountries": ["us", "ca"],
  "allowedCities": []
}
```
*Allows access from United States and Canada only*

### City-Only Restrictions
```json
{
  "allowedCountries": [],
  "allowedCities": ["new york", "los angeles"]
}
```
*Allows access from New York and Los Angeles only*

### Combined Restrictions
```json
{
  "allowedCountries": ["us"],
  "allowedCities": ["new york", "chicago"]
}
```
*Allows access from New York or Chicago, but only if in the United States*

### No Restrictions
```json
{
  "allowedCountries": [],
  "allowedCities": []
}
```
*No location restrictions applied*

## Testing

### Test Script
Run the comprehensive test script:
```powershell
.\scripts\test_geolocation_restrictions.ps1 -ApiHost "http://localhost:8080"
```

### Test Cases
1. **Country-only restrictions**: Validates country code enforcement
2. **City-only restrictions**: Validates city name enforcement
3. **Combined restrictions**: Validates AND logic
4. **No restrictions**: Validates unrestricted access
5. **Invalid inputs**: Validates error handling
6. **Geolocation failures**: Validates fallback behavior

### Manual Testing
1. **Send restricted email**: Use compose modal to set location restrictions
2. **View from allowed location**: Should succeed
3. **View from blocked location**: Should return 403 with clear error message
4. **VPN testing**: Test with VPN to simulate different locations

## Error Handling

### Common Error Responses

#### Invalid Country Code (400 Bad Request)
```json
{
  "error": "Invalid country code. Must be ISO 3166-1 alpha-2 format (e.g., US, CA, GB)"
}
```

#### Invalid City Name (400 Bad Request)
```json
{
  "error": "Invalid city name. Must be 2-100 characters, letters, spaces, hyphens, and apostrophes only"
}
```

#### Geolocation Failure (403 Forbidden)
```json
{
  "error": "Access blocked: Unable to verify your location.",
  "code": "geo_restricted"
}
```

#### Location Blocked (403 Forbidden)
```json
{
  "error": "Access blocked: Your country (US) is not in the allowed countries.",
  "code": "geo_restricted",
  "message": "Access blocked based on sender's location restrictions."
}
```

## Monitoring and Logging

### Access Logs
All geolocation-related access attempts are logged:
```
2024-01-15 10:30:00 INFO: Geolocation check passed for IP 192.168.1.1 (new york, US)
2024-01-15 10:31:00 WARN: Geolocation access blocked for IP 192.168.1.2 (chicago, US): Access blocked: Your city (chicago) is not in the allowed cities.
```

### Metrics
- **Geolocation requests**: Number of IP lookups performed
- **Access blocks**: Number of location-based access denials
- **Geolocation failures**: Number of failed IP lookups

## Future Enhancements

### Planned Features
1. **MaxMind GeoLite2 Integration**: Local geolocation database for better performance
2. **Geolocation Caching**: Cache results to reduce API calls
3. **Advanced Location Types**: Support for regions, states, and postal codes
4. **Time-based Restrictions**: Allow access only during specific time zones
5. **Location History**: Track and display access location history

### Performance Optimizations
1. **Connection Pooling**: Optimize HTTP client for geolocation requests
2. **Batch Processing**: Handle multiple geolocation requests efficiently
3. **Rate Limiting**: Implement intelligent rate limiting for geolocation API

## Troubleshooting

### Common Issues

#### Geolocation Service Unavailable
- **Symptom**: All location-restricted emails are blocked
- **Cause**: ipapi.co service is down or rate limit exceeded
- **Solution**: Check service status and consider upgrading to paid tier

#### Incorrect Location Detection
- **Symptom**: Users in allowed locations are being blocked
- **Cause**: VPN usage or corporate network routing
- **Solution**: Educate users about VPN restrictions or adjust allowed locations

#### Performance Issues
- **Symptom**: Slow email access times
- **Cause**: Geolocation API latency
- **Solution**: Implement caching or consider local geolocation database

### Debug Commands
```bash
# Test geolocation service
curl "http://ip-api.com/json/8.8.8.8?fields=status,message,countryCode,city,query"

# Check database schema
sqlite3 /var/db/secure-email.db ".schema emails"

# View geolocation restrictions
sqlite3 /var/db/secure-email.db "SELECT email_id, allowed_countries, allowed_cities FROM emails WHERE allowed_countries IS NOT NULL OR allowed_cities IS NOT NULL;"
```

## Conclusion

The geolocation restrictions feature provides a powerful additional layer of security for sensitive emails. By allowing senders to restrict access based on physical location, the system ensures that confidential information remains accessible only to authorized users in specific geographic areas.

The implementation is designed to be secure, user-friendly, and compliant with privacy regulations. The feature integrates seamlessly with the existing security infrastructure and provides clear feedback to users when access is denied.
