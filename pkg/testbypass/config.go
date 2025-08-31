package testbypass

import (
	"os"
	"strconv"
)

// Config holds test bypass configuration
type Config struct {
	Enabled      bool
	TestEmail    string
	TestPassword string
	TestUserID   string
}

// LoadConfig loads test bypass configuration from environment variables
func LoadConfig() *Config {
	enabled := false
	if envVal := os.Getenv("TEST_BYPASS"); envVal != "" {
		if val, err := strconv.ParseBool(envVal); err == nil {
			enabled = val
		}
	}

	return &Config{
		Enabled:      enabled,
		TestEmail:    "test@example.com",
		TestPassword: "Test1234!",
		TestUserID:   "test-user-12345",
	}
}

// IsTestUser checks if the given email is the test user
func (c *Config) IsTestUser(email string) bool {
	return c.Enabled && email == c.TestEmail
}

// IsTestUserID checks if the given user ID is the test user
func (c *Config) IsTestUserID(userID string) bool {
	return c.Enabled && userID == c.TestUserID
}
