# Amazon SES Test Script

This script tests Amazon SES email sending using the verified domain `securesystem.email`.

## Prerequisites

1. **Python 3.7+** installed
2. **AWS credentials** configured
3. **boto3** Python library installed
4. **Verified domain** in Amazon SES

## Setup

### 1. Install Dependencies

```bash
pip install -r requirements.txt
```

### 2. Configure AWS Credentials

Choose one of these methods:

#### Option A: AWS CLI (Recommended)
```bash
aws configure
```
Enter your:
- AWS Access Key ID
- AWS Secret Access Key
- Default region: `us-east-1`
- Default output format: `json`

#### Option B: Environment Variables
```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_DEFAULT_REGION=us-east-1
```

#### Option C: IAM Role (if running on EC2)
The script will automatically use IAM role credentials if available.

## Running the Test

### Basic Test
```bash
python ses_test.py
```

### Expected Output
```
🚀 Amazon SES Test Script for Secure Email MVP
==================================================
🔍 Checking SES verification status...
   Domain securesystem.email: Success
   ✅ Domain is verified!
   Email test@securesystem.email: Success
   ✅ Email is verified!

🔧 Creating SES client...
📊 Checking SES account status...
   ✅ SES Quota: 0/200 emails sent in last 24h
📝 Creating email message...
📤 Sending test email...
   From: test@securesystem.email
   To: cpigusch@gmail.com
   Subject: SES Test Email - Secure Email MVP

✅ SUCCESS: Email sent successfully!
📧 Message ID: 0000018a12345678-12345678-1234-1234-1234-123456789012-000000
📬 Check your inbox at: cpigusch@gmail.com

🔍 Verification Steps:
   1. Check your Gmail inbox for the test email
   2. Open the email and click 'Show original' (three dots menu)
   3. Look for these authentication headers:
      - Authentication-Results: spf=PASS
      - Authentication-Results: dkim=PASS
      - Authentication-Results: dmarc=PASS

🎉 Test completed successfully!
📧 Message ID: 0000018a12345678-12345678-1234-1234-1234-123456789012-000000
📬 Check your email at: cpigusch@gmail.com
```

## Verification Steps

### 1. Check Email Delivery
- Look for the test email in your Gmail inbox
- Check spam folder if not found

### 2. Verify Email Authentication
1. Open the email in Gmail
2. Click the three dots menu (⋮)
3. Select "Show original"
4. Look for these headers:

```
Authentication-Results: mx.google.com;
       dkim=PASS header.i=@securesystem.email;
       spf=PASS (google.com: domain of test@securesystem.email designates 54.240.27.0/24 as permitted sender) smtp.mailfrom=test@securesystem.email;
       dmarc=PASS (p=QUARANTINE sp=QUARANTINE dis=NONE) header.from=securesystem.email
```

### 3. Expected Results
- ✅ **SPF = PASS**: Sender Policy Framework verification passed
- ✅ **DKIM = PASS**: DomainKeys Identified Mail signature verified
- ✅ **DMARC = PASS**: Domain-based Message Authentication, Reporting & Conformance passed

## Troubleshooting

### Common Errors

#### 1. NoCredentialsError
```
❌ ERROR: AWS credentials not found!
```
**Solution**: Configure AWS credentials using `aws configure` or environment variables.

#### 2. MessageRejected
```
❌ SES ERROR: MessageRejected
```
**Solutions**:
- Check if domain is verified in SES console
- Verify you're not in SES sandbox mode
- Check if recipient email is verified (if in sandbox)

#### 3. MailFromDomainNotVerified
```
❌ SES ERROR: MailFromDomainNotVerified
```
**Solution**: Verify the MAIL FROM domain in SES console.

#### 4. ConfigurationSetDoesNotExist
```
❌ SES ERROR: ConfigurationSetDoesNotExist
```
**Solution**: Remove configuration set reference or create the configuration set.

### SES Console Verification

1. Go to [Amazon SES Console](https://console.aws.amazon.com/ses/)
2. Select region: **US East (N. Virginia)**
3. Check **Verified identities**:
   - Domain: `securesystem.email` should show "Verified"
   - Email: `test@securesystem.email` should show "Verified"

### DNS Verification

Ensure these DNS records are properly configured:

#### SPF Record
```
securesystem.email. IN TXT "v=spf1 include:amazonses.com -all"
```

#### DKIM Records
```
[selector]._domainkey.securesystem.email. IN CNAME [selector].dkim.amazonses.com.
```

#### MX Record (for bounce handling)
```
bounce.securesystem.email. IN MX 10 feedback-smtp.us-east-1.amazonses.com.
```

## Customization

### Change Recipient Email
Edit the `TO_EMAIL` variable in `ses_test.py`:
```python
TO_EMAIL = 'your-email@gmail.com'
```

### Change Sender Email
Edit the `FROM_EMAIL` variable in `ses_test.py`:
```python
FROM_EMAIL = 'your-name@securesystem.email'
```

### Change AWS Region
Edit the `AWS_REGION` variable in `ses_test.py`:
```python
AWS_REGION = 'us-west-2'  # or your preferred region
```

## Security Notes

- The script uses `send_raw_email` for maximum control over email headers
- Both plain text and HTML versions are included for better deliverability
- The script includes comprehensive error handling and troubleshooting information
- AWS credentials should be kept secure and never committed to version control

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Verify AWS credentials and permissions
3. Check SES console for verification status
4. Review AWS CloudTrail logs for detailed error information

