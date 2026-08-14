package vent

import (
	"embed"
	"net/http"
)

//go:embed static
var static embed.FS

// StaticDirHandler returns an HTTP handler that serves static files embedded in the vent package.
func StaticDirHandler() http.Handler {
	fileServer := http.FileServerFS(static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fileServer.ServeHTTP(w, r)
	})
}
