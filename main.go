package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alec-moore-se/ToyServerHTTP/internal/database"
	"github.com/alec-moore-se/ToyServerHTTP/internal/database/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

type UserType struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		_ = fmt.Errorf("Database failed to open: %w", err)
		return
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       os.Getenv("PLATFORM"),
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.createUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirp)

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
		Email string `json:"email"`
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

	fmt.Fprintf(w, `{"id": "%s", "created_at": "%s", "updated_at": "%s", "email": "%s"}`, user.ID, user.CreatedAt, user.UpdatedAt, user.Email)
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
		Body    string        `json:"body"`
		User_id uuid.NullUUID `json:"user_id"`
	}
	type returners struct {
		Valid       bool   `json:"valid"`
		Error       string `json:"error"`
		CleanedBody string `json:"cleaned_body"`
	}

	type userChirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := jsonBody{}
	returner := returners{}
	userChirps := userChirp{}
	err := decoder.Decode(&params)
	if err != nil {
		returner.Error = "Something went wrong"
		w.WriteHeader(500)
		return
	}

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
	clean_body := strings.ToLower(params.Body)
	listOfWordsClean := strings.Split(clean_body, " ")
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

	userChirps = cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{returner.CleanedBody, params.User_id})

	data, err := json.Marshal(returner)
	if err != nil {
		returner.Error = "Something went wrong"
		w.WriteHeader(500)
		return
	}
	w.Write(data)
}
