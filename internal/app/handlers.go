package app

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/maevlava/chirpy/internal/auth"
	"github.com/maevlava/chirpy/internal/database"
	httputil "github.com/maevlava/chirpy/internal/delivery/httputil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}
type RefreshTokenResponse struct {
	Token string `json:"token"`
}

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

func (app *Application) HandlerUsers(w http.ResponseWriter, r *http.Request) {
	type EmailParam struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	param := EmailParam{}
	err := decoder.Decode(&param)
	if err != nil {
		_ = httputil.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
	}

	hashedPassword, err := auth.HashPassword(param.Password)
	newUser := database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Email:          param.Email,
		HashedPassword: hashedPassword,
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
func (app *Application) HandlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	//update params
	type UpdateUserParam struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// auth
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(tokenString, app.Config.JWTSecret)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)
	param := UpdateUserParam{}
	err = decoder.Decode(&param)
	if err != nil {
		httputil.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
	}

	hashedPassword, err := auth.HashPassword(param.Password)
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updateUserParam := database.UpdateUserParams{
		ID:             userID,
		Email:          param.Email,
		HashedPassword: hashedPassword,
	}
	updatedUser, err := app.Config.DB.UpdateUser(r.Context(), updateUserParam)
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	userResponse := UserResponse{
		Email:     updatedUser.Email,
		UpdatedAt: updatedUser.UpdatedAt.Format(time.RFC3339),
	}
	httputil.RespondWithJSON(w, http.StatusOK, userResponse)
	return
}

func (app *Application) HandlerChirps(w http.ResponseWriter, r *http.Request) {
	// authentication
	tokenstring, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(tokenstring, app.Config.JWTSecret)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	const maxChirpLength = 200
	type ChripParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := ChripParams{}

	err = decoder.Decode(&params)
	if err != nil {
		httputil.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	} else if len(params.Body) > maxChirpLength {
		httputil.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	if params.Body == "" {
		httputil.RespondWithError(w, http.StatusBadRequest, "Chirp body cannot be empty")
		return
	}
	if userID == uuid.Nil {
		httputil.RespondWithError(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}

	// if valid
	createChirpParams := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      params.Body,
		UserID:    userID,
	}
	createdChirp, err := app.Config.DB.CreateChirp(r.Context(), createChirpParams)
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	httputil.RespondWithJSON(w, http.StatusCreated, createdChirp)
}
func (app *Application) HandlerGetChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	var err error

	authorIDStr := r.URL.Query().Get("author_id")
	if authorIDStr == "" {
		chirps, err = app.Config.DB.GetAllChirps(r.Context())
		if err != nil {
			httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
	} else {
		authorID, err := uuid.Parse(authorIDStr)
		if err != nil {
			httputil.RespondWithError(w, http.StatusBadRequest, "Author ID is invalid")
			return
		}
		chirps, err = app.Config.DB.GetChirpsByAuthor(r.Context(), authorID)
	}
	sortOrder := r.URL.Query().Get("sort")

	if strings.ToLower(sortOrder) == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}

	httputil.RespondWithJSON(w, http.StatusOK, chirps)
}
func (app *Application) HandlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIdPath := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(chirpIdPath)
	if err != nil {
		log.Fatalf("Error parsing chirpId: %v", err)
		return
	}

	chirp, err := app.Config.DB.GetChirpById(r.Context(), chirpId)
	if err != nil {
		httputil.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	httputil.RespondWithJSON(w, http.StatusOK, chirp)
}
func (app *Application) HandlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	//auth
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userID, err := auth.ValidateJWT(tokenString, app.Config.JWTSecret)
	if err != nil {
		httputil.RespondWithError(w, http.StatusForbidden, err.Error())
		return
	}
	chirpIdPath := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(chirpIdPath)
	if err != nil {
		log.Fatalf("Error parsing chirpId: %v", err)
		return
	}
	chirp, err := app.Config.DB.GetChirpById(r.Context(), chirpId)
	if err != nil {
		httputil.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if chirp.UserID != userID {
		httputil.RespondWithError(w, http.StatusForbidden, "User ID inconsistent")
		return
	}

	err = app.Config.DB.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.RespondWithJSON(w, http.StatusNoContent, "")

}
func (app *Application) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	// parse the request
	type LoginParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := LoginParams{}
	err := decoder.Decode(&params)
	if err != nil {
		httputil.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
	}
	// get user by email
	user, err := app.Config.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}
	// match password with the database
	err = auth.CheckPassword(user.HashedPassword, params.Password)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, "Password incorrect")
		return
	}

	// make JWT
	accessTokenDuration := 1 * time.Hour
	accessTokenString, err := auth.MakeJWT(user.ID, app.Config.JWTSecret, accessTokenDuration)
	if err != nil {
		log.Printf("ERROR generating access token for user %s: %v", user.Email, err)
		httputil.RespondWithError(w, http.StatusInternalServerError, "Could not generate access token")
		return
	}
	// make refresh token
	refreshTokenString, err := auth.RefreshToken()
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// safe refresh token to database 60 days
	refreshTokenExpiry := time.Now().UTC().Add(60 * 24 * time.Hour)
	_, err = app.Config.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: refreshTokenExpiry,
	})
	if err != nil {
		log.Printf("ERROR storing refresh token for user %s: %v", user.ID, err)
		httputil.RespondWithError(w, http.StatusInternalServerError, "Could not store session")
		return
	}

	// if match return OK status and json
	response := UserResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
		Email:        user.Email,
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
		IsChirpyRed:  user.IsChirpyRed,
	}

	w.WriteHeader(http.StatusOK)
	httputil.RespondWithJSON(w, http.StatusOK, response)

}
func (app *Application) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {

	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid token")
		return
	}

	user, err := app.Config.DB.GetUserForRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired session")
		return
	}

	newAccessTokenDuration := 1 * time.Hour
	newAccessTokenString, err := auth.MakeJWT(user.ID, app.Config.JWTSecret, newAccessTokenDuration)
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, "Could not refresh token")
		return
	}

	response := RefreshTokenResponse{Token: newAccessTokenString}
	httputil.RespondWithJSON(w, http.StatusOK, response)
}
func (app *Application) HandlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid token")
		return
	}

	err = app.Config.DB.RevokeRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, "Could not revoke session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *Application) HandlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	// auth api key
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		httputil.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid apikey")
		return
	}
	if apiKey != app.Config.PolkaApiKey {
		httputil.RespondWithError(w, http.StatusUnauthorized, "Invalid apikey")
		return
	}

	type PolkaRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	params := PolkaRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		httputil.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		httputil.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	_, err = app.Config.DB.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		httputil.RespondWithError(w, http.StatusNotFound, "Could not upgrade user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// util
