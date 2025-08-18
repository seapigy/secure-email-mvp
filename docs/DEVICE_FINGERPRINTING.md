# Device Fingerprinting & Trusted Devices

## Overview

**Micro-Iteration 4.14** implements device fingerprinting and trusted devices functionality for the SecureChat Email system. This feature allows emails to be restricted to specific devices that have been previously authorized through MFA verification, providing an additional layer of security beyond location-based restrictions.

## Features

### 🔐 **Device Fingerprinting**
- **Deterministic Fingerprinting**: Creates unique device identifiers based on User-Agent, IP subnet, and optional browser hints
- **Privacy-Preserving**: Uses IP subnets (/24 for IPv4, /64 for IPv6) instead of exact IP addresses
- **Stable Identification**: Normalizes User-Agent strings to handle minor browser updates
- **Argon2id Hashing**: Securely hashes fingerprints using email ID as salt

### 🛡️ **Trusted Devices Management**
- **Per-Email Control**: Each email can be configured to require trusted devices only
- **MFA Integration**: Devices can only be trusted after successful MFA verification
- **Access Tracking**: Monitors device usage patterns and access counts
- **Device Removal**: Supports removing devices from trusted lists

### 🔄 **Security Flow Integration**
- **Updated Security Order**: JWT → Revoke/Time → MFA/Decoy → Geofence → **Trusted Device** → Read-once
- **Self-Destruct Integration**: Failed device checks increment self-destruct counter
- **Generic Error Handling**: Prevents information leakage through standardized error messages

## Architecture

### Database Schema

#### New Columns in `emails` Table
```sql
-- Add trusted_devices_only column to emails table
ALTER TABLE emails ADD COLUMN trusted_devices_only BOOLEAN DEFAULT FALSE;
```

#### New `trusted_devices` Table
```sql
CREATE TABLE trusted_devices (
    id TEXT PRIMARY KEY,                    -- UUID for the trusted device record
    email_id TEXT NOT NULL,                 -- Foreign key to emails table
    device_hash TEXT NOT NULL,              -- Argon2id hash of device fingerprint
    device_fingerprint TEXT NOT NULL,       -- Original fingerprint (for debugging)
    user_agent TEXT,                        -- User-Agent string
    ip_address TEXT,                        -- IP address (for audit purposes)
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,  -- When device was trusted
    last_used_at DATETIME,                  -- Last successful access from this device
    access_count INTEGER DEFAULT 0,         -- Number of successful accesses
    
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);
```

#### Indexes for Performance
```sql
CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_id ON trusted_devices(email_id);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_device_hash ON trusted_devices(device_hash);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_hash ON trusted_devices(email_id, device_hash);
CREATE INDEX IF NOT EXISTS idx_emails_trusted_devices_only ON emails(trusted_devices_only);
```

### Service Architecture

#### DeviceFingerprintService Interface
```go
type DeviceFingerprintService interface {
    GenerateFingerprint(userAgent, clientIP string, browserHints map[string]string) (string, error)
    HashFingerprint(fingerprint, emailID string) (string, error)
    IsDeviceTrusted(emailID, deviceHash string) (bool, error)
    TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP string) error
    UpdateDeviceAccess(emailID, deviceHash string) error
    GetDeviceInfo(emailID, deviceHash string) (*DeviceInfo, error)
    GetTrustedDevices(emailID string) ([]*DeviceInfo, error)
    RemoveTrustedDevice(emailID, deviceHash string) error
}
```

#### Components
- **DeviceFingerprintServiceImpl**: Main implementation with database integration
- **MockDeviceFingerprintService**: Mock implementation for testing
- **DeviceInfo**: Data structure for device information

## Implementation Details

### Device Fingerprinting Algorithm

#### Fingerprint Generation
1. **User-Agent Normalization**: Convert to lowercase, remove version numbers
2. **IP Subnet Extraction**: Use /24 for IPv4, /64 for IPv6 for privacy
3. **Browser Hints Integration**: Optional additional device characteristics
4. **Deterministic Ordering**: Sort components for consistent fingerprints
5. **SHA-256 Hashing**: Create final fingerprint hash

#### Fingerprint Hashing
- **Argon2id**: Use email ID as salt for additional security
- **Parameters**: time=1, memory=64MB, threads=4, keyLen=32
- **Deterministic**: Same fingerprint + email ID always produces same hash

### Security Check Integration

#### Updated Security Flow
```go
// Step 9.6: Check device fingerprinting restrictions (Micro-Iteration 4.14)
if srv.deviceFingerprintService != nil {
    // Check if email requires trusted devices only
    var trustedDevicesOnly bool
    err = srv.db.QueryRow("SELECT trusted_devices_only FROM emails WHERE email_id = ?", emailID).Scan(&trustedDevicesOnly)
    
    if trustedDevicesOnly {
        // Generate device fingerprint
        fingerprint, err := srv.deviceFingerprintService.GenerateFingerprint(userAgent, clientIP, nil)
        
        // Hash the fingerprint
        deviceHash, err := srv.deviceFingerprintService.HashFingerprint(fingerprint, emailID)
        
        // Check if device is trusted
        isTrusted, err := srv.deviceFingerprintService.IsDeviceTrusted(emailID, deviceHash)
        
        if !isTrusted {
            // Deny access with generic error
            return "Email has been revoked or cannot be accessed"
        }
        
        // Update device access tracking
        srv.deviceFingerprintService.UpdateDeviceAccess(emailID, deviceHash)
    }
}
```

### API Endpoints

#### Trust Device Endpoint
```
POST /api/email/{id}/trust-device
```

**Purpose**: Trust the current device for a specific email after successful MFA verification

**Request Headers**:
- `Authorization: Bearer <JWT_TOKEN>`
- `User-Agent`: Browser/device identification
- `X-Forwarded-For`: Client IP address

**Response**:
```json
{
    "success": true,
    "message": "Device trusted successfully"
}
```

**Security Requirements**:
- Valid JWT token required
- User must be sender or recipient of the email
- Device fingerprinting service must be available

## Usage Examples

### Enabling Device Fingerprinting for an Email

#### Database Update
```sql
UPDATE emails 
SET trusted_devices_only = TRUE 
WHERE email_id = 'your-email-id';
```

#### API Usage
```bash
# Trust current device for an email
curl -X POST "https://api.securesystem.email/api/email/your-email-id/trust-device" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
```

### Device Fingerprinting Flow

1. **Email Creation**: Set `trusted_devices_only = TRUE` for sensitive emails
2. **First Access**: User attempts to access email with device fingerprinting enabled
3. **MFA Verification**: User completes MFA verification
4. **Device Trusting**: User calls trust device endpoint to authorize current device
5. **Subsequent Access**: Device is automatically recognized and allowed access
6. **New Device**: Different device requires new MFA + trust process

## Testing

### Unit Tests
- **MockDeviceFingerprintService**: Comprehensive test coverage for all service methods
- **Fingerprint Generation**: Tests deterministic fingerprint creation
- **Device Trust Management**: Tests device trust, untrust, and access tracking
- **Error Handling**: Tests various error scenarios

### Integration Tests
- **End-to-End Flow**: Tests complete device fingerprinting integration
- **Server Integration**: Tests with actual server components
- **Settings Management**: Tests device fingerprinting enable/disable

### Test Coverage
```bash
# Run device fingerprinting tests
go test ./pkg/devicefingerprint/... -v
go test ./cmd/api/device_fingerprinting_integration_test.go -v
```

## Configuration

### Environment Variables
No additional environment variables required. Device fingerprinting uses existing database and service infrastructure.

### Database Migration
The device fingerprinting migration is automatically applied during server startup:
```sql
-- Apply device fingerprinting migration (Micro-Iteration 4.14)
-- Located in: schema/migrate_add_device_fingerprinting.sql
```

## Security Considerations

### Privacy Protection
- **IP Subnet Usage**: Uses /24 and /64 subnets instead of exact IPs
- **User-Agent Normalization**: Removes version numbers for stability
- **No Raw Data Storage**: Only hashed fingerprints stored in database

### Security Measures
- **Argon2id Hashing**: Industry-standard password hashing for fingerprint storage
- **Email-Specific Salting**: Each email uses its ID as salt for fingerprint hashing
- **Generic Error Messages**: Prevents information leakage through error responses
- **Audit Logging**: Comprehensive logging of device trust events

### Threat Mitigation
- **Device Spoofing**: Fingerprint includes multiple device characteristics
- **IP Spoofing**: Uses subnet-based identification for resilience
- **Brute Force**: Argon2id provides protection against fingerprint guessing
- **Information Leakage**: Generic error messages prevent device enumeration

## Monitoring and Logging

### Audit Events
- **device_trust**: Logs device trust operations (success/failure)
- **email_access**: Includes device fingerprinting results in access logs

### Log Fields
- **Device Hash**: Hashed device fingerprint (never raw data)
- **User Agent**: Browser/device identification
- **IP Address**: Client IP for audit purposes
- **Access Count**: Number of successful accesses from device

### Monitoring Metrics
- **Trusted Device Count**: Number of trusted devices per email
- **Device Access Patterns**: Frequency and timing of device access
- **Failed Device Checks**: Number of untrusted device access attempts

## Future Enhancements

### Planned Features
- **Device Expiration**: Automatic removal of unused trusted devices
- **Device Categories**: Different trust levels (high, medium, low)
- **Advanced Fingerprinting**: Additional device characteristics (screen resolution, timezone, etc.)
- **Device Analytics**: Detailed usage analytics and anomaly detection

### Potential Improvements
- **Machine Learning**: Anomaly detection for suspicious device patterns
- **Risk Scoring**: Dynamic risk assessment based on device characteristics
- **Geolocation Integration**: Combine device fingerprinting with location data
- **Multi-Device Support**: Enhanced management of multiple trusted devices

## Troubleshooting

### Common Issues

#### Device Not Recognized
- **Cause**: Device fingerprint changed due to browser update or IP change
- **Solution**: Re-trust the device through MFA verification

#### Fingerprint Generation Errors
- **Cause**: Invalid IP address or User-Agent format
- **Solution**: Check client IP and User-Agent headers

#### Database Errors
- **Cause**: Migration not applied or database corruption
- **Solution**: Verify migration status and database integrity

### Debug Information
- **Fingerprint Components**: Logged during development for debugging
- **Device Hash**: Available in audit logs for troubleshooting
- **Access Patterns**: Tracked for security analysis

## API Reference

### DeviceFingerprintService Methods

#### GenerateFingerprint
```go
func (d *DeviceFingerprintServiceImpl) GenerateFingerprint(
    userAgent, clientIP string, 
    browserHints map[string]string) (string, error)
```
Creates a deterministic device fingerprint from device characteristics.

#### HashFingerprint
```go
func (d *DeviceFingerprintServiceImpl) HashFingerprint(
    fingerprint, emailID string) (string, error)
```
Creates an Argon2id hash of the fingerprint using email ID as salt.

#### IsDeviceTrusted
```go
func (d *DeviceFingerprintServiceImpl) IsDeviceTrusted(
    emailID, deviceHash string) (bool, error)
```
Checks if a device hash is trusted for a specific email.

#### TrustDevice
```go
func (d *DeviceFingerprintServiceImpl) TrustDevice(
    emailID, deviceHash, fingerprint, userAgent, clientIP string) error
```
Adds a device to the trusted devices list.

#### UpdateDeviceAccess
```go
func (d *DeviceFingerprintServiceImpl) UpdateDeviceAccess(
    emailID, deviceHash string) error
```
Updates the last access time and count for a trusted device.

#### GetDeviceInfo
```go
func (d *DeviceFingerprintServiceImpl) GetDeviceInfo(
    emailID, deviceHash string) (*DeviceInfo, error)
```
Retrieves information about a specific trusted device.

#### GetTrustedDevices
```go
func (d *DeviceFingerprintServiceImpl) GetTrustedDevices(
    emailID string) ([]*DeviceInfo, error)
```
Retrieves all trusted devices for an email.

#### RemoveTrustedDevice
```go
func (d *DeviceFingerprintServiceImpl) RemoveTrustedDevice(
    emailID, deviceHash string) error
```
Removes a device from the trusted devices list.

## Conclusion

Device fingerprinting and trusted devices provide a robust additional layer of security for the SecureChat Email system. By combining device identification with MFA verification, this feature ensures that sensitive emails can only be accessed from authorized devices, significantly reducing the risk of unauthorized access.

The implementation follows security best practices with privacy-preserving fingerprinting, secure hashing, comprehensive audit logging, and generic error handling. The modular design allows for easy testing and future enhancements while maintaining compatibility with existing security features.
