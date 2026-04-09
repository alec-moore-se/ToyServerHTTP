package models

import (
	"encoding/json"
	"fmt"
	"github.com/alec-moore-se/ToyServerHTTP/internal/auth"
	"github.com/google/uuid"
	"net/http"
)

func (cfg *ApiConfig) PolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Error getting api key")
		fmt.Println(err)
		return
	}

	if apiKey != cfg.polkaSecret {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Error getting api key")
		fmt.Println(err)
		return
	}

	type polkaSent struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	var p polkaSent

	err = decoder.Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error creating json")
		return
	}

	if p.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userID := p.Data.UserID

	err = cfg.db.PutUserChirpyRed(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Error putting user chirpy red")
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
