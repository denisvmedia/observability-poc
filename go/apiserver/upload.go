package apiserver

import (
	"encoding/json"
	"net/http"

	"github.com/denisvmedia/observability-poc/registry"
	"github.com/denisvmedia/observability-poc/services/ingestion"
)

const maxUploadBytes = 32 << 20 // 32 MB

type uploadResponse struct {
	RowsInserted int      `json:"rows_inserted"`
	RowsSkipped  int      `json:"rows_skipped"`
	Errors       []string `json:"errors"`
}

func uploadHandler(reg registry.SessionRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			writeJSONError(w, http.StatusBadRequest, "request too large or not multipart")
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "missing file field")
			return
		}
		defer file.Close()

		result, err := ingestion.Ingest(r.Context(), file, reg)
		if err != nil {
			writeJSONError(w, http.StatusUnprocessableEntity, "ingestion failed: "+err.Error())
			return
		}

		errors := result.Errors
		if errors == nil {
			errors = make([]string, 0)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(uploadResponse{
			RowsInserted: result.RowsInserted,
			RowsSkipped:  result.RowsSkipped,
			Errors:       errors,
		})
	}
}
