package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

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

	userWNoPass := database.CreateUserRow{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
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
}
