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

	apiMux.HandleFunc("GET /healthz", metricsMiddleware(http.HandlerFunc(app.HandlerReadiness)).ServeHTTP)
	apiMux.HandleFunc("POST /chirps", metricsMiddleware(http.HandlerFunc(app.HandlerChirps)).ServeHTTP)
	apiMux.HandleFunc("POST /users", metricsMiddleware(http.HandlerFunc(app.HandlerUsers)).ServeHTTP)
	return apiMux
}
func serveAdminMux(app *app.Application) *http.ServeMux {
	adminMux := http.NewServeMux()

	adminMux.HandleFunc("GET /metrics", app.HandlerMetrics)
	adminMux.HandleFunc("POST /reset", app.HandlerResetUsers)

	return adminMux
}
