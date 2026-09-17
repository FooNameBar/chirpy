package auth

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
	tNow := time.Now().UTC()
	tExp := tNow.Add(expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(tNow),
		ExpiresAt: jwt.NewNumericDate(tExp),
		Subject:   userID.String(),
	})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("wrong signing method")
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("jwt.ParseWithClaims: %v\n", err)
	}
	uuidStr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("token.Claims.GetSubject: %v\n", err)
	}
	userId, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("uuid.Parse: %v\n", err)
	}
	return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenStr := headers.Get("Authorization")
	if tokenStr == ""{
		return "", fmt.Errorf("No token in headers")
	}

	return strings.TrimSpace(strings.Trim(tokenStr, "Bearer")), nil
}

func MakeRefreshToken() string {
	data := make([]byte, 32)
	n, err := rand.Read(data)
	if err != nil || n != 32 {
	}

	return hex.EncodeToString(data)
}
