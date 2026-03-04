//go:build with_frontend

package apiserver

import (
	"io/fs"
	"net/http"
	"net/http/httptest"

	"github.com/denisvmedia/observability-poc/frontend"
)

func frontendHandler() http.Handler {
	dist := frontend.GetDist()
	fsys, _ := fs.Sub(dist, "dist")
	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := httptest.NewRecorder()
		fileServer.ServeHTTP(recorder, r)

		if recorder.Code == http.StatusOK {
			for k, v := range recorder.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(recorder.Code)
			_, _ = w.Write(recorder.Body.Bytes())
			return
		}

		// SPA fallback: return index.html for all unmatched routes.
		data, err := dist.ReadFile("dist/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})
}

