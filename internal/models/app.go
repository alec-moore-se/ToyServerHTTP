package models

import (
	"github.com/alec-moore-se/ToyServerHTTP/internal/database"
	"github.com/google/uuid"
	"sync/atomic"
	"time"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaSecret    string
}

type UserType struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

// ------------- Setters --------------
func (cfg *ApiConfig) SetFileserverHits(hits *atomic.Int32) {
	cfg.fileserverHits.Store(hits.Load())
}

func (cfg *ApiConfig) SetDb(db *database.Queries) {
	cfg.db = db
}

func (cfg *ApiConfig) SetPlatform(platform string) {
	cfg.platform = platform
}

func (cfg *ApiConfig) SetJwtSecret(jwtSecret string) {
	cfg.jwtSecret = jwtSecret
}

func (cfg *ApiConfig) SetPolkaSecret(polkaSecret string) {
	cfg.polkaSecret = polkaSecret
}
