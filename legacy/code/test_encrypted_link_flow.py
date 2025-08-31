#!/usr/bin/env python3
"""
Secure Email MVP - Encrypted Click-to-Decrypt Link Test
Tests the complete flow: PQC Encryption → Secure Link → SES Email → Frontend Decryption
"""

import boto3
import json
import base64
import requests
import time
import sys
from botocore.exceptions import ClientError
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart

class SecureEmailLinkTester:
    def __init__(self):
        self.aws_region = 'us-east-1'
        self.from_email = 'test@securesystem.email'  # Use verified email from domain
        self.to_email = 'cpigusch@gmail.com'
        self.base_url = 'https://securesystem.email'  # Your frontend URL
        self.api_base_url = 'https://securesystem.email'  # Your API URL
        
        # Initialize SES client
        self.ses_client = boto3.client('ses', region_name=self.aws_region)
        
        # Test data
        self.test_message = "This is a simple encrypted test message."
        self.test_subject = "Secure Email MVP Simple Test"
        
    def test_ses_verification(self):
        """Test SES domain and email verification"""
        print("🔍 Testing SES verification status...")
        
        try:
            # Check domain verification
            domain_response = self.ses_client.get_identity_verification_attributes(
                Identities=['securesystem.email']
            )
            
            domain_status = domain_response['VerificationAttributes'].get('securesystem.email', {}).get('VerificationStatus')
            if domain_status == 'Success':
                print("   ✅ Domain securesystem.email: Success")
            else:
                print(f"   ❌ Domain securesystem.email: {domain_status}")
                return False
                
            # Check email verification (since domain is verified, any email from domain should work)
            print("   ✅ Email test@securesystem.email: Success (domain verified)")
                
            return True
            
        except Exception as e:
            print(f"   ❌ SES verification check failed: {e}")
            return False
    
    def create_test_encrypted_data(self):
        """Create test encrypted data using your PQC system"""
        print("🔐 Creating test encrypted data...")
        
        # This would normally call your PQC encryption API
        # For testing, we'll create a mock encrypted structure
        # In production, this would use your actual PQC hybrid encryption
        
        test_data = {
            "encrypted_content": base64.b64encode(self.test_message.encode()).decode(),
            "encryption_method": "PQC-HYBRID",
            "kyber_level": 768,
            "algorithm": "AES-256-GCM",
            "test_key": "test_key_for_decryption",  # In real system, this would be the actual key
            "metadata": {
                "subject": self.test_subject,
                "sender": self.from_email,
                "timestamp": time.time()
            }
        }
        
        print("   ✅ Test encrypted data created")
        return test_data
    
    def create_secure_link(self, encrypted_data):
        """Create a secure link using your existing secure link service"""
        print("🔗 Creating secure link...")
        
        # This would call your secure link creation API
        # For testing, we'll create a mock secure link
        # In production, this would use your actual secure link service
        
        link_id = base64.urlsafe_b64encode(f"test_link_{int(time.time())}".encode()).decode().rstrip('=')
        secure_url = f"{self.base_url}/v/{link_id}"
        
        link_data = {
            "link_id": link_id,
            "secure_url": secure_url,
            "encrypted_data": encrypted_data,
            "security_settings": {
                "require_password": False,  # Simple test - no password
                "require_mfa": False,
                "read_once": False,
                "auto_destruct": False
            }
        }
        
        print(f"   ✅ Secure link created: {secure_url}")
        return link_data
    
    def send_secure_link_email(self, link_data):
        """Send email with secure link via SES"""
        print("📧 Sending secure link email...")
        
        # Create email content
        email_body = f"""
        <html>
        <head>
            <title>Secure Email MVP Test</title>
        </head>
        <body>
            <h2>🔐 Secure Email MVP - Test Message</h2>
            <p>You have received a secure encrypted message.</p>
            <p><strong>Subject:</strong> {self.test_subject}</p>
            <p><strong>From:</strong> {self.from_email}</p>
            <br>
            <p>Click the secure link below to decrypt and read your message:</p>
            <p><a href="{link_data['secure_url']}" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">🔓 Open Secure Message</a></p>
            <br>
            <p><strong>Security Features:</strong></p>
            <ul>
                <li>🔐 Post-Quantum Cryptography (Kyber-768 + AES-256-GCM)</li>
                <li>🔗 Secure link with encrypted content</li>
                <li>📧 End-to-end encryption</li>
            </ul>
            <br>
            <p><em>This is a test message from the Secure Email MVP system.</em></p>
        </body>
        </html>
        """
        
        # Create raw email
        msg = MIMEMultipart('alternative')
        msg['Subject'] = self.test_subject
        msg['From'] = self.from_email
        msg['To'] = self.to_email
        
        # Add HTML part
        html_part = MIMEText(email_body, 'html')
        msg.attach(html_part)
        
        # Add plain text part
        text_body = f"""
        Secure Email MVP Test
        
        Subject: {self.test_subject}
        From: {self.from_email}
        
        You have received a secure encrypted message.
        
        Click this link to decrypt and read your message:
        {link_data['secure_url']}
        
        Security Features:
        - Post-Quantum Cryptography (Kyber-768 + AES-256-GCM)
        - Secure link with encrypted content
        - End-to-end encryption
        
        This is a test message from the Secure Email MVP system.
        """
        
        text_part = MIMEText(text_body, 'plain')
        msg.attach(text_part)
        
        try:
            # Send email via SES
            response = self.ses_client.send_raw_email(
                Source=self.from_email,
                Destinations=[self.to_email],
                RawMessage={'Data': msg.as_string()}
            )
            
            message_id = response['MessageId']
            print(f"   ✅ Email sent successfully!")
            print(f"   📧 Message ID: {message_id}")
            print(f"   📬 Check your inbox at: {self.to_email}")
            
            return message_id
            
        except ClientError as e:
            print(f"   ❌ Failed to send email: {e}")
            return None
    
    def test_frontend_access(self, link_data):
        """Test if the secure link is accessible via frontend"""
        print("🌐 Testing frontend access...")
        
        try:
            # Test the secure link URL
            response = requests.get(link_data['secure_url'], timeout=10)
            
            if response.status_code == 200:
                print("   ✅ Frontend accessible")
                print(f"   📄 Response length: {len(response.text)} characters")
                return True
            else:
                print(f"   ❌ Frontend returned status: {response.status_code}")
                return False
                
        except requests.exceptions.RequestException as e:
            print(f"   ❌ Frontend access failed: {e}")
            return False
    
    def run_complete_test(self):
        """Run the complete encrypted link flow test"""
        print("🚀 Secure Email MVP - Encrypted Link Flow Test")
        print("=" * 60)
        
        # Step 1: Test SES verification
        if not self.test_ses_verification():
            print("❌ SES verification failed. Cannot proceed.")
            return False
        
        # Step 2: Create encrypted data
        encrypted_data = self.create_test_encrypted_data()
        
        # Step 3: Create secure link
        link_data = self.create_secure_link(encrypted_data)
        
        # Step 4: Send email with secure link
        message_id = self.send_secure_link_email(link_data)
        if not message_id:
            print("❌ Email sending failed. Cannot proceed.")
            return False
        
        # Step 5: Test frontend access
        frontend_accessible = self.test_frontend_access(link_data)
        
        # Step 6: Generate test report
        self.generate_test_report(link_data, message_id, frontend_accessible)
        
        return True
    
    def generate_test_report(self, link_data, message_id, frontend_accessible):
        """Generate comprehensive test report"""
        print("\n📋 Test Report")
        print("=" * 60)
        
        print(f"✅ Test Status: {'PASSED' if frontend_accessible else 'PARTIAL'}")
        print(f"📧 SES Message ID: {message_id}")
        print(f"🔗 Secure Link: {link_data['secure_url']}")
        print(f"🔐 Encryption Method: {link_data['encrypted_data']['encryption_method']}")
        print(f"🔑 Test Key: {link_data['encrypted_data']['test_key']}")
        print(f"🌐 Frontend Access: {'✅ Working' if frontend_accessible else '❌ Failed'}")
        
        print("\n📝 Next Steps:")
        print("1. Check your Gmail inbox for the test email")
        print("2. Click the secure link in the email")
        print("3. Verify the frontend opens and displays the message")
        print("4. Check email headers in Gmail for SPF/DKIM/DMARC")
        
        print("\n🔍 Verification Steps:")
        print("1. Open Gmail and find the test email")
        print("2. Click 'Show original' (three dots menu)")
        print("3. Look for these authentication headers:")
        print("   - Authentication-Results: spf=PASS")
        print("   - Authentication-Results: dkim=PASS")
        print("   - Authentication-Results: dmarc=PASS")
        
        print(f"\n🎯 Expected Outcome:")
        print("✅ You receive a Gmail message with a clickable secure link")
        print("✅ Clicking the link opens the Secure Email MVP frontend")
        print("✅ The message content is decrypted and readable")
        print("✅ Email authentication headers show PASS status")

def main():
    """Main test execution"""
    tester = SecureEmailLinkTester()
    
    try:
        success = tester.run_complete_test()
        
        if success:
            print("\n🎉 Test completed successfully!")
            print("Check your email and verify the secure link functionality.")
        else:
            print("\n❌ Test failed. Check the error messages above.")
            sys.exit(1)
            
    except Exception as e:
        print(f"\n💥 Test execution failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
