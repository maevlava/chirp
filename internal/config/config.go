package config

import (
	"github.com/maevlava/chirpy/internal/database"
	"log"
	"os"
	"sync/atomic"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
	WebStaticDir   string
	DB             *database.Queries
	JWTSecret      string
	PolkaApiKey    string
}

func Load() *ApiConfig {
	secret := os.Getenv("JWT_SECRET")
	PolkaAPIKey := os.Getenv("POLKA_KEY")
	if secret == "" || PolkaAPIKey == "" {
		log.Fatal("env missing value")
	}
	return &ApiConfig{
		WebStaticDir: "./web/static",
		JWTSecret:    secret,
		PolkaApiKey:  PolkaAPIKey,
	}
}
