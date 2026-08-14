package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLength   = 16
)

// HashIt hashes a plaintext password using argon2id and returns a PHC-formatted string.
// The output format is: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
func HashIt(plaintext string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(plaintext), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	saltB64Len := base64.RawStdEncoding.EncodedLen(len(salt))
	hashB64Len := base64.RawStdEncoding.EncodedLen(len(hash))

	b := make([]byte, 0, 64+saltB64Len+hashB64Len)
	b = append(b, "$argon2id$v="...)
	b = strconv.AppendInt(b, int64(argon2.Version), 10)
	b = append(b, "$m="...)
	b = strconv.AppendInt(b, int64(argonMemory), 10)
	b = append(b, ",t="...)
	b = strconv.AppendInt(b, int64(argonTime), 10)
	b = append(b, ",p="...)
	b = strconv.AppendInt(b, int64(argonThreads), 10)
	b = append(b, '$')

	saltStart := len(b)
	b = b[:saltStart+saltB64Len]
	base64.RawStdEncoding.Encode(b[saltStart:], salt)

	b = append(b, '$')

	hashStart := len(b)
	b = b[:hashStart+hashB64Len]
	base64.RawStdEncoding.Encode(b[hashStart:], hash)

	return string(b), nil
}

// VerifyHash verifies a plaintext password against a PHC-formatted argon2id hash.
// Parameters (m, t, p) are parsed from the stored hash, enabling safe parameter
// migration — old hashes remain verifiable after changing default constants.
func VerifyHash(plaintext, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return false, errors.New("invalid hash algorithm")
	}

	if parts[2] != "v=19" {
		return false, errors.New("invalid hash version")
	}

	memory, time, threads, err := parseParams(parts[3])
	if err != nil {
		return false, fmt.Errorf("invalid hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt: %w", err)
	}
	if len(salt) == 0 {
		return false, errors.New("invalid salt: empty")
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}
	if len(decodedHash) == 0 {
		return false, errors.New("invalid hash: empty")
	}

	hashToVerify := argon2.IDKey(
		[]byte(plaintext),
		salt,
		time,
		memory,
		threads,
		uint32(len(decodedHash)),
	)

	return subtle.ConstantTimeCompare(decodedHash, hashToVerify) == 1, nil
}

// parseParams extracts m, t, p from a string like "m=65536,t=1,p=4".
func parseParams(s string) (memory uint32, time uint32, threads uint8, err error) {
	pairs := strings.Split(s, ",")
	if len(pairs) != 3 {
		return 0, 0, 0, errors.New("expected 3 parameter pairs")
	}

	seen := make(map[string]bool, 3)

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			return 0, 0, 0, fmt.Errorf("malformed parameter: %s", pair)
		}

		if seen[kv[0]] {
			return 0, 0, 0, fmt.Errorf("duplicate parameter: %s", kv[0])
		}
		seen[kv[0]] = true

		val, parseErr := strconv.ParseUint(kv[1], 10, 32)
		if parseErr != nil {
			return 0, 0, 0, fmt.Errorf("invalid value for %s: %w", kv[0], parseErr)
		}

		switch kv[0] {
		case "m":
			memory = uint32(val)
		case "t":
			time = uint32(val)
		case "p":
			if val > 255 {
				return 0, 0, 0, errors.New("threads parameter exceeds uint8 range")
			}
			threads = uint8(val)
		default:
			return 0, 0, 0, fmt.Errorf("unknown parameter: %s", kv[0])
		}
	}

	if memory == 0 || time == 0 || threads == 0 {
		return 0, 0, 0, errors.New("all parameters (m, t, p) must be non-zero")
	}

	if memory < 8*uint32(threads) {
		return 0, 0, 0, errors.New("memory parameter must be at least 8 times threads")
	}

	return memory, time, threads, nil
}
