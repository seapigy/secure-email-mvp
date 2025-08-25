package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pquerna/otp/totp"
)

func main() {
	secret := "JBSWY3DPEHPK3PXP"
	now := time.Now()

	fmt.Printf("🔐 TOTP Generation Test\n")
	fmt.Printf("======================\n")
	fmt.Printf("Secret: %s\n", secret)
	fmt.Printf("Current time: %v\n", now)
	fmt.Printf("Current time (UTC): %v\n", now.UTC())

	// Generate current TOTP code
	currentCode, err := totp.GenerateCode(secret, now)
	if err != nil {
		log.Fatalf("Failed to generate TOTP code: %v", err)
	}

	fmt.Printf("\n🎯 Generated TOTP Code: %s\n", currentCode)

	// Test validation
	valid, err := totp.ValidateCustom(currentCode, secret, now, totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    6,
		Algorithm: 1, // SHA1
	})

	if err != nil {
		log.Printf("Validation error: %v", err)
	} else {
		fmt.Printf("✅ Validation result: %t\n", valid)
	}

	// Test with our test code
	testCode := "216705"
	valid2, err := totp.ValidateCustom(testCode, secret, now, totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    6,
		Algorithm: 1, // SHA1
	})

	if err != nil {
		log.Printf("Test code validation error: %v", err)
	} else {
		fmt.Printf("🧪 Test code '%s' validation: %t\n", testCode, valid2)
	}

	// Test with different time steps
	fmt.Printf("\n⏰ Testing different time steps:\n")
	for step := -2; step <= 2; step++ {
		testTime := now.Add(time.Duration(step*30) * time.Second)
		code, _ := totp.GenerateCode(secret, testTime)
		valid, _ := totp.ValidateCustom(testCode, secret, testTime, totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    6,
			Algorithm: 1,
		})
		fmt.Printf("  Step %+d: Code=%s, Valid=%t\n", step, code, valid)
	}
}
