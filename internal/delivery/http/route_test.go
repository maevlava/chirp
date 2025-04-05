// In internal/delivery/http/routes_test.go
package http_test // Use _test package convention

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// Adjust these import paths to match your project structure
	"github.com/maevlava/chirpy/internal/app"
	"github.com/maevlava/chirpy/internal/config"
	httpdelivery "github.com/maevlava/chirpy/internal/delivery/http"
)

// TestHealthzMethods specifically checks the method handling for /healthz
func TestHealthzMethods(t *testing.T) {
	// --- Setup ---
	// Create minimal dependencies needed JUST to initialize the router
	// Use the same Load/NewApplication as your main.go
	cfg := config.Load() // Assumes Load() exists and returns *config.ApiConfig
	// Ensure WebStaticDir is set if NewRouter depends on it, even if not used in this test path
	if cfg.WebStaticDir == "" {
		cfg.WebStaticDir = "../../../web/static" // Adjust relative path if needed from this test file's location
	}
	appInstance := app.NewApplication(cfg)

	// Use the EXACT NewRouter function from your routes.go file
	// *** IMPORTANT: Ensure the version of NewRouter being tested ***
	// *** DOES NOT have the metricsMiddleware applied to /healthz ***
	// *** It should have: mux.Handle("GET /healthz", readinessHandler) ***
	testRouter := httpdelivery.NewRouter(appInstance)

	// --- Test GET Request ---
	reqGet, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}
	rrGet := httptest.NewRecorder()
	testRouter.ServeHTTP(rrGet, reqGet)

	// Check GET status code
	if status := rrGet.Code; status != http.StatusOK {
		t.Errorf("GET /healthz handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// --- Test POST Request ---
	reqPost, err := http.NewRequest("POST", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}
	rrPost := httptest.NewRecorder()
	testRouter.ServeHTTP(rrPost, reqPost)

	// Check POST status code <<<< THIS IS THE CRITICAL CHECK >>>>
	expectedStatus := http.StatusMethodNotAllowed
	if status := rrPost.Code; status != expectedStatus {
		// Include the response body in the error if it's not 405, might give clues
		t.Errorf("POST /healthz handler returned wrong status code: got %v want %v\nResponse Body: %s",
			status, expectedStatus, rrPost.Body.String())
	}
}
