package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type eventReq struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handleUpgradeUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)

	var eReq eventReq
	err := decoder.Decode(&eReq)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if eReq.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), eReq.Data.UserID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = cfg.db.UpdateChirpyRed(r.Context(), user.ID)
	if err != nil {
		log.Printf("Something went wrong upgrading the user to chirpy red %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
