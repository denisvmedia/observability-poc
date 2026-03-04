package apiserver

import (
	"encoding/json"
	"net/http"

	"github.com/denisvmedia/observability-poc/registry"
)

func versionsHandler(reg registry.SessionRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		versions, err := reg.ListVersions(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list versions"}`, http.StatusInternalServerError)
			return
		}

		// Always return an array, never null.
		if versions == nil {
			versions = make([]string, 0)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(versions)
	}
}
