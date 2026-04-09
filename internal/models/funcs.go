package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alec-moore-se/ToyServerHTTP/internal/auth"
	"github.com/alec-moore-se/ToyServerHTTP/internal/database"
	"net/http"
	"time"
)

func (cfg *ApiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Content-Type is not application/json")
		return
	}

	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	var p params

	err := decoder.Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error creating json")
		return
	}

	hp, err := auth.HashPassword(p.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error hashing password")
		fmt.Println(err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          p.Email,
		HashedPassword: hp,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating user\n")
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserType{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	})
}

func (cfg *ApiConfig) Login(w http.ResponseWriter, r *http.Request) {
	type login struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	var p login
	err := decoder.Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error creating json")
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	user, err := cfg.db.GetUserwEmail(r.Context(), p.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error getting user")
		fmt.Println(err)
		return
	}
	match, err := auth.CheckPasswordHash(p.Password, user.HashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error checking password")
		fmt.Printf("User Pass: %s\n", user.HashedPassword)
		fmt.Printf("tried Pass: %s\n", p.Password)
		fmt.Println(err)
		return
	}
	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error making JWT")
		fmt.Println(err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error making refresh token")
		fmt.Println(err)
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: sql.NullTime{Time: time.Now().Add(time.Hour * 24 * 60), Valid: true},
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating refresh token")
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(UserType{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt.Time,
		UpdatedAt:    user.UpdatedAt.Time,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed.Bool,
	})
}

func (cfg *ApiConfig) Refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	type params struct {
		RefreshToken string `json:"refresh_token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fullToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if fullToken.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if fullToken.ExpiresAt.Time.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, errr := cfg.db.GetUserIDwToken(r.Context(), token)
	if errr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jwtToken, err := auth.MakeJWT(userID, cfg.jwtSecret, time.Hour)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"token\": " + "\"" + jwtToken + "\"}"))
}

func (cfg *ApiConfig) Revoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), token)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *ApiConfig) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Content-Type is not application/json")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Error getting token")
		fmt.Println(err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Error getting user ID")
		fmt.Println(err)
		return
	}

	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	var p params

	err = decoder.Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error creating json")
		return
	}

	user, err := cfg.db.GetUser(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error getting user")
		fmt.Println(err)
		return
	}

	hp, err := auth.HashPassword(p.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error hashing password")
		fmt.Println(err)
		return
	}

	err = cfg.db.UpdateUserEmailaPass(r.Context(), database.UpdateUserEmailaPassParams{
		ID:             userID,
		Email:          p.Email,
		HashedPassword: hp,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating user")
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserType{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   time.Now(),
		Email:       p.Email,
		Token:       token,
		IsChirpyRed: user.IsChirpyRed.Bool,
	})
}
