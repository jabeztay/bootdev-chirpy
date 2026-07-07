package internal

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("hash missing argon2id prefix: %q", hash)
	}
}

func TestHashPasswordUniqueSalt(t *testing.T) {
	// Same input must yield different hashes because the salt is random.
	h1, err := HashPassword("samepass")
	if err != nil {
		t.Fatal(err)
	}
	h2, err := HashPassword("samepass")
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h2 {
		t.Error("expected different hashes for same password (random salt)")
	}
}

func TestCheckPasswordHash_Match(t *testing.T) {
	password := "p@ssw0rd123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash: %v", err)
	}
	if !match {
		t.Error("expected password to match its own hash")
	}
}

func TestCheckPasswordHash_NoMatch(t *testing.T) {
	hash, err := HashPassword("rightpassword")
	if err != nil {
		t.Fatal(err)
	}
	match, err := CheckPasswordHash("wrongpassword", hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Error("expected wrong password not to match")
	}
}

func TestCheckPasswordHash_MalformedHash(t *testing.T) {
	_, err := CheckPasswordHash("password", "not-a-valid-hash")
	if err == nil {
		t.Error("expected error for malformed hash string")
	}
}
