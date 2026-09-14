package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/FooNameBar/chirpy/internal/database"
	"github.com/google/uuid"
)

type reqBody struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type errResp struct {
	Error string `json:"error"`
}

var bannedWords []string = []string{"kerfuffle", "sharbert", "fornax"}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var data reqBody
	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("Something went wrong decoding %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var respData []byte
	var status int
	if len(data.Body) > 140 {
		respData, err = json.Marshal(errResp{Error: "Chirp is too long"})
		if err != nil {
			log.Printf("json.Marshal error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		status = http.StatusBadRequest
	} else {
		corrected := maskBannedWords(data.Body)
		chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			UserID: data.UserID,
			Body: corrected,
		})
		if err != nil {
			log.Printf("db.CreateChirp error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respData, err = json.Marshal(chirp)
		if err != nil {
			log.Printf("json.Marshal error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		status = http.StatusCreated
	}
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(status)
	w.Write(respData)
}

func maskBannedWords(msg string) string {
	mask := "****"
	corrected := strings.Split(msg, " ")
	for _, bw := range bannedWords {
		for i, w := range corrected {
			lowered := strings.ToLower(w)
			if lowered == bw {
				corrected[i] = mask
			}
		}
	}
	return strings.Join(corrected, " ")
}
