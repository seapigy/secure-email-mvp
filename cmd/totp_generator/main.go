package main

import (
	"encoding/base32"
	"fmt"
	"os"
	"time"

	"github.com/pquerna/otp/totp"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: totp_generator <base32_secret>")
		fmt.Println("Example: totp_generator JBSWY3DPEHPK3PXP")
		os.Exit(1)
	}

	secret := os.Args[1]

	// Validate that the secret is valid base32
	_, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		fmt.Printf("Error: Invalid base32 secret: %v\n", err)
		os.Exit(1)
	}

	// Generate TOTP code
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		fmt.Printf("Error generating TOTP code: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(code)
}












