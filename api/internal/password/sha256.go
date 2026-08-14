package password

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashSHA256 produces a hex-encoded SHA-256 digest of the input.
// Suitable for API key verification, content checksums, etc.
// NOT suitable for password hashing — use HashIt (argon2id) instead.
func HashSHA256(plaintext string) string {
	hash := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(hash[:])
}
