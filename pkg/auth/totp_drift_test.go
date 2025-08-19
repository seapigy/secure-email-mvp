package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestTOTPDriftTolerance(t *testing.T) {
	// Test configuration with 30-second drift tolerance
	config := TOTPConfig{
		Period:                30,
		Skew:                  1,
		Digits:                6,
		Algorithm:             "SHA1",
		DriftToleranceSeconds: 30,
		MaxDriftSteps:         2,
	}

	// Generate a test secret
	secret := "JBSWY3DPEHPK3PXP" // Test secret for consistent testing

	tests := []struct {
		name        string
		timeOffset  time.Duration
		shouldValid bool
		description string
	}{
		{
			name:        "Current Time (0s)",
			timeOffset:  0,
			shouldValid: true,
			description: "TOTP should validate at current time",
		},
		{
			name:        "30 Seconds Ago (-30s)",
			timeOffset:  -30 * time.Second,
			shouldValid: true,
			description: "TOTP should validate within drift tolerance",
		},
		{
			name:        "30 Seconds Future (+30s)",
			timeOffset:  30 * time.Second,
			shouldValid: true,
			description: "TOTP should validate within drift tolerance",
		},
		{
			name:        "60 Seconds Ago (-60s)",
			timeOffset:  -60 * time.Second,
			shouldValid: false,
			description: "TOTP should be rejected outside drift tolerance",
		},
		{
			name:        "60 Seconds Future (+60s)",
			timeOffset:  60 * time.Second,
			shouldValid: false,
			description: "TOTP should be rejected outside drift tolerance",
		},
		{
			name:        "90 Seconds Ago (-90s)",
			timeOffset:  -90 * time.Second,
			shouldValid: false,
			description: "TOTP should be rejected well outside drift tolerance",
		},
		{
			name:        "90 Seconds Future (+90s)",
			timeOffset:  90 * time.Second,
			shouldValid: false,
			description: "TOTP should be rejected well outside drift tolerance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate TOTP code at the specified time offset
			checkTime := time.Now().Add(tt.timeOffset)
			code, err := totp.GenerateCode(secret, checkTime)
			if err != nil {
				t.Fatalf("Failed to generate TOTP code: %v", err)
			}

			// Validate the code using our drift-tolerant function
			valid := validateTOTPWithConfig(code, secret, config)

			if valid != tt.shouldValid {
				t.Errorf("%s: Expected validation to be %t, got %t", tt.description, tt.shouldValid, valid)
			} else {
				t.Logf("✓ %s: Validation result %t (expected %t)", tt.description, valid, tt.shouldValid)
			}
		})
	}
}

func TestTOTPDriftToleranceEdgeCases(t *testing.T) {
	// Test configuration with different drift tolerance settings
	configs := []struct {
		name   string
		config TOTPConfig
	}{
		{
			name: "Default 30s tolerance",
			config: TOTPConfig{
				Period:                30,
				DriftToleranceSeconds: 30,
				MaxDriftSteps:         2,
				Digits:                6,
				Algorithm:             "SHA1",
			},
		},
		{
			name: "Strict 15s tolerance",
			config: TOTPConfig{
				Period:                30,
				DriftToleranceSeconds: 15,
				MaxDriftSteps:         1,
				Digits:                6,
				Algorithm:             "SHA1",
			},
		},
		{
			name: "Loose 60s tolerance",
			config: TOTPConfig{
				Period:                30,
				DriftToleranceSeconds: 60,
				MaxDriftSteps:         4,
				Digits:                6,
				Algorithm:             "SHA1",
			},
		},
	}

	secret := "JBSWY3DPEHPK3PXP"

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			// Test edge cases around the tolerance boundary
			boundaryOffset := time.Duration(cfg.config.DriftToleranceSeconds) * time.Second

			testCases := []struct {
				offset      time.Duration
				shouldValid bool
				description string
			}{
				{
					offset:      boundaryOffset - time.Second,
					shouldValid: true,
					description: "Just inside tolerance boundary",
				},
				{
					offset:      boundaryOffset,
					shouldValid: false,
					description: "At tolerance boundary (should fail)",
				},
				{
					offset:      boundaryOffset + time.Second,
					shouldValid: false,
					description: "Just outside tolerance boundary",
				},
			}

			for _, tc := range testCases {
				// Test both positive and negative offsets
				for _, sign := range []int{1, -1} {
					actualOffset := time.Duration(sign) * tc.offset
					checkTime := time.Now().Add(actualOffset)

					code, err := totp.GenerateCode(secret, checkTime)
					if err != nil {
						t.Fatalf("Failed to generate TOTP code: %v", err)
					}

					valid := validateTOTPWithConfig(code, secret, cfg.config)

					if valid != tc.shouldValid {
						t.Errorf("Offset %v: Expected validation to be %t, got %t (%s)",
							actualOffset, tc.shouldValid, valid, tc.description)
					} else {
						t.Logf("✓ Offset %v: %s - Validation result %t", actualOffset, tc.description, valid)
					}
				}
			}
		})
	}
}

func TestTOTPDriftToleranceWithSystemClockSkew(t *testing.T) {
	// Test that our drift tolerance handles system clock skew scenarios
	config := TOTPConfig{
		Period:                30,
		DriftToleranceSeconds: 30,
		MaxDriftSteps:         2,
		Digits:                6,
		Algorithm:             "SHA1",
	}

	secret := "JBSWY3DPEHPK3PXP"

	// Simulate different clock skew scenarios
	skewScenarios := []struct {
		name        string
		clockSkew   time.Duration
		shouldValid bool
	}{
		{
			name:        "No clock skew",
			clockSkew:   0,
			shouldValid: true,
		},
		{
			name:        "5 seconds fast",
			clockSkew:   5 * time.Second,
			shouldValid: true,
		},
		{
			name:        "5 seconds slow",
			clockSkew:   -5 * time.Second,
			shouldValid: true,
		},
		{
			name:        "25 seconds fast",
			clockSkew:   25 * time.Second,
			shouldValid: true,
		},
		{
			name:        "25 seconds slow",
			clockSkew:   -25 * time.Second,
			shouldValid: true,
		},
		{
			name:        "35 seconds fast (outside tolerance)",
			clockSkew:   35 * time.Second,
			shouldValid: false,
		},
		{
			name:        "35 seconds slow (outside tolerance)",
			clockSkew:   -35 * time.Second,
			shouldValid: false,
		},
	}

	for _, scenario := range skewScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Generate code at current time
			code, err := totp.GenerateCode(secret, time.Now())
			if err != nil {
				t.Fatalf("Failed to generate TOTP code: %v", err)
			}

			// Simulate clock skew by adjusting the validation time
			skewedTime := time.Now().Add(scenario.clockSkew)

			// Create a modified config that simulates the skewed time
			// We'll modify the validation function to use the skewed time
			valid := validateTOTPWithConfigSkewed(code, secret, config, skewedTime)

			if valid != scenario.shouldValid {
				t.Errorf("Clock skew %v: Expected validation to be %t, got %t",
					scenario.clockSkew, scenario.shouldValid, valid)
			} else {
				t.Logf("✓ Clock skew %v: Validation result %t (expected %t)",
					scenario.clockSkew, valid, scenario.shouldValid)
			}
		})
	}
}

// validateTOTPWithConfigSkewed is a test helper that simulates clock skew
func validateTOTPWithConfigSkewed(code, secret string, config TOTPConfig, skewedTime time.Time) bool {
	// Try validation with RFC 6238 compliant drift tolerance
	// Check current time step and surrounding steps within drift tolerance
	for step := -config.MaxDriftSteps; step <= config.MaxDriftSteps; step++ {
		// Calculate time for this step using the skewed time
		checkTime := skewedTime.Add(time.Duration(step*int(config.Period)) * time.Second)

		// Validate TOTP with custom parameters for this time step
		valid, err := totp.ValidateCustom(code, secret, checkTime, totp.ValidateOpts{
			Period:    config.Period,
			Skew:      0, // No additional skew since we're manually checking steps
			Digits:    6, // Default to 6 digits
			Algorithm: 1, // Default to SHA1 for compatibility
		})

		if err != nil {
			continue
		}

		if valid {
			return true
		}
	}

	return false
}

func TestTOTPDriftTolerancePerformance(t *testing.T) {
	// Test that drift tolerance doesn't significantly impact performance
	config := TOTPConfig{
		Period:                30,
		DriftToleranceSeconds: 30,
		MaxDriftSteps:         2,
		Digits:                6,
		Algorithm:             "SHA1",
	}

	secret := "JBSWY3DPEHPK3PXP"
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("Failed to generate TOTP code: %v", err)
	}

	// Benchmark the validation performance
	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		validateTOTPWithConfig(code, secret, config)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)

	t.Logf("Performance test: %d iterations in %v (avg: %v per validation)",
		iterations, duration, avgDuration)

	// Ensure average validation time is under 1ms
	if avgDuration > time.Millisecond {
		t.Errorf("Average validation time %v exceeds 1ms threshold", avgDuration)
	} else {
		t.Logf("✓ Performance test passed: average validation time %v", avgDuration)
	}
}
