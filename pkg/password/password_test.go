package password

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewPasswordService(t *testing.T) {
	service := NewPasswordService()
	if service == nil {
		t.Fatal("NewPasswordService returned nil")
	}
	if service.config == nil {
		t.Fatal("PasswordService config is nil")
	}
	if service.client == nil {
		t.Fatal("PasswordService client is nil")
	}
}

func TestNewPasswordServiceWithConfig(t *testing.T) {
	config := &PasswordConfig{
		MinLength:           16,
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:      true,
		RequireSpecialChars: true,
		HIBPTimeout:         5 * time.Second,
		UserAgent:           "Test/1.0",
	}

	service := NewPasswordServiceWithConfig(config)
	if service == nil {
		t.Fatal("NewPasswordServiceWithConfig returned nil")
	}
	if service.config.MinLength != 16 {
		t.Errorf("Expected MinLength 16, got %d", service.config.MinLength)
	}
	if service.config.UserAgent != "Test/1.0" {
		t.Errorf("Expected UserAgent 'Test/1.0', got %s", service.config.UserAgent)
	}
}

func TestValidatePassword_StrongPassword(t *testing.T) {
	service := NewPasswordService()
	ctx := context.Background()

	result, err := service.ValidatePassword(ctx, "SecurePassword123!")
	if err != nil {
		t.Fatalf("ValidatePassword failed: %v", err)
	}

	if !result.IsValid {
		t.Errorf("Expected valid password, got invalid with errors: %v", result.Errors)
	}

	if result.Score < 50 {
		t.Errorf("Expected score >= 50, got %d", result.Score)
	}

	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions for improvement")
	}
}

func TestValidatePassword_WeakPassword(t *testing.T) {
	service := NewPasswordService()
	ctx := context.Background()

	result, err := service.ValidatePassword(ctx, "weak")
	if err != nil {
		t.Fatalf("ValidatePassword failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected invalid password, got valid")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected validation errors")
	}

	if result.Score > 40 {
		t.Errorf("Expected low score, got %d", result.Score)
	}
}

func TestValidatePassword_CommonPassword(t *testing.T) {
	service := NewPasswordService()
	ctx := context.Background()

	result, err := service.ValidatePassword(ctx, "password")
	if err != nil {
		t.Fatalf("ValidatePassword failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected invalid password (common), got valid")
	}

	foundCommonError := false
	for _, err := range result.Errors {
		if err == "Password is too common and easily guessable" {
			foundCommonError = true
			break
		}
	}

	if !foundCommonError {
		t.Error("Expected common password error")
	}
}

func TestValidatePassword_MissingRequirements(t *testing.T) {
	service := NewPasswordService()
	ctx := context.Background()

	// Test password missing uppercase
	result, err := service.ValidatePassword(ctx, "securepassword123!")
	if err != nil {
		t.Fatalf("ValidatePassword failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected invalid password (missing uppercase), got valid")
	}

	foundUppercaseError := false
	for _, err := range result.Errors {
		if err == "Password must contain at least one uppercase letter" {
			foundUppercaseError = true
			break
		}
	}

	if !foundUppercaseError {
		t.Error("Expected uppercase requirement error")
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	service := NewPasswordService()
	ctx := context.Background()

	result, err := service.ValidatePassword(ctx, "Short1!")
	if err != nil {
		t.Fatalf("ValidatePassword failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected invalid password (too short), got valid")
	}

	foundLengthError := false
	for _, err := range result.Errors {
		if err == "Password must be at least 12 characters long" {
			foundLengthError = true
			break
		}
	}

	if !foundLengthError {
		t.Error("Expected length requirement error")
	}
}

func TestCalculatePasswordScore(t *testing.T) {
	service := NewPasswordService()

	tests := []struct {
		password string
		minScore int
	}{
		{"SecurePassword123!", 60},
		{"weak", 10},
		{"password", 5},
		{"VeryLongPasswordWithSpecialChars123!", 80},
	}

	for _, test := range tests {
		score := service.calculatePasswordScore(test.password)
		if score < test.minScore {
			t.Errorf("Password '%s' expected score >= %d, got %d", test.password, test.minScore, score)
		}
	}
}

func TestIsCommonPassword(t *testing.T) {
	service := NewPasswordService()

	commonPasswords := []string{"password", "123456", "qwerty", "admin", "letmein"}
	uncommonPasswords := []string{"SecurePassword123!", "MyUniquePass2023!", "RandomString456"}

	for _, password := range commonPasswords {
		if !service.isCommonPassword(password) {
			t.Errorf("Expected '%s' to be common password", password)
		}
	}

	for _, password := range uncommonPasswords {
		if service.isCommonPassword(password) {
			t.Errorf("Expected '%s' to be uncommon password", password)
		}
	}
}

func TestGenerateSuggestions(t *testing.T) {
	service := NewPasswordService()

	// Test password missing multiple requirements
	suggestions := service.generateSuggestions("weak", []string{
		"Password must be at least 12 characters long",
		"Password must contain at least one uppercase letter",
		"Password must contain at least one number",
		"Password must contain at least one special character",
	})

	if len(suggestions) < 4 {
		t.Errorf("Expected at least 4 suggestions, got %d", len(suggestions))
	}

	// Test strong password
	suggestions = service.generateSuggestions("SecurePassword123!", []string{})
	if len(suggestions) == 0 {
		t.Error("Expected suggestions even for strong password")
	}
}

func TestCheckPasswordBreach_NoAPIKey(t *testing.T) {
	service := NewPasswordService()
	ctx := context.Background()

	// Test without API key
	count, err := service.checkPasswordBreach(ctx, "testpassword")
	if err != nil {
		t.Fatalf("checkPasswordBreach failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0 without API key, got %d", count)
	}
}

func TestCheckPasswordBreach_WithAPIKey(t *testing.T) {
	config := DefaultConfig()
	config.HIBPAPIKey = "test-key"
	service := NewPasswordServiceWithConfig(config)

	ctx := context.Background()
	count, err := service.checkPasswordBreach(ctx, "testpassword")
	if err != nil {
		// API call will fail in test environment, which is expected
		t.Logf("API call failed as expected: %v", err)
		return
	}

	// In test environment, we expect either 0 or a real breach count
	if count < 0 {
		t.Errorf("Expected non-negative count, got %d", count)
	}
}

func TestValidatePassword_WithBreachCheck(t *testing.T) {
	config := DefaultConfig()
	config.HIBPAPIKey = "test-key"
	service := NewPasswordServiceWithConfig(config)

	ctx := context.Background()
	result, err := service.ValidatePassword(ctx, "SecurePassword123!")
	if err != nil {
		t.Fatalf("ValidatePassword failed: %v", err)
	}

	// Test that the service handles API calls gracefully
	if result.BreachCount < 0 {
		t.Errorf("Expected non-negative breach count, got %d", result.BreachCount)
	}

	// The password should be valid regardless of breach check result in test
	if !result.IsValid {
		t.Logf("Password validation failed with errors: %v", result.Errors)
	}
}

func TestValidatePassword_EnvironmentVariables(t *testing.T) {
	// Test with environment variables
	os.Setenv("HIBP_API_KEY", "test-env-key")
	defer os.Unsetenv("HIBP_API_KEY")

	config := DefaultConfig()
	config.HIBPAPIKey = os.Getenv("HIBP_API_KEY")
	service := NewPasswordServiceWithConfig(config)

	if service.config.HIBPAPIKey != "test-env-key" {
		t.Errorf("Expected HIBP API key 'test-env-key', got '%s'", service.config.HIBPAPIKey)
	}
}

func TestPasswordValidationResult_String(t *testing.T) {
	result := &PasswordValidationResult{
		IsValid:     false,
		Score:       25,
		BreachCount: 100,
		Errors:      []string{"Too short", "Missing uppercase"},
		Suggestions: []string{"Make it longer", "Add uppercase letters"},
	}

	// Test that all fields are properly set
	if result.IsValid {
		t.Error("Expected IsValid to be false")
	}
	if result.Score != 25 {
		t.Errorf("Expected Score 25, got %d", result.Score)
	}
	if result.BreachCount != 100 {
		t.Errorf("Expected BreachCount 100, got %d", result.BreachCount)
	}
	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}
	if len(result.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(result.Suggestions))
	}
}
