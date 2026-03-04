//go:build with_frontend

package apiserver

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/denisvmedia/observability-poc/frontend"
)

func frontendHandler() http.Handler {
	dist := frontend.GetDist()
	fsys, _ := fs.Sub(dist, "dist")
	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Derive the clean relative path for an fs.Stat lookup.
		// path.Clean prevents path traversal; TrimPrefix strips the leading slash.
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}

		if _, err := fs.Stat(fsys, name); err == nil {
			// File exists in the embedded FS — serve it directly.
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: unmatched paths (e.g. /dashboard) get index.html.
		data, err := dist.ReadFile("dist/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})
}

