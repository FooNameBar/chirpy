package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/FooNameBar/chirpy/internal/auth"
	"github.com/FooNameBar/chirpy/internal/database"
)

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refTokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Something went wrong getting bearer token %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refToken, err := cfg.db.GetRefreshToken(r.Context(), refTokenStr)
	if err != nil {
		log.Printf("Something went wrong getting refresh token: %v, %v\n", refTokenStr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if time.Now().After(refToken.ExpiresAt) || (refToken.RevokedAt.Valid && time.Now().After(refToken.RevokedAt.Time)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refTokenStr)
	if err != nil {
		log.Printf("Something went wrong getting user by refresh token %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		log.Printf("Something went wrong making jwt token %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := struct {
		Token string `json:"token"`
	}{Token: token}

	resData, err := json.Marshal(res)
	if err != nil {
		log.Printf("Something went wrong making marshaling json %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resData)
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	refTokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Something went wrong getting bearer token %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refToken, err := cfg.db.GetRefreshToken(r.Context(), refTokenStr)
	if err != nil {
		log.Printf("Something went wrong getting refresh token %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = cfg.db.RevokeToken(r.Context(), database.RevokeTokenParams{
		UpdatedAt: time.Now(),
		RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Token:     refToken.Token,
	})

	if err != nil {
		log.Printf("Something went wrong revoking refresh token %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
