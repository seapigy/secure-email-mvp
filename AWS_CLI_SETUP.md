# AWS CLI Setup Guide for Secure Email MVP

This guide provides step-by-step instructions for installing and configuring AWS CLI for the Secure Email MVP project.

## 🚀 Quick Start

### Prerequisites
- Windows 11/10, macOS, or Linux
- PowerShell (Windows) or Terminal (macOS/Linux)
- AWS Account with SES access
- IAM user with appropriate permissions

### Required IAM Permissions
Your AWS user needs these permissions for SES testing:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ses:SendEmail",
                "ses:SendRawEmail",
                "ses:GetSendQuota",
                "ses:GetIdentityVerificationAttributes",
                "ses:ListIdentities",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        }
    ]
}
```

## 📦 Installation

### Windows (Recommended)
```powershell
# Install using winget
winget install Amazon.AWSCLI

# Or download from AWS website
# https://aws.amazon.com/cli/
```

### macOS
```bash
# Install using Homebrew
brew install awscli

# Or download installer
curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
sudo installer -pkg AWSCLIV2.pkg -target /
```

### Linux
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install awscli

# Or download installer
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

## ⚙️ Configuration

### Method 1: Automated Setup (Recommended)
Use our setup script with your AWS credentials:

```powershell
# Windows PowerShell
.\setup_aws_cli.ps1 -AccessKeyId "AKIA..." -SecretAccessKey "your_secret_key"
```

### Method 2: Manual Configuration
```bash
# Configure AWS CLI
aws configure

# Enter your credentials when prompted:
# AWS Access Key ID: AKIA...
# AWS Secret Access Key: your_secret_key
# Default region name: us-east-1
# Default output format: json
```

### Method 3: Environment Variables
```bash
# Set environment variables
export AWS_ACCESS_KEY_ID="AKIA..."
export AWS_SECRET_ACCESS_KEY="your_secret_key"
export AWS_DEFAULT_REGION="us-east-1"
```

## 🔍 Verification

### 1. Check Installation
```bash
aws --version
# Expected: aws-cli/2.x.x Python/x.x.x
```

### 2. Test Authentication
```bash
aws sts get-caller-identity
# Expected: JSON with Account, UserId, and ARN
```

### 3. Test SES Access
```bash
aws ses get-send-quota --region us-east-1
# Expected: JSON with Max24HourSend, SentLast24Hours, etc.
```

### 4. List Verified Identities
```bash
aws ses list-identities --region us-east-1
# Expected: List of verified domains and email addresses
```

## 🧪 Testing

### Run SES Test Script
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
```

## 🔒 Security Configuration

### 1. File Locations
AWS CLI stores credentials in:
- **Windows**: `%USERPROFILE%\.aws\`
- **macOS/Linux**: `~/.aws/`

### 2. File Structure
```
.aws/
├── credentials    # Access keys (NEVER commit this)
└── config         # Region and output format
```

### 3. Git Protection
Our `.gitignore` includes:
```
# AWS credentials and configuration
.aws/
.aws/credentials
.aws/config
aws-credentials.json
aws-config.json
```

### 4. Credential Rotation
Rotate your AWS access keys regularly:
1. Create new access key in AWS Console
2. Update local configuration
3. Delete old access key
4. Test with new credentials

## 🚀 CI/CD Configuration

### GitHub Actions
Our workflow (`.github/workflows/aws-ses-test.yml`) uses GitHub Secrets:

#### Required Secrets
1. `AWS_ACCESS_KEY_ID`: Your AWS Access Key ID
2. `AWS_SECRET_ACCESS_KEY`: Your AWS Secret Access Key

#### Setting Up Secrets
1. Go to your GitHub repository
2. Navigate to Settings → Secrets and variables → Actions
3. Add the required secrets

#### Workflow Features
- ✅ Automatic testing on push/PR
- ✅ Manual trigger support
- ✅ SES quota checking
- ✅ Identity verification
- ✅ Email sending test

### Other CI/CD Platforms

#### GitLab CI
```yaml
variables:
  AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID
  AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY
  AWS_DEFAULT_REGION: us-east-1

test-ses:
  script:
    - pip install -r requirements.txt
    - python ses_test.py
```

#### Azure DevOps
```yaml
variables:
  AWS_ACCESS_KEY_ID: $(AWS_ACCESS_KEY_ID)
  AWS_SECRET_ACCESS_KEY: $(AWS_SECRET_ACCESS_KEY)
  AWS_DEFAULT_REGION: us-east-1

steps:
- script: |
    pip install -r requirements.txt
    python ses_test.py
```

## 🛠️ Troubleshooting

### Common Issues

#### 1. "aws: command not found"
**Solution**: Install AWS CLI or add to PATH
```bash
# Windows: Restart terminal after installation
# macOS/Linux: Add to PATH in ~/.bashrc or ~/.zshrc
export PATH=$PATH:/usr/local/bin
```

#### 2. "Unable to locate credentials"
**Solution**: Configure credentials
```bash
aws configure
# Or set environment variables
export AWS_ACCESS_KEY_ID="your_key"
export AWS_SECRET_ACCESS_KEY="your_secret"
```

#### 3. "Access Denied" for SES
**Solution**: Check IAM permissions
- Verify user has SES permissions
- Check if SES is enabled in your region
- Ensure domain/email is verified

#### 4. "MessageRejected" error
**Solution**: Check SES configuration
- Verify domain in SES console
- Check if in sandbox mode
- Verify recipient email (if in sandbox)

### Debug Commands
```bash
# Check AWS CLI version
aws --version

# Test authentication
aws sts get-caller-identity

# Check SES status
aws ses get-send-quota --region us-east-1

# List verified identities
aws ses list-identities --region us-east-1

# Check domain verification
aws ses get-identity-verification-attributes \
  --identities securesystem.email --region us-east-1
```

## 📚 Additional Resources

### AWS Documentation
- [AWS CLI Installation](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- [AWS CLI Configuration](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html)
- [SES Developer Guide](https://docs.aws.amazon.com/ses/latest/dg/Welcome.html)

### Security Best Practices
- [AWS Security Best Practices](https://docs.aws.amazon.com/wellarchitected/latest/security-pillar/welcome.html)
- [IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [SES Security](https://docs.aws.amazon.com/ses/latest/dg/security.html)

### Support
- [AWS Support](https://aws.amazon.com/support/)
- [SES Forum](https://forums.aws.amazon.com/forum.jspa?forumID=90)
- [GitHub Issues](https://github.com/your-repo/issues)

## 🎯 Next Steps

After completing AWS CLI setup:

1. **Test SES Configuration**
   ```bash
   python ses_test.py
   ```

2. **Verify Email Authentication**
   - Check Gmail headers for SPF/DKIM/DMARC
   - Verify domain reputation

3. **Production Deployment**
   - Use IAM roles instead of access keys
   - Set up monitoring and alerting
   - Configure SES sending limits

4. **Security Hardening**
   - Enable SES reputation monitoring
   - Set up bounce and complaint handling
   - Configure DMARC reporting

---

**Remember**: Never commit AWS credentials to version control. Always use environment variables or IAM roles in production environments.
