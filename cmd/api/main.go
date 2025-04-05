package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/maevlava/chirpy/internal/app"
	"github.com/maevlava/chirpy/internal/config"
	"github.com/maevlava/chirpy/internal/database"
	httpdelivery "github.com/maevlava/chirpy/internal/delivery/http"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8080"

func main() {
	err := godotenv.Load()
	if err != nil {
		_ = fmt.Errorf("Error loading .env file")
	}

	cfg := config.Load()
	loadDB(cfg)

	appInstance := app.NewApplication(cfg)

	router := httpdelivery.NewRouter(appInstance)
	server := http.Server{
		Addr:    ":" + defaultPort,
		Handler: router,
	}

	fmt.Println("Server listening on port ", defaultPort)
	log.Fatal(server.ListenAndServe())
}

func loadDB(cfg *config.ApiConfig) {
	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)
	cfg.DB = database.New(db)
}
