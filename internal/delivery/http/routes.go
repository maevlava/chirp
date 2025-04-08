package http

import (
	"github.com/maevlava/chirpy/internal/app"
	"net/http"
)

func NewRouter(app *app.Application) http.Handler {
	mux := serveFileServerMux(app)
	apiMux := serveApiMux(app)
	adminMux := serveAdminMux(app)

	apiHandler := http.StripPrefix("/api", apiMux)
	adminHandler := http.StripPrefix("/admin", adminMux)

	mux.Handle("/admin/", adminHandler)
	mux.Handle("/api/", apiHandler)

	return mux
}
func serveFileServerMux(app *app.Application) *http.ServeMux {
	//fileServerPath

	mux := http.NewServeMux()

	metricsMiddleware := app.MiddlewareMetricsInc
	fileSeverHandler := http.FileServer(http.Dir(app.Config.WebStaticDir))
	handlerWithMetrics := metricsMiddleware(fileSeverHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/app", http.StatusFound)
			return
		}
	})

	webAppHandler := metricsMiddleware(http.HandlerFunc(app.HandlerWebApp))
	mux.Handle("/app", webAppHandler)

	mux.Handle("/app/", http.StripPrefix("/app/", handlerWithMetrics))

	return mux
}
func serveApiMux(app *app.Application) *http.ServeMux {
	// non file server paths
	metricsMiddleware := app.MiddlewareMetricsInc
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /chirps", metricsMiddleware(http.HandlerFunc(app.HandlerGetChirps)).ServeHTTP)
	apiMux.HandleFunc("POST /chirps", metricsMiddleware(http.HandlerFunc(app.HandlerChirps)).ServeHTTP)
	apiMux.HandleFunc("PUT /users", metricsMiddleware(http.HandlerFunc(app.HandlerUserUpdate)).ServeHTTP)
	apiMux.HandleFunc("GET /chirps/{chirpId}", metricsMiddleware(http.HandlerFunc(app.HandlerGetChirpByID)).ServeHTTP)
	apiMux.HandleFunc("DELETE /chirps/{chirpId}", metricsMiddleware(http.HandlerFunc(app.HandlerDeleteChirpByID)).ServeHTTP)
	apiMux.HandleFunc("GET /healthz", metricsMiddleware(http.HandlerFunc(app.HandlerReadiness)).ServeHTTP)
	apiMux.HandleFunc("POST /login", metricsMiddleware(http.HandlerFunc(app.HandlerLogin)).ServeHTTP)
	apiMux.HandleFunc("POST /users", metricsMiddleware(http.HandlerFunc(app.HandlerUsers)).ServeHTTP)
	apiMux.HandleFunc("POST /refresh", metricsMiddleware(http.HandlerFunc(app.HandlerRefreshToken)).ServeHTTP)
	apiMux.HandleFunc("POST /revoke", metricsMiddleware(http.HandlerFunc(app.HandlerRevokeToken)).ServeHTTP)
	apiMux.HandleFunc("POST /polka/webhooks", metricsMiddleware(http.HandlerFunc(app.HandlerPolkaWebhooks)).ServeHTTP)
	return apiMux
}
func serveAdminMux(app *app.Application) *http.ServeMux {
	adminMux := http.NewServeMux()

	adminMux.HandleFunc("GET /metrics", app.HandlerMetrics)
	adminMux.HandleFunc("POST /reset", app.HandlerResetUsers)

	return adminMux
}
