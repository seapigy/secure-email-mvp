package auth

// DO NOT EDIT EXISTING CODE - new file added
// Email service for sending verification codes and recovery keys

import (
	"fmt"
	"log"
)

// SendVerificationEmail sends a verification code to the user's email
func SendVerificationEmail(email, code, username string) error {
	subject := "Verify Your SecureMail Account - Recovery Key Coming"
	body := fmt.Sprintf(`
Hello %s,

Welcome to SecureMail! Please verify your external email address to complete your account setup.

Your verification code is: %s

This code will expire in 30 minutes.

⚠️ CRITICAL SECURITY NOTICE:
Once you verify this email, you will receive your RECOVERY KEY via email. This recovery key is EXTREMELY IMPORTANT and must be kept secure:

• Store it in a secure password manager
• Write it down and store it in a safe place
• Never share it with anyone
• You will need it to recover your account if you lose access
• This is the ONLY time you will receive this key

If you don't verify this email, your account will remain unverified and you won't receive your recovery key.

If you didn't create this account, please ignore this email.

Best regards,
The SecureMail Team
`, username, code)

	// TODO: Integrate with actual email service (SendGrid, AWS SES, etc.)
	// For now, log the email content for development
	log.Printf("EMAIL_VERIFICATION_TO=%s SUBJECT=%s BODY=%s", email, subject, body)

	return nil
}

// SendRecoveryKeyEmail sends the recovery key to the user's verified email
func SendRecoveryKeyEmail(email, recoveryKey, username string) error {
	subject := "Your SecureMail Recovery Key"
	body := fmt.Sprintf(`
Hello %s,

Your email has been verified! Here is your recovery key for your SecureMail account:

RECOVERY KEY: %s

⚠️ IMPORTANT SECURITY INFORMATION:
- Store this recovery key in a secure location
- You will need this key along with your external email to recover your account
- Never share this key with anyone
- This is the only time you will receive this key

Your account details:
- Username: %s
- External Email: %s
- Primary Email: %s@securesystem.email

If you didn't request this, please contact support immediately.

Best regards,
The SecureMail Team
`, username, recoveryKey, username, email, username)

	// TODO: Integrate with actual email service (SendGrid, AWS SES, etc.)
	// For now, log the email content for development
	log.Printf("EMAIL_RECOVERY_KEY_TO=%s SUBJECT=%s BODY=%s", email, subject, body)

	return nil
}
