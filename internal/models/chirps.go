package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

func (cfg *ApiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Content-Type is not application/json")
		return
	}

	type jsonBody struct {
		Body string `json:"body"`
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
		w.WriteHeader(500)
		returner.Error = "Something went wrong"
		data, _ := json.Marshal(returner)
		w.Write(data)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		returner.Error = "Error getting token"
		data, _ := json.Marshal(returner)
		w.Write(data)
		return
	}

	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		returner.Error = "Error getting user ID"
		data, _ := json.Marshal(returner)
		w.Write(data)
		return
	}

	if len(params.Body) > 140 {
		returner.Error = "Chirp is too long"
		data, err := json.Marshal(returner)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
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
		UserID: userUUID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating chirp")
		fmt.Println(err)
		return
	}

	resp := userChirp{
		ID:        userChirps.ID,
		CreatedAt: userChirps.CreatedAt,
		UpdatedAt: userChirps.UpdatedAt,
		Body:      userChirps.Body,
		UserID:    userChirps.UserID,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (cfg *ApiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	author_id := r.URL.Query().Get("author_id")
	userID, err := uuid.Parse(author_id)
	if err != nil {
		userID = uuid.Nil
	}

	sortQuery := r.URL.Query().Get("sort")
	if sortQuery != "desc" {
		sortQuery = "asc"
	}

	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error getting chirps")
		fmt.Println(err)
		return
	}

	var resp []userChirp
	for _, chirp := range chirps {
		if userID == uuid.Nil || userID == chirp.UserID {
			resp = append(resp, userChirp{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			})
		}
	}

	if sortQuery == "desc" {
		sort.Slice(resp, func(i, j int) bool {
			return resp[i].CreatedAt.After(resp[j].CreatedAt)
		})
	} else {
		sort.Slice(resp, func(i, j int) bool {
			return resp[i].CreatedAt.Before(resp[j].CreatedAt)
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (cfg *ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing chirpID")
		fmt.Println(err)
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

func (cfg *ApiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Error getting token 1")
		fmt.Println(err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Error getting user ID")
		fmt.Println(err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing chirpID")
		fmt.Println(err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Error getting chirp")
		fmt.Println(err)
		return
	}

	if chirp.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Error getting chirp")
		fmt.Println(err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error deleting chirp")
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
