package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/alec-moore-se/ToyServerHTTP/internal/auth"
	"github.com/alec-moore-se/ToyServerHTTP/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

type UserType struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		_ = fmt.Errorf("database failed to open: %w", err)
		return
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       os.Getenv("PLATFORM"),
		jwtSecret:      os.Getenv("JWT_SECRET"),
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.createUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirp)
	mux.HandleFunc("POST /api/login", apiCfg.Login)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if r.Header.Get("Content-Type") != "application/json" {
		fmt.Fprintf(w, "Content-Type is not application/json")
		w.WriteHeader(http.StatusBadRequest)
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
		fmt.Fprintf(w, "Error creating json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), p.Email)
	fmt.Println(user)

	if err != nil {
		fmt.Fprintf(w, "Error creating user")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.HashedPassword, err = auth.HashPassword(p.Password)
	if err != nil {
		fmt.Fprintf(w, "Error hashing password")
	}
	err = cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             user.ID,
		HashedPassword: user.HashedPassword,
	},
	)
	if err != nil {
		fmt.Fprintf(w, "Error hashing password")
		fmt.Println(err)
		user.HashedPassword = ""
	}

	fmt.Fprintf(w, `{"id": "%s", "created_at": "%s", "updated_at": "%s", "email": "%s"}`,
		user.ID, user.CreatedAt.Time, user.UpdatedAt.Time, user.Email)
}

func (cfg *apiConfig) Login(w http.ResponseWriter, r *http.Request) {
	type login struct {
		Email     string        `json:"email"`
		Password  string        `json:"password"`
		ExpiresIn time.Duration `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	var p login
	err := decoder.Decode(&p)
	if err != nil {
		fmt.Fprintf(w, "Error creating json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.ExpiresIn == 0 || p.ExpiresIn > time.Hour {
		p.ExpiresIn = time.Hour
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	user, err := cfg.db.GetUserwEmail(r.Context(), p.Email)
	if err != nil {
		fmt.Fprintf(w, "Error getting user")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	match, err := auth.CheckPasswordHash(p.Password, user.HashedPassword)
	if err != nil {
		fmt.Fprintf(w, "Error checking password")
		fmt.Printf("User Pass: %s\n", user.HashedPassword)
		fmt.Printf("tried Pass: %s\n", p.Password)
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, p.ExpiresIn)
	if err != nil {
		fmt.Fprintf(w, "Error making JWT")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(UserType{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email,
		Token:     token,
	})
}
