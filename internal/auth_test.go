package internal

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "correct-secret"

	// A valid, non-expired token we can reuse.
	validToken, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed during setup: %v", err)
	}

	// An already-expired token: negative duration puts exp in the past.
	expiredToken, err := MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed during setup: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			tokenSecret: secret,
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "expired token",
			tokenString: expiredToken,
			tokenSecret: secret,
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong-secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "malformed token string",
			tokenString: "not.a.jwt",
			tokenSecret: secret,
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() userID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}
