package app

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/maevlava/chirpy/internal/database"
	httputil "github.com/maevlava/chirpy/internal/delivery/httputil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (app *Application) HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}
func (app *Application) HandlerWebApp(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(app.Config.WebStaticDir, "index.html")
	http.ServeFile(w, r, indexPath)
}

func (app *Application) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	serverHits := app.Config.FileServerHits.Load()

	metricsPath := filepath.Join(app.Config.WebStaticDir, "admin/metrics.html")
	htmlBytes, err := os.ReadFile(metricsPath)
	if err != nil {
		log.Printf("Error reading metrics file: %v", err)
		return
	}

	metricsHTMLContent := string(htmlBytes)
	htmlResponse := fmt.Sprintf(metricsHTMLContent, serverHits)
	_, err = w.Write([]byte(htmlResponse))
	if err != nil {
		log.Printf("Error writing admin metrics response: %v", err)
		return
	}

}
func (app *Application) HandlerValidate(w http.ResponseWriter, r *http.Request) {
	type ValidateParameters struct {
		Body string `json:"body"`
	}
	type CleanedValidateChirpResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := ValidateParameters{}

	err := decoder.Decode(&params)
	// if something went wrong
	if err != nil {
		httputil.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	// if chirp is too long
	const maxChirpLength = 100
	if len(params.Body) > maxChirpLength {
		httputil.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
	}

	cleaned := cleanProfanity(params.Body)
	respPayload := CleanedValidateChirpResponse{
		CleanedBody: cleaned,
	}
	httputil.RespondWithJSON(w, http.StatusOK, respPayload)
}

func (app *Application) HandlerUsers(w http.ResponseWriter, r *http.Request) {
	type EmailParam struct {
		Email string `json:"email"`
	}
	type UserResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	param := EmailParam{}
	err := decoder.Decode(&param)
	if err != nil {
		_ = httputil.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
	}

	newUser := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Email:     param.Email,
	}

	db := app.Config.DB
	createdUser, err := db.CreateUser(r.Context(), newUser)
	if err != nil {
		_ = httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	response := UserResponse{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: createdUser.UpdatedAt.Format(time.RFC3339),
		Email:     createdUser.Email,
	}

	_ = httputil.RespondWithJSON(w, http.StatusCreated, response)
}
func (app *Application) HandlerResetUsers(w http.ResponseWriter, r *http.Request) {
	platform := os.Getenv("PLATFORM")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	_ = app.Config.DB.DeleteAllUsers(r.Context())
	w.WriteHeader(http.StatusOK)

	app.Config.FileServerHits.Store(0)
}

// util
func cleanProfanity(text string) string {
	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	replacement := "****"

	words := strings.Split(text, " ")

	for i, word := range words {
		if _, found := profaneWords[strings.ToLower(word)]; found {
			words[i] = replacement
		}
	}

	return strings.Join(words, " ")
}
