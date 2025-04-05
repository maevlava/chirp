package app

import (
	"github.com/maevlava/chirpy/internal/config"
)

type Application struct {
	Config *config.ApiConfig
}

func NewApplication(cfg *config.ApiConfig) *Application {
	return &Application{
		Config: cfg,
	}
}
