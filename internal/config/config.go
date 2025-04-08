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
}

func Load() *ApiConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("jwt secret not defined")
	}
	return &ApiConfig{
		WebStaticDir: "./web/static",
		JWTSecret:    secret,
	}
}
