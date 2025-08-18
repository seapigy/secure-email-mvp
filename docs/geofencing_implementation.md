# Geofencing & Location-Based Access Control

## Overview

Micro-Iteration 4.13 implements comprehensive geofencing functionality for the Secure Email MVP, allowing email senders to restrict access to specific geographic regions or IP ranges. This feature provides an additional layer of security by preventing unauthorized access from restricted locations.

## Features

### Core Functionality

- **Country-based restrictions**: Restrict access to specific countries using ISO 3166-1 alpha-2 country codes
- **IP range restrictions**: Restrict access to specific CIDR ranges (e.g., corporate networks)
- **Combined restrictions**: Support for both country and IP range restrictions simultaneously
- **Violation tracking**: Monitor and track geofence violation attempts
- **Self-destruct integration**: Failed geofence checks increment self-destruct counter
- **Generic error handling**: Prevents information leakage through standardized error messages

### Security Features

- **Defense in depth**: Geofencing works alongside existing security features (MFA, read-once, etc.)
- **Audit logging**: All geofence violations are logged with detailed metadata
- **Violation counter**: Track failed attempts for potential self-destruct triggers
- **Location privacy**: Client location data is not exposed in error messages

## Architecture

### Database Schema

```sql
-- Geofencing fields added to emails table
ALTER TABLE emails ADD COLUMN allowed_countries TEXT;
ALTER TABLE emails ADD COLUMN allowed_ip_ranges TEXT;
ALTER TABLE emails ADD COLUMN geofence_violations INTEGER DEFAULT 0;
ALTER TABLE emails ADD COLUMN geofence_last_violation DATETIME;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_emails_geofence_countries ON emails(allowed_countries);
CREATE INDEX IF NOT EXISTS idx_emails_geofence_ip_ranges ON emails(allowed_ip_ranges);
CREATE INDEX IF NOT EXISTS idx_emails_geofence_violations ON emails(geofence_violations, geofence_last_violation);
```

### Service Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP Handler  │───▶│ GeofencingService│───▶│ GeolocationSvc  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Database      │
                       │   (SQLite)      │
                       └─────────────────┘
```

### Components

1. **GeofencingService** (`pkg/geofencing/geofencing.go`)
   - Core geofencing logic and database operations
   - Violation tracking and counter management
   - Settings validation and formatting

2. **GeolocationService** (`pkg/geolocation/geolocation.go`)
   - IP-to-location mapping interface
   - Country and IP range validation
   - Mock implementation for testing

3. **HTTP Handler Integration** (`cmd/api/get_email_by_id_handler.go`)
   - Geofencing check in security flow
   - Error handling and audit logging
   - Integration with existing security features

## Implementation Details

### Security Check Order

The geofencing check is integrated into the existing security flow:

1. JWT authentication & sender/recipient check
2. Remote revoke, time lock, expiration checks
3. MFA-on-open & Decoy message validation
4. **Geofence check (NEW)**
5. Read-once consumption & secure retrieval

### Geofencing Logic

```go
func (g *GeofencingService) CheckGeofenceAccess(emailID, clientIP string) (*GeofenceResult, error) {
    // 1. Get geofencing settings from database
    allowedCountries, allowedIPRanges, err := g.getGeofencingSettings(emailID)
    
    // 2. If no restrictions, allow access
    if len(allowedCountries) == 0 && len(allowedIPRanges) == 0 {
        return &GeofenceResult{Allowed: true}, nil
    }
    
    // 3. Check IP range restrictions first (more specific)
    if len(allowedIPRanges) > 0 {
        if !g.geolocationSvc.IsIPInRange(clientIP, allowedIPRanges) {
            g.incrementGeofenceViolations(emailID)
            return &GeofenceResult{Allowed: false, Reason: "geofence_ip_blocked"}, nil
        }
    }
    
    // 4. Check country restrictions
    if len(allowedCountries) > 0 {
        location, err := g.geolocationSvc.GetLocation(clientIP)
        if err != nil {
            return &GeofenceResult{Allowed: false, Reason: "geofence_location_unknown"}, nil
        }
        
        if !g.geolocationSvc.IsCountryAllowed(location.Country, allowedCountries) {
            g.incrementGeofenceViolations(emailID)
            return &GeofenceResult{Allowed: false, Reason: "geofence_country_blocked"}, nil
        }
    }
    
    return &GeofenceResult{Allowed: true}, nil
}
```

### Data Storage

Geofencing settings are stored as JSON arrays in the database:

```json
// Allowed countries
["US", "CA", "GB"]

// Allowed IP ranges
["192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"]
```

### Error Handling

All geofencing failures return the generic error message:
```
"Email has been revoked or cannot be accessed"
```

This prevents information leakage about:
- Whether the email exists
- What geofencing restrictions are in place
- The client's actual location

## Usage Examples

### Setting Geofencing Restrictions

```go
// Allow only US and Canada
allowedCountries := []string{"US", "CA"}
allowedIPRanges := []string{}

err := geofencingSvc.SetGeofencingSettings(emailID, allowedCountries, allowedIPRanges)
```

### Corporate Network Restriction

```go
// Allow only corporate IP ranges
allowedCountries := []string{}
allowedIPRanges := []string{"192.168.1.0/24", "10.0.0.0/8"}

err := geofencingSvc.SetGeofencingSettings(emailID, allowedCountries, allowedIPRanges)
```

### Combined Restrictions

```go
// Allow US users from specific IP ranges
allowedCountries := []string{"US"}
allowedIPRanges := []string{"192.168.1.0/24", "203.0.113.0/24"}

err := geofencingSvc.SetGeofencingSettings(emailID, allowedCountries, allowedIPRanges)
```

## Testing

### Unit Tests

- **GeofencingService**: Core geofencing logic and database operations
- **GeolocationService**: IP-to-location mapping and validation
- **Validation**: Country codes and CIDR range validation

### Integration Tests

- **Complete flow**: End-to-end geofencing with HTTP handlers
- **Violation tracking**: Counter increment and reset functionality
- **Error handling**: Generic error message verification

### Mock Services

```go
// Mock geolocation service for testing
mockGeo := geolocation.NewMockGeolocationService()
mockGeo.SetLocation("192.168.1.1", &geolocation.Location{
    Country: "US",
    City:    "New York",
    IP:      "192.168.1.1",
})

// Mock geofencing service
geofencingSvc := geofencing.NewGeofencingService(db, mockGeo)
```

## Configuration

### Environment Variables

No additional environment variables are required for basic geofencing functionality. The system uses a simple geolocation service by default.

### Geolocation Provider

The current implementation uses a simple IP-to-country mapping. For production use, consider integrating with:

- **MaxMind GeoLite2**: Free geolocation database
- **IP-API**: Free tier with 1,000 requests/day
- **Cloudflare IP Geolocation**: Built-in with Cloudflare

### Custom Geolocation Service

To use a different geolocation provider, implement the `GeolocationService` interface:

```go
type GeolocationService interface {
    GetLocation(ip string) (*Location, error)
    IsCountryAllowed(clientCountry string, allowedCountries []string) bool
    IsIPInRange(clientIP string, allowedRanges []string) bool
}
```

## Security Considerations

### Privacy Protection

- Client location data is never exposed in error messages
- Geolocation lookups are performed server-side only
- Audit logs contain location data for security analysis but are not exposed to clients

### Attack Prevention

- **IP spoofing**: Relies on proper proxy configuration (X-Forwarded-For headers)
- **Geolocation bypass**: Uses multiple validation layers
- **Information leakage**: Generic error messages prevent enumeration attacks

### Performance

- Geofencing checks are performed after authentication but before content retrieval
- IP range checks are performed before geolocation lookups (faster)
- Database queries are optimized with proper indexing

## Monitoring and Analytics

### Audit Logging

All geofencing events are logged with:

```json
{
  "event_type": "email_access",
  "outcome": "failure",
  "details": "Geofencing blocked: geofence_country_blocked",
  "ip_address": "203.0.113.1",
  "country": "AU",
  "city": "Sydney"
}
```

### Violation Tracking

- **Counter**: `geofence_violations` field tracks failed attempts
- **Timestamp**: `geofence_last_violation` records when violations occur
- **Integration**: Violations can trigger self-destruct functionality

### Metrics

Key metrics to monitor:

- Geofence violation rates by country/IP range
- False positive rates (legitimate users blocked)
- Geolocation service availability and accuracy
- Performance impact of geofencing checks

## Future Enhancements

### Planned Features

1. **Advanced Geolocation**: Integration with MaxMind GeoLite2
2. **Time-based Restrictions**: Allow/deny access during specific time windows
3. **Risk-based Scoring**: Dynamic geofencing based on threat intelligence
4. **Whitelist Management**: User-friendly interface for managing restrictions
5. **Geofencing Templates**: Predefined restriction sets for common use cases

### Integration Opportunities

- **Threat Intelligence**: Dynamic geofencing based on threat feeds
- **User Behavior Analytics**: Adaptive restrictions based on user patterns
- **Compliance Frameworks**: Automated geofencing for regulatory requirements
- **Multi-factor Geofencing**: Combine location with device fingerprinting

## Troubleshooting

### Common Issues

1. **False Positives**: Users in allowed countries/IPs being blocked
   - Check proxy configuration and X-Forwarded-For headers
   - Verify geolocation service accuracy
   - Review CIDR range definitions

2. **Performance Impact**: Slow email access due to geofencing
   - Optimize database queries with proper indexing
   - Consider caching geolocation results
   - Review geolocation service response times

3. **Configuration Errors**: Invalid country codes or CIDR ranges
   - Validate settings before applying
   - Check error logs for validation failures
   - Use the validation functions provided

### Debugging

Enable debug logging to troubleshoot geofencing issues:

```go
log.Printf("Geofencing check - Email: %s, IP: %s, Result: %v", emailID, clientIP, result)
```

### Support

For geofencing-related issues:

1. Check audit logs for detailed error information
2. Verify geolocation service configuration
3. Review database schema and migration status
4. Test with mock services to isolate issues

## Conclusion

The geofencing implementation provides a robust, secure, and scalable solution for location-based access control. It integrates seamlessly with existing security features while maintaining privacy and preventing information leakage. The modular design allows for easy customization and future enhancements.
