# SMTP Integration for Secure Email MVP

This document describes the Amazon SES SMTP integration implemented for the Secure Email MVP backend.

## Overview

The SMTP integration allows the Secure Email MVP to send notification emails to recipients when secure emails are created. This is implemented using Amazon SES (Simple Email Service) via SMTP for reliable email delivery.

## Features

- **Secure SMTP Connection**: TLS-encrypted connection to Amazon SES
- **Environment-based Configuration**: All credentials stored in environment variables
- **Error Handling**: Graceful handling of SMTP failures without crashing the server
- **Modular Design**: Easy to swap SES for other SMTP providers
- **Health Check Endpoints**: Test SMTP connectivity and configuration status

## Configuration

### Environment Variables

Set the following environment variables in your `.env` file:

```bash
# Amazon SES SMTP Configuration
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=ses-smtp-user.20250821-123501
SMTP_PASSWORD=your-ses-smtp-password
SMTP_FROM=secure-email@securesystem.email
```

### Amazon SES Setup

1. **Verify Domain/Email**: Ensure your sending domain or email is verified in Amazon SES
2. **Create SMTP Credentials**: Generate SMTP username and password in the SES console
3. **Configure Sending Limits**: Set appropriate sending limits for your use case

## API Endpoints

### Test SMTP Connectivity

**POST** `/api/email/test-smtp`

Test SMTP connectivity by sending a test email.

**Request Body:**
```json
{
  "test_email": "recipient@example.com"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "SMTP test successful - test email sent to recipient@example.com",
  "config": {
    "host": "email-smtp.us-east-1.amazonaws.com",
    "port": "587",
    "from": "secure-email@securesystem.email"
  }
}
```

### Check SMTP Configuration

**GET** `/api/email/smtp-config`

Check SMTP configuration status without sending emails.

**Response:**
```json
{
  "status": "configured",
  "message": "SMTP configuration appears to be complete",
  "config": {
    "SMTP_HOST": "email-smtp.us-east-1.amazonaws.com",
    "SMTP_PORT": "587",
    "SMTP_USERNAME": "***HIDDEN***",
    "SMTP_PASSWORD": "***HIDDEN***",
    "SMTP_FROM": "secure-email@securesystem.email"
  },
  "missing": []
}
```

## Implementation Details

### SMTP Mailer (`pkg/mailer/smtp_mailer.go`)

The core SMTP functionality is implemented in the `SMTPMailer` struct:

```go
type SMTPMailer struct {
    config *SMTPConfig
    client *smtp.Client
}
```

**Key Methods:**
- `NewSMTPMailer()`: Initialize SMTP mailer with environment variables
- `SendEmail(msg EmailMessage)`: Send an email via SMTP
- `TestConnection(testEmail string)`: Test SMTP connectivity
- `GetConfig()`: Get configuration (without sensitive data)

### Email Security Service Integration

The SMTP mailer is integrated into the `EmailSecurityService`:

```go
type EmailSecurityService struct {
    // ... other fields
    smtpMailer *mailer.SMTPMailer
}
```

**Notification Email:**
When a secure email is sent, a notification email is automatically sent to the recipient with the following content:

```
Subject: You have a new secure message

Hello,

You have received a secure message. Please log into SecureSystem.email to view it.

Message ID: email_1234567890_abcdefgh

Best regards,
Secure Email MVP System
```

### Error Handling

The SMTP integration includes comprehensive error handling:

1. **Configuration Errors**: Validates all required environment variables
2. **Connection Errors**: Handles SMTP connection failures gracefully
3. **Authentication Errors**: Manages SMTP authentication issues
4. **Non-blocking**: SMTP failures don't prevent email creation

**Error Logging:**
```
⚠️ SMTP mailer not available: SMTP_HOST environment variable is required
❌ Failed to send notification email to user@example.com: failed to connect to SMTP server
✅ SMTP notification email sent successfully to: user@example.com
```

## Testing

### Unit Tests

Run the SMTP mailer unit tests:

```bash
go test -v ./pkg/mailer/
```

### Integration Tests

Run SMTP integration tests (requires valid SES credentials):

```bash
INTEGRATION_TESTS=1 go test -v ./cmd/api/ -run TestSMTPIntegration
```

### Manual Testing

1. **Test Configuration:**
   ```bash
   curl -X GET http://localhost:8080/api/email/smtp-config
   ```

2. **Test SMTP Connectivity:**
   ```bash
   curl -X POST http://localhost:8080/api/email/test-smtp \
     -H "Content-Type: application/json" \
     -d '{"test_email": "your-email@example.com"}'
   ```

## Security Considerations

1. **Credential Protection**: SMTP credentials are never exposed in API responses
2. **TLS Encryption**: All SMTP connections use TLS encryption
3. **Environment Variables**: Credentials stored in environment variables, not in code
4. **Error Sanitization**: Error messages don't expose sensitive information

## Troubleshooting

### Common Issues

1. **"SMTP_HOST environment variable is required"**
   - Ensure all SMTP environment variables are set in `.env` file

2. **"Failed to connect to SMTP server"**
   - Verify SMTP_HOST and SMTP_PORT are correct
   - Check network connectivity to Amazon SES

3. **"Failed to authenticate"**
   - Verify SMTP_USERNAME and SMTP_PASSWORD are correct
   - Ensure SES credentials are active

4. **"Email not verified"**
   - Verify sending email address in Amazon SES console
   - Check SES sending limits and account status

### Debugging

Enable detailed logging by checking server logs for SMTP-related messages:

```
📧 SMTP Mailer initialized with host: email-smtp.us-east-1.amazonaws.com:587
📧 Attempting to send email to: user@example.com
✅ Email sent successfully to: user@example.com
```

## Future Enhancements

1. **Multiple SMTP Providers**: Support for other SMTP services (SendGrid, Mailgun, etc.)
2. **Email Templates**: Customizable notification email templates
3. **Retry Logic**: Automatic retry for failed email sends
4. **Rate Limiting**: SMTP rate limiting to prevent abuse
5. **Email Tracking**: Track email delivery and bounce rates

## Migration from Mock Email

The SMTP integration replaces the previous mock email functionality:

- **Before**: Email notifications were logged but not sent
- **After**: Real email notifications sent via Amazon SES SMTP

The system gracefully handles cases where SMTP is not configured by falling back to logging only.







