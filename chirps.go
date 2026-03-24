package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alec-moore-se/ToyServerHTTP/internal/auth"
	"github.com/alec-moore-se/ToyServerHTTP/internal/database"
	"github.com/google/uuid"
)

type userChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if r.Header.Get("Content-Type") != "application/json" {
		fmt.Fprintf(w, "Content-Type is not application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type jsonBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type returners struct {
		Valid       bool   `json:"valid"`
		Error       string `json:"error"`
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := jsonBody{}
	returner := returners{}
	err := decoder.Decode(&params)
	if err != nil {
		returner.Error = "Something went wrong"
		w.WriteHeader(500)
		return
	}

	//--------------
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		returner.Error = "Error getting token"
	}
	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		returner.Error = "Error validating token"
	}
	if params.UserID != userUUID {
		returner.Error = "User ID does not match"
	}
	//--------------

	if len(params.Body) > 140 {
		returner.Error = "Chirp is too long"
		w.WriteHeader(400)
		data, err := json.Marshal(returner)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write(data)
		return
	}
	cleanBody := strings.ToLower(params.Body)
	listOfWordsClean := strings.Split(cleanBody, " ")
	listOfWordsDirty := strings.Split(params.Body, " ")
	for i := range listOfWordsClean {
		if listOfWordsClean[i] == "kerfuffle" {
			listOfWordsClean[i] = "****"
		}
		if listOfWordsClean[i] == "sharbert" {
			listOfWordsClean[i] = "****"
		}
		if listOfWordsClean[i] == "fornax" {
			listOfWordsClean[i] = "****"
		}
		if listOfWordsClean[i] != "****" {
			listOfWordsClean[i] = listOfWordsDirty[i]
		}
	}
	returner.CleanedBody = strings.Join(listOfWordsClean, " ")

	w.Header().Set("Content-Type", "application/json")
	returner.Valid = true

	userChirps, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   returner.CleanedBody,
		UserID: params.UserID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := userChirp{
		ID:        userChirps.ID,
		CreatedAt: userChirps.CreatedAt,
		UpdatedAt: userChirps.UpdatedAt,
		Body:      userChirps.Body,
		UserID:    userChirps.UserID,
	}

	json.NewEncoder(w).Encode(resp)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		fmt.Fprintf(w, "Error getting chirps")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var resp []userChirp
	for _, chirp := range chirps {
		resp = append(resp, userChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	json.NewEncoder(w).Encode(resp)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		fmt.Fprintf(w, "Error parsing chirpID")
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Error getting chirp")
		fmt.Println(err)
		return
	}

	resp := userChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
