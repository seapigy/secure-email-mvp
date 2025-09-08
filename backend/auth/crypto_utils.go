package auth

// DO NOT EDIT EXISTING CODE - new file added
// Crypto helper utilities: Argon2id password hashing + token helpers.

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id params - tune for production (these are conservative; increase mem/time for real production)
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

// HashPassword returns encoded Argon2id hash in the format:
// $argon2id$v=19$m=65536,t=3,p=4$<salt_b64>$<hash_b64>
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password empty")
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Hash)
	return encoded, nil
}

// VerifyPassword compares plaintext password with encoded hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	// encodedHash format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}
	// parts[3] => params like v=19$m=65536,t=3,p=4 (we ignore parsing v)
	params := parts[3]
	saltB64 := parts[4]
	hashB64 := parts[5]

	// parse params: extract memory and time and threads. We'll fallback to defaults on parse error.
	var memory uint32 = argonMemory
	var timeParam uint32 = argonTime
	var threads uint8 = argonThreads

	// simple param parse
	// params like m=65536,t=3,p=4
	for _, p := range strings.Split(params, ",") {
		if strings.HasPrefix(p, "m=") {
			var m uint32
			fmt.Sscanf(p, "m=%d", &m)
			if m > 0 {
				memory = m
			}
		} else if strings.HasPrefix(p, "t=") {
			var t uint32
			fmt.Sscanf(p, "t=%d", &t)
			if t > 0 {
				timeParam = t
			}
		} else if strings.HasPrefix(p, "p=") {
			var pth uint32
			fmt.Sscanf(p, "p=%d", &pth)
			if pth > 0 {
				threads = uint8(pth)
			}
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, err
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, err
	}

	hash := argon2.IDKey([]byte(password), salt, timeParam, memory, threads, uint32(len(expectedHash)))
	if len(hash) != len(expectedHash) {
		return false, nil
	}
	// constant time compare
	diff := 0
	for i := 0; i < len(hash); i++ {
		diff |= int(hash[i] ^ expectedHash[i])
	}
	return diff == 0, nil
}

// GenerateRandomToken returns base64 raw token of n bytes.
func GenerateRandomToken(n int) (string, error) {
	if n <= 0 {
		n = 32
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(b), nil
}

// HashToken returns the SHA256 hex of the provided token for storage.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
