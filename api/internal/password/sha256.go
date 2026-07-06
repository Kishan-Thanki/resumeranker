package password

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashSHA256(plaintext string) string {
	hash := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(hash[:])
}
