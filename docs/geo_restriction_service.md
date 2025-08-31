# Geo-Restriction Service Documentation

## Overview

The Geo-Restriction Service is a comprehensive location-based access control system for Secure Email MVP that allows users to configure granular geo-restrictions for their emails. This service extends the existing geolocation functionality to support configurable allow/block lists for countries and cities, providing enhanced security and access control.

## Features

### Core Features

- **Configurable Allow/Block Lists**: Support for both allow and block rules with multiple countries and cities
- **Flexible Rule Types**: Allow rules (whitelist) and block rules (blacklist) with different matching logic
- **Strict Mode**: Option to require both country and city matches (AND logic) or either match (OR logic)
- **Default Actions**: Configurable default behavior when no rules are defined
- **Violation Tracking**: Automatic tracking of geo-restriction violations with timestamps
- **Real-time Enforcement**: Immediate enforcement during email access attempts
- **Audit Logging**: Comprehensive logging of all geo-restriction events

### Advanced Features

- **JSON-based Configuration**: Flexible rule and configuration storage using JSON
- **Rule Validation**: Comprehensive validation of country codes and city names
- **Normalization**: Automatic normalization of country codes and city names for consistent matching
- **Error Handling**: Graceful handling of geolocation failures with configurable fallback behavior
- **Performance Optimization**: Efficient rule evaluation with early termination

## Architecture

### Components

1. **GeoRestrictionService**: Core service for rule management and access evaluation
2. **GeoRestrictionRule**: Data structure representing individual geo-restriction rules
3. **GeoRestrictionConfig**: Configuration settings for the geo-restriction system
4. **API Handlers**: HTTP handlers for managing geo-restrictions via REST API
5. **Database Integration**: SQLite storage for rules, configuration, and violation tracking

### Data Flow

1. **Rule Creation**: Users create geo-restriction rules via API
2. **Rule Storage**: Rules are validated, normalized, and stored in JSON format
3. **Access Request**: When email access is requested, geolocation is determined
4. **Rule Evaluation**: Rules are evaluated in order with appropriate logic
5. **Access Decision**: Access is granted or denied based on rule evaluation
6. **Violation Tracking**: Failed attempts are logged and tracked

## API Endpoints

### Rule Management

#### GET /api/email/{id}/geo-restrictions
Retrieves all geo-restriction rules for an email.

**Response:**
```json
{
  "success": true,
  "message": "Geo-restriction rules retrieved successfully",
  "rules": [
    {
      "id": "rule1",
      "email_id": "email123",
      "type": "allow",
      "countries": ["us", "ca"],
      "cities": ["new york", "toronto"],
      "description": "Allow access from US and Canada",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### POST /api/email/{id}/geo-restrictions
Creates a new geo-restriction rule.

**Request Body:**
```json
{
  "email_id": "email123",
  "type": "allow",
  "countries": ["us", "ca"],
  "cities": ["new york", "toronto"],
  "description": "Allow access from US and Canada"
}
```

#### PUT /api/email/{id}/geo-restrictions/{ruleId}
Updates an existing geo-restriction rule.

**Request Body:**
```json
{
  "email_id": "email123",
  "type": "allow",
  "countries": ["us", "ca", "gb"],
  "cities": ["new york", "toronto", "london"],
  "description": "Updated: Allow access from US, Canada, and UK"
}
```

#### DELETE /api/email/{id}/geo-restrictions/{ruleId}
Deletes a geo-restriction rule.

### Configuration Management

#### GET /api/email/{id}/geo-restrictions/config
Retrieves geo-restriction configuration for an email.

**Response:**
```json
{
  "success": true,
  "message": "Geo-restriction configuration retrieved successfully",
  "config": {
    "enabled": true,
    "default_action": "allow",
    "strict_mode": false,
    "log_violations": true,
    "block_on_geolocation_failure": true
  }
}
```

#### PUT /api/email/{id}/geo-restrictions/config
Updates geo-restriction configuration.

**Request Body:**
```json
{
  "email_id": "email123",
  "enabled": true,
  "default_action": "allow",
  "strict_mode": false,
  "log_violations": true,
  "block_on_geolocation_failure": true
}
```

### Status and Monitoring

#### GET /api/email/{id}/geo-restrictions/status
Retrieves comprehensive status information for geo-restrictions.

**Response:**
```json
{
  "email_id": "email123",
  "enabled": true,
  "rules_count": 2,
  "violations_count": 0,
  "last_violation": null,
  "config": {
    "enabled": true,
    "default_action": "allow",
    "strict_mode": false,
    "log_violations": true,
    "block_on_geolocation_failure": true
  },
  "current_location": {
    "country": "us",
    "city": "new york",
    "ip": "192.168.1.1"
  },
  "access_allowed": true,
  "access_reason": ""
}
```

## Configuration Options

### GeoRestrictionConfig

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | true | Enable/disable geo-restrictions |
| `default_action` | string | "allow" | Default action when no rules exist |
| `strict_mode` | boolean | false | Require both country and city match |
| `log_violations` | boolean | true | Log violation attempts |
| `block_on_geolocation_failure` | boolean | true | Block access on geolocation failure |

### Rule Types

#### Allow Rules (Whitelist)
- **Purpose**: Explicitly allow access from specific locations
- **Logic**: Access granted if location matches any allow rule
- **Example**: Allow access from US and Canada

#### Block Rules (Blacklist)
- **Purpose**: Explicitly block access from specific locations
- **Logic**: Access denied if location matches any block rule
- **Example**: Block access from Russia and China

### Matching Logic

#### Non-Strict Mode (Default)
- **Country Match**: Access granted if country matches any rule
- **City Match**: Access granted if city matches any rule
- **Combination**: Either country OR city match is sufficient

#### Strict Mode
- **Country Match**: Country must match rule
- **City Match**: City must match rule
- **Combination**: Both country AND city must match

## Usage Examples

### Basic Allow Rule
```json
{
  "type": "allow",
  "countries": ["us", "ca"],
  "cities": ["new york", "toronto"],
  "description": "Allow access from US and Canada"
}
```

### Basic Block Rule
```json
{
  "type": "block",
  "countries": ["ru", "cn"],
  "description": "Block access from Russia and China"
}
```

### Country-Only Rule
```json
{
  "type": "allow",
  "countries": ["us", "ca", "gb"],
  "description": "Allow access from US, Canada, and UK"
}
```

### City-Only Rule
```json
{
  "type": "allow",
  "cities": ["new york", "london", "toronto"],
  "description": "Allow access from specific cities"
}
```

### Mixed Rule Set
```json
[
  {
    "type": "allow",
    "countries": ["us", "ca"],
    "cities": ["new york", "toronto"],
    "description": "Allow access from US and Canada"
  },
  {
    "type": "block",
    "countries": ["ru", "cn"],
    "description": "Block access from Russia and China"
  }
]
```

## Database Schema

### Enhanced Geo-Restriction Fields

The following fields are added to the `emails` table:

| Field | Type | Description |
|-------|------|-------------|
| `geo_restriction_rules` | TEXT | JSON array of geo-restriction rules |
| `geo_restriction_config` | TEXT | JSON object with configuration settings |
| `geo_restriction_enabled` | INTEGER | Boolean flag to enable/disable geo-restrictions |
| `geo_restriction_violations` | INTEGER | Counter for tracking violation attempts |
| `geo_restriction_last_violation` | DATETIME | Timestamp of the last violation attempt |

### Example Database Content

```sql
-- Example geo_restriction_rules JSON
[
  {
    "id": "rule1",
    "type": "allow",
    "countries": ["us", "ca"],
    "cities": ["new york", "toronto"],
    "description": "Allow access from US and Canada",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]

-- Example geo_restriction_config JSON
{
  "enabled": true,
  "default_action": "allow",
  "strict_mode": false,
  "log_violations": true,
  "block_on_geolocation_failure": true
}
```

## Security Considerations

### Privacy Protection
- **No PII Storage**: Only location metadata is stored, no personal information
- **Anonymized Logging**: Violation logs contain only location and timestamp data
- **Secure Transmission**: All API communications use HTTPS

### Access Control
- **Ownership Verification**: Users can only manage geo-restrictions for their own emails
- **JWT Authentication**: All endpoints require valid JWT authentication
- **Input Validation**: Comprehensive validation of all rule inputs

### Error Handling
- **Graceful Degradation**: System continues to function even if geolocation fails
- **Generic Error Messages**: Error responses don't leak sensitive information
- **Fallback Behavior**: Configurable behavior when geolocation services are unavailable

## Integration with Existing Systems

### Geolocation Service
- **Leverages Existing Infrastructure**: Uses the existing geolocation service
- **Consistent API**: Maintains compatibility with existing geolocation endpoints
- **Performance Optimization**: Reuses geolocation results across multiple checks

### Notification System
- **Violation Alerts**: Integrates with notification system for violation alerts
- **Audit Logging**: Logs all geo-restriction events for audit purposes
- **Event Tracking**: Tracks geo-restriction events in the access event history

### Brute Force Protection
- **Violation Counting**: Integrates with brute force protection system
- **Lockout Logic**: Contributes to overall security scoring
- **Reset Mechanisms**: Violation counts reset on successful access

## Testing

### Unit Tests
Comprehensive unit tests cover:
- Rule validation and normalization
- Access evaluation logic
- Configuration management
- JSON serialization/deserialization
- Error handling scenarios

### Integration Tests
PowerShell integration tests verify:
- API endpoint functionality
- Rule creation and management
- Configuration updates
- Status monitoring
- End-to-end access control

### Test Scenarios
- Allow rule matching
- Block rule matching
- Mixed rule sets
- Strict mode behavior
- Geolocation failure handling
- Violation tracking

## Monitoring and Logging

### Log Events
- **Rule Creation**: Log when new rules are created
- **Rule Updates**: Log when existing rules are modified
- **Rule Deletion**: Log when rules are removed
- **Access Attempts**: Log all access attempts with geo-restriction results
- **Violations**: Log failed access attempts due to geo-restrictions
- **Configuration Changes**: Log when configuration is updated

### Metrics
- **Rule Count**: Number of active rules per email
- **Violation Rate**: Frequency of geo-restriction violations
- **Access Success Rate**: Percentage of successful access attempts
- **Geolocation Success Rate**: Percentage of successful geolocation lookups

## Troubleshooting

### Common Issues

#### Geolocation Failures
- **Symptom**: Access denied due to geolocation failure
- **Cause**: External geolocation service unavailable
- **Solution**: Check geolocation service status, review `block_on_geolocation_failure` setting

#### Rule Not Working
- **Symptom**: Expected rule behavior not observed
- **Cause**: Rule configuration or matching logic issue
- **Solution**: Verify rule syntax, check strict mode setting, validate country/city codes

#### Performance Issues
- **Symptom**: Slow response times
- **Cause**: Large number of rules or inefficient evaluation
- **Solution**: Optimize rule count, review rule complexity

### Debug Information
- **Status Endpoint**: Use `/api/email/{id}/geo-restrictions/status` for current state
- **Log Analysis**: Review application logs for detailed error information
- **Rule Validation**: Verify rule syntax and configuration

## Future Enhancements

### Planned Features
- **IP Range Support**: Support for IP range-based restrictions
- **Time-based Rules**: Rules that apply only during specific time periods
- **Advanced Matching**: Regular expression support for city names
- **Rule Templates**: Predefined rule templates for common scenarios
- **Bulk Operations**: API endpoints for bulk rule management

### Performance Improvements
- **Caching**: Cache geolocation results for improved performance
- **Rule Optimization**: Optimize rule evaluation algorithms
- **Database Indexing**: Improve database query performance

### Security Enhancements
- **Rate Limiting**: Add rate limiting for geo-restriction API endpoints
- **Advanced Validation**: Enhanced input validation and sanitization
- **Audit Trail**: Comprehensive audit trail for all geo-restriction operations

## Conclusion

The Geo-Restriction Service provides a powerful and flexible location-based access control system for Secure Email MVP. With its comprehensive API, configurable rules, and robust security features, it enables users to implement sophisticated geo-restriction policies while maintaining ease of use and system reliability.

The service integrates seamlessly with existing security features and provides the foundation for advanced location-based security controls in the future.





















