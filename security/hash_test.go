package security

import "testing"

func TestHashPassword_And_CheckPasswordHash(t *testing.T) {
	password := "gophersRule123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if len(hash) == 0 {
		t.Fatalf("expected non-empty hash string")
	}

	if !CheckPasswordHash(password, hash) {
		t.Fatalf("expected correct password to match hash, but check failed")
	}

	if CheckPasswordHash("wrongPassword!", hash) {
		t.Fatalf("expected incorrect password to fail check, but it succeeded")
	}
}

func TestHashPassword_DistinctHashes(t *testing.T) {
	password := "samePassword"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("HashPassword failed: %v, %v", err1, err2)
	}

	if hash1 == hash2 {
		t.Fatalf("expected different hashes due to unique salts, got identical hashes")
	}
}
