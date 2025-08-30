package main

import (
	"fmt"
	"strings"
)

// Simple test functions to verify the signup v2 implementation
func testPasswordHashing() {
	fmt.Println("Testing password hashing...")

	password := "SecurePass123!"
	hashedPassword, err := hashPasswordWithArgon2id(password)
	if err != nil {
		fmt.Printf("❌ Password hashing failed: %v\n", err)
		return
	}

	if !strings.HasPrefix(hashedPassword, "$argon2id$") {
		fmt.Println("❌ Hash format is incorrect")
		return
	}

	if hashedPassword == password {
		fmt.Println("❌ Hash equals original password")
		return
	}

	fmt.Println("✅ Password hashing works correctly")
}

func testPasswordValidation() {
	fmt.Println("Testing password validation...")

	tests := []struct {
		password string
		expected bool
	}{
		{"SecurePass123!", true},
		{"weak", false},
		{"nouppercase123!", false},
		{"NOLOWERCASE123!", false},
		{"NoNumbers!", false},
		{"NoSpecial123", false},
		{"", false},
	}

	for _, test := range tests {
		result := isValidPasswordStrength(test.password)
		if result != test.expected {
			fmt.Printf("❌ Password '%s': expected %v, got %v\n", test.password, test.expected, result)
			return
		}
	}

	fmt.Println("✅ Password validation works correctly")
}

func testPlanValidation() {
	fmt.Println("Testing plan validation...")

	tests := []struct {
		plan     string
		expected bool
	}{
		{"free", true},
		{"paid", true},
		{"company", true},
		{"invalid", false},
		{"", false},
	}

	for _, test := range tests {
		result := isValidPlan(test.plan)
		if result != test.expected {
			fmt.Printf("❌ Plan '%s': expected %v, got %v\n", test.plan, test.expected, result)
			return
		}
	}

	fmt.Println("✅ Plan validation works correctly")
}

func testEmailValidation() {
	fmt.Println("Testing email validation...")

	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user@domain.org", true},
		{"invalid-email", false},
		{"@domain.com", false},
		{"user@", false},
		{"", false},
		{strings.Repeat("a", 255) + "@example.com", false}, // Too long
		{"a@b", false}, // Too short
	}

	for _, test := range tests {
		result := isValidEmailFormat(test.email)
		if result != test.expected {
			fmt.Printf("❌ Email '%s': expected %v, got %v\n", test.email, test.expected, result)
			return
		}
	}

	fmt.Println("✅ Email validation works correctly")
}

func testSignupV2Implementation() {
	fmt.Println("🧪 Testing Signup V2 Implementation")
	fmt.Println("====================================")

	testPasswordHashing()
	testPasswordValidation()
	testPlanValidation()
	testEmailValidation()

	fmt.Println("\n✅ All tests completed!")
}
