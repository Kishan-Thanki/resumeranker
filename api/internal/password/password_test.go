package password_test

import (
	"strings"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/password"
)

func TestHashItAndVerifyHash(t *testing.T) {
	t.Parallel()

	t.Run("successful hash and verify", func(t *testing.T) {
		plain := "mySuperSecretPassword123!"

		hash1, err := password.HashIt(plain)
		if err != nil {
			t.Fatalf("unexpected error hashing: %v", err)
		}

		if hash1 == "" {
			t.Fatal("expected a hash, got empty string")
		}

		match, err := password.VerifyHash(plain, hash1)
		if err != nil {
			t.Fatalf("unexpected error verifying: %v", err)
		}
		if !match {
			t.Fatal("expected match to be true")
		}
	})

	t.Run("different salts produce different hashes", func(t *testing.T) {
		plain := "password"

		hash1, err := password.HashIt(plain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hash2, err := password.HashIt(plain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hash1 == hash2 {
			t.Fatal("expected hashes to be different due to random salt")
		}
	})

	t.Run("incorrect password fails", func(t *testing.T) {
		plain := "correctPassword"
		wrongPlain := "CorrectPassword"

		hash, _ := password.HashIt(plain)

		match, err := password.VerifyHash(wrongPlain, hash)
		if err != nil {
			t.Fatalf("unexpected error verifying: %v", err)
		}
		if match {
			t.Fatal("expected match to be false for wrong password")
		}
	})

	t.Run("empty password hashes and verifies", func(t *testing.T) {
		hash, err := password.HashIt("")
		if err != nil {
			t.Fatalf("unexpected error hashing empty password: %v", err)
		}
		if hash == "" {
			t.Fatal("expected a hash, got empty string")
		}

		match, err := password.VerifyHash("", hash)
		if err != nil {
			t.Fatalf("unexpected error verifying empty password: %v", err)
		}
		if !match {
			t.Fatal("expected match to be true for empty password")
		}

		match, err = password.VerifyHash("notempty", hash)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if match {
			t.Fatal("expected match to be false for non-empty password against empty hash")
		}
	})

	t.Run("malformed hash returns safe error", func(t *testing.T) {
		plain := "password"
		malformedHashes := []string{
			"invalid-format",
			"$argon2id$v=19$m=65536,t=1,p=4$invalidbase64salt$invalidbase64hash",
			"$argon2id$v=19$m=0,t=1,p=4$c2FsdA$aGFzaA",
			"$argon2id$v=19$m=abc,t=1,p=4$c2FsdA$aGFzaA",
			"$argon2i$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA",
			"$argon2id$v=16$m=65536,t=1,p=4$c2FsdA$aGFzaA",
			"$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$",
			"$argon2id$v=19$m=65536,t=1,p=4$$aGFzaA",
			"$argon2id$v=19$m=65536,t=1,p=256$c2FsdA$aGFzaA",
			"$argon2id$v=19$m=16,t=1,p=4$c2FsdA$aGFzaA",
			"$argon2id$v=19$m=65536,t=1,p=4,m=65536$c2FsdA$aGFzaA",
		}

		for _, malformed := range malformedHashes {
			match, err := password.VerifyHash(plain, malformed)
			if match {
				t.Errorf("expected match to be false for malformed hash: %s", malformed)
			}
			if err == nil {
				t.Errorf("expected error for malformed hash: %s", malformed)
			}
		}
	})
}

func TestHashSHA256(t *testing.T) {
	t.Parallel()

	t.Run("consistent hashing", func(t *testing.T) {
		plain := "hello world"
		expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

		result := password.HashSHA256(plain)
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}

		if result != strings.ToLower(result) {
			t.Error("expected lowercase hex string")
		}
	})
}

func BenchmarkHashIt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = password.HashIt("benchmarkpassword")
	}
}

func BenchmarkVerifyHash(b *testing.B) {
	b.ReportAllocs()
	hash, _ := password.HashIt("benchmarkpassword")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = password.VerifyHash("benchmarkpassword", hash)
	}
}
