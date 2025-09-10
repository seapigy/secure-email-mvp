# Oracle Cloud Infrastructure CLI Setup Guide

## 1. Install OCI CLI

### On Windows (PowerShell):
```powershell
# Download and install OCI CLI
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/oracle/oci-cli/master/scripts/install/install.ps1" -OutFile "install.ps1"
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
.\install.ps1 -AcceptAllDefaults
```

### On Linux/macOS:
```bash
bash -c "$(curl -L https://raw.githubusercontent.com/oracle/oci-cli/master/scripts/install/install.sh)"
```

## 2. Generate API Key Pair

### Option A: Using OCI CLI (Recommended)
```bash
# Generate a new key pair
mkdir -p ~/.oci
openssl genrsa -out ~/.oci/oci_api_key.pem 2048
chmod 600 ~/.oci/oci_api_key.pem
openssl rsa -pubout -in ~/.oci/oci_api_key.pem -out ~/.oci/oci_api_key_public.pem

# Get the fingerprint
openssl rsa -pubout -outform DER -in ~/.oci/oci_api_key.pem | openssl md5 -c
```

### Option B: Using OpenSSL directly
```bash
# Generate private key
openssl genrsa -out ~/.oci/oci_api_key.pem 2048
chmod 600 ~/.oci/oci_api_key.pem

# Generate public key
openssl rsa -pubout -in ~/.oci/oci_api_key.pem -out ~/.oci/oci_api_key_public.pem

# Get fingerprint
openssl rsa -pubout -outform DER -in ~/.oci/oci_api_key.pem | openssl md5 -c
```

## 3. Upload Public Key to OCI Console

1. Go to [OCI Console](https://cloud.oracle.com)
2. Navigate to **Identity & Security** → **Users**
3. Click on your user
4. Go to **API Keys** section
5. Click **Add API Key**
6. Upload the public key file (`~/.oci/oci_api_key_public.pem`)
7. Copy the **Fingerprint** that appears

## 4. Get Your OCIDs

### User OCID:
1. Go to **Identity & Security** → **Users**
2. Click on your user
3. Copy the **OCID** from the user details

### Tenancy OCID:
1. Go to **Administration** → **Tenancy Details**
2. Copy the **OCID** from the tenancy information

### Compartment OCID:
1. Go to **Identity & Security** → **Compartments**
2. Click on your compartment (usually the root compartment)
3. Copy the **OCID**

## 5. Create ~/.oci/config File

Create the file at `~/.oci/config` with the following content:

```ini
[DEFAULT]
user=ocid1.user.oc1..aaaaaaaa...your-user-ocid-here
fingerprint=aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp
key_file=~/.oci/oci_api_key.pem
tenancy=ocid1.tenancy.oc1..aaaaaaaa...your-tenancy-ocid-here
region=us-ashburn-1
pass_phrase=
```

### Windows Path:
- File location: `C:\Users\{username}\.oci\config`
- Key file location: `C:\Users\{username}\.oci\oci_api_key.pem`

### Linux/macOS Path:
- File location: `~/.oci/config`
- Key file location: `~/.oci/oci_api_key.pem`

## 6. Test Configuration

```bash
# Test the configuration
oci iam user get --user-id $(oci iam user list --query 'data[0].id' --raw-output)

# List compartments
oci iam compartment list

# List availability domains
oci iam availability-domain list
```

## 7. Set Up GitHub Secrets

For the CI/CD pipeline, you need to set these secrets in your GitHub repository:

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Add the following secrets:

### OCI_CONFIG
```bash
# Base64 encode your config file
cat ~/.oci/config | base64 -w 0
```

### OCI_PRIVATE_KEY
```bash
# Base64 encode your private key
cat ~/.oci/oci_api_key.pem | base64 -w 0
```

### OCI_INSTANCE_ID
```bash
# Get your compute instance ID (after creating it)
oci compute instance list --compartment-id <your-compartment-id>
```

### OCI_COMPARTMENT_ID
```bash
# Your compartment OCID (from step 4)
```

## 8. Common Issues and Solutions

### Permission Denied
```bash
chmod 600 ~/.oci/config
chmod 600 ~/.oci/oci_api_key.pem
```

### Invalid Fingerprint
- Make sure you copied the fingerprint correctly from the OCI console
- The fingerprint should be in the format: `aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp`

### Region Issues
- Make sure the region in your config matches your OCI tenancy region
- Common regions: `us-ashburn-1`, `us-phoenix-1`, `eu-frankfurt-1`, `uk-london-1`

### Key File Not Found
- Ensure the key file path in the config is correct
- Use absolute paths if relative paths don't work

## 9. Security Best Practices

1. **Never commit your private key or config file to version control**
2. **Use environment variables for sensitive data in production**
3. **Rotate your API keys regularly**
4. **Use least privilege principle for user permissions**
5. **Enable MFA on your OCI account**

## 10. Troubleshooting

### Check OCI CLI version
```bash
oci --version
```

### Verify configuration
```bash
oci setup config --list
```

### Test connectivity
```bash
oci iam user get --user-id $(oci iam user list --query 'data[0].id' --raw-output)
```

If you encounter any issues, check the [OCI CLI documentation](https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm) for detailed troubleshooting steps.
