// Package apiserver wires together the HTTP router and all handler functions.
package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/denisvmedia/observability-poc/registry"
)

// New constructs the HTTP handler for the observability API.
// It mounts all routes and attaches the required middleware.
func New(reg registry.SessionRegistry) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Get("/healthz", healthzHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/upload", uploadHandler(reg))
		r.Get("/versions", versionsHandler(reg))
		r.Get("/dashboard", dashboardHandler(reg))
	})

	r.Handle("/*", frontendHandler())

	return r
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "ok")
}

// writeJSONError writes a JSON-encoded {"error": msg} response.
// Using json.NewEncoder ensures msg is properly escaped, avoiding any risk
// of malformed JSON from special characters in error strings.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: msg})
}
