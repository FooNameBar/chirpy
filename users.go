package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/FooNameBar/chirpy/internal/auth"
	"github.com/FooNameBar/chirpy/internal/database"
)

type userAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)

	var uReq userAuth
	err := decoder.Decode(&uReq)
	if err != nil {
		log.Printf("Something went wrong decoding %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hashedPass, err := auth.HashPassword(uReq.Password)
	if err != nil {
		log.Printf("Something went wrong hashing %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          uReq.Email,
		HashedPassword: hashedPass,
	})
	if err != nil {
		log.Printf("Something went wrong creating the user %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uData, err := json.Marshal(user)
	if err != nil {
		log.Printf("Something went wrong marshaling %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(uData)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	cfg.fileserveHits = atomic.Int32{}
	w.WriteHeader(http.StatusOK)
	count, err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		log.Printf("Something went wrong resetting %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Hits reset to 0\n%d users deleted\n", count)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var uReq userAuth
	err := decoder.Decode(&uReq)
	if err != nil {
		log.Printf("Something went wrong decoding %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), uReq.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	matches, err := auth.CheckPasswordHash(uReq.Password, user.HashedPassword)
	if err != nil || !matches {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		log.Printf("Something went wrong making jwt %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	userWNoPass := struct {
		database.CreateUserRow
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}{
		CreateUserRow: database.CreateUserRow{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	}

	resData, err := json.Marshal(userWNoPass)
	if err != nil {
		log.Printf("Something went wrong marshaling %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resData)

	_, err = cfg.db.AddRefreshToken(r.Context(), database.AddRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		log.Printf("Something went wrong adding refresh token %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) handlerUpdateLoginInfo(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var uReq userAuth
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&uReq)
	if err != nil {
		log.Printf("Something went wrong decoding json %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hashed_password, err := auth.HashPassword(uReq.Password)
	if err != nil {
		log.Printf("Something went wrong hashing password %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	updtdUser, err := cfg.db.UpdateEmailPassword(r.Context(), database.UpdateEmailPasswordParams{
		Email: uReq.Email,
		HashedPassword: hashed_password,
		ID: userId,
	})

	resData, err := json.Marshal(updtdUser)
	if err != nil {
		log.Printf("Something went wrong marshaling json %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resData)
}
