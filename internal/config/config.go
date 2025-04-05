package config

import (
	"github.com/maevlava/chirpy/internal/database"
	"sync/atomic"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
	WebStaticDir   string
	DB             *database.Queries
}

func Load() *ApiConfig {
	return &ApiConfig{
		WebStaticDir: "./web/static",
	}
}
