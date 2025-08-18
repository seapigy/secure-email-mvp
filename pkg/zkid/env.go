package zkid

import (
	"encoding/hex"
	"os"
	"strconv"
)

// ConfigFromEnv loads ZKID config from environment variables
// ZKID_ENABLED (bool), ZKID_MASTER_KEY (hex 64), ZKID_EMAIL_HASH_PEPPER (hex), ZKID_RECOVERY_PEPPER (hex)
func ConfigFromEnv() *Config {
	enabled, _ := strconv.ParseBool(os.Getenv("ZKID_ENABLED"))
	cfg := &Config{Enabled: enabled}
	if !enabled {
		return cfg
	}
	if mk := os.Getenv("ZKID_MASTER_KEY"); mk != "" {
		if b, err := hex.DecodeString(mk); err == nil && len(b) == 32 {
			cfg.MasterKey = b
		}
	}
	if pep := os.Getenv("ZKID_EMAIL_HASH_PEPPER"); pep != "" {
		if b, err := hex.DecodeString(pep); err == nil {
			cfg.EmailHashPepper = b
		}
	}
	if rpep := os.Getenv("ZKID_RECOVERY_PEPPER"); rpep != "" {
		if b, err := hex.DecodeString(rpep); err == nil {
			cfg.RecoveryPepper = b
		}
	}
	return cfg
}
