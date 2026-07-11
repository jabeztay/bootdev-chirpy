package internal

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	userId, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")

	if authHeader == "" {
		return "", fmt.Errorf("auth header is missing")
	}

	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return "", fmt.Errorf("auth header must use bearer")
	}

	token := authHeader[len("bearer "):]
	token = strings.TrimSpace(token)

	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	encodedStr := hex.EncodeToString(key)
	return encodedStr
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")

	if authHeader == "" {
		return "", fmt.Errorf("auth header is missing")
	}

	if !strings.HasPrefix(strings.ToLower(authHeader), "apikey ") {
		return "", fmt.Errorf("auth header must use ApiKey")
	}

	token := authHeader[len("ApiKey "):]
	token = strings.TrimSpace(token)

	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}
