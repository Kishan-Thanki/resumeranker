package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLength   = 16
)

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

	return unsafe.String(unsafe.SliceData(b), len(b)), nil
}

func VerifyHash(plaintext, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	hashToVerify := argon2.IDKey([]byte(plaintext), salt, argonTime, argonMemory, argonThreads, uint32(len(decodedHash)))

	if subtle.ConstantTimeCompare(decodedHash, hashToVerify) == 1 {
		return true, nil
	}
	return false, nil
}
