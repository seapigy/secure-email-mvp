package auth

import (
	"time"

	"github.com/pquerna/otp/totp"
)

// GenerateTOTPCode generates a valid TOTP code for the given secret
// This is used for testing purposes to generate valid TOTP codes
func GenerateTOTPCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

// GenerateTOTPCodeAtTime generates a valid TOTP code for the given secret at a specific time
// This is used for testing purposes to generate valid TOTP codes at specific timestamps
func GenerateTOTPCodeAtTime(secret string, t time.Time) (string, error) {
	return totp.GenerateCode(secret, t)
}

// ValidateTOTPCode validates a TOTP code against a secret
// This is used for testing purposes to validate TOTP codes
func ValidateTOTPCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

