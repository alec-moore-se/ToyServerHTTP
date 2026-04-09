package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/alec-moore-se/ToyServerHTTP/internal/database"
	app "github.com/alec-moore-se/ToyServerHTTP/internal/models"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

	apiCfg := app.ApiConfig{}
	apiCfg.SetFileserverHits(&atomic.Int32{})
	apiCfg.SetDb(dbQueries)
	apiCfg.SetPlatform(os.Getenv("PLATFORM"))
	apiCfg.SetJwtSecret(os.Getenv("JWT_SECRET"))
	apiCfg.SetPolkaSecret(os.Getenv("POLKA_KEY"))

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", app.HandlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirp)
	mux.HandleFunc("POST /api/login", apiCfg.Login)
	mux.HandleFunc("POST /api/refresh", apiCfg.Refresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.Revoke)
	mux.HandleFunc("PUT /api/users/", apiCfg.UpdateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.DeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.PolkaWebhook)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
