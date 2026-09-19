package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRoundTrip(t *testing.T) {
	id := uuid.New()
	secret := "This is a secret"
	token, err := MakeJWT(id, secret, time.Second*30)
	if err != nil {
		t.Fatalf("MakeJWT: %v\n", err)
	}

	returnedId, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT: %v\n", err)
	}

	if id != returnedId {
		t.Fatalf("Returned\t%v\nOriginal\t%v\n", id, returnedId)
	}
}

func TestExpiredToken(t *testing.T) {
	id := uuid.New()
	secret := "This is a secret"
	token, err := MakeJWT(id, secret, time.Second*-5)
	if err != nil {
		t.Fatalf("MakeJWT: %v\n", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatalf("Validated an expired token\n")
	}
}

func TestSecretMatches(t *testing.T) {
	id := uuid.New()
	secret := "This is a secret"
	token, err := MakeJWT(id, secret, time.Second*30)
	if err != nil {
		t.Fatalf("MakeJWT: %v\n", err)
	}

	_, err = ValidateJWT(token, "Wrong secret")
	if err == nil {
		t.Fatalf("Validated with non matching secrets\n")
	}
}

func TestGetBearerToken(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v\n", err)
	}
	header := req.Header

	token, err := MakeJWT(uuid.New(), "This is a secret", time.Hour*1)
	if err != nil {
		t.Fatalf("MakeJWT: %v\n", err)
	}
	header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	returnedStr, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("GetBearerToken: %v\n", err)
	}

	if token != returnedStr {
		t.Fatalf("\nOriginal:\t%s\nReturned:\t%s\n", token, returnedStr)
	}
}

func TestGetBearerTokenRefresh(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v\n", err)
	}
	header := req.Header

	refToken := MakeRefreshToken()
	header.Add("Authorization", fmt.Sprintf("Bearer %s", refToken))

	returnedStr, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("GetBearerToken: %v\n", err)
	}

	if refToken != returnedStr {
		t.Fatalf("\nOriginal:\t%s\nReturned:\t%s\n", refToken, returnedStr)
	}
}

func TestGetAPIKey(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v\n", err)
	}
	header := req.Header

	key := "f271c81ff7084ee5b99a5091b42d486e"
	header.Add("Authorization", fmt.Sprintf("ApiKey %s", key))

	returnedStr, err := GetAPIKey(header)
	if err != nil {
		t.Fatalf("GetAPIKey: %v\n", err)
	}

	if key != returnedStr {
		t.Fatalf("\nOriginal:\t%s\nReturned:\t%s\n", key, returnedStr)
	}
}
