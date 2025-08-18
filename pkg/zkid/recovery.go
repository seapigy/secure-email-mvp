package zkid

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/argon2"
)

// GenerateRecoveryCodes creates N one-time recovery codes and stores their hashes
func (s *Service) GenerateRecoveryCodes(userID string, n int) ([]string, error) {
	if !s.config.Enabled {
		return nil, errors.New("zkid disabled")
	}
	if n <= 0 {
		n = 10
	}
	codes := make([]string, 0, n)
	for i := 0; i < n; i++ {
		raw := make([]byte, 16)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		code := hex.EncodeToString(raw)
		salt := make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return nil, err
		}
		hash := argon2.IDKey([]byte(code), append(salt, s.config.RecoveryPepper...), 3, 64*1024, 2, 32)
		id := fmt.Sprintf("rc_%d", time.Now().UnixNano())
		if _, err := s.db.Exec(`INSERT INTO zkid_recovery_codes (id,user_id,salt,hash,used,created_at) VALUES (?,?,?,?,0,CURRENT_TIMESTAMP)`, id, userID, salt, hash); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, nil
}

// ValidateAndConsumeRecoveryCode validates a code for a user and marks it used atomically
func (s *Service) ValidateAndConsumeRecoveryCode(userID, code string) (bool, error) {
	if !s.config.Enabled {
		return false, errors.New("zkid disabled")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(`SELECT id,salt,hash FROM zkid_recovery_codes WHERE user_id=? AND used=0`, userID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var salt, hash []byte
		if err := rows.Scan(&id, &salt, &hash); err != nil {
			return false, err
		}
		computed := argon2.IDKey([]byte(code), append(salt, s.config.RecoveryPepper...), 3, 64*1024, 2, 32)
		if len(computed) == len(hash) {
			equal := true
			for i := range computed {
				if computed[i] != hash[i] {
					equal = false
					break
				}
			}
			if equal {
				if _, err := tx.Exec(`UPDATE zkid_recovery_codes SET used=1, used_at=CURRENT_TIMESTAMP WHERE id=?`, id); err != nil {
					return false, err
				}
				if err := tx.Commit(); err != nil {
					return false, err
				}
				return true, nil
			}
		}
	}
	return false, nil
}
